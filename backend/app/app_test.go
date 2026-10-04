package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/caiofdev/sortly/backend/apperr"
	"github.com/caiofdev/sortly/backend/organizer"
	"github.com/caiofdev/sortly/backend/organizer/criteria"
	"github.com/caiofdev/sortly/backend/settings"
	"github.com/caiofdev/sortly/backend/store"
	"github.com/caiofdev/sortly/backend/undo"
)

var testNow = time.Date(2026, 3, 5, 12, 0, 0, 0, time.UTC)

type fakeOrganizer struct {
	result organizer.Result
	err    error
	got    organizer.Request
	calls  int
	during func()
}

func (f *fakeOrganizer) Organize(_ context.Context, req organizer.Request) (organizer.Result, error) {
	f.got = req
	f.calls++
	if f.during != nil {
		f.during()
	}
	return f.result, f.err
}

type fakeUndoer struct {
	result undo.Result
	err    error
}

func (f fakeUndoer) Undo(context.Context) (undo.Result, error) { return f.result, f.err }

type fakeRecords struct {
	op  *store.Operation
	err error
}

func (f fakeRecords) Load() (*store.Operation, error) { return f.op, f.err }

type fakeSettings struct {
	current settings.Settings
	err     error
}

func (f *fakeSettings) Get() settings.Settings { return f.current }

func (f *fakeSettings) SetLanguage(lang string) (settings.Settings, error) {
	if f.err != nil {
		return f.current, f.err
	}
	f.current.Language = lang
	return f.current, nil
}

func (f *fakeSettings) SetCriterion(key string, enabled bool) (settings.Settings, error) {
	if f.err != nil {
		return f.current, f.err
	}
	f.current.Options, _ = f.current.Options.With(key, enabled)
	return f.current, nil
}

// newTestApp preenche as dependências que o teste não informou com fakes neutros.
func newTestApp(d Deps) (*App, *bytes.Buffer) {
	var logs bytes.Buffer
	d.Logger = slog.New(slog.NewTextHandler(&logs, nil))
	d.Now = func() time.Time { return testNow }
	if d.Records == nil {
		d.Records = fakeRecords{}
	}
	if d.Settings == nil {
		d.Settings = &fakeSettings{current: settings.Default()}
	}
	return New(d), &logs
}

func last(s ViewState) Notification {
	if len(s.Notifications) == 0 {
		return Notification{}
	}
	return s.Notifications[0]
}

func TestNewDefaults(t *testing.T) {
	a := New(Deps{})
	if a.log == nil || a.ctx == nil || a.deps.Now == nil {
		t.Fatal("New deveria preencher logger, contexto e relógio padrão")
	}
}

func TestGetState(t *testing.T) {
	record := &store.Operation{SourceFolderPath: `C:\origem`, DestinationFolderPath: `C:\destino`,
		MovedItems: []store.MovedItem{{From: "a", To: "b"}}}
	cases := []struct {
		name    string
		records fakeRecords
		want    ViewState
		notice  string
	}{
		{"sem registro", fakeRecords{}, ViewState{}, ""},
		{"registro com 1 item recupera os caminhos", fakeRecords{op: record},
			ViewState{SourceFolderPath: `C:\origem`, DestinationFolderPath: `C:\destino`, HasUndo: true}, CodeRecovered},
		{"registro com 0 itens não oferece desfazer", fakeRecords{op: &store.Operation{SourceFolderPath: "x"}}, ViewState{}, ""},
		{"registro corrompido", fakeRecords{err: store.ErrCorrupted}, ViewState{}, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a, _ := newTestApp(Deps{Records: tc.records})
			got := a.GetState()
			if got.SourceFolderPath != tc.want.SourceFolderPath || got.DestinationFolderPath != tc.want.DestinationFolderPath ||
				got.HasUndo != tc.want.HasUndo || last(got).Code != tc.notice {
				t.Fatalf("estado = %+v, want %+v com aviso %q", got, tc.want, tc.notice)
			}
		})
	}
}

func TestGetStateLoadsRecordOnce(t *testing.T) {
	op := &store.Operation{SourceFolderPath: "x", MovedItems: []store.MovedItem{{From: "a", To: "b"}}}
	a, _ := newTestApp(Deps{Records: fakeRecords{op: op}})
	a.GetState()
	if got := a.GetState(); len(got.Notifications) != 1 {
		t.Fatalf("o aviso de recuperação deveria aparecer uma vez só: %+v", got.Notifications)
	}
}

func TestGetStateIncludesSettings(t *testing.T) {
	prefs := &fakeSettings{current: settings.Settings{Language: "en", Options: criteria.Options{ByDate: true}}}
	a, _ := newTestApp(Deps{Settings: prefs})
	got := a.GetState().Settings
	if got.Language != "en" || len(got.Criteria) != len(criteria.Keys) || !got.Criteria[3].Locked {
		t.Fatalf("Settings = %+v", got)
	}
}

func TestSelectFolders(t *testing.T) {
	var titles []string
	picked := map[string]string{SourceDialogTitle: `C:\Fotos`, DestinationDialogTitle: `C:\Destino`}
	a, _ := newTestApp(Deps{PickDir: func(_ context.Context, title string) (string, error) {
		titles = append(titles, title)
		return picked[title], nil
	}})

	a.SelectSource()
	got := a.SelectDestination()

	if got.SourceFolderPath != `C:\Fotos` || got.DestinationFolderPath != `C:\Destino` {
		t.Fatalf("estado = %+v", got)
	}
	if len(titles) != 2 || titles[0] != SourceDialogTitle || titles[1] != DestinationDialogTitle {
		t.Fatalf("títulos = %v", titles)
	}
}

func TestSelectFolderCancelled(t *testing.T) {
	a, _ := newTestApp(Deps{PickDir: func(context.Context, string) (string, error) { return "", nil }})
	a.state.SourceFolderPath = `C:\anterior`
	if got := a.SelectSource(); got.SourceFolderPath != `C:\anterior` || len(got.Notifications) != 0 {
		t.Fatalf("cancelar não deveria mudar nada: %+v", got)
	}
}

func TestSelectFolderError(t *testing.T) {
	a, logs := newTestApp(Deps{PickDir: func(context.Context, string) (string, error) {
		return "", errors.New("diálogo quebrou")
	}})
	got := last(a.SelectDestination())
	if got.Kind != KindError || got.Code != apperr.CodeUnexpected || got.Action != ActionSelectDestination || !got.At.Equal(testNow) {
		t.Fatalf("aviso = %+v", got)
	}
	if !strings.Contains(logs.String(), "diálogo quebrou") {
		t.Errorf("o detalhe deveria ir para o log: %s", logs.String())
	}
}

func TestDropPaths(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}

	t.Run("sem itens não faz nada", func(t *testing.T) {
		a, _ := newTestApp(Deps{})
		if got := a.DropPaths(nil); got.SourceFolderPath != "" || len(got.Notifications) != 0 {
			t.Fatalf("estado = %+v", got)
		}
	})

	t.Run("arquivo: a pasta dele vira a origem; usa o primeiro item", func(t *testing.T) {
		a, _ := newTestApp(Deps{})
		got := a.DropPaths([]string{file, filepath.Join(dir, "outro")})
		if got.SourceFolderPath != dir || last(got).Code != CodeSourceDropped || last(got).Path != dir {
			t.Fatalf("estado = %+v", got)
		}
	})

	t.Run("item inexistente vira aviso com código", func(t *testing.T) {
		a, _ := newTestApp(Deps{})
		got := a.DropPaths([]string{filepath.Join(dir, "sumiu")})
		if got.SourceFolderPath != "" || last(got).Code != "DROPPED_MISSING" || last(got).Action != ActionDrop {
			t.Fatalf("estado = %+v", got)
		}
	})
}

func TestOrganizeWithoutSource(t *testing.T) {
	org := &fakeOrganizer{}
	a, _ := newTestApp(Deps{Organizer: org})
	got := a.Organize()
	if org.calls != 0 || last(got).Code != CodeSourceRequired || last(got).Kind != KindError {
		t.Fatalf("estado = %+v, chamadas = %d", got, org.calls)
	}
}

func TestOrganizeUsesSourceAsDefaultDestination(t *testing.T) {
	org := &fakeOrganizer{result: organizer.Result{MovedFiles: 3, CanUndo: true}}
	prefs := &fakeSettings{current: settings.Settings{Options: criteria.Options{ByDate: true}}}
	a, logs := newTestApp(Deps{Organizer: org, Settings: prefs})
	a.state.SourceFolderPath = `C:\origem`

	got := a.Organize()

	want := organizer.Request{SourceFolderPath: `C:\origem`, DestinationFolderPath: `C:\origem`, Options: criteria.Options{ByDate: true}}
	if org.got != want {
		t.Fatalf("pedido = %+v, want %+v", org.got, want)
	}
	n := last(got)
	if !got.HasUndo || got.Busy != "" || n.Code != CodeOrganizeDone || n.Kind != KindOrganize || n.Organize.MovedFiles != 3 {
		t.Fatalf("estado = %+v, aviso = %+v", got, n)
	}
	if !strings.Contains(logs.String(), "movidos=3") {
		t.Errorf("log = %s", logs.String())
	}
}

func TestOrganizeErrorBecomesNotification(t *testing.T) {
	org := &fakeOrganizer{err: fmt.Errorf("%w: %q", organizer.ErrInvalidSource, `C:\nada`)}
	a, logs := newTestApp(Deps{Organizer: org})
	a.state.SourceFolderPath, a.state.DestinationFolderPath, a.state.HasUndo = `C:\nada`, `C:\destino`, true

	got := a.Organize()

	if last(got).Code != "INVALID_SOURCE" || !got.HasUndo || got.Busy != "" || org.got.DestinationFolderPath != `C:\destino` {
		t.Fatalf("estado = %+v", got)
	}
	if !strings.Contains(logs.String(), "nada") {
		t.Errorf("o detalhe deveria ir para o log: %s", logs.String())
	}
}

func TestOrganizeEmitsBusyWhileRunning(t *testing.T) {
	var emitted []ViewState
	org := &fakeOrganizer{result: organizer.Result{CanUndo: true}}
	a, _ := newTestApp(Deps{Organizer: org, Emit: func(_ context.Context, s ViewState) { emitted = append(emitted, s) }})
	a.startup(context.Background())
	a.state.SourceFolderPath = `C:\origem`
	org.during = func() {
		if len(emitted) == 0 || emitted[len(emitted)-1].Busy != BusyOrganize {
			t.Errorf("durante a organização, o último estado emitido deveria estar ocupado: %+v", emitted)
		}
	}

	a.Organize()

	if emitted[len(emitted)-1].Busy != "" {
		t.Fatalf("o último estado emitido deveria estar livre: %+v", emitted[len(emitted)-1])
	}
}

func TestActionInProgressBlocksAnother(t *testing.T) {
	org := &fakeOrganizer{}
	a, _ := newTestApp(Deps{Organizer: org, Undoer: fakeUndoer{}})
	a.state.SourceFolderPath = `C:\origem`
	a.work.Lock()
	defer a.work.Unlock()

	a.Organize()
	got := a.Undo()

	if org.calls != 0 || len(got.Notifications) != 0 {
		t.Fatalf("com outra ação em andamento nada deveria acontecer: chamadas = %d, estado = %+v", org.calls, got)
	}
}

func TestUndo(t *testing.T) {
	a, _ := newTestApp(Deps{Undoer: fakeUndoer{result: undo.Result{RestoredFiles: 2}}})
	a.state.HasUndo = true
	got := a.Undo()
	if got.HasUndo || got.Busy != "" || last(got).Code != CodeUndoDone || last(got).Kind != KindRestore || last(got).Undo.RestoredFiles != 2 {
		t.Fatalf("estado = %+v", got)
	}

	empty, _ := newTestApp(Deps{Undoer: fakeUndoer{err: undo.ErrNothingToUndo}})
	if got := last(empty.Undo()); got.Code != "NOTHING_TO_UNDO" || got.Action != ActionUndo {
		t.Fatalf("aviso = %+v", got)
	}
}

func TestNotificationsLimitAndOrder(t *testing.T) {
	a, _ := newTestApp(Deps{PickDir: func(context.Context, string) (string, error) { return "", errors.New("x") }})
	for range MaxNotifications {
		a.SelectSource()
	}
	got := a.GetState().Notifications
	if len(got) != MaxNotifications || got[0].ID != MaxNotifications || got[MaxNotifications-1].ID != 1 {
		t.Fatalf("80 avisos: %d, primeiro %d, último %d", len(got), got[0].ID, got[len(got)-1].ID)
	}

	got = a.SelectSource().Notifications
	if len(got) != MaxNotifications || got[0].ID != MaxNotifications+1 || got[MaxNotifications-1].ID != 2 {
		t.Fatalf("81º aviso deveria descartar o mais antigo: %d, primeiro %d, último %d", len(got), got[0].ID, got[len(got)-1].ID)
	}

	if cleared := a.ClearNotifications(); len(cleared.Notifications) != 0 {
		t.Fatalf("após limpar: %+v", cleared.Notifications)
	}
}

func TestSnapshotDoesNotShareNotifications(t *testing.T) {
	a, _ := newTestApp(Deps{PickDir: func(context.Context, string) (string, error) { return "", errors.New("x") }})
	got := a.SelectSource()
	got.Notifications[0].Code = "ALTERADO"
	if a.GetState().Notifications[0].Code == "ALTERADO" {
		t.Fatal("o estado devolvido não pode compartilhar a lista interna")
	}
}

func TestSetLanguage(t *testing.T) {
	a, _ := newTestApp(Deps{})
	if got := a.SetLanguage("en"); got.Settings.Language != "en" || len(got.Notifications) != 0 {
		t.Fatalf("estado = %+v", got)
	}

	failing, _ := newTestApp(Deps{Settings: &fakeSettings{current: settings.Default(), err: settings.ErrInvalidLanguage}})
	if got := last(failing.SetLanguage("fr")); got.Code != "INVALID_LANGUAGE" || got.Action != ActionSettings {
		t.Fatalf("aviso = %+v", got)
	}
}

func TestSetCriterion(t *testing.T) {
	a, _ := newTestApp(Deps{})
	got := a.SetCriterion("byDate", true).Settings.Criteria
	if got[3].Key != "byDate" || !got[3].Enabled {
		t.Fatalf("critérios = %+v", got)
	}

	failing, _ := newTestApp(Deps{Settings: &fakeSettings{current: settings.Default(), err: settings.ErrLastCriterion}})
	if got := last(failing.SetCriterion("byExtension", false)); got.Code != "LAST_CRITERION" {
		t.Fatalf("aviso = %+v", got)
	}
}

func TestEmitOnlyAfterStartup(t *testing.T) {
	calls := 0
	a, _ := newTestApp(Deps{Emit: func(context.Context, ViewState) { calls++ }})
	a.GetState()
	if calls != 0 {
		t.Fatal("sem o runtime do Wails (antes do startup), nada deveria ser emitido")
	}
	a.startup(context.Background())
	a.GetState()
	if calls != 1 {
		t.Fatalf("emissões depois do startup = %d, want 1", calls)
	}
}

func TestNewDefault(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	a, err := NewDefault(nil)
	if err != nil {
		t.Fatal(err)
	}
	d := a.deps
	if d.Organizer == nil || d.Undoer == nil || d.Records == nil || d.Settings == nil || d.PickDir == nil || d.Emit == nil {
		t.Fatalf("dependências = %+v", d)
	}
	if want := filepath.Join(home, ".sortly", "last-operation.json"); d.Records.(*store.FileStore).Path() != want {
		t.Fatalf("registro em %s, want %s", d.Records.(*store.FileStore).Path(), want)
	}
}

func TestNewDefaultSavesSettingsNextToRecord(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	a, err := NewDefault(nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := a.SetLanguage("en"); got.Settings.Language != "en" {
		t.Fatalf("estado = %+v", got)
	}
	if _, err := os.Stat(filepath.Join(home, ".sortly", "settings.json")); err != nil {
		t.Fatalf("preferências deveriam ir para ~/.sortly/settings.json: %v", err)
	}
}

func TestNewDefaultWithoutHome(t *testing.T) {
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")
	t.Setenv("home", "")
	if _, err := NewDefault(nil); err == nil {
		t.Fatal("sem pasta do usuário, NewDefault deveria falhar")
	}
}

func TestStateJSONAlwaysHasLists(t *testing.T) {
	a, _ := newTestApp(Deps{})
	for name, state := range map[string]ViewState{"inicial": a.GetState(), "após limpar": a.ClearNotifications()} {
		data, err := json.Marshal(state)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), `"notifications":[]`) || !strings.Contains(string(data), `"criteria":[`) {
			t.Errorf("%s: listas deveriam ser arrays, não null: %s", name, data)
		}
	}
}

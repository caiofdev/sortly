package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/caiofdev/sortly/backend/apperr"
	"github.com/caiofdev/sortly/backend/organizer"
	"github.com/caiofdev/sortly/backend/organizer/criteria"
	"github.com/caiofdev/sortly/backend/settings"
	"github.com/caiofdev/sortly/backend/store"
	"github.com/caiofdev/sortly/backend/undo"
)

type fakeOrganizer struct {
	result organizer.Result
	err    error
	got    organizer.Request
}

func (f *fakeOrganizer) Organize(_ context.Context, req organizer.Request) (organizer.Result, error) {
	f.got = req
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

func newTestApp(d Deps) (*App, *bytes.Buffer) {
	var logs bytes.Buffer
	d.Logger = slog.New(slog.NewTextHandler(&logs, nil))
	return New(d), &logs
}

func TestNewWithoutLogger(t *testing.T) {
	a := New(Deps{})
	if a.log == nil || a.ctx == nil {
		t.Fatal("New deveria preencher logger e contexto padrão")
	}
}

func TestSelectFolders(t *testing.T) {
	var titles []string
	picked := map[string]string{SourceDialogTitle: `C:\Fotos`, DestinationDialogTitle: ""}
	a, _ := newTestApp(Deps{PickDir: func(_ context.Context, title string) (string, error) {
		titles = append(titles, title)
		return picked[title], nil
	}})

	if got, err := a.SelectSourceFolder(); got != `C:\Fotos` || err != nil {
		t.Errorf("origem = (%q, %v)", got, err)
	}
	if got, err := a.SelectDestinationFolder(); got != "" || err != nil {
		t.Errorf("destino cancelado = (%q, %v), want vazio sem erro", got, err)
	}
	if len(titles) != 2 || titles[0] != SourceDialogTitle || titles[1] != DestinationDialogTitle {
		t.Errorf("títulos = %v", titles)
	}

}

func TestSelectFolderError(t *testing.T) {
	failing, logs := newTestApp(Deps{PickDir: func(context.Context, string) (string, error) {
		return "", errors.New("diálogo indisponível")
	}})
	if _, err := failing.SelectSourceFolder(); err == nil || err.Error() != apperr.CodeUnexpected {
		t.Errorf("erro do diálogo = %v, want UNEXPECTED", err)
	}
	if !strings.Contains(logs.String(), "diálogo indisponível") {
		t.Errorf("o erro completo deveria ir para o log: %s", logs.String())
	}
}

func TestResolveDroppedPathFacade(t *testing.T) {
	a, _ := newTestApp(Deps{})
	dir := t.TempDir()

	if got, err := a.ResolveDroppedPath(dir); err != nil || got.SourceFolderPath != dir {
		t.Errorf("pasta = (%+v, %v)", got, err)
	}
	if _, err := a.ResolveDroppedPath(""); err == nil || err.Error() != "DROPPED_INVALID" {
		t.Errorf("vazio = %v, want DROPPED_INVALID", err)
	}
}

func TestGetLastOrganizationState(t *testing.T) {
	oneItem := &store.Operation{
		SourceFolderPath: `C:\origem`, DestinationFolderPath: `C:\destino`,
		MovedItems: []store.MovedItem{{From: "a", To: "b"}},
	}
	cases := []struct {
		name    string
		records fakeRecords
		want    OrganizationState
	}{
		{"1 item: pode desfazer", fakeRecords{op: oneItem}, OrganizationState{true, `C:\origem`, `C:\destino`}},
		{"0 itens: nada a desfazer", fakeRecords{op: &store.Operation{SourceFolderPath: `C:\origem`}}, OrganizationState{}},
		{"sem registro", fakeRecords{}, OrganizationState{}},
		{"corrompido", fakeRecords{err: store.ErrCorrupted}, OrganizationState{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a, _ := newTestApp(Deps{Records: tc.records})
			if got := a.GetLastOrganizationState(); got != tc.want {
				t.Fatalf("estado = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestOrganizeFiles(t *testing.T) {
	t.Run("sucesso", func(t *testing.T) {
		org := &fakeOrganizer{result: organizer.Result{MovedFiles: 3, CanUndo: true}}
		prefs := &fakeSettings{current: settings.Settings{Options: criteria.Options{ByDate: true}}}
		a, logs := newTestApp(Deps{Organizer: org, Settings: prefs})

		got, err := a.OrganizeFiles(`C:\origem`, `C:\destino`)

		want := organizer.Request{SourceFolderPath: `C:\origem`, DestinationFolderPath: `C:\destino`, Options: criteria.Options{ByDate: true}}
		if err != nil || got.MovedFiles != 3 || org.got != want {
			t.Fatalf("OrganizeFiles = (%+v, %v)", got, err)
		}
		if !strings.Contains(logs.String(), "movidos=3") {
			t.Errorf("log = %s", logs.String())
		}
	})

	t.Run("erro com código chega ao frontend só como código", func(t *testing.T) {
		org := &fakeOrganizer{err: fmt.Errorf("%w: %q", organizer.ErrInvalidSource, `C:\nada`)}
		a, logs := newTestApp(Deps{Organizer: org, Settings: &fakeSettings{}})

		_, err := a.OrganizeFiles("", "")

		if err == nil || err.Error() != "INVALID_SOURCE" {
			t.Fatalf("err = %v, want INVALID_SOURCE (sem prefixos)", err)
		}
		if !strings.Contains(logs.String(), "nada") {
			t.Errorf("o detalhe deveria ir para o log: %s", logs.String())
		}
	})

	t.Run("erro sem código vira UNEXPECTED", func(t *testing.T) {
		a, _ := newTestApp(Deps{Organizer: &fakeOrganizer{err: errors.New("pânico controlado")}, Settings: &fakeSettings{}})
		if _, err := a.OrganizeFiles("", ""); err == nil || err.Error() != apperr.CodeUnexpected {
			t.Fatalf("err = %v, want UNEXPECTED", err)
		}
	})
}

func TestGetSettings(t *testing.T) {
	a, _ := newTestApp(Deps{Settings: &fakeSettings{current: settings.Default()}})
	got := a.GetSettings()
	if got.Language != settings.DefaultLanguage || len(got.Criteria) != len(criteria.Keys) {
		t.Fatalf("GetSettings = %+v", got)
	}
}

func TestSetLanguage(t *testing.T) {
	a, _ := newTestApp(Deps{Settings: &fakeSettings{current: settings.Default()}})
	if got, err := a.SetLanguage("en"); err != nil || got.Language != "en" {
		t.Fatalf("SetLanguage = (%+v, %v)", got, err)
	}

	failing, _ := newTestApp(Deps{Settings: &fakeSettings{err: settings.ErrInvalidLanguage}})
	if _, err := failing.SetLanguage("fr"); err == nil || err.Error() != "INVALID_LANGUAGE" {
		t.Fatalf("err = %v, want INVALID_LANGUAGE", err)
	}
}

func TestSetCriterion(t *testing.T) {
	a, _ := newTestApp(Deps{Settings: &fakeSettings{current: settings.Default()}})
	got, err := a.SetCriterion("byDate", true)
	if err != nil || got.Criteria[3].Key != "byDate" || !got.Criteria[3].Enabled {
		t.Fatalf("SetCriterion = (%+v, %v)", got, err)
	}

	failing, _ := newTestApp(Deps{Settings: &fakeSettings{err: settings.ErrLastCriterion}})
	if _, err := failing.SetCriterion("byExtension", false); err == nil || err.Error() != "LAST_CRITERION" {
		t.Fatalf("err = %v, want LAST_CRITERION", err)
	}
}

func TestUndoLastOrganization(t *testing.T) {
	a, _ := newTestApp(Deps{Undoer: fakeUndoer{result: undo.Result{RestoredFiles: 2}}})
	if got, err := a.UndoLastOrganization(); err != nil || got.RestoredFiles != 2 {
		t.Fatalf("Undo = (%+v, %v)", got, err)
	}

	empty, _ := newTestApp(Deps{Undoer: fakeUndoer{err: undo.ErrNothingToUndo}})
	if _, err := empty.UndoLastOrganization(); err == nil || err.Error() != "NOTHING_TO_UNDO" {
		t.Fatalf("err = %v, want NOTHING_TO_UNDO", err)
	}
}

func TestNewDefault(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	a, err := NewDefault(nil)
	if err != nil || a.deps.Organizer == nil || a.deps.Undoer == nil || a.deps.Records == nil || a.deps.Settings == nil || a.deps.PickDir == nil {
		t.Fatalf("NewDefault = (%+v, %v)", a, err)
	}
	if got := a.GetLastOrganizationState(); got.HasUndo {
		t.Fatalf("estado = %+v", got)
	}
	if want := filepath.Join(home, ".sortly", "last-operation.json"); a.deps.Records.(*store.FileStore).Path() != want {
		t.Fatalf("registro em %s, want %s", a.deps.Records.(*store.FileStore).Path(), want)
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
	if _, err := a.SetLanguage("en"); err != nil {
		t.Fatal(err)
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

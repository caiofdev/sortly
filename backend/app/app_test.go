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
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/wailsapp/wails/v2/pkg/options"

	"github.com/caiofdev/sortly/backend/apperr"
	"github.com/caiofdev/sortly/backend/history"
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
	// Roda no lugar do executor: recebe o contexto (para Cancelar) e o callback de
	// progresso (#78).
	run func(ctx context.Context, report organizer.Reporter)

	preview     func(ctx context.Context) (organizer.Preview, error)
	previewMu   sync.Mutex
	previewReqs []organizer.Request
}

func (f *fakeOrganizer) Preview(ctx context.Context, req organizer.Request) (organizer.Preview, error) {
	f.previewMu.Lock()
	f.previewReqs = append(f.previewReqs, req)
	f.previewMu.Unlock()
	if f.preview == nil {
		return organizer.Preview{TotalFiles: 1, Folders: []organizer.FolderCount{{Name: "pdf", Count: 1}}}, nil
	}
	return f.preview(ctx)
}

func (f *fakeOrganizer) previewCalls() int {
	f.previewMu.Lock()
	defer f.previewMu.Unlock()
	return len(f.previewReqs)
}

func (f *fakeOrganizer) OrganizeReporting(ctx context.Context, req organizer.Request, report organizer.Reporter) (organizer.Result, error) {
	f.got = req
	f.calls++
	if f.during != nil {
		f.during()
	}
	if f.run != nil {
		f.run(ctx, report)
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

type fakeHistory struct {
	entries []history.Entry
	err     error
	undone  int
}

func (f *fakeHistory) List() []history.Entry { return append([]history.Entry{}, f.entries...) }

func (f *fakeHistory) Add(e history.Entry) error {
	if f.err != nil {
		return f.err
	}
	f.entries = append([]history.Entry{e}, f.entries...)
	return nil
}

func (f *fakeHistory) MarkLastUndone() error {
	f.undone++
	if f.err != nil {
		return f.err
	}
	if len(f.entries) > 0 {
		f.entries[0].Status = history.StatusUndone
	}
	return nil
}

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

func (f *fakeSettings) SetTheme(theme string) (settings.Settings, error) {
	if f.err != nil {
		return f.current, f.err
	}
	f.current.Theme = theme
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
	d.Now = func() time.Time { return testNow }
	if d.Records == nil {
		d.Records = fakeRecords{}
	}
	if d.Settings == nil {
		d.Settings = &fakeSettings{current: settings.Default()}
	}
	if d.Organizer == nil {
		d.Organizer = &fakeOrganizer{}
	}
	if d.History == nil {
		d.History = &fakeHistory{}
	}
	if d.Background == nil {
		d.Background = func(task func()) { task() }
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
	if !got.HasUndo || got.Busy != "" || n.Code != CodeOrganizeDone || n.Kind != KindSuccess || n.Organize.MovedFiles != 3 {
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

func TestOrganizeErrorAndUndo(t *testing.T) {
	tests := []struct {
		name   string
		result organizer.Result
		err    error
		want   bool
	}{
		// Regressão (#53): o desfazer apontaria para a organização anterior.
		{"registro não salvo desabilita o desfazer anterior",
			organizer.Result{MovedFiles: 2}, fmt.Errorf("%w: disco cheio", organizer.ErrRecordNotSaved), false},
		{"erro com arquivos movidos e registro salvo habilita o desfazer",
			organizer.Result{MovedFiles: 1, CanUndo: true}, context.Canceled, true},
		{"erro sem arquivos movidos mantém o desfazer anterior",
			organizer.Result{}, organizer.ErrNoCriteria, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, _ := newTestApp(Deps{Organizer: &fakeOrganizer{result: tt.result, err: tt.err}})
			a.state.SourceFolderPath, a.state.HasUndo = `C:\origem`, true

			got := a.Organize()

			if got.HasUndo != tt.want || last(got).Kind != KindError {
				t.Fatalf("HasUndo = %v, want %v; aviso = %+v", got.HasUndo, tt.want, last(got))
			}
		})
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

func TestOrganizeReportsProgressThrottled(t *testing.T) {
	clock := testNow
	var progress []organizer.Progress
	org := &fakeOrganizer{result: organizer.Result{MovedFiles: 3, CanUndo: true}}
	org.run = func(_ context.Context, report organizer.Reporter) {
		steps := []struct {
			after time.Duration
			p     organizer.Progress
		}{
			{0, organizer.Progress{Done: 0, Total: 3, File: "a.pdf", Folder: "pdf"}},
			{49 * time.Millisecond, organizer.Progress{Done: 1, Total: 3, File: "b.pdf", Folder: "pdf"}},
			{50 * time.Millisecond, organizer.Progress{Done: 2, Total: 3, File: "c.txt", Folder: "txt"}},
			{time.Millisecond, organizer.Progress{Done: 3, Total: 3}},
		}
		for _, s := range steps {
			clock = clock.Add(s.after)
			report(s.p)
		}
	}
	a, _ := newTestApp(Deps{Organizer: org, Emit: func(_ context.Context, s ViewState) {
		if s.Busy == BusyOrganize && s.Progress.Total > 0 {
			progress = append(progress, s.Progress)
		}
	}})
	a.deps.Now = func() time.Time { return clock }
	a.startup(context.Background())
	a.state.SourceFolderPath = `C:\origem`

	got := a.Organize()

	// 0 sai sempre; 1 chega 49 ms depois e é pulado; 2 chega 99 ms depois do
	// último emitido; 3 é o fim e sai mesmo 1 ms depois.
	var done []int
	for _, p := range progress {
		done = append(done, p.Done)
	}
	if !reflect.DeepEqual(done, []int{0, 2, 3}) || progress[0].File != "a.pdf" {
		t.Fatalf("progresso emitido = %+v", progress)
	}
	if got.Progress != (organizer.Progress{}) {
		t.Fatalf("depois de organizar, o progresso zera: %+v", got.Progress)
	}
}

func TestCancel(t *testing.T) {
	tests := []struct {
		name     string
		result   organizer.Result
		err      error
		wantCode string
		wantKind string
		wantUndo bool
	}{
		{"cancelar vira aviso com o que foi movido", organizer.Result{MovedFiles: 1, CanUndo: true, Canceled: true},
			context.Canceled, CodeOrganizeCanceled, KindInfo, true},
		{"registro não salvo vale mais que cancelar", organizer.Result{MovedFiles: 1, Canceled: true},
			errors.Join(context.Canceled, organizer.ErrRecordNotSaved), "RECORD_NOT_SAVED", KindError, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var a *App
			org := &fakeOrganizer{result: tt.result, err: tt.err}
			org.run = func(ctx context.Context, _ organizer.Reporter) {
				a.Cancel()
				if ctx.Err() == nil {
					t.Error("Cancel deveria cancelar o contexto da organização")
				}
			}
			a, _ = newTestApp(Deps{Organizer: org})
			a.state.SourceFolderPath = `C:\origem`

			got := a.Organize()

			n := last(got)
			if n.Code != tt.wantCode || n.Kind != tt.wantKind || got.HasUndo != tt.wantUndo {
				t.Fatalf("estado = %+v, aviso = %+v", got, n)
			}
		})
	}
}

func TestLastResult(t *testing.T) {
	tests := []struct {
		name   string
		result organizer.Result
		err    error
		want   bool
	}{
		{"arquivos movidos mostram o Concluído", organizer.Result{MovedFiles: 1, CanUndo: true}, nil, true},
		{"com falhas, mas algo movido, também", organizer.Result{MovedFiles: 2, FailedFiles: 1, CanUndo: true}, nil, true},
		{"nada movido fica na tela inicial", organizer.Result{}, nil, false},
		{"cancelado fica na tela inicial", organizer.Result{MovedFiles: 3, Canceled: true, CanUndo: true}, context.Canceled, false},
		{"erro fica na tela inicial", organizer.Result{MovedFiles: 1}, organizer.ErrRecordNotSaved, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, _ := newTestApp(Deps{Organizer: &fakeOrganizer{result: tt.result, err: tt.err}})
			a.state.SourceFolderPath = `C:\origem`

			got := a.Organize()

			if (got.LastResult != nil) != tt.want {
				t.Fatalf("lastResult = %+v, want presente = %v", got.LastResult, tt.want)
			}
		})
	}
}

func TestLastResultIsCleared(t *testing.T) {
	tests := []struct {
		name   string
		action func(a *App) ViewState
	}{
		{"desfazer", func(a *App) ViewState { return a.Undo() }},
		{"organizar outra pasta", func(a *App) ViewState { return a.StartOver() }},
		{"escolher outra origem", func(a *App) ViewState { return a.SelectSource() }},
		{"arrastar outra origem", func(a *App) ViewState { return a.DropPaths([]string{t.TempDir()}) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, _ := newTestApp(Deps{
				Organizer: &fakeOrganizer{result: organizer.Result{MovedFiles: 1, CanUndo: true}},
				Undoer:    fakeUndoer{result: undo.Result{RestoredFiles: 1}},
				PickDir:   func(context.Context, string) (string, error) { return `C:\outra`, nil },
			})
			a.state.SourceFolderPath = `C:\origem`
			a.Organize()

			if got := tt.action(a); got.LastResult != nil {
				t.Fatalf("o Concluído deveria sumir: %+v", got.LastResult)
			}
		})
	}
}

func TestStartOver(t *testing.T) {
	a, _ := newTestApp(Deps{Organizer: &fakeOrganizer{result: organizer.Result{MovedFiles: 1, CanUndo: true}}})
	a.state.SourceFolderPath, a.state.DestinationFolderPath = `C:\origem`, `C:\destino`
	a.Organize()

	got := a.StartOver()

	if got.SourceFolderPath != "" || got.DestinationFolderPath != `C:\destino` || !got.HasUndo || got.Preview.Status != PreviewNone {
		t.Fatalf("volta à tela inicial sem origem, com o destino e o desfazer: %+v", got)
	}
}

func TestOpenDestinationWithoutPath(t *testing.T) {
	opened := 0
	a, _ := newTestApp(Deps{
		Organizer:  &fakeOrganizer{result: organizer.Result{MovedFiles: 1, CanUndo: true}},
		OpenFolder: func(string) error { opened++; return nil },
	})
	a.state.SourceFolderPath = "origem"
	a.Organize()

	if a.OpenDestination(); opened != 0 {
		t.Fatal("resultado sem destino não abre nada")
	}
}

func TestOpenDestination(t *testing.T) {
	done := organizer.Result{MovedFiles: 1, CanUndo: true, DestinationFolderPath: `C:\destino`}
	tests := []struct {
		name     string
		organize bool
		openErr  error
		want     []string
		wantCode string
	}{
		{"sem resultado, não abre nada", false, nil, nil, ""},
		{"abre o destino do resultado", true, nil, []string{`C:\destino`}, ""},
		{"falha vira aviso", true, ErrDestinationNotFound, []string{`C:\destino`}, "DESTINATION_NOT_FOUND"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var opened []string
			a, _ := newTestApp(Deps{
				Organizer:  &fakeOrganizer{result: done},
				OpenFolder: func(path string) error { opened = append(opened, path); return tt.openErr },
			})
			a.state.SourceFolderPath = `C:\origem`
			if tt.organize {
				a.Organize()
			}

			got := a.OpenDestination()

			if !reflect.DeepEqual(opened, tt.want) {
				t.Fatalf("abriu %v, want %v", opened, tt.want)
			}
			if tt.wantCode != "" && (last(got).Code != tt.wantCode || last(got).Action != ActionOpenDestination) {
				t.Fatalf("aviso = %+v", last(got))
			}
		})
	}
}

func TestHistory(t *testing.T) {
	tests := []struct {
		name       string
		result     organizer.Result
		err        error
		wantStatus string
	}{
		{"organizar entra como concluída", organizer.Result{MovedFiles: 3, CanUndo: true}, nil, history.StatusDone},
		{"interrompida com arquivos movidos entra", organizer.Result{MovedFiles: 2, CanUndo: true, Canceled: true}, context.Canceled, history.StatusCanceled},
		{"nada movido não entra", organizer.Result{}, nil, ""},
		{"interrompida sem nada movido não entra", organizer.Result{Canceled: true}, context.Canceled, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hist := &fakeHistory{}
			a, _ := newTestApp(Deps{Organizer: &fakeOrganizer{result: organizer.Result{
				SourceFolderPath: `C:\origem`, DestinationFolderPath: `C:\destino`,
				MovedFiles: tt.result.MovedFiles, CanUndo: tt.result.CanUndo, Canceled: tt.result.Canceled,
			}, err: tt.err}, History: hist})
			a.state.SourceFolderPath = `C:\origem`

			got := a.Organize().History

			if tt.wantStatus == "" {
				if len(got) != 0 {
					t.Fatalf("histórico = %+v, want vazio", got)
				}
				return
			}
			want := history.Entry{At: testNow, SourceFolderPath: `C:\origem`, DestinationFolderPath: `C:\destino`,
				MovedFiles: tt.result.MovedFiles, Status: tt.wantStatus}
			if len(got) != 1 || got[0] != want {
				t.Fatalf("histórico = %+v, want [%+v]", got, want)
			}
		})
	}
}

func TestHistoryAfterUndo(t *testing.T) {
	hist := &fakeHistory{entries: []history.Entry{{MovedFiles: 1, Status: history.StatusDone}}}
	a, _ := newTestApp(Deps{Undoer: fakeUndoer{result: undo.Result{RestoredFiles: 1}}, History: hist})

	if got := a.Undo().History; got[0].Status != history.StatusUndone {
		t.Fatalf("o desfazer marca a mais recente: %+v", got)
	}

	failing, _ := newTestApp(Deps{Undoer: fakeUndoer{err: undo.ErrNothingToUndo}, History: hist})
	failing.Undo()
	if hist.undone != 1 {
		t.Fatalf("desfazer que falhou não marca nada: %d marcações", hist.undone)
	}
}

func TestHistoryFailureOnlyLogs(t *testing.T) {
	hist := &fakeHistory{err: errors.New("disco cheio")}
	a, logs := newTestApp(Deps{
		Organizer: &fakeOrganizer{result: organizer.Result{MovedFiles: 1, CanUndo: true}},
		Undoer:    fakeUndoer{result: undo.Result{RestoredFiles: 1}},
		History:   hist,
	})
	a.state.SourceFolderPath = `C:\origem`

	organized := a.Organize()
	undone := a.Undo()

	if last(organized).Code != CodeOrganizeDone || last(undone).Code != CodeUndoDone {
		t.Fatalf("falha no histórico não pode virar aviso: %+v / %+v", last(organized), last(undone))
	}
	if strings.Count(logs.String(), "disco cheio") != 2 {
		t.Fatalf("as duas falhas deveriam ir para o log: %s", logs.String())
	}
}

func TestCancelWithoutOrganizeDoesNothing(t *testing.T) {
	a, _ := newTestApp(Deps{})
	if got := a.Cancel(); got.Busy != "" || len(got.Notifications) != 0 {
		t.Fatalf("estado = %+v", got)
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
	if got.HasUndo || got.Busy != "" || last(got).Code != CodeUndoDone || last(got).Kind != KindInfo || last(got).Undo.RestoredFiles != 2 {
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

func TestUnread(t *testing.T) {
	a, _ := newTestApp(Deps{PickDir: func(context.Context, string) (string, error) { return "", errors.New("x") }})
	if a.GetState().Unread {
		t.Fatal("sem avisos, nada deveria estar não lido")
	}
	if !a.SelectSource().Unread {
		t.Fatal("um aviso novo deveria ficar não lido")
	}
	read := a.MarkNotificationsRead()
	if read.Unread || len(read.Notifications) != 1 {
		t.Fatalf("marcar como lidas apaga o ponto e mantém a lista: %+v", read)
	}
	if !a.SelectSource().Unread {
		t.Fatal("um aviso depois de ler deveria ficar não lido de novo")
	}
	if a.ClearNotifications().Unread {
		t.Fatal("limpar não deixa nada para ler")
	}
}

func TestNotificationKind(t *testing.T) {
	organize := []struct {
		name   string
		result organizer.Result
		want   string
	}{
		{"arquivos movidos", organizer.Result{MovedFiles: 1}, KindSuccess},
		{"nada movido", organizer.Result{}, KindInfo},
		{"uma falha", organizer.Result{MovedFiles: 3, FailedFiles: 1}, KindError},
	}
	for _, tt := range organize {
		t.Run("organizar: "+tt.name, func(t *testing.T) {
			if got := organizeKind(tt.result); got != tt.want {
				t.Fatalf("organizeKind = %q, want %q", got, tt.want)
			}
		})
	}
	undone := []struct {
		name   string
		result undo.Result
		want   string
	}{
		{"sem falhas", undo.Result{RestoredFiles: 2}, KindInfo},
		{"uma falha", undo.Result{RestoredFiles: 2, FailedFiles: 1}, KindError},
	}
	for _, tt := range undone {
		t.Run("desfazer: "+tt.name, func(t *testing.T) {
			if got := undoKind(tt.result); got != tt.want {
				t.Fatalf("undoKind = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPreviewTriggers(t *testing.T) {
	dir := func(context.Context, string) (string, error) { return `C:\origem`, nil }
	tests := []struct {
		name   string
		action func(a *App)
		want   int
	}{
		{"escolher a origem", func(a *App) { a.SelectSource() }, 1},
		{"escolher o destino", func(a *App) { a.SelectDestination() }, 1},
		{"arrastar", func(a *App) { a.DropPaths([]string{t.TempDir()}) }, 1},
		{"mudar um critério", func(a *App) { a.SetCriterion("byDate", true) }, 1},
		{"organizar", func(a *App) { a.Organize() }, 1},
		{"desfazer", func(a *App) { a.Undo() }, 1},
		{"trocar o idioma não refaz", func(a *App) { a.SetLanguage("en") }, 0},
		{"critério recusado não refaz", func(a *App) {
			a.deps.Settings.(*fakeSettings).err = settings.ErrLastCriterion
			a.SetCriterion("byExtension", false)
		}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			org := &fakeOrganizer{}
			a, _ := newTestApp(Deps{Organizer: org, Undoer: fakeUndoer{}, PickDir: dir})
			a.state.SourceFolderPath = `C:\origem`

			tt.action(a)

			if got := org.previewCalls(); got != tt.want {
				t.Fatalf("prévias = %d, want %d", got, tt.want)
			}
		})
	}
}

// Sem o Background dos testes: a prévia roda numa goroutine e chega sozinha (#77).
func TestPreviewRunsInBackground(t *testing.T) {
	org := &fakeOrganizer{}
	a := New(Deps{Organizer: org, Records: fakeRecords{}, Settings: &fakeSettings{current: settings.Default()}, History: &fakeHistory{},
		PickDir: func(context.Context, string) (string, error) { return `C:\origem`, nil }})

	a.SelectSource()

	deadline := time.Now().Add(5 * time.Second)
	for a.GetState().Preview.Status != PreviewReady {
		if time.Now().After(deadline) {
			t.Fatalf("a prévia não ficou pronta: %+v", a.GetState().Preview)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestPreviewWithoutSource(t *testing.T) {
	org := &fakeOrganizer{}
	a, _ := newTestApp(Deps{Organizer: org, PickDir: func(context.Context, string) (string, error) { return "", nil }})

	a.SetCriterion("byDate", true)
	got := a.SelectSource()

	if org.previewCalls() != 0 || got.Preview.Status != PreviewNone {
		t.Fatalf("sem origem não há prévia: chamadas = %d, estado = %+v", org.previewCalls(), got.Preview)
	}
}

func TestPreviewReady(t *testing.T) {
	var emitted []ViewState
	org := &fakeOrganizer{}
	prefs := &fakeSettings{current: settings.Settings{Options: criteria.Options{ByDate: true}}}
	a, _ := newTestApp(Deps{Organizer: org, Settings: prefs, Emit: func(_ context.Context, s ViewState) { emitted = append(emitted, s) }})
	a.startup(context.Background())
	a.state.DestinationFolderPath = `C:\destino`

	a.DropPaths([]string{t.TempDir()})

	if len(emitted) != 2 || emitted[0].Preview.Status != PreviewLoading {
		t.Fatalf("primeiro o estado contando, depois a prévia: %+v", emitted)
	}
	ready := emitted[1].Preview
	if ready.Status != PreviewReady || ready.TotalFiles != 1 || ready.Folders[0].Name != "pdf" {
		t.Fatalf("prévia = %+v", ready)
	}
	req := org.previewReqs[0]
	if req.DestinationFolderPath != `C:\destino` || req.Options != (criteria.Options{ByDate: true}) {
		t.Fatalf("pedido da prévia = %+v", req)
	}
}

func TestPreviewOnlyNewestCounts(t *testing.T) {
	var queued []func()
	calls := 0
	org := &fakeOrganizer{preview: func(ctx context.Context) (organizer.Preview, error) {
		calls++
		if err := ctx.Err(); err != nil {
			return organizer.Preview{}, err
		}
		return organizer.Preview{TotalFiles: calls, Folders: []organizer.FolderCount{}}, nil
	}}
	a, _ := newTestApp(Deps{Organizer: org, Background: func(task func()) { queued = append(queued, task) }})
	a.state.SourceFolderPath = `C:\origem`

	a.SetCriterion("byDate", true)
	a.SetCriterion("bySize", true)
	queued[1]()
	queued[0]()

	got := a.GetState()
	if got.Preview.Status != PreviewReady || got.Preview.TotalFiles != 1 || len(got.Notifications) != 0 {
		t.Fatalf("a prévia antiga (cancelada) não pode sobrescrever a nova: %+v", got)
	}
}

func TestPreviewStaleResultIsIgnored(t *testing.T) {
	var queued []func()
	org := &fakeOrganizer{preview: func(context.Context) (organizer.Preview, error) {
		return organizer.Preview{TotalFiles: 99, Folders: []organizer.FolderCount{}}, nil
	}}
	a, _ := newTestApp(Deps{Organizer: org, Background: func(task func()) { queued = append(queued, task) }})
	a.state.SourceFolderPath = `C:\origem`

	a.SetCriterion("byDate", true)
	a.Organize()
	queued[0]()

	if got := a.GetState().Preview; got.Status == PreviewReady {
		t.Fatalf("prévia de antes da organização apareceu: %+v", got)
	}
}

func TestPreviewError(t *testing.T) {
	org := &fakeOrganizer{preview: func(context.Context) (organizer.Preview, error) {
		return organizer.Preview{}, fmt.Errorf("%w: sem permissão", organizer.ErrInvalidSource)
	}}
	a, _ := newTestApp(Deps{Organizer: org})
	a.state.SourceFolderPath = `C:\origem`

	if loading := a.SetCriterion("byDate", true); loading.Preview.Status != PreviewLoading {
		t.Fatalf("o binding devolve o estado contando: %+v", loading.Preview)
	}

	got := a.GetState()
	if got.Preview.Status != PreviewNone || last(got).Code != "INVALID_SOURCE" || last(got).Action != ActionPreview {
		t.Fatalf("estado = %+v, aviso = %+v", got.Preview, last(got))
	}
}

// Regressão (#77): um panic ao ler um arquivo na prévia derrubava o app inteiro.
func TestPreviewPanicBecomesNotification(t *testing.T) {
	org := &fakeOrganizer{preview: func(context.Context) (organizer.Preview, error) { panic("mp4 malformado") }}
	a, logs := newTestApp(Deps{Organizer: org})
	a.state.SourceFolderPath = `C:\origem`

	a.SetCriterion("byDate", true)

	got := a.GetState()
	if got.Preview.Status != PreviewNone || last(got).Code != apperr.CodeUnexpected || last(got).Action != ActionPreview {
		t.Fatalf("estado = %+v, aviso = %+v", got.Preview, last(got))
	}
	if !strings.Contains(logs.String(), "mp4 malformado") {
		t.Fatalf("o panic deveria ir para o log: %s", logs.String())
	}
}

func TestPreviewStartsAfterRecovery(t *testing.T) {
	org := &fakeOrganizer{}
	op := &store.Operation{SourceFolderPath: `C:\origem`, MovedItems: []store.MovedItem{{From: "a", To: "b"}}}
	a, _ := newTestApp(Deps{Organizer: org, Records: fakeRecords{op: op}})

	a.GetState()

	if org.previewCalls() != 1 || a.GetState().Preview.Status != PreviewReady {
		t.Fatalf("a origem recuperada deveria ganhar prévia: chamadas = %d", org.previewCalls())
	}
}

func TestOrganizeHidesPreviewWhileRunning(t *testing.T) {
	org := &fakeOrganizer{result: organizer.Result{CanUndo: true}}
	a, _ := newTestApp(Deps{Organizer: org})
	a.state.SourceFolderPath = `C:\origem`
	org.during = func() {
		if p := a.GetState().Preview; p.Status != PreviewNone {
			t.Errorf("durante a organização, a prévia deveria sumir: %+v", p)
		}
	}

	a.Organize()
	if got := a.GetState(); got.Preview.Status != PreviewReady {
		t.Fatalf("depois de organizar, a prévia volta: %+v", got.Preview)
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

func TestStateVersionGrowsOnEveryChange(t *testing.T) {
	var emitted []uint64
	a, _ := newTestApp(Deps{Emit: func(_ context.Context, s ViewState) { emitted = append(emitted, s.Version) }})
	a.startup(context.Background())

	first, second, third := a.GetState(), a.ClearNotifications(), a.SetLanguage("en")

	if first.Version != 1 || second.Version != 2 || third.Version != 3 {
		t.Fatalf("versões = %d, %d, %d; want 1, 2, 3", first.Version, second.Version, third.Version)
	}
	if !reflect.DeepEqual(emitted, []uint64{1, 2, 3}) {
		t.Fatalf("emitidas = %v, want [1 2 3]", emitted)
	}
}

// Regressão (#55): ações simultâneas recebem versões distintas, para que a
// interface consiga descartar o estado que chegar atrasado.
func TestStateVersionUniqueUnderConcurrency(t *testing.T) {
	a, _ := newTestApp(Deps{})
	const n = 50
	versions := make(chan uint64, n)
	var wg sync.WaitGroup
	for range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			versions <- a.ClearNotifications().Version
		}()
	}
	wg.Wait()
	close(versions)

	seen := map[uint64]bool{}
	for v := range versions {
		if seen[v] || v < 1 || v > n {
			t.Fatalf("versão %d repetida ou fora de 1..%d", v, n)
		}
		seen[v] = true
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

func TestSetTheme(t *testing.T) {
	light := options.RGBA{R: 0xf7, G: 0xf7, B: 0xf2, A: 255}
	tests := []struct {
		name    string
		started bool
		err     error
		want    []options.RGBA
	}{
		{"depois do startup, a janela troca de cor", true, nil, []options.RGBA{light}},
		{"antes do startup, não há janela", false, nil, nil},
		{"tema recusado não pinta", true, settings.ErrInvalidTheme, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var painted []options.RGBA
			a, _ := newTestApp(Deps{
				Settings: &fakeSettings{current: settings.Default(), err: tt.err},
				Paint:    func(_ context.Context, c options.RGBA) { painted = append(painted, c) },
			})
			if tt.started {
				a.startup(context.Background())
			}

			got := a.SetTheme("light")

			if !reflect.DeepEqual(painted, tt.want) {
				t.Fatalf("pintou %v, want %v", painted, tt.want)
			}
			if tt.err == nil && got.Settings.Theme != "light" {
				t.Fatalf("tema = %q", got.Settings.Theme)
			}
			if tt.err != nil && last(got).Code != "INVALID_THEME" {
				t.Fatalf("aviso = %+v", last(got))
			}
		})
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
		if !strings.Contains(string(data), `"notifications":[]`) || !strings.Contains(string(data), `"criteria":[`) ||
			!strings.Contains(string(data), `"folders":[]`) || !strings.Contains(string(data), `"history":[]`) {
			t.Errorf("%s: listas deveriam ser arrays, não null: %s", name, data)
		}
	}
}

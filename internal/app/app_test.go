package app

// função                     | CC | casos
// New                        |  2 | com e sem logger
// SelectSourceFolder         |  1 | TestSelectFolders (título de origem)
// SelectDestinationFolder    |  1 | TestSelectFolders (título de destino)
// pick                       |  1 | TestSelectFolders (escolhido; cancelado), TestSelectFolderError
// ResolveDroppedPath         |  1 | TestResolveDroppedPathFacade
// GetLastOrganizationState   |  3 | TestGetLastOrganizationState: com registro; sem itens; corrompido
// OrganizeFiles              |  2 | TestOrganizeFiles: sucesso; erro com código; erro sem código
// UndoLastOrganization       |  2 | TestUndoLastOrganization: sucesso; nada para desfazer
// toFrontend                 |  2 | via os testes acima (nil e erro)
// NewDefault                 |  2 | TestNewDefault: com e sem pasta do usuário
//
// Valor-limite: registro com 0 itens (sem desfazer) e 1 item (com desfazer).

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"testing"

	"github.com/caiofdev/sortly/internal/apperr"
	"github.com/caiofdev/sortly/internal/organizer"
	"github.com/caiofdev/sortly/internal/store"
	"github.com/caiofdev/sortly/internal/undo"
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
		a, logs := newTestApp(Deps{Organizer: org})
		req := organizer.Request{SourceFolderPath: `C:\origem`}

		got, err := a.OrganizeFiles(req)

		if err != nil || got.MovedFiles != 3 || org.got != req {
			t.Fatalf("OrganizeFiles = (%+v, %v)", got, err)
		}
		if !strings.Contains(logs.String(), "movidos=3") {
			t.Errorf("log = %s", logs.String())
		}
	})

	t.Run("erro com código chega ao frontend só como código", func(t *testing.T) {
		org := &fakeOrganizer{err: fmt.Errorf("%w: %q", organizer.ErrInvalidSource, `C:\nada`)}
		a, logs := newTestApp(Deps{Organizer: org})

		_, err := a.OrganizeFiles(organizer.Request{})

		if err == nil || err.Error() != "INVALID_SOURCE" {
			t.Fatalf("err = %v, want INVALID_SOURCE (sem prefixos)", err)
		}
		if !strings.Contains(logs.String(), "nada") {
			t.Errorf("o detalhe deveria ir para o log: %s", logs.String())
		}
	})

	t.Run("erro sem código vira UNEXPECTED", func(t *testing.T) {
		a, _ := newTestApp(Deps{Organizer: &fakeOrganizer{err: errors.New("pânico controlado")}})
		if _, err := a.OrganizeFiles(organizer.Request{}); err == nil || err.Error() != apperr.CodeUnexpected {
			t.Fatalf("err = %v, want UNEXPECTED", err)
		}
	})
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
	if err != nil || a.deps.Organizer == nil || a.deps.Undoer == nil || a.deps.Records == nil || a.deps.PickDir == nil {
		t.Fatalf("NewDefault = (%+v, %v)", a, err)
	}
	// Sem registro, nada a desfazer.
	if got := a.GetLastOrganizationState(); got.HasUndo {
		t.Fatalf("estado = %+v", got)
	}
	if want := filepath.Join(home, ".sortly", "last-operation.json"); a.deps.Records.(*store.FileStore).Path() != want {
		t.Fatalf("registro em %s, want %s", a.deps.Records.(*store.FileStore).Path(), want)
	}

	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")
	t.Setenv("home", "")
	if _, err := NewDefault(nil); err == nil {
		t.Fatal("sem pasta do usuário, NewDefault deveria falhar")
	}
}

package organizer

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/caiofdev/sortly/backend/apperr"
	"github.com/caiofdev/sortly/backend/metadata"
	"github.com/caiofdev/sortly/backend/organizer/criteria"
	"github.com/caiofdev/sortly/backend/store"
)

func TestValidate(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "arquivo.txt")
	writeAt(t, file, "x", testTime)

	cases := []struct {
		name     string
		req      Request
		wantErr  error
		wantDest string
	}{
		{"origem vazia", Request{}, ErrInvalidSource, ""},
		{"origem inexistente", Request{SourceFolderPath: filepath.Join(dir, "nada")}, ErrInvalidSource, ""},
		{"origem é arquivo", Request{SourceFolderPath: file}, ErrInvalidSource, ""},
		{"destino é arquivo", Request{SourceFolderPath: dir, DestinationFolderPath: file}, ErrInvalidDestination, ""},
		{"nenhum critério", Request{SourceFolderPath: dir}, ErrNoCriteria, ""},
		{"destino vazio usa a origem", Request{SourceFolderPath: dir, Options: criteria.Default}, nil, dir},
		{"destino inexistente é aceito", Request{SourceFolderPath: dir, DestinationFolderPath: filepath.Join(dir, "novo"), Options: criteria.Default}, nil, filepath.Join(dir, "novo")},
		{"um critério que não é a extensão", Request{SourceFolderPath: dir, Options: criteria.Options{ByDate: true}}, nil, dir},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, dst, _, err := validate(tc.req)
			if !errors.Is(err, tc.wantErr) || dst != tc.wantDest {
				t.Fatalf("validate = (%q, %v), want (%q, %v)", dst, err, tc.wantDest, tc.wantErr)
			}
		})
	}
}

func TestOrganizeEndToEnd(t *testing.T) {
	src := t.TempDir()
	for _, f := range []string{"a.pdf", "b.txt", "LEIAME"} {
		writeAt(t, filepath.Join(src, f), f, testTime)
	}
	mkdir(t, filepath.Join(src, "subpasta"))
	svc, st := newTestService(t)

	got, err := svc.Organize(context.Background(), Request{SourceFolderPath: src, Options: criteria.Default})
	if err != nil {
		t.Fatal(err)
	}

	want := Result{
		SourceFolderPath: src, DestinationFolderPath: src,
		ProcessedFiles: 3, MovedFiles: 2, IgnoredWithoutExtension: 1, IgnoredFolders: 1, CanUndo: true,
	}
	if got != want {
		t.Fatalf("Organize =\n%+v\nwant\n%+v", got, want)
	}
	assertTree(t, src, []string{"LEIAME", "pdf/a.pdf", "subpasta/", "txt/b.txt"})

	op, err := st.Load()
	if err != nil || len(op.MovedItems) != 2 || op.SourceFolderPath != src {
		t.Fatalf("registro = (%+v, %v)", op, err)
	}
}

func TestOrganizeNothingMovedKeepsPreviousUndo(t *testing.T) {
	src := t.TempDir()
	writeAt(t, filepath.Join(src, "a.txt"), "a", testTime)
	svc, st := newTestService(t)
	if _, err := svc.Organize(context.Background(), Request{SourceFolderPath: src, Options: criteria.Default}); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(st.Path())

	// Segunda vez: só pastas na origem, nada a mover (#8).
	got, err := svc.Organize(context.Background(), Request{SourceFolderPath: src, Options: criteria.Default})

	after, _ := os.ReadFile(st.Path())
	if err != nil || got.MovedFiles != 0 || !got.CanUndo || string(before) != string(after) {
		t.Fatalf("Organize = (%+v, %v); registro mudou: %v (#8)", got, err, string(before) != string(after))
	}
}

func TestOrganizeNothingMovedWithoutPreviousRecord(t *testing.T) {
	svc, st := newTestService(t)
	got, err := svc.Organize(context.Background(), Request{SourceFolderPath: t.TempDir(), Options: criteria.Default})
	if err != nil || got.CanUndo {
		t.Fatalf("Organize = (%+v, %v), want CanUndo false", got, err)
	}
	if _, statErr := os.Stat(st.Path()); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("nenhum registro deveria ser criado, stat = %v", statErr)
	}
}

func TestOrganizePartialFailureIsRecorded(t *testing.T) {
	src := t.TempDir()
	writeAt(t, filepath.Join(src, "a.txt"), "a", testTime)
	writeAt(t, filepath.Join(src, "b.txt"), "b", testTime)
	svc, st := newTestService(t)
	realMove := svc.executor.move
	svc.executor.move = func(from, to string) (string, error) {
		if strings.HasSuffix(from, "b.txt") {
			return "", os.ErrPermission
		}
		return realMove(from, to)
	}

	got, err := svc.Organize(context.Background(), Request{SourceFolderPath: src, Options: criteria.Default})

	op, _ := st.Load()
	if err != nil || got.MovedFiles != 1 || got.FailedFiles != 1 || !got.CanUndo || len(op.MovedItems) != 1 {
		t.Fatalf("Organize = (%+v, %v), registro = %+v (#8)", got, err, op)
	}
}

func TestOrganizeRecordNotSaved(t *testing.T) {
	src := t.TempDir()
	writeAt(t, filepath.Join(src, "a.txt"), "a", testTime)
	blocker := filepath.Join(t.TempDir(), "arquivo")
	writeAt(t, blocker, "", testTime)
	svc := NewService(Deps{Metadata: metadata.Reader{}, Store: store.New(filepath.Join(blocker, "x.json"), nil), Location: testLoc})

	got, err := svc.Organize(context.Background(), Request{SourceFolderPath: src, Options: criteria.Default})

	if !errors.Is(err, ErrRecordNotSaved) || apperr.CodeOf(err) != "RECORD_NOT_SAVED" {
		t.Fatalf("err = %v, want RECORD_NOT_SAVED", err)
	}
	if got.MovedFiles != 1 || got.CanUndo {
		t.Fatalf("Result = %+v: o arquivo foi movido, mas não há como desfazer", got)
	}
}

type saveFailingStore struct {
	previous *store.Operation
}

func (s *saveFailingStore) Load() (*store.Operation, error) { return s.previous, nil }
func (s *saveFailingStore) Save(store.Operation) error      { return errors.New("disco cheio") }
func (s *saveFailingStore) Clear() error                    { s.previous = nil; return nil }

// Regressão (#53): depois de RECORD_NOT_SAVED, o registro da organização
// anterior não pode continuar disponível para desfazer.
func TestOrganizeRecordNotSavedClearsPrevious(t *testing.T) {
	src := t.TempDir()
	writeAt(t, filepath.Join(src, "a.txt"), "a", testTime)
	st := &saveFailingStore{previous: &store.Operation{MovedItems: []store.MovedItem{{From: "x", To: "y"}}}}
	svc := NewService(Deps{Metadata: metadata.Reader{}, Store: st, Location: testLoc})

	got, err := svc.Organize(context.Background(), Request{SourceFolderPath: src, Options: criteria.Default})

	if !errors.Is(err, ErrRecordNotSaved) || got.CanUndo {
		t.Fatalf("Organize = (%+v, %v), want RECORD_NOT_SAVED sem desfazer", got, err)
	}
	if st.previous != nil {
		t.Fatalf("registro anterior continua: %+v", st.previous)
	}
}

func TestOrganizeErrorsHaveCodes(t *testing.T) {
	svc, _ := newTestService(t)
	_, err := svc.Organize(context.Background(), Request{})
	if apperr.CodeOf(err) != "INVALID_SOURCE" {
		t.Fatalf("CodeOf = %q, want INVALID_SOURCE (#8)", apperr.CodeOf(err))
	}
}

func TestOrganizeUnreadableSource(t *testing.T) {
	svc, _ := newTestService(t)
	src := t.TempDir()
	svc.planner.readDir = func(string) ([]os.DirEntry, error) { return nil, os.ErrPermission }
	if _, err := svc.Organize(context.Background(), Request{SourceFolderPath: src, Options: criteria.Default}); !errors.Is(err, ErrInvalidSource) {
		t.Fatalf("err = %v, want ErrInvalidSource", err)
	}
}

func TestResultJSONHasNoMessage(t *testing.T) {
	data, err := json.Marshal(Result{})
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]any
	_ = json.Unmarshal(data, &fields)
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	want := []string{"canUndo", "canceled", "destinationFolderPath", "failedFiles", "ignoredFolders",
		"ignoredWithoutExtension", "movedFiles", "processedFiles", "sourceFolderPath", "unchangedFiles"}
	if !reflect.DeepEqual(keys, want) {
		t.Fatalf("campos = %v, want %v (#8: sem \"message\")", keys, want)
	}
}

func TestNewServiceDefaultLocation(t *testing.T) {
	svc := NewService(Deps{Metadata: metadata.Reader{}, Store: store.New(filepath.Join(t.TempDir(), "x.json"), nil)})
	if svc.planner == nil || svc.executor == nil {
		t.Fatal("serviço incompleto")
	}
}

func newTestService(t *testing.T) (*Service, *store.FileStore) {
	t.Helper()
	st := store.New(filepath.Join(t.TempDir(), ".sortly", "last-operation.json"), nil)
	return NewService(Deps{Metadata: metadata.Reader{}, Store: st, Location: testLoc}), st
}

// Pastas terminam em "/" (#8).
func assertTree(t *testing.T, root string, want []string) {
	t.Helper()
	got := listTree(t, root)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("árvore =\n%v\nwant\n%v", got, want)
	}
}

func listTree(t *testing.T, root string) []string {
	t.Helper()
	var got []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || path == root {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			if entries, _ := os.ReadDir(path); len(entries) > 0 {
				return nil
			}
			rel += "/"
		}
		got = append(got, rel)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(got)
	return got
}

func TestOrganizeCanceled(t *testing.T) {
	t.Run("no meio: o que foi movido fica no registro", func(t *testing.T) {
		src := t.TempDir()
		writeAt(t, filepath.Join(src, "a.txt"), "a", testTime)
		writeAt(t, filepath.Join(src, "b.txt"), "b", testTime)
		svc, st := newTestService(t)
		ctx, cancel := context.WithCancel(context.Background())
		realMove := svc.executor.move
		svc.executor.move = func(from, to string) (string, error) {
			final, err := realMove(from, to)
			cancel()
			return final, err
		}

		got, err := svc.OrganizeReporting(ctx, Request{SourceFolderPath: src, Options: criteria.Default}, func(Progress) {})

		op, _ := st.Load()
		if !errors.Is(err, context.Canceled) || !got.Canceled || got.MovedFiles != 1 || !got.CanUndo || len(op.MovedItems) != 1 {
			t.Fatalf("Organize = (%+v, %v), registro = %+v", got, err, op)
		}
	})

	t.Run("no planejamento: nada movido", func(t *testing.T) {
		src := t.TempDir()
		writeAt(t, filepath.Join(src, "a.txt"), "a", testTime)
		svc, _ := newTestService(t)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		got, err := svc.Organize(ctx, Request{SourceFolderPath: src, Options: criteria.Default})

		want := Result{SourceFolderPath: src, DestinationFolderPath: src, Canceled: true}
		if !errors.Is(err, context.Canceled) || got != want {
			t.Fatalf("Organize = (%+v, %v), want %+v", got, err, want)
		}
		assertTree(t, src, []string{"a.txt"})
	})

	t.Run("outro erro no planejamento não é cancelamento", func(t *testing.T) {
		svc, _ := newTestService(t)
		svc.planner.readDir = func(string) ([]os.DirEntry, error) { return nil, os.ErrPermission }
		got, err := svc.Organize(context.Background(), Request{SourceFolderPath: t.TempDir(), Options: criteria.Default})
		if !errors.Is(err, ErrInvalidSource) || got.Canceled {
			t.Fatalf("Organize = (%+v, %v)", got, err)
		}
	})
}

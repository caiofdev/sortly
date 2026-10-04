package organizer

// função             | CC | casos
// NewService         |  2 | com e sem fuso
// Service.Organize   |  3 | TestOrganize*: validação falha; origem ilegível; fluxo completo
// Service.record     |  3 | TestOrganize*: nada movido (B2); salvo; falha ao salvar
// Service.canUndo    |  3 | TestOrganize*: movido e salvo; movido sem salvar; nada movido com e sem registro anterior
// validate           |  7 | TestValidate (9 casos)
// isDir              |  1 | via TestValidate
//
// Regressões: B1 (falha parcial ainda grava o journal), B2 (organizar sem mover nada não
// apaga o desfazer anterior), B6 (erros com código), B7 (resultado sem mensagem), B8
// (arquivo já no lugar não é renomeado).

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
	"github.com/caiofdev/sortly/backend/store"
)

func TestValidate(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "arquivo.txt")
	writeAt(t, file, "x", testTime)
	no := false

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
		{"nenhum critério", Request{SourceFolderPath: dir, OrganizationOptions: RawOptions{ByExtension: &no}}, ErrNoCriteria, ""},
		{"destino vazio usa a origem", Request{SourceFolderPath: dir}, nil, dir},
		{"destino inexistente é aceito", Request{SourceFolderPath: dir, DestinationFolderPath: filepath.Join(dir, "novo")}, nil, filepath.Join(dir, "novo")},
		{"byExtension ausente basta", Request{SourceFolderPath: dir, DestinationFolderPath: dir}, nil, dir},
		{"um critério além da extensão desligada", Request{SourceFolderPath: dir, OrganizationOptions: RawOptions{ByDate: true, ByExtension: &no}}, nil, dir},
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

	got, err := svc.Organize(context.Background(), Request{SourceFolderPath: src})
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
	if _, err := svc.Organize(context.Background(), Request{SourceFolderPath: src}); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(st.Path())

	// Segunda vez: só pastas na origem, nada a mover.
	got, err := svc.Organize(context.Background(), Request{SourceFolderPath: src})

	after, _ := os.ReadFile(st.Path())
	if err != nil || got.MovedFiles != 0 || !got.CanUndo || string(before) != string(after) {
		t.Fatalf("Organize = (%+v, %v); registro mudou: %v (B2)", got, err, string(before) != string(after))
	}
}

func TestOrganizeNothingMovedWithoutPreviousRecord(t *testing.T) {
	svc, st := newTestService(t)
	got, err := svc.Organize(context.Background(), Request{SourceFolderPath: t.TempDir()})
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

	got, err := svc.Organize(context.Background(), Request{SourceFolderPath: src})

	op, _ := st.Load()
	if err != nil || got.MovedFiles != 1 || got.FailedFiles != 1 || !got.CanUndo || len(op.MovedItems) != 1 {
		t.Fatalf("Organize = (%+v, %v), registro = %+v (B1)", got, err, op)
	}
}

func TestOrganizeRecordNotSaved(t *testing.T) {
	src := t.TempDir()
	writeAt(t, filepath.Join(src, "a.txt"), "a", testTime)
	blocker := filepath.Join(t.TempDir(), "arquivo")
	writeAt(t, blocker, "", testTime)
	svc := NewService(Deps{Metadata: metadata.Reader{}, Store: store.New(filepath.Join(blocker, "x.json"), nil), Location: testLoc})

	got, err := svc.Organize(context.Background(), Request{SourceFolderPath: src})

	if !errors.Is(err, ErrRecordNotSaved) || apperr.CodeOf(err) != "RECORD_NOT_SAVED" {
		t.Fatalf("err = %v, want RECORD_NOT_SAVED", err)
	}
	if got.MovedFiles != 1 || got.CanUndo {
		t.Fatalf("Result = %+v: o arquivo foi movido, mas não há como desfazer", got)
	}
}

func TestOrganizeErrorsHaveCodes(t *testing.T) {
	svc, _ := newTestService(t)
	_, err := svc.Organize(context.Background(), Request{})
	if apperr.CodeOf(err) != "INVALID_SOURCE" {
		t.Fatalf("CodeOf = %q, want INVALID_SOURCE (B6)", apperr.CodeOf(err))
	}
}

func TestOrganizeUnreadableSource(t *testing.T) {
	svc, _ := newTestService(t)
	src := t.TempDir()
	svc.planner.readDir = func(string) ([]os.DirEntry, error) { return nil, os.ErrPermission }
	if _, err := svc.Organize(context.Background(), Request{SourceFolderPath: src}); !errors.Is(err, ErrInvalidSource) {
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
	want := []string{"canUndo", "destinationFolderPath", "failedFiles", "ignoredFolders",
		"ignoredWithoutExtension", "movedFiles", "processedFiles", "sourceFolderPath", "unchangedFiles"}
	if !reflect.DeepEqual(keys, want) {
		t.Fatalf("campos = %v, want %v (B7: sem \"message\")", keys, want)
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

// assertTree compara a árvore de arquivos (pastas terminam em "/").
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

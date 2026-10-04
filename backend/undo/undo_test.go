package undo

// função                 | CC | casos
// NewService             |  2 | com e sem logger
// Service.Undo           |  2 | TestUndoNothingToUndo; demais testes
// Service.load           |  4 | sem registro; corrompido; lista vazia; erro de leitura; válido
// Service.inversePlan    |  3 | TestUndoReverseOrder; TestUndoSkipsMissing
// countRenamed           |  4 | TestUndoRenamesWhenOccupied; TestUndoRestoresAndCleans (sem renomear)
// remaining              |  6 | TestUndoPartialFailure; TestUndoCanceled; TestUndoSkipsMissing (pulado não fica)
// Service.updateRecord   |  4 | tudo desfeito; sobra item; Clear falha; Save falha
// Service.cleanup        |  4 | com raiz; sem raiz (TestUndoWithoutRoot); raiz = origem (destino vazio)
// removeEmptyAncestors   |  3 | para exatamente na raiz; caixa diferente no Windows; fora da raiz
// cleanupCandidates      |  4 | TestCleanupCandidates: só createdFolders; só pastas dos itens; repetidas
//
// Valor-limite: 0 itens no registro (nada para desfazer) e 1 item; pasta com 0 e 1 arquivo;
// subida que termina exatamente na raiz; registro vazio, corrompido e legado.

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/caiofdev/sortly/backend/apperr"
	"github.com/caiofdev/sortly/backend/fs/files"

	"github.com/caiofdev/sortly/backend/organizer"
	"github.com/caiofdev/sortly/backend/store"
)

func TestUndoNothingToUndo(t *testing.T) {
	cases := map[string]string{
		"sem registro":      "",
		"registro vazio":    " ",
		"registro inválido": "{quebrado",
		"lista com 0 itens": `{"movedItems":[]}`,
		"lista ausente":     `{"sourceFolderPath":"C:\\x"}`,
		"JSON null":         "null",
	}
	for name, content := range cases {
		t.Run(name, func(t *testing.T) {
			st := newStore(t)
			if content != "" {
				writeFile(t, st.Path(), content)
			}
			_, err := NewService(st, nil).Undo(context.Background())
			if !errors.Is(err, ErrNothingToUndo) || apperr.CodeOf(err) != "NOTHING_TO_UNDO" {
				t.Fatalf("err = %v, want NOTHING_TO_UNDO", err)
			}
		})
	}
}

func TestUndoLoadError(t *testing.T) {
	st := newStore(t)
	mkdir(t, st.Path()) // o caminho do registro é uma pasta: erro de leitura
	_, err := NewService(st, nil).Undo(context.Background())
	if err == nil || errors.Is(err, ErrNothingToUndo) || apperr.CodeOf(err) != apperr.CodeUnexpected {
		t.Fatalf("err = %v, want erro inesperado", err)
	}
}

func TestUndoRestoresAndCleans(t *testing.T) {
	src, dst := t.TempDir(), t.TempDir()
	st := withRecord(t, src, dst,
		move(t, src, dst, "a.pdf", "pdf/pages-1"),
		move(t, src, dst, "b.pdf", "pdf/pages-3"),
	)

	got, err := NewService(st, nil).Undo(context.Background())

	if err != nil || got != (Result{RestoredFiles: 2}) {
		t.Fatalf("Undo = (%+v, %v)", got, err)
	}
	assertTree(t, src, "a.pdf", "b.pdf")
	assertTree(t, dst) // pastas criadas removidas; a raiz do destino fica (mesmo vazia)
	assertNoRecord(t, st)
}

func TestUndoSkipsMissing(t *testing.T) {
	src, dst := t.TempDir(), t.TempDir()
	a := move(t, src, dst, "a.txt", "txt")
	b := move(t, src, dst, "b.txt", "txt")
	if err := os.Remove(b.To); err != nil { // usuário apagou um dos arquivos organizados
		t.Fatal(err)
	}
	st := withRecord(t, src, dst, a, b)

	got, err := NewService(st, nil).Undo(context.Background())

	if err != nil || got != (Result{RestoredFiles: 1, SkippedMissing: 1}) {
		t.Fatalf("Undo = (%+v, %v)", got, err)
	}
	assertNoRecord(t, st) // o arquivo pulado não fica pendente
}

func TestUndoRenamesWhenOccupied(t *testing.T) {
	src, dst := t.TempDir(), t.TempDir()
	a := move(t, src, dst, "a.txt", "txt")
	writeFile(t, a.From, "novo arquivo com o mesmo nome")
	st := withRecord(t, src, dst, a)

	got, err := NewService(st, nil).Undo(context.Background())

	if err != nil || got != (Result{RestoredFiles: 1, RenamedOnRestore: 1}) {
		t.Fatalf("Undo = (%+v, %v)", got, err)
	}
	assertContent(t, a.From, "novo arquivo com o mesmo nome")
	assertContent(t, filepath.Join(src, "a (1).txt"), "a.txt")
}

func TestUndoPreservesNonEmptyFolder(t *testing.T) {
	src, dst := t.TempDir(), t.TempDir()
	a := move(t, src, dst, "a.txt", "txt/date-2026-03-05")
	writeFile(t, filepath.Join(dst, "txt", "do-usuario.txt"), "x") // 1 arquivo do usuário
	st := withRecord(t, src, dst, a)

	if _, err := NewService(st, nil).Undo(context.Background()); err != nil {
		t.Fatal(err)
	}
	assertTree(t, dst, "txt/do-usuario.txt") // date-* (vazia) removida; txt (1 arquivo) preservada
}

func TestUndoLegacyRecordWithoutCreatedFolders(t *testing.T) {
	src, dst := t.TempDir(), t.TempDir()
	a := move(t, src, dst, "a.txt", "txt/size-1mb")
	st := newStore(t)
	saveRecord(t, st, store.Operation{SourceFolderPath: src, DestinationFolderPath: dst, MovedItems: []store.MovedItem{a}})

	if _, err := NewService(st, nil).Undo(context.Background()); err != nil {
		t.Fatal(err)
	}
	assertTree(t, dst)
}

func TestUndoDestinationEmptyUsesSourceAsRoot(t *testing.T) {
	src := t.TempDir()
	a := move(t, src, src, "a.txt", "txt")
	st := newStore(t)
	saveRecord(t, st, store.Operation{SourceFolderPath: src, MovedItems: []store.MovedItem{a}})

	if _, err := NewService(st, nil).Undo(context.Background()); err != nil {
		t.Fatal(err)
	}
	assertTree(t, src, "a.txt")
}

func TestUndoWithoutRoot(t *testing.T) {
	src, dst := t.TempDir(), t.TempDir()
	a := move(t, src, dst, "a.txt", "txt/sub")
	st := newStore(t)
	saveRecord(t, st, store.Operation{MovedItems: []store.MovedItem{a}})
	svc := NewService(st, nil)
	var removed []string
	svc.removeDir = func(dir string) bool { removed = append(removed, dir); return true }

	if _, err := svc.Undo(context.Background()); err != nil {
		t.Fatal(err)
	}
	// Sem raiz conhecida, só a pasta do arquivo é tentada, sem subir.
	if want := []string{filepath.Join(dst, "txt", "sub")}; !reflect.DeepEqual(removed, want) {
		t.Fatalf("removidas = %v, want %v", removed, want)
	}
}

func TestRemoveEmptyAncestors(t *testing.T) {
	root := filepath.Join(t.TempDir(), "Destino")
	deep := filepath.Join(root, "pdf", "pages-3")
	outside := filepath.Join(filepath.Dir(root), "outra")
	svc := NewService(newStore(t), nil)
	var removed []string
	svc.removeDir = func(dir string) bool { removed = append(removed, dir); return true }

	svc.removeEmptyAncestors(deep, root)
	svc.removeEmptyAncestors(outside, root)

	want := []string{deep, filepath.Join(root, "pdf")} // para exatamente na raiz; fora da raiz, nada
	if !reflect.DeepEqual(removed, want) {
		t.Fatalf("removidas = %v, want %v", removed, want)
	}
}

func TestRemoveEmptyAncestorsCaseInsensitiveRoot(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("caixa diferente só é o mesmo caminho no Windows")
	}
	root := filepath.Join(t.TempDir(), "Destino")
	mkdir(t, filepath.Join(root, "pdf"))

	NewService(newStore(t), nil).removeEmptyAncestors(filepath.Join(root, "pdf"), strings.ToUpper(root))

	if !files.Exists(root) || files.Exists(filepath.Join(root, "pdf")) {
		t.Fatal("a raiz não pode ser removida por ter caixa diferente; a subpasta sim")
	}
}

func TestUndoReverseOrder(t *testing.T) {
	src, dst := t.TempDir(), t.TempDir()
	items := []store.MovedItem{
		move(t, src, dst, "a.txt", "txt"), move(t, src, dst, "b.txt", "txt"), move(t, src, dst, "c.txt", "txt"),
	}
	st := withRecord(t, src, dst, items...)
	svc := NewService(st, nil)
	fake := &fakeMover{}
	svc.mover = fake

	if _, err := svc.Undo(context.Background()); err != nil {
		t.Fatal(err)
	}
	var order []string
	for _, m := range fake.plan.Moves {
		order = append(order, filepath.Base(m.To))
	}
	if want := []string{"c.txt", "b.txt", "a.txt"}; !reflect.DeepEqual(order, want) {
		t.Fatalf("ordem = %v, want %v", order, want)
	}
}

func TestUndoPartialFailure(t *testing.T) {
	src, dst := t.TempDir(), t.TempDir()
	a, b := move(t, src, dst, "a.txt", "txt"), move(t, src, dst, "b.txt", "txt")
	st := withRecord(t, src, dst, a, b)
	svc := NewService(st, nil)
	svc.mover = &fakeMover{failOn: "b.txt"}

	got, err := svc.Undo(context.Background())
	if err != nil || got != (Result{RestoredFiles: 1, FailedFiles: 1, CanUndo: true}) {
		t.Fatalf("1º Undo = (%+v, %v)", got, err)
	}
	op, _ := st.Load()
	if len(op.MovedItems) != 1 || op.MovedItems[0] != b {
		t.Fatalf("registro deveria guardar só o item que falhou: %+v", op)
	}

	// Um novo desfazer, sem falha, termina o trabalho.
	got, err = NewService(st, nil).Undo(context.Background())
	if err != nil || got != (Result{RestoredFiles: 1}) {
		t.Fatalf("2º Undo = (%+v, %v)", got, err)
	}
	assertTree(t, src, "a.txt", "b.txt")
	assertNoRecord(t, st)
}

func TestUndoCanceled(t *testing.T) {
	src, dst := t.TempDir(), t.TempDir()
	a, b := move(t, src, dst, "a.txt", "txt"), move(t, src, dst, "b.txt", "txt")
	st := withRecord(t, src, dst, a, b)
	svc := NewService(st, nil)
	svc.mover = &fakeMover{cancelAfter: 1}

	got, err := svc.Undo(context.Background())

	if !errors.Is(err, context.Canceled) || got.RestoredFiles != 1 || !got.CanUndo {
		t.Fatalf("Undo = (%+v, %v)", got, err)
	}
	op, _ := st.Load()
	if len(op.MovedItems) != 1 || op.MovedItems[0] != a { // ordem inversa: b voltou, a ficou pendente
		t.Fatalf("registro = %+v, want só o item não tentado", op)
	}
}

func TestUpdateRecordFailuresAreLogged(t *testing.T) {
	var logs bytes.Buffer
	log := slog.New(slog.NewTextHandler(&logs, nil))
	op := &store.Operation{MovedItems: []store.MovedItem{{From: "a", To: "b"}}}
	svc := NewService(failingStore{}, log)

	if svc.updateRecord(op, nil) {
		t.Error("sem pendências, CanUndo deveria ser falso")
	}
	if !svc.updateRecord(op, op.MovedItems) {
		t.Error("com pendências, CanUndo deveria ser verdadeiro")
	}
	for _, want := range []string{"apagar o registro", "regravar o registro"} {
		if !strings.Contains(logs.String(), want) {
			t.Errorf("log não contém %q: %s", want, logs.String())
		}
	}
}

func TestCleanupCandidates(t *testing.T) {
	op := &store.Operation{
		CreatedFolders: []string{filepath.Join("d", "pdf"), filepath.Join("d", "png")},
		MovedItems: []store.MovedItem{
			{To: filepath.Join("d", "pdf", "a.pdf")},
			{To: filepath.Join("d", "txt", "b.txt")},
		},
	}
	want := []string{filepath.Join("d", "pdf"), filepath.Join("d", "png"), filepath.Join("d", "txt")}
	if got := cleanupCandidates(op); !reflect.DeepEqual(got, want) {
		t.Fatalf("cleanupCandidates = %v, want %v", got, want)
	}
}

// --- apoio ---

// fakeMover move de verdade, mas pode falhar num arquivo ou cancelar depois de N movimentos.
type fakeMover struct {
	plan        organizer.Plan
	failOn      string
	cancelAfter int
}

func (f *fakeMover) Apply(_ context.Context, plan organizer.Plan) (organizer.Outcome, error) {
	f.plan = plan
	var out organizer.Outcome
	for i, m := range plan.Moves {
		if f.cancelAfter > 0 && i == f.cancelAfter {
			return out, context.Canceled
		}
		if f.failOn != "" && strings.HasSuffix(m.From, f.failOn) {
			out.FailedFiles++
			continue
		}
		final, err := files.MoveUnique(m.From, m.To)
		if err != nil {
			return out, err
		}
		out.MovedItems = append(out.MovedItems, store.MovedItem{From: m.From, To: final})
	}
	return out, nil
}

type failingStore struct{}

func (failingStore) Load() (*store.Operation, error) { return nil, nil }
func (failingStore) Save(store.Operation) error      { return errors.New("disco cheio") }
func (failingStore) Clear() error                    { return errors.New("sem permissão") }

func newStore(t *testing.T) *store.FileStore {
	t.Helper()
	return store.New(filepath.Join(t.TempDir(), "last-operation.json"), nil)
}

// move cria src/name e o move para dst/sub/name, como a organização faria.
func move(t *testing.T, src, dst, name, sub string) store.MovedItem {
	t.Helper()
	from := filepath.Join(src, name)
	to := filepath.Join(dst, filepath.FromSlash(sub), name)
	writeFile(t, from, name)
	mkdir(t, filepath.Dir(to))
	if err := os.Rename(from, to); err != nil {
		t.Fatal(err)
	}
	return store.MovedItem{From: from, To: to}
}

func withRecord(t *testing.T, src, dst string, items ...store.MovedItem) *store.FileStore {
	t.Helper()
	st := newStore(t)
	folders := map[string]bool{}
	var created []string
	for _, it := range items {
		if dir := filepath.Dir(it.To); !folders[dir] {
			folders[dir] = true
			created = append(created, dir)
		}
	}
	saveRecord(t, st, store.Operation{SourceFolderPath: src, DestinationFolderPath: dst, MovedItems: items, CreatedFolders: created})
	return st
}

func saveRecord(t *testing.T, st *store.FileStore, op store.Operation) {
	t.Helper()
	if err := st.Save(op); err != nil {
		t.Fatal(err)
	}
}

func assertNoRecord(t *testing.T, st *store.FileStore) {
	t.Helper()
	if op, err := st.Load(); op != nil || err != nil {
		t.Fatalf("o registro deveria ter sido apagado: (%+v, %v)", op, err)
	}
}

func assertContent(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil || string(got) != want {
		t.Fatalf("%s = (%q, %v), want %q", path, got, err, want)
	}
}

// assertTree compara os arquivos sob root (caminhos relativos com "/").
func assertTree(t *testing.T, root string, want ...string) {
	t.Helper()
	var got []string
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			rel, _ := filepath.Rel(root, path)
			got = append(got, filepath.ToSlash(rel))
		}
		if err == nil && d.IsDir() && path != root {
			if entries, _ := os.ReadDir(path); len(entries) == 0 {
				rel, _ := filepath.Rel(root, path)
				got = append(got, filepath.ToSlash(rel)+"/")
			}
		}
		return err
	})
	if len(got) == 0 && len(want) == 0 {
		return
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("árvore de %s =\n%v\nwant\n%v", root, got, want)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	mkdir(t, filepath.Dir(path))
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func mkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

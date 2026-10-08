package undo

import (
	"bytes"
	"context"
	"encoding/json"
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
	mkdir(t, st.Path()) // o caminho do registro é uma pasta: erro de leitura (#9)
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
	assertTree(t, dst) // pastas criadas removidas; a raiz do destino fica, mesmo vazia (#9)
	assertNoRecord(t, st)
}

func TestUndoSkipsMissing(t *testing.T) {
	src, dst := t.TempDir(), t.TempDir()
	a := move(t, src, dst, "a.txt", "txt")
	b := move(t, src, dst, "b.txt", "txt")
	if err := os.Remove(b.To); err != nil { // usuário apagou um dos arquivos organizados (#9)
		t.Fatal(err)
	}
	st := withRecord(t, src, dst, a, b)

	got, err := NewService(st, nil).Undo(context.Background())

	if err != nil || got != (Result{RestoredFiles: 1, SkippedMissing: 1}) {
		t.Fatalf("Undo = (%+v, %v)", got, err)
	}
	assertNoRecord(t, st) // o arquivo pulado não fica pendente (#9)
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
	writeFile(t, filepath.Join(dst, "txt", "do-usuario.txt"), "x") // 1 arquivo do usuário (#9)
	st := withRecord(t, src, dst, a)

	if _, err := NewService(st, nil).Undo(context.Background()); err != nil {
		t.Fatal(err)
	}
	assertTree(t, dst, "txt/do-usuario.txt") // date-* (vazia) removida; txt (1 arquivo) preservada (#9)
}

// Regressão (#57): pasta que já existia no destino não é removida pelo
// desfazer, mesmo vazia; as criadas pela organização são.
func TestUndoKeepsFoldersThatAlreadyExisted(t *testing.T) {
	src, dst := t.TempDir(), t.TempDir()
	mkdir(t, filepath.Join(dst, "pdf")) // vazia, do usuário (#57)
	a := move(t, src, dst, "a.pdf", "pdf/pages-1")
	st := newStore(t)
	saveRecord(t, st, store.Operation{SourceFolderPath: src, DestinationFolderPath: dst, MovedItems: []store.MovedItem{a},
		CreatedFolders: []string{filepath.Join(dst, "pdf", "pages-1")}})

	if _, err := NewService(st, nil).Undo(context.Background()); err != nil {
		t.Fatal(err)
	}
	assertTree(t, dst, "pdf/")
}

func TestUndoCreatedFolderOutsideRootIsKept(t *testing.T) {
	src, dst := t.TempDir(), t.TempDir()
	outside := filepath.Join(t.TempDir(), "vazia")
	mkdir(t, outside)
	a := move(t, src, dst, "a.txt", "txt")
	st := newStore(t)
	saveRecord(t, st, store.Operation{SourceFolderPath: src, DestinationFolderPath: dst, MovedItems: []store.MovedItem{a},
		CreatedFolders: []string{filepath.Join(dst, "txt"), outside, dst}})

	if _, err := NewService(st, nil).Undo(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !files.Exists(outside) || !files.Exists(dst) || files.Exists(filepath.Join(dst, "txt")) {
		t.Fatal("só a pasta dentro da raiz pode ser removida; a raiz e o que está fora ficam")
	}
}

func TestUndoLegacyRecordWithoutCreatedFolders(t *testing.T) {
	src, dst := t.TempDir(), t.TempDir()
	a := move(t, src, dst, "a.txt", "txt/size-1mb")
	st := newStore(t)
	// Sem o campo createdFolders, como nos registros antigos: o store sempre grava
	// a lista, então o JSON é escrito à mão (#57).
	legacy, err := json.Marshal(map[string]any{
		"sourceFolderPath": src, "destinationFolderPath": dst,
		"movedItems": []map[string]string{{"from": a.From, "to": a.To}},
	})
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, st.Path(), string(legacy))

	if _, err := NewService(st, nil).Undo(context.Background()); err != nil {
		t.Fatal(err)
	}
	assertTree(t, dst)
}

func TestUndoDestinationEmptyUsesSourceAsRoot(t *testing.T) {
	src := t.TempDir()
	a := move(t, src, src, "a.txt", "txt")
	st := newStore(t)
	saveRecord(t, st, store.Operation{SourceFolderPath: src, MovedItems: []store.MovedItem{a},
		CreatedFolders: []string{filepath.Join(src, "txt")}})

	if _, err := NewService(st, nil).Undo(context.Background()); err != nil {
		t.Fatal(err)
	}
	assertTree(t, src, "a.txt")
}

func TestUndoWithoutRootTouchesNothing(t *testing.T) {
	src, dst := t.TempDir(), t.TempDir()
	a := move(t, src, dst, "a.txt", "txt/sub")
	st := newStore(t)
	saveRecord(t, st, store.Operation{MovedItems: []store.MovedItem{a}})
	svc := NewService(st, nil)
	var removed []string
	svc.removeDir = func(dir string) bool { removed = append(removed, dir); return true }

	got, err := svc.Undo(context.Background())

	if err != nil || got != (Result{SkippedMissing: 1}) || len(removed) != 0 {
		t.Fatalf("Undo = (%+v, %v), removidas = %v", got, err, removed)
	}
	assertTree(t, dst, "txt/sub/a.txt")
}

// Regressão (#54): um registro editado ou corrompido não pode mover arquivos
// para fora das pastas registradas.
func TestUndoIgnoresItemsOutsideRecordedFolders(t *testing.T) {
	tests := []struct {
		name      string
		from, to  func(src, dst, outside string) string
		wantMoved int
	}{
		{"dentro das pastas volta",
			func(src, _, _ string) string { return filepath.Join(src, "a.txt") },
			func(_, dst, _ string) string { return filepath.Join(dst, "txt", "a.txt") }, 1},
		{"from fora da origem",
			func(_, _, out string) string { return filepath.Join(out, "a.txt") },
			func(_, dst, _ string) string { return filepath.Join(dst, "txt", "a.txt") }, 0},
		{"to fora do destino",
			func(src, _, _ string) string { return filepath.Join(src, "a.txt") },
			func(_, _, out string) string { return filepath.Join(out, "a.txt") }, 0},
		{"from em pasta irmã com o mesmo prefixo",
			func(src, _, _ string) string { return filepath.Join(src+"2", "a.txt") },
			func(_, dst, _ string) string { return filepath.Join(dst, "txt", "a.txt") }, 0},
		{"from é a própria origem",
			func(src, _, _ string) string { return src },
			func(_, dst, _ string) string { return filepath.Join(dst, "txt", "a.txt") }, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := t.TempDir()
			src, dst, outside := filepath.Join(base, "dados"), filepath.Join(base, "destino"), filepath.Join(base, "fora")
			item := store.MovedItem{From: tt.from(src, dst, outside), To: tt.to(src, dst, outside)}
			// Todas as pastas existem: sem a validação, o movimento daria certo (#54).
			for _, dir := range []string{src, src + "2", outside, filepath.Dir(item.To)} {
				mkdir(t, dir)
			}
			writeFile(t, item.To, "conteúdo")
			st := newStore(t)
			saveRecord(t, st, store.Operation{SourceFolderPath: src, DestinationFolderPath: dst, MovedItems: []store.MovedItem{item}})
			var logs bytes.Buffer

			got, err := NewService(st, slog.New(slog.NewTextHandler(&logs, nil))).Undo(context.Background())

			if err != nil || got.RestoredFiles != tt.wantMoved || got.SkippedMissing != 1-tt.wantMoved {
				t.Fatalf("Undo = (%+v, %v)", got, err)
			}
			if tt.wantMoved == 0 {
				assertContent(t, item.To, "conteúdo")
				if !strings.Contains(logs.String(), "fora das pastas registradas") {
					t.Errorf("o item ignorado deveria ir para o log: %s", logs.String())
				}
			}
		})
	}
}

func TestStrictlyInside(t *testing.T) {
	root := filepath.Join(t.TempDir(), "dados")
	cases := map[string]bool{
		filepath.Join(root, "a.txt"): true,
		root:                         false,
		root + "2":                   false,
		filepath.Dir(root):           false,
	}
	for path, want := range cases {
		if got := strictlyInside(path, root); got != want {
			t.Errorf("strictlyInside(%q) = %v, want %v", path, got, want)
		}
	}
}

func TestRootOf(t *testing.T) {
	if got := rootOf(&store.Operation{SourceFolderPath: "o", DestinationFolderPath: "d"}); got != "d" {
		t.Errorf("com destino = %q, want d", got)
	}
	if got := rootOf(&store.Operation{SourceFolderPath: "o"}); got != "o" {
		t.Errorf("sem destino = %q, want o", got)
	}
}

func TestWithinRecordedFolders(t *testing.T) {
	src, dst := filepath.Join(t.TempDir(), "dados"), filepath.Join(t.TempDir(), "destino")
	in := store.MovedItem{From: filepath.Join(src, "a.txt"), To: filepath.Join(dst, "txt", "a.txt")}
	tests := []struct {
		name string
		op   store.Operation
		item store.MovedItem
		want bool
	}{
		{"dentro", store.Operation{SourceFolderPath: src, DestinationFolderPath: dst}, in, true},
		{"sem origem no registro", store.Operation{DestinationFolderPath: dst}, in, false},
		{"from fora", store.Operation{SourceFolderPath: src, DestinationFolderPath: dst},
			store.MovedItem{From: filepath.Join(dst, "a.txt"), To: in.To}, false},
		{"to fora", store.Operation{SourceFolderPath: src, DestinationFolderPath: dst},
			store.MovedItem{From: in.From, To: filepath.Join(src, "a.txt")}, false},
		{"to é o próprio destino", store.Operation{SourceFolderPath: src, DestinationFolderPath: dst},
			store.MovedItem{From: in.From, To: dst}, false},
		{"destino vazio usa a origem", store.Operation{SourceFolderPath: src},
			store.MovedItem{From: in.From, To: filepath.Join(src, "txt", "a.txt")}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := withinRecordedFolders(&tt.op, tt.item); got != tt.want {
				t.Fatalf("withinRecordedFolders = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemoveEmptyAncestors(t *testing.T) {
	root := filepath.Join(t.TempDir(), "Destino")
	tests := []struct {
		name  string
		start string
		want  []string
	}{
		{"fora da raiz: nada", filepath.Join(filepath.Dir(root), "outra"), nil},
		{"a própria raiz: nada", root, nil},
		{"um nível: só a pasta", filepath.Join(root, "pdf"), []string{filepath.Join(root, "pdf")}},
		{"vários níveis: sobe até a raiz, exclusive", filepath.Join(root, "pdf", "pages-3"),
			[]string{filepath.Join(root, "pdf", "pages-3"), filepath.Join(root, "pdf")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(newStore(t), nil)
			var removed []string
			svc.removeDir = func(dir string) bool { removed = append(removed, dir); return true }

			svc.removeEmptyAncestors(tt.start, root)

			if !reflect.DeepEqual(removed, tt.want) {
				t.Fatalf("removidas = %v, want %v", removed, tt.want)
			}
		})
	}
}

func TestRemoveEmptyAncestorsCaseInsensitiveRoot(t *testing.T) {
	if !caseInsensitivePlatform() {
		t.Skip("caixa diferente só é o mesmo caminho no Windows e no macOS")
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

	// Um novo desfazer, sem falha, termina o trabalho (#9).
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
	if len(op.MovedItems) != 1 || op.MovedItems[0] != a { // ordem inversa: b voltou, a ficou pendente (#9)
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

func TestItemFolders(t *testing.T) {
	pdf, txt := filepath.Join("d", "pdf"), filepath.Join("d", "txt")
	tests := []struct {
		name string
		op   store.Operation
		want []string
	}{
		{"sem arquivos", store.Operation{}, nil},
		{"uma pasta por arquivo", store.Operation{MovedItems: []store.MovedItem{
			{To: filepath.Join(pdf, "a.pdf")}, {To: filepath.Join(txt, "b.txt")},
		}}, []string{pdf, txt}},
		{"pasta repetida entra uma vez, na ordem", store.Operation{MovedItems: []store.MovedItem{
			{To: filepath.Join(txt, "a.txt")}, {To: filepath.Join(pdf, "b.pdf")}, {To: filepath.Join(txt, "c.txt")},
		}}, []string{txt, pdf}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := itemFolders(&tt.op); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("itemFolders = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeepestFirst(t *testing.T) {
	d := filepath.Join("d")
	pdf, pages := filepath.Join(d, "pdf"), filepath.Join(d, "pdf", "pages-3")
	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{"vazia", nil, nil},
		{"mãe antes da filha", []string{d, pdf, pages}, []string{pages, pdf, d}},
		{"já na ordem", []string{pages, pdf}, []string{pages, pdf}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deepestFirst(tt.in)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("deepestFirst = %v, want %v", got, tt.want)
			}
		})
	}
}

// --- apoio ---

// Move de verdade, mas pode falhar num arquivo ou cancelar depois de N movimentos (#9).
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

// Cria src/name e o move para dst/sub/name, como a organização faria (#9).
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

// Como a organização gravaria quando todas as pastas abaixo de dst foram criadas por ela
// (#57).
func withRecord(t *testing.T, src, dst string, items ...store.MovedItem) *store.FileStore {
	t.Helper()
	st := newStore(t)
	folders := map[string]bool{}
	var created []string
	for _, it := range items {
		for dir := filepath.Dir(it.To); dir != dst && !folders[dir]; dir = filepath.Dir(dir) {
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

// Caminhos relativos com "/" (#9).
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

func caseInsensitivePlatform() bool { return runtime.GOOS == "windows" || runtime.GOOS == "darwin" }

package tests

// Ponta a ponta com "Incluir subpastas": os arquivos de todos os níveis vão para
// as pastas de critério, as subpastas que ficaram vazias somem, e o desfazer as
// recria ao devolver os arquivos (#83).

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/caiofdev/sortly/backend/metadata"
	"github.com/caiofdev/sortly/backend/organizer"
	"github.com/caiofdev/sortly/backend/organizer/criteria"
	"github.com/caiofdev/sortly/backend/store"
	"github.com/caiofdev/sortly/backend/undo"
)

func fileList(t *testing.T, root string) []string {
	t.Helper()
	var got []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			rel, _ := filepath.Rel(root, path)
			got = append(got, filepath.ToSlash(rel))
		}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(got)
	return got
}

func TestSubfoldersThenUndoRestoresTree(t *testing.T) {
	src, dst := t.TempDir(), t.TempDir()
	for _, name := range []string{"a.pdf", "fotos/b.png", "fotos/2026/c.png", "docs/d.txt"} {
		write(t, filepath.Join(src, filepath.FromSlash(name)), name)
	}
	before := fileList(t, src)
	st := store.New(filepath.Join(t.TempDir(), "last-operation.json"), nil)

	res, err := organizer.NewService(organizer.Deps{Metadata: metadata.Reader{}, Store: st}).Organize(context.Background(),
		organizer.Request{SourceFolderPath: src, DestinationFolderPath: dst, Options: criteria.Default, IncludeSubfolders: true})
	if err != nil || res.MovedFiles != 4 {
		t.Fatalf("Organize = (%+v, %v)", res, err)
	}
	if entries, _ := os.ReadDir(src); len(entries) != 0 {
		t.Fatalf("as subpastas vazias deveriam ter sido removidas: %v", entries)
	}

	got, err := undo.NewService(st, nil).Undo(context.Background())

	if err != nil || got.RestoredFiles != 4 {
		t.Fatalf("Undo = (%+v, %v)", got, err)
	}
	if after := fileList(t, src); len(after) != len(before) {
		t.Fatalf("árvore depois do desfazer = %v, want %v", after, before)
	}
	for i := range before {
		if fileList(t, src)[i] != before[i] {
			t.Fatalf("árvore depois do desfazer = %v, want %v", fileList(t, src), before)
		}
	}
}

package tests

// Teste ponta a ponta: organizar com todos os critérios e desfazer devolve a
// árvore idêntica à original (mesmos caminhos, conteúdos e datas de modificação),
// remove todas as pastas criadas e apaga o registro.

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/caiofdev/sortly/backend/metadata"
	"github.com/caiofdev/sortly/backend/organizer"
	"github.com/caiofdev/sortly/backend/organizer/criteria"
	"github.com/caiofdev/sortly/backend/store"
	"github.com/caiofdev/sortly/backend/undo"
)

func TestOrganizeThenUndoRestoresIdenticalTree(t *testing.T) {
	for _, tc := range []struct {
		name    string
		inPlace bool
	}{
		{"destino separado", false},
		{"na própria pasta", true},
	} {
		t.Run(tc.name, func(t *testing.T) { roundTrip(t, tc.inPlace) })
	}
}

func roundTrip(t *testing.T, inPlace bool) {
	t.Helper()
	src := t.TempDir()
	copyFixtures(t, src)
	dst := filepath.Join(t.TempDir(), "Organizados")
	if inPlace {
		dst = src
	}
	before := treeHash(t, src)
	st := store.New(filepath.Join(t.TempDir(), "last-operation.json"), nil)
	yes := true
	req := organizer.Request{
		SourceFolderPath:      src,
		DestinationFolderPath: dst,
		OrganizationOptions: criteria.RawOptions{
			ByExtension: &yes, ByDate: true, BySize: true, ByResolution: true, ByDuration: true, ByPages: true,
		},
	}

	org, err := organizer.NewService(organizer.Deps{Metadata: metadata.Reader{}, Store: st}).Organize(context.Background(), req)
	if err != nil || org.MovedFiles == 0 {
		t.Fatalf("Organize = (%+v, %v)", org, err)
	}
	if treeHash(t, src) == before {
		t.Fatal("a organização deveria ter mudado a árvore")
	}

	got, err := undo.NewService(st, nil).Undo(context.Background())
	if err != nil || got.RestoredFiles != org.MovedFiles || got.CanUndo {
		t.Fatalf("Undo = (%+v, %v), want %d restaurados", got, err, org.MovedFiles)
	}
	if after := treeHash(t, src); after != before {
		t.Fatalf("árvore depois do desfazer difere da original\nantes:  %s\ndepois: %s", before, after)
	}
	if !inPlace {
		assertEmptyTree(t, dst)
	}
	assertNoRecord(t, st)
}

// copyFixtures copia as fixtures de metadados (imagens, mp4, documentos) para
// dir, mais arquivos sem extensão e uma subpasta, todos com a mesma data.
func copyFixtures(t *testing.T, dir string) {
	t.Helper()
	root := filepath.Join("..", "metadata", "testdata")
	for _, group := range []string{"image", "mp4", "pages"} {
		entries, err := os.ReadDir(filepath.Join(root, group))
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range entries {
			copyFile(t, filepath.Join(root, group, e.Name()), filepath.Join(dir, e.Name()))
		}
	}
	for _, name := range []string{"LEIAME", ".gitignore", "nota.txt"} {
		writeFile(t, filepath.Join(dir, name), name)
	}
	writeFile(t, filepath.Join(dir, "subpasta", "dentro.txt"), "não deve ser tocado")
	mtime := time.Date(2026, 3, 5, 12, 0, 0, 0, time.Local)
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if !e.IsDir() {
			_ = os.Chtimes(filepath.Join(dir, e.Name()), mtime, mtime)
		}
	}
}

func copyFile(t *testing.T, from, to string) {
	t.Helper()
	data, err := os.ReadFile(from)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, to, string(data))
}

// treeHash resume a árvore: caminho relativo, hash do conteúdo e data de modificação.
func treeHash(t *testing.T, root string) string {
	t.Helper()
	var lines []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()
		h := sha256.New()
		if _, err := io.Copy(h, f); err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		lines = append(lines, fmt.Sprintf("%s %x %d", filepath.ToSlash(rel), h.Sum(nil), info.ModTime().UnixNano()))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(lines)
	sum := sha256.Sum256([]byte(fmt.Sprint(lines)))
	return fmt.Sprintf("%d arquivos, %s", len(lines), hex.EncodeToString(sum[:8]))
}

func assertNoRecord(t *testing.T, st *store.FileStore) {
	t.Helper()
	if op, err := st.Load(); op != nil || err != nil {
		t.Fatalf("o registro deveria ter sido apagado: (%+v, %v)", op, err)
	}
}

// assertEmptyTree falha se sobrou algum arquivo ou pasta dentro de root.
func assertEmptyTree(t *testing.T, root string) {
	t.Helper()
	entries, err := os.ReadDir(root)
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	for _, e := range entries {
		t.Errorf("sobrou %s em %s", e.Name(), root)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

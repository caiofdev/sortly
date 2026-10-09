package organizer

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/caiofdev/sortly/backend/organizer/criteria"
	"github.com/caiofdev/sortly/backend/store"
)

// Origem com arquivos em três níveis e uma subpasta só com outra subpasta (#83).
func treeSource(t *testing.T) string {
	t.Helper()
	src := t.TempDir()
	writeAt(t, filepath.Join(src, "a.pdf"), "a", testTime)
	writeAt(t, filepath.Join(src, "fotos", "b.png"), "b", testTime)
	writeAt(t, filepath.Join(src, "fotos", "2026", "c.png"), "c", testTime)
	writeAt(t, filepath.Join(src, "docs", "d.txt"), "d", testTime)
	return src
}

func TestPlanTree(t *testing.T) {
	src := treeSource(t)
	dst := t.TempDir()

	got, err := newTestPlanner().PlanTree(context.Background(), src, dst, criteria.Options{ByExtension: true})

	if err != nil || got.ProcessedFiles != 4 || len(got.Moves) != 4 || got.IgnoredFolders != 0 {
		t.Fatalf("PlanTree = (%+v, %v)", got, err)
	}
	want := map[string]string{
		"a.pdf": filepath.Join(dst, "pdf", "a.pdf"),
		"b.png": filepath.Join(dst, "png", "b.png"),
		"c.png": filepath.Join(dst, "png", "c.png"),
		"d.txt": filepath.Join(dst, "txt", "d.txt"),
	}
	for _, m := range got.Moves {
		if want[filepath.Base(m.From)] != m.To {
			t.Errorf("%s → %s, want %s", m.From, m.To, want[filepath.Base(m.From)])
		}
	}
}

func TestPlanTreeSkipsDestinationInsideSource(t *testing.T) {
	src := treeSource(t)
	dst := filepath.Join(src, "Organizados")
	writeAt(t, filepath.Join(dst, "pdf", "velho.pdf"), "v", testTime)

	got, err := newTestPlanner().PlanTree(context.Background(), src, dst, criteria.Options{ByExtension: true})

	if err != nil || got.ProcessedFiles != 4 || got.IgnoredFolders != 1 {
		t.Fatalf("o destino dentro da origem não é percorrido: (%+v, %v)", got, err)
	}
}

func TestPlanTreeSpecialEntries(t *testing.T) {
	src := t.TempDir()
	p := newTestPlanner()
	p.readDir = func(dir string) ([]fs.DirEntry, error) {
		if dir == src {
			return []fs.DirEntry{
				fakeEntry{name: "atalho", mode: fs.ModeSymlink},
				fakeEntry{name: "trancada", mode: fs.ModeDir},
			}, nil
		}
		return nil, os.ErrPermission
	}

	got, err := p.PlanTree(context.Background(), src, src, criteria.Options{ByExtension: true})

	// O atalho não é seguido nem contado; a subpasta ilegível conta como ignorada (#83).
	if err != nil || got.ProcessedFiles != 0 || got.IgnoredFolders != 1 {
		t.Fatalf("PlanTree = (%+v, %v)", got, err)
	}
}

func TestPlanTreeErrors(t *testing.T) {
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	tests := []struct {
		name string
		ctx  context.Context
		src  func(t *testing.T) string
		want error
	}{
		{"origem ilegível", context.Background(), func(t *testing.T) string { return filepath.Join(t.TempDir(), "nada") }, ErrInvalidSource},
		{"cancelado", canceled, treeSource, context.Canceled},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newTestPlanner().PlanTree(tt.ctx, tt.src(t), t.TempDir(), criteria.Options{ByExtension: true})
			if !errors.Is(err, tt.want) {
				t.Fatalf("err = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestOrganizeIncludingSubfolders(t *testing.T) {
	tests := []struct {
		name     string
		inPlace  bool
		wantSrc  []string
		wantDest []string
	}{
		{"destino separado: as subpastas vazias somem", false,
			[]string{"guardar/LEIAME"},
			[]string{"pdf/a.pdf", "png/b.png", "png/c.png", "txt/d.txt"}},
		{"na própria pasta", true,
			[]string{"guardar/LEIAME", "pdf/a.pdf", "png/b.png", "png/c.png", "txt/d.txt"}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := treeSource(t)
			writeAt(t, filepath.Join(src, "guardar", "LEIAME"), "fica", testTime)
			dst := t.TempDir()
			if tt.inPlace {
				dst = src
			}
			svc, _ := newTestService(t)

			got, err := svc.Organize(context.Background(), Request{SourceFolderPath: src, DestinationFolderPath: dst,
				Options: criteria.Default, IncludeSubfolders: true})

			if err != nil || got.MovedFiles != 4 || got.IgnoredWithoutExtension != 1 {
				t.Fatalf("Organize = (%+v, %v)", got, err)
			}
			assertTree(t, src, tt.wantSrc)
			if !tt.inPlace {
				assertTree(t, dst, tt.wantDest)
			}
		})
	}
}

func TestPreviewIncludingSubfolders(t *testing.T) {
	src := treeSource(t)
	svc, _ := newTestService(t)
	for _, tt := range []struct {
		subfolders bool
		want       int
	}{{false, 1}, {true, 4}} {
		got, err := svc.Preview(context.Background(), Request{SourceFolderPath: src, Options: criteria.Default, IncludeSubfolders: tt.subfolders})
		if err != nil || got.TotalFiles != tt.want {
			t.Errorf("subpastas = %v: Preview = (%+v, %v), want %d arquivos", tt.subfolders, got, err, tt.want)
		}
	}
}

func TestRemoveEmptied(t *testing.T) {
	tests := []struct {
		name  string
		dirs  []string
		files []string
		moved []string
		want  []string
	}{
		{"arquivo da raiz: a origem nunca é removida", nil, nil, []string{"x.txt"}, nil},
		{"cadeia vazia some inteira", []string{"a/b/c"}, nil, []string{"a/b/c/x.txt"}, nil},
		{"para na pasta que tem outra coisa", []string{"a/b/c"}, []string{"a/fica.txt"}, []string{"a/b/c/x.txt"}, []string{"a/fica.txt"}},
		{"várias pastas e arquivos repetidos na mesma", []string{"a/b", "c"}, []string{"c/fica.txt"},
			[]string{"a/b/x.txt", "a/b/y.txt", "c/z.txt"}, []string{"c/fica.txt"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := t.TempDir()
			for _, d := range tt.dirs {
				if err := os.MkdirAll(filepath.Join(src, filepath.FromSlash(d)), 0o755); err != nil {
					t.Fatal(err)
				}
			}
			for _, f := range tt.files {
				writeAt(t, filepath.Join(src, filepath.FromSlash(f)), "x", testTime)
			}
			var moved []store.MovedItem
			for _, f := range tt.moved {
				moved = append(moved, store.MovedItem{From: filepath.Join(src, filepath.FromSlash(f))})
			}
			svc, _ := newTestService(t)

			svc.removeEmptied(src, moved)

			if _, err := os.Stat(src); err != nil {
				t.Fatal("a origem não pode ser removida")
			}
			assertTree(t, src, tt.want)
		})
	}
}

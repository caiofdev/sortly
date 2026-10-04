package organizer

// função            | CC | casos
// NewPlanner        |  1 | todos
// Planner.Plan      |  5 | TestPlanReadDirError; TestPlanCanceled; TestPlanScenarios; TestPlanRuleError
// Planner.active    |  3 | TestPlanScenarios (critérios ligados e desligados)
// Plan.add          |  6 | TestPlanScenarios + TestPlanSpecialEntries: pasta; não regular; Info falha; sem extensão; erro de regra; ok
// Plan.addMove      |  2 | TestPlanScenarios: move; destino = origem (B8)
// newFile           |  1 | via cenários
// segmentsFor       |  4 | TestPlanScenarios: regra não se aplica; aplica; erro
//
// Valor-limite: pasta vazia; só subpastas; 1 arquivo; destino = origem sem subpasta gerada.

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/caiofdev/sortly/backend/metadata"
)

func TestPlanScenarios(t *testing.T) {
	cases := []struct {
		name  string
		files []string
		dirs  []string
		opts  Options
		dst   string // "" = destino igual à origem
		want  Plan   // Source/Destination preenchidos no teste; Moves relativos
	}{
		{name: "pasta vazia", opts: Options{ByExtension: true}, want: Plan{}},
		{name: "só subpastas", dirs: []string{"a", "b"}, opts: Options{ByExtension: true}, want: Plan{IgnoredFolders: 2}},
		{
			name:  "extensão: sem extensão é ignorado",
			files: []string{"foto.JPG", "LEIAME", ".gitignore", "backup.tar.gz"},
			opts:  Options{ByExtension: true},
			want: Plan{ProcessedFiles: 4, IgnoredWithoutExtension: 2, Moves: []Move{
				{"backup.tar.gz", filepath.Join("gz", "backup.tar.gz")},
				{"foto.JPG", filepath.Join("jpg", "foto.JPG")},
			}},
		},
		{
			name:  "sem extensão ligado, o arquivo sem extensão é movido pelos outros critérios",
			files: []string{"LEIAME"},
			opts:  Options{BySize: true},
			want:  Plan{ProcessedFiles: 1, Moves: []Move{{"LEIAME", filepath.Join("size-1mb", "LEIAME")}}},
		},
		{
			name:  "todos os critérios: aninhamento na ordem do registry",
			files: []string{"a.pdf", "b.mp4", "c.png", "d.txt"},
			opts:  Options{ByExtension: true, ByDate: true, BySize: true, ByResolution: true, ByDuration: true, ByPages: true},
			want: Plan{ProcessedFiles: 4, Moves: []Move{
				{"a.pdf", filepath.Join("pdf", "date-2026-03-05", "size-1mb", "pages-12", "a.pdf")},
				{"b.mp4", filepath.Join("mp4", "date-2026-03-05", "size-1mb", "duration-00h01m00s", "b.mp4")},
				{"c.png", filepath.Join("png", "date-2026-03-05", "size-1mb", "3x2", "c.png")},
				{"d.txt", filepath.Join("txt", "date-2026-03-05", "size-1mb", "d.txt")},
			}},
		},
		{
			name:  "destino = origem e nenhuma subpasta gerada: arquivo fica onde está (B8)",
			files: []string{"nota.txt", "foto.png"},
			opts:  Options{ByResolution: true},
			want: Plan{ProcessedFiles: 2, UnchangedFiles: 1, Moves: []Move{
				{"foto.png", filepath.Join("3x2", "foto.png")},
			}},
		},
		{
			name:  "destino diferente e nenhuma subpasta gerada: vai para a raiz do destino",
			files: []string{"nota.txt"},
			opts:  Options{ByResolution: true},
			dst:   "outro",
			want:  Plan{ProcessedFiles: 1, Moves: []Move{{"nota.txt", "nota.txt"}}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			src := t.TempDir()
			for _, d := range tc.dirs {
				mkdir(t, filepath.Join(src, d))
			}
			for _, f := range tc.files {
				writeAt(t, filepath.Join(src, f), "x", testTime)
			}
			dst := src
			if tc.dst != "" {
				dst = filepath.Join(t.TempDir(), tc.dst)
			}

			got, err := newTestPlanner().Plan(context.Background(), src, dst, tc.opts)
			if err != nil {
				t.Fatal(err)
			}
			want := absolutePlan(tc.want, src, dst)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("Plan =\n%+v\nwant\n%+v", got, want)
			}
		})
	}
}

func TestPlanSpecialEntries(t *testing.T) {
	src := t.TempDir()
	p := newTestPlanner()
	p.readDir = func(string) ([]fs.DirEntry, error) {
		return []fs.DirEntry{
			fakeEntry{name: "atalho", mode: fs.ModeSymlink},
			fakeEntry{name: "sumiu.txt", infoErr: fs.ErrNotExist},
		}, nil
	}

	got, err := p.Plan(context.Background(), src, src, Options{ByExtension: true})
	if err != nil {
		t.Fatal(err)
	}
	// O atalho é ignorado sem contar; o arquivo que sumiu conta como processado, mas não é movido.
	if got.ProcessedFiles != 1 || len(got.Moves) != 0 || got.IgnoredFolders != 0 {
		t.Fatalf("Plan = %+v", got)
	}
}

func TestPlanReadDirError(t *testing.T) {
	_, err := newTestPlanner().Plan(context.Background(), filepath.Join(t.TempDir(), "nada"), "x", Options{ByExtension: true})
	if !errors.Is(err, ErrInvalidSource) {
		t.Fatalf("err = %v, want ErrInvalidSource", err)
	}
}

func TestPlanCanceled(t *testing.T) {
	src := t.TempDir()
	writeAt(t, filepath.Join(src, "a.txt"), "x", testTime)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := newTestPlanner().Plan(ctx, src, src, Options{ByExtension: true}); !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}

func TestPlanRuleError(t *testing.T) {
	src := t.TempDir()
	writeAt(t, filepath.Join(src, "a.txt"), "x", testTime)
	p := NewPlanner([]SegmentRule{failingRule{}})

	if _, err := p.Plan(context.Background(), src, src, Options{ByExtension: true}); !errors.Is(err, errRead) {
		t.Fatalf("err = %v, want errRead", err)
	}
}

// --- apoio ---

var (
	testLoc  = time.FixedZone("BRT", -3*3600)
	testTime = time.Date(2026, 3, 5, 12, 0, 0, 0, testLoc)
)

func newTestPlanner() *Planner {
	meta := fakeMeta{size: metadata.Size{Width: 3, Height: 2}, seconds: 60, pages: 12}
	return NewPlanner(NewRules(meta, testLoc))
}

func absolutePlan(p Plan, src, dst string) Plan {
	p.Source, p.Destination = src, dst
	for i, m := range p.Moves {
		p.Moves[i] = Move{From: filepath.Join(src, m.From), To: filepath.Join(dst, m.To)}
	}
	return p
}

type failingRule struct{}

func (failingRule) Enabled(Options) bool { return true }
func (failingRule) Applies(File) bool    { return true }
func (failingRule) Segment(context.Context, File) (string, error) {
	return "", errRead
}

type fakeEntry struct {
	name    string
	mode    fs.FileMode
	infoErr error
}

func (e fakeEntry) Name() string               { return e.name }
func (e fakeEntry) IsDir() bool                { return e.mode.IsDir() }
func (e fakeEntry) Type() fs.FileMode          { return e.mode.Type() }
func (e fakeEntry) Info() (fs.FileInfo, error) { return nil, e.infoErr }

func mkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

func writeAt(t *testing.T, path, content string, mtime time.Time) {
	t.Helper()
	mkdir(t, filepath.Dir(path))
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, mtime, mtime); err != nil {
		t.Fatal(err)
	}
}

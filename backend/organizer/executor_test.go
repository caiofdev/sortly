package organizer

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/caiofdev/sortly/backend/store"
)

func TestApplySuccess(t *testing.T) {
	src, dst := t.TempDir(), t.TempDir()
	for _, f := range []string{"a.pdf", "b.pdf", "c.png"} {
		writeAt(t, filepath.Join(src, f), f, testTime)
	}
	writeAt(t, filepath.Join(dst, "pdf", "b.pdf"), "já existia", testTime)
	plan := Plan{Moves: []Move{
		{filepath.Join(src, "a.pdf"), filepath.Join(dst, "pdf", "a.pdf")},
		{filepath.Join(src, "b.pdf"), filepath.Join(dst, "pdf", "b.pdf")},
		{filepath.Join(src, "c.png"), filepath.Join(dst, "png", "c.png")},
	}}

	out, err := NewExecutor(nil).Apply(context.Background(), plan)
	if err != nil {
		t.Fatal(err)
	}

	wantMoved := []store.MovedItem{
		{From: filepath.Join(src, "a.pdf"), To: filepath.Join(dst, "pdf", "a.pdf")},
		{From: filepath.Join(src, "b.pdf"), To: filepath.Join(dst, "pdf", "b (1).pdf")},
		{From: filepath.Join(src, "c.png"), To: filepath.Join(dst, "png", "c.png")},
	}
	// pdf/ já existia: só png/ foi criada (#57).
	wantFolders := []string{filepath.Join(dst, "png")}
	if !equalMoved(out.MovedItems, wantMoved) || !equalStrings(out.CreatedFolders, wantFolders) || out.FailedFiles != 0 {
		t.Fatalf("Outcome = %+v", out)
	}
	assertFile(t, filepath.Join(dst, "pdf", "b.pdf"), "já existia")
	assertFile(t, filepath.Join(dst, "pdf", "b (1).pdf"), "b.pdf")
}

func TestApplyFailureMidBatch(t *testing.T) {
	src, dst := t.TempDir(), t.TempDir()
	for _, f := range []string{"a.txt", "b.txt", "c.txt"} {
		writeAt(t, filepath.Join(src, f), f, testTime)
	}
	var logs bytes.Buffer
	e := NewExecutor(slog.New(slog.NewTextHandler(&logs, nil)))
	realMove := e.move
	e.move = func(from, to string) (string, error) {
		if strings.HasSuffix(from, "b.txt") {
			return "", os.ErrPermission
		}
		return realMove(from, to)
	}
	plan := Plan{Moves: []Move{
		{filepath.Join(src, "a.txt"), filepath.Join(dst, "a.txt")},
		{filepath.Join(src, "b.txt"), filepath.Join(dst, "b.txt")},
		{filepath.Join(src, "c.txt"), filepath.Join(dst, "c.txt")},
	}}

	out, err := e.Apply(context.Background(), plan)

	if err != nil || out.FailedFiles != 1 || len(out.MovedItems) != 2 {
		t.Fatalf("Apply = (%+v, %v), want 2 movidos e 1 falha (B1)", out, err)
	}
	assertFile(t, filepath.Join(src, "b.txt"), "b.txt")
	if !strings.Contains(logs.String(), "b.txt") {
		t.Fatalf("a falha deveria ir para o log: %q", logs.String())
	}
}

func TestApplyMkdirFails(t *testing.T) {
	src := t.TempDir()
	writeAt(t, filepath.Join(src, "a.txt"), "x", testTime)
	blocker := filepath.Join(t.TempDir(), "arquivo")
	writeAt(t, blocker, "", testTime)

	out, err := NewExecutor(nil).Apply(context.Background(), Plan{Moves: []Move{
		{filepath.Join(src, "a.txt"), filepath.Join(blocker, "txt", "a.txt")},
	}})
	if err != nil || out.FailedFiles != 1 || len(out.MovedItems) != 0 {
		t.Fatalf("Apply = (%+v, %v)", out, err)
	}
}

func TestApplyRecordsEveryCreatedLevel(t *testing.T) {
	src := t.TempDir()
	dst := filepath.Join(t.TempDir(), "novo-destino")
	writeAt(t, filepath.Join(src, "a.pdf"), "a", testTime)
	writeAt(t, filepath.Join(src, "b.pdf"), "b", testTime)
	plan := Plan{Moves: []Move{
		{filepath.Join(src, "a.pdf"), filepath.Join(dst, "pdf", "pages-1", "a.pdf")},
		{filepath.Join(src, "b.pdf"), filepath.Join(dst, "pdf", "pages-2", "b.pdf")},
	}}

	out, err := NewExecutor(nil).Apply(context.Background(), plan)

	want := []string{
		filepath.Join(dst, "pdf", "pages-1"), filepath.Join(dst, "pdf"), dst,
		filepath.Join(dst, "pdf", "pages-2"),
	}
	if err != nil || !equalStrings(out.CreatedFolders, want) {
		t.Fatalf("CreatedFolders = %v (%v), want %v", out.CreatedFolders, err, want)
	}
}

func TestApplyMoveFailsKeepsCreatedFolder(t *testing.T) {
	src, dst := t.TempDir(), t.TempDir()
	e := NewExecutor(nil)
	e.move = func(string, string) (string, error) { return "", os.ErrPermission }

	out, _ := e.Apply(context.Background(), Plan{Moves: []Move{
		{filepath.Join(src, "a.txt"), filepath.Join(dst, "txt", "a.txt")},
	}})

	if want := []string{filepath.Join(dst, "txt")}; out.FailedFiles != 1 || !equalStrings(out.CreatedFolders, want) {
		t.Fatalf("Outcome = %+v, want a pasta criada registrada mesmo com a falha", out)
	}
}

func TestMissingDirs(t *testing.T) {
	root := filepath.VolumeName(os.TempDir()) + string(filepath.Separator)
	a := filepath.Join(root, "a")
	ab := filepath.Join(a, "b")
	tests := []struct {
		name     string
		existing map[string]bool
		want     []string
	}{
		{"já existe", map[string]bool{ab: true, a: true, root: true}, nil},
		{"só a última falta", map[string]bool{a: true, root: true}, []string{ab}},
		{"dois níveis faltam", map[string]bool{root: true}, []string{ab, a}},
		{"nem a raiz do volume existe: para nela", map[string]bool{}, []string{ab, a, root}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewExecutor(nil)
			e.exists = func(p string) bool { return tt.existing[p] }
			if got := e.missingDirs(ab); !equalStrings(got, tt.want) {
				t.Fatalf("missingDirs = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApplyCanceled(t *testing.T) {
	src, dst := t.TempDir(), t.TempDir()
	writeAt(t, filepath.Join(src, "a.txt"), "a", testTime)
	writeAt(t, filepath.Join(src, "b.txt"), "b", testTime)
	ctx, cancel := context.WithCancel(context.Background())
	e := NewExecutor(nil)
	realMove := e.move
	e.move = func(from, to string) (string, error) {
		defer cancel() // cancela depois do primeiro movimento
		return realMove(from, to)
	}

	out, err := e.Apply(ctx, Plan{Moves: []Move{
		{filepath.Join(src, "a.txt"), filepath.Join(dst, "a.txt")},
		{filepath.Join(src, "b.txt"), filepath.Join(dst, "b.txt")},
	}})
	if !errors.Is(err, context.Canceled) || len(out.MovedItems) != 1 {
		t.Fatalf("Apply = (%+v, %v), want journal com 1 item e context.Canceled", out, err)
	}
}

func equalMoved(a, b []store.MovedItem) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func assertFile(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil || string(got) != want {
		t.Fatalf("%s = (%q, %v), want %q", path, got, err, want)
	}
}

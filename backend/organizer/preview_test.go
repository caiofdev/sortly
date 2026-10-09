package organizer

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/caiofdev/sortly/backend/organizer/criteria"
	"github.com/caiofdev/sortly/backend/store"
)

func planWith(dst string, files int, folders map[string]int) Plan {
	plan := Plan{Destination: dst, ProcessedFiles: files}
	for folder, n := range folders {
		for i := range n {
			to := filepath.Join(dst, folder, fmt.Sprintf("f%d.txt", i))
			if folder == "" {
				to = filepath.Join(dst, fmt.Sprintf("f%d.txt", i))
			}
			plan.Moves = append(plan.Moves, Move{To: to})
		}
	}
	return plan
}

func folderNames(n int) map[string]int {
	folders := map[string]int{}
	for i := range n {
		folders[fmt.Sprintf("p%d", i)] = n - i
	}
	return folders
}

func TestSummarize(t *testing.T) {
	dst := filepath.Join("C:", "destino")
	tests := []struct {
		name string
		plan Plan
		want Preview
	}{
		{"nada para mover", planWith(dst, 2, nil), Preview{TotalFiles: 2, Folders: []FolderCount{}}},
		{"raiz do destino e empate pelo nome", planWith(dst, 4, map[string]int{"txt": 1, "": 1, "pdf": 2}),
			Preview{TotalFiles: 4, Folders: []FolderCount{{"pdf", 2}, {"", 1}, {"txt", 1}}}},
		{"5 pastas", planWith(dst, 15, folderNames(5)),
			Preview{TotalFiles: 15, Folders: []FolderCount{{"p0", 5}, {"p1", 4}, {"p2", 3}, {"p3", 2}, {"p4", 1}}}},
		{"6 pastas, sem outras", planWith(dst, 21, folderNames(6)),
			Preview{TotalFiles: 21, Folders: []FolderCount{{"p0", 6}, {"p1", 5}, {"p2", 4}, {"p3", 3}, {"p4", 2}, {"p5", 1}}}},
		{"8 pastas: as duas menores viram outras", planWith(dst, 36, folderNames(8)),
			Preview{TotalFiles: 36, Folders: []FolderCount{{"p0", 8}, {"p1", 7}, {"p2", 6}, {"p3", 5}, {"p4", 4}, {"p5", 3}}, OtherFiles: 3}},
		{"7 pastas: a menor vira outras", planWith(dst, 28, folderNames(7)),
			Preview{TotalFiles: 28, Folders: []FolderCount{{"p0", 7}, {"p1", 6}, {"p2", 5}, {"p3", 4}, {"p4", 3}, {"p5", 2}}, OtherFiles: 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := summarize(tt.plan); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("summarize =\n%+v\nwant\n%+v", got, tt.want)
			}
		})
	}
}

func TestFirstFolder(t *testing.T) {
	dst := filepath.Join("C:", "destino")
	tests := []struct {
		to, want string
	}{
		{filepath.Join(dst, "pdf", "date-2026-03-05", "a.pdf"), "pdf"},
		{filepath.Join(dst, "pdf", "a.pdf"), "pdf"},
		{filepath.Join(dst, "a.pdf"), ""},
	}
	for _, tt := range tests {
		if got := firstFolder(dst, tt.to); got != tt.want {
			t.Errorf("firstFolder(%q) = %q, want %q", tt.to, got, tt.want)
		}
	}
}

func TestPreview(t *testing.T) {
	src := t.TempDir()
	for _, f := range []string{"a.pdf", "b.pdf", "c.txt", "LEIAME"} {
		writeAt(t, filepath.Join(src, f), f, testTime)
	}
	svc, _ := newTestService(t)

	got, err := svc.Preview(context.Background(), Request{SourceFolderPath: src, Options: criteria.Default})

	want := Preview{TotalFiles: 4, Folders: []FolderCount{{"pdf", 2}, {"txt", 1}}}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("Preview = (%+v, %v), want %+v", got, err, want)
	}
	assertTree(t, src, []string{"LEIAME", "a.pdf", "b.pdf", "c.txt"})
}

func TestPreviewErrors(t *testing.T) {
	src := t.TempDir()
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	writeAt(t, filepath.Join(src, "a.txt"), "a", testTime)

	tests := []struct {
		name       string
		ctx        context.Context
		req        Request
		unreadable bool
		want       error
	}{
		{"origem inválida", context.Background(), Request{}, false, ErrInvalidSource},
		{"origem ilegível", context.Background(), Request{SourceFolderPath: src, Options: criteria.Default}, true, ErrInvalidSource},
		{"prévia cancelada", canceled, Request{SourceFolderPath: src, Options: criteria.Default}, false, context.Canceled},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _ := newTestService(t)
			if tt.unreadable {
				svc.planner.readDir = func(string) ([]os.DirEntry, error) { return nil, os.ErrPermission }
			}
			if _, err := svc.Preview(tt.ctx, tt.req); !errors.Is(err, tt.want) {
				t.Fatalf("err = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestMovedFolders(t *testing.T) {
	dst := filepath.Join("C:", "destino")
	moved := []store.MovedItem{
		{From: "a", To: filepath.Join(dst, "txt", "nota (1).txt")},
		{From: "b", To: filepath.Join(dst, "txt", "b.txt")},
		{From: "c", To: filepath.Join(dst, "pdf", "date-2026-03-05", "c.pdf")},
	}
	tests := []struct {
		name       string
		moved      []store.MovedItem
		want       []FolderCount
		wantOthers int
	}{
		{"nada movido", nil, []FolderCount{}, 0},
		{"pelo nome final e pela pasta de 1º nível", moved, []FolderCount{{"txt", 2}, {"pdf", 1}}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, others := movedFolders(dst, tt.moved)
			if !reflect.DeepEqual(got, tt.want) || others != tt.wantOthers {
				t.Fatalf("movedFolders = (%+v, %d), want (%+v, %d)", got, others, tt.want, tt.wantOthers)
			}
		})
	}
}

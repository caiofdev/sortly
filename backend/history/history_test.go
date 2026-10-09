package history

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

var testAt = time.Date(2026, 3, 5, 12, 0, 0, 0, time.UTC)

func entry(n int, status string) Entry {
	return Entry{At: testAt.Add(time.Duration(n) * time.Minute), SourceFolderPath: fmt.Sprintf("C:/o%d", n), MovedFiles: n, Status: status}
}

func newService(t *testing.T, content string) (*Service, *bytes.Buffer) {
	t.Helper()
	path := filepath.Join(t.TempDir(), ".sortly", "history.json")
	if content != "" {
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	var logs bytes.Buffer
	return New(path, slog.New(slog.NewTextHandler(&logs, nil))), &logs
}

func TestList(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []Entry
		logged  string
	}{
		{"sem arquivo", "", []Entry{}, ""},
		{"JSON null", "null", []Entry{}, ""},
		{"corrompido", "{", []Entry{}, "corrompido"},
		{"válido", `[{"at":"2026-03-05T12:01:00Z","sourceFolderPath":"C:/o1","destinationFolderPath":"","movedFiles":1,"status":"done"}]`,
			[]Entry{entry(1, StatusDone)}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, logs := newService(t, tt.content)
			if got := s.List(); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("List = %+v, want %+v", got, tt.want)
			}
			if !strings.Contains(logs.String(), tt.logged) {
				t.Fatalf("log = %q, want %q", logs.String(), tt.logged)
			}
		})
	}
}

func TestListUnreadableIsLogged(t *testing.T) {
	s, logs := newService(t, "")
	// Uma pasta no lugar do arquivo faz a leitura falhar (#80).
	if err := os.MkdirAll(s.path, 0o700); err != nil {
		t.Fatal(err)
	}
	if got := s.List(); len(got) != 0 || !strings.Contains(logs.String(), "ilegível") {
		t.Fatalf("List = %+v, log = %q", got, logs.String())
	}
}

func TestAddKeepsNewestFirstAndPersists(t *testing.T) {
	s, _ := newService(t, "")
	for i := 1; i <= 2; i++ {
		if err := s.Add(entry(i, StatusDone)); err != nil {
			t.Fatal(err)
		}
	}
	want := []Entry{entry(2, StatusDone), entry(1, StatusDone)}
	if got := New(s.path, nil).List(); !reflect.DeepEqual(got, want) {
		t.Fatalf("relido do disco = %+v, want %+v", got, want)
	}
}

func TestAddLimit(t *testing.T) {
	tests := []struct {
		before, want int
	}{
		{MaxEntries - 2, MaxEntries - 1},
		{MaxEntries - 1, MaxEntries},
		{MaxEntries, MaxEntries},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d antes", tt.before), func(t *testing.T) {
			s, _ := newService(t, "")
			for i := range tt.before {
				_ = s.Add(entry(i, StatusDone))
			}
			_ = s.Add(entry(999, StatusDone))
			got := s.List()
			if len(got) != tt.want || got[0].MovedFiles != 999 {
				t.Fatalf("%d entradas, primeira %+v", len(got), got[0])
			}
		})
	}
}

func TestMarkLastUndone(t *testing.T) {
	tests := []struct {
		name   string
		before []Entry
		want   []Entry
	}{
		{"sem histórico não muda nada", nil, []Entry{}},
		{"marca só a mais recente", []Entry{entry(1, StatusDone), entry(2, StatusDone)},
			[]Entry{entry(2, StatusUndone), entry(1, StatusDone)}},
		{"interrompida também pode ser desfeita", []Entry{entry(1, StatusCanceled)}, []Entry{entry(1, StatusUndone)}},
		{"já desfeita fica como está", []Entry{entry(1, StatusDone), entry(2, StatusUndone)},
			[]Entry{entry(2, StatusUndone), entry(1, StatusDone)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, _ := newService(t, "")
			for _, e := range tt.before {
				_ = s.Add(e)
			}
			if err := s.MarkLastUndone(); err != nil {
				t.Fatal(err)
			}
			if got := New(s.path, nil).List(); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("List = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestSaveFailureKeepsPrevious(t *testing.T) {
	dir := t.TempDir()
	blocker := filepath.Join(dir, "arquivo")
	if err := os.WriteFile(blocker, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	var logs bytes.Buffer
	s := New(filepath.Join(blocker, "history.json"), slog.New(slog.NewTextHandler(&logs, nil)))

	if err := s.Add(entry(1, StatusDone)); err == nil {
		t.Fatal("gravar dentro de um arquivo deveria falhar")
	}
	if len(s.List()) != 0 || !strings.Contains(logs.String(), "não salvo") {
		t.Fatalf("lista = %+v, log = %q", s.List(), logs.String())
	}
	if err := s.MarkLastUndone(); err != nil {
		t.Fatalf("sem histórico, marcar não grava nada: %v", err)
	}
}

func TestListDoesNotShareTheInternalList(t *testing.T) {
	s, _ := newService(t, "")
	_ = s.Add(entry(1, StatusDone))
	s.List()[0].Status = "alterado"
	if s.List()[0].Status != StatusDone {
		t.Fatal("a lista devolvida não pode compartilhar a interna")
	}
}

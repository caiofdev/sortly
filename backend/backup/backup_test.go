package backup

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func newStore(t *testing.T) (*Store, string, *bytes.Buffer) {
	t.Helper()
	root := filepath.Join(t.TempDir(), "substituidos")
	var logs bytes.Buffer
	s := New(root, slog.New(slog.NewTextHandler(&logs, nil)))
	s.now = func() time.Time { return time.Unix(0, 42) }
	return s, root, &logs
}

func TestNewFolder(t *testing.T) {
	s, root, _ := newStore(t)
	if got := s.NewFolder(); got != filepath.Join(root, "42") {
		t.Fatalf("NewFolder = %q", got)
	}
	if _, err := os.Stat(root); !os.IsNotExist(err) {
		t.Fatal("a pasta só é criada quando o primeiro arquivo é guardado")
	}
}

func TestContains(t *testing.T) {
	s, root, _ := newStore(t)
	tests := []struct {
		path string
		want bool
	}{
		{filepath.Join(root, "42"), true},
		{filepath.Join(root, "42", "pdf", "a.pdf"), true},
		{root, false},
		{filepath.Dir(root), false},
		{filepath.Join(filepath.Dir(root), "outra"), false},
	}
	for _, tt := range tests {
		if got := s.Contains(tt.path); got != tt.want {
			t.Errorf("Contains(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
	if (&Store{}).Contains(root) {
		t.Error("sem raiz, nada está dentro")
	}
}

func TestRemove(t *testing.T) {
	s, root, logs := newStore(t)
	folder := filepath.Join(root, "42")
	if err := os.MkdirAll(filepath.Join(folder, "pdf"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(folder, "pdf", "a.pdf"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(filepath.Dir(root), "do-usuario")
	if err := os.MkdirAll(outside, 0o700); err != nil {
		t.Fatal(err)
	}

	s.Remove("")
	s.Remove(outside)
	s.Remove(folder)

	if _, err := os.Stat(folder); !os.IsNotExist(err) {
		t.Fatal("a pasta da organização deveria ter sido apagada")
	}
	if _, err := os.Stat(outside); err != nil {
		t.Fatal("pasta fora da raiz não pode ser apagada")
	}
	if !strings.Contains(logs.String(), "fora da raiz") {
		t.Fatalf("log = %q", logs.String())
	}
}

func TestRemoveFailureIsLogged(t *testing.T) {
	s, root, logs := newStore(t)
	s.removeAll = func(string) error { return os.ErrPermission }
	s.Remove(filepath.Join(root, "42"))
	if !strings.Contains(logs.String(), "não apagado") {
		t.Fatalf("log = %q", logs.String())
	}
}

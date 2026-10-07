package logging

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultDir(t *testing.T) {
	stubConfigDir(t, "/config", nil)
	if got, err := DefaultDir(); err != nil || got != filepath.Join("/config", "Sortly", "logs") {
		t.Fatalf("DefaultDir = (%q, %v)", got, err)
	}

	stubConfigDir(t, "", errors.New("sem HOME"))
	if _, err := DefaultDir(); err == nil {
		t.Fatal("esperava erro")
	}
}

func TestOpen(t *testing.T) {
	t.Run("cria a pasta e grava", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "Sortly", "logs")
		logger, closeFn, err := Open(dir)
		if err != nil {
			t.Fatal(err)
		}
		logger.Info("organização concluída", "movidos", 3)
		if err := closeFn(); err != nil {
			t.Fatal(err)
		}
		data, _ := os.ReadFile(filepath.Join(dir, fileName))
		if !strings.Contains(string(data), "organização concluída") || !strings.Contains(string(data), "movidos=3") {
			t.Fatalf("log = %q", data)
		}
	})

	t.Run("pasta não pode ser criada", func(t *testing.T) {
		blocker := filepath.Join(t.TempDir(), "arquivo")
		writeFile(t, blocker, 0)
		if _, _, err := Open(filepath.Join(blocker, "logs")); err == nil {
			t.Fatal("esperava erro")
		}
	})

	t.Run("arquivo de log não abre", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.MkdirAll(filepath.Join(dir, fileName, "x"), 0o755); err != nil { // sortly.log é uma pasta (#10)
			t.Fatal(err)
		}
		if _, _, err := Open(dir); err == nil {
			t.Fatal("esperava erro")
		}
	})
}

func TestOpenDefault(t *testing.T) {
	t.Run("pasta padrão", func(t *testing.T) {
		base := t.TempDir()
		stubConfigDir(t, base, nil)
		var stderr bytes.Buffer
		logger, closeFn := OpenDefault(&stderr)
		logger.Info("ok")
		_ = closeFn()
		if _, err := os.Stat(filepath.Join(base, "Sortly", "logs", fileName)); err != nil || stderr.Len() != 0 {
			t.Fatalf("stat = %v, stderr = %q", err, stderr.String())
		}
	})

	t.Run("sem pasta de configuração usa stderr", func(t *testing.T) {
		stubConfigDir(t, "", errors.New("sem HOME"))
		var stderr bytes.Buffer
		logger, closeFn := OpenDefault(&stderr)
		logger.Info("ainda registra")
		if err := closeFn(); err != nil || !strings.Contains(stderr.String(), "ainda registra") {
			t.Fatalf("stderr = %q", stderr.String())
		}
	})

	t.Run("pasta padrão não abre usa stderr", func(t *testing.T) {
		blocker := filepath.Join(t.TempDir(), "arquivo")
		writeFile(t, blocker, 0)
		stubConfigDir(t, blocker, nil)
		var stderr bytes.Buffer
		_, _ = OpenDefault(&stderr)
		if !strings.Contains(stderr.String(), "indisponível") {
			t.Fatalf("stderr = %q", stderr.String())
		}
	})
}

func TestRotate(t *testing.T) {
	cases := []struct {
		name        string
		size        int64
		wantRotated bool
	}{
		{"limite - 1 mantém", MaxSize - 1, false},
		{"limite rotaciona", MaxSize, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), fileName)
			writeFile(t, path, tc.size)
			rotate(path, MaxSize)
			_, err := os.Stat(path + ".1")
			if rotated := err == nil; rotated != tc.wantRotated {
				t.Fatalf("rotacionou = %v, want %v", rotated, tc.wantRotated)
			}
		})
	}

	t.Run("inexistente", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), fileName)
		rotate(path, MaxSize)
		if _, err := os.Stat(path + ".1"); err == nil {
			t.Fatal("não deveria criar .1")
		}
	})
}

func stubConfigDir(t *testing.T, dir string, err error) {
	t.Helper()
	old := userConfigDir
	userConfigDir = func() (string, error) { return dir, err }
	t.Cleanup(func() { userConfigDir = old })
}

func writeFile(t *testing.T, path string, size int64) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := errors.Join(f.Truncate(size), f.Close()); err != nil {
		t.Fatal(err)
	}
}

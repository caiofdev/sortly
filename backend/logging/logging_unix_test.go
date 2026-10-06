//go:build !windows

package logging

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestOpenRestrictsPermissions(t *testing.T) {
	t.Run("pasta e arquivo novos só do dono", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "Sortly", "logs")
		_, closeFn, err := Open(dir)
		if err != nil {
			t.Fatal(err)
		}
		_ = closeFn()
		assertMode(t, filepath.Dir(dir), dirMode)
		assertMode(t, dir, dirMode)
		assertMode(t, filepath.Join(dir, fileName), fileMode)
	})

	// Regressão (#59): logs criados por versões anteriores eram legíveis por
	// outros usuários, e o modo do OpenFile não muda um arquivo existente.
	t.Run("pasta e arquivos antigos são corrigidos", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.Chmod(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(dir, fileName)
		for _, f := range []string{path, path + ".1"} {
			if err := os.WriteFile(f, []byte("antigo\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			if err := os.Chmod(f, 0o644); err != nil {
				t.Fatal(err)
			}
		}

		_, closeFn, err := Open(dir)
		if err != nil {
			t.Fatal(err)
		}
		_ = closeFn()
		assertMode(t, dir, dirMode)
		assertMode(t, path, fileMode)
		assertMode(t, path+".1", fileMode)
	})
}

func TestRestrict(t *testing.T) {
	tests := []struct {
		name  string
		files func(dir string) []string
	}{
		{"só a pasta", func(string) []string { return nil }},
		{"arquivo ausente é ignorado", func(dir string) []string { return []string{filepath.Join(dir, "nada.log")} }},
		{"arquivo existente fica só do dono", func(dir string) []string {
			f := filepath.Join(dir, "a.log")
			if err := os.WriteFile(f, nil, 0o644); err != nil {
				t.Fatal(err)
			}
			return []string{f}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.Chmod(dir, 0o755); err != nil {
				t.Fatal(err)
			}
			files := tt.files(dir)

			restrict(dir, files...)

			assertMode(t, dir, dirMode)
			for _, f := range files {
				if _, err := os.Stat(f); err == nil {
					assertMode(t, f, fileMode)
				}
			}
		})
	}
}

func assertMode(t *testing.T, path string, want fs.FileMode) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != want {
		t.Fatalf("modo de %s = %o, want %o", path, got, want)
	}
}

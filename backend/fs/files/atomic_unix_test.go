//go:build !windows

package files

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteAtomicOwnerOnly(t *testing.T) {
	dir := filepath.Join(t.TempDir(), ".sortly")
	path := filepath.Join(dir, "settings.json")

	if err := WriteAtomic(path, []byte("{}")); err != nil {
		t.Fatal(err)
	}

	for p, want := range map[string]os.FileMode{dir: 0o700, path: 0o600} {
		info, err := os.Stat(p)
		if err != nil || info.Mode().Perm() != want {
			t.Fatalf("modo de %s = (%v, %v), want %o", p, info, err, want)
		}
	}
}

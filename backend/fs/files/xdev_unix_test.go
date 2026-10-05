//go:build !windows

package files

import (
	"path/filepath"
	"testing"
)

func TestSyncParentDir(t *testing.T) {
	dir := t.TempDir()
	if err := syncParentDir(dir); err != nil {
		t.Fatalf("pasta existente: %v", err)
	}
	if err := syncParentDir(filepath.Join(dir, "nao-existe")); err == nil {
		t.Fatal("pasta inexistente: esperava erro")
	}
}

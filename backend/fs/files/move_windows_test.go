//go:build windows

package files

// Regressão: nomes que terminam em ponto ("arquivo.") são movidos sem perder o ponto,
// como na versão Electron (achado na checagem de paridade da issue #8).

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/caiofdev/sortly/backend/fs/paths"
)

func TestMoveUniqueTrailingDot(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "arquivo.")
	if err := os.WriteFile(paths.Native(src), []byte("ponto"), 0o644); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(dir, "size-1mb", "arquivo.")
	if err := MkdirAll(filepath.Dir(dst)); err != nil {
		t.Fatal(err)
	}

	got, err := MoveUnique(src, dst)
	if err != nil || got != dst {
		t.Fatalf("MoveUnique = (%q, %v), want %q", got, err, dst)
	}
	data, err := os.ReadFile(paths.Native(dst))
	if err != nil || string(data) != "ponto" {
		t.Fatalf("destino = (%q, %v)", data, err)
	}
	if _, err := os.Stat(paths.Native(src)); !os.IsNotExist(err) {
		t.Fatalf("origem deveria ter sumido: %v", err)
	}
}

package fsutil

// função          | CC | casos
// Exists          |  1 | TestExists: arquivo; pasta; inexistente
// RemoveEmptyDir  |  3 | TestRemoveEmptyDir: vazia; com conteúdo; arquivo; inexistente
//
// Valor-limite: pasta com 0 itens (removida) e com 1 item (preservada).

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExists(t *testing.T) {
	dir := t.TempDir()
	file := writeFile(t, dir, "a.txt", "x")

	if !Exists(file) || !Exists(dir) {
		t.Error("arquivo e pasta existentes deveriam existir")
	}
	if Exists(filepath.Join(dir, "nada")) {
		t.Error("caminho inexistente não deveria existir")
	}
}

func TestRemoveEmptyDir(t *testing.T) {
	t.Run("pasta vazia (0 itens) é removida", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "vazia")
		if err := os.Mkdir(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if !RemoveEmptyDir(dir) || Exists(dir) {
			t.Fatal("pasta vazia deveria ser removida")
		}
	})

	t.Run("pasta com 1 item é preservada", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "a.txt", "x")
		if RemoveEmptyDir(dir) || !Exists(dir) {
			t.Fatal("pasta com conteúdo não pode ser removida")
		}
	})

	t.Run("arquivo nunca é apagado", func(t *testing.T) {
		file := writeFile(t, t.TempDir(), "a.txt", "x")
		if RemoveEmptyDir(file) || !Exists(file) {
			t.Fatal("RemoveEmptyDir não pode apagar arquivos")
		}
	})

	t.Run("inexistente é ignorado", func(t *testing.T) {
		if RemoveEmptyDir(filepath.Join(t.TempDir(), "nada")) {
			t.Fatal("não deveria informar remoção")
		}
	})
}

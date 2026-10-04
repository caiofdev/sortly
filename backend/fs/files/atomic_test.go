package files

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteAtomic(t *testing.T) {
	t.Run("cria a pasta e grava", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "nova", "dados.json")
		if err := WriteAtomic(path, []byte(`{"a":1}`)); err != nil {
			t.Fatal(err)
		}
		assertContent(t, path, `{"a":1}`)
		assertNoTemp(t, path)
	})

	t.Run("substitui o arquivo existente", func(t *testing.T) {
		path := writeFile(t, t.TempDir(), "dados.json", "antigo")
		if err := WriteAtomic(path, []byte("novo")); err != nil {
			t.Fatal(err)
		}
		assertContent(t, path, "novo")
	})

	t.Run("pasta não pode ser criada", func(t *testing.T) {
		blocker := writeFile(t, t.TempDir(), "arquivo", "")
		if err := WriteAtomic(filepath.Join(blocker, "dados.json"), []byte("x")); err == nil {
			t.Fatal("esperava erro ao criar pasta dentro de um arquivo")
		}
	})

	t.Run("troca falha: arquivo anterior intacto e temporário removido", func(t *testing.T) {
		path := writeFile(t, t.TempDir(), "dados.json", "antigo")
		stubRename(t, func(string, string) error { return errors.New("disco cheio") })

		if err := WriteAtomic(path, []byte("novo")); err == nil {
			t.Fatal("esperava erro na troca do arquivo")
		}
		assertContent(t, path, "antigo")
		assertNoTemp(t, path)
	})

	t.Run("temporário não pode ser criado", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "dados.json")
		if err := os.Chmod(dir, 0o500); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.Chmod(dir, 0o700) })
		if probe, err := os.CreateTemp(dir, "sonda-*"); err == nil {
			_ = probe.Close()
			_ = os.Remove(probe.Name())
			t.Skip("o sistema permite criar arquivos numa pasta somente leitura (Windows)")
		}
		if err := WriteAtomic(path, []byte("x")); err == nil {
			t.Fatal("esperava erro ao criar o temporário")
		}
	})
}

func assertNoTemp(t *testing.T, path string) {
	t.Helper()
	matches, _ := filepath.Glob(filepath.Join(filepath.Dir(path), "."+filepath.Base(path)+"-*.tmp"))
	if len(matches) > 0 {
		t.Fatalf("temporários esquecidos: %v", matches)
	}
}

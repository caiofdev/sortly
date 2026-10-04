package files

// função          | CC | casos
// Move            |  3 | TestMove: mesmo volume; erro que não é de volume; outro volume (fallback)
// MoveUnique      |  3 | TestMoveUnique: livre; ocupado (não sobrescreve); reserva falha; movimento falha
// isCrossDevice   |  1 | TestIsCrossDevice
// copyThenRemove  |  4 | TestCopyThenRemove: sucesso; origem inexistente; cópia falha; origem não removível
// copyFile        |  5 | TestCopyFile: sucesso com mtime; sobrescreve destino; origem não abre; destino não abre; leitura falha
//
// Regressão B3: mover entre volumes (rename com EXDEV / ERROR_NOT_SAME_DEVICE) não pode falhar.
// Limite: mesmo volume vs outro volume; falha na última etapa (remover a origem) não duplica o arquivo.
// O caminho "falha ao fechar o destino" de copyFile não é reproduzível de forma portável.

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMove(t *testing.T) {
	t.Run("mesmo volume", func(t *testing.T) {
		dir := t.TempDir()
		src, dst := writeFile(t, dir, "a.txt", "conteúdo"), filepath.Join(dir, "b.txt")

		if err := Move(src, dst); err != nil {
			t.Fatal(err)
		}
		assertMoved(t, src, dst, "conteúdo")
	})

	t.Run("erro que não é de volume é devolvido sem fallback", func(t *testing.T) {
		dir := t.TempDir()
		src := writeFile(t, dir, "a.txt", "x")
		denied := &os.LinkError{Op: "rename", Old: src, New: "y", Err: fs.ErrPermission}
		stubRename(t, func(string, string) error { return denied })

		if err := Move(src, filepath.Join(dir, "b.txt")); !errors.Is(err, fs.ErrPermission) {
			t.Fatalf("err = %v, want ErrPermission", err)
		}
		assertExists(t, src)
	})

	t.Run("outro volume usa cópia e preserva a data (B3)", func(t *testing.T) {
		dir := t.TempDir()
		src, dst := writeFile(t, dir, "a.txt", "conteúdo"), filepath.Join(dir, "b.txt")
		mtime := time.Date(2026, 3, 5, 23, 59, 59, 0, time.Local)
		if err := os.Chtimes(src, mtime, mtime); err != nil {
			t.Fatal(err)
		}
		stubCrossDevice(t)

		if err := Move(src, dst); err != nil {
			t.Fatal(err)
		}
		assertMoved(t, src, dst, "conteúdo")
		assertModTime(t, dst, mtime)
	})
}

func TestMoveUnique(t *testing.T) {
	t.Run("destino livre", func(t *testing.T) {
		dir := t.TempDir()
		src, dst := writeFile(t, dir, "a.txt", "novo"), filepath.Join(dir, "sub-a.txt")

		got, err := MoveUnique(src, dst)
		if err != nil || got != dst {
			t.Fatalf("MoveUnique = (%q, %v), want (%q, nil)", got, err, dst)
		}
		assertMoved(t, src, dst, "novo")
	})

	t.Run("destino ocupado não é sobrescrito", func(t *testing.T) {
		dir := t.TempDir()
		src := writeFile(t, dir, "a.txt", "novo")
		existing := writeFile(t, dir, "b.txt", "antigo")

		got, err := MoveUnique(src, existing)
		if err != nil {
			t.Fatal(err)
		}
		if want := filepath.Join(dir, "b (1).txt"); got != want {
			t.Fatalf("MoveUnique = %q, want %q", got, want)
		}
		assertMoved(t, src, got, "novo")
		assertContent(t, existing, "antigo")
	})

	t.Run("reserva falha", func(t *testing.T) {
		dir := t.TempDir()
		src := writeFile(t, dir, "a.txt", "x")

		if _, err := MoveUnique(src, filepath.Join(dir, "nao-existe", "a.txt")); err == nil {
			t.Fatal("esperava erro de pasta inexistente")
		}
		assertExists(t, src)
	})

	t.Run("movimento falha e a reserva é desfeita", func(t *testing.T) {
		dir := t.TempDir()
		dst := filepath.Join(dir, "b.txt")

		if _, err := MoveUnique(filepath.Join(dir, "nao-existe.txt"), dst); err == nil {
			t.Fatal("esperava erro de origem inexistente")
		}
		assertNotExists(t, dst)
	})
}

func TestIsCrossDevice(t *testing.T) {
	if !isCrossDevice(&os.LinkError{Op: "rename", Err: errCrossDevice}) {
		t.Error("LinkError com errCrossDevice deveria ser reconhecido")
	}
	if isCrossDevice(fs.ErrPermission) {
		t.Error("ErrPermission não é erro de volume")
	}
}

func TestCopyThenRemove(t *testing.T) {
	t.Run("sucesso", func(t *testing.T) {
		dir := t.TempDir()
		src, dst := writeFile(t, dir, "a.txt", "x"), filepath.Join(dir, "b.txt")

		if err := copyThenRemove(src, dst); err != nil {
			t.Fatal(err)
		}
		assertMoved(t, src, dst, "x")
	})

	t.Run("origem inexistente", func(t *testing.T) {
		dir := t.TempDir()
		if err := copyThenRemove(filepath.Join(dir, "nada"), filepath.Join(dir, "b")); err == nil {
			t.Fatal("esperava erro")
		}
	})

	t.Run("cópia falha: destino apagado, origem intacta", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "pasta")
		if err := os.Mkdir(src, 0o755); err != nil {
			t.Fatal(err)
		}
		dst := filepath.Join(dir, "b")

		if err := copyThenRemove(src, dst); err == nil {
			t.Fatal("esperava erro ao ler uma pasta como arquivo")
		}
		assertExists(t, src)
		assertNotExists(t, dst)
	})

	t.Run("origem não removível: cópia apagada, sem duplicar", func(t *testing.T) {
		dir := t.TempDir()
		src, dst := writeFile(t, dir, "a.txt", "x"), filepath.Join(dir, "b.txt")
		old := removeFile
		removeFile = func(string) error { return fs.ErrPermission }
		t.Cleanup(func() { removeFile = old })

		if err := copyThenRemove(src, dst); !errors.Is(err, fs.ErrPermission) {
			t.Fatalf("err = %v, want ErrPermission", err)
		}
		assertContent(t, src, "x")
		assertNotExists(t, dst)
	})
}

func TestCopyFile(t *testing.T) {
	t.Run("copia conteúdo e data de modificação", func(t *testing.T) {
		dir := t.TempDir()
		src := writeFile(t, dir, "a.txt", "conteúdo")
		mtime := time.Date(2024, 2, 29, 0, 0, 0, 0, time.Local)
		if err := os.Chtimes(src, mtime, mtime); err != nil {
			t.Fatal(err)
		}
		dst := filepath.Join(dir, "b.txt")

		if err := copyFile(src, dst, stat(t, src)); err != nil {
			t.Fatal(err)
		}
		assertContent(t, dst, "conteúdo")
		assertModTime(t, dst, mtime)
	})

	t.Run("sobrescreve o destino reservado", func(t *testing.T) {
		dir := t.TempDir()
		src := writeFile(t, dir, "a.txt", "curto")
		dst := writeFile(t, dir, "b.txt", "conteúdo bem mais longo")

		if err := copyFile(src, dst, stat(t, src)); err != nil {
			t.Fatal(err)
		}
		assertContent(t, dst, "curto")
	})

	t.Run("origem não abre", func(t *testing.T) {
		dir := t.TempDir()
		info := stat(t, writeFile(t, dir, "modelo", ""))

		if err := copyFile(filepath.Join(dir, "nada"), filepath.Join(dir, "b"), info); err == nil {
			t.Fatal("esperava erro")
		}
	})

	t.Run("destino não abre", func(t *testing.T) {
		dir := t.TempDir()
		src := writeFile(t, dir, "a.txt", "x")

		if err := copyFile(src, filepath.Join(dir, "nao-existe", "b"), stat(t, src)); err == nil {
			t.Fatal("esperava erro")
		}
	})

	t.Run("leitura da origem falha", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "pasta")
		if err := os.Mkdir(src, 0o755); err != nil {
			t.Fatal(err)
		}

		if err := copyFile(src, filepath.Join(dir, "b"), stat(t, src)); err == nil {
			t.Fatal("esperava erro ao ler uma pasta")
		}
	})
}

func stubRename(t *testing.T, fn func(string, string) error) {
	t.Helper()
	old := rename
	rename = fn
	t.Cleanup(func() { rename = old })
}

// stubCrossDevice simula o erro do sistema ao renomear entre volumes.
func stubCrossDevice(t *testing.T) {
	stubRename(t, func(oldPath, newPath string) error {
		return &os.LinkError{Op: "rename", Old: oldPath, New: newPath, Err: errCrossDevice}
	})
}

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func stat(t *testing.T, path string) fs.FileInfo {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	return info
}

func assertMoved(t *testing.T, src, dst, content string) {
	t.Helper()
	assertNotExists(t, src)
	assertContent(t, dst, content)
}

func assertContent(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ler %s: %v", path, err)
	}
	if !bytes.Equal(got, []byte(want)) {
		t.Fatalf("%s = %q, want %q", path, got, want)
	}
}

func assertModTime(t *testing.T, path string, want time.Time) {
	t.Helper()
	if got := stat(t, path).ModTime(); !got.Equal(want) {
		t.Fatalf("mtime de %s = %v, want %v", path, got, want)
	}
}

func assertExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("%s deveria existir: %v", path, err)
	}
}

func assertNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("%s não deveria existir (stat err = %v)", path, err)
	}
}

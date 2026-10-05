package files

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/caiofdev/sortly/backend/fs/paths"
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

	t.Run("disco não confirma a gravação: cópia apagada, origem intacta", func(t *testing.T) {
		dir := t.TempDir()
		src, dst := writeFile(t, dir, "a.txt", "x"), filepath.Join(dir, "b.txt")
		stubSync(t, func(*os.File) error { return errors.New("E/S") })

		if err := copyThenRemove(src, dst); err == nil {
			t.Fatal("esperava erro")
		}
		assertContent(t, src, "x")
		assertNotExists(t, dst)
	})

	t.Run("pasta do destino não confirma a gravação: cópia apagada, origem intacta", func(t *testing.T) {
		dir := t.TempDir()
		src, dst := writeFile(t, dir, "a.txt", "x"), filepath.Join(dir, "b.txt")
		stubSyncDir(t, func(string) error { return errors.New("E/S") })

		if err := copyThenRemove(src, dst); err == nil {
			t.Fatal("esperava erro")
		}
		assertContent(t, src, "x")
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

// Regressão (#52): a origem só pode ser apagada depois de a cópia e a entrada
// dela na pasta estarem gravadas em disco.
func TestCopyThenRemoveSyncsBeforeRemovingSource(t *testing.T) {
	dir := t.TempDir()
	src, dst := writeFile(t, dir, "a.txt", "x"), filepath.Join(dir, "b.txt")
	var events []string
	stubSync(t, func(f *os.File) error {
		events = append(events, "sync")
		return f.Sync()
	})
	stubSyncDir(t, func(d string) error {
		events = append(events, "syncdir")
		return syncParentDir(d)
	})
	old := removeFile
	removeFile = func(name string) error {
		events = append(events, "remove")
		return old(name)
	}
	t.Cleanup(func() { removeFile = old })

	if err := copyThenRemove(src, dst); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(events, " "); got != "sync syncdir remove" {
		t.Fatalf("eventos = [%s], want [sync syncdir remove]", got)
	}
	assertMoved(t, src, dst, "x")
}

func TestCopyDurably(t *testing.T) {
	t.Run("sucesso grava a pasta do destino", func(t *testing.T) {
		dir := t.TempDir()
		src, dst := writeFile(t, dir, "a.txt", "x"), filepath.Join(dir, "b.txt")
		var synced string
		stubSyncDir(t, func(d string) error { synced = d; return nil })

		if err := copyDurably(src, dst, stat(t, src)); err != nil {
			t.Fatal(err)
		}
		assertContent(t, dst, "x")
		if want := paths.Native(dir); synced != want {
			t.Fatalf("pasta gravada = %q, want %q", synced, want)
		}
	})

	t.Run("cópia falha: pasta não é gravada", func(t *testing.T) {
		dir := t.TempDir()
		info := stat(t, writeFile(t, dir, "modelo", ""))
		stubSyncDir(t, func(string) error { t.Error("syncDir não deveria rodar"); return nil })

		if err := copyDurably(filepath.Join(dir, "nada"), filepath.Join(dir, "b"), info); err == nil {
			t.Fatal("esperava erro")
		}
	})

	t.Run("pasta não confirma a gravação", func(t *testing.T) {
		dir := t.TempDir()
		src := writeFile(t, dir, "a.txt", "x")
		stubSyncDir(t, func(string) error { return errors.New("E/S") })

		if err := copyDurably(src, filepath.Join(dir, "b.txt"), stat(t, src)); err == nil {
			t.Fatal("esperava erro quando o Sync da pasta falha")
		}
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

	t.Run("disco não confirma a gravação", func(t *testing.T) {
		dir := t.TempDir()
		src := writeFile(t, dir, "a.txt", "x")
		stubSync(t, func(*os.File) error { return errors.New("E/S") })

		if err := copyFile(src, filepath.Join(dir, "b.txt"), stat(t, src)); err == nil {
			t.Fatal("esperava erro quando o Sync falha")
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

func stubSync(t *testing.T, fn func(*os.File) error) {
	t.Helper()
	old := syncFile
	syncFile = fn
	t.Cleanup(func() { syncFile = old })
}

func stubSyncDir(t *testing.T, fn func(string) error) {
	t.Helper()
	old := syncDir
	syncDir = fn
	t.Cleanup(func() { syncDir = old })
}

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

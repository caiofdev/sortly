package files

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/caiofdev/sortly/backend/fs/paths"
)

// Substituíveis nos testes para simular falhas que não dá para provocar de
// forma portável: outro volume, origem que não pode ser removida, disco que não
// confirma a gravação (#5, #52).
var (
	rename     = os.Rename
	removeFile = os.Remove
	syncFile   = (*os.File).Sync
	syncDir    = syncParentDir
)

// Substitui dst se ele existir (um nome reservado com Reserve). Entre volumes o
// rename falha, e o arquivo é copiado, com a data de modificação preservada,
// antes de a origem ser removida (#5).
func Move(src, dst string) error {
	err := rename(paths.Native(src), paths.Native(dst))
	if err == nil || !isCrossDevice(err) {
		return err
	}
	return copyThenRemove(src, dst)
}

// Nunca sobrescreve: usa o primeiro nome livre a partir de dst (veja Reserve) e
// desfaz a reserva se o movimento falhar (#5).
func MoveUnique(src, dst string) (string, error) {
	final, err := Reserve(dst)
	if err != nil {
		return "", err
	}
	if err := Move(src, final); err != nil {
		_ = os.Remove(paths.Native(final))
		return "", err
	}
	return final, nil
}

// Com paths.Native, para nomes que o Win32 normalizaria (#8).
func MkdirAll(path string) error {
	return os.MkdirAll(paths.Native(path), 0o755)
}

func isCrossDevice(err error) bool {
	return errors.Is(err, errCrossDevice)
}

// Em qualquer falha, a cópia é apagada e a origem fica intacta: o arquivo
// nunca fica duplicado nem perdido (#5).
func copyThenRemove(src, dst string) error {
	info, err := os.Stat(paths.Native(src))
	if err != nil {
		return err
	}
	if err := copyDurably(src, dst, info); err != nil {
		_ = os.Remove(paths.Native(dst))
		return err
	}
	if err := removeFile(paths.Native(src)); err != nil {
		_ = os.Remove(paths.Native(dst))
		return fmt.Errorf("files: cópia feita, mas a origem não pôde ser removida: %w", err)
	}
	return nil
}

func copyDurably(src, dst string, info fs.FileInfo) error {
	if err := copyFile(src, dst, info); err != nil {
		return err
	}
	return syncDir(paths.Native(filepath.Dir(dst)))
}

func copyFile(src, dst string, info fs.FileInfo) error {
	in, err := os.Open(paths.Native(src))
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	out, err := os.OpenFile(paths.Native(dst), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	// Sem o Sync, a cópia pode estar só no cache do sistema quando a origem for
	// apagada: uma queda de energia ou um pendrive removido perderia o arquivo (#52).
	if err := errors.Join(syncFile(out), out.Close()); err != nil {
		return err
	}
	return os.Chtimes(paths.Native(dst), info.ModTime(), info.ModTime())
}

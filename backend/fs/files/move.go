package files

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/caiofdev/sortly/backend/fs/paths"
)

// Substituíveis nos testes para simular falhas que não dá para provocar de
// forma portável (outro volume, origem que não pode ser removida).
var (
	rename     = os.Rename
	removeFile = os.Remove
)

// Move move src para dst, substituindo dst se ele existir (por exemplo, um
// nome reservado com Reserve). Se src e dst estão em volumes diferentes, o
// rename falha; nesse caso o arquivo é copiado, a data de modificação é
// preservada e a origem é removida.
func Move(src, dst string) error {
	err := rename(paths.Native(src), paths.Native(dst))
	if err == nil || !isCrossDevice(err) {
		return err
	}
	return copyThenRemove(src, dst)
}

// MoveUnique move src para o primeiro nome livre a partir de dst (veja
// Reserve) e devolve o caminho final. Nunca sobrescreve um arquivo existente.
// Se o movimento falhar, a reserva é desfeita.
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

// MkdirAll cria a pasta e as intermediárias (com paths.Native no Windows).
func MkdirAll(path string) error {
	return os.MkdirAll(paths.Native(path), 0o755)
}

func isCrossDevice(err error) bool {
	return errors.Is(err, errCrossDevice)
}

// copyThenRemove é o fallback entre volumes. Em qualquer falha, a cópia é
// apagada e a origem continua intacta, então o arquivo nunca fica duplicado
// nem perdido.
func copyThenRemove(src, dst string) error {
	info, err := os.Stat(paths.Native(src))
	if err != nil {
		return err
	}
	if err := copyFile(src, dst, info); err != nil {
		_ = os.Remove(paths.Native(dst))
		return err
	}
	if err := removeFile(paths.Native(src)); err != nil {
		_ = os.Remove(paths.Native(dst))
		return fmt.Errorf("files: cópia feita, mas a origem não pôde ser removida: %w", err)
	}
	return nil
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
	if err := out.Close(); err != nil {
		return err
	}
	return os.Chtimes(paths.Native(dst), info.ModTime(), info.ModTime())
}

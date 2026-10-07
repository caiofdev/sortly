//go:build !windows

package files

import (
	"errors"
	"os"
	"syscall"
)

var errCrossDevice error = syscall.EXDEV

// No Unix, o Sync do arquivo não grava o nome na pasta: depois de uma queda, o
// conteúdo pode estar no disco sem nenhuma pasta apontando para ele (#52).
func syncParentDir(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	return errors.Join(d.Sync(), d.Close())
}

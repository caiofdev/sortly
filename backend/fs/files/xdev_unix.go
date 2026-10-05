//go:build !windows

package files

import (
	"errors"
	"os"
	"syscall"
)

var errCrossDevice error = syscall.EXDEV

// syncParentDir grava em disco a entrada da pasta onde a cópia foi criada. No
// Unix, o Sync do arquivo não garante o nome: depois de uma queda, o conteúdo
// pode estar no disco sem nenhuma pasta apontando para ele.
func syncParentDir(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	return errors.Join(d.Sync(), d.Close())
}

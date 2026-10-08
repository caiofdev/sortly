package files

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/caiofdev/sortly/backend/fs/paths"
)

// O temporário fica na mesma pasta porque o rename só é atômico dentro do mesmo
// volume, e o Sync antes do rename garante que um crash nunca deixe o arquivo
// pela metade: em qualquer falha, o anterior fica intacto. A pasta é criada só
// para o dono, como o temporário (os.CreateTemp usa 0o600), porque os dados do
// app trazem caminhos de arquivos do usuário (#6, #59).
func WriteAtomic(path string, data []byte) error {
	dir := paths.Native(filepath.Dir(path))
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+"-*.tmp")
	if err != nil {
		return err
	}
	_, writeErr := tmp.Write(data)
	if err := errors.Join(writeErr, tmp.Sync(), tmp.Close()); err != nil {
		_ = os.Remove(tmp.Name())
		return err
	}
	if err := rename(tmp.Name(), paths.Native(path)); err != nil {
		_ = os.Remove(tmp.Name())
		return err
	}
	return nil
}

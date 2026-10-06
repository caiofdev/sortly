package files

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/caiofdev/sortly/backend/fs/paths"
)

// WriteAtomic grava data em path sem nunca deixar um arquivo pela metade:
// escreve um temporário na mesma pasta (o rename só é atômico dentro do mesmo
// volume), força a gravação em disco e o renomeia por cima do atual. Cria a
// pasta se preciso, só para o dono, como o temporário (os.CreateTemp usa
// 0o600): os dados do app trazem caminhos de arquivos do usuário. Em qualquer
// falha, o arquivo anterior fica intacto.
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

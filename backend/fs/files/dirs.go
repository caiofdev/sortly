package files

import (
	"os"

	"github.com/caiofdev/sortly/backend/fs/paths"
)

// Exists informa se há algo (arquivo ou pasta) no caminho.
func Exists(path string) bool {
	_, err := os.Lstat(paths.Native(path))
	return err == nil
}

// RemoveEmptyDir remove a pasta somente se ela existir, for uma pasta e
// estiver vazia. Qualquer outra situação (não existe, é arquivo, tem
// conteúdo, sem permissão) é ignorada: nunca apaga arquivos nem conteúdo.
func RemoveEmptyDir(path string) bool {
	native := paths.Native(path)
	info, err := os.Lstat(native)
	if err != nil || !info.IsDir() {
		return false
	}
	return os.Remove(native) == nil
}

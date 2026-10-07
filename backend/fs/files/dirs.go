package files

import (
	"os"

	"github.com/caiofdev/sortly/backend/fs/paths"
)

func Exists(path string) bool {
	_, err := os.Lstat(paths.Native(path))
	return err == nil
}

// Nunca apaga arquivos nem conteúdo: se o caminho não existe, é arquivo, tem
// conteúdo ou não dá permissão, nada acontece (#9).
func RemoveEmptyDir(path string) bool {
	native := paths.Native(path)
	info, err := os.Lstat(native)
	if err != nil || !info.IsDir() {
		return false
	}
	return os.Remove(native) == nil
}

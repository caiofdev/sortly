package fsutil

import "os"

// Exists informa se há algo (arquivo ou pasta) no caminho.
func Exists(path string) bool {
	_, err := os.Lstat(nativePath(path))
	return err == nil
}

// RemoveEmptyDir remove a pasta somente se ela existir, for uma pasta e
// estiver vazia. Qualquer outra situação (não existe, é arquivo, tem
// conteúdo, sem permissão) é ignorada: nunca apaga arquivos nem conteúdo.
func RemoveEmptyDir(path string) bool {
	native := nativePath(path)
	info, err := os.Lstat(native)
	if err != nil || !info.IsDir() {
		return false
	}
	return os.Remove(native) == nil
}

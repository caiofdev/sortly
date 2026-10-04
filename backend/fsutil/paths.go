package fsutil

import (
	"path/filepath"
	"runtime"
	"strings"
)

// caseInsensitive reflete o sistema de arquivos padrão de cada plataforma:
// no Windows, "C:\Fotos" e "c:\fotos" são o mesmo caminho.
var caseInsensitive = runtime.GOOS == "windows"

// PathsEqual informa se dois caminhos apontam para o mesmo lugar,
// depois de torná-los absolutos e limpos.
func PathsEqual(a, b string) bool {
	return pathsEqual(a, b, caseInsensitive)
}

// IsInside informa se child é igual a root ou está dentro dele.
// "/a/bc" não está dentro de "/a/b".
func IsInside(child, root string) bool {
	return isInside(child, root, caseInsensitive)
}

func pathsEqual(a, b string, fold bool) bool {
	return normalize(a, fold) == normalize(b, fold)
}

func isInside(child, root string, fold bool) bool {
	c, r := normalize(child, fold), normalize(root, fold)
	if c == r {
		return true
	}
	// Raízes como "C:\" ou "/" já terminam com separador.
	if !strings.HasSuffix(r, string(filepath.Separator)) {
		r += string(filepath.Separator)
	}
	return strings.HasPrefix(c, r)
}

func normalize(p string, fold bool) string {
	abs, err := filepath.Abs(p)
	if err != nil {
		abs = filepath.Clean(p)
	}
	if fold {
		return strings.ToLower(abs)
	}
	return abs
}

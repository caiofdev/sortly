package paths

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Sistemas de arquivos padrão que não diferenciam maiúsculas: NTFS no Windows e
// APFS no macOS. Lá, "C:\Fotos" e "c:\fotos" (ou "~/Downloads" e "~/downloads")
// são o mesmo caminho (#58).
var caseInsensitiveOS = map[string]bool{"windows": true, "darwin": true}

var caseInsensitive = caseInsensitiveOS[runtime.GOOS]

// Iguais depois de absolutos e limpos (sem diferenciar maiúsculas no Windows e
// no macOS) ou, quando os dois existem, o mesmo arquivo no disco. A segunda forma
// cobre volumes fora do padrão da plataforma, como um pendrive FAT no Linux (#5, #58).
func Equal(a, b string) bool {
	return equal(a, b, caseInsensitive) || sameFile(a, b)
}

// child igual a root também conta como dentro; "/a/bc" não está dentro de
// "/a/b". Sem diferenciar maiúsculas no Windows e no macOS (#5, #58).
func IsInside(child, root string) bool {
	return isInside(child, root, caseInsensitive)
}

// Chave de mapa para caminhos: dois caminhos com a mesma chave são iguais para
// Equal (sem o teste de mesmo arquivo no disco) (#82).
func Key(p string) string {
	return normalize(p, caseInsensitive)
}

func equal(a, b string, fold bool) bool {
	return normalize(a, fold) == normalize(b, fold)
}

func isInside(child, root string, fold bool) bool {
	c, r := normalize(child, fold), normalize(root, fold)
	if c == r {
		return true
	}
	// Raízes como "C:\" ou "/" já terminam com separador (#5).
	if !strings.HasSuffix(r, string(filepath.Separator)) {
		r += string(filepath.Separator)
	}
	return strings.HasPrefix(c, r)
}

// Lstat para não seguir atalhos: um link e o seu alvo não são o mesmo lugar (#58).
func sameFile(a, b string) bool {
	ia, err := os.Lstat(Native(a))
	if err != nil {
		return false
	}
	ib, err := os.Lstat(Native(b))
	return err == nil && os.SameFile(ia, ib)
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

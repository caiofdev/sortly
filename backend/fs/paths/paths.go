package paths

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// caseInsensitiveOS são as plataformas cujo sistema de arquivos padrão não
// diferencia maiúsculas: NTFS no Windows e APFS no macOS. Lá, "C:\Fotos" e
// "c:\fotos" (ou "~/Downloads" e "~/downloads") são o mesmo caminho.
var caseInsensitiveOS = map[string]bool{"windows": true, "darwin": true}

var caseInsensitive = caseInsensitiveOS[runtime.GOOS]

// Equal informa se dois caminhos apontam para o mesmo lugar: iguais depois de
// absolutos e limpos (sem diferenciar maiúsculas no Windows e no macOS) ou,
// quando os dois existem, o mesmo arquivo no disco. A segunda forma cobre
// volumes que fogem do padrão da plataforma, como um pendrive FAT no Linux.
func Equal(a, b string) bool {
	return equal(a, b, caseInsensitive) || sameFile(a, b)
}

// IsInside informa se child é igual a root ou está dentro dele, sem
// diferenciar maiúsculas no Windows e no macOS. "/a/bc" não está dentro de "/a/b".
func IsInside(child, root string) bool {
	return isInside(child, root, caseInsensitive)
}

func equal(a, b string, fold bool) bool {
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

// sameFile usa Lstat para não seguir atalhos: um link e o seu alvo não são o
// mesmo lugar.
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

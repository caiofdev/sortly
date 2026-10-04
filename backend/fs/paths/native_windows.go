//go:build windows

package paths

import (
	"os"
	"path/filepath"
	"strings"
)

// Native converte para o formato estendido do Windows (\\?\C:\...), que
// desliga a normalização do Win32. Sem isso, nomes que terminam em ponto ou
// espaço ("arquivo.", vindos de Linux/macOS) perdem o final e o arquivo
// "some". O caminho devolvido aos chamadores continua no formato normal.
func Native(p string) string {
	if p == "" || strings.HasPrefix(p, `\\?\`) {
		return p
	}
	// filepath.Abs usa GetFullPathName, que também remove o ponto final;
	// por isso o caminho absoluto é montado de forma apenas léxica.
	abs := p
	if !filepath.IsAbs(p) {
		wd, err := os.Getwd()
		if err != nil {
			return p
		}
		abs = filepath.Join(wd, p)
	}
	abs = filepath.Clean(abs)
	if strings.HasPrefix(abs, `\\`) {
		return `\\?\UNC\` + abs[2:]
	}
	return `\\?\` + abs
}

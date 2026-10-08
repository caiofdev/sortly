package paths

import "path/filepath"

// Segue exatamente o path.extname do Node, usado na versão 1.0, que difere de
// filepath.Ext em arquivos ocultos: Ext(".gitignore") == "" (#5).
//
//	"foto.JPG" -> ".JPG"    "backup.tar.gz" -> ".gz"    "README" -> ""
//	"arquivo." -> "."       ".gitignore" -> ""          "..." -> "."
func Ext(name string) string {
	base := filepath.Base(name)
	startDot, preDotState := scanDots(base)

	if startDot == -1 || preDotState == 0 {
		return ""
	}
	// ".." não tem extensão (#5).
	if preDotState == 1 && startDot == len(base)-1 && startDot == 1 {
		return ""
	}
	return base[startDot:]
}

// Percorre o nome da direita para a esquerda, como o Node. startDot é a posição
// do último ponto (-1 se não houver); preDotState diz o que vem antes dele:
// 0 = nada, 1 = só pontos, -1 = algum outro caractere (#5).
func scanDots(base string) (startDot, preDotState int) {
	startDot = -1
	for i := len(base) - 1; i >= 0; i-- {
		if base[i] == '.' && startDot == -1 {
			startDot = i
			continue
		}
		if base[i] == '.' {
			preDotState = 1
			continue
		}
		if startDot != -1 {
			preDotState = -1
		}
	}
	return startDot, preDotState
}

package fsutil

import "path/filepath"

// Ext devolve a extensão do nome, com o ponto, seguindo exatamente o
// path.extname do Node usado na versão Electron. Difere de filepath.Ext em
// arquivos ocultos: Ext(".gitignore") == "" (filepath.Ext devolve ".gitignore").
//
//	"foto.JPG" -> ".JPG"    "backup.tar.gz" -> ".gz"    "README" -> ""
//	"arquivo." -> "."       ".gitignore" -> ""          "..." -> "."
func Ext(name string) string {
	base := filepath.Base(name)
	startDot, preDotState := scanDots(base)

	if startDot == -1 || preDotState == 0 {
		return ""
	}
	// ".." não tem extensão.
	if preDotState == 1 && startDot == len(base)-1 && startDot == 1 {
		return ""
	}
	return base[startDot:]
}

// scanDots percorre o nome da direita para a esquerda como o Node faz.
// startDot é a posição do último ponto (-1 se não houver). preDotState indica
// o que vem antes dele: 0 = nada, 1 = só pontos, -1 = algum outro caractere.
func scanDots(base string) (startDot, preDotState int) {
	startDot = -1
	for i := len(base) - 1; i >= 0; i-- {
		switch {
		case base[i] == '.' && startDot == -1:
			startDot = i
		case base[i] == '.':
			preDotState = 1
		case startDot != -1:
			preDotState = -1
		}
	}
	return startDot, preDotState
}

// Package metadata lê as informações usadas pelos critérios de organização:
// resolução de imagens, duração de vídeos mp4 e número de páginas de documentos.
//
// Só lê valores; os nomes de pasta ("1920x1080", "pages-12", "unknown") são do
// organizador. O comportamento segue as bibliotecas da versão 1.0 (image-size,
// music-metadata, pdf-lib e JSZip), conferido contra elas nas mesmas fixtures (#7).
package metadata

import (
	"context"
	"errors"
)

var (
	// O arquivo foi lido, mas a informação não existe ou é inválida: dimensão zero,
	// duração zero, contagem de páginas ausente (#7).
	ErrNoMetadata = errors.New("metadata: informação indisponível")

	// Sem leitor para o tipo de arquivo, como .doc (#7).
	ErrUnsupported = errors.New("metadata: tipo de arquivo não suportado")
)

type Size struct {
	Width  int
	Height int
}

// O valor zero está pronto para uso (#7).
type Reader struct{}

func (Reader) Resolution(path string) (Size, error) {
	return Resolution(path)
}

func (Reader) Duration(path string) (float64, error) {
	return Duration(path)
}

func (Reader) Pages(ctx context.Context, path, ext string) (int, error) {
	return Pages(ctx, path, ext)
}

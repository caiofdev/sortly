// Package metadata lê as informações usadas pelos critérios de organização:
// resolução de imagens, duração de vídeos mp4 e número de páginas de documentos.
//
// O pacote só lê valores; os nomes de pasta ("1920x1080", "pages-12",
// "duration-00h01m00s", "unknown") são responsabilidade do organizador.
// O comportamento segue as bibliotecas da versão 1.0 (image-size,
// music-metadata, pdf-lib e JSZip), conferido contra elas nas mesmas fixtures.
package metadata

import (
	"context"
	"errors"
)

var (
	// ErrNoMetadata indica que o arquivo foi lido, mas a informação não existe
	// ou é inválida (dimensão zero, duração zero, contagem de páginas ausente).
	ErrNoMetadata = errors.New("metadata: informação indisponível")

	// ErrUnsupported indica que não há leitor para o tipo de arquivo (ex.: .doc).
	ErrUnsupported = errors.New("metadata: tipo de arquivo não suportado")
)

// Size é a largura e altura de uma imagem, em pixels.
type Size struct {
	Width  int
	Height int
}

// Reader agrupa os leitores do pacote, para ser injetado no organizador.
// O valor zero está pronto para uso.
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

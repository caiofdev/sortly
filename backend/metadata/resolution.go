package metadata

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"image"
	"io"
	"os"

	// Decodificadores registrados em image.DecodeConfig (#7).
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/webp"
)

// Assinatura, cabeçalho do arquivo e largura/altura do DIB (#7).
const bmpHeaderSize = 26

// Lê só o cabeçalho, e o formato vem do conteúdo, não da extensão (como o
// image-size): um ".png" com conteúdo JPEG é lido como JPEG (#7).
func Resolution(path string) (Size, error) {
	f, err := os.Open(path)
	if err != nil {
		return Size{}, err
	}
	defer func() { _ = f.Close() }()

	r := bufio.NewReader(f)
	size, err := decodeSize(r)
	if err != nil {
		return Size{}, err
	}
	if size.Width <= 0 || size.Height <= 0 {
		return Size{}, ErrNoMetadata
	}
	return size, nil
}

func decodeSize(r *bufio.Reader) (Size, error) {
	if sig, _ := r.Peek(2); string(sig) == "BM" {
		return bmpSize(r)
	}
	cfg, _, err := image.DecodeConfig(r)
	if err != nil {
		return Size{}, fmt.Errorf("%w: %w", ErrNoMetadata, err)
	}
	return Size{Width: cfg.Width, Height: cfg.Height}, nil
}

// Como o image-size: a altura negativa (imagem de cima para baixo) vale o valor
// absoluto. Não depende de o BMP usar uma compressão suportada por um
// decodificador completo (#7).
func bmpSize(r io.Reader) (Size, error) {
	header := make([]byte, bmpHeaderSize)
	if _, err := io.ReadFull(r, header); err != nil {
		return Size{}, fmt.Errorf("%w: bmp: %w", ErrNoMetadata, err)
	}
	width := int64(binary.LittleEndian.Uint32(header[18:22]))
	height := int64(int32(binary.LittleEndian.Uint32(header[22:26])))
	if height < 0 {
		height = -height
	}
	return Size{Width: int(width), Height: int(height)}, nil
}

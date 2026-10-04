package metadata

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

// maxMoovSize limita a leitura do índice do vídeo (box "moov"). Ele costuma
// ter alguns KB a poucos MB; o conteúdo de mídia ("mdat") nunca é lido.
const maxMoovSize = 64 << 20

var errBadBox = errors.New("box mp4 inválido")

// Duration devolve a duração do mp4 em segundos.
//
// Usa a duração da primeira faixa de áudio, como a versão Electron
// (music-metadata). Se o vídeo não tem áudio, usa a duração do filme
// (box "mvhd"), que é a que o player mostra.
func Duration(path string) (float64, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer func() { _ = f.Close() }()

	moov, err := findTopLevelBox(f, "moov")
	if err != nil {
		return 0, fmt.Errorf("%w: mp4: %w", ErrNoMetadata, err)
	}
	if seconds, ok := firstAudioTrackDuration(moov); ok {
		return seconds, nil
	}
	if seconds, ok := headerSeconds(childBox(moov, "mvhd")); ok {
		return seconds, nil
	}
	return 0, ErrNoMetadata
}

// findTopLevelBox percorre os boxes do arquivo pulando o conteúdo, sem
// carregá-lo, e devolve o payload do primeiro box do tipo pedido.
func findTopLevelBox(r io.ReadSeeker, want string) ([]byte, error) {
	for {
		typ, size, err := readBoxHeader(r)
		if err != nil {
			return nil, err
		}
		if typ == want {
			return readPayload(r, size)
		}
		if _, err := r.Seek(size, io.SeekCurrent); err != nil {
			return nil, err
		}
	}
}

func readPayload(r io.ReadSeeker, size int64) ([]byte, error) {
	if size < 0 {
		return io.ReadAll(io.LimitReader(r, maxMoovSize))
	}
	if size > maxMoovSize {
		return nil, fmt.Errorf("%w: moov com %d bytes", errBadBox, size)
	}
	payload := make([]byte, size)
	_, err := io.ReadFull(r, payload)
	return payload, err
}

// readBoxHeader lê tamanho e tipo do próximo box e devolve o tamanho do
// payload. Tamanho 1 indica "largesize" (64 bits); tamanho 0 indica que o
// box vai até o fim do arquivo (devolvido como -1).
func readBoxHeader(r io.Reader) (string, int64, error) {
	var header [8]byte
	if _, err := io.ReadFull(r, header[:]); err != nil {
		return "", 0, err
	}
	size := int64(binary.BigEndian.Uint32(header[:4]))
	typ := string(header[4:])
	headerLen := int64(8)

	switch size {
	case 0:
		return typ, -1, nil
	case 1:
		var large [8]byte
		if _, err := io.ReadFull(r, large[:]); err != nil {
			return "", 0, err
		}
		size = int64(binary.BigEndian.Uint64(large[:]))
		headerLen = 16
	}
	if size < headerLen {
		return "", 0, fmt.Errorf("%w: %q com tamanho %d", errBadBox, typ, size)
	}
	return typ, size - headerLen, nil
}

// boxes divide um payload nos boxes filhos. Para no primeiro box malformado.
func boxes(data []byte) map[string][][]byte {
	children := map[string][][]byte{}
	for len(data) >= 8 {
		size := int(binary.BigEndian.Uint32(data[:4]))
		if size < 8 || size > len(data) {
			break
		}
		typ := string(data[4:8])
		children[typ] = append(children[typ], data[8:size])
		data = data[size:]
	}
	return children
}

// childBox devolve o payload do primeiro filho com o caminho de tipos dado
// (ex.: "mdia", "mdhd"), ou nil.
func childBox(data []byte, path ...string) []byte {
	for _, typ := range path {
		found := boxes(data)[typ]
		if len(found) == 0 {
			return nil
		}
		data = found[0]
	}
	return data
}

// firstAudioTrackDuration repete a regra do music-metadata: a duração vem do
// "mdhd" da primeira faixa de áudio (handler "soun"/"audi" com pelo menos um canal).
func firstAudioTrackDuration(moov []byte) (float64, bool) {
	for _, trak := range boxes(moov)["trak"] {
		if isAudioTrack(trak) {
			return headerSeconds(childBox(trak, "mdia", "mdhd"))
		}
	}
	return 0, false
}

func isAudioTrack(trak []byte) bool {
	hdlr := childBox(trak, "mdia", "hdlr")
	if len(hdlr) < 12 {
		return false
	}
	handler := string(hdlr[8:12])
	if handler != "soun" && handler != "audi" {
		return false
	}
	return audioChannels(childBox(trak, "mdia", "minf", "stbl", "stsd")) > 0
}

// audioChannels lê o número de canais da primeira descrição de amostra de som.
// stsd: versão/flags (4) + quantidade (4) + entrada [tamanho (4) + formato (4)
// + reservado (6) + índice (2) + versão (2) + revisão (2) + fornecedor (4) + canais (2)].
func audioChannels(stsd []byte) int {
	const channelsOffset = 8 + 8 + 8 + 8
	if len(stsd) < channelsOffset+2 {
		return 0
	}
	return int(binary.BigEndian.Uint16(stsd[channelsOffset:]))
}

// headerSeconds lê escala de tempo e duração de um "mvhd" ou "mdhd"
// (versão 0 com campos de 32 bits ou versão 1 com campos de 64 bits).
func headerSeconds(header []byte) (float64, bool) {
	if len(header) < 20 {
		return 0, false
	}
	var timescale, duration uint64
	if header[0] == 1 {
		if len(header) < 32 {
			return 0, false
		}
		timescale = uint64(binary.BigEndian.Uint32(header[20:24]))
		duration = binary.BigEndian.Uint64(header[24:32])
	} else {
		timescale = uint64(binary.BigEndian.Uint32(header[12:16]))
		duration = uint64(binary.BigEndian.Uint32(header[16:20]))
	}
	if timescale == 0 || duration == 0 {
		return 0, false
	}
	return float64(duration) / float64(timescale), true
}

package metadata

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

// O índice do vídeo (box "moov") costuma ter de alguns KB a poucos MB; o
// conteúdo de mídia ("mdat") nunca é lido (#7).
const maxMoovSize = 64 << 20

var errBadBox = errors.New("box mp4 inválido")

// Em segundos, pela duração da primeira faixa de áudio, como a versão 1.0
// (music-metadata). Sem áudio, usa a duração do filme (box "mvhd"), que é a que
// o player mostra (#7).
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

// Pula o conteúdo dos boxes sem carregá-lo (#7).
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

// Tamanho 1 indica "largesize" (64 bits); tamanho 0 indica que o box vai até o
// fim do arquivo, devolvido como -1 (#7).
func readBoxHeader(r io.Reader) (string, int64, error) {
	var header [8]byte
	if _, err := io.ReadFull(r, header[:]); err != nil {
		return "", 0, err
	}
	size := int64(binary.BigEndian.Uint32(header[:4]))
	typ := string(header[4:])
	headerLen := int64(8)

	if size == 0 {
		return typ, -1, nil
	}
	if size == 1 {
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

// Para no primeiro box malformado (#7).
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

// path é a sequência de tipos (ex.: "mdia", "mdhd"); nil se não houver (#7).
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

// Mesma regra do music-metadata: a duração vem do "mdhd" da primeira faixa de
// áudio (handler "soun"/"audi" com pelo menos um canal) (#7).
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

// Layout do stsd: versão/flags (4) + quantidade (4) + entrada [tamanho (4) +
// formato (4) + reservado (6) + índice (2) + versão (2) + revisão (2) +
// fornecedor (4) + canais (2)] (#7).
func audioChannels(stsd []byte) int {
	const channelsOffset = 8 + 8 + 8 + 8
	if len(stsd) < channelsOffset+2 {
		return 0
	}
	return int(binary.BigEndian.Uint16(stsd[channelsOffset:]))
}

// "mvhd" ou "mdhd", na versão 0 (campos de 32 bits) ou 1 (campos de 64 bits) (#7).
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

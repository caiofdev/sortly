package metadata

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"path/filepath"
	"testing"
)

// Fixtures em que a versão 1.0 (music-metadata) devolvia "duration-unknown"
// e o Go encontra a duração (#7).
const electronUnknown = true

func TestDurationFixtures(t *testing.T) {
	cases := []struct {
		file        string
		want        float64 // 0 = sem duração (ErrNoMetadata) (#7)
		improvement bool    // versão 1.0: duration-unknown (#7)
		why         string
	}{
		{"audio-0.4s.mp4", 0.4, false, ""},
		{"audio-0.5s.mp4", 0.5, false, ""},
		{"audio-59s.mp4", 59, false, ""},
		{"audio-60s.mp4", 60, false, ""},
		{"audio-3599s.mp4", 3599, false, ""},
		{"audio-differs-from-movie.mp4", 59.98, false, "vale a faixa de áudio, não o filme (60,02 s)"},
		{"audio-only.mp4", 5, false, ""},
		{"moov-at-end.mp4", 125.5, false, "índice depois do conteúdo"},
		{"largesize-mdat.mp4", 7, false, "box de 64 bits antes do índice"},
		{"audio-3600s-v1.mp4", 3600, electronUnknown, "music-metadata ignora o mdhd versão 1"},
		{"video-only-59s.mp4", 59, electronUnknown, "sem áudio: usa a duração do filme"},
		{"video-only-v1.mp4", 90, electronUnknown, "sem áudio: mvhd versão 1"},
		{"audio-zero-channels.mp4", 8, electronUnknown, "faixa de som sem canais não conta como áudio"},
		{"audio-zero-duration.mp4", 9, electronUnknown, "áudio sem duração: usa a do filme"},
		{"zero-duration.mp4", 0, false, ""},
		{"zero-timescale.mp4", 0, false, ""},
		{"no-moov.mp4", 0, false, ""},
		{"truncated.mp4", 0, false, ""},
		{"not-mp4.mp4", 0, false, ""},
	}

	for _, tc := range cases {
		t.Run(tc.file, func(t *testing.T) {
			got, err := Duration(filepath.Join("testdata", "mp4", tc.file))
			if tc.want == 0 {
				if !errors.Is(err, ErrNoMetadata) {
					t.Fatalf("Duration = (%v, %v), want ErrNoMetadata", got, err)
				}
				return
			}
			if err != nil || got != tc.want {
				t.Fatalf("Duration = (%v, %v), want %v (%s)", got, err, tc.want, tc.why)
			}
		})
	}

	t.Run("arquivo inexistente", func(t *testing.T) {
		if _, err := Duration(filepath.Join(t.TempDir(), "nada.mp4")); err == nil {
			t.Fatal("esperava erro")
		}
	})
}

func TestReadBoxHeader(t *testing.T) {
	cases := []struct {
		name     string
		data     []byte
		wantType string
		wantSize int64
		wantErr  bool
	}{
		{"normal", rawBox(20, "moov"), "moov", 12, false},
		{"tamanho 0 vai até o fim", rawBox(0, "mdat"), "mdat", -1, false},
		{"largesize", append(rawBox(1, "mdat"), be64(40)...), "mdat", 24, false},
		{"largesize truncado", append(rawBox(1, "mdat"), 0, 0), "", 0, true},
		{"tamanho menor que o cabeçalho", rawBox(4, "free"), "", 0, true},
		{"fim do arquivo", nil, "", 0, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			typ, size, err := readBoxHeader(bytes.NewReader(tc.data))
			if (err != nil) != tc.wantErr || typ != tc.wantType || size != tc.wantSize {
				t.Fatalf("readBoxHeader = (%q, %d, %v), want (%q, %d, erro=%v)",
					typ, size, err, tc.wantType, tc.wantSize, tc.wantErr)
			}
		})
	}
}

func TestReadPayload(t *testing.T) {
	t.Run("tamanho conhecido", func(t *testing.T) {
		got, err := readPayload(bytes.NewReader([]byte("abcdef")), 3)
		if err != nil || string(got) != "abc" {
			t.Fatalf("readPayload = (%q, %v)", got, err)
		}
	})
	t.Run("até o fim do arquivo", func(t *testing.T) {
		got, err := readPayload(bytes.NewReader([]byte("abcdef")), -1)
		if err != nil || string(got) != "abcdef" {
			t.Fatalf("readPayload = (%q, %v)", got, err)
		}
	})
	t.Run("acima do limite", func(t *testing.T) {
		if _, err := readPayload(bytes.NewReader(nil), maxMoovSize+1); !errors.Is(err, errBadBox) {
			t.Fatalf("err = %v, want errBadBox", err)
		}
	})
	t.Run("moov que vai até o fim do arquivo", func(t *testing.T) {
		file := append(rawBox(0, "moov"), mvhd(0, 1000, 2500)...)
		moov, err := findTopLevelBox(bytes.NewReader(file), "moov")
		if err != nil {
			t.Fatal(err)
		}
		if got, ok := headerSeconds(childBox(moov, "mvhd")); !ok || got != 2.5 {
			t.Fatalf("duração = (%v, %v), want 2.5", got, ok)
		}
	})
	t.Run("sem o box procurado", func(t *testing.T) {
		if _, err := findTopLevelBox(bytes.NewReader(rawBox(8, "free")), "moov"); !errors.Is(err, io.EOF) {
			t.Fatalf("err = %v, want EOF", err)
		}
	})
}

func TestBoxes(t *testing.T) {
	two := append(box("free", []byte("ab")), box("skip", []byte("c"))...)
	cases := []struct {
		name string
		data []byte
		want int
	}{
		{"vazio", nil, 0},
		{"dois filhos", two, 2},
		{"tamanho menor que 8", rawBox(7, "free"), 0},
		{"tamanho maior que o restante", rawBox(100, "free"), 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := 0
			for _, list := range boxes(tc.data) {
				got += len(list)
			}
			if got != tc.want {
				t.Fatalf("boxes = %d filhos, want %d", got, tc.want)
			}
		})
	}
}

func TestChildBox(t *testing.T) {
	data := box("mdia", box("mdhd", []byte("x")))
	if got := childBox(data, "mdia", "mdhd"); string(got) != "x" {
		t.Errorf("caminho existente = %q", got)
	}
	if got := childBox(data, "mdia", "hdlr"); got != nil {
		t.Errorf("caminho inexistente = %q, want nil", got)
	}
	if got := childBox(data); !bytes.Equal(got, data) {
		t.Errorf("caminho vazio deveria devolver o próprio dado")
	}
}

func TestIsAudioTrack(t *testing.T) {
	cases := []struct {
		name string
		trak []byte
		want bool
	}{
		{"sem hdlr", box("mdia", nil), false},
		{"vídeo", trakWith("vide", 2), false},
		{"som com 2 canais", trakWith("soun", 2), true},
		{"áudio (audi) com 1 canal", trakWith("audi", 1), true},
		{"som sem canais", trakWith("soun", 0), false},
		{"som sem descrição de amostra", box("mdia", box("hdlr", hdlr("soun"))), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isAudioTrack(tc.trak); got != tc.want {
				t.Fatalf("isAudioTrack = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestHeaderSeconds(t *testing.T) {
	cases := []struct {
		name   string
		header []byte
		want   float64
		wantOK bool
	}{
		{"curto demais", make([]byte, 19), 0, false},
		{"versão 0", mvhd(0, 1000, 59000)[8:], 59, true},
		{"versão 1", mvhd(1, 90000, 3600*90000)[8:], 3600, true},
		{"versão 1 curta", append([]byte{1}, make([]byte, 30)...), 0, false},
		{"escala de tempo 0", mvhd(0, 0, 1000)[8:], 0, false},
		{"duração 0", mvhd(0, 1000, 0)[8:], 0, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := headerSeconds(tc.header)
			if got != tc.want || ok != tc.wantOK {
				t.Fatalf("headerSeconds = (%v, %v), want (%v, %v)", got, ok, tc.want, tc.wantOK)
			}
		})
	}
}

// --- construtores de boxes para os testes ---

func rawBox(size uint32, typ string) []byte {
	return append(be32(size), typ...)
}

func box(typ string, payload []byte) []byte {
	return append(rawBox(uint32(8+len(payload)), typ), payload...)
}

func mvhd(version byte, timescale uint32, duration uint64) []byte {
	var b bytes.Buffer
	b.Write([]byte{version, 0, 0, 0})
	if version == 1 {
		b.Write(make([]byte, 16))
		b.Write(be32(timescale))
		b.Write(be64(duration))
	} else {
		b.Write(make([]byte, 8))
		b.Write(be32(timescale))
		b.Write(be32(uint32(duration)))
	}
	b.Write(make([]byte, 80))
	return box("mvhd", b.Bytes())
}

func hdlr(handler string) []byte {
	return append(append(make([]byte, 8), handler...), make([]byte, 12)...)
}

func trakWith(handler string, channels uint16) []byte {
	entry := make([]byte, 28)
	binary.BigEndian.PutUint16(entry[16:], channels)
	stsd := append(make([]byte, 8), box("mp4a", entry)...)
	stbl := box("stbl", box("stsd", stsd))
	return box("mdia", append(box("hdlr", hdlr(handler)), box("minf", stbl)...))
}

func be32(v uint32) []byte { return binary.BigEndian.AppendUint32(nil, v) }
func be64(v uint64) []byte { return binary.BigEndian.AppendUint64(nil, v) }

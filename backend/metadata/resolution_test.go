package metadata

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// Os valores esperados são os do image-size usado na versão 1.0 (#7).
func TestResolutionFixtures(t *testing.T) {
	cases := []struct {
		file string
		want Size // zero = "unknown" na versão 1.0 (#7)
	}{
		{"1x1.png", Size{1, 1}},
		{"1x1-lossless.webp", Size{1, 1}},
		{"1x1-lossy.webp", Size{1, 1}},
		{"3x2.png", Size{3, 2}},
		{"3x2.jpg", Size{3, 2}},
		{"3x2.gif", Size{3, 2}},
		{"3x2.bmp", Size{3, 2}},
		{"3x2-top-down.bmp", Size{3, 2}},
		{"jpeg-content.png", Size{3, 2}},
		{"empty.png", Size{}},
		{"truncated.png", Size{}},
		{"not-an-image.jpg", Size{}},
	}

	for _, tc := range cases {
		t.Run(tc.file, func(t *testing.T) {
			got, err := Resolution(filepath.Join("testdata", "image", tc.file))
			if tc.want == (Size{}) {
				if !errors.Is(err, ErrNoMetadata) {
					t.Fatalf("Resolution = (%v, %v), want ErrNoMetadata", got, err)
				}
				return
			}
			if err != nil || got != tc.want {
				t.Fatalf("Resolution = (%v, %v), want %v", got, err, tc.want)
			}
		})
	}
}

func TestResolutionEdgeCases(t *testing.T) {
	t.Run("arquivo inexistente", func(t *testing.T) {
		if _, err := Resolution(filepath.Join(t.TempDir(), "nada.png")); err == nil {
			t.Fatal("esperava erro")
		}
	})

	for _, tc := range []struct {
		name          string
		width, height int32
	}{
		{"largura 0", 0, 2},
		{"altura 0", 3, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "img.bmp")
			if err := os.WriteFile(path, bmpHeader(tc.width, tc.height), 0o644); err != nil {
				t.Fatal(err)
			}
			if _, err := Resolution(path); !errors.Is(err, ErrNoMetadata) {
				t.Fatalf("err = %v, want ErrNoMetadata", err)
			}
		})
	}
}

func TestBMPSize(t *testing.T) {
	cases := []struct {
		name    string
		data    []byte
		want    Size
		wantErr bool
	}{
		{"normal", bmpHeader(640, 480), Size{640, 480}, false},
		{"altura negativa (top-down)", bmpHeader(640, -480), Size{640, 480}, false},
		{"cabeçalho truncado", bmpHeader(640, 480)[:bmpHeaderSize-1], Size{}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := bmpSize(bytes.NewReader(tc.data))
			if (err != nil) != tc.wantErr || got != tc.want {
				t.Fatalf("bmpSize = (%v, %v), want (%v, erro=%v)", got, err, tc.want, tc.wantErr)
			}
		})
	}
}

func bmpHeader(width, height int32) []byte {
	var b bytes.Buffer
	b.WriteString("BM")
	_ = binary.Write(&b, binary.LittleEndian, []uint32{54, 0, 54, 40})
	_ = binary.Write(&b, binary.LittleEndian, []int32{width, height})
	b.Write(make([]byte, 28))
	return b.Bytes()
}

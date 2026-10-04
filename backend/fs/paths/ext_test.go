package paths

// função    | CC | casos
// Ext       |  6 | TestExt: sem ponto; ponto inicial (oculto); ".."; "..."; extensão simples; dupla; ponto final; caminho com pasta
// scanDots  |  6 | TestScanDots: vazio; sem ponto; um ponto; dois pontos; só pontos; ponto inicial
//
// Valor-limite: posição do ponto — primeiro caractere (".gitignore"), último ("arquivo."),
// único caractere antes do ponto ("a.b"), nomes só com pontos ("..", "...").

import (
	"path/filepath"
	"testing"
)

func TestExt(t *testing.T) {
	cases := []struct {
		name string
		want string
	}{
		{"README", ""},
		{".gitignore", ""},
		{".", ""},
		{"..", ""},
		{"...", "."},
		{"a.b", ".b"},
		{"foto.JPG", ".JPG"},
		{"backup.tar.gz", ".gz"},
		{"arquivo.", "."},
		{".config.json", ".json"},
		{"..hidden", ".hidden"},
		{filepath.Join("pasta.com.ponto", "README"), ""},
		{filepath.Join("pasta", "relatorio.pdf"), ".pdf"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Ext(tc.name); got != tc.want {
				t.Errorf("Ext(%q) = %q, want %q", tc.name, got, tc.want)
			}
		})
	}
}

func TestScanDots(t *testing.T) {
	cases := []struct {
		base         string
		wantStartDot int
		wantPreDot   int
	}{
		{"", -1, 0},
		{"README", -1, 0},
		{"a.b", 1, -1},
		{"a.b.c", 3, -1},
		{"..", 1, 1},
		{".gitignore", 0, 0},
	}

	for _, tc := range cases {
		t.Run(tc.base, func(t *testing.T) {
			startDot, preDot := scanDots(tc.base)
			if startDot != tc.wantStartDot || preDot != tc.wantPreDot {
				t.Errorf("scanDots(%q) = (%d, %d), want (%d, %d)",
					tc.base, startDot, preDot, tc.wantStartDot, tc.wantPreDot)
			}
		})
	}
}

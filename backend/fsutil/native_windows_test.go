//go:build windows

package fsutil

// função      | CC | casos
// nativePath  |  6 | TestNativePath: vazio; já estendido; absoluto com ponto final; UNC; relativo; com ".."
//
// Regressão: nomes que terminam em ponto ("arquivo.") são movidos sem perder o ponto,
// como na versão Electron (achado na checagem de paridade da issue #8).

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNativePath(t *testing.T) {
	wd, _ := os.Getwd()
	cases := []struct{ in, want string }{
		{"", ""},
		{`\\?\C:\x`, `\\?\C:\x`},
		{`C:\Users\ana\arquivo.`, `\\?\C:\Users\ana\arquivo.`},
		{`\\servidor\share\a.txt`, `\\?\UNC\servidor\share\a.txt`},
		{"rel", `\\?\` + filepath.Join(wd, "rel")},
		{`C:\a\..\b\nota.`, `\\?\C:\b\nota.`},
	}
	for _, tc := range cases {
		if got := nativePath(tc.in); got != tc.want {
			t.Errorf("nativePath(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestMoveUniqueTrailingDot(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "arquivo.")
	if err := os.WriteFile(nativePath(src), []byte("ponto"), 0o644); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(dir, "size-1mb", "arquivo.")
	if err := MkdirAll(filepath.Dir(dst)); err != nil {
		t.Fatal(err)
	}

	got, err := MoveUnique(src, dst)
	if err != nil || got != dst {
		t.Fatalf("MoveUnique = (%q, %v), want %q", got, err, dst)
	}
	data, err := os.ReadFile(nativePath(dst))
	if err != nil || string(data) != "ponto" {
		t.Fatalf("destino = (%q, %v)", data, err)
	}
	if _, err := os.Stat(nativePath(src)); !os.IsNotExist(err) {
		t.Fatalf("origem deveria ter sumido: %v", err)
	}
}

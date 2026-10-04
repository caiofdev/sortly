package fsutil

// função      | CC | casos
// PathsEqual  |  1 | TestExportedPathHelpers
// IsInside    |  1 | TestExportedPathHelpers
// pathsEqual  |  1 | TestPathsEqual (5 casos)
// isInside    |  3 | TestIsInside: igual; dentro; prefixo sem separador; pai; raiz com separador; caixa
// normalize   |  3 | TestNormalize: absoluto; relativo; com e sem ignorar caixa
//
// Valor-limite: "/a/b" vs "/a/bc" (prefixo de texto sem ser subpasta) e raiz do volume
// ("C:\" ou "/"), que já termina com separador.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPathsEqual(t *testing.T) {
	base := t.TempDir()

	cases := []struct {
		name string
		a, b string
		fold bool
		want bool
	}{
		{"mesmo caminho", filepath.Join(base, "Fotos"), filepath.Join(base, "Fotos"), false, true},
		{"caminho não limpo", filepath.Join(base, "x", "..", "Fotos") + string(filepath.Separator), filepath.Join(base, "Fotos"), false, true},
		{"caminhos diferentes", filepath.Join(base, "Fotos"), filepath.Join(base, "Videos"), false, false},
		{"caixa diferente, sensível", filepath.Join(base, "Fotos"), filepath.Join(base, "FOTOS"), false, false},
		{"caixa diferente, insensível", filepath.Join(base, "Fotos"), filepath.Join(base, "FOTOS"), true, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := pathsEqual(tc.a, tc.b, tc.fold); got != tc.want {
				t.Errorf("pathsEqual(%q, %q, %v) = %v, want %v", tc.a, tc.b, tc.fold, got, tc.want)
			}
		})
	}
}

func TestIsInside(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "destino")
	volumeRoot := filepath.VolumeName(base) + string(filepath.Separator)

	cases := []struct {
		name        string
		child, root string
		fold        bool
		want        bool
	}{
		{"igual à raiz", root, root, false, true},
		{"subpasta direta", filepath.Join(root, "pdf"), root, false, true},
		{"subpasta profunda", filepath.Join(root, "pdf", "pages-3"), root, false, true},
		{"prefixo de texto, mas pasta irmã", root + "2", root, false, false},
		{"pai não está dentro do filho", base, root, false, false},
		{"raiz do volume termina com separador", filepath.Join(volumeRoot, "x"), volumeRoot, false, true},
		{"caixa diferente, sensível", filepath.Join(strings.ToUpper(root), "pdf"), root, false, strings.ToUpper(root) == root},
		{"caixa diferente, insensível", filepath.Join(strings.ToUpper(root), "pdf"), root, true, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isInside(tc.child, tc.root, tc.fold); got != tc.want {
				t.Errorf("isInside(%q, %q, %v) = %v, want %v", tc.child, tc.root, tc.fold, got, tc.want)
			}
		})
	}
}

func TestNormalize(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if got, want := normalize("Rel", false), filepath.Join(wd, "Rel"); got != want {
		t.Errorf("relativo: normalize = %q, want %q", got, want)
	}

	abs := filepath.Join(t.TempDir(), "Pasta")
	if got := normalize(abs, false); got != abs {
		t.Errorf("absoluto: normalize = %q, want %q", got, abs)
	}
	if got, want := normalize(abs, true), strings.ToLower(abs); got != want {
		t.Errorf("insensível: normalize = %q, want %q", got, want)
	}
}

func TestExportedPathHelpers(t *testing.T) {
	root := t.TempDir()
	child := filepath.Join(root, "a")

	if !PathsEqual(root, root) {
		t.Error("PathsEqual(root, root) = false")
	}
	if !IsInside(child, root) {
		t.Error("IsInside(child, root) = false")
	}
	if IsInside(root, child) {
		t.Error("IsInside(root, child) = true")
	}
	// No Windows a comparação ignora maiúsculas; nos demais sistemas, não.
	if got := PathsEqual(child, strings.ToUpper(child)); got != caseInsensitive && strings.ToUpper(child) != child {
		t.Errorf("PathsEqual com caixa diferente = %v, caseInsensitive = %v", got, caseInsensitive)
	}
}

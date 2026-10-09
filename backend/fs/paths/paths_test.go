package paths

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEqual(t *testing.T) {
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
			if got := equal(tc.a, tc.b, tc.fold); got != tc.want {
				t.Errorf("equal(%q, %q, %v) = %v, want %v", tc.a, tc.b, tc.fold, got, tc.want)
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

	if !Equal(root, root) {
		t.Error("Equal(root, root) = false")
	}
	if !IsInside(child, root) {
		t.Error("IsInside(child, root) = false")
	}
	if IsInside(root, child) {
		t.Error("IsInside(root, child) = true")
	}
	// No Windows e no macOS a comparação ignora maiúsculas; no Linux, não (#58).
	if got := Equal(child, strings.ToUpper(child)); got != caseInsensitive && strings.ToUpper(child) != child {
		t.Errorf("Equal com caixa diferente = %v, caseInsensitive = %v", got, caseInsensitive)
	}
}

func TestCaseInsensitiveOS(t *testing.T) {
	for goos, want := range map[string]bool{"windows": true, "darwin": true, "linux": false, "freebsd": false} {
		if got := caseInsensitiveOS[goos]; got != want {
			t.Errorf("caseInsensitiveOS[%q] = %v, want %v", goos, got, want)
		}
	}
}

func TestSameFile(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(a, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "link.txt")
	if err := os.Link(a, link); err != nil {
		t.Fatal(err)
	}
	other := filepath.Join(dir, "b.txt")
	if err := os.WriteFile(other, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	missing := filepath.Join(dir, "nada.txt")
	tests := []struct {
		name string
		a, b string
		want bool
	}{
		{"o mesmo arquivo por outro nome (link físico)", a, link, true},
		{"arquivos diferentes", a, other, false},
		{"o primeiro não existe", missing, a, false},
		{"o segundo não existe", a, missing, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sameFile(tt.a, tt.b); got != tt.want {
				t.Fatalf("sameFile = %v, want %v", got, tt.want)
			}
		})
	}
	if !Equal(a, link) {
		t.Error("Equal deveria reconhecer o mesmo arquivo por outro caminho")
	}
}

func TestKey(t *testing.T) {
	a := filepath.Join(t.TempDir(), "Pasta", "a.txt")
	if Key(a) != Key(filepath.Join(filepath.Dir(a), ".", "a.txt")) {
		t.Fatal("caminhos iguais depois de limpos têm a mesma chave")
	}
	upper := filepath.Join(filepath.Dir(a), "A.TXT")
	if (Key(a) == Key(upper)) != caseInsensitive {
		t.Fatalf("maiúsculas: chaves iguais = %v, want %v (sistema sem diferenciar = %v)", Key(a) == Key(upper), caseInsensitive, caseInsensitive)
	}
}

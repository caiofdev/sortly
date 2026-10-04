package app

// função              | CC | casos
// resolveDroppedPath  |  5 | TestResolveDroppedPath: vazio; inexistente; pasta; arquivo; nem arquivo nem pasta

import (
	"errors"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestResolveDroppedPath(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "foto.jpg")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name    string
		path    string
		want    string
		wantErr error
	}{
		{"vazio", "", "", ErrDroppedInvalid},
		{"não existe mais", filepath.Join(dir, "sumiu.txt"), "", ErrDroppedMissing},
		{"pasta vira a própria origem", dir, dir, nil},
		{"arquivo usa a pasta que o contém", file, dir, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := resolveDroppedPath(tc.path)
			if got != tc.want || !errors.Is(err, tc.wantErr) {
				t.Fatalf("resolveDroppedPath = (%q, %v), want (%q, %v)", got, err, tc.want, tc.wantErr)
			}
		})
	}
}

func TestResolveDroppedPathUnsupported(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("sockets Unix como arquivo especial não são portáveis para o Windows")
	}
	path := filepath.Join(t.TempDir(), "s.sock")
	l, err := net.Listen("unix", path)
	if err != nil {
		t.Skip("socket Unix indisponível:", err)
	}
	defer func() { _ = l.Close() }()

	if _, err := resolveDroppedPath(path); !errors.Is(err, ErrDroppedUnsupported) {
		t.Fatalf("err = %v, want ErrDroppedUnsupported", err)
	}
}

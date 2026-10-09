package app

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/caiofdev/sortly/backend/apperr"
)

type fakeInfo struct{ dir bool }

func (f fakeInfo) Name() string       { return "destino" }
func (f fakeInfo) Size() int64        { return 0 }
func (f fakeInfo) Mode() os.FileMode  { return 0 }
func (f fakeInfo) ModTime() time.Time { return time.Time{} }
func (f fakeInfo) IsDir() bool        { return f.dir }
func (f fakeInfo) Sys() any           { return nil }

func TestFolderOpener(t *testing.T) {
	dir := func(string) (os.FileInfo, error) { return fakeInfo{dir: true}, nil }
	tests := []struct {
		name     string
		goos     string
		stat     func(string) (os.FileInfo, error)
		startErr error
		wantProg string
		wantCode string
	}{
		{"Windows abre no Explorer", "windows", dir, nil, "explorer", ""},
		{"macOS abre no Finder", "darwin", dir, nil, "open", ""},
		{"Linux usa o xdg-open", "linux", dir, nil, "xdg-open", ""},
		{"pasta apagada", "windows", func(string) (os.FileInfo, error) { return nil, os.ErrNotExist }, nil, "", "DESTINATION_NOT_FOUND"},
		{"virou arquivo", "windows", func(string) (os.FileInfo, error) { return fakeInfo{}, nil }, nil, "", "DESTINATION_NOT_FOUND"},
		{"programa não abre", "linux", dir, errors.New("xdg-open ausente"), "xdg-open", apperr.CodeUnexpected},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var prog, arg string
			o := folderOpener{goos: tt.goos, stat: tt.stat, start: func(name string, args ...string) error {
				prog, arg = name, args[0]
				return tt.startErr
			}}

			err := o.open(`C:\destino`)

			if prog != tt.wantProg || (prog != "" && arg != `C:\destino`) {
				t.Fatalf("abriu %q %q, want %q", prog, arg, tt.wantProg)
			}
			if tt.wantCode == "" && err != nil || tt.wantCode != "" && apperr.CodeOf(err) != tt.wantCode {
				t.Fatalf("err = %v (código %q), want %q", err, apperr.CodeOf(err), tt.wantCode)
			}
		})
	}
}

func TestStartDetached(t *testing.T) {
	if err := startDetached(os.Args[0], "-test.run=^$"); err != nil {
		t.Fatalf("o próprio binário de teste deveria iniciar: %v", err)
	}
	if err := startDetached("programa-que-nao-existe-sortly"); err == nil {
		t.Fatal("programa inexistente deveria falhar")
	}
}

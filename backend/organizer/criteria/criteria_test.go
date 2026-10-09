package criteria

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/caiofdev/sortly/backend/metadata"
)

type fakeMeta struct {
	size     metadata.Size
	sizeErr  error
	seconds  float64
	durErr   error
	pages    int
	pagesErr error
}

func (f fakeMeta) Resolution(string) (metadata.Size, error) { return f.size, f.sizeErr }
func (f fakeMeta) Duration(string) (float64, error)         { return f.seconds, f.durErr }
func (f fakeMeta) Pages(context.Context, string, string) (int, error) {
	return f.pages, f.pagesErr
}

var errRead = errors.New("falha de leitura")

func TestNewOrder(t *testing.T) {
	rules := New(fakeMeta{}, time.UTC)
	want := []string{"criteria.typeRule", "criteria.extensionRule", "criteria.dateRule", "criteria.sizeRule",
		"criteria.resolutionRule", "criteria.durationRule", "criteria.pagesRule"}
	for i, r := range rules {
		if got := reflect.TypeOf(r).String(); got != want[i] {
			t.Errorf("regra %d = %s, want %s", i, got, want[i])
		}
	}
}

func TestFileExt(t *testing.T) {
	cases := map[string]string{
		"arquivo": "", "arquivo.": "", ".gitignore": "", "A.TAR.GZ": "gz", "Foto.JPG": "jpg", "relatório.pdf": "pdf",
	}
	for name, want := range cases {
		if got := fileExt(name); got != want {
			t.Errorf("fileExt(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestSegmentOrUnknown(t *testing.T) {
	if got := segmentOrUnknown("pages-", "3", nil); got != "pages-3" {
		t.Errorf("sem erro = %q", got)
	}
	if got := segmentOrUnknown("", "3x2", errRead); got != "unknown" {
		t.Errorf("com erro = %q", got)
	}
}

func TestEnabledAndApplies(t *testing.T) {
	all := Options{ByType: true, ByDuration: true, ByPages: true, ByResolution: true, ByDate: true, BySize: true, ByExtension: true}
	for _, r := range New(fakeMeta{}, time.UTC) {
		if !r.Enabled(all) || r.Enabled(Options{}) {
			t.Errorf("%T: Enabled não segue a opção", r)
		}
	}

	cases := []struct {
		rule Rule
		ext  string
		want bool
	}{
		{extensionRule{}, "", true},
		{dateRule{}, "txt", true},
		{sizeRule{}, "txt", true},
		{resolutionRule{}, "webp", true},
		{resolutionRule{}, "pdf", false},
		{durationRule{}, "mp4", true},
		{durationRule{}, "mov", false},
		{pagesRule{}, "doc", true},
		{pagesRule{}, "jpg", false},
	}
	for _, tc := range cases {
		if got := tc.rule.Applies(File{Ext: tc.ext}); got != tc.want {
			t.Errorf("%T.Applies(%q) = %v, want %v", tc.rule, tc.ext, got, tc.want)
		}
	}
}

func TestNewFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Foto.JPG")
	if err := os.WriteFile(path, []byte("abc"), 0o600); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	got := NewFile(dir, info)
	if got.Name != "Foto.JPG" || got.Path != path || got.Ext != "jpg" || got.Size != 3 || !got.ModTime.Equal(info.ModTime()) {
		t.Errorf("NewFile = %+v", got)
	}
}

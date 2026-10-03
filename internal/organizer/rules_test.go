package organizer

// função                   | CC | casos
// NewRules                 |  1 | TestNewRulesOrder
// fileExt                  |  1 | TestFileExt (6)
// segmentOrUnknown         |  2 | TestSegmentOrUnknown
// extensionRule.Segment    |  2 | TestExtensionRule: com e sem extensão
// dateRule.Segment         |  1 | TestDateRule (5 fronteiras)
// sizeRule.Segment         |  1 | TestSizeRule (7 fronteiras)
// resolutionRule.Segment   |  1 | TestResolutionRule: ok; falha
// durationRule.Segment     |  3 | TestDurationRule: ok; falha; duração 0 sem erro
// formatDuration           |  1 | TestFormatDuration (10 fronteiras)
// pagesRule.Segment        |  2 | TestPagesRule: ok; falha; contexto cancelado
// *.Enabled / *.Applies    |  1 | TestEnabledAndApplies
// setOf                    |  2 | via regras
//
// Valor-limite: tamanho (0 B, 1 B, 1 MB, 1 MB + 1, 2 MB, 2 MB + 1); data (23:59:59 e
// 00:00:00 locais, virada de ano, 29/02); duração (0,4 / 0,5 / 59 / 60 / 3599 / 3600 s);
// extensão ("arquivo", "arquivo.", ".gitignore", "A.TAR.GZ").

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/caiofdev/sortly/internal/metadata"
)

// fakeMeta devolve valores fixos para os critérios de metadados.
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

func TestNewRulesOrder(t *testing.T) {
	rules := NewRules(fakeMeta{}, time.UTC)
	want := []string{"organizer.extensionRule", "organizer.dateRule", "organizer.sizeRule",
		"organizer.resolutionRule", "organizer.durationRule", "organizer.pagesRule"}
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

func TestExtensionRule(t *testing.T) {
	got, err := extensionRule{}.Segment(context.Background(), File{Ext: "pdf"})
	if err != nil || got != "pdf" {
		t.Errorf("com extensão = (%q, %v)", got, err)
	}
	if _, err := (extensionRule{}).Segment(context.Background(), File{}); !errors.Is(err, ErrSkipNoExtension) {
		t.Errorf("sem extensão: err = %v, want ErrSkipNoExtension", err)
	}
}

func TestDateRule(t *testing.T) {
	loc := time.FixedZone("BRT", -3*3600)
	cases := []struct {
		name string
		at   time.Time
		want string
	}{
		{"23:59:59 local", time.Date(2026, 3, 5, 23, 59, 59, 0, loc), "date-2026-03-05"},
		{"00:00:00 local", time.Date(2026, 3, 6, 0, 0, 0, 0, loc), "date-2026-03-06"},
		{"virada de ano", time.Date(2025, 12, 31, 23, 59, 59, 0, loc), "date-2025-12-31"},
		{"29 de fevereiro", time.Date(2024, 2, 29, 12, 0, 0, 0, loc), "date-2024-02-29"},
		{"UTC já é o dia seguinte, local não", time.Date(2026, 3, 6, 1, 30, 0, 0, time.UTC), "date-2026-03-05"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := dateRule{loc: loc}.Segment(context.Background(), File{ModTime: tc.at})
			if got != tc.want {
				t.Fatalf("Segment = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestSizeRule(t *testing.T) {
	cases := []struct {
		bytes int64
		want  string
	}{
		{0, "size-1mb"},
		{1, "size-1mb"},
		{1048575, "size-1mb"},
		{1048576, "size-1mb"},
		{1048577, "size-2mb"},
		{2097152, "size-2mb"},
		{2097153, "size-3mb"},
	}
	for _, tc := range cases {
		got, _ := sizeRule{}.Segment(context.Background(), File{Size: tc.bytes})
		if got != tc.want {
			t.Errorf("%d bytes = %q, want %q", tc.bytes, got, tc.want)
		}
	}
}

func TestResolutionRule(t *testing.T) {
	ok := resolutionRule{meta: fakeMeta{size: metadata.Size{Width: 1920, Height: 1080}}}
	if got, _ := ok.Segment(context.Background(), File{}); got != "1920x1080" {
		t.Errorf("ok = %q", got)
	}
	bad := resolutionRule{meta: fakeMeta{sizeErr: errRead}}
	if got, _ := bad.Segment(context.Background(), File{}); got != "unknown" {
		t.Errorf("falha = %q, want unknown", got)
	}
}

func TestDurationRule(t *testing.T) {
	cases := []struct {
		name string
		meta fakeMeta
		want string
	}{
		{"ok", fakeMeta{seconds: 61}, "duration-00h01m01s"},
		{"falha de leitura", fakeMeta{durErr: errRead}, "duration-unknown"},
		{"duração 0 sem erro", fakeMeta{seconds: 0}, "duration-unknown"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := durationRule{meta: tc.meta}.Segment(context.Background(), File{})
			if got != tc.want {
				t.Fatalf("Segment = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	cases := []struct {
		seconds float64
		want    string
	}{
		{-1, "00h00m00s"},
		{0.4, "00h00m00s"},
		{0.5, "00h00m01s"},
		{59, "00h00m59s"},
		{59.5, "00h01m00s"},
		{60, "00h01m00s"},
		{3599, "00h59m59s"},
		{3600, "01h00m00s"},
		{59.98, "00h01m00s"},
		{360000, "100h00m00s"},
	}
	for _, tc := range cases {
		if got := formatDuration(tc.seconds); got != tc.want {
			t.Errorf("formatDuration(%v) = %q, want %q", tc.seconds, got, tc.want)
		}
	}
}

func TestPagesRule(t *testing.T) {
	ok := pagesRule{meta: fakeMeta{pages: 12}}
	if got, err := ok.Segment(context.Background(), File{}); got != "pages-12" || err != nil {
		t.Errorf("ok = (%q, %v)", got, err)
	}
	bad := pagesRule{meta: fakeMeta{pagesErr: metadata.ErrUnsupported}}
	if got, _ := bad.Segment(context.Background(), File{Ext: "doc"}); got != "pages-unknown" {
		t.Errorf(".doc = %q, want pages-unknown", got)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := ok.Segment(ctx, File{}); !errors.Is(err, context.Canceled) {
		t.Errorf("contexto cancelado: err = %v", err)
	}
}

func TestEnabledAndApplies(t *testing.T) {
	all := Options{ByDuration: true, ByPages: true, ByResolution: true, ByDate: true, BySize: true, ByExtension: true}
	for _, r := range NewRules(fakeMeta{}, time.UTC) {
		if !r.Enabled(all) || r.Enabled(Options{}) {
			t.Errorf("%T: Enabled não segue a opção", r)
		}
	}

	cases := []struct {
		rule SegmentRule
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

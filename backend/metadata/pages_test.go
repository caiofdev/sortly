package metadata

import (
	"archive/zip"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// Os valores esperados são os do pdf-lib e do JSZip usados na versão 1.0 (#7).
func TestPagesFixtures(t *testing.T) {
	cases := []struct {
		file    string
		want    int // 0 = "pages-unknown" na versão 1.0 (#7)
		wantErr error
	}{
		{"1-page.pdf", 1, nil},
		{"3-pages-classic-xref.pdf", 3, nil},
		{"3-pages-object-streams.pdf", 3, nil},
		{"truncated.pdf", 0, ErrNoMetadata},
		{"not-a-pdf.pdf", 0, ErrNoMetadata},
		{"1-page.docx", 1, nil},
		{"12-pages.docx", 12, nil},
		{"lowercase-tag.docx", 5, nil},
		{"zero-pages.docx", 0, ErrNoMetadata},
		{"no-pages-tag.docx", 0, ErrNoMetadata},
		{"no-app-xml.docx", 0, ErrNoMetadata},
		{"not-a-zip.docx", 0, ErrNoMetadata},
		{"3-pages.odt", 3, nil},
		{"zero-pages.odt", 0, ErrNoMetadata},
		{"no-page-count.odt", 0, ErrNoMetadata},
		{"legacy.doc", 0, ErrUnsupported},
	}

	for _, tc := range cases {
		t.Run(tc.file, func(t *testing.T) {
			path := filepath.Join("testdata", "pages", tc.file)
			got, err := Pages(context.Background(), path, filepath.Ext(tc.file)[1:])
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("Pages = (%d, %v), want %v", got, err, tc.wantErr)
				}
				return
			}
			if err != nil || got != tc.want {
				t.Fatalf("Pages = (%d, %v), want %d", got, err, tc.want)
			}
		})
	}
}

func TestPagesExtensionIsCaseInsensitive(t *testing.T) {
	got, err := Pages(context.Background(), filepath.Join("testdata", "pages", "12-pages.docx"), "DOCX")
	if err != nil || got != 12 {
		t.Fatalf("Pages(DOCX) = (%d, %v), want 12", got, err)
	}
}

func TestPagesMissingFile(t *testing.T) {
	for _, ext := range []string{"pdf", "docx"} {
		if _, err := Pages(context.Background(), filepath.Join(t.TempDir(), "nada."+ext), ext); err == nil {
			t.Errorf("%s inexistente: esperava erro", ext)
		}
	}
}

func TestPagesCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Pages(ctx, filepath.Join("testdata", "pages", "1-page.pdf"), "pdf"); err == nil {
		t.Fatal("contexto cancelado deveria interromper a leitura do pdf")
	}
}

func TestZipRegexOverflow(t *testing.T) {
	path := filepath.Join(t.TempDir(), "enorme.docx")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	w := zip.NewWriter(f)
	entry, _ := w.Create("docProps/app.xml")
	_, _ = entry.Write([]byte("<Pages>99999999999999999999999</Pages>"))
	if err := errors.Join(w.Close(), f.Close()); err != nil {
		t.Fatal(err)
	}

	if _, err := Pages(context.Background(), path, "docx"); !errors.Is(err, ErrNoMetadata) {
		t.Fatalf("err = %v, want ErrNoMetadata", err)
	}
}

func TestReader(t *testing.T) {
	var r Reader
	if size, err := r.Resolution(filepath.Join("testdata", "image", "3x2.png")); err != nil || size != (Size{3, 2}) {
		t.Errorf("Resolution = (%v, %v)", size, err)
	}
	if secs, err := r.Duration(filepath.Join("testdata", "mp4", "audio-59s.mp4")); err != nil || secs != 59 {
		t.Errorf("Duration = (%v, %v)", secs, err)
	}
	if n, err := r.Pages(context.Background(), filepath.Join("testdata", "pages", "3-pages.odt"), "odt"); err != nil || n != 3 {
		t.Errorf("Pages = (%v, %v)", n, err)
	}
}

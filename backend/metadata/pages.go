package metadata

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

// Protege contra arquivos compactados maliciosos (zip bomb) (#7).
const maxXMLEntrySize = 1 << 20

type PageCounter interface {
	CountPages(ctx context.Context, path string) (int, error)
}

// Chave: extensão minúscula, sem ponto. ".doc" não tem contador e o
// organizador o coloca em "pages-unknown" (#7).
var pageCounters = map[string]PageCounter{
	"pdf":  pdfCounter{},
	"docx": zipRegexCounter{entry: "docProps/app.xml", pattern: regexp.MustCompile(`(?i)<Pages>(\d+)</Pages>`)},
	"odt":  zipRegexCounter{entry: "meta.xml", pattern: regexp.MustCompile(`(?i)meta:page-count="(\d+)"`)},
}

// ext vem sem ponto. Sem contador para a extensão, ErrUnsupported; contagem
// ausente ou zero, ErrNoMetadata (#7).
func Pages(ctx context.Context, path, ext string) (int, error) {
	counter, ok := pageCounters[strings.ToLower(ext)]
	if !ok {
		return 0, fmt.Errorf("%w: %q", ErrUnsupported, ext)
	}
	n, err := counter.CountPages(ctx, path)
	if err != nil {
		return 0, err
	}
	if n <= 0 {
		return 0, ErrNoMetadata
	}
	return n, nil
}

type pdfCounter struct{}

func (pdfCounter) CountPages(ctx context.Context, path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer func() { _ = f.Close() }()

	// Modo sem estado: o pdfcpu não lê nem cria configuração na pasta do usuário (#7).
	conf, err := api.LoadConfiguration(api.ConfigurationOptions{Mode: api.ConfigurationModeStateless})
	if err != nil {
		return 0, err
	}
	n, err := api.PageCount(ctx, f, conf)
	if err != nil {
		return 0, fmt.Errorf("%w: pdf: %w", ErrNoMetadata, err)
	}
	return n, nil
}

// docx e odt guardam a contagem num XML de metadados dentro do zip; o primeiro
// grupo da expressão é o número de páginas (#7).
type zipRegexCounter struct {
	entry   string
	pattern *regexp.Regexp
}

func (c zipRegexCounter) CountPages(_ context.Context, path string) (int, error) {
	xml, err := readZipEntry(path, c.entry)
	if err != nil {
		return 0, err
	}
	match := c.pattern.FindSubmatch(xml)
	if match == nil {
		return 0, ErrNoMetadata
	}
	n, err := strconv.Atoi(string(match[1]))
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrNoMetadata, err)
	}
	return n, nil
}

func readZipEntry(path, name string) ([]byte, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("%w: zip: %w", ErrNoMetadata, err)
	}
	defer func() { _ = zr.Close() }()

	f, err := zr.Open(name)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %w", ErrNoMetadata, name, err)
	}
	defer func() { _ = f.Close() }()

	return io.ReadAll(io.LimitReader(f, maxXMLEntrySize))
}

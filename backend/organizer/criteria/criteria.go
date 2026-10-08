// Package criteria define os critérios de organização (Strategy + registry,
// ADR 0003): cada critério ativo que se aplica a um arquivo gera uma subpasta,
// sempre na ordem de New (#8).
package criteria

import (
	"context"
	"errors"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/caiofdev/sortly/backend/fs/paths"
	"github.com/caiofdev/sortly/backend/metadata"
)

// O arquivo não é movido e conta em IgnoredWithoutExtension (#8).
var ErrSkipNoExtension = errors.New("criteria: arquivo sem extensão")

type File struct {
	Name    string
	Path    string
	Ext     string // minúscula, sem ponto; "" quando não há extensão (#8)
	Size    int64
	ModTime time.Time
}

func NewFile(dir string, info fs.FileInfo) File {
	return File{
		Name:    info.Name(),
		Path:    filepath.Join(dir, info.Name()),
		Ext:     fileExt(info.Name()),
		Size:    info.Size(),
		ModTime: info.ModTime(),
	}
}

type MetadataReader interface {
	Resolution(path string) (metadata.Size, error)
	Duration(path string) (float64, error)
	Pages(ctx context.Context, path, ext string) (int, error)
}

type Rule interface {
	Enabled(opts Options) bool
	// Ex.: resolução só vale para imagens (#8).
	Applies(f File) bool
	// Só devolve erro para pular o arquivo (ErrSkipNoExtension) ou quando o contexto é
	// cancelado (#8).
	Segment(ctx context.Context, f File) (string, error)
}

// Ordem de aninhamento das pastas: extensão → data → tamanho → resolução →
// duração → páginas (#8).
func New(meta MetadataReader, loc *time.Location) []Rule {
	return []Rule{
		extensionRule{},
		dateRule{loc: loc},
		sizeRule{},
		resolutionRule{meta: meta},
		durationRule{meta: meta},
		pagesRule{meta: meta},
	}
}

func fileExt(name string) string {
	return strings.ToLower(strings.TrimPrefix(paths.Ext(name), "."))
}

func segmentOrUnknown(prefix, segment string, err error) string {
	if err != nil {
		return prefix + "unknown"
	}
	return prefix + segment
}

func setOf(values ...string) map[string]bool {
	set := make(map[string]bool, len(values))
	for _, v := range values {
		set[v] = true
	}
	return set
}

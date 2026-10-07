// Package criteria define os critérios de organização (Strategy + registry,
// ADR 0003): cada critério ativo que se aplica a um arquivo gera uma subpasta,
// sempre na ordem de New.
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

// ErrSkipNoExtension sinaliza que o arquivo não tem extensão e o critério de
// extensão está ligado: ele não é movido (conta em IgnoredWithoutExtension).
var ErrSkipNoExtension = errors.New("criteria: arquivo sem extensão")

// File é o que as regras sabem sobre um arquivo da pasta de origem.
type File struct {
	Name    string
	Path    string
	Ext     string // minúscula, sem ponto; "" quando não há extensão
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

// MetadataReader lê os metadados usados pelos critérios. É satisfeito por metadata.Reader.
type MetadataReader interface {
	Resolution(path string) (metadata.Size, error)
	Duration(path string) (float64, error)
	Pages(ctx context.Context, path, ext string) (int, error)
}

type Rule interface {
	Enabled(opts Options) bool
	// Applies informa se o critério vale para o arquivo (ex.: resolução só para imagens).
	Applies(f File) bool
	// Segment devolve o nome da subpasta. Só devolve erro para pular o
	// arquivo (ErrSkipNoExtension) ou quando o contexto é cancelado.
	Segment(ctx context.Context, f File) (string, error)
}

// New monta o registry na ordem de aninhamento das pastas:
// extensão → data → tamanho → resolução → duração → páginas.
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

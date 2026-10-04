package organizer

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/caiofdev/sortly/backend/fs/paths"
	"github.com/caiofdev/sortly/backend/metadata"
)

const bytesPerMB = 1024 * 1024

// Extensões que cada critério de metadados considera (iguais às da versão Electron).
var (
	imageExtensions = setOf("png", "jpg", "jpeg", "gif", "webp", "bmp")
	videoExtensions = setOf("mp4")
	// ".doc" não tem leitor de páginas, mas continua aqui: vai para "pages-unknown".
	pagedExtensions = setOf("pdf", "doc", "docx", "odt")
)

// File é o que as regras sabem sobre um arquivo da pasta de origem.
type File struct {
	Name    string
	Path    string
	Ext     string // minúscula, sem ponto; "" quando não há extensão
	Size    int64
	ModTime time.Time
}

// MetadataReader lê os metadados usados pelos critérios. É satisfeito por metadata.Reader.
type MetadataReader interface {
	Resolution(path string) (metadata.Size, error)
	Duration(path string) (float64, error)
	Pages(ctx context.Context, path, ext string) (int, error)
}

// SegmentRule é um critério de organização (Strategy). Cada regra ativa que
// se aplica ao arquivo gera uma subpasta.
type SegmentRule interface {
	// Enabled informa se o critério está ligado nas opções.
	Enabled(opts Options) bool
	// Applies informa se o critério vale para o arquivo (ex.: resolução só para imagens).
	Applies(f File) bool
	// Segment devolve o nome da subpasta. Só devolve erro para pular o
	// arquivo (ErrSkipNoExtension) ou quando o contexto é cancelado.
	Segment(ctx context.Context, f File) (string, error)
}

// NewRules monta o registry na ordem de aninhamento das pastas:
// extensão → data → tamanho → resolução → duração → páginas.
func NewRules(meta MetadataReader, loc *time.Location) []SegmentRule {
	return []SegmentRule{
		extensionRule{},
		dateRule{loc: loc},
		sizeRule{},
		resolutionRule{meta: meta},
		durationRule{meta: meta},
		pagesRule{meta: meta},
	}
}

// fileExt devolve a extensão como a versão Electron: path.extname, sem o ponto, minúscula.
func fileExt(name string) string {
	return strings.ToLower(strings.TrimPrefix(paths.Ext(name), "."))
}

// segmentOrUnknown devolve o segmento ou, se a leitura falhou, prefix + "unknown".
func segmentOrUnknown(prefix, segment string, err error) string {
	if err != nil {
		return prefix + "unknown"
	}
	return prefix + segment
}

type extensionRule struct{}

func (extensionRule) Enabled(o Options) bool { return o.ByExtension }
func (extensionRule) Applies(File) bool      { return true }

func (extensionRule) Segment(_ context.Context, f File) (string, error) {
	if f.Ext == "" {
		return "", ErrSkipNoExtension
	}
	return f.Ext, nil
}

type dateRule struct{ loc *time.Location }

func (dateRule) Enabled(o Options) bool { return o.ByDate }
func (dateRule) Applies(File) bool      { return true }

// Segment usa a data de modificação no fuso local, como a versão Electron.
func (r dateRule) Segment(_ context.Context, f File) (string, error) {
	return "date-" + f.ModTime.In(r.loc).Format("2006-01-02"), nil
}

type sizeRule struct{}

func (sizeRule) Enabled(o Options) bool { return o.BySize }
func (sizeRule) Applies(File) bool      { return true }

// Segment arredonda para cima em MB, com mínimo de 1: 0 B e 1 MB exato viram "size-1mb".
func (sizeRule) Segment(_ context.Context, f File) (string, error) {
	mb := max(1, (f.Size+bytesPerMB-1)/bytesPerMB)
	return fmt.Sprintf("size-%dmb", mb), nil
}

type resolutionRule struct{ meta MetadataReader }

func (resolutionRule) Enabled(o Options) bool { return o.ByResolution }
func (resolutionRule) Applies(f File) bool    { return imageExtensions[f.Ext] }

func (r resolutionRule) Segment(_ context.Context, f File) (string, error) {
	size, err := r.meta.Resolution(f.Path)
	return segmentOrUnknown("", fmt.Sprintf("%dx%d", size.Width, size.Height), err), nil
}

type durationRule struct{ meta MetadataReader }

func (durationRule) Enabled(o Options) bool { return o.ByDuration }
func (durationRule) Applies(f File) bool    { return videoExtensions[f.Ext] }

func (r durationRule) Segment(_ context.Context, f File) (string, error) {
	seconds, err := r.meta.Duration(f.Path)
	if err == nil && seconds <= 0 {
		err = metadata.ErrNoMetadata
	}
	return segmentOrUnknown("duration-", formatDuration(seconds), err), nil
}

// formatDuration arredonda como o Math.round do JavaScript (0,5 sobe) e
// formata como HHhMMmSSs.
func formatDuration(seconds float64) string {
	total := int64(math.Floor(math.Max(0, seconds) + 0.5))
	return fmt.Sprintf("%02dh%02dm%02ds", total/3600, total%3600/60, total%60)
}

type pagesRule struct{ meta MetadataReader }

func (pagesRule) Enabled(o Options) bool { return o.ByPages }
func (pagesRule) Applies(f File) bool    { return pagedExtensions[f.Ext] }

// Segment interrompe só se o contexto for cancelado; qualquer outro erro vira "pages-unknown".
func (r pagesRule) Segment(ctx context.Context, f File) (string, error) {
	pages, err := r.meta.Pages(ctx, f.Path, f.Ext)
	if ctxErr := ctx.Err(); ctxErr != nil {
		return "", ctxErr
	}
	return segmentOrUnknown("pages-", fmt.Sprint(pages), err), nil
}

func setOf(values ...string) map[string]bool {
	set := make(map[string]bool, len(values))
	for _, v := range values {
		set[v] = true
	}
	return set
}

package criteria

import (
	"context"
	"fmt"
)

// ".doc" não tem leitor de páginas, mas continua aqui: vai para "pages-unknown" (#8).
var pagedExtensions = setOf("pdf", "doc", "docx", "odt")

type pagesRule struct{ meta MetadataReader }

func (pagesRule) Enabled(o Options) bool { return o.ByPages }
func (pagesRule) Applies(f File) bool    { return pagedExtensions[f.Ext] }

// Só interrompe se o contexto for cancelado; qualquer outro erro vira "pages-unknown" (#8).
func (r pagesRule) Segment(ctx context.Context, f File) (string, error) {
	pages, err := r.meta.Pages(ctx, f.Path, f.Ext)
	if ctxErr := ctx.Err(); ctxErr != nil {
		return "", ctxErr
	}
	return segmentOrUnknown("pages-", fmt.Sprint(pages), err), nil
}

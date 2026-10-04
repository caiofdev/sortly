package criteria

import (
	"context"
	"errors"
	"testing"

	"github.com/caiofdev/sortly/backend/metadata"
)

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

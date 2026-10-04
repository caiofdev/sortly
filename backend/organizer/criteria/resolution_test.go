package criteria

import (
	"context"
	"testing"

	"github.com/caiofdev/sortly/backend/metadata"
)

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

package criteria

import (
	"context"
	"errors"
	"testing"
)

func TestExtensionRule(t *testing.T) {
	got, err := extensionRule{}.Segment(context.Background(), File{Ext: "pdf"})
	if err != nil || got != "pdf" {
		t.Errorf("com extensão = (%q, %v)", got, err)
	}
	if _, err := (extensionRule{}).Segment(context.Background(), File{}); !errors.Is(err, ErrSkipNoExtension) {
		t.Errorf("sem extensão: err = %v, want ErrSkipNoExtension", err)
	}
}

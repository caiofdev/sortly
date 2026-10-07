package criteria

import (
	"context"
	"fmt"
)

const bytesPerMB = 1024 * 1024

type sizeRule struct{}

func (sizeRule) Enabled(o Options) bool { return o.BySize }
func (sizeRule) Applies(File) bool      { return true }

// Arredonda para cima em MB, com mínimo de 1: 0 B e 1 MB exato viram "size-1mb" (#8).
func (sizeRule) Segment(_ context.Context, f File) (string, error) {
	mb := max(1, (f.Size+bytesPerMB-1)/bytesPerMB)
	return fmt.Sprintf("size-%dmb", mb), nil
}

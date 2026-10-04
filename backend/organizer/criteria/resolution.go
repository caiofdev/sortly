package criteria

import (
	"context"
	"fmt"
)

var imageExtensions = setOf("png", "jpg", "jpeg", "gif", "webp", "bmp")

type resolutionRule struct{ meta MetadataReader }

func (resolutionRule) Enabled(o Options) bool { return o.ByResolution }
func (resolutionRule) Applies(f File) bool    { return imageExtensions[f.Ext] }

func (r resolutionRule) Segment(_ context.Context, f File) (string, error) {
	size, err := r.meta.Resolution(f.Path)
	return segmentOrUnknown("", fmt.Sprintf("%dx%d", size.Width, size.Height), err), nil
}

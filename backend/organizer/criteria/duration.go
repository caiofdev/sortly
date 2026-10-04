package criteria

import (
	"context"
	"fmt"
	"math"

	"github.com/caiofdev/sortly/backend/metadata"
)

var videoExtensions = setOf("mp4")

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

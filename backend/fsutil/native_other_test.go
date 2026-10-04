//go:build !windows

package fsutil

// função      | CC | casos
// nativePath  |  1 | TestNativePath

import "testing"

func TestNativePath(t *testing.T) {
	for _, p := range []string{"", "/tmp/arquivo.", "rel"} {
		if got := nativePath(p); got != p {
			t.Errorf("nativePath(%q) = %q, want inalterado", p, got)
		}
	}
}

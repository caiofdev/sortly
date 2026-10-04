//go:build !windows

package paths

// função      | CC | casos
// Native  |  1 | TestNativePath

import "testing"

func TestNativePath(t *testing.T) {
	for _, p := range []string{"", "/tmp/arquivo.", "rel"} {
		if got := Native(p); got != p {
			t.Errorf("Native(%q) = %q, want inalterado", p, got)
		}
	}
}

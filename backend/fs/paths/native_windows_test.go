//go:build windows

package paths

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNativePath(t *testing.T) {
	wd, _ := os.Getwd()
	cases := []struct{ in, want string }{
		{"", ""},
		{`\\?\C:\x`, `\\?\C:\x`},
		{`C:\Users\ana\arquivo.`, `\\?\C:\Users\ana\arquivo.`},
		{`\\servidor\share\a.txt`, `\\?\UNC\servidor\share\a.txt`},
		{"rel", `\\?\` + filepath.Join(wd, "rel")},
		{`C:\a\..\b\nota.`, `\\?\C:\b\nota.`},
	}
	for _, tc := range cases {
		if got := Native(tc.in); got != tc.want {
			t.Errorf("Native(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

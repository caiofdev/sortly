//go:build !windows

package paths

// Native não muda nada fora do Windows.
func Native(p string) string { return p }

//go:build !windows

package fsutil

// nativePath não muda nada fora do Windows.
func nativePath(p string) string { return p }

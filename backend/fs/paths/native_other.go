//go:build !windows

package paths

// Fora do Windows não há formato estendido (#8).
func Native(p string) string { return p }

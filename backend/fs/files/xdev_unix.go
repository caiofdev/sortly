//go:build !windows

package files

import "syscall"

// errCrossDevice é o erro do rename entre sistemas de arquivos diferentes.
var errCrossDevice error = syscall.EXDEV

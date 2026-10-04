//go:build !windows

package files

import "syscall"

var errCrossDevice error = syscall.EXDEV

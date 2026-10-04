//go:build windows

package fsutil

import "syscall"

// errCrossDevice é ERROR_NOT_SAME_DEVICE, devolvido pelo MoveFileEx ao mover
// entre unidades diferentes.
var errCrossDevice error = syscall.Errno(17)

//go:build windows

package files

import "syscall"

// ERROR_NOT_SAME_DEVICE, devolvido pelo MoveFileEx ao mover entre unidades (#5).
var errCrossDevice error = syscall.Errno(17)

// No Windows não é preciso: o NTFS registra a criação do arquivo no próprio
// journal, e o FlushFileBuffers numa pasta exige acesso de escrita que o
// os.Open não pede (#52).
func syncParentDir(string) error { return nil }

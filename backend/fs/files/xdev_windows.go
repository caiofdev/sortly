//go:build windows

package files

import "syscall"

// errCrossDevice é ERROR_NOT_SAME_DEVICE, devolvido pelo MoveFileEx ao mover
// entre unidades diferentes.
var errCrossDevice error = syscall.Errno(17)

// syncParentDir não faz nada no Windows: o NTFS registra a criação do arquivo
// no próprio journal, e o FlushFileBuffers numa pasta exige acesso de escrita
// que o os.Open não pede.
func syncParentDir(string) error { return nil }

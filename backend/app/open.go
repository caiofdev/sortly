package app

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/caiofdev/sortly/backend/apperr"
)

var ErrDestinationNotFound = apperr.New("DESTINATION_NOT_FOUND", "app: a pasta de destino não existe mais")

// O gerenciador de arquivos de cada sistema; nos demais (Linux), o xdg-open
// escolhe o do ambiente gráfico (#79).
var fileManagers = map[string]string{"windows": "explorer", "darwin": "open"}

type folderOpener struct {
	goos  string
	stat  func(string) (os.FileInfo, error)
	start func(name string, args ...string) error
}

func newFolderOpener() folderOpener {
	return folderOpener{goos: runtime.GOOS, stat: os.Stat, start: startDetached}
}

func (o folderOpener) open(path string) error {
	if info, err := o.stat(path); err != nil || !info.IsDir() {
		return fmt.Errorf("%w: %q", ErrDestinationNotFound, path)
	}
	name, ok := fileManagers[o.goos]
	if !ok {
		name = "xdg-open"
	}
	if err := o.start(name, path); err != nil {
		return fmt.Errorf("app: abrir %q com %s: %w", path, name, err)
	}
	return nil
}

// Sem esperar o código de saída: o explorer sai com 1 mesmo quando abre a pasta;
// o Wait só libera o processo (#79).
func startDetached(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() { _ = cmd.Wait() }()
	return nil
}

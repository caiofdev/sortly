package app

import (
	"os"
	"path/filepath"

	"github.com/caiofdev/sortly/backend/apperr"
)

var (
	ErrDroppedInvalid     = apperr.New("DROPPED_INVALID", "app: item arrastado inválido")
	ErrDroppedMissing     = apperr.New("DROPPED_MISSING", "app: item arrastado não existe mais")
	ErrDroppedUnsupported = apperr.New("DROPPED_UNSUPPORTED", "app: só arquivos e pastas podem ser arrastados")
)

// Um arquivo solto vale pela pasta que o contém (#12).
func resolveDroppedPath(path string) (string, error) {
	if path == "" {
		return "", ErrDroppedInvalid
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", ErrDroppedMissing
	}
	if info.IsDir() {
		return path, nil
	}
	if info.Mode().IsRegular() {
		return filepath.Dir(path), nil
	}
	return "", ErrDroppedUnsupported
}

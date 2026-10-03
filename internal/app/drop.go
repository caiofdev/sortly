package app

import (
	"os"
	"path/filepath"

	"github.com/caiofdev/sortly/internal/apperr"
)

// Erros de arrastar e soltar (ADR 0004).
var (
	ErrDroppedInvalid     = apperr.New("DROPPED_INVALID", "app: item arrastado inválido")
	ErrDroppedMissing     = apperr.New("DROPPED_MISSING", "app: item arrastado não existe mais")
	ErrDroppedUnsupported = apperr.New("DROPPED_UNSUPPORTED", "app: só arquivos e pastas podem ser arrastados")
)

// resolveDroppedPath devolve a pasta de origem para um item arrastado: a
// própria pasta, ou a pasta que contém o arquivo.
func resolveDroppedPath(path string) (string, error) {
	if path == "" {
		return "", ErrDroppedInvalid
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", ErrDroppedMissing
	}
	switch {
	case info.IsDir():
		return path, nil
	case info.Mode().IsRegular():
		return filepath.Dir(path), nil
	default:
		return "", ErrDroppedUnsupported
	}
}

// Package logging configura o log do app (log/slog) em arquivo, na pasta de
// configuração do usuário:
//
//	Windows: %AppData%\Sortly\logs\sortly.log
//	macOS:   ~/Library/Application Support/Sortly/logs/sortly.log
//	Linux:   ~/.config/Sortly/logs/sortly.log
package logging

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

// MaxSize é o tamanho a partir do qual o log atual vira sortly.log.1 ao abrir o app.
const MaxSize = 5 << 20

const fileName = "sortly.log"

// userConfigDir é substituível nos testes.
var userConfigDir = os.UserConfigDir

// DefaultDir devolve a pasta de logs do app.
func DefaultDir() (string, error) {
	base, err := userConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "Sortly", "logs"), nil
}

// Open abre (ou cria) o arquivo de log em dir. A função devolvida fecha o arquivo.
func Open(dir string) (*slog.Logger, func() error, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, nil, err
	}
	path := filepath.Join(dir, fileName)
	rotate(path, MaxSize)

	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, nil, err
	}
	return newLogger(f), f.Close, nil
}

// OpenDefault abre o log na pasta padrão. Se não for possível, registra no
// stderr: o app nunca deixa de abrir por causa do log.
func OpenDefault(stderr io.Writer) (*slog.Logger, func() error) {
	dir, err := DefaultDir()
	if err == nil {
		logger, closeFn, openErr := Open(dir)
		if openErr == nil {
			return logger, closeFn
		}
		err = openErr
	}
	logger := newLogger(stderr)
	logger.Warn("log em arquivo indisponível; usando stderr", "err", err)
	return logger, func() error { return nil }
}

func newLogger(w io.Writer) *slog.Logger {
	return slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{Level: slog.LevelInfo}))
}

// rotate guarda o log anterior como .1 quando ele atinge max bytes,
// substituindo um .1 mais antigo.
func rotate(path string, max int64) {
	info, err := os.Stat(path)
	if err != nil || info.Size() < max {
		return
	}
	_ = os.Rename(path, path+".1")
}

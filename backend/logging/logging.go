// Package logging grava o log do app (log/slog) num arquivo na pasta de
// configuração do usuário (#10):
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

// Ao abrir o app, o log atual vira sortly.log.1 a partir deste tamanho (#10).
const MaxSize = 5 << 20

const fileName = "sortly.log"

// O log registra caminhos de arquivos do usuário: pasta e arquivos são só do
// dono. No Windows, o modo não se aplica; as permissões vêm da pasta do perfil (#59).
const (
	dirMode  = 0o700
	fileMode = 0o600
)

// Substituível nos testes (#10).
var userConfigDir = os.UserConfigDir

func DefaultDir() (string, error) {
	base, err := userConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "Sortly", "logs"), nil
}

func Open(dir string) (*slog.Logger, func() error, error) {
	if err := os.MkdirAll(dir, dirMode); err != nil {
		return nil, nil, err
	}
	path := filepath.Join(dir, fileName)
	rotate(path, MaxSize)

	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, fileMode)
	if err != nil {
		return nil, nil, err
	}
	restrict(dir, path, path+".1")
	return newLogger(f), f.Close, nil
}

// Sem a pasta padrão, registra no stderr: o app nunca deixa de abrir por causa do log
// (#10).
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

// O modo do MkdirAll e do OpenFile só vale na criação: pasta e arquivos criados
// por versões anteriores (0o755/0o644) são corrigidos aqui. Uma falha (arquivo
// ausente, outro dono) não impede o log (#59).
func restrict(dir string, files ...string) {
	_ = os.Chmod(dir, dirMode)
	for _, f := range files {
		_ = os.Chmod(f, fileMode)
	}
}

// Substitui um .1 mais antigo (#10).
func rotate(path string, max int64) {
	info, err := os.Stat(path)
	if err != nil || info.Size() < max {
		return
	}
	_ = os.Rename(path, path+".1")
}

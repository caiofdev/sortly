// Package backup guarda os arquivos substituídos por uma organização em
// ~/.sortly/substituidos/<organização>/, para que o desfazer possa devolvê-los.
// Só apaga o que estiver dentro dessa raiz (ADR 0008, #82).
package backup

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/caiofdev/sortly/backend/fs/paths"
)

type Store struct {
	root      string
	now       func() time.Time
	removeAll func(string) error
	log       *slog.Logger
}

// Com log nil, nada é registrado (#82).
func New(root string, log *slog.Logger) *Store {
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	return &Store{root: root, now: time.Now, removeAll: os.RemoveAll, log: log}
}

// Uma pasta por organização, com o instante em nanossegundos no nome: duas
// organizações nunca dividem a pasta, e o desfazer de uma não mexe na outra.
// A pasta só é criada no disco quando o primeiro arquivo é guardado (#82).
func (s *Store) NewFolder() string {
	return filepath.Join(s.root, strconv.FormatInt(s.now().UnixNano(), 10))
}

// Um registro editado à mão poderia apontar para qualquer pasta; fora da raiz,
// nada é apagado nem lido como backup (#54, #82).
func (s *Store) Contains(path string) bool {
	return s.root != "" && paths.IsInside(path, s.root) && !paths.Equal(path, s.root)
}

// Pasta vazia ou fora da raiz não faz nada. Falhas vão só para o log: na pior
// hipótese, sobra uma pasta de backup em ~/.sortly (#82).
func (s *Store) Remove(folder string) {
	if folder == "" {
		return
	}
	if !s.Contains(folder) {
		s.log.Warn("pasta de backup fora da raiz; não apagada", "pasta", folder, "raiz", s.root)
		return
	}
	if err := s.removeAll(paths.Native(folder)); err != nil {
		s.log.Error("backup dos substituídos não apagado", "pasta", folder, "err", fmt.Errorf("backup: %w", err))
	}
}

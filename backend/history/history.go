// Package history guarda as últimas organizações em ~/.sortly/history.json, para
// a página Histórico. É separado do registro do desfazer (last-operation.json),
// que continua no formato da 1.0 e de um nível só (ADR 0007, #80).
package history

import (
	"encoding/json"
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/caiofdev/sortly/backend/fs/files"
)

const MaxEntries = 50

// Valores de Entry.Status (#80).
const (
	StatusDone     = "done"
	StatusUndone   = "undone"
	StatusCanceled = "canceled"
)

type Entry struct {
	At                    time.Time `json:"at"`
	SourceFolderPath      string    `json:"sourceFolderPath"`
	DestinationFolderPath string    `json:"destinationFolderPath"`
	MovedFiles            int       `json:"movedFiles"`
	Status                string    `json:"status"`
}

// Seguro para uso concorrente: os bindings do Wails rodam em goroutines separadas (#80).
type Service struct {
	mu      sync.Mutex
	path    string
	log     *slog.Logger
	entries []Entry
	loaded  bool
}

// Com log nil, nada é registrado (#80).
func New(path string, log *slog.Logger) *Service {
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	return &Service{path: path, log: log}
}

// Da mais recente para a mais antiga; sempre uma lista, nunca nil (#80).
func (s *Service) List() []Entry {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Entry{}, s.load()...)
}

// As que passam de MaxEntries saem (#80).
func (s *Service) Add(e Entry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	next := append([]Entry{e}, s.load()...)
	if len(next) > MaxEntries {
		next = next[:MaxEntries]
	}
	return s.save(next)
}

// O desfazer é de um nível: desfaz sempre a organização mais recente. Sem
// histórico (registro de antes da 2.x) ou com ela já desfeita, não muda nada (#80).
func (s *Service) MarkLastUndone() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	current := s.load()
	if len(current) == 0 || current[0].Status == StatusUndone {
		return nil
	}
	next := append([]Entry{}, current...)
	next[0].Status = StatusUndone
	return s.save(next)
}

// Lido só na primeira chamada. Arquivo ausente vale como vazio; ilegível ou
// corrompido também, com log: o histórico nunca impede o app de abrir (#80).
func (s *Service) load() []Entry {
	if s.loaded {
		return s.entries
	}
	s.loaded = true
	s.entries = []Entry{}
	data, err := os.ReadFile(s.path)
	if errors.Is(err, fs.ErrNotExist) {
		return s.entries
	}
	if err != nil {
		s.log.Warn("histórico ilegível; começando vazio", "path", s.path, "err", err)
		return s.entries
	}
	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		s.log.Warn("histórico corrompido; começando vazio", "path", s.path, "err", err)
		return s.entries
	}
	if entries != nil {
		s.entries = entries
	}
	return s.entries
}

// Se a gravação falhar, a lista em memória continua a anterior (#80).
func (s *Service) save(next []Entry) error {
	// Marshal de uma lista de structs com tipos simples não falha (#80).
	data, _ := json.Marshal(next)
	if err := files.WriteAtomic(s.path, data); err != nil {
		s.log.Error("histórico não salvo", "path", s.path, "err", err)
		return err
	}
	s.entries = next
	return nil
}

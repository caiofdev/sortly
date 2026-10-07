// Package store guarda o registro da última organização, usado pelo desfazer.
// O formato é o mesmo da versão 1.0, para que um desfazer pendente continue
// funcionando depois da atualização (#6).
package store

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/caiofdev/sortly/backend/fs/files"
)

// O arquivo existe, mas não é um registro válido; para o usuário, equivale a "nada para
// desfazer" (#6).
var ErrCorrupted = errors.New("store: registro da última operação corrompido")

type MovedItem struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// A ordem dos campos é a mesma do JSON gravado pela versão 1.0 (#6).
type Operation struct {
	SourceFolderPath      string      `json:"sourceFolderPath"`
	DestinationFolderPath string      `json:"destinationFolderPath"`
	MovedItems            []MovedItem `json:"movedItems"`
	// Ausente em registros antigos; nesse caso fica nil (#6, #57).
	CreatedFolders []string `json:"createdFolders"`
}

func (o *Operation) CanUndo() bool {
	return o != nil && len(o.MovedItems) > 0
}

// As interfaces que o consomem ficam nos pacotes de organização e desfazer (#6).
type FileStore struct {
	path string
	log  *slog.Logger
}

// O mesmo caminho da versão 1.0 (#6).
func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("store: pasta do usuário: %w", err)
	}
	return filepath.Join(home, ".sortly", "last-operation.json"), nil
}

// Com log nil, nada é registrado (#6).
func New(path string, log *slog.Logger) *FileStore {
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	return &FileStore{path: path, log: log}
}

func (s *FileStore) Path() string {
	return s.path
}

// Sem arquivo (ou com o JSON null), devolve (nil, nil); arquivo vazio ou inválido,
// ErrCorrupted (#6).
func (s *FileStore) Load() (*Operation, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, s.fail("ler", err)
	}

	var op *Operation
	if err := json.Unmarshal(data, &op); err != nil {
		s.log.Warn("registro da última operação ignorado", "path", s.path, "err", err)
		return nil, fmt.Errorf("%w: %w", ErrCorrupted, err)
	}
	return op, nil
}

// Com files.WriteAtomic, um crash no meio da gravação deixa o registro anterior
// intacto, nunca um JSON pela metade (#6).
func (s *FileStore) Save(op Operation) error {
	data, err := encode(op)
	if err != nil {
		return s.fail("codificar", err)
	}
	if err := files.WriteAtomic(s.path, data); err != nil {
		return s.fail("gravar", err)
	}
	return nil
}

// Não ter registro não é erro (#6).
func (s *FileStore) Clear() error {
	if err := os.Remove(s.path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return s.fail("apagar", err)
	}
	return nil
}

func (s *FileStore) fail(action string, err error) error {
	s.log.Error("registro da última operação", "acao", action, "path", s.path, "err", err)
	return fmt.Errorf("store: %s %s: %w", action, s.path, err)
}

// O mesmo JSON do JSON.stringify da versão 1.0: compacto, sem escapar &, < e >
// (o encoding/json escapa por padrão) e com listas vazias como [] em vez de null (#6).
func encode(op Operation) ([]byte, error) {
	if op.MovedItems == nil {
		op.MovedItems = []MovedItem{}
	}
	if op.CreatedFolders == nil {
		op.CreatedFolders = []string{}
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(op); err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(buf.Bytes(), []byte("\n")), nil
}

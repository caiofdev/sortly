// Package store guarda o registro da última organização, usado pelo desfazer.
//
// O formato do arquivo é o mesmo da versão Electron
// (electron/repositories/lastOperationRepository.js), para que um desfazer
// pendente continue funcionando depois da atualização.
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
)

// ErrCorrupted indica que o arquivo existe, mas não é um registro válido.
// Para o usuário, equivale a "nada para desfazer".
var ErrCorrupted = errors.New("store: registro da última operação corrompido")

// MovedItem é um arquivo movido pela organização.
type MovedItem struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// Operation é o registro da última organização. A ordem dos campos é a mesma
// do JSON gravado pela versão Electron.
type Operation struct {
	SourceFolderPath      string      `json:"sourceFolderPath"`
	DestinationFolderPath string      `json:"destinationFolderPath"`
	MovedItems            []MovedItem `json:"movedItems"`
	// Ausente em registros antigos; nesse caso fica nil.
	CreatedFolders []string `json:"createdFolders"`
}

// CanUndo informa se há arquivos movidos para desfazer.
func (o *Operation) CanUndo() bool {
	return o != nil && len(o.MovedItems) > 0
}

// FileStore grava o registro num arquivo JSON. As interfaces que o consomem
// ficam nos pacotes de organização e desfazer.
type FileStore struct {
	path string
	log  *slog.Logger
}

// Substituível nos testes para simular falha na troca atômica do arquivo.
var renameFile = os.Rename

// DefaultPath devolve ~/.sortly/last-operation.json, o mesmo caminho da versão Electron.
func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("store: pasta do usuário: %w", err)
	}
	return filepath.Join(home, ".sortly", "last-operation.json"), nil
}

// New cria um FileStore no caminho indicado. Com log nil, nada é registrado.
func New(path string, log *slog.Logger) *FileStore {
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	return &FileStore{path: path, log: log}
}

// Path devolve o caminho do arquivo.
func (s *FileStore) Path() string {
	return s.path
}

// Load lê o registro. Sem arquivo (ou com o JSON null), devolve (nil, nil).
// Arquivo vazio ou inválido devolve ErrCorrupted.
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

// Save grava o registro de forma atômica: escreve um arquivo temporário na
// mesma pasta e o renomeia por cima do atual. Um crash no meio da gravação
// deixa o registro anterior intacto, nunca um JSON pela metade.
func (s *FileStore) Save(op Operation) error {
	data, err := encode(op)
	if err != nil {
		return s.fail("codificar", err)
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return s.fail("criar pasta", err)
	}
	if err := s.writeAtomic(data); err != nil {
		return s.fail("gravar", err)
	}
	return nil
}

// Clear apaga o registro. Não ter registro não é erro.
func (s *FileStore) Clear() error {
	if err := os.Remove(s.path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return s.fail("apagar", err)
	}
	return nil
}

func (s *FileStore) writeAtomic(data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(s.path), ".last-operation-*.tmp")
	if err != nil {
		return err
	}
	_, writeErr := tmp.Write(data)
	if err := errors.Join(writeErr, tmp.Sync(), tmp.Close()); err != nil {
		_ = os.Remove(tmp.Name())
		return err
	}
	if err := renameFile(tmp.Name(), s.path); err != nil {
		_ = os.Remove(tmp.Name())
		return err
	}
	return nil
}

func (s *FileStore) fail(action string, err error) error {
	s.log.Error("registro da última operação", "acao", action, "path", s.path, "err", err)
	return fmt.Errorf("store: %s %s: %w", action, s.path, err)
}

// encode gera o mesmo JSON do JSON.stringify da versão Electron: compacto,
// sem escapar &, < e > (o encoding/json escapa por padrão) e com listas
// vazias como [] em vez de null.
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

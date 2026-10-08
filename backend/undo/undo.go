// Package undo desfaz a última organização: devolve cada arquivo ao lugar de
// origem, na ordem inversa, e remove as pastas criadas que ficaram vazias. Cada
// item do registro é um movimento (Command); desfazer é executar os movimentos
// inversos com o mesmo Executor da organização (#9).
package undo

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"slices"

	"github.com/caiofdev/sortly/backend/apperr"
	"github.com/caiofdev/sortly/backend/fs/files"
	"github.com/caiofdev/sortly/backend/fs/paths"

	"github.com/caiofdev/sortly/backend/organizer"
	"github.com/caiofdev/sortly/backend/store"
)

var ErrNothingToUndo = apperr.New("NOTHING_TO_UNDO", "undo: nenhuma organização para desfazer")

// Sem mensagem pronta: o frontend monta o texto (ADR 0004, #9).
type Result struct {
	RestoredFiles    int  `json:"restoredFiles"`
	RenamedOnRestore int  `json:"renamedOnRestore"`
	SkippedMissing   int  `json:"skippedMissing"`
	FailedFiles      int  `json:"failedFiles"`
	CanUndo          bool `json:"canUndo"`
}

type RecordStore interface {
	Load() (*store.Operation, error)
	Save(store.Operation) error
	Clear() error
}

// Precisa mover sem sobrescrever, como o *organizer.Executor (#9).
type Mover interface {
	Apply(ctx context.Context, plan organizer.Plan) (organizer.Outcome, error)
}

type Service struct {
	store     RecordStore
	mover     Mover
	exists    func(string) bool
	removeDir func(string) bool
	log       *slog.Logger
}

// Com log nil, nada é registrado (#9).
func NewService(st RecordStore, log *slog.Logger) *Service {
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	return &Service{
		store:     st,
		mover:     organizer.NewExecutor(log),
		exists:    files.Exists,
		removeDir: files.RemoveEmptyDir,
		log:       log,
	}
}

// Arquivos que não estão mais no destino são pulados (SkippedMissing). Se o
// lugar original estiver ocupado, o arquivo volta com " (n)" no nome
// (RenamedOnRestore). Arquivos que falharem continuam no registro, para que um
// novo desfazer tente de novo (CanUndo fica verdadeiro) (#9).
func (s *Service) Undo(ctx context.Context) (Result, error) {
	op, err := s.load()
	if err != nil {
		return Result{}, err
	}

	plan, skipped := s.inversePlan(op)
	out, applyErr := s.mover.Apply(ctx, plan)

	result := Result{
		RestoredFiles:    len(out.MovedItems),
		RenamedOnRestore: countRenamed(plan, out),
		SkippedMissing:   skipped,
		FailedFiles:      out.FailedFiles,
	}
	s.cleanup(op)
	result.CanUndo = s.updateRecord(op, remaining(op, plan, out))
	return result, applyErr
}

// Um registro corrompido conta como "nada para desfazer", como na versão 1.0 (#9).
func (s *Service) load() (*store.Operation, error) {
	op, err := s.store.Load()
	if err != nil && !errors.Is(err, store.ErrCorrupted) {
		return nil, fmt.Errorf("undo: %w", err)
	}
	if err != nil || !op.CanUndo() {
		return nil, ErrNothingToUndo
	}
	return op, nil
}

// Do último movimento para o primeiro. Conta como pulados os arquivos que não
// estão mais onde a organização os deixou e os que estão fora das pastas do
// registro (#9, #54).
func (s *Service) inversePlan(op *store.Operation) (organizer.Plan, int) {
	var plan organizer.Plan
	skipped := 0
	for i := len(op.MovedItems) - 1; i >= 0; i-- {
		item := op.MovedItems[i]
		if !s.restorable(op, item) {
			skipped++
			continue
		}
		plan.Moves = append(plan.Moves, organizer.Move{From: item.To, To: item.From})
	}
	return plan, skipped
}

// Um registro editado ou corrompido com caminhos fora das pastas registradas
// faria o desfazer mover arquivos para qualquer lugar; esses itens são pulados
// e vão para o log (#54).
func (s *Service) restorable(op *store.Operation, item store.MovedItem) bool {
	if !withinRecordedFolders(op, item) {
		s.log.Warn("item do registro fora das pastas registradas; ignorado",
			"origem", op.SourceFolderPath, "destino", rootOf(op), "from", item.From, "to", item.To)
		return false
	}
	return s.exists(item.To)
}

// from dentro da origem e to dentro do destino (destino vazio é a própria
// origem), sem ser a própria pasta (#54).
func withinRecordedFolders(op *store.Operation, item store.MovedItem) bool {
	return op.SourceFolderPath != "" &&
		strictlyInside(item.From, op.SourceFolderPath) &&
		strictlyInside(item.To, rootOf(op))
}

func strictlyInside(path, root string) bool {
	return paths.IsInside(path, root) && !paths.Equal(path, root)
}

// Destino vazio no registro é a própria origem (#54).
func rootOf(op *store.Operation) string {
	if op.DestinationFolderPath == "" {
		return op.SourceFolderPath
	}
	return op.DestinationFolderPath
}

// Voltou com outro nome quando o lugar original estava ocupado: o destino final
// difere do pedido no plano inverso (#9).
func countRenamed(plan organizer.Plan, out organizer.Outcome) int {
	requested := make(map[string]string, len(plan.Moves))
	for _, m := range plan.Moves {
		requested[m.From] = m.To
	}
	n := 0
	for _, moved := range out.MovedItems {
		if moved.To != requested[moved.From] {
			n++
		}
	}
	return n
}

// Os itens que não voltaram nem foram pulados: os que falharam ou não chegaram
// a ser tentados, com o contexto cancelado (#9).
func remaining(op *store.Operation, plan organizer.Plan, out organizer.Outcome) []store.MovedItem {
	restored := make(map[string]bool, len(out.MovedItems))
	for _, moved := range out.MovedItems {
		restored[moved.From] = true
	}
	inPlan := make(map[string]bool, len(plan.Moves))
	for _, m := range plan.Moves {
		inPlan[m.From] = true
	}

	var left []store.MovedItem
	for _, item := range op.MovedItems {
		if inPlan[item.To] && !restored[item.To] {
			left = append(left, item)
		}
	}
	return left
}

// Falhas aqui só vão para o log: na pior hipótese o registro antigo fica, e um
// novo desfazer pula o que já voltou (#9).
func (s *Service) updateRecord(op *store.Operation, left []store.MovedItem) bool {
	if len(left) == 0 {
		if err := s.store.Clear(); err != nil {
			s.log.Error("não foi possível apagar o registro após desfazer", "err", err)
		}
		return false
	}
	pending := *op
	pending.MovedItems = left
	if err := s.store.Save(pending); err != nil {
		s.log.Error("não foi possível regravar o registro com as falhas do desfazer", "err", err)
	}
	return true
}

// Só as pastas criadas pela organização, se ficaram vazias, da mais funda para a
// mais rasa; pastas que já existiam não estão na lista. Registros antigos, sem
// createdFolders, não dizem o que foi criado: sobe-se de cada pasta de arquivo
// até a raiz. Nada fora da raiz é tocado (#57).
func (s *Service) cleanup(op *store.Operation) {
	root := rootOf(op)
	if root == "" {
		return
	}
	if op.CreatedFolders == nil {
		for _, dir := range itemFolders(op) {
			s.removeEmptyAncestors(dir, root)
		}
		return
	}
	for _, dir := range deepestFirst(op.CreatedFolders) {
		if strictlyInside(dir, root) {
			s.removeDir(dir)
		}
	}
}

// Cada filha vem antes da mãe: o caminho da filha é sempre mais longo (#57).
func deepestFirst(dirs []string) []string {
	sorted := slices.Clone(dirs)
	slices.SortStableFunc(sorted, func(a, b string) int { return len(b) - len(a) })
	return sorted
}

// Para em root, sem removê-la. A comparação ignora maiúsculas no Windows e no
// macOS, então a raiz nunca é removida por diferença de caixa (#9, #58).
func (s *Service) removeEmptyAncestors(start, root string) {
	for dir := start; paths.IsInside(dir, root) && !paths.Equal(dir, root); dir = filepath.Dir(dir) {
		s.removeDir(dir)
	}
}

func itemFolders(op *store.Operation) []string {
	seen := map[string]bool{}
	var dirs []string
	for _, item := range op.MovedItems {
		if dir := filepath.Dir(item.To); !seen[dir] {
			seen[dir] = true
			dirs = append(dirs, dir)
		}
	}
	return dirs
}

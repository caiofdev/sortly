// Package undo desfaz a última organização: devolve cada arquivo ao lugar de
// origem, na ordem inversa, e remove as pastas que ficaram vazias.
//
// Cada item do registro é um movimento (Command); desfazer é executar os
// movimentos inversos com o mesmo Executor da organização.
package undo

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/caiofdev/sortly/backend/apperr"
	"github.com/caiofdev/sortly/backend/fsutil"
	"github.com/caiofdev/sortly/backend/organizer"
	"github.com/caiofdev/sortly/backend/store"
)

// ErrNothingToUndo indica que não há organização para desfazer (ADR 0004).
var ErrNothingToUndo = apperr.New("NOTHING_TO_UNDO", "undo: nenhuma organização para desfazer")

// Result resume o desfazer. Não traz mensagem pronta (o frontend monta o texto).
type Result struct {
	RestoredFiles    int  `json:"restoredFiles"`
	RenamedOnRestore int  `json:"renamedOnRestore"`
	SkippedMissing   int  `json:"skippedMissing"`
	FailedFiles      int  `json:"failedFiles"`
	CanUndo          bool `json:"canUndo"`
}

// RecordStore guarda o registro da última organização. É satisfeito por *store.FileStore.
type RecordStore interface {
	Load() (*store.Operation, error)
	Save(store.Operation) error
	Clear() error
}

// Mover executa movimentos sem sobrescrever. É satisfeito por *organizer.Executor.
type Mover interface {
	Apply(ctx context.Context, plan organizer.Plan) (organizer.Outcome, error)
}

// Service desfaz a última organização.
type Service struct {
	store     RecordStore
	mover     Mover
	exists    func(string) bool
	removeDir func(string) bool
	log       *slog.Logger
}

// NewService cria o serviço. Com log nil, nada é registrado.
func NewService(st RecordStore, log *slog.Logger) *Service {
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	return &Service{
		store:     st,
		mover:     organizer.NewExecutor(log),
		exists:    fsutil.Exists,
		removeDir: fsutil.RemoveEmptyDir,
		log:       log,
	}
}

// Undo devolve os arquivos da última organização ao lugar de origem.
//
// Arquivos que não estão mais no destino são pulados (SkippedMissing). Se o
// lugar original estiver ocupado, o arquivo volta com " (n)" no nome
// (RenamedOnRestore). Arquivos que falharem continuam no registro, para que
// um novo desfazer tente de novo (CanUndo fica verdadeiro).
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

// load devolve o registro ou ErrNothingToUndo. Um registro corrompido conta
// como "nada para desfazer", como na versão Electron.
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

// inversePlan monta os movimentos inversos, do último para o primeiro, e
// conta os arquivos que não estão mais onde a organização os deixou.
func (s *Service) inversePlan(op *store.Operation) (organizer.Plan, int) {
	var plan organizer.Plan
	skipped := 0
	for i := len(op.MovedItems) - 1; i >= 0; i-- {
		item := op.MovedItems[i]
		if !s.exists(item.To) {
			skipped++
			continue
		}
		plan.Moves = append(plan.Moves, organizer.Move{From: item.To, To: item.From})
	}
	return plan, skipped
}

// countRenamed conta os arquivos que voltaram com outro nome porque o lugar
// original estava ocupado: o destino final difere do pedido no plano inverso.
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

// remaining devolve os itens do registro original que não voltaram, nem foram
// pulados: os que falharam ou não chegaram a ser tentados (contexto cancelado).
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

// updateRecord apaga o registro quando tudo foi desfeito, ou o regrava só com
// o que falhou. Falhas aqui são só registradas no log: na pior hipótese o
// registro antigo fica, e um novo desfazer pula o que já voltou.
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

// cleanup remove as pastas que ficaram vazias, subindo de cada pasta usada
// pela organização até a raiz do destino (sem removê-la). Pastas com outros
// arquivos são preservadas.
func (s *Service) cleanup(op *store.Operation) {
	root := op.DestinationFolderPath
	if root == "" {
		root = op.SourceFolderPath
	}
	for _, dir := range cleanupCandidates(op) {
		if root == "" {
			s.removeDir(dir)
			continue
		}
		s.removeEmptyAncestors(dir, root)
	}
}

// removeEmptyAncestors sobe de start até root (exclusive). A comparação ignora
// maiúsculas no Windows, então a raiz nunca é removida por diferença de caixa.
func (s *Service) removeEmptyAncestors(start, root string) {
	for dir := start; fsutil.IsInside(dir, root) && !fsutil.PathsEqual(dir, root); dir = filepath.Dir(dir) {
		s.removeDir(dir)
	}
}

// cleanupCandidates junta as pastas registradas e as pastas dos arquivos
// movidos (registros antigos não têm createdFolders), sem repetir.
func cleanupCandidates(op *store.Operation) []string {
	seen := map[string]bool{}
	var dirs []string
	add := func(dir string) {
		if !seen[dir] {
			seen[dir] = true
			dirs = append(dirs, dir)
		}
	}
	for _, dir := range op.CreatedFolders {
		add(dir)
	}
	for _, item := range op.MovedItems {
		add(filepath.Dir(item.To))
	}
	return dirs
}

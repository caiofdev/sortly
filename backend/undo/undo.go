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
	"slices"

	"github.com/caiofdev/sortly/backend/apperr"
	"github.com/caiofdev/sortly/backend/fs/files"
	"github.com/caiofdev/sortly/backend/fs/paths"

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
		exists:    files.Exists,
		removeDir: files.RemoveEmptyDir,
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
// como "nada para desfazer", como na versão 1.0.
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
// conta os arquivos que não estão mais onde a organização os deixou ou que
// estão fora das pastas do registro.
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

// restorable informa se o item pode voltar. Um registro editado ou corrompido
// com caminhos fora das pastas registradas faria o desfazer mover arquivos para
// qualquer lugar; esses itens são pulados e vão para o log.
func (s *Service) restorable(op *store.Operation, item store.MovedItem) bool {
	if !withinRecordedFolders(op, item) {
		s.log.Warn("item do registro fora das pastas registradas; ignorado",
			"origem", op.SourceFolderPath, "destino", rootOf(op), "from", item.From, "to", item.To)
		return false
	}
	return s.exists(item.To)
}

// withinRecordedFolders exige from dentro da origem e to dentro do destino
// (destino vazio é a própria origem), sem ser a própria pasta.
func withinRecordedFolders(op *store.Operation, item store.MovedItem) bool {
	return op.SourceFolderPath != "" &&
		strictlyInside(item.From, op.SourceFolderPath) &&
		strictlyInside(item.To, rootOf(op))
}

func strictlyInside(path, root string) bool {
	return paths.IsInside(path, root) && !paths.Equal(path, root)
}

// rootOf devolve a pasta onde a organização criou as subpastas.
func rootOf(op *store.Operation) string {
	if op.DestinationFolderPath == "" {
		return op.SourceFolderPath
	}
	return op.DestinationFolderPath
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

// cleanup remove as pastas criadas pela organização que ficaram vazias, da
// mais funda para a mais rasa. Pastas que já existiam não estão na lista e
// ficam. Registros antigos, sem createdFolders, não dizem o que foi
// criado: sobe-se de cada pasta de arquivo até a raiz. Nada fora da raiz é
// tocado, e pastas com conteúdo nunca são removidas.
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

// deepestFirst ordena as pastas para que cada filha venha antes da mãe: o
// caminho da filha é sempre mais longo que o da mãe.
func deepestFirst(dirs []string) []string {
	sorted := slices.Clone(dirs)
	slices.SortStableFunc(sorted, func(a, b string) int { return len(b) - len(a) })
	return sorted
}

// removeEmptyAncestors sobe de start até root (exclusive). A comparação ignora
// maiúsculas no Windows e no macOS, então a raiz nunca é removida por
// diferença de caixa.
func (s *Service) removeEmptyAncestors(start, root string) {
	for dir := start; paths.IsInside(dir, root) && !paths.Equal(dir, root); dir = filepath.Dir(dir) {
		s.removeDir(dir)
	}
}

// itemFolders devolve as pastas onde os arquivos foram parar, sem repetir.
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

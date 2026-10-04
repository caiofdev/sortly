package organizer

import (
	"context"
	"log/slog"
	"path/filepath"

	"github.com/caiofdev/sortly/internal/fsutil"
	"github.com/caiofdev/sortly/internal/store"
)

// Outcome é o journal da execução: o que foi movido de fato.
type Outcome struct {
	MovedItems     []store.MovedItem
	CreatedFolders []string
	FailedFiles    int
}

// Executor aplica um plano, movendo um arquivo por vez sem sobrescrever nada.
type Executor struct {
	mkdirAll func(string) error
	move     func(src, dst string) (string, error)
	log      *slog.Logger
}

// NewExecutor cria o executor. Com log nil, nada é registrado.
func NewExecutor(log *slog.Logger) *Executor {
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	return &Executor{mkdirAll: fsutil.MkdirAll, move: fsutil.MoveUnique, log: log}
}

// Apply move os arquivos do plano. Uma falha em um arquivo não interrompe os
// demais: ela é registrada no log e contada em FailedFiles. Cada movimento
// concluído entra no journal, então tudo o que foi movido pode ser desfeito
// (B1). Se o contexto for cancelado, devolve o journal parcial e o erro.
func (e *Executor) Apply(ctx context.Context, plan Plan) (Outcome, error) {
	var out Outcome
	seen := map[string]bool{}

	for _, m := range plan.Moves {
		if err := ctx.Err(); err != nil {
			return out, err
		}
		final, err := e.moveOne(m)
		if err != nil {
			out.FailedFiles++
			e.log.Error("falha ao mover arquivo", "de", m.From, "para", m.To, "err", err)
			continue
		}
		out.MovedItems = append(out.MovedItems, store.MovedItem{From: m.From, To: final})
		if dir := filepath.Dir(final); !seen[dir] {
			seen[dir] = true
			out.CreatedFolders = append(out.CreatedFolders, dir)
		}
	}
	return out, nil
}

func (e *Executor) moveOne(m Move) (string, error) {
	if err := e.mkdirAll(filepath.Dir(m.To)); err != nil {
		return "", err
	}
	return e.move(m.From, m.To)
}

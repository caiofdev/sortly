package organizer

import (
	"context"
	"log/slog"
	"path/filepath"

	"github.com/caiofdev/sortly/backend/fs/files"
	"github.com/caiofdev/sortly/backend/store"
)

// O journal da execução: só o que foi movido de fato (#8).
type Outcome struct {
	MovedItems []store.MovedItem
	// Só as pastas que não existiam antes, da mais funda para a mais rasa em cada
	// movimento: o desfazer remove apenas essas (#57).
	CreatedFolders []string
	FailedFiles    int
}

// Um arquivo por vez, sem sobrescrever nada (#8).
type Executor struct {
	mkdirAll func(string) error
	exists   func(string) bool
	move     func(src, dst string) (string, error)
	log      *slog.Logger
}

// Com log nil, nada é registrado (#8).
func NewExecutor(log *slog.Logger) *Executor {
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	return &Executor{mkdirAll: files.MkdirAll, exists: files.Exists, move: files.MoveUnique, log: log}
}

// Uma falha num arquivo não interrompe os demais: vai para o log e conta em
// FailedFiles. Cada movimento concluído entra no journal, então tudo o que foi
// movido pode ser desfeito; com o contexto cancelado, devolve o journal parcial
// e o erro (#8).
func (e *Executor) Apply(ctx context.Context, plan Plan) (Outcome, error) {
	var out Outcome
	for _, m := range plan.Moves {
		if err := ctx.Err(); err != nil {
			return out, err
		}
		final, created, err := e.moveOne(m)
		out.CreatedFolders = append(out.CreatedFolders, created...)
		if err != nil {
			out.FailedFiles++
			e.log.Error("falha ao mover arquivo", "de", m.From, "para", m.To, "err", err)
			continue
		}
		out.MovedItems = append(out.MovedItems, store.MovedItem{From: m.From, To: final})
	}
	return out, nil
}

// As pastas criadas são devolvidas mesmo se o movimento falhar, para que o desfazer as
// limpe (#57).
func (e *Executor) moveOne(m Move) (final string, created []string, err error) {
	dir := filepath.Dir(m.To)
	missing := e.missingDirs(dir)
	if err := e.mkdirAll(dir); err != nil {
		return "", nil, err
	}
	final, err = e.move(m.From, m.To)
	return final, missing, err
}

// dir e as pastas acima dele que ainda não existem, da mais funda para a mais rasa (#57).
func (e *Executor) missingDirs(dir string) []string {
	var missing []string
	for d := dir; !e.exists(d); {
		missing = append(missing, d)
		parent := filepath.Dir(d)
		if parent == d {
			break
		}
		d = parent
	}
	return missing
}

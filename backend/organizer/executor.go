package organizer

import (
	"context"
	"log/slog"
	"path/filepath"

	"github.com/caiofdev/sortly/backend/fs/files"
	"github.com/caiofdev/sortly/backend/store"
)

// Outcome é o journal da execução: o que foi movido de fato.
type Outcome struct {
	MovedItems []store.MovedItem
	// CreatedFolders são só as pastas que não existiam antes, da mais funda
	// para a mais rasa em cada movimento: o desfazer remove apenas essas.
	CreatedFolders []string
	FailedFiles    int
}

// Executor aplica um plano, movendo um arquivo por vez sem sobrescrever nada.
type Executor struct {
	mkdirAll func(string) error
	exists   func(string) bool
	move     func(src, dst string) (string, error)
	log      *slog.Logger
}

// NewExecutor cria o executor. Com log nil, nada é registrado.
func NewExecutor(log *slog.Logger) *Executor {
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	return &Executor{mkdirAll: files.MkdirAll, exists: files.Exists, move: files.MoveUnique, log: log}
}

// Apply move os arquivos do plano. Uma falha em um arquivo não interrompe os
// demais: ela é registrada no log e contada em FailedFiles. Cada movimento
// concluído entra no journal, então tudo o que foi movido pode ser desfeito
// (B1). Se o contexto for cancelado, devolve o journal parcial e o erro.
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

// moveOne cria a pasta do destino e move o arquivo. As pastas criadas são
// devolvidas mesmo se o movimento falhar, para que o desfazer as limpe.
func (e *Executor) moveOne(m Move) (final string, created []string, err error) {
	dir := filepath.Dir(m.To)
	missing := e.missingDirs(dir)
	if err := e.mkdirAll(dir); err != nil {
		return "", nil, err
	}
	final, err = e.move(m.From, m.To)
	return final, missing, err
}

// missingDirs devolve dir e as pastas acima dele que ainda não existem, da
// mais funda para a mais rasa.
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

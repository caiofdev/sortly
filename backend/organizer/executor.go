package organizer

import (
	"context"
	"errors"
	"log/slog"
	"path/filepath"

	"github.com/caiofdev/sortly/backend/fs/files"
	"github.com/caiofdev/sortly/backend/fs/paths"
	"github.com/caiofdev/sortly/backend/store"
)

// Valores de Plan.Duplicates: o que fazer quando o destino já tem um arquivo com o
// mesmo nome. Renomear ("nome (1).ext") é o padrão e o comportamento da 1.0 (#82).
const (
	DuplicatesRename  = "rename"
	DuplicatesSkip    = "skip"
	DuplicatesReplace = "replace"
)

var errSkippedDuplicate = errors.New("organizer: arquivo com o mesmo nome no destino; ignorado")

// Strategy (ADR 0003): renomear não precisa de política, é o movimento de sempre (#82).
type duplicatePolicy func(e *Executor, m Move, plan Plan, out *Outcome) (string, error)

var duplicatePolicies = map[string]duplicatePolicy{
	DuplicatesSkip:    skipDuplicate,
	DuplicatesReplace: replaceDuplicate,
}

func ValidDuplicates(policy string) bool {
	_, ok := duplicatePolicies[policy]
	return ok || policy == DuplicatesRename
}

// O journal da execução: só o que foi movido de fato (#8).
type Outcome struct {
	MovedItems []store.MovedItem
	// Só as pastas que não existiam antes, da mais funda para a mais rasa em cada
	// movimento: o desfazer remove apenas essas (#57).
	CreatedFolders []string
	FailedFiles    int
	// Com "Ignorar", ficaram na origem; com "Substituir", o existente foi para o
	// backup antes (ADR 0008, #82).
	SkippedDuplicates int
	ReplacedItems     []store.ReplacedItem
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

// File é só o nome do arquivo; Folder é a pasta de destino relativa ao destino
// da organização ("" = a raiz). Done = Total marca o fim (#78). CategoryFolders diz
// se o 1º nível de Folder é uma categoria do Tipo, que a interface traduz (#84).
type Progress struct {
	Done            int    `json:"done"`
	Total           int    `json:"total"`
	File            string `json:"file"`
	Folder          string `json:"folder"`
	CategoryFolders bool   `json:"categoryFolders"`
}

// Chamado antes de cada movimento e uma vez no fim, na goroutine da organização (#78).
type Reporter func(Progress)

func (e *Executor) Apply(ctx context.Context, plan Plan) (Outcome, error) {
	return e.ApplyReporting(ctx, plan, func(Progress) {})
}

// Uma falha num arquivo não interrompe os demais: vai para o log e conta em
// FailedFiles. Cada movimento concluído entra no journal, então tudo o que foi
// movido pode ser desfeito; com o contexto cancelado (Cancelar na tela), devolve
// o journal parcial e o erro (#8, #78).
func (e *Executor) ApplyReporting(ctx context.Context, plan Plan, report Reporter) (Outcome, error) {
	var out Outcome
	total := len(plan.Moves)
	placed := map[string]bool{}
	for i, m := range plan.Moves {
		if err := ctx.Err(); err != nil {
			return out, err
		}
		report(Progress{
			Done: i, Total: total, File: filepath.Base(m.From),
			Folder: relativeFolder(plan.Destination, m.To), CategoryFolders: plan.CategoryFolders,
		})
		final, created, err := e.moveOne(m, plan, &out, placed)
		out.CreatedFolders = append(out.CreatedFolders, created...)
		if errors.Is(err, errSkippedDuplicate) {
			out.SkippedDuplicates++
			continue
		}
		if err != nil {
			out.FailedFiles++
			e.log.Error("falha ao mover arquivo", "de", m.From, "para", m.To, "err", err)
			continue
		}
		placed[paths.Key(final)] = true
		out.MovedItems = append(out.MovedItems, store.MovedItem{From: m.From, To: final})
	}
	report(Progress{Done: total, Total: total})
	return out, nil
}

// To sempre começa com o destino do plano (Plan.addMove), então o Rel não falha (#78).
func relativeFolder(dst, to string) string {
	rel, _ := filepath.Rel(dst, filepath.Dir(to))
	if rel == "." {
		return ""
	}
	return rel
}

// As pastas criadas são devolvidas mesmo se o movimento falhar, para que o desfazer as
// limpe (#57).
func (e *Executor) moveOne(m Move, plan Plan, out *Outcome, placed map[string]bool) (final string, created []string, err error) {
	dir := filepath.Dir(m.To)
	missing := e.missingDirs(dir)
	if err := e.mkdirAll(dir); err != nil {
		return "", nil, err
	}
	final, err = e.place(m, plan, out, placed)
	return final, missing, err
}

// Só conta como duplicado o arquivo que já estava no destino: um que esta mesma
// organização acabou de colocar ali (dois arquivos com o mesmo nome indo para a
// mesma pasta) é renomeado, para nunca ignorar nem substituir o próprio trabalho (#82).
func (e *Executor) place(m Move, plan Plan, out *Outcome, placed map[string]bool) (string, error) {
	policy, ok := duplicatePolicies[plan.Duplicates]
	if ok && e.exists(m.To) && !placed[paths.Key(m.To)] {
		return policy(e, m, plan, out)
	}
	return e.move(m.From, m.To)
}

func skipDuplicate(*Executor, Move, Plan, *Outcome) (string, error) {
	return "", errSkippedDuplicate
}

// O existente vai primeiro para o backup, na mesma posição relativa ao destino;
// se o novo não puder ocupar o lugar, o existente volta, e nada se perde. Sem
// pasta de backup, renomeia (#82).
func replaceDuplicate(e *Executor, m Move, plan Plan, out *Outcome) (string, error) {
	if plan.BackupFolder == "" {
		return e.move(m.From, m.To)
	}
	rel, _ := filepath.Rel(plan.Destination, m.To)
	target := filepath.Join(plan.BackupFolder, rel)
	if err := e.mkdirAll(filepath.Dir(target)); err != nil {
		return "", err
	}
	backup, err := e.move(m.To, target)
	if err != nil {
		return "", err
	}
	final, err := e.move(m.From, m.To)
	if err != nil {
		if _, back := e.move(backup, m.To); back != nil {
			e.log.Error("substituído não voltou do backup", "backup", backup, "para", m.To, "err", back)
		}
		return "", err
	}
	out.ReplacedItems = append(out.ReplacedItems, store.ReplacedItem{Path: final, Backup: backup})
	return final, nil
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

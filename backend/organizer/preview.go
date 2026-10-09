package organizer

import (
	"context"
	"path/filepath"
	"sort"
	"strings"
)

// Sete, para as sete categorias do critério Tipo caberem sempre, como no
// protótipo; o resto vira um chip "demais pastas" (#77, #81).
const MaxPreviewFolders = 7

// Name vazio são os arquivos que vão para a raiz do destino, sem subpasta (#77).
type FolderCount struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type Preview struct {
	TotalFiles      int           `json:"totalFiles"`
	Folders         []FolderCount `json:"folders"`
	OtherFiles      int           `json:"otherFiles"`
	CategoryFolders bool          `json:"categoryFolders"`
}

// Planeja como o Organize, sem mover nada: a prévia mostra exatamente as pastas
// que a organização criaria. Os critérios de metadados leem os arquivos, então
// numa pasta grande isto demora e respeita o ctx (#77).
func (s *Service) Preview(ctx context.Context, req Request) (Preview, error) {
	src, dst, opts, err := validate(req)
	if err != nil {
		return Preview{}, err
	}
	plan, err := s.planner.Plan(ctx, src, dst, opts)
	if err != nil {
		return Preview{}, err
	}
	return summarize(plan), nil
}

func summarize(plan Plan) Preview {
	targets := make([]string, 0, len(plan.Moves))
	for _, m := range plan.Moves {
		targets = append(targets, m.To)
	}
	folders, others := countFolders(plan.Destination, targets)
	return Preview{TotalFiles: plan.ProcessedFiles, Folders: folders, OtherFiles: others, CategoryFolders: plan.CategoryFolders}
}

// Agrupa pela pasta de 1º nível, da com mais arquivos para a com menos; no empate,
// pelo nome, para a ordem não mudar entre uma prévia e outra. Passando de
// MaxPreviewFolders, o resto soma em others. A prévia e a tela Concluído usam o
// mesmo resumo (#77, #79).
func countFolders(dst string, targets []string) (folders []FolderCount, others int) {
	counts := map[string]int{}
	for _, to := range targets {
		counts[firstFolder(dst, to)]++
	}
	folders = make([]FolderCount, 0, len(counts))
	for name, n := range counts {
		folders = append(folders, FolderCount{Name: name, Count: n})
	}
	sort.Slice(folders, func(i, j int) bool {
		if folders[i].Count != folders[j].Count {
			return folders[i].Count > folders[j].Count
		}
		return folders[i].Name < folders[j].Name
	})

	if len(folders) <= MaxPreviewFolders {
		return folders, 0
	}
	for _, f := range folders[MaxPreviewFolders:] {
		others += f.Count
	}
	return folders[:MaxPreviewFolders], others
}

// To sempre começa com o destino do plano (Plan.addMove), então o Rel não falha (#77).
func firstFolder(dst, to string) string {
	rel, _ := filepath.Rel(dst, to)
	folder, _, nested := strings.Cut(rel, string(filepath.Separator))
	if !nested {
		return ""
	}
	return folder
}

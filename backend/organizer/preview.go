package organizer

import (
	"context"
	"path/filepath"
	"sort"
	"strings"
)

// Como no protótipo: até 6 pastas e um chip "outras" com o resto (#77).
const MaxPreviewFolders = 6

// Name vazio são os arquivos que vão para a raiz do destino, sem subpasta (#77).
type FolderCount struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type Preview struct {
	TotalFiles int           `json:"totalFiles"`
	Folders    []FolderCount `json:"folders"`
	OtherFiles int           `json:"otherFiles"`
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

// Da pasta com mais arquivos para a com menos; no empate, pelo nome, para a
// ordem não mudar entre uma prévia e outra (#77).
func summarize(plan Plan) Preview {
	counts := map[string]int{}
	for _, m := range plan.Moves {
		counts[firstFolder(plan.Destination, m.To)]++
	}
	folders := make([]FolderCount, 0, len(counts))
	for name, n := range counts {
		folders = append(folders, FolderCount{Name: name, Count: n})
	}
	sort.Slice(folders, func(i, j int) bool {
		if folders[i].Count != folders[j].Count {
			return folders[i].Count > folders[j].Count
		}
		return folders[i].Name < folders[j].Name
	})

	p := Preview{TotalFiles: plan.ProcessedFiles, Folders: folders}
	if len(folders) > MaxPreviewFolders {
		for _, f := range folders[MaxPreviewFolders:] {
			p.OtherFiles += f.Count
		}
		p.Folders = folders[:MaxPreviewFolders]
	}
	return p
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

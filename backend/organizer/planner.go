package organizer

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/caiofdev/sortly/backend/fs/paths"
	"github.com/caiofdev/sortly/backend/organizer/criteria"
)

// To é o destino desejado; o nome final ganha " (n)" se estiver ocupado (#8).
type Move struct {
	From string
	To   string
}

type Plan struct {
	Source                  string
	Destination             string
	Moves                   []Move
	ProcessedFiles          int
	IgnoredWithoutExtension int
	IgnoredFolders          int
	// Arquivos que já estão no lugar certo (destino = origem e nenhum critério
	// gerou subpasta): não são movidos nem renomeados (#8).
	UnchangedFiles int
	// Com o critério Tipo, as pastas de 1º nível são categorias (images,
	// documents…), que a interface traduz (#81).
	CategoryFolders bool
	// Política de duplicados e, com "Substituir", a pasta do backup desta
	// organização; o Service preenche depois de planejar (#82).
	Duplicates   string
	BackupFolder string
}

// Não altera nada no disco (#8).
type Planner struct {
	rules   []criteria.Rule
	readDir func(string) ([]fs.DirEntry, error)
}

func NewPlanner(rules []criteria.Rule) *Planner {
	return &Planner{rules: rules, readDir: os.ReadDir}
}

// Lê só o nível superior de src. Subpastas são ignoradas e contadas; itens que
// não são arquivo regular (como atalhos simbólicos) são ignorados sem contar (#8).
func (p *Planner) Plan(ctx context.Context, src, dst string, opts criteria.Options) (Plan, error) {
	entries, err := p.readDir(src)
	if err != nil {
		return Plan{}, fmt.Errorf("%w: %w", ErrInvalidSource, err)
	}

	rules := p.active(opts)
	plan := Plan{Source: src, Destination: dst, CategoryFolders: opts.ByType}
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return Plan{}, err
		}
		if err := plan.add(ctx, entry, rules); err != nil {
			return Plan{}, err
		}
	}
	return plan, nil
}

// Com "Incluir subpastas": os arquivos de qualquer nível vão para as mesmas
// pastas de critério no destino. Não segue links (atalhos e junções não são
// pastas para o ReadDir), pula a pasta de destino quando ela fica dentro da
// origem e conta como ignorada uma subpasta que não pôde ser lida (#83).
func (p *Planner) PlanTree(ctx context.Context, src, dst string, opts criteria.Options) (Plan, error) {
	entries, err := p.readDir(src)
	if err != nil {
		return Plan{}, fmt.Errorf("%w: %w", ErrInvalidSource, err)
	}
	plan := Plan{Source: src, Destination: dst, CategoryFolders: opts.ByType}
	if err := p.walk(ctx, src, entries, &plan, p.active(opts)); err != nil {
		return Plan{}, err
	}
	return plan, nil
}

func (p *Planner) walk(ctx context.Context, dir string, entries []fs.DirEntry, plan *Plan, rules []criteria.Rule) error {
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := p.visit(ctx, dir, entry, plan, rules); err != nil {
			return err
		}
	}
	return nil
}

func (p *Planner) visit(ctx context.Context, dir string, entry fs.DirEntry, plan *Plan, rules []criteria.Rule) error {
	if !entry.IsDir() {
		return plan.addFile(ctx, dir, entry, rules)
	}
	child := filepath.Join(dir, entry.Name())
	if hiddenFolder(entry.Name()) || paths.Equal(child, plan.Destination) {
		plan.IgnoredFolders++
		return nil
	}
	entries, err := p.readDir(child)
	if err != nil {
		plan.IgnoredFolders++
		return nil
	}
	return p.walk(ctx, child, entries, plan, rules)
}

// Pastas ocultas guardam o funcionamento de outros programas: espalhar o
// conteúdo de um .git corromperia o repositório, e o .sortly tem o registro do
// próprio desfazer (#83).
func hiddenFolder(name string) bool {
	return strings.HasPrefix(name, ".")
}

func (p *Planner) active(opts criteria.Options) []criteria.Rule {
	var rules []criteria.Rule
	for _, r := range p.rules {
		if r.Enabled(opts) {
			rules = append(rules, r)
		}
	}
	return rules
}

func (plan *Plan) add(ctx context.Context, entry fs.DirEntry, rules []criteria.Rule) error {
	if entry.IsDir() {
		plan.IgnoredFolders++
		return nil
	}
	return plan.addFile(ctx, plan.Source, entry, rules)
}

func (plan *Plan) addFile(ctx context.Context, dir string, entry fs.DirEntry, rules []criteria.Rule) error {
	if !entry.Type().IsRegular() {
		return nil
	}
	plan.ProcessedFiles++

	info, err := entry.Info()
	if err != nil {
		// Sumiu entre a listagem e a leitura: segue com os demais (#8).
		return nil
	}
	file := criteria.NewFile(dir, info)

	segments, err := segmentsFor(ctx, rules, file)
	if errors.Is(err, criteria.ErrSkipNoExtension) {
		plan.IgnoredWithoutExtension++
		return nil
	}
	if err != nil {
		return err
	}
	plan.addMove(file, segments)
	return nil
}

func (plan *Plan) addMove(file criteria.File, segments []string) {
	parts := append(append([]string{plan.Destination}, segments...), file.Name)
	to := filepath.Join(parts...)
	if paths.Equal(file.Path, to) {
		plan.UnchangedFiles++
		return
	}
	plan.Moves = append(plan.Moves, Move{From: file.Path, To: to})
}

func segmentsFor(ctx context.Context, rules []criteria.Rule, f criteria.File) ([]string, error) {
	segments := make([]string, 0, len(rules))
	for _, r := range rules {
		if !r.Applies(f) {
			continue
		}
		segment, err := r.Segment(ctx, f)
		if err != nil {
			return nil, err
		}
		segments = append(segments, segment)
	}
	return segments, nil
}

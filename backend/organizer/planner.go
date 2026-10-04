package organizer

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/caiofdev/sortly/backend/fs/paths"
)

// Move é um arquivo a mover. To é o destino desejado; o nome final pode
// ganhar um sufixo " (n)" se já estiver ocupado.
type Move struct {
	From string
	To   string
}

// Plan é o resultado do planejamento: o que mover e o que foi ignorado.
type Plan struct {
	Source                  string
	Destination             string
	Moves                   []Move
	ProcessedFiles          int
	IgnoredWithoutExtension int
	IgnoredFolders          int
	// UnchangedFiles: arquivos que já estão no lugar certo (destino = origem e
	// nenhum critério gerou subpasta). Não são movidos nem renomeados.
	UnchangedFiles int
}

// Planner decide para onde cada arquivo vai, sem alterar nada no disco.
type Planner struct {
	rules   []SegmentRule
	readDir func(string) ([]fs.DirEntry, error)
}

// NewPlanner cria o planejador com o registry de regras dado (veja NewRules).
func NewPlanner(rules []SegmentRule) *Planner {
	return &Planner{rules: rules, readDir: os.ReadDir}
}

// Plan lê só o nível superior de src e calcula os movimentos para dst.
// Subpastas são ignoradas e contadas; itens que não são arquivo regular
// (como atalhos simbólicos) são ignorados sem contar.
func (p *Planner) Plan(ctx context.Context, src, dst string, opts Options) (Plan, error) {
	entries, err := p.readDir(src)
	if err != nil {
		return Plan{}, fmt.Errorf("%w: %w", ErrInvalidSource, err)
	}

	rules := p.active(opts)
	plan := Plan{Source: src, Destination: dst}
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

func (p *Planner) active(opts Options) []SegmentRule {
	var rules []SegmentRule
	for _, r := range p.rules {
		if r.Enabled(opts) {
			rules = append(rules, r)
		}
	}
	return rules
}

func (plan *Plan) add(ctx context.Context, entry fs.DirEntry, rules []SegmentRule) error {
	if entry.IsDir() {
		plan.IgnoredFolders++
		return nil
	}
	if !entry.Type().IsRegular() {
		return nil
	}
	plan.ProcessedFiles++

	info, err := entry.Info()
	if err != nil {
		// Sumiu entre a listagem e a leitura: segue com os demais, como antes.
		return nil
	}
	file := newFile(plan.Source, info)

	segments, err := segmentsFor(ctx, rules, file)
	if errors.Is(err, ErrSkipNoExtension) {
		plan.IgnoredWithoutExtension++
		return nil
	}
	if err != nil {
		return err
	}
	plan.addMove(file, segments)
	return nil
}

func (plan *Plan) addMove(file File, segments []string) {
	parts := append(append([]string{plan.Destination}, segments...), file.Name)
	to := filepath.Join(parts...)
	if paths.Equal(file.Path, to) {
		plan.UnchangedFiles++
		return
	}
	plan.Moves = append(plan.Moves, Move{From: file.Path, To: to})
}

func newFile(dir string, info fs.FileInfo) File {
	return File{
		Name:    info.Name(),
		Path:    filepath.Join(dir, info.Name()),
		Ext:     fileExt(info.Name()),
		Size:    info.Size(),
		ModTime: info.ModTime(),
	}
}

func segmentsFor(ctx context.Context, rules []SegmentRule, f File) ([]string, error) {
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

// Package organizer organiza os arquivos de uma pasta em subpastas, segundo os
// critérios escolhidos: valida o pedido, planeja (Planner), executa (Executor)
// e grava o registro para desfazer. As regras estão em
// docs/organization-rules.md (#8).
package organizer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/caiofdev/sortly/backend/organizer/criteria"
	"github.com/caiofdev/sortly/backend/store"
)

// Options vem das preferências salvas (#44).
type Request struct {
	SourceFolderPath      string
	DestinationFolderPath string
	Options               criteria.Options
}

// Sem mensagem pronta: o frontend monta o texto no idioma do usuário (ADR 0004, #8).
type Result struct {
	SourceFolderPath        string `json:"sourceFolderPath"`
	DestinationFolderPath   string `json:"destinationFolderPath"`
	ProcessedFiles          int    `json:"processedFiles"`
	MovedFiles              int    `json:"movedFiles"`
	FailedFiles             int    `json:"failedFiles"`
	UnchangedFiles          int    `json:"unchangedFiles"`
	IgnoredWithoutExtension int    `json:"ignoredWithoutExtension"`
	IgnoredFolders          int    `json:"ignoredFolders"`
	CanUndo                 bool   `json:"canUndo"`
}

type RecordStore interface {
	Load() (*store.Operation, error)
	Save(store.Operation) error
	Clear() error
}

type Deps struct {
	Metadata criteria.MetadataReader
	Store    RecordStore
	Location *time.Location // fuso das pastas por data; nil = fuso local (#8)
	Logger   *slog.Logger
}

type Service struct {
	planner  *Planner
	executor *Executor
	store    RecordStore
}

func NewService(d Deps) *Service {
	loc := d.Location
	if loc == nil {
		loc = time.Local
	}
	return &Service{
		planner:  NewPlanner(criteria.New(d.Metadata, loc)),
		executor: NewExecutor(d.Logger),
		store:    d.Store,
	}
}

// Mesmo com erro (falha ao gravar o registro, contexto cancelado), o Result
// descreve o que foi feito, porque arquivos podem ter sido movidos (#8).
func (s *Service) Organize(ctx context.Context, req Request) (Result, error) {
	src, dst, opts, err := validate(req)
	if err != nil {
		return Result{}, err
	}
	plan, err := s.planner.Plan(ctx, src, dst, opts)
	if err != nil {
		return Result{}, err
	}

	out, applyErr := s.executor.Apply(ctx, plan)
	saveErr := s.record(plan, out)

	result := Result{
		SourceFolderPath:        src,
		DestinationFolderPath:   dst,
		ProcessedFiles:          plan.ProcessedFiles,
		MovedFiles:              len(out.MovedItems),
		FailedFiles:             out.FailedFiles,
		UnchangedFiles:          plan.UnchangedFiles,
		IgnoredWithoutExtension: plan.IgnoredWithoutExtension,
		IgnoredFolders:          plan.IgnoredFolders,
		CanUndo:                 s.canUndo(out, saveErr),
	}
	return result, errors.Join(applyErr, saveErr)
}

// Só grava se algo foi movido: uma organização sem movimentos preserva o desfazer da
// anterior (#8).
func (s *Service) record(plan Plan, out Outcome) error {
	if len(out.MovedItems) == 0 {
		return nil
	}
	op := store.Operation{
		SourceFolderPath:      plan.Source,
		DestinationFolderPath: plan.Destination,
		MovedItems:            out.MovedItems,
		CreatedFolders:        out.CreatedFolders,
	}
	if err := s.store.Save(op); err != nil {
		// O registro que ficou é da organização anterior: se sobrevivesse, o próximo
		// desfazer (inclusive após reabrir o app) desfaria a errada. A falha ao apagar
		// já vai para o log do store (#53).
		_ = s.store.Clear()
		return fmt.Errorf("%w: %w", ErrRecordNotSaved, err)
	}
	return nil
}

func (s *Service) canUndo(out Outcome, saveErr error) bool {
	if len(out.MovedItems) > 0 {
		return saveErr == nil
	}
	previous, err := s.store.Load()
	return err == nil && previous.CanUndo()
}

// A ordem (origem, destino, critérios) decide qual erro o usuário vê quando há
// mais de um problema. Destino vazio usa a própria origem; destino inexistente
// será criado (#8).
func validate(req Request) (src, dst string, opts criteria.Options, err error) {
	src = req.SourceFolderPath
	if src == "" || !isDir(src) {
		return "", "", criteria.Options{}, fmt.Errorf("%w: %q", ErrInvalidSource, src)
	}
	dst = req.DestinationFolderPath
	if dst == "" {
		dst = src
	}
	if info, statErr := os.Stat(dst); statErr == nil && !info.IsDir() {
		return "", "", criteria.Options{}, fmt.Errorf("%w: %q", ErrInvalidDestination, dst)
	}
	opts = req.Options
	if !opts.Any() {
		return "", "", criteria.Options{}, ErrNoCriteria
	}
	return src, dst, opts, nil
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

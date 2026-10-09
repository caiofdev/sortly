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

// Options e Duplicates vêm das preferências salvas (#44, #82).
type Request struct {
	SourceFolderPath      string
	DestinationFolderPath string
	Options               criteria.Options
	Duplicates            string
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
	Canceled                bool   `json:"canceled"`
	// Só o que foi movido de fato, para a tela Concluído (#79).
	Folders         []FolderCount `json:"folders"`
	OtherFiles      int           `json:"otherFiles"`
	CategoryFolders bool          `json:"categoryFolders"`
	// Com "Ignorar" e "Substituir" (#82).
	SkippedDuplicates int `json:"skippedDuplicates"`
	ReplacedFiles     int `json:"replacedFiles"`
}

// Onde ficam os substituídos de cada organização (#82).
type BackupStore interface {
	NewFolder() string
	Remove(folder string)
}

// Sem backup configurado, "Substituir" renomeia: nunca sobrescreve sem ter como
// devolver (#82).
type noBackups struct{}

func (noBackups) NewFolder() string { return "" }
func (noBackups) Remove(string)     {}

type RecordStore interface {
	Load() (*store.Operation, error)
	Save(store.Operation) error
	Clear() error
}

type Deps struct {
	Metadata criteria.MetadataReader
	Store    RecordStore
	Location *time.Location // fuso das pastas por data; nil = fuso local (#8)
	Backups  BackupStore    // nil = sem backup; "Substituir" renomeia (#82)
	Logger   *slog.Logger
}

type Service struct {
	planner  *Planner
	executor *Executor
	store    RecordStore
	backups  BackupStore
	log      *slog.Logger
}

func NewService(d Deps) *Service {
	loc := d.Location
	if loc == nil {
		loc = time.Local
	}
	backups := d.Backups
	if backups == nil {
		backups = noBackups{}
	}
	return &Service{
		planner:  NewPlanner(criteria.New(d.Metadata, loc)),
		executor: NewExecutor(d.Logger),
		store:    d.Store,
		backups:  backups,
		log:      logger(d.Logger),
	}
}

func (s *Service) Organize(ctx context.Context, req Request) (Result, error) {
	return s.OrganizeReporting(ctx, req, func(Progress) {})
}

// Mesmo com erro (falha ao gravar o registro, contexto cancelado), o Result
// descreve o que foi feito, porque arquivos podem ter sido movidos. Cancelado
// no planejamento, nada foi movido (#8, #78).
func (s *Service) OrganizeReporting(ctx context.Context, req Request, report Reporter) (Result, error) {
	src, dst, opts, err := validate(req)
	if err != nil {
		return Result{}, err
	}
	plan, err := s.planner.Plan(ctx, src, dst, opts)
	if err != nil {
		canceled := errors.Is(err, context.Canceled)
		return Result{SourceFolderPath: src, DestinationFolderPath: dst, Canceled: canceled, Folders: []FolderCount{}}, err
	}

	s.prepare(&plan, req.Duplicates)
	out, applyErr := s.executor.ApplyReporting(ctx, plan, report)
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
		Canceled:                errors.Is(applyErr, context.Canceled),
	}
	result.Folders, result.OtherFiles = movedFolders(dst, out.MovedItems)
	result.CategoryFolders = plan.CategoryFolders
	result.SkippedDuplicates = out.SkippedDuplicates
	result.ReplacedFiles = len(out.ReplacedItems)
	return result, errors.Join(applyErr, saveErr)
}

// Política desconhecida (preferência de uma versão futura) vale como renomear (#82).
func (s *Service) prepare(plan *Plan, duplicates string) {
	if !ValidDuplicates(duplicates) {
		duplicates = DuplicatesRename
	}
	plan.Duplicates = duplicates
	if duplicates == DuplicatesReplace {
		plan.BackupFolder = s.backups.NewFolder()
	}
}

func logger(l *slog.Logger) *slog.Logger {
	if l == nil {
		return slog.New(slog.DiscardHandler)
	}
	return l
}

func movedFolders(dst string, moved []store.MovedItem) ([]FolderCount, int) {
	targets := make([]string, 0, len(moved))
	for _, item := range moved {
		targets = append(targets, item.To)
	}
	return countFolders(dst, targets)
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
	if len(out.ReplacedItems) > 0 {
		op.ReplacedItems = out.ReplacedItems
		op.BackupFolder = plan.BackupFolder
	}
	previous := s.previousBackup()
	if err := s.store.Save(op); err != nil {
		// O registro que ficou é da organização anterior: se sobrevivesse, o próximo
		// desfazer (inclusive após reabrir o app) desfaria a errada. A falha ao apagar
		// já vai para o log do store (#53).
		_ = s.store.Clear()
		s.backups.Remove(previous)
		// Sem registro, o desfazer não devolve os substituídos desta organização; eles
		// continuam no backup, e o log diz onde (ADR 0008, #82).
		if op.BackupFolder != "" {
			s.log.Error("substituídos sem registro para desfazer", "pasta", op.BackupFolder, "arquivos", len(op.ReplacedItems))
		}
		return fmt.Errorf("%w: %w", ErrRecordNotSaved, err)
	}
	s.backups.Remove(previous)
	return nil
}

// O backup só serve ao desfazer do registro que o criou: quando esse registro é
// substituído ou apagado, os substituídos dele não têm mais como voltar (ADR 0008, #82).
func (s *Service) previousBackup() string {
	previous, err := s.store.Load()
	if err != nil || previous == nil {
		return ""
	}
	return previous.BackupFolder
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

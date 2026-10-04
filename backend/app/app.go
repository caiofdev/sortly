// Package app contém a fachada exposta ao frontend pelo Wails e a
// configuração da janela. A fachada só delega aos serviços de backend/ e
// converte os erros em códigos (ADR 0004); a lógica de negócio fica nos pacotes.
package app

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/caiofdev/sortly/backend/apperr"
	"github.com/caiofdev/sortly/backend/organizer"
	"github.com/caiofdev/sortly/backend/store"
	"github.com/caiofdev/sortly/backend/undo"
)

// Títulos dos seletores de pasta (os mesmos da versão Electron).
const (
	SourceDialogTitle      = "Selecione a pasta com os arquivos para organizar"
	DestinationDialogTitle = "Selecione a pasta de destino para receber os arquivos organizados"
)

// Organizer organiza uma pasta. É satisfeito por *organizer.Service.
type Organizer interface {
	Organize(ctx context.Context, req organizer.Request) (organizer.Result, error)
}

// Undoer desfaz a última organização. É satisfeito por *undo.Service.
type Undoer interface {
	Undo(ctx context.Context) (undo.Result, error)
}

// RecordReader lê o registro da última organização. É satisfeito por *store.FileStore.
type RecordReader interface {
	Load() (*store.Operation, error)
}

// DirectoryPicker abre o seletor de pasta e devolve o caminho escolhido ("" se cancelado).
type DirectoryPicker func(ctx context.Context, title string) (string, error)

// Deps são as dependências da fachada.
type Deps struct {
	Organizer Organizer
	Undoer    Undoer
	Records   RecordReader
	PickDir   DirectoryPicker
	Logger    *slog.Logger
}

// OrganizationState informa se há uma organização para desfazer ao abrir o app.
type OrganizationState struct {
	HasUndo               bool   `json:"hasUndo"`
	SourceFolderPath      string `json:"sourceFolderPath"`
	DestinationFolderPath string `json:"destinationFolderPath"`
}

// DroppedPath é a pasta de origem definida por arrastar e soltar.
type DroppedPath struct {
	SourceFolderPath string `json:"sourceFolderPath"`
}

// App é a fachada exposta ao frontend. Seus métodos públicos viram os
// bindings JavaScript gerados em frontend/wailsjs/go/app/App.js.
type App struct {
	ctx  context.Context
	deps Deps
	log  *slog.Logger
	// busy impede organizar e desfazer ao mesmo tempo.
	busy sync.Mutex
}

// New cria a fachada com as dependências dadas.
func New(d Deps) *App {
	log := d.Logger
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	return &App{ctx: context.Background(), deps: d, log: log}
}

// startup guarda o contexto do runtime do Wails, necessário para diálogos.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// SelectSourceFolder abre o seletor da pasta de origem.
func (a *App) SelectSourceFolder() (string, error) {
	return a.pick(SourceDialogTitle)
}

// SelectDestinationFolder abre o seletor da pasta de destino.
func (a *App) SelectDestinationFolder() (string, error) {
	return a.pick(DestinationDialogTitle)
}

func (a *App) pick(title string) (string, error) {
	path, err := a.deps.PickDir(a.ctx, title)
	return path, a.toFrontend("selecionar pasta", err)
}

// ResolveDroppedPath define a origem a partir de um item arrastado.
func (a *App) ResolveDroppedPath(path string) (DroppedPath, error) {
	src, err := resolveDroppedPath(path)
	return DroppedPath{SourceFolderPath: src}, a.toFrontend("arrastar e soltar", err)
}

// GetLastOrganizationState informa se a última organização pode ser
// desfeita. Um registro corrompido conta como "nada para desfazer".
func (a *App) GetLastOrganizationState() OrganizationState {
	op, err := a.deps.Records.Load()
	if err != nil {
		a.log.Warn("registro da última organização ignorado", "err", err)
		return OrganizationState{}
	}
	if !op.CanUndo() {
		return OrganizationState{}
	}
	return OrganizationState{
		HasUndo:               true,
		SourceFolderPath:      op.SourceFolderPath,
		DestinationFolderPath: op.DestinationFolderPath,
	}
}

// OrganizeFiles organiza a pasta de origem segundo os critérios escolhidos.
func (a *App) OrganizeFiles(req organizer.Request) (organizer.Result, error) {
	a.busy.Lock()
	defer a.busy.Unlock()

	result, err := a.deps.Organizer.Organize(a.ctx, req)
	if err == nil {
		a.log.Info("organização concluída", "origem", result.SourceFolderPath, "destino", result.DestinationFolderPath,
			"movidos", result.MovedFiles, "falhas", result.FailedFiles)
	}
	return result, a.toFrontend("organizar", err)
}

// UndoLastOrganization desfaz a última organização.
func (a *App) UndoLastOrganization() (undo.Result, error) {
	a.busy.Lock()
	defer a.busy.Unlock()

	result, err := a.deps.Undoer.Undo(a.ctx)
	if err == nil {
		a.log.Info("desfazer concluído", "restaurados", result.RestoredFiles, "falhas", result.FailedFiles)
	}
	return result, a.toFrontend("desfazer", err)
}

// toFrontend registra o erro completo no log e devolve ao frontend só o
// código estável (ex.: "NOTHING_TO_UNDO"), que ele traduz para o idioma do usuário.
func (a *App) toFrontend(action string, err error) error {
	if err == nil {
		return nil
	}
	code := apperr.CodeOf(err)
	a.log.Error("falha", "acao", action, "codigo", code, "err", err)
	return errors.New(code)
}

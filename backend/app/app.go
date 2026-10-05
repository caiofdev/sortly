// Package app contém a fachada exposta ao frontend pelo Wails e a
// configuração da janela. A fachada guarda o estado da tela (ViewState) e
// delega as regras aos serviços de backend/; o frontend só renderiza (ADR 0005).
package app

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/caiofdev/sortly/backend/apperr"
	"github.com/caiofdev/sortly/backend/organizer"
	"github.com/caiofdev/sortly/backend/settings"
	"github.com/caiofdev/sortly/backend/store"
	"github.com/caiofdev/sortly/backend/undo"
)

// Títulos dos seletores de pasta.
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

// SettingsStore lê e altera as preferências. É satisfeito por *settings.Service.
type SettingsStore interface {
	Get() settings.Settings
	SetLanguage(lang string) (settings.Settings, error)
	SetCriterion(key string, enabled bool) (settings.Settings, error)
}

// DirectoryPicker abre o seletor de pasta e devolve o caminho escolhido ("" se cancelado).
type DirectoryPicker func(ctx context.Context, title string) (string, error)

// Emitter avisa a interface de que o estado mudou.
type Emitter func(ctx context.Context, state ViewState)

// Deps são as dependências da fachada. Logger, Emit e Now são opcionais.
type Deps struct {
	Organizer Organizer
	Undoer    Undoer
	Records   RecordReader
	Settings  SettingsStore
	PickDir   DirectoryPicker
	Emit      Emitter
	Now       func() time.Time
	Logger    *slog.Logger
}

// App é a fachada exposta ao frontend. Seus métodos públicos viram os
// bindings JavaScript gerados em frontend/wailsjs/go/app/App.js; todos
// devolvem o estado completo da tela, e os erros viram notificações.
type App struct {
	ctx     context.Context
	started bool
	deps    Deps
	log     *slog.Logger

	// work impede organizar e desfazer ao mesmo tempo; mu protege state.
	work   sync.Mutex
	mu     sync.Mutex
	state  ViewState
	loaded bool
	nextID int
}

// New cria a fachada com as dependências dadas.
func New(d Deps) *App {
	if d.Logger == nil {
		d.Logger = slog.New(slog.DiscardHandler)
	}
	if d.Now == nil {
		d.Now = time.Now
	}
	return &App{ctx: context.Background(), deps: d, log: d.Logger}
}

// startup guarda o contexto do runtime do Wails, necessário para diálogos e eventos.
func (a *App) startup(ctx context.Context) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ctx = ctx
	a.started = true
}

// runtimeContext devolve o contexto do Wails guardado no startup, que roda em
// outra goroutine.
func (a *App) runtimeContext() context.Context {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.ctx
}

// GetState devolve o estado atual. Na primeira chamada, recupera a última
// organização para que ela possa ser desfeita.
func (a *App) GetState() ViewState {
	return a.update(func(*ViewState) {})
}

// SelectSource abre o seletor da pasta de origem.
func (a *App) SelectSource() ViewState {
	return a.pick(SourceDialogTitle, ActionSelectSource, func(s *ViewState, p string) { s.SourceFolderPath = p })
}

// SelectDestination abre o seletor da pasta de destino.
func (a *App) SelectDestination() ViewState {
	return a.pick(DestinationDialogTitle, ActionSelectDestination, func(s *ViewState, p string) { s.DestinationFolderPath = p })
}

func (a *App) pick(title, action string, set func(*ViewState, string)) ViewState {
	path, err := a.deps.PickDir(a.runtimeContext(), title)
	return a.update(func(s *ViewState) {
		switch {
		case err != nil:
			a.fail(s, action, err)
		case path != "":
			set(s, path)
		}
	})
}

// DropPaths define a origem a partir dos itens soltos no painel; usa o primeiro.
func (a *App) DropPaths(paths []string) ViewState {
	if len(paths) == 0 {
		return a.GetState()
	}
	src, err := resolveDroppedPath(paths[0])
	return a.update(func(s *ViewState) {
		if err != nil {
			a.fail(s, ActionDrop, err)
			return
		}
		s.SourceFolderPath = src
		a.notify(s, Notification{Kind: KindInfo, Code: CodeSourceDropped, Action: ActionDrop, Path: src})
	})
}

// Organize organiza a pasta de origem com os critérios salvos. Destino vazio
// usa a própria origem. Com outra ação em andamento, não faz nada.
func (a *App) Organize() ViewState {
	if !a.work.TryLock() {
		return a.GetState()
	}
	defer a.work.Unlock()

	current := a.GetState()
	if current.SourceFolderPath == "" {
		return a.update(func(s *ViewState) {
			a.notify(s, Notification{Kind: KindError, Code: CodeSourceRequired, Action: ActionOrganize})
		})
	}
	req := organizer.Request{
		SourceFolderPath:      current.SourceFolderPath,
		DestinationFolderPath: current.DestinationFolderPath,
		Options:               a.deps.Settings.Get().Options,
	}
	if req.DestinationFolderPath == "" {
		req.DestinationFolderPath = req.SourceFolderPath
	}

	a.update(func(s *ViewState) { s.Busy = BusyOrganize })
	result, err := a.deps.Organizer.Organize(a.runtimeContext(), req)
	return a.update(func(s *ViewState) {
		s.Busy = ""
		a.applyOrganize(s, result, err)
	})
}

// applyOrganize leva o resultado ao estado. Com arquivos movidos, o Result vale
// mesmo com erro: se o registro não foi salvo, o desfazer anterior não pode
// continuar disponível, porque desfaria a organização errada.
func (a *App) applyOrganize(s *ViewState, result organizer.Result, err error) {
	if err == nil || result.MovedFiles > 0 {
		s.HasUndo = result.CanUndo
	}
	if err != nil {
		a.fail(s, ActionOrganize, err)
		return
	}
	a.log.Info("organização concluída", "origem", result.SourceFolderPath, "destino", result.DestinationFolderPath,
		"movidos", result.MovedFiles, "falhas", result.FailedFiles)
	a.notify(s, Notification{Kind: KindOrganize, Code: CodeOrganizeDone, Action: ActionOrganize, Organize: &result})
}

// Undo desfaz a última organização. Com outra ação em andamento, não faz nada.
func (a *App) Undo() ViewState {
	if !a.work.TryLock() {
		return a.GetState()
	}
	defer a.work.Unlock()

	a.update(func(s *ViewState) { s.Busy = BusyRestore })
	result, err := a.deps.Undoer.Undo(a.runtimeContext())
	return a.update(func(s *ViewState) {
		s.Busy = ""
		if err != nil {
			a.fail(s, ActionUndo, err)
			return
		}
		a.log.Info("desfazer concluído", "restaurados", result.RestoredFiles, "falhas", result.FailedFiles)
		s.HasUndo = result.CanUndo
		a.notify(s, Notification{Kind: KindRestore, Code: CodeUndoDone, Action: ActionUndo, Undo: &result})
	})
}

// ClearNotifications apaga o histórico de notificações.
func (a *App) ClearNotifications() ViewState {
	return a.update(func(s *ViewState) { s.Notifications = nil })
}

// SetLanguage troca o idioma.
func (a *App) SetLanguage(lang string) ViewState {
	_, err := a.deps.Settings.SetLanguage(lang)
	return a.afterSettings(err)
}

// SetCriterion liga ou desliga um critério.
func (a *App) SetCriterion(key string, enabled bool) ViewState {
	_, err := a.deps.Settings.SetCriterion(key, enabled)
	return a.afterSettings(err)
}

func (a *App) afterSettings(err error) ViewState {
	return a.update(func(s *ViewState) {
		if err != nil {
			a.fail(s, ActionSettings, err)
		}
	})
}

// update aplica change ao estado, emite o estado novo e o devolve. As
// preferências entram na hora, lidas do serviço, para nunca ficarem defasadas.
func (a *App) update(change func(*ViewState)) ViewState {
	a.mu.Lock()
	a.ensureLoaded()
	change(&a.state)
	a.state.Version++
	snapshot := a.state
	// Lista sempre presente (nunca null no JSON): a interface percorre direto.
	snapshot.Notifications = append([]Notification{}, a.state.Notifications...)
	// Lidas sob o lock para que uma versão maior nunca traga preferências mais antigas.
	snapshot.Settings = a.deps.Settings.Get().View()
	ctx, started := a.ctx, a.started
	a.mu.Unlock()

	if started && a.deps.Emit != nil {
		a.deps.Emit(ctx, snapshot)
	}
	return snapshot
}

// ensureLoaded recupera, uma única vez, a última organização desfazível.
// Um registro corrompido conta como "nada para desfazer".
func (a *App) ensureLoaded() {
	if a.loaded {
		return
	}
	a.loaded = true
	op, err := a.deps.Records.Load()
	if err != nil {
		a.log.Warn("registro da última organização ignorado", "err", err)
		return
	}
	if !op.CanUndo() {
		return
	}
	a.state.HasUndo = true
	a.state.SourceFolderPath = op.SourceFolderPath
	a.state.DestinationFolderPath = op.DestinationFolderPath
	a.notify(&a.state, Notification{Kind: KindInfo, Code: CodeRecovered, Action: ActionStartup})
}

func (a *App) notify(s *ViewState, n Notification) {
	a.nextID++
	n.ID = a.nextID
	n.At = a.deps.Now()
	s.push(n)
}

// fail registra o erro completo no log e notifica só o código estável (ex.:
// "NOTHING_TO_UNDO"), que a interface traduz para o idioma do usuário.
func (a *App) fail(s *ViewState, action string, err error) {
	code := apperr.CodeOf(err)
	a.log.Error("falha", "acao", action, "codigo", code, "err", err)
	a.notify(s, Notification{Kind: KindError, Code: code, Action: action})
}

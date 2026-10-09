// Package app é a fachada exposta ao frontend pelo Wails e a configuração da
// janela: guarda o estado da tela (ViewState) e delega as regras aos serviços;
// o frontend só renderiza (ADR 0005, #45).
package app

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/options"

	"github.com/caiofdev/sortly/backend/apperr"
	"github.com/caiofdev/sortly/backend/organizer"
	"github.com/caiofdev/sortly/backend/settings"
	"github.com/caiofdev/sortly/backend/store"
	"github.com/caiofdev/sortly/backend/undo"
)

const (
	SourceDialogTitle      = "Selecione a pasta com os arquivos para organizar"
	DestinationDialogTitle = "Selecione a pasta de destino para receber os arquivos organizados"
)

type Organizer interface {
	Organize(ctx context.Context, req organizer.Request) (organizer.Result, error)
}

type Undoer interface {
	Undo(ctx context.Context) (undo.Result, error)
}

type RecordReader interface {
	Load() (*store.Operation, error)
}

type SettingsStore interface {
	Get() settings.Settings
	SetLanguage(lang string) (settings.Settings, error)
	SetCriterion(key string, enabled bool) (settings.Settings, error)
	SetTheme(theme string) (settings.Settings, error)
}

// Devolve "" quando o usuário cancela o seletor (#10).
type DirectoryPicker func(ctx context.Context, title string) (string, error)

type Emitter func(ctx context.Context, state ViewState)

type BackgroundPainter func(ctx context.Context, colour options.RGBA)

// Logger, Emit, Paint e Now são opcionais (#10).
type Deps struct {
	Organizer Organizer
	Undoer    Undoer
	Records   RecordReader
	Settings  SettingsStore
	PickDir   DirectoryPicker
	Emit      Emitter
	Paint     BackgroundPainter
	Now       func() time.Time
	Logger    *slog.Logger
}

// Os métodos públicos viram os bindings em frontend/wailsjs/go/app/App.js;
// todos devolvem o estado completo da tela, e os erros viram notificações (ADR 0005, #45).
type App struct {
	ctx     context.Context
	started bool
	deps    Deps
	log     *slog.Logger

	// work impede organizar e desfazer ao mesmo tempo; mu protege state (#45).
	work   sync.Mutex
	mu     sync.Mutex
	state  ViewState
	loaded bool
	nextID int
}

func New(d Deps) *App {
	if d.Logger == nil {
		d.Logger = slog.New(slog.DiscardHandler)
	}
	if d.Now == nil {
		d.Now = time.Now
	}
	return &App{ctx: context.Background(), deps: d, log: d.Logger}
}

// O contexto do runtime do Wails é necessário para diálogos e eventos (#3).
func (a *App) startup(ctx context.Context) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ctx = ctx
	a.started = true
}

// Lido sob o lock porque o startup roda em outra goroutine (#45).
func (a *App) runtimeContext() context.Context {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.ctx
}

// Na primeira chamada, recupera a última organização para que ela possa ser desfeita (#45).
func (a *App) GetState() ViewState {
	return a.update(func(*ViewState) {})
}

func (a *App) SelectSource() ViewState {
	return a.pick(SourceDialogTitle, ActionSelectSource, func(s *ViewState, p string) { s.SourceFolderPath = p })
}

func (a *App) SelectDestination() ViewState {
	return a.pick(DestinationDialogTitle, ActionSelectDestination, func(s *ViewState, p string) { s.DestinationFolderPath = p })
}

func (a *App) pick(title, action string, set func(*ViewState, string)) ViewState {
	path, err := a.deps.PickDir(a.runtimeContext(), title)
	return a.update(func(s *ViewState) {
		if err != nil {
			a.fail(s, action, err)
			return
		}
		if path != "" {
			set(s, path)
		}
	})
}

// Só o primeiro item solto vale (#12).
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

// Destino vazio usa a própria origem. Com outra ação em andamento, não faz nada (#45).
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

// Com arquivos movidos, o Result vale mesmo com erro: se o registro não foi
// salvo, o desfazer anterior não pode continuar disponível, porque desfaria a
// organização errada (#53).
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
	a.notify(s, Notification{Kind: organizeKind(result), Code: CodeOrganizeDone, Action: ActionOrganize, Organize: &result})
}

// Falha parcial conta como erro, para o aviso ficar na tela até o usuário vê-lo;
// sem nada movido, o aviso é neutro (#75).
func organizeKind(r organizer.Result) string {
	if r.FailedFiles > 0 {
		return KindError
	}
	if r.MovedFiles == 0 {
		return KindInfo
	}
	return KindSuccess
}

func undoKind(r undo.Result) string {
	if r.FailedFiles > 0 {
		return KindError
	}
	return KindInfo
}

// Com outra ação em andamento, não faz nada (#45).
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
		a.notify(s, Notification{Kind: undoKind(result), Code: CodeUndoDone, Action: ActionUndo, Undo: &result})
	})
}

func (a *App) ClearNotifications() ViewState {
	return a.update(func(s *ViewState) {
		s.Notifications = nil
		s.Unread = false
	})
}

func (a *App) MarkNotificationsRead() ViewState {
	return a.update(func(s *ViewState) { s.Unread = false })
}

func (a *App) SetLanguage(lang string) ViewState {
	_, err := a.deps.Settings.SetLanguage(lang)
	return a.afterSettings(err)
}

func (a *App) SetCriterion(key string, enabled bool) ViewState {
	_, err := a.deps.Settings.SetCriterion(key, enabled)
	return a.afterSettings(err)
}

func (a *App) SetTheme(theme string) ViewState {
	st, err := a.deps.Settings.SetTheme(theme)
	if err == nil {
		a.paint(st.Theme)
	}
	return a.afterSettings(err)
}

// Sem o runtime do Wails (antes do startup), não há janela para pintar (#76).
func (a *App) paint(theme string) {
	a.mu.Lock()
	ctx, started := a.ctx, a.started
	a.mu.Unlock()
	if started && a.deps.Paint != nil {
		a.deps.Paint(ctx, backgroundFor(theme))
	}
}

func (a *App) afterSettings(err error) ViewState {
	return a.update(func(s *ViewState) {
		if err != nil {
			a.fail(s, ActionSettings, err)
		}
	})
}

// As preferências entram em todo estado, lidas do serviço, para nunca ficarem defasadas
// (#44).
func (a *App) update(change func(*ViewState)) ViewState {
	a.mu.Lock()
	a.ensureLoaded()
	change(&a.state)
	a.state.Version++
	snapshot := a.state
	// Lista sempre presente (nunca null no JSON): a interface percorre direto (#45).
	snapshot.Notifications = append([]Notification{}, a.state.Notifications...)
	// Sob o lock, para que uma versão maior nunca traga preferências mais antigas (#55).
	snapshot.Settings = a.deps.Settings.Get().View()
	ctx, started := a.ctx, a.started
	a.mu.Unlock()

	if started && a.deps.Emit != nil {
		a.deps.Emit(ctx, snapshot)
	}
	return snapshot
}

// Só uma vez; um registro corrompido conta como "nada para desfazer" (#45).
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

// O erro completo vai para o log; a interface recebe só o código estável
// (ex.: "NOTHING_TO_UNDO") e o traduz (ADR 0004, #45).
func (a *App) fail(s *ViewState, action string, err error) {
	code := apperr.CodeOf(err)
	a.log.Error("falha", "acao", action, "codigo", code, "err", err)
	a.notify(s, Notification{Kind: KindError, Code: code, Action: action})
}

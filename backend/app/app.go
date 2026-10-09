// Package app é a fachada exposta ao frontend pelo Wails e a configuração da
// janela: guarda o estado da tela (ViewState) e delega as regras aos serviços;
// o frontend só renderiza (ADR 0005, #45).
package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/options"

	"github.com/caiofdev/sortly/backend/apperr"
	"github.com/caiofdev/sortly/backend/history"
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
	OrganizeReporting(ctx context.Context, req organizer.Request, report organizer.Reporter) (organizer.Result, error)
	Preview(ctx context.Context, req organizer.Request) (organizer.Preview, error)
}

type Undoer interface {
	Undo(ctx context.Context) (undo.Result, error)
}

type RecordReader interface {
	Load() (*store.Operation, error)
}

type HistoryStore interface {
	List() []history.Entry
	Add(e history.Entry) error
	MarkLastUndone() error
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

// Abre a pasta no gerenciador de arquivos do sistema (#79).
type FolderOpener func(path string) error

// Logger, Emit, Paint, Now, Background e OpenFolder são opcionais. Background roda as tarefas
// em segundo plano (a prévia): por padrão, numa goroutine; os testes a rodam na hora
// (#10, #77).
type Deps struct {
	Organizer  Organizer
	Undoer     Undoer
	Records    RecordReader
	Settings   SettingsStore
	History    HistoryStore
	PickDir    DirectoryPicker
	Emit       Emitter
	Paint      BackgroundPainter
	OpenFolder FolderOpener
	Now        func() time.Time
	Background func(task func())
	Logger     *slog.Logger
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

	// Só a prévia mais recente vale: uma nova cancela a anterior, e o resultado de
	// uma antiga que termine depois é descartado pelo número (#77).
	cancelPreview context.CancelFunc
	previewSeq    uint64
	// Disparadas depois de soltar o lock, porque chamam update de novo (#77).
	pending []func()

	cancelWork   context.CancelFunc
	lastProgress time.Time
}

func New(d Deps) *App {
	if d.Logger == nil {
		d.Logger = slog.New(slog.DiscardHandler)
	}
	if d.Now == nil {
		d.Now = time.Now
	}
	if d.Background == nil {
		d.Background = func(task func()) { go task() }
	}
	if d.OpenFolder == nil {
		d.OpenFolder = newFolderOpener().open
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
			s.LastResult = nil
			a.startPreview(s)
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
		s.LastResult = nil
		a.startPreview(s)
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

	ctx, cancel := context.WithCancel(a.runtimeContext())
	defer cancel()
	// A prévia leria arquivos que estão sendo movidos; volta quando a organização acaba (#77).
	a.update(func(s *ViewState) {
		s.Busy = BusyOrganize
		s.Progress = organizer.Progress{}
		s.LastResult = nil
		a.cancelWork = cancel
		a.stopPreview(s)
	})
	result, err := a.deps.Organizer.OrganizeReporting(ctx, req, a.reportProgress)
	return a.update(func(s *ViewState) {
		s.Busy = ""
		s.Progress = organizer.Progress{}
		a.cancelWork = nil
		a.applyOrganize(s, result, err)
		a.startPreview(s)
	})
}

// Volta à tela inicial: sem origem e sem o Concluído. O destino escolhido fica,
// e o desfazer da organização continua disponível (#79).
func (a *App) StartOver() ViewState {
	return a.update(func(s *ViewState) {
		s.SourceFolderPath = ""
		s.LastResult = nil
		a.stopPreview(s)
	})
}

// Só o destino do último resultado, guardado aqui: o frontend não manda caminho,
// então a tela não consegue abrir uma pasta qualquer (#79).
func (a *App) OpenDestination() ViewState {
	a.mu.Lock()
	var path string
	if a.state.LastResult != nil {
		path = a.state.LastResult.DestinationFolderPath
	}
	a.mu.Unlock()
	if path == "" {
		return a.GetState()
	}
	err := a.deps.OpenFolder(path)
	return a.update(func(s *ViewState) {
		if err != nil {
			a.fail(s, ActionOpenDestination, err)
		}
	})
}

// O que já foi movido fica no registro e pode ser desfeito. Sem organização em
// andamento, não faz nada (#78).
func (a *App) Cancel() ViewState {
	return a.update(func(*ViewState) {
		if a.cancelWork != nil {
			a.cancelWork()
		}
	})
}

// No máximo um estado a cada 50 ms (~20/s): numa pasta com milhares de arquivos,
// um evento por arquivo inundaria o WebView. O primeiro e o último sempre saem (#78).
const progressInterval = 50 * time.Millisecond

func (a *App) reportProgress(p organizer.Progress) {
	a.mu.Lock()
	now := a.deps.Now()
	skip := p.Done > 0 && p.Done < p.Total && now.Sub(a.lastProgress) < progressInterval
	if !skip {
		a.lastProgress = now
	}
	a.mu.Unlock()
	if skip {
		return
	}
	a.update(func(s *ViewState) { s.Progress = p })
}

// Com arquivos movidos, o Result vale mesmo com erro: se o registro não foi
// salvo, o desfazer anterior não pode continuar disponível, porque desfaria a
// organização errada (#53).
func (a *App) applyOrganize(s *ViewState, result organizer.Result, err error) {
	if err == nil || result.MovedFiles > 0 {
		s.HasUndo = result.CanUndo
	}
	a.addHistory(result)
	// Cancelar não é falha; o aviso diz quantos arquivos já foram movidos. Se o
	// registro não foi salvo, a falha vale mais (#53, #78).
	if result.Canceled && !errors.Is(err, organizer.ErrRecordNotSaved) {
		a.log.Info("organização cancelada", "origem", result.SourceFolderPath, "movidos", result.MovedFiles)
		a.notify(s, Notification{Kind: KindInfo, Code: CodeOrganizeCanceled, Action: ActionOrganize, Organize: &result})
		return
	}
	if err != nil {
		a.fail(s, ActionOrganize, err)
		return
	}
	a.log.Info("organização concluída", "origem", result.SourceFolderPath, "destino", result.DestinationFolderPath,
		"movidos", result.MovedFiles, "falhas", result.FailedFiles)
	a.notify(s, Notification{Kind: organizeKind(result), Code: CodeOrganizeDone, Action: ActionOrganize, Organize: &result})
	if result.MovedFiles > 0 {
		s.LastResult = &result
	}
}

// Só organizações que moveram arquivos entram, como no registro do desfazer: assim
// a mais recente do histórico é sempre a que o desfazer desfaz. O histórico é
// secundário; uma falha ao gravá-lo vai só para o log (ADR 0007, #80).
func (a *App) addHistory(result organizer.Result) {
	if result.MovedFiles == 0 {
		return
	}
	status := history.StatusDone
	if result.Canceled {
		status = history.StatusCanceled
	}
	err := a.deps.History.Add(history.Entry{
		At:                    a.deps.Now(),
		SourceFolderPath:      result.SourceFolderPath,
		DestinationFolderPath: result.DestinationFolderPath,
		MovedFiles:            result.MovedFiles,
		Status:                status,
	})
	if err != nil {
		a.log.Warn("histórico não salvo", "err", err)
	}
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

	a.update(func(s *ViewState) {
		s.Busy = BusyRestore
		a.stopPreview(s)
	})
	result, err := a.deps.Undoer.Undo(a.runtimeContext())
	return a.update(func(s *ViewState) {
		s.Busy = ""
		a.startPreview(s)
		if err != nil {
			a.fail(s, ActionUndo, err)
			return
		}
		a.log.Info("desfazer concluído", "restaurados", result.RestoredFiles, "falhas", result.FailedFiles)
		s.HasUndo = result.CanUndo
		s.LastResult = nil
		if err := a.deps.History.MarkLastUndone(); err != nil {
			a.log.Warn("histórico não atualizado após desfazer", "err", err)
		}
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
	return a.update(func(s *ViewState) {
		if err != nil {
			a.fail(s, ActionSettings, err)
			return
		}
		a.startPreview(s)
	})
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
	if snapshot.Preview.Folders == nil {
		snapshot.Preview = noPreview(snapshot.Preview.Status)
	}
	// Sob o lock, para que uma versão maior nunca traga preferências mais antigas (#55).
	snapshot.Settings = a.deps.Settings.Get().View()
	snapshot.History = a.deps.History.List()
	ctx, started := a.ctx, a.started
	tasks := a.pending
	a.pending = nil
	a.mu.Unlock()

	if started && a.deps.Emit != nil {
		a.deps.Emit(ctx, snapshot)
	}
	for _, task := range tasks {
		a.deps.Background(task)
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
	a.startPreview(&a.state)
	a.notify(&a.state, Notification{Kind: KindInfo, Code: CodeRecovered, Action: ActionStartup})
}

// Chamado dentro de update, sob o lock: cancela a prévia anterior e calcula a
// nova em segundo plano. A tela recebe o resultado pelo evento de estado (#77).
func (a *App) startPreview(s *ViewState) {
	a.stopPreview(s)
	if s.SourceFolderPath == "" {
		return
	}
	ctx, cancel := context.WithCancel(a.ctx)
	a.cancelPreview = cancel
	seq := a.previewSeq
	s.Preview = noPreview(PreviewLoading)
	req := organizer.Request{
		SourceFolderPath:      s.SourceFolderPath,
		DestinationFolderPath: s.DestinationFolderPath,
		Options:               a.deps.Settings.Get().Options,
	}
	a.pending = append(a.pending, func() {
		p, err := a.runPreview(ctx, req)
		a.update(func(s *ViewState) { a.finishPreview(s, seq, p, err) })
	})
}

// A prévia lê arquivos do usuário (mp4, pdf, docx) numa goroutine própria, onde
// um panic encerraria o app: vira um erro inesperado e vai para o log (#77).
func (a *App) runPreview(ctx context.Context, req organizer.Request) (p organizer.Preview, err error) {
	defer func() {
		if r := recover(); r != nil {
			a.log.Error("panic na prévia", "origem", req.SourceFolderPath, "panic", r, "stack", string(debug.Stack()))
			err = fmt.Errorf("app: prévia: %v", r)
		}
	}()
	return a.deps.Organizer.Preview(ctx, req)
}

func (a *App) stopPreview(s *ViewState) {
	a.previewSeq++
	if a.cancelPreview != nil {
		a.cancelPreview()
		a.cancelPreview = nil
	}
	s.Preview = noPreview(PreviewNone)
}

// Cancelada por uma prévia nova não é erro; origem ilegível vira aviso (#77).
func (a *App) finishPreview(s *ViewState, seq uint64, p organizer.Preview, err error) {
	if seq != a.previewSeq {
		return
	}
	a.stopPreview(s)
	if errors.Is(err, context.Canceled) {
		return
	}
	if err != nil {
		a.fail(s, ActionPreview, err)
		return
	}
	s.Preview = PreviewState{Status: PreviewReady, Preview: p}
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

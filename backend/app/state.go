package app

import (
	"time"

	"github.com/caiofdev/sortly/backend/history"
	"github.com/caiofdev/sortly/backend/organizer"
	"github.com/caiofdev/sortly/backend/settings"
	"github.com/caiofdev/sortly/backend/undo"
)

// Durante organizar e desfazer, é por este evento que a interface sabe que a ação começou
// (#45).
const StateEvent = "sortly:state"

const MaxNotifications = 80

// Status do aviso: o ponto no painel e a variante do toast; os de erro ficam
// na tela até o usuário fechar (#75).
const (
	KindSuccess = "success"
	KindInfo    = "info"
	KindError   = "error"
)

// Códigos que não são erros; os de erro vêm de apperr (ADR 0004, #45).
const (
	CodeRecovered        = "RECOVERED_LAST_ORGANIZATION"
	CodeSourceDropped    = "SOURCE_DROPPED"
	CodeOrganizeDone     = "ORGANIZE_DONE"
	CodeOrganizeCanceled = "ORGANIZE_CANCELED"
	CodeUndoDone         = "UNDO_DONE"
	CodeSourceRequired   = "SOURCE_REQUIRED"
)

// Num erro sem tradução própria, a interface usa o texto padrão da ação (#45).
const (
	ActionStartup           = "startup"
	ActionSelectSource      = "selectSource"
	ActionSelectDestination = "selectDestination"
	ActionDrop              = "drop"
	ActionOrganize          = "organize"
	ActionUndo              = "undo"
	ActionSettings          = "settings"
	ActionPreview           = "preview"
	ActionOpenDestination   = "openDestination"
)

// Valores de PreviewState.Status: a interface mostra "contando" enquanto a prévia
// roda em segundo plano (#77).
const (
	PreviewNone    = ""
	PreviewLoading = "loading"
	PreviewReady   = "ready"
)

type PreviewState struct {
	Status string `json:"status"`
	organizer.Preview
}

// Folders nunca é nil: a interface percorre a lista direto (#77).
func noPreview(status string) PreviewState {
	return PreviewState{Status: status, Preview: organizer.Preview{Folders: []organizer.FolderCount{}}}
}

// Valores de ViewState.Busy (#45).
const (
	BusyOrganize = "organize"
	BusyRestore  = "restore"
)

// A interface monta o texto no idioma atual a partir do código e dos dados (ADR 0004, #45).
type Notification struct {
	ID       int               `json:"id"`
	Kind     string            `json:"kind"`
	Code     string            `json:"code"`
	Action   string            `json:"action,omitempty"`
	Path     string            `json:"path,omitempty"`
	Organize *organizer.Result `json:"organize,omitempty"`
	Undo     *undo.Result      `json:"undo,omitempty"`
	At       time.Time         `json:"at"`
}

// Version cresce a cada estado entregue. O estado sai do lock antes de ser
// emitido e devolvido, então duas ações quase simultâneas podem entregá-lo fora
// de ordem; a interface descarta o que tiver versão menor que a que já mostra (#55).
type ViewState struct {
	Version               uint64             `json:"version"`
	SourceFolderPath      string             `json:"sourceFolderPath"`
	DestinationFolderPath string             `json:"destinationFolderPath"`
	HasUndo               bool               `json:"hasUndo"`
	Busy                  string             `json:"busy"`
	Unread                bool               `json:"unread"`
	Preview               PreviewState       `json:"preview"`
	Progress              organizer.Progress `json:"progress"`
	// Preenchido quando a organização move arquivos: a tela mostra o Concluído
	// até Organizar outra pasta ou Desfazer (#79).
	LastResult    *organizer.Result `json:"lastResult"`
	History       []history.Entry   `json:"history"`
	Settings      settings.View     `json:"settings"`
	Notifications []Notification    `json:"notifications"`
}

// A mais recente fica no topo; as que passam de MaxNotifications saem (#45).
func (s *ViewState) push(n Notification) {
	s.Unread = true
	s.Notifications = append([]Notification{n}, s.Notifications...)
	if len(s.Notifications) > MaxNotifications {
		s.Notifications = s.Notifications[:MaxNotifications]
	}
}

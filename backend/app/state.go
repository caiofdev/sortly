package app

import (
	"time"

	"github.com/caiofdev/sortly/backend/organizer"
	"github.com/caiofdev/sortly/backend/settings"
	"github.com/caiofdev/sortly/backend/undo"
)

// Durante organizar e desfazer, é por este evento que a interface sabe que a ação começou
// (#45).
const StateEvent = "sortly:state"

const MaxNotifications = 80

// Definem a cor do aviso na interface (#45).
const (
	KindInfo     = "info"
	KindOrganize = "organize"
	KindRestore  = "restore"
	KindError    = "error"
)

// Códigos que não são erros; os de erro vêm de apperr (ADR 0004, #45).
const (
	CodeRecovered      = "RECOVERED_LAST_ORGANIZATION"
	CodeSourceDropped  = "SOURCE_DROPPED"
	CodeOrganizeDone   = "ORGANIZE_DONE"
	CodeUndoDone       = "UNDO_DONE"
	CodeSourceRequired = "SOURCE_REQUIRED"
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
)

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
	Version               uint64         `json:"version"`
	SourceFolderPath      string         `json:"sourceFolderPath"`
	DestinationFolderPath string         `json:"destinationFolderPath"`
	HasUndo               bool           `json:"hasUndo"`
	Busy                  string         `json:"busy"`
	Settings              settings.View  `json:"settings"`
	Notifications         []Notification `json:"notifications"`
}

// A mais recente fica no topo; as que passam de MaxNotifications saem (#45).
func (s *ViewState) push(n Notification) {
	s.Notifications = append([]Notification{n}, s.Notifications...)
	if len(s.Notifications) > MaxNotifications {
		s.Notifications = s.Notifications[:MaxNotifications]
	}
}

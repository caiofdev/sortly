package app

import (
	"time"

	"github.com/caiofdev/sortly/backend/organizer"
	"github.com/caiofdev/sortly/backend/settings"
	"github.com/caiofdev/sortly/backend/undo"
)

// StateEvent é o evento emitido a cada mudança de estado. Durante organizar e
// desfazer, é por ele que a interface sabe que a ação começou.
const StateEvent = "sortly:state"

// MaxNotifications limita o histórico de notificações; as mais antigas saem.
const MaxNotifications = 80

// Tipos de notificação; definem a cor do aviso na interface.
const (
	KindInfo     = "info"
	KindOrganize = "organize"
	KindRestore  = "restore"
	KindError    = "error"
)

// Códigos de notificação que não são erros do backend. A interface traduz
// cada código; os de erro são os de apperr (ADR 0004).
const (
	CodeRecovered      = "RECOVERED_LAST_ORGANIZATION"
	CodeSourceDropped  = "SOURCE_DROPPED"
	CodeOrganizeDone   = "ORGANIZE_DONE"
	CodeUndoDone       = "UNDO_DONE"
	CodeSourceRequired = "SOURCE_REQUIRED"
)

// Ações que geram notificações. Num erro sem tradução própria, a interface
// usa o texto padrão da ação.
const (
	ActionStartup           = "startup"
	ActionSelectSource      = "selectSource"
	ActionSelectDestination = "selectDestination"
	ActionDrop              = "drop"
	ActionOrganize          = "organize"
	ActionUndo              = "undo"
	ActionSettings          = "settings"
)

// Ações em andamento (ViewState.Busy).
const (
	BusyOrganize = "organize"
	BusyRestore  = "restore"
)

// Notification é um aviso estruturado: a interface monta o texto no idioma
// atual a partir do código e dos dados.
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

// ViewState é tudo o que a tela mostra.
type ViewState struct {
	SourceFolderPath      string         `json:"sourceFolderPath"`
	DestinationFolderPath string         `json:"destinationFolderPath"`
	HasUndo               bool           `json:"hasUndo"`
	Busy                  string         `json:"busy"`
	Settings              settings.View  `json:"settings"`
	Notifications         []Notification `json:"notifications"`
}

// push acrescenta a notificação no topo e descarta as que passam do limite.
func (s *ViewState) push(n Notification) {
	s.Notifications = append([]Notification{n}, s.Notifications...)
	if len(s.Notifications) > MaxNotifications {
		s.Notifications = s.Notifications[:MaxNotifications]
	}
}

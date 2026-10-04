package app

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/caiofdev/sortly/backend/metadata"
	"github.com/caiofdev/sortly/backend/organizer"
	"github.com/caiofdev/sortly/backend/settings"
	"github.com/caiofdev/sortly/backend/store"
	"github.com/caiofdev/sortly/backend/undo"
)

// NewDefault monta a fachada com os serviços reais: registro em
// ~/.sortly/last-operation.json, preferências em ~/.sortly/settings.json,
// leitura de metadados, organizador, desfazer e o seletor de pasta nativo do Wails.
func NewDefault(log *slog.Logger) (*App, error) {
	recordPath, err := store.DefaultPath()
	if err != nil {
		return nil, fmt.Errorf("app: %w", err)
	}
	records := store.New(recordPath, log)
	settingsPath := filepath.Join(filepath.Dir(recordPath), "settings.json")

	return New(Deps{
		Organizer: organizer.NewService(organizer.Deps{Metadata: metadata.Reader{}, Store: records, Logger: log}),
		Undoer:    undo.NewService(records, log),
		Records:   records,
		Settings:  settings.New(settingsPath, log),
		PickDir:   nativeDirectoryPicker,
		Emit:      emitState,
		Logger:    log,
	}), nil
}

// nativeDirectoryPicker abre o seletor de pasta do sistema. Só funciona com o
// runtime do Wails em execução (contexto recebido no startup).
func nativeDirectoryPicker(ctx context.Context, title string) (string, error) {
	return runtime.OpenDirectoryDialog(ctx, runtime.OpenDialogOptions{Title: title})
}

// emitState envia o estado para a interface pelo evento StateEvent.
func emitState(ctx context.Context, state ViewState) {
	runtime.EventsEmit(ctx, StateEvent, state)
}

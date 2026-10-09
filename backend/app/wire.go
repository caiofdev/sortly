package app

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/caiofdev/sortly/backend/metadata"
	"github.com/caiofdev/sortly/backend/organizer"
	"github.com/caiofdev/sortly/backend/settings"
	"github.com/caiofdev/sortly/backend/store"
	"github.com/caiofdev/sortly/backend/undo"
)

// Registro em ~/.sortly/last-operation.json e preferências em
// ~/.sortly/settings.json, a mesma pasta da versão 1.0 (#10, #44).
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
		Paint:     paintBackground,
		Logger:    log,
	}), nil
}

// Só funciona com o runtime do Wails em execução (contexto recebido no startup) (#10).
func nativeDirectoryPicker(ctx context.Context, title string) (string, error) {
	return runtime.OpenDirectoryDialog(ctx, runtime.OpenDialogOptions{Title: title})
}

func emitState(ctx context.Context, state ViewState) {
	runtime.EventsEmit(ctx, StateEvent, state)
}

func paintBackground(ctx context.Context, c options.RGBA) {
	runtime.WindowSetBackgroundColour(ctx, c.R, c.G, c.B, c.A)
}

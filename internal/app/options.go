package app

import (
	"io/fs"

	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

// Valores da janela iguais aos da versão Electron (electron/bootstrap/appBootstrap.js).
const (
	appTitle        = "Sortly"
	windowWidth     = 980
	windowHeight    = 700
	windowMinWidth  = 820
	windowMinHeight = 600
)

// Arrastar e soltar: o Wails só entrega os caminhos quando o arquivo é solto
// sobre um elemento com este estilo (o painel de arrastar e soltar).
const (
	DropTargetProperty = "--wails-drop-target"
	DropTargetValue    = "drop"
)

// backgroundColour é o #0f172a usado enquanto a interface carrega.
var backgroundColour = options.RGBA{R: 15, G: 23, B: 42, A: 255}

// Options monta a configuração da janela e do servidor de assets.
// A janela abre maximizada e sem menu, como na versão Electron.
func Options(a *App, assets fs.FS) *options.App {
	bg := backgroundColour

	return &options.App{
		Title:            appTitle,
		Width:            windowWidth,
		Height:           windowHeight,
		MinWidth:         windowMinWidth,
		MinHeight:        windowMinHeight,
		WindowStartState: options.Maximised,
		BackgroundColour: &bg,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		DragAndDrop: &options.DragAndDrop{
			EnableFileDrop:  true,
			CSSDropProperty: DropTargetProperty,
			CSSDropValue:    DropTargetValue,
		},
		OnStartup: a.startup,
		Bind: []interface{}{
			a,
		},
	}
}

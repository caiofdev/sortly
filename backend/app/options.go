package app

import (
	"io/fs"

	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

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
// A janela abre maximizada e sem menu. icon é o PNG
// do app (build/appicon.png), usado como ícone da janela no Linux; no Windows
// e no macOS o ícone vem do executável e do pacote .app.
func Options(a *App, assets fs.FS, icon []byte) *options.App {
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
		// Sem aceleração por GPU no WebView: a interface é estática e renderiza igual
		// por software. No Windows, o processo de GPU do WebView2 respondia por ~80%
		// da memória comprometida (docs/benchmark.md). No Linux, Never é o padrão do
		// Wails quando options.Linux é nil (wailsapp/wails#2977) e precisa ser repetido
		// aqui porque o ícone exige options.Linux. O macOS não tem opção equivalente.
		Windows: &windows.Options{
			WebviewGpuIsDisabled: true,
		},
		Linux: &linux.Options{
			Icon:             icon,
			ProgramName:      "sortly",
			WebviewGpuPolicy: linux.WebviewGpuPolicyNever,
		},
		OnStartup: a.startup,
		Bind: []interface{}{
			a,
		},
	}
}

package app

import (
	"io/fs"

	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/windows"

	"github.com/caiofdev/sortly/backend/settings"
)

const (
	appTitle        = "Sortly"
	windowWidth     = 980
	windowHeight    = 700
	windowMinWidth  = 820
	windowMinHeight = 600
)

// O Wails só entrega os caminhos quando o arquivo é solto sobre um elemento
// com este estilo: o painel de arrastar e soltar (#12).
const (
	DropTargetProperty = "--wails-drop-target"
	DropTargetValue    = "drop"
)

// O bg-000 de cada tema: a janela abre e troca de tema sem piscar outra cor
// enquanto a interface carrega (#74, #76).
var themeBackgrounds = map[string]options.RGBA{
	"dark":  {R: 0, G: 0, B: 0, A: 255},
	"light": {R: 0xf7, G: 0xf7, B: 0xf2, A: 255},
}

func backgroundFor(theme string) options.RGBA {
	if bg, ok := themeBackgrounds[theme]; ok {
		return bg
	}
	return themeBackgrounds[settings.DefaultTheme]
}

// A janela abre maximizada e sem menu. icon é o PNG do app (build/appicon.png),
// usado só no Linux: no Windows e no macOS o ícone vem do executável e do
// pacote .app (#13).
func Options(a *App, assets fs.FS, icon []byte) *options.App {
	bg := backgroundFor(a.deps.Settings.Get().Theme)

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
		// aqui porque o ícone exige options.Linux. O macOS não tem opção equivalente (#34).
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

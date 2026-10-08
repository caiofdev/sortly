package app

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
)

func TestOptionsWindow(t *testing.T) {
	opts := Options(New(Deps{}), fstest.MapFS{}, testIcon)

	cases := []struct {
		name      string
		got, want any
	}{
		{"title", opts.Title, "Sortly"},
		{"width", opts.Width, 980},
		{"height", opts.Height, 700},
		{"min width", opts.MinWidth, 820},
		{"min height", opts.MinHeight, 600},
		{"start state", opts.WindowStartState, options.Maximised},
		{"background", *opts.BackgroundColour, options.RGBA{R: 0, G: 0, B: 0, A: 255}},
		{"menu", opts.Menu == nil, true},
		{"frameless", opts.Frameless, false},
		{"arrastar e soltar", opts.DragAndDrop.EnableFileDrop, true},
		{"propriedade do alvo", opts.DragAndDrop.CSSDropProperty, "--wails-drop-target"},
		{"valor do alvo", opts.DragAndDrop.CSSDropValue, "drop"},
		{"drop do WebView ativo", opts.DragAndDrop.DisableWebViewDrop, false},
		{"ícone da janela no Linux", string(opts.Linux.Icon), string(testIcon)},
		{"nome do programa no Linux", opts.Linux.ProgramName, "sortly"},
		{"sem GPU no WebView2 (Windows)", opts.Windows.WebviewGpuIsDisabled, true},
		{"sem GPU no WebKitGTK (Linux)", opts.Linux.WebviewGpuPolicy, linux.WebviewGpuPolicyNever},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Errorf("%s = %v, want %v", tc.name, tc.got, tc.want)
			}
		})
	}
}

func TestOptionsWiresApp(t *testing.T) {
	app := New(Deps{})
	assets := fstest.MapFS{"index.html": &fstest.MapFile{Data: []byte("<html></html>")}}

	opts := Options(app, assets, testIcon)

	if opts.AssetServer == nil || opts.AssetServer.Assets == nil {
		t.Fatal("asset server sem assets")
	}
	if len(opts.Bind) != 1 || opts.Bind[0] != app {
		t.Fatalf("Bind = %v, want [app]", opts.Bind)
	}
	if opts.OnStartup == nil {
		t.Fatal("OnStartup não configurado")
	}
}

func TestOptionsDoesNotShareBackground(t *testing.T) {
	first := Options(New(Deps{}), fstest.MapFS{}, testIcon)
	first.BackgroundColour.R = 255

	second := Options(New(Deps{}), fstest.MapFS{}, testIcon)
	if *second.BackgroundColour != backgroundColour {
		t.Fatalf("cor de fundo compartilhada entre instâncias: %+v", *second.BackgroundColour)
	}
}

func TestStartupStoresContext(t *testing.T) {
	app := New(Deps{})
	ctx := context.WithValue(context.Background(), runtimeKey{}, "runtime")

	Options(app, fstest.MapFS{}, testIcon).OnStartup(ctx)

	if app.ctx != ctx {
		t.Fatal("startup não guardou o contexto do runtime")
	}
}

type runtimeKey struct{}

var testIcon = []byte("PNG de teste")

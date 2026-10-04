package main

import (
	"embed"
	"os"

	"github.com/wailsapp/wails/v2"

	"github.com/caiofdev/sortly/backend/app"
	"github.com/caiofdev/sortly/backend/logging"
)

// O embed precisa ficar na raiz: o go:embed não aceita caminhos com "..".
//
//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var icon []byte

func main() {
	log, closeLog := logging.OpenDefault(os.Stderr)
	defer func() { _ = closeLog() }()

	a, err := app.NewDefault(log)
	if err != nil {
		log.Error("não foi possível iniciar o Sortly", "err", err)
		os.Exit(1)
	}
	log.Info("Sortly iniciado")

	if err := wails.Run(app.Options(a, assets, icon)); err != nil {
		log.Error("erro do Wails", "err", err)
		os.Exit(1)
	}
}

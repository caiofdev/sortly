package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"

	"github.com/caiofdev/sortly/internal/app"
)

// O embed precisa ficar na raiz: o go:embed não aceita caminhos com "..".
//
//go:embed all:frontend/dist
var assets embed.FS

func main() {
	if err := wails.Run(app.Options(app.New(), assets)); err != nil {
		log.Fatal(err)
	}
}

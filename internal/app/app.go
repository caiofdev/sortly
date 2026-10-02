// Package app contém a fachada exposta ao frontend pelo Wails e a
// configuração da janela. A lógica de negócio fica nos demais pacotes de internal/.
package app

import "context"

// App é a fachada exposta ao frontend. Os métodos de organização e desfazer
// serão adicionados quando os serviços do backend existirem.
type App struct {
	ctx context.Context
}

// New cria a fachada da aplicação.
func New() *App {
	return &App{}
}

// startup guarda o contexto do runtime do Wails, necessário para diálogos e eventos.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

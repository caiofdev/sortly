# ADR 0001 — Migrar de Electron para Wails v2

- **Status:** Aceita
- **Data:** 2026-10-02
- **Autores:** Caio Reis, Claude

## Contexto

O Sortly 1.0 é um app Electron 31. Cada instância carrega um Chromium e um runtime Node.js completos, o que custa dezenas a centenas de MB de RAM e gera um instalador grande para um app cuja interface é uma única tela.

O backend é pequeno: 6 chamadas IPC de requisição/resposta, sem eventos, com lógica de sistema de arquivos e leitura de metadados.

## Decisão

Reescrever o app com **Wails v2** (versão estável):

- **Backend em Go**, compilado num único binário.
- **Frontend React** reaproveitado, renderizado pelo WebView nativo do sistema: WebView2 (Windows), WKWebView (macOS) e WebKitGTK (Linux).
- **Interface visualmente idêntica** nesta migração; o redesign é uma tarefa separada.
- Plataformas-alvo: Windows (instalador NSIS), macOS e Linux.
- O desenvolvimento acontece na branch `wails-rewrite` e só entra em `main` com paridade comprovada ([#15](https://github.com/caiofdev/sortly/issues/15)).

## Alternativas consideradas

| Alternativa | Por que não |
|---|---|
| Otimizar o Electron | O custo base (Chromium + Node) continua; o ganho é marginal |
| Tauri (Rust) | Mesmo modelo de WebView nativo, mas exige Rust; Go é mais simples para a equipe |
| Wails v3 | Melhor suporte a múltiplas janelas e tray, mas a API ainda está mudando. O app tem uma janela só |

## Consequências

**Positivas**
- Instalador 9× menor (8,8 MB × 82 MB), 15× menos espaço em disco (17 MB × 259 MB) e organização ~2× mais rápida, medidos em [#15](https://github.com/caiofdev/sortly/issues/15) ([benchmark.md](../benchmark.md)).
- Backend tipado e testável com a ferramenta padrão do Go.
- Remove dependências Node (`music-metadata`, `pdf-lib`, `jszip`, `image-size`).

**Negativas / riscos**
- **Memória no Windows:** o WebView2 também é Chromium, com a mesma arquitetura de processos. Com a GPU ligada, a memória ficava próxima da do Electron, porque o processo de GPU respondia por ~80% da memória comprometida. Com a GPU do WebView2 desligada ([#34](https://github.com/caiofdev/sortly/issues/34)), a memória privada cai para 177 MB em repouso (Electron: 260 MB) e 200 MB no pico (Electron: 388 MB), sem diferença visual. A renderização passa a ser por software, suficiente para uma interface estática.
- Diferenças de renderização entre WebViews. Mitigação: conferência visual nas três plataformas ([#13](https://github.com/caiofdev/sortly/issues/13)).
- O drag and drop precisa ser reimplementado com a API do Wails ([#12](https://github.com/caiofdev/sortly/issues/12)).
- As preferências salvas no `localStorage` não são migradas, porque a origem do WebView muda.
- Desenvolvedores precisam de Go e da Wails CLI.

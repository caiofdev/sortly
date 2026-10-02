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
- Consumo de memória e tamanho de instalador bem menores (será medido em [#15](https://github.com/caiofdev/sortly/issues/15)).
- Backend tipado e testável com a ferramenta padrão do Go.
- Remove dependências Node (`music-metadata`, `pdf-lib`, `jszip`, `image-size`).

**Negativas / riscos**
- Diferenças de renderização entre WebViews. Mitigação: conferência visual nas três plataformas ([#13](https://github.com/caiofdev/sortly/issues/13)).
- O drag and drop precisa ser reimplementado com a API do Wails ([#12](https://github.com/caiofdev/sortly/issues/12)).
- As preferências salvas no `localStorage` não são migradas, porque a origem do WebView muda.
- Desenvolvedores precisam de Go e da Wails CLI.

# Arquitetura

> Autores: Caio Reis, Claude

O Sortly usa **Wails v2**: backend em Go e frontend React num WebView nativo. Por que não Electron: [ADR 0001](adr/0001-electron-para-wails.md). As regras de negócio estão em [organization-rules.md](organization-rules.md).

## 1. Visão geral

O Wails usa o WebView nativo do sistema (WebView2 no Windows, WKWebView no macOS, WebKitGTK no Linux), em vez de embutir um Chromium e um Node completos. O backend vira um único binário Go.

```mermaid
flowchart LR
  subgraph Frontend["Frontend — React 18 + design system Sortly (WebView nativo)"]
    UI[Componentes<br/>e i18n] --> VS[useViewState]
  end
  VS -- "bindings: ação → ViewState" --> App
  App -. "evento sortly:state" .-> VS
  subgraph Backend["Backend — Go"]
    App[App<br/>ViewState · notificações] --> Org[organizer<br/>Planner · Executor]
    Org --> Crit[organizer/criteria<br/>uma regra por critério]
    App --> Undo[undo]
    Crit --> Meta[metadata<br/>imagem · mp4 · páginas]
    Org --> Files[fs/files<br/>Move · MoveUnique · Reserve]
    Undo --> Files
    Files --> Paths[fs/paths<br/>Ext · Equal · IsInside · Native]
    Org --> Store[store<br/>OperationStore]
    Undo --> Store
    App --> Log[logging<br/>slog]
  end
  Store --> JSON[(~/.sortly/<br/>last-operation.json)]
  App --> Settings[settings]
  Settings --> Prefs[(~/.sortly/<br/>settings.json)]
```

## 2. Pacotes Go

| Pacote | Responsabilidade |
|---|---|
| `main` (`main.go`) | Embute `frontend/dist`, monta as dependências e chama `wails.Run`. Fica na raiz porque o `go:embed` não aceita `..` e a CLI do Wails v2 compila o pacote da pasta do `wails.json` |
| `backend/app` | Fachada `App` exposta ao frontend: guarda o estado da tela (`ViewState`, notificações), delega as regras aos serviços e emite `sortly:state`; também as opções da janela (`Options`) e a composição das dependências (`wire.go`) |
| `backend/organizer` | Validação do pedido, `Planner` (calcula o plano, sem efeitos colaterais), `Executor` (aplica os movimentos e mantém o journal), erros com código ([ADR 0004](adr/0004-erros-com-codigo.md)) |
| `backend/organizer/criteria` | Os seis critérios, um por arquivo, atrás da interface `Rule` e do registry `New` ([ADR 0003](adr/0003-strategy-regras.md)); `Options` (com a ordem de exibição `Keys` e o acesso por chave) e `File` |
| `backend/history` | Histórico das organizações em `~/.sortly/history.json` (últimas 50, a mais recente primeiro, `files.WriteAtomic`), separado do registro do desfazer ([ADR 0007](adr/0007-historico-separado-do-desfazer.md)). Arquivo ausente ou corrompido vale como vazio, com log |
| `backend/settings` | Preferências (idioma, tema e critérios) em `~/.sortly/settings.json`: lê uma vez, aceita só valores válidos, recusa desligar o último critério e grava com `files.WriteAtomic`; monta a visão para a tela (`View`) |
| `backend/metadata` | Leitura de resolução de imagens, duração de mp4 e contagem de páginas (`PageCounter` por extensão) |
| `backend/undo` | Reverte o journal da última operação e remove as pastas que ficaram vazias |
| `backend/store` | `OperationStore`: lê e grava o registro da última operação de forma atômica, no mesmo formato JSON da versão 1.0 |
| `backend/fs/files` | Operações em disco: `WriteAtomic` (temporário + sync + rename), `Move` (rename com fallback entre volumes), `MoveUnique`/`Reserve` (nome livre com criação exclusiva), `MkdirAll`, `Exists`, `RemoveEmptyDir` |
| `backend/fs/paths` | Caminhos sem tocar no disco: `Ext` (como o `path.extname` do Node), `Equal`/`IsInside` (caixa de cada sistema) e `Native` (prefixo `\\?\` no Windows) |
| `backend/apperr` | Erros com código estável ([ADR 0004](adr/0004-erros-com-codigo.md)) |
| `backend/logging` | Configuração do `log/slog` em arquivo |

Princípios:

- **Injeção de dependências por construtor.** Interfaces pequenas, declaradas no pacote que as consome, permitem usar fakes nos testes.
- **Planejar separado de executar.** O `Planner` não toca no disco, o que o torna fácil de testar e prepara um futuro modo de pré-visualização.
- **`context.Context`** em operações longas, já preparado para cancelamento e progresso.
- **Complexidade ciclomática ≤ 10 por função**, verificada no CI.

## 3. Contrato com o frontend (bindings)

O frontend só renderiza ([ADR 0005](adr/0005-estado-da-tela-no-backend.md)). A fachada `backend/app.App` guarda o estado da tela, e **todo binding devolve o estado completo** (`ViewState`). Os métodos viram funções JavaScript em `frontend/wailsjs/go/app/App.js`, com tipos em `models.ts`.

| Binding | O que faz |
|---|---|
| `GetState()` | Devolve o estado. Na primeira chamada, recupera a última organização desfazível e avisa (`RECOVERED_LAST_ORGANIZATION`) |
| `SelectSource()` / `SelectDestination()` | Abre o seletor de pasta; cancelar não muda nada |
| `DropPaths(paths)` | Define a origem a partir do primeiro item solto (a pasta, ou a pasta do arquivo) |
| `Organize()` | Organiza a origem no destino (vazio = a própria origem) com os critérios salvos. Sem origem: `SOURCE_REQUIRED` |
| `Cancel()` | Cancela a organização em andamento; o que já foi movido fica no registro e pode ser desfeito. Sem organização em andamento, não faz nada |
| `Undo()` | Desfaz a última organização (e sai do Concluído) |
| `OpenDestination()` | Abre no gerenciador de arquivos (explorer, open, xdg-open) o destino do `lastResult`. O caminho vem do estado no Go, nunca do JS; pasta apagada vira `DESTINATION_NOT_FOUND` |
| `StartOver()` | "Organizar outra pasta": limpa a origem e o `lastResult`; o destino e o desfazer continuam |
| `MarkNotificationsRead()` | Apaga o ponto de não lidas do sino; a lista continua |
| `ClearNotifications()` | Apaga o histórico de notificações |
| `SetLanguage(language)` / `SetTheme(theme)` / `SetCriterion(key, enabled)` | Alteram as preferências (`backend/settings`). `SetTheme` também troca a cor de fundo da janela |

```
ViewState {
  version: number                            // cresce a cada estado entregue
  sourceFolderPath, destinationFolderPath: string
  hasUndo: bool
  busy: "" | "organize" | "restore"          // ação em andamento
  unread: bool                               // há notificação nova desde a última leitura
  preview: { status: "" | "loading" | "ready", totalFiles, folders: [{ name, count }], otherFiles }
  progress: { done, total, file, folder }    // durante organizar; total 0 = ainda planejando
  lastResult: Result | null                  // a tela Concluído; preenchido quando a organização move arquivos
  history: [{ at, sourceFolderPath, destinationFolderPath, movedFiles, status: "done" | "undone" | "canceled" }]
  settings: { language, theme: "dark" | "light", criteria: [{ key, enabled, locked }] }
  notifications: [{ id, kind: "success" | "info" | "error", code, action, path?, organize?, undo?, at }]   // até 80, a mais recente primeiro
}
```

- **Erros não rejeitam a promessa:** viram uma notificação `kind: "error"` com o código (`INVALID_SOURCE`, `NOTHING_TO_UNDO`, `DROPPED_MISSING`, `LAST_CRITERION`, `UNEXPECTED`…) e a ação que falhou. O detalhe vai para o log. O frontend traduz o código, ou usa o texto padrão da ação quando o código não tem tradução própria (ADR 0004).
- **Notificações de sucesso** também são códigos com dados: `ORGANIZE_DONE` (com `organize`: movidos, falhas, ignorados…), `UNDO_DONE` (com `undo`) e `SOURCE_DROPPED` (com `path`).
- **Status (`kind`):** o backend decide se o aviso é sucesso, neutro ou erro. Organizar ou desfazer com algum arquivo que falhou é `error`, e organizar sem nada movido é `info`. A interface usa o `kind` para o ponto do painel e a variante do toast; o toast de erro fica até o usuário fechar.
- **Prévia da origem:** ao definir a origem ou o destino, mudar um critério, ao fim de organizar e de desfazer e na origem recuperada ao abrir, o `App` planeja em segundo plano (`organizer.Service.Preview`, o mesmo `Planner` da organização, sem mover nada). Uma prévia nova cancela a anterior, e um resultado antigo que chegue depois é descartado. O binding devolve `status: "loading"`, e o resultado chega pelo evento de estado. `folders` traz até 6 pastas de 1º nível, da maior para a menor (nome vazio = raiz do destino); o resto soma em `otherFiles`. Origem ilegível vira uma notificação de erro com a ação `preview`. Durante organizar e desfazer, a prévia some.
- **Progresso e Cancelar:** o `Executor` reporta o progresso antes de cada movimento e no fim (`organizer.Progress`: feitos, total, nome do arquivo e pasta relativa ao destino). O `App` emite no máximo um estado a cada 50 ms (~20/s); o primeiro e o último sempre saem. `Cancel` cancela o `context` da organização: o resultado sai com `canceled: true`, o journal parcial é gravado, e a notificação `ORGANIZE_CANCELED` diz quantos arquivos foram movidos. Se o registro não pôde ser salvo, vale o erro `RECORD_NOT_SAVED`.
- **Concluído:** quando a organização move ao menos um arquivo (mesmo com falhas), o `lastResult` guarda o resultado, com `folders` (as pastas de 1º nível do que foi movido de fato, no mesmo resumo da prévia) e `otherFiles`. Cancelada, com erro ou sem nada movido, a tela continua na inicial. Desfazer, escolher outra origem e `StartOver` limpam o `lastResult`.
- **Histórico:** só as organizações que moveram arquivos entram (`done` ou, se interrompidas, `canceled`), as mesmas que gravam o registro do desfazer; desfazer marca a mais recente como `undone`. Uma falha ao gravar o histórico vai só para o log (ADR 0007).
- **Evento `sortly:state`:** emitido a cada mudança, com o estado inteiro. É por ele que a tela mostra "Organizando…" enquanto a chamada de `Organize` ainda não terminou.
- **Estados fora de ordem:** evento e retorno do binding saem do lock antes de chegar à tela, então duas ações quase simultâneas podem entregá-los fora de ordem. O `useViewState` só troca o estado por um de `version` maior.
- **Organizar e desfazer não rodam juntos:** uma chamada durante a outra devolve o estado sem fazer nada.
- **Listas são sempre arrays** no JSON, nunca `null`.

**Log:** `backend/logging` grava em `sortly.log` na pasta de configuração do usuário (`%AppData%\Sortly\logs` no Windows, `~/Library/Application Support/Sortly/logs` no macOS, `~/.config/Sortly/logs` no Linux). Ao passar de 5 MB, o arquivo vira `sortly.log.1` na próxima abertura. Como o log traz caminhos de arquivos do usuário, no macOS e no Linux a pasta é `0o700` e os arquivos `0o600` (logs de versões anteriores são corrigidos ao abrir); a pasta `~/.sortly` também é criada só para o dono. Se a pasta não puder ser usada, o log vai para o stderr e o app abre normalmente.

## 4. Frontend

A interface segue o design system "Sortly" (amarelo `#F5E600`, fonte pixel; [ADR 0006](adr/0006-design-system-no-lugar-do-tailwind.md)), e o frontend não tem regra de negócio ([ADR 0005](adr/0005-estado-da-tela-no-backend.md)). O tema escuro é o principal. O claro vem do `data-theme="light"` no `<html>`, que o `App` aplica a partir do estado. Ao contrário do design system, no tema claro a sidebar fica branca, e não preta; a janela já abre com o fundo do tema salvo:

| Módulo | Responsabilidade |
|---|---|
| `hooks/useViewState.js` | Único módulo que importa os bindings e o runtime do Wails. Espelho do `ViewState`: estado inicial, evento `sortly:state`, arquivos soltos (`DropPaths`) e as ações; fora do Wails, fica no estado inicial |
| `i18n/` | Textos PT/EN; `notifications.js` transforma notificações estruturadas em título e texto no idioma atual |
| `views/` | Shell (`OrganizerView`: sidebar, cabeçalho e a página aberta) e as páginas Organizar e Configurações |
| `components/` | Peças do design system: Sidebar, PageHead, Dropzone, PathField, botões, Switch, notificações, toasts e ícones |
| `styles/` | `tokens.css` (temas), `components.css` (classes `st-*` do design system, sem edição) e `app.css` (fonte embutida e ajustes) |

A estrutura de pastas do repositório está em [development.md](development.md#3-estrutura-do-repositório).

## 5. Compatibilidade com a versão 1.0

- Os nomes das pastas criadas são idênticos.
- O formato de `~/.sortly/last-operation.json` é o mesmo, então um desfazer pendente da versão 1.0 funciona na 2.0.
- Os textos em português exibidos ao usuário continuam os mesmos.
- As preferências da 1.0 ficavam no `localStorage` do Electron e **não** são migradas. Na 2.0 elas ficam em `~/.sortly/settings.json`, fora do WebView. O usuário volta aos padrões (português, critério Extensão) uma única vez.

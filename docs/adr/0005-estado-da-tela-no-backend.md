# ADR 0005 — Estado da tela no backend; o frontend só renderiza

- **Status:** Aceita
- **Data:** 2026-10-04
- **Autores:** Caio Reis, Claude
- **Substitui:** [ADR 0002](0002-gateway-frontend.md)

## Contexto

Depois da migração, o frontend ainda guardava regra e estado: um controller com origem, destino, ação em andamento e "pode desfazer"; validações ("selecione a origem antes de organizar"); o histórico de notificações (limite, ids, horário); e a tradução de erros por ação. Essa lógica era testada em JavaScript e duplicava decisões que o Go já tomava. As preferências já tinham ido para o Go na #44.

## Decisão

- O Go guarda o **`ViewState`**: origem, destino, desfazer disponível, ação em andamento (`busy`), preferências e até 80 notificações.
- **Cada binding devolve o estado completo**: `GetState`, `SelectSource`, `SelectDestination`, `DropPaths`, `Organize`, `Undo`, `ClearNotifications`, `SetLanguage` e `SetCriterion`. O frontend troca o estado inteiro pelo que recebeu.
- **Erros viram notificações** com código e ação, sem rejeitar a promessa. O frontend não tem tratamento de erro.
- **Notificações são estruturadas** (`{kind, code, action, path, organize, undo, at}`). O texto continua no frontend, no idioma atual (ADR 0004 mantida: o backend não produz texto para o usuário).
- **Evento `sortly:state`**: emitido a cada mudança. É o que mostra "Organizando…" enquanto a chamada de `Organize` não termina.
- **`version`** cresce a cada estado entregue, sob o lock. Evento e retorno podem chegar fora de ordem, e o frontend descarta o estado com versão menor ou igual à que já mostra (#55).
- Organizar e desfazer não rodam ao mesmo tempo: uma segunda chamada durante a primeira devolve o estado sem fazer nada.
- O gateway continua sendo o único módulo que importa os bindings, mas não normaliza erros nem tem regra (veja a atualização abaixo). A assinatura dos arquivos soltos fica no JS porque o filtro `--wails-drop-target` é feito pelo runtime JS do Wails; ela só chama `DropPaths`.

## Alternativas consideradas

| Alternativa | Por que não |
|---|---|
| Manter o controller no frontend | Regra de tela testada em JS e duplicada com o Go; era o que a issue #45 queria eliminar |
| Só eventos, sem retorno nos bindings | Cada ação precisaria esperar um evento para saber o resultado; o retorno direto é mais simples e o evento cobre só o meio da ação |
| Texto das notificações pronto no Go | Divide o i18n entre Go e JS (os rótulos da tela continuam no JS) e contraria a ADR 0004 |
| Filtrar o drop no Go (`runtime.OnFileDrop`) | O Go recebe todo drop na janela; o filtro por alvo (`--wails-drop-target`) só existe no runtime JS |

## Consequências

- O frontend fica com componentes, textos, a abertura e o fechamento dos painéis, o destaque ao arrastar e as assinaturas de eventos do Wails.
- As regras da tela ganham testes em Go, com o resto do backend; os testes do frontend cobrem renderização e tradução.
- Toda lista do `ViewState` precisa ser um array no JSON, nunca `null` (teste `TestStateJSONAlwaysHasLists`).
- A tela só aparece depois do primeiro `GetState`, então abre direto no idioma salvo.
- O horário das notificações é formatado no idioma atual, não no idioma do momento do aviso.

## Atualização (2026-10-07, #61)

O `services/sortlyGateway.js` foi removido. Depois desta decisão, ele só repassava chamadas: não normalizava erros, não isolava outro runtime, e os fakes que justificavam a camada são feitos com `vi.mock` do Vitest. Agora o `hooks/useViewState.js` é o único módulo que importa os bindings (`wailsjs/go/app/App`) e o runtime do Wails, e trata a ausência do backend (`npm run dev` no navegador): sem `window.runtime`, não assina eventos e as ações deixam o estado inicial.

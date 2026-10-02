# Code review — versão Electron 1.0

> Autores: Caio Reis, Claude · Data: 2026-10-02 · Base: commit `3da065f`

Revisão feita antes da reescrita em Wails. Cada achado tem a issue que o resolve. Os defeitos (B1–B7) recebem **teste de regressão** obrigatório na reescrita.

## 1. Defeitos

| ID | Severidade | Onde | Problema | Correção | Issue |
|---|---|---|---|---|---|
| B1 | Alta | `organizeFilesService.js:212-260` | Se um movimento falha no meio do lote, a exceção interrompe a função antes de `setLastOperation`. Os arquivos já movidos **não podem ser desfeitos**. | Journal: registrar cada movimento concluído e persistir mesmo em falha parcial. O resultado informa movidos e falhas. | [#8](https://github.com/caiofdev/sortly/issues/8) |
| B2 | Alta | `organizeFilesService.js:255` | Organizar uma pasta onde nada é movido sobrescreve o registro com uma lista vazia e **apaga o desfazer da operação anterior**. | Persistir somente se houve ao menos um movimento. | [#8](https://github.com/caiofdev/sortly/issues/8) |
| B3 | Média | `fsAdapter.js:11-13` | `fs.rename` falha ao mover entre discos ou volumes diferentes (EXDEV). | `Move` com fallback: copia, preserva a data de modificação e remove a origem. | [#5](https://github.com/caiofdev/sortly/issues/5) |
| B4 | Média | `pathModel.js:13` | `while (true)` sem limite e corrida entre "o nome existe?" e "mover" (TOCTOU). | Limite de tentativas e criação exclusiva do destino. | [#5](https://github.com/caiofdev/sortly/issues/5) |
| B5 | Média | `lastOperationRepository.js:30-40` | A gravação não é atômica (um crash pode deixar JSON corrompido) e os erros são descartados silenciosamente. | Gravar em arquivo temporário + rename; registrar erros em log. | [#6](https://github.com/caiofdev/sortly/issues/6) |
| B6 | Baixa | serviços do backend | Mensagens de erro misturam português ("Pasta inválida.") e inglês ("Invalid dropped item."). Com a interface em inglês, o usuário vê texto em português. | Erros com código; o frontend traduz ([ADR 0004](adr/0004-erros-com-codigo.md)). | [#8](https://github.com/caiofdev/sortly/issues/8) |
| B7 | Baixa | `organizeFilesService.js:269`, `undoOrganizationService.js:89` | O backend monta um campo `message` em português que a interface nunca usa (`feedbackCopy` gera o próprio texto). | Remover `message` do contrato: o backend devolve só dados. | [#8](https://github.com/caiofdev/sortly/issues/8) |

## 2. Backend — design

| Achado | Princípio | Refatoração | Issue |
|---|---|---|---|
| `organizeFiles` valida, varre, classifica, move, persiste e formata mensagem numa única função | SRP | `Validate` → `Planner.Plan` (sem efeitos) → `Executor.Apply` → `Store.Save` | [#8](https://github.com/caiofdev/sortly/issues/8) |
| `buildSegments` é uma cadeia de `if`; cada critério novo exige editar a função (complexidade alta) | OCP | Strategy `SegmentRule` + registry ordenado ([ADR 0003](adr/0003-strategy-regras.md)) | [#8](https://github.com/caiofdev/sortly/issues/8) |
| Contagem de páginas por `if/else` de extensão; docx e odt repetem "abrir zip → ler entrada → regex" | DRY, OCP | `map[string]PageCounter`; um único `zipRegexCounter{entry, pattern}` com duas instâncias | [#7](https://github.com/caiofdev/sortly/issues/7) |
| `try/catch → 'xxx-unknown'` repetido em três funções | DRY | Helper `segmentOrUnknown(prefix, fn)` | [#7](https://github.com/caiofdev/sortly/issues/7) |
| Checagem de mp4 duplicada (`buildSegments:173` e `getDurationFolderName:74`) | DRY | `Applies()` da regra é o único lugar | [#8](https://github.com/caiofdev/sortly/issues/8) |
| `.doc` está na lista de paginados, mas não tem leitor | Clareza | Só extensões com contador registrado; o resultado continua `pages-unknown` | [#7](https://github.com/caiofdev/sortly/issues/7) |
| Desfazer reimplementa o laço de movimentos | DRY | Command: `MoveCommand{From, To}`; desfazer executa os inversos com o mesmo `Executor` | [#9](https://github.com/caiofdev/sortly/issues/9) |
| Repositório com estado global de módulo (`lastOperationRepository.js:5-6`) | DIP, testabilidade | `OperationStore` injetado por construtor, com caminho configurável | [#6](https://github.com/caiofdev/sortly/issues/6) |
| Nenhum log | Observabilidade | `log/slog` em arquivo | [#10](https://github.com/caiofdev/sortly/issues/10) |
| Sem cancelamento ou progresso | Extensibilidade | `context.Context` nas operações longas | [#8](https://github.com/caiofdev/sortly/issues/8) |

## 3. Frontend — lógica (sem mudança visual)

| Achado | Onde | Princípio | Refatoração | Issue |
|---|---|---|---|---|
| Acesso direto à global `window.electronAPI` | `useFileOrganizerController.js` | DIP | `services/sortlyGateway.js` ([ADR 0002](adr/0002-gateway-frontend.md)) | [#11](https://github.com/caiofdev/sortly/issues/11) |
| `handleSelectSourceFolder` e `handleSelectDestinationFolder` são quase idênticos | `useFileOrganizerController.js:59-85` | DRY | `selectFolder(gatewayFn, setter, errorKey)` | [#11](https://github.com/caiofdev/sortly/issues/11) |
| `handleOrganizeFiles` e `handleUndoLastOrganization` repetem loading → try → feedback → finally | `useFileOrganizerController.js:102-142` | DRY | `runAction(kind, fn, successFormatter, errorKey)` | [#11](https://github.com/caiofdev/sortly/issues/11) |
| A view converte `feedback` em notificação via `useEffect` + `processedFeedbackIdsRef` + `onClearFeedback` | `OrganizerView.jsx:32-58` | SRP | `useNotifications()` chamado pelo controller; a view fica só apresentacional | [#11](https://github.com/caiofdev/sortly/issues/11) |
| Geração de IDs duplicada | `controller:16`, `OrganizerView.jsx:50` | DRY | `utils/createId.js` com `crypto.randomUUID()` | [#11](https://github.com/caiofdev/sortly/issues/11) |
| Mesmo padrão de `localStorage` em dois hooks, sem try/catch na escrita | `useLanguagePreference.js`, `useOrganizationOptions.js` | DRY | `usePersistentState(key, default, validate)` | [#11](https://github.com/caiofdev/sortly/issues/11) |
| Regra "pelo menos um critério" em três lugares | hook, `OrganizerSettingsPanel.jsx:2,38`, backend | DRY | `domain/organizationOptions.js` (o backend mantém a validação defensiva) | [#11](https://github.com/caiofdev/sortly/issues/11) |
| Fallback de idioma `copy[language] \|\| copy['pt-BR']` e mapeamento de locale repetidos | controller, view | DRY | `getCopy(dict, language)`, `toLocale(language)` | [#11](https://github.com/caiofdev/sortly/issues/11) |
| `getItemTone` é uma cadeia de `if`; `restore` e `error` retornam o mesmo valor | `NotificationsCenter.jsx:1-15` | Simplicidade | Mapa de lookup com constante compartilhada | [#11](https://github.com/caiofdev/sortly/issues/11) |
| `useEffect` usa `copy` fora da lista de dependências | `useFileOrganizerController.js:57` | Correção | Corrigido com ESLint `react-hooks/exhaustive-deps` | [#11](https://github.com/caiofdev/sortly/issues/11) |
| Código morto: `FeedbackPanel.jsx` nunca é importado; chaves `destinationButton`, `feedbackProgressRunning`, `feedbackProgressDone` sem uso | `components/`, `i18n/organizerCopy.js` | Limpeza | Remover | [#11](https://github.com/caiofdev/sortly/issues/11) |
| Números mágicos (80 notificações, 72 caracteres) | view, `FolderPathsPanel` | Clareza | Constantes nomeadas | [#11](https://github.com/caiofdev/sortly/issues/11) |

## 4. Infraestrutura

| Achado | Correção | Issue |
|---|---|---|
| Fonte Inter carregada do Google Fonts em tempo de execução (`src/index.css:1`); não funciona offline | Embutir com `@fontsource/inter` | [#13](https://github.com/caiofdev/sortly/issues/13) |
| Drag and drop usa `File.path`, que só existe no Electron | Evento de drop nativo do Wails | [#12](https://github.com/caiofdev/sortly/issues/12) |
| Copyright do instalador diz "Jorge"; o autor é Caio Fernandes dos Reis | Corrigir os metadados | [#14](https://github.com/caiofdev/sortly/issues/14) |
| Sem testes, sem lint, sem CI | golangci-lint, ESLint, Vitest e GitHub Actions | [#4](https://github.com/caiofdev/sortly/issues/4) |

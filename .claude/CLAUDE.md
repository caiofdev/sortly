# Sortly — guia para o Claude

App de desktop que organiza os arquivos de uma pasta em subpastas por critérios (extensão, data, tamanho, resolução, duração de mp4, páginas) e desfaz a última organização. Feito em **Wails v2**: backend em Go, frontend React 18 + Vite + Tailwind num WebView nativo. Repositório `caiofdev/sortly`, branch principal `main`.

Responda e escreva (código, comentários, docs, commits, PRs) em **português do Brasil**.

## Estrutura

```
main.go             ponto de entrada: embute frontend/dist e chama wails.Run (precisa ficar na raiz por causa do go:embed)
wails.json          configuração do Wails e versão do app (info.productVersion)
backend/            backend Go (pacotes por responsabilidade; teste sempre ao lado do código)
  app/              fachada exposta ao frontend (bindings), drop nativo, opções da janela, composição (wire.go)
  organizer/        Planner puro, Executor com journal, serviço
    criteria/       critérios (Strategy + registry): um arquivo por critério, opções, File
  undo/             desfazer pelo inverso do journal
  metadata/         resolução de imagem, duração de mp4, páginas (pdf/docx/odt)
  store/            registro da última organização (~/.sortly/last-operation.json)
  settings/         preferências (idioma e critérios) em ~/.sortly/settings.json
  fs/paths/         caminhos: Ext, Equal, IsInside, Native (prefixo \\?\ do Windows)
  fs/files/         disco: WriteAtomic, Move (fallback entre volumes), MoveUnique, Reserve, pastas
  apperr/           erros com código estável
  logging/          slog em arquivo
  tests/            testes de integração (só API pública, disco de verdade)
frontend/src/       React só de apresentação: components/, views/, hooks/useViewState (único que importa wailsjs/), i18n/; teste ao lado de cada módulo
frontend/tests/     testes de integração do frontend (App inteiro)
frontend/wailsjs/   bindings gerados pelo Wails (versionados; regenerados por wails dev/build)
build/              ícones, manifesto Windows, Info.plist, NSIS (windows/installer), nfpm (linux)
scripts/            check-coverage.sh, cccases/ (CC × casos), benchmark/ e parity/ (PowerShell)
docs/               architecture.md, organization-rules.md, development.md (inclui testes e fluxo), release.md, benchmark.md, adr/, images/
```

## Comandos

```bash
wails dev                         # app em desenvolvimento (no Linux: -tags webkit2_41)
wails build                       # build/bin/Sortly(.exe)

go test ./...                     # testes do backend
golangci-lint run ./...           # lint, gofmt/goimports, gocyclo/cyclop ≤ 10
bash scripts/check-coverage.sh    # cobertura de backend/ ≥ 85%
go run ./scripts/cccases          # CC × casos de teste por função

cd frontend && npm run lint && npm run format:check && npm test   # ESLint (complexity 10), Prettier, Vitest
```

## Fluxo de trabalho

1. Toda mudança nasce de uma issue. Crie a branch **ligada à issue** (para fechar sozinha no merge):
   `gh issue develop N --base main --name sortly-N-descricao-curta --checkout`
2. Commits semânticos com o número da issue: `tipo(sortly-N): descrição no imperativo`. Tipos: `feat`, `fix`, `refactor`, `perf`, `test`, `docs`, `ci`, `build`, `chore`. Todo commit termina com a linha de coautoria do Claude.
3. Antes de abrir o PR, rode a skill `/sortly-review` sobre o diff e corrija os achados.
4. PR para `main` com o template `.github/pull_request_template.md` preenchido (issue, tipo, resumo, mudanças por área, testes com CC × casos, interface, checklist, riscos, autores) e o rodapé do Claude Code.
5. Mudanças que o usuário percebe vão no `CHANGELOG.md` → `[Não lançado]`, em linguagem simples.
6. Não crie tags nem releases sem pedido explícito.

**Coautoria:** Caio Reis e Claude são coautores de tudo: commits (trailer), PRs (seção Autores), issues (linha final "Autores: Caio Reis (@caiofdev) & Claude (coautor)"), README e CHANGELOG.

## Regras de código

- **Complexidade ciclomática ≤ 10** por função (Go e JS); a CI falha acima disso.
- **Testes:** cada função com CC = N tem pelo menos N casos (confira com `go run ./scripts/cccases`); fronteiras testadas em limite − 1, limite e limite + 1; todo bug corrigido ganha teste de regressão. Go: testes em tabela, `t.TempDir()`, fakes das interfaces. Frontend: Vitest + React Testing Library.
- **Go:** injeção de dependências por construtor; interfaces pequenas declaradas no pacote que as usa; erros com `%w` e códigos de `apperr`; `context.Context` em operações longas.
- **Comentários:** explique o *porquê* de trechos difíceis e documente identificadores exportados (godoc de uma linha, cobrado pelo `revive`); não descreva o que o código já diz nem conte a história da migração. Testes não têm tabela de CC: use `go run ./scripts/cccases`.
- **Frontend só renderiza (ADR 0005):** o estado da tela mora em `backend/app` (`ViewState`); cada binding devolve o estado completo e erros viram notificações. Nada de regra, validação ou estado de negócio no JS. Toda lista enviada ao frontend é array no JSON, nunca `null`.

## Invariantes (não quebre sem uma ADR nova)

- Nomes das pastas criadas e o formato de `~/.sortly/last-operation.json` são os da versão 1.0 (um desfazer pendente da 1.0 funciona na 2.0). Especificação em `docs/organization-rules.md`.
- O backend devolve **dados e códigos de erro**, nunca texto para o usuário; o i18n fica inteiro no frontend (ADR 0004).
- A interface não muda visualmente sem pedido explícito.
- Decisões registradas em `docs/adr/`: 0001 Electron → Wails, 0002 gateway no frontend (substituída), 0003 Strategy + registry nas regras, 0004 erros com código, 0005 estado da tela no backend.

## Armadilhas do ambiente

- A máquina de desenvolvimento é Windows. Os scripts `.ps1` rodam no Windows PowerShell 5.1 e precisam de **UTF-8 com BOM** por causa dos acentos.
- No Linux, Wails e `go vet`/`go test` do pacote `app` precisam da tag `webkit2_41`.
- `.go`, `.sh` e `go.mod` usam LF (`.gitattributes`); o resto pode vir com CRLF do checkout.
- O loader do WebView2 no Wails ignora `WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS`; para automatizar a interface no Windows use UI Automation (veja `scripts/parity/windows.ps1`).
- Para rodar o app sem tocar nos dados reais, aponte `USERPROFILE`, `APPDATA` e `LOCALAPPDATA` para uma pasta temporária.

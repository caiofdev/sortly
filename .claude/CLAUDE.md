# Sortly — guia para o Claude

App de desktop que organiza os arquivos de uma pasta em subpastas por critérios (extensão, data, tamanho, resolução, duração de mp4, páginas) e desfaz a última organização. Feito em **Wails v2**: backend em Go, frontend React 18 + Vite + Tailwind num WebView nativo. Repositório `caiofdev/sortly`, branch principal `main`.

Responda e escreva (código, comentários, docs, commits, PRs) em **português do Brasil**.

## Estrutura

```
main.go             ponto de entrada: embute frontend/dist e chama wails.Run (precisa ficar na raiz por causa do go:embed)
wails.json          configuração do Wails e versão do app (info.productVersion)
internal/           backend Go
  app/              fachada exposta ao frontend (bindings), drop nativo, opções da janela, composição (wire.go)
  organizer/        regras (Strategy + registry), Planner puro, Executor com journal, serviço
  undo/             desfazer pelo inverso do journal
  metadata/         resolução de imagem, duração de mp4, páginas (pdf/docx/odt)
  store/            registro da última organização (~/.sortly/last-operation.json), gravação atômica
  fsutil/           mover (com fallback entre volumes), nome único, caminhos
  apperr/           erros com código estável
  logging/          slog em arquivo
frontend/src/       React: components/, views/, controllers/, hooks/, services/, domain/, i18n/
frontend/wailsjs/   bindings gerados pelo Wails (versionados; regenerados por wails dev/build)
build/              ícones, manifesto Windows, Info.plist, NSIS (windows/installer), nfpm (linux)
scripts/            check-coverage.sh, cccases/ (CC × casos), benchmark/ e parity/ (PowerShell)
docs/               architecture.md, organization-rules.md, development.md (inclui testes e fluxo), release.md, benchmark.md, adr/, images/
```

A milestone "Refinamento do backend" (#39–#45) muda parte disso: `internal/` vira `backend/` (#42) e o estado da tela vai para o Go (#45). Atualize este arquivo quando essas issues entrarem.

## Comandos

```bash
wails dev                         # app em desenvolvimento (no Linux: -tags webkit2_41)
wails build                       # build/bin/Sortly(.exe)

go test ./...                     # testes do backend
golangci-lint run ./...           # lint, gofmt/goimports, gocyclo/cyclop ≤ 10
bash scripts/check-coverage.sh    # cobertura de internal/ ≥ 85%
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
- **Comentários:** explique o *porquê* de trechos difíceis e documente identificadores exportados; não descreva o que o código já diz.
- **Frontend:** a direção é o frontend só renderizar (#45); não acrescente regra de negócio no JS.

## Invariantes (não quebre sem uma ADR nova)

- Nomes das pastas criadas e o formato de `~/.sortly/last-operation.json` são os da versão 1.0 (um desfazer pendente da 1.0 funciona na 2.0). Especificação em `docs/organization-rules.md`.
- O backend devolve **dados e códigos de erro**, nunca texto para o usuário; o i18n fica inteiro no frontend (ADR 0004).
- A interface não muda visualmente sem pedido explícito.
- Decisões registradas em `docs/adr/`: 0001 Electron → Wails, 0002 gateway no frontend, 0003 Strategy + registry nas regras, 0004 erros com código.

## Armadilhas do ambiente

- A máquina de desenvolvimento é Windows. Os scripts `.ps1` rodam no Windows PowerShell 5.1 e precisam de **UTF-8 com BOM** por causa dos acentos.
- No Linux, Wails e `go vet`/`go test` do pacote `app` precisam da tag `webkit2_41`.
- `.go`, `.sh` e `go.mod` usam LF (`.gitattributes`); o resto pode vir com CRLF do checkout.
- O loader do WebView2 no Wails ignora `WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS`; para automatizar a interface no Windows use UI Automation (veja `scripts/parity/windows.ps1`).
- Para rodar o app sem tocar nos dados reais, aponte `USERPROFILE`, `APPDATA` e `LOCALAPPDATA` para uma pasta temporária.

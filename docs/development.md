# Desenvolvimento

> Autores: Caio Reis, Claude

Como preparar o ambiente, rodar o app e contribuir. A arquitetura está em [architecture.md](architecture.md), os testes em [testing.md](testing.md) e a geração de versões em [release.md](release.md).

## 1. Pré-requisitos

| Ferramenta | Versão | Observação |
|---|---|---|
| Go | 1.25+ | versão em [`go.mod`](../go.mod) |
| Node.js | 20+ | a CI usa a 22 |
| Wails CLI | v2 | `go install github.com/wailsapp/wails/v2/cmd/wails@latest` |
| golangci-lint | v2 | lint e complexidade do Go |

Por sistema:

- **Windows:** o WebView2 já vem no Windows 10/11 atualizados. Para gerar o instalador, também o [NSIS](https://nsis.sourceforge.io/) (`makensis` no `PATH`).
- **macOS:** Xcode Command Line Tools (`xcode-select --install`).
- **Linux (Debian/Ubuntu):** `sudo apt-get install libgtk-3-dev libwebkit2gtk-4.1-dev`. Use a build tag `webkit2_41` em todos os comandos do Wails e do Go (`-tags webkit2_41`).

Confira tudo com:

```bash
wails doctor
```

## 2. Rodando o app

```bash
git clone https://github.com/caiofdev/sortly.git
cd sortly
wails dev        # instala as dependências do frontend na primeira vez
```

O `wails dev` compila o Go, sobe o Vite e abre a janela. Mudanças no frontend recarregam na hora; mudanças no Go recompilam e reabrem o app. Na janela de desenvolvimento, o botão direito abre o "Inspecionar" do WebView.

Para gerar o executável de produção:

```bash
wails build      # build/bin/Sortly.exe (Windows), Sortly.app (macOS) ou Sortly (Linux)
```

O instalador e os pacotes de cada sistema estão em [release.md](release.md).

### Só o frontend no navegador

`npm run dev` em `frontend/` abre a interface no navegador, sem backend: as ações mostram o erro padrão. Serve para ajustar layout; para testar o fluxo completo, use `wails dev`.

## 3. Estrutura do repositório

```
sortly/
  main.go                 # ponto de entrada: embute frontend/dist e chama wails.Run
  wails.json              # configuração do Wails e versão do app (info.productVersion)
  internal/               # backend Go (detalhes em architecture.md)
  frontend/
    src/                  # React + Tailwind
    wailsjs/              # bindings gerados pelo Wails (versionados)
  build/                  # ícones, manifesto do Windows, Info.plist, NSIS e nfpm
  scripts/
    check-coverage.sh     # cobertura mínima de internal/
    benchmark/memory.ps1  # benchmark de memória no Windows (benchmark.md)
    parity/windows.ps1    # roteiro de paridade no Windows (checklist-paridade.md)
  docs/                   # esta documentação
  .github/                # CI, release e template de PR
```

### Bindings do frontend

Os métodos públicos de `internal/app.App` viram funções em `frontend/wailsjs/go/app/App.js`. O `wails dev` e o `wails build` regeneram esses arquivos; depois de mudar a assinatura de um método, rode um dos dois e versione o resultado. O frontend não importa os bindings diretamente: tudo passa por `services/sortlyGateway.js` ([ADR 0002](adr/0002-gateway-frontend.md)).

## 4. Testes e qualidade

Resumo dos comandos (critérios completos em [testing.md](testing.md)):

```bash
go test ./...                    # backend
golangci-lint run ./...          # lint, formatação e complexidade ≤ 10
bash scripts/check-coverage.sh   # cobertura de internal/ (mínimo 85%)

cd frontend
npm run lint                     # ESLint (complexidade ≤ 10)
npm run format:check             # Prettier
npm test                         # Vitest
```

## 5. Dados do app na máquina

| O quê | Onde |
|---|---|
| Registro do último organizar (desfazer) | `~/.sortly/last-operation.json`, o mesmo da versão 1.0 |
| Log | `%AppData%\Sortly\logs` (Windows), `~/Library/Application Support/Sortly/logs` (macOS), `~/.config/Sortly/logs` (Linux) |
| Idioma e critérios | `localStorage` do WebView |

Para testar sem mexer nos seus dados, abra o app com outra pasta de usuário. No Windows, defina `USERPROFILE`, `APPDATA` e `LOCALAPPDATA` apontando para uma pasta temporária, como fazem os scripts em `scripts/`.

## 6. Convenções

- **Claude Code:** o [`.claude/CLAUDE.md`](../.claude/CLAUDE.md) resume estrutura, comandos e fluxo para o Claude. Antes de abrir um PR, rode a skill `/sortly-review` ([`.claude/skills/sortly-review`](../.claude/skills/sortly-review/SKILL.md)): ela roda lint, testes e `scripts/cccases`, percorre checklists de Go, React, testes, design, sistema operacional e segurança e devolve achados por severidade. A revisão humana continua obrigatória.

- **Branches:** uma por issue, `sortly-N-descricao-curta`, criada e ligada à issue com:

  ```bash
  gh issue develop N --base main --name sortly-N-descricao-curta --checkout
  ```

- **Commits:** `tipo(sortly-N): descrição` (tipos: `feat`, `fix`, `refactor`, `perf`, `test`, `docs`, `ci`, `build`, `chore`). Todo commit leva a linha de coautoria; o modelo [`.gitmessage`](../.gitmessage) já a traz:

  ```bash
  git config commit.template .gitmessage
  ```

- **Pull requests:** abrem preenchidos com o [template](../.github/pull_request_template.md). O fluxo está em [docs/README.md](README.md#fluxo-de-contribuição).
- **Código:** comentários e documentação em português; complexidade ciclomática ≤ 10 por função; Go formatado com `gofmt`/`goimports` e JavaScript com Prettier.
- **CHANGELOG:** o que o usuário percebe vai em `[Não lançado]`, em linguagem simples.

# Desenvolvimento

> Autores: Caio Reis, Claude

Como preparar o ambiente, rodar o app, testar e contribuir. A arquitetura está em [architecture.md](architecture.md), as regras de negócio em [organization-rules.md](organization-rules.md) e a geração de versões em [release.md](release.md).

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
  backend/                # backend Go (pacotes em architecture.md)
    app/                  # fachada do Wails, drop, janela, composição
    organizer/            # planejamento e execução; criteria/ com um critério por arquivo
    undo/  store/  metadata/  apperr/  logging/
    fs/paths/             # caminhos (extensão, comparação, prefixo do Windows)
    fs/files/             # mover, reservar nome, pastas
    tests/                # testes de integração (só API pública)
  frontend/
    src/                  # React + Tailwind
    wailsjs/              # bindings gerados pelo Wails (versionados)
  build/                  # ícones, manifesto do Windows, Info.plist, NSIS e nfpm
  scripts/
    check-coverage.sh     # cobertura mínima de backend/
    cccases/              # complexidade ciclomática × casos de teste
    benchmark/memory.ps1  # benchmark de memória no Windows (benchmark.md)
    parity/windows.ps1    # roteiro de paridade pela interface no Windows
  docs/                   # esta documentação
  .github/                # CI, release e template de PR
```

### Bindings do frontend

Os métodos públicos de `backend/app.App` viram funções em `frontend/wailsjs/go/app/App.js`. O `wails dev` e o `wails build` regeneram esses arquivos; depois de mudar a assinatura de um método, rode um dos dois e versione o resultado. O frontend não importa os bindings diretamente: tudo passa por `services/sortlyGateway.js` ([ADR 0002](adr/0002-gateway-frontend.md)).

## 4. Testes e qualidade

Todo pull request precisa passar pela CI ([`.github/workflows/ci.yml`](../.github/workflows/ci.yml)).

### 4.1 Complexidade ciclomática (CC)

A complexidade ciclomática conta os caminhos independentes de uma função: começa em 1 e soma 1 para cada `if`, `for`, `case`, `&&`, `||` e `?:`.

- **Limite:** nenhuma função pode ter CC acima de **10**. Se passar, a função deve ser dividida antes do merge.
  - Go: `gocyclo` e `cyclop` (via `golangci-lint`).
  - JavaScript: regra `complexity` do ESLint.
- **Casos de teste:** uma função com CC = N precisa de **pelo menos N casos**, cobrindo os caminhos básicos independentes.
- **Rastreabilidade:** [`scripts/cccases`](../scripts/cccases) cruza a CC de cada função Go com os casos de teste que a chamam e aponta as funções com casos < CC. A CI publica o relatório no resumo do job "Go · lint" (informativo, não bloqueia). No frontend, a contagem é manual (cada `it` ou linha de `it.each`).

  ```bash
  go run ./scripts/cccases            # funções com CC ≥ 2 em backend/
  go run ./scripts/cccases -min 1 backend/undo
  go run ./scripts/cccases -strict    # sai com código 1 se houver casos < CC
  ```

  A contagem é heurística (linhas de tabela percorridas com `range`, `t.Run` e ifs de asserção fora de laço). "Sem teste direto" indica função exercitada só por outras funções, o que nem sempre é problema.

  A tabela CC × casos das funções alteradas vai na seção "Testes" do pull request.

### 4.2 Análise de valor-limite

Os erros costumam aparecer nas fronteiras. Para cada fronteira testamos **o limite − 1, o limite e o limite + 1**.

| Área | Fronteira | Casos |
|---|---|---|
| Tamanho | 1 MB (1 048 576 bytes) | 0 B, 1 B, 1 048 576, 1 048 577, 2 097 152, 2 097 153 |
| Duração | arredondamento e troca de unidade | 0 s, 0,4 s, 0,5 s, 59 s, 60 s, 3 599 s, 3 600 s |
| Data | troca de dia (fuso local) | 23:59:59, 00:00:00, virada de ano, 29/02 |
| Extensão | ponto no nome | `arquivo`, `arquivo.`, `.gitignore`, `A.TAR.GZ` |
| Conflito de nomes | nome ocupado | livre, `(1)`, `(1)` e `(2)` ocupados, limite de tentativas |
| Critérios | pelo menos 1 marcado | 0, 1 e 2 marcados; todos os 6 |
| Cobertura (CI) | 85% | 84,9% falha; 85,0% passa |
| Interface | truncar caminho com mais de 72 caracteres; histórico de 80 notificações | 72 / 73 caracteres; 80 / 81 notificações |

As regras completas, com os valores esperados, estão em [organization-rules.md](organization-rules.md).

### 4.3 Comportamento unitário e integração

- **Go:** testes em tabela (`t.Run`), `t.TempDir()` para tocar no disco e fakes para as interfaces (`FileSystem`, `OperationStore`, …).
- **Integração:** organizar e depois desfazer deve devolver a árvore de arquivos idêntica à original (comparada por hash).
- **Regressão:** todo bug corrigido ganha um teste que falharia no código antigo.
- **Frontend:** Vitest + React Testing Library + jsdom. Os hooks são testados com `renderHook`; o backend é mockado.

### 4.4 Cobertura

- **Mínimo de 85%** nos pacotes Go em `backend/`, verificado por [`scripts/check-coverage.sh`](../scripts/check-coverage.sh).
- O pacote `main` (inicialização do Wails) não entra na meta.

### 4.5 Como rodar localmente

**Go (na raiz)**

```bash
go vet ./...
go test ./...
golangci-lint run ./...          # lint, formatação e complexidade
bash scripts/check-coverage.sh   # cobertura de backend/ (mínimo 85%)
go test -race ./...              # Linux e macOS (no Windows exige gcc)
```

**Frontend (em `frontend/`)**

```bash
npm run lint          # ESLint (react-hooks, complexidade ≤ 10)
npm run format:check  # Prettier (npm run format corrige)
npm test              # Vitest
npm run test:watch    # Vitest em modo observação
```

### 4.6 O que a CI executa

| Job | Sistema | Passos |
|---|---|---|
| Go · lint | Ubuntu | `golangci-lint` (gofmt, goimports, govet, staticcheck, errcheck, gocyclo, cyclop, errorlint, revive, …) e o relatório de `scripts/cccases` no resumo do job (informativo) |
| Go · testes | Ubuntu, Windows, macOS | `go vet`, `go test` (com `-race` no Linux e no macOS), cobertura de `backend/` no Ubuntu |
| Frontend | Ubuntu | `npm ci`, ESLint, Prettier, Vitest, `vite build` |

No Linux, o Wails precisa de `libgtk-3-dev` e `libwebkit2gtk-4.1-dev`, além da build tag `webkit2_41`.

O workflow [`release.yml`](../.github/workflows/release.yml) roda quando uma tag `v*` é criada. Ele gera os executáveis das três plataformas e cria um rascunho de release.

## 5. Dados do app na máquina

| O quê | Onde |
|---|---|
| Registro do último organizar (desfazer) | `~/.sortly/last-operation.json`, o mesmo da versão 1.0 |
| Log | `%AppData%\Sortly\logs` (Windows), `~/Library/Application Support/Sortly/logs` (macOS), `~/.config/Sortly/logs` (Linux) |
| Idioma e critérios | `~/.sortly/settings.json` (some com a pasta `~/.sortly`, não com os dados do WebView) |

Para testar sem mexer nos seus dados, abra o app com outra pasta de usuário. No Windows, defina `USERPROFILE`, `APPDATA` e `LOCALAPPDATA` apontando para uma pasta temporária, como fazem os scripts em `scripts/`.

## 6. Fluxo de contribuição

1. Toda mudança nasce de uma issue. Crie a branch a partir de `main`, ligada à issue para ela fechar sozinha no merge:

   ```bash
   gh issue develop N --base main --name sortly-N-descricao-curta --checkout
   ```

2. Faça commits no padrão `tipo(sortly-N): descrição` (tipos: `feat`, `fix`, `refactor`, `perf`, `test`, `docs`, `ci`, `build`, `chore`). Todo commit leva a linha de coautoria; o modelo [`.gitmessage`](../.gitmessage) já a traz: `git config commit.template .gitmessage`.
3. Rode a skill `/sortly-review` do Claude Code ([`.claude/skills/sortly-review`](../.claude/skills/sortly-review/SKILL.md)) e corrija os achados. Ela roda lint, testes e `scripts/cccases`, percorre checklists de Go, React, testes, design, sistema operacional e segurança e devolve achados por severidade. A revisão humana continua obrigatória.
4. Abra o pull request para `main`. Ele abre preenchido com o [template](../.github/pull_request_template.md): issue (`Closes #N`), tipo, resumo, mudanças por área, defeitos corrigidos com o teste de regressão de cada um, testes (tabela CC × casos e valor-limite), interface (sem mudança visual ou capturas de antes e depois), checklist, riscos e autores. Seções que não se aplicam podem ser removidas; as tabelas de testes são obrigatórias em PRs de código.
5. Atualize `CHANGELOG.md` → `[Não lançado]` quando o usuário perceber a mudança, em linguagem simples.

**Código:** comentários e documentação em português. Comentário explica o *porquê* de um trecho difícil (formato de arquivo, comportamento do sistema operacional, compatibilidade com a 1.0); o que o código já diz não ganha comentário. Identificadores exportados do Go têm godoc de uma linha começando pelo nome (regra `exported` do `revive`). Os testes não trazem tabela de CC: o relatório do `cccases` faz esse papel. Complexidade ciclomática ≤ 10 por função; Go formatado com `gofmt`/`goimports` e JavaScript com Prettier. Decisões de arquitetura com troca real (algo que poderia ter sido diferente) ganham uma ADR em [`adr/`](adr/) (`NNNN-titulo-curto.md`: Contexto, Decisão, Alternativas, Consequências); uma ADR não é editada para mudar de ideia, e sim substituída por outra.

O [`.claude/CLAUDE.md`](../.claude/CLAUDE.md) resume estrutura, comandos e este fluxo para o Claude.

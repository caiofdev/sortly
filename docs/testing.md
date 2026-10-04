# Testes e qualidade

> Autores: Caio Reis, Claude

Todo pull request precisa passar pela CI ([`.github/workflows/ci.yml`](../.github/workflows/ci.yml)). Este documento explica os critérios usados nos testes e como rodar tudo localmente.

## 1. Critérios

### 1.1 Complexidade ciclomática (CC)

A complexidade ciclomática conta os caminhos independentes de uma função: começa em 1 e soma 1 para cada `if`, `for`, `case`, `&&`, `||` e `?:`.

- **Limite:** nenhuma função pode ter CC acima de **10**. Se passar, a função deve ser dividida antes do merge.
  - Go: `gocyclo` e `cyclop` (via `golangci-lint`).
  - JavaScript: regra `complexity` do ESLint.
- **Casos de teste:** uma função com CC = N precisa de **pelo menos N casos**, cobrindo os caminhos básicos independentes.
- **Rastreabilidade:** [`scripts/cccases`](../scripts/cccases) cruza a CC de cada função Go com os casos de teste que a chamam e aponta as funções com casos < CC. A CI publica o relatório no resumo do job "Go · lint" (informativo, não bloqueia). No frontend, a contagem é manual (cada `it` ou linha de `it.each`).

  ```bash
  go run ./scripts/cccases            # funções com CC ≥ 2 em internal/
  go run ./scripts/cccases -min 1 internal/undo
  go run ./scripts/cccases -strict    # sai com código 1 se houver casos < CC
  ```

  A contagem é heurística (linhas de tabela percorridas com `range`, `t.Run` e ifs de asserção fora de laço). "Sem teste direto" indica função exercitada só por outras funções, o que nem sempre é problema. As tabelas `função | CC | casos` nos arquivos de teste estão sendo substituídas por este relatório (#43).

  A tabela CC × casos das funções alteradas vai na seção "Testes" do pull request.

### 1.2 Análise de valor-limite

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

### 1.3 Comportamento unitário e integração

- **Go:** testes em tabela (`t.Run`), `t.TempDir()` para tocar no disco e fakes para as interfaces (`FileSystem`, `OperationStore`, …).
- **Integração:** organizar e depois desfazer deve devolver a árvore de arquivos idêntica à original (comparada por hash).
- **Regressão:** cada defeito B1–B8 de [code-review.md](code-review.md) ganha um teste que falharia no código antigo.
- **Frontend:** Vitest + React Testing Library + jsdom. Os hooks são testados com `renderHook`; o backend é mockado.

### 1.4 Cobertura

- **Mínimo de 85%** nos pacotes Go em `internal/`, verificado por [`scripts/check-coverage.sh`](../scripts/check-coverage.sh).
- O pacote `main` (inicialização do Wails) não entra na meta.

## 2. Como rodar localmente

### Go (na raiz)

```bash
go vet ./...
go test ./...
golangci-lint run ./...          # lint, formatação e complexidade
bash scripts/check-coverage.sh   # cobertura de internal/ (mínimo 85%)
go test -race ./...              # Linux e macOS (no Windows exige gcc)
```

### Frontend (em `frontend/`)

```bash
npm run lint          # ESLint (react-hooks, complexidade ≤ 10)
npm run format:check  # Prettier (npm run format corrige)
npm test              # Vitest
npm run test:watch    # Vitest em modo observação
```

## 3. O que a CI executa

| Job | Sistema | Passos |
|---|---|---|
| Go · lint | Ubuntu | `golangci-lint` (gofmt, goimports, govet, staticcheck, errcheck, gocyclo, cyclop, errorlint, …) |
| Go · testes | Ubuntu, Windows, macOS | `go vet`, `go test` (com `-race` no Linux e no macOS), cobertura de `internal/` no Ubuntu |
| Frontend | Ubuntu | `npm ci`, ESLint, Prettier, Vitest, `vite build` |

No Linux, o Wails precisa de `libgtk-3-dev` e `libwebkit2gtk-4.1-dev`, além da build tag `webkit2_41`.

O workflow [`release.yml`](../.github/workflows/release.yml) roda quando uma tag `v*` é criada. Ele gera os executáveis das três plataformas e cria um rascunho de release.

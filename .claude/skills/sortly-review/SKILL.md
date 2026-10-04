---
name: sortly-review
description: Code review do Sortly (Go + Wails + React) antes de abrir ou aprovar um PR. Cobre padrões de Go e React usados no mercado, testes (complexidade ciclomática e valor-limite), padrões de projeto, sistema de arquivos/SO, segurança e as decisões do projeto (ADRs). Use quando pedirem "revise", "code review", "/sortly-review", antes de abrir um PR do Sortly, ou ao revisar um PR/branch/diff deste repositório.
---

# Code review do Sortly

Objetivo: encontrar o que uma revisão humana costuma deixar passar — casos de borda, comportamento do sistema operacional, concorrência, regras do projeto quebradas em silêncio — e reportar só o que se sustenta com um cenário concreto. O Caio revisa depois; esta revisão não substitui a dele, ela a antecede.

## Argumentos

- Sem argumento: revisa o diff da branch atual contra `main` (`git diff main...HEAD`) mais as mudanças não commitadas.
- Número de PR (`#37`, `37`): `gh pr diff N` e `gh pr view N`.
- Branch, commit ou caminho: revisa esse alvo.
- `corrigir`: depois do relatório, aplica as correções dos achados confirmados (com testes) e roda as verificações de novo.

## Processo

1. **Contexto.** Leia o `.claude/CLAUDE.md`, a issue ligada à branch (`gh issue view N`), o diff inteiro e, para cada arquivo alterado, o arquivo completo e o teste dele. Se o diff tocar regras de organização ou o registro do desfazer, leia `docs/organization-rules.md`; se tocar uma decisão de arquitetura, a ADR correspondente.

2. **Verificações objetivas.** Rode o que se aplica ao diff e transforme cada falha em achado:
   - Go: `golangci-lint run ./...`, `go test ./...`, `bash scripts/check-coverage.sh`
   - `go run ./scripts/cccases` — toda função **alterada ou nova** com `⚠ casos < CC` é achado; `↪ sem teste direto` é achado só se a função tiver lógica própria que nenhum teste exercita (confira a cobertura com `go test -coverprofile`).
   - Frontend (se tocado): `cd frontend && npm run lint && npm run format:check && npm test`
   - Em JS, conte os casos à mão: cada `it`/linha de `it.each` que chama a função.

3. **Revisão por área.** Carregue só as referências das áreas tocadas e percorra os checklists:
   - Código Go → [references/go.md](references/go.md)
   - Código React/JS → [references/react.md](references/react.md)
   - Qualquer teste ou código sem teste → [references/testes.md](references/testes.md)
   - Estrutura, novos tipos/pacotes, refatorações → [references/design.md](references/design.md)
   - Arquivos, caminhos, mover/renomear, registro em disco → [references/sistema-operacional.md](references/sistema-operacional.md)
   - Entrada do frontend, leitura de arquivos do usuário (mp4, pdf, zip) → [references/seguranca.md](references/seguranca.md)
   - Sempre: [references/aprendizados.md](references/aprendizados.md) — erros que já passaram antes.

4. **Confirmação.** Para cada suspeita, escreva o cenário concreto: entrada/estado → resultado errado. Se não conseguir, descarte. Quando for barato, prove com um teste que falha (não o deixe no código sem pedir). Não reporte estilo que o lint já cobre nem preferência pessoal sem custo concreto.

5. **Relatório** no formato abaixo. Sem achados, diga isso e liste o que foi verificado.

## Severidade

| Nível | Quando |
|---|---|
| 🔴 Crítico | Perda ou corrupção de dados do usuário, desfazer impossível, crash, falha de segurança |
| 🟠 Alto | Resultado errado num caso plausível, regra do projeto ou ADR violada, teste que não testa o que diz |
| 🟡 Médio | Caso de borda raro mal tratado, casos < CC, falta de teste de regressão, acoplamento que vai doer na próxima mudança |
| 🔵 Baixo | Clareza, nome, duplicação pequena, comentário desnecessário |
| 💡 Sugestão | Melhoria opcional, sem defeito |

## Formato do relatório

```markdown
## Revisão: <alvo>

**Verificações:** lint ✅ · testes ✅ · cobertura 91% ✅ · cccases: 1 função com casos < CC ⚠ · frontend —

| # | Sev. | Arquivo | Achado |
|---|---|---|---|
| 1 | 🔴 | internal/undo/undo.go:142 | Desfazer apaga pasta com arquivo do usuário |

### 1. 🔴 Desfazer apaga pasta com arquivo do usuário — `internal/undo/undo.go:142`
**Cenário:** organizar → usuário cria `notas.txt` em `pdf/` → desfazer → `pdf/` é removida com o arquivo.
**Por quê:** `RemoveAll` em vez de remover só pasta vazia.
**Correção:** usar `fsutil.RemoveEmptyDir`; teste de regressão com arquivo extra na pasta.
```

Ordene do mais grave para o menos grave. Cite `arquivo:linha`. Uma correção por achado, curta.

## Aprimoramento contínuo

Quando o Caio (ou um bug em produção) apontar algo que esta revisão deixou passar, acrescente em `references/aprendizados.md` uma regra **generalizada** (não o caso específico), com a data e a origem (PR/issue). Se a regra pertencer a uma área, acrescente também no checklist da área. Revise a lista quando ela crescer: junte duplicadas, remova o que virou lint.

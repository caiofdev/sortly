<!--
Título do PR: tipo(sortly-N): descrição curta
Ex.: feat(sortly-12): adiciona drag and drop nativo
-->

## 🔗 Issue

Closes #<!-- número -->

## 🏷️ Tipo de mudança

- [ ] `feat` — nova funcionalidade
- [ ] `fix` — correção de defeito
- [ ] `refactor` — melhoria interna sem mudança de comportamento
- [ ] `test` — testes
- [ ] `docs` — documentação
- [ ] `ci` / `build` / `chore` — infraestrutura e manutenção

## 📝 Resumo

<!-- O que este PR faz e por quê, em 2 a 4 frases. -->

## 🔧 Mudanças

<!-- Lista objetiva do que mudou, agrupada por área. -->

**Backend (Go)**
-

**Frontend (React)**
-

**Docs / Infra**
-

## 🐞 Defeitos corrigidos

<!-- Se aplicável, cite os achados do docs/code-review.md (B1–B7) e o teste de regressão de cada um. Apague a seção se não houver. -->

| ID | Correção | Teste de regressão |
|---|---|---|
| | | |

## 🧪 Testes

### Complexidade ciclomática

<!-- Uma linha por função nova ou alterada. Casos ≥ CC; CC ≤ 10. -->

| Função | CC | Casos |
|---|---|---|
| | | |

### Valor-limite

<!-- Fronteiras testadas (limite − 1, limite, limite + 1). -->

| Fronteira | Casos |
|---|---|
| | |

### Como verificar

```bash
go test ./...
npm test
```

<!-- Passos manuais, se houver (ex.: wails dev → arrastar uma pasta → organizar → desfazer). -->

- [ ] Testes automatizados passando localmente
- [ ] Cobertura de `internal/...` ≥ 85%
- [ ] Verificação manual feita (descrita acima)

## 🖼️ Interface

- [ ] Sem mudança visual
- [ ] Com mudança visual (antes / depois abaixo)

<!--
| Antes | Depois |
|---|---|
| ![antes]() | ![depois]() |
-->

## ✅ Checklist

- [ ] Commits no padrão `tipo(sortly-N): descrição`
- [ ] `golangci-lint` e ESLint sem erros (complexidade ≤ 10)
- [ ] `docs/` atualizado quando aplicável (arquitetura, regras, ADR)
- [ ] `CHANGELOG.md` → `[Não lançado]` atualizado, se o usuário percebe a mudança
- [ ] Sem código morto, `console.log` ou TODOs esquecidos
- [ ] Compatibilidade mantida (nomes de pastas, formato do `last-operation.json`)

## ⚠️ Riscos e observações

<!-- Pontos de atenção para quem revisa, decisões em aberto, trabalho futuro. -->

## 👥 Autores

- Caio Reis (@caiofdev)
- Claude (coautor)

---

🤖 Generated with [Claude Code](https://claude.com/claude-code)

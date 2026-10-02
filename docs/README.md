# Documentação do Sortly

> Autores: Caio Reis, Claude

Índice da documentação técnica do projeto. O [README da raiz](../README.md) apresenta o app para usuários; aqui ficam os detalhes para quem desenvolve.

## Visão geral

| Documento | Conteúdo |
|---|---|
| [architecture.md](architecture.md) | Arquitetura atual (Electron) e arquitetura alvo (Wails v2 + Go + React), com diagramas |
| [organization-rules.md](organization-rules.md) | Regras de negócio: critérios de organização, nomes de pastas, conflitos e desfazer |
| [code-review.md](code-review.md) | Revisão do código atual: defeitos (B1–B7) e refatorações planejadas |
| [testing.md](testing.md) | Estratégia de testes (complexidade ciclomática e valor-limite), CI e como rodar localmente |

## Decisões de arquitetura (ADRs)

| ADR | Título | Status |
|---|---|---|
| [0001](adr/0001-electron-para-wails.md) | Migrar de Electron para Wails v2 | Aceita |
| [0002](adr/0002-gateway-frontend.md) | Gateway no frontend para acesso ao backend | Aceita |
| [0003](adr/0003-strategy-regras.md) | Strategy + Registry para os critérios de organização | Aceita |
| [0004](adr/0004-erros-com-codigo.md) | Erros do backend identificados por código | Aceita |

Novas decisões seguem o mesmo formato (`adr/NNNN-titulo-curto.md`): Contexto, Decisão, Alternativas consideradas, Consequências.

## Fluxo de contribuição

1. Escolha uma issue da milestone e crie a branch a partir de `wails-rewrite`: `sortly-N-descricao-curta`.
2. Faça commits no padrão `tipo(sortly-N): descrição`. O modelo em [`.gitmessage`](../.gitmessage) já traz a linha de coautoria; para ativá-lo, rode `git config commit.template .gitmessage`.
3. Abra o pull request para `wails-rewrite`. Ele abre preenchido com o [template](../.github/pull_request_template.md). Preencha:
   - **Issue** (`Closes #N`) e **tipo de mudança**;
   - **Resumo** e **mudanças** por área (Backend Go, Frontend React, Docs/Infra);
   - **Defeitos corrigidos** (B1–B7 de [code-review.md](code-review.md)), com o teste de regressão de cada um;
   - **Testes**: tabela `Função | CC | Casos` (nº de casos ≥ complexidade ciclomática, CC ≤ 10) e tabela de valor-limite;
   - **Interface**: sem mudança visual, ou capturas de antes e depois;
   - **Checklist**, **riscos** e **autores**.
4. Atualize `CHANGELOG.md` → `[Não lançado]` quando o usuário perceber a mudança.

Seções que não se aplicam podem ser marcadas como "Não se aplica" ou removidas. As tabelas de testes, porém, são obrigatórias em PRs de código.

## Em breve

Os documentos abaixo serão criados conforme as issues da milestone [Wails rewrite](https://github.com/caiofdev/sortly/milestone/1) avançam:

- `development.md` — pré-requisitos, comandos de desenvolvimento e convenções de commit
- `release.md` — empacotamento por plataforma, versionamento e changelog

## Imagens

Capturas de tela e GIFs usados no README ficam em [`images/`](images/).

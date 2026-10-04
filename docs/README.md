# Documentação do Sortly

> Autores: Caio Reis, Claude

Índice da documentação técnica do projeto. O [README da raiz](../README.md) apresenta o app para usuários; aqui ficam os detalhes para quem desenvolve.

## Visão geral

| Documento | Conteúdo |
|---|---|
| [architecture.md](architecture.md) | Arquitetura (Wails v2 + Go + React), contrato com o frontend e a versão 1.0 (Electron) como referência |
| [development.md](development.md) | Pré-requisitos, como rodar o app, estrutura do repositório e convenções |
| [organization-rules.md](organization-rules.md) | Regras de negócio: critérios de organização, nomes de pastas, conflitos e desfazer |
| [benchmark.md](benchmark.md) | Memória, velocidade e tamanho: Electron 1.0 × Wails 2.0 (Windows) |
| [checklist-paridade.md](checklist-paridade.md) | Roteiro de paridade por plataforma, com os resultados no Windows |
| [code-review.md](code-review.md) | Revisão do código da versão 1.0: defeitos (B1–B8) e refatorações aplicadas na 2.0 |
| [release.md](release.md) | Versionamento, como gerar uma versão e o que cada pacote (Windows, macOS, Linux) contém |
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

1. Escolha uma issue e crie a branch a partir de `main`, ligada à issue: `gh issue develop N --base main --name sortly-N-descricao-curta --checkout`.
2. Faça commits no padrão `tipo(sortly-N): descrição`. O modelo em [`.gitmessage`](../.gitmessage) já traz a linha de coautoria; para ativá-lo, rode `git config commit.template .gitmessage`.
3. Abra o pull request para `main`. Ele abre preenchido com o [template](../.github/pull_request_template.md). Preencha:
   - **Issue** (`Closes #N`) e **tipo de mudança**;
   - **Resumo** e **mudanças** por área (Backend Go, Frontend React, Docs/Infra);
   - **Defeitos corrigidos** (B1–B8 de [code-review.md](code-review.md)), com o teste de regressão de cada um;
   - **Testes**: tabela `Função | CC | Casos` (nº de casos ≥ complexidade ciclomática, CC ≤ 10) e tabela de valor-limite;
   - **Interface**: sem mudança visual, ou capturas de antes e depois;
   - **Checklist**, **riscos** e **autores**.
4. Atualize `CHANGELOG.md` → `[Não lançado]` quando o usuário perceber a mudança.

Seções que não se aplicam podem ser marcadas como "Não se aplica" ou removidas. As tabelas de testes, porém, são obrigatórias em PRs de código.

## Imagens

Capturas de tela e GIFs usados no README ficam em [`images/`](images/).

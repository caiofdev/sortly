# Documentação do Sortly

> Autores: Caio Reis, Claude

Índice da documentação técnica do projeto. O [README da raiz](../README.md) apresenta o app para usuários; aqui ficam os detalhes para quem desenvolve.

## Visão geral

| Documento | Conteúdo |
|---|---|
| [architecture.md](architecture.md) | Arquitetura atual (Electron) e arquitetura alvo (Wails v2 + Go + React), com diagramas |
| [organization-rules.md](organization-rules.md) | Regras de negócio: critérios de organização, nomes de pastas, conflitos e desfazer |
| [code-review.md](code-review.md) | Revisão do código atual: defeitos (B1–B7) e refatorações planejadas |

## Decisões de arquitetura (ADRs)

| ADR | Título | Status |
|---|---|---|
| [0001](adr/0001-electron-para-wails.md) | Migrar de Electron para Wails v2 | Aceita |
| [0002](adr/0002-gateway-frontend.md) | Gateway no frontend para acesso ao backend | Aceita |
| [0003](adr/0003-strategy-regras.md) | Strategy + Registry para os critérios de organização | Aceita |
| [0004](adr/0004-erros-com-codigo.md) | Erros do backend identificados por código | Aceita |

Novas decisões seguem o mesmo formato (`adr/NNNN-titulo-curto.md`): Contexto, Decisão, Alternativas consideradas, Consequências.

## Em breve

Os documentos abaixo serão criados conforme as issues da milestone [Wails rewrite](https://github.com/caiofdev/sortly/milestone/1) avançam:

- `testing.md` — estratégia de testes (complexidade ciclomática e valor-limite), como rodar e metas de cobertura
- `development.md` — pré-requisitos, comandos de desenvolvimento e convenções de commit
- `release.md` — empacotamento por plataforma, versionamento e changelog

## Imagens

Capturas de tela e GIFs usados no README ficam em [`images/`](images/).

# ADR 0004 — Erros do backend identificados por código

- **Status:** Aceita
- **Data:** 2026-10-02
- **Autores:** Caio Reis, Claude

## Contexto

Na versão Electron, o backend lança erros com texto fixo, em português em alguns casos ("Pasta inválida.") e em inglês em outros ("Invalid dropped item."). A interface exibe `error.message` diretamente. Com o idioma em inglês, o usuário vê mensagens em português; com o idioma em português, às vezes vê mensagens em inglês. O backend também devolve um campo `message` nos resultados que a interface ignora (achados B6 e B7 em [code-review.md](../code-review.md)).

## Decisão

- O backend **não produz texto para o usuário**. Ele devolve dados e erros identificados por **código estável**:

  | Código | Situação |
  |---|---|
  | `INVALID_SOURCE` | Pasta de origem ausente ou inválida |
  | `INVALID_DESTINATION` | Pasta de destino inválida |
  | `NO_CRITERIA` | Nenhum critério de organização ativo |
  | `NOTHING_TO_UNDO` | Não há organização para desfazer |
  | `DROPPED_INVALID` | Item arrastado inválido |
  | `DROPPED_MISSING` | Item arrastado não existe mais |
  | `DROPPED_UNSUPPORTED` | Item arrastado não é arquivo nem pasta |
  | `RECORD_NOT_SAVED` | Arquivos movidos, mas o registro para desfazer não foi gravado |
  | `UNEXPECTED` | Erro inesperado (detalhes vão para o log) |

- Em Go, são erros sentinela criados com `apperr.New(código, mensagem)` (`internal/apperr`), verificados com `errors.Is`. `apperr.CodeOf(err)` devolve o código para o frontend (`UNEXPECTED` quando não há).
- O gateway do frontend ([ADR 0002](0002-gateway-frontend.md)) converte o código no texto do idioma atual, via `feedbackCopy`.
- Os textos em português continuam **idênticos** aos atuais.
- O campo `message` sai dos resultados de `organizeFiles` e `undoLastOrganization`.

## Consequências

- Todas as mensagens passam a respeitar o idioma escolhido.
- Os testes do backend verificam códigos, não textos.
- É preciso manter a tabela de códigos sincronizada entre Go e `feedbackCopy`; um teste no frontend garante que todo código tenha tradução nos dois idiomas.

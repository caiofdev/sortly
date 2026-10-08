# Testes

## Complexidade ciclomática (basis path testing)

CC = 1 + `if` + `for` + `range` + `case` não-default + `&&` + `||` (definição do gocyclo; `scripts/cccases` usa a mesma).

- [ ] Nenhuma função passa de CC 10 (lint).
- [ ] Toda função **nova ou alterada** tem casos ≥ CC (`go run ./scripts/cccases`). Os casos precisam percorrer **caminhos diferentes** — dez entradas que passam pelo mesmo `if` contam como uma para esse caminho.
- [ ] Para cada `if`, existe um caso que entra e um que não entra; para cada `&&`/`||`, um caso em que cada operando decide o resultado (cobertura de condição, não só de decisão).
- [ ] Mapa de lookup (o projeto não usa `switch`): um caso por chave que muda o resultado e um para a chave ausente; teste que as chaves do mapa batem com a lista pública (ex.: `criteria.Keys`).
- [ ] Laço: zero, uma e várias iterações.
- [ ] Caminho de erro de cada chamada que pode falhar (use fakes que falham: `failOn`, `cancelAfter`).

## Análise de valor-limite

Para cada fronteira: **limite − 1, limite, limite + 1**. Fronteiras do Sortly:

| Área | Fronteira |
|---|---|
| Tamanho | 0 B, 1 B, 1 048 576, 1 048 577, 2 097 152, 2 097 153 |
| Duração | 0 s, 0,4 / 0,5 s (arredondamento), 59 / 60 s, 3 599 / 3 600 s |
| Data | 23:59:59 / 00:00:00 no fuso local, virada de ano, 29/02 |
| Extensão | `arquivo`, `arquivo.`, `.gitignore`, `A.TAR.GZ`, maiúsculas |
| Conflito de nomes | livre, `(1)` ocupado, `(1)` e `(2)` ocupados, limite de tentativas |
| Critérios | 0, 1, 2 e 6 marcados; desmarcar o último |
| Notificações | 80 / 81 |
| Caminho exibido | 72 / 73 caracteres |
| Cobertura | 84,9% / 85,0% |

- [ ] Fronteira nova introduzida pelo diff (constante, limite, `<` vs `<=`) tem os três casos.
- [ ] Off-by-one: confira `<` vs `<=`, `len-1`, índices em fatias, arredondamento (`math.Round` vs truncar).

## Partição de equivalência

- [ ] Uma entrada representativa de cada classe: válida, inválida, vazia, nil, enorme, com Unicode, com espaços, com separador de caminho.

## Qualidade dos testes

- [ ] O teste falharia se o código estivesse errado? (Asserção fraca: só `err == nil`, só tamanho da lista, `want` calculado com a mesma função testada.)
- [ ] Nome/descrição do caso diz o cenário, e a mensagem de falha mostra entrada, obtido e esperado.
- [ ] Determinístico: sem depender de hora atual, fuso, locale, ordem de mapa, ordem de `ReadDir` em outros SOs, rede, `sleep`.
- [ ] Hermético: `t.TempDir()`, nada em `~/.sortly` ou pastas reais; variáveis de ambiente com `t.Setenv`.
- [ ] `t.Helper()` em helpers; `t.Fatal` só quando continuar não faz sentido.
- [ ] Teste específico de SO com build tag ou `runtime.GOOS` + `t.Skip` explicando por quê. Arquivo `!windows` não compila na máquina Windows: rode `GOOS=linux golangci-lint run` e `GOOS=linux go test -c` e, se houver WSL, execute o binário lá.
- [ ] Fixture de formato antigo (registro da 1.0, preferências) é escrita **crua** no teste, não pelo código que normaliza: o `store` grava `createdFolders: []` mesmo com nil, então salvar pelo store não simula um arquivo sem o campo.
- [ ] Fakes simples e explícitos em vez de mocks com expectativas frágeis.
- [ ] Golden files para formatos que precisam ser idênticos (registro JSON da 1.0).
- [ ] Bug corrigido → teste de regressão que falha no código antigo.
- [ ] Integração: organizar → desfazer devolve a árvore idêntica (hash), inclusive com falha no meio.

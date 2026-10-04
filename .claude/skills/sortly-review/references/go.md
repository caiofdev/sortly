# Go

Base: [Effective Go](https://go.dev/doc/effective_go), [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments), [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md), [Google Go Style](https://google.github.io/styleguide/go/).

## Erros

- [ ] Todo erro é tratado ou propagado; nada de `_ =` em erro que importa (Close de arquivo **escrito** importa: o erro de flush aparece no Close).
- [ ] Contexto ao propagar: `fmt.Errorf("mover %s: %w", path, err)` — com `%w`, nunca `%v`, para `errors.Is/As` funcionar.
- [ ] Comparação com `errors.Is` / `errors.As`, nunca `==` em erro embrulhado nem comparação de `err.Error()`.
- [ ] Erros que chegam ao frontend saem com código de `apperr` (ADR 0004); erro sem código vira `UNEXPECTED` e precisa ir para o log com detalhe.
- [ ] Sem `panic` em caminho de erro esperado; `panic` só para invariante de programação.
- [ ] Mensagens de erro em minúsculas, sem pontuação final (convenção Go).
- [ ] Erro tratado uma vez: ou loga, ou devolve — não os dois (log duplicado).

## Recursos e defer

- [ ] Todo `Open`/`Create` tem `Close`; arquivo aberto para escrita verifica o erro do `Close`.
- [ ] `defer` dentro de laço segura os recursos até a função acabar — extraia uma função.
- [ ] `defer f.Close()` antes de checar o erro do `Open` é bug (f nil).
- [ ] Arquivo temporário removido em todo caminho de erro.

## Concorrência

- Os bindings do Wails rodam em goroutines separadas: dois cliques rápidos chamam o método duas vezes **em paralelo**.
- [ ] Estado compartilhado em `App`/serviços protegido (mutex) ou imutável; organizar e desfazer não rodam juntos.
- [ ] Goroutine iniciada tem fim garantido (contexto, canal fechado); sem vazamento.
- [ ] `context.Context` é o primeiro parâmetro, propagado, nunca guardado em struct (exceto o ctx do Wails no `startup`); laços longos checam `ctx.Err()`.
- [ ] Mapas acessados de várias goroutines estão protegidos.
- [ ] Testes com `-race` passam (CI no Linux e macOS).

## APIs e tipos

- [ ] Interfaces pequenas, declaradas no pacote que **consome**; construtores devolvem tipos concretos ("accept interfaces, return structs").
- [ ] Dependências por construtor (DI); nada de estado global mutável nem `init()` com efeito colateral.
- [ ] Zero value útil ou construtor obrigatório documentado.
- [ ] Receptor consistente no tipo (todos ponteiro ou todos valor); ponteiro quando muta ou o struct é grande.
- [ ] Slices/mapas recebidos e guardados são copiados se o chamador puder alterá-los depois.
- [ ] `nil` slice vs vazio: o JSON do registro exige `[]` (veja `store.encode`).
- [ ] Exportar só o necessário; identificador exportado tem godoc começando pelo nome.
- [ ] Nomes curtos e sem gagueira (`store.Store` ruim, `store.FileStore` ok); sem pacotes `util`/`common`/`helpers`.
- [ ] Sem parâmetro booleano que muda o comportamento da função (prefira duas funções ou opções).

## Desempenho (pasta grande: 10 000+ arquivos)

- [ ] Nada O(n²) sobre a lista de arquivos (busca linear dentro de laço → mapa).
- [ ] Não lê o arquivo inteiro na memória para extrair metadado (mp4, pdf, zip): leia cabeçalhos/trechos.
- [ ] `append` em laço com tamanho conhecido usa `make([]T, 0, n)`.
- [ ] Nenhuma chamada de sistema repetida sem necessidade (`Stat` duas vezes no mesmo arquivo).

## Wails

- [ ] Métodos públicos de `App` viram bindings: não exponha método sem querer (o Wails gera JS para todos).
- [ ] Mudou assinatura de binding → `frontend/wailsjs` regenerado e versionado, gateway/frontend atualizados.
- [ ] Nada bloqueia o `startup`/`domReady` (a janela fica branca).
- [ ] Opções da janela por plataforma: Linux exige a tag `webkit2_41`; `options.Linux` não nil desliga padrões (ex.: política de GPU).

## Formatação e lint

`gofmt`, `goimports`, `errorlint`, `gocyclo`/`cyclop` ≤ 10, `unparam`, `unconvert` e o conjunto padrão rodam na CI — não reporte o que eles já pegam; reporte quando o código **desliga** um lint (`//nolint`) sem justificativa.

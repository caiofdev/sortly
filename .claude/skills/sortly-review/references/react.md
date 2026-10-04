# React e JavaScript

Base: [react.dev](https://react.dev) (especialmente "You Might Not Need an Effect" e "Rules of Hooks"), [Testing Library — guiding principles](https://testing-library.com/docs/guiding-principles).

## Papel do frontend no Sortly

- [ ] O frontend renderiza e traduz; regra de negócio (validação, decisão, limites, estado da organização) mora no Go (#45). Lógica nova no JS é achado 🟠, salvo apresentação pura.
- [ ] Acesso ao backend só pelo gateway (`services/sortlyGateway.js`), nunca `window.go` direto em componente.
- [ ] Textos vêm do i18n (`i18n/*`), em PT e EN; nenhum texto fixo em componente; chaves novas existem nos dois idiomas.
- [ ] Erro do backend traduzido pelo código (`describeError`), nunca exibindo `error.message` cru.

## Hooks

- [ ] Hooks só no topo do componente/hook, nunca em condição ou laço.
- [ ] Dependências de `useEffect`/`useMemo`/`useCallback` completas (o lint `react-hooks/exhaustive-deps` está ligado; `eslint-disable` sem motivo é achado).
- [ ] Efeito que assina algo (evento do Wails, timer, listener) devolve a função de limpeza — senão a assinatura duplica a cada montagem (StrictMode monta duas vezes em dev).
- [ ] Efeito usado para **derivar** estado de props/estado é desnecessário: calcule durante o render.
- [ ] Efeito que reage a evento do usuário deveria estar no handler.
- [ ] Atualização baseada no valor anterior usa a forma funcional (`setX(prev => ...)`).
- [ ] Promise resolvida depois de desmontar não atualiza estado (ou é ignorada com segurança).

## Estado

- [ ] Uma fonte de verdade: nada duplicado entre estado local, props e backend.
- [ ] Estado mínimo; o resto é derivado.
- [ ] Objetos/arrays de estado não são mutados (`push`, atribuição direta): crie novos.
- [ ] `key` estável e única em listas (id, não índice quando a lista muda de ordem/tamanho).

## Componentes e acessibilidade

- [ ] Componente com uma responsabilidade; props explícitas, sem espalhar `{...props}` desconhecidas.
- [ ] Elemento clicável é `<button>` (não `div` com `onClick`); botão só com ícone tem `aria-label`.
- [ ] Inputs com `<label>` associado; checkbox controlado (`checked` + `onChange`).
- [ ] Foco visível e navegação por teclado funcionando nos painéis (configurações, notificações).
- [ ] Mudança visual não pedida é achado (a interface é congelada sem pedido explícito).

## JavaScript

- [ ] `===`; sem coerção implícita surpreendente (`'' || padrão` engole string vazia válida? `??` quando só null/undefined devem cair no padrão).
- [ ] `async` sem `try/catch` onde a rejeição precisa virar notificação.
- [ ] Datas/números formatados com `toLocale` do idioma escolhido, não do sistema.
- [ ] Sem código morto, `console.log` ou export não usado.

## Testes (Vitest + Testing Library)

- [ ] Consultas por papel/texto (`getByRole`, `getByText`), não por classe ou estrutura.
- [ ] Testa comportamento visível, não detalhe de implementação (estado interno, nome de função).
- [ ] Gateway falso injetado; nenhum teste depende de `window.go` real.
- [ ] `await`/`waitFor` em toda atualização assíncrona (sem warnings de `act`).
- [ ] Limites: último critério marcado, caminho de 72/73 caracteres, 80/81 notificações.

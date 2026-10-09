# React e JavaScript

Base: [react.dev](https://react.dev) (especialmente "You Might Not Need an Effect" e "Rules of Hooks"), [Testing Library — guiding principles](https://testing-library.com/docs/guiding-principles).

## Papel do frontend no Sortly

- [ ] O frontend renderiza e traduz; regra de negócio (validação, decisão, limites, estado da organização) mora no Go (ADR 0005). Lógica nova no JS é achado 🟠, salvo apresentação pura.
- [ ] Acesso ao backend só pelo `hooks/useViewState.js`, o único módulo que importa `wailsjs/` (bindings e runtime); nunca `window.go` direto em componente. Fora do Wails (`npm run dev`), o hook fica no estado inicial.
- [ ] Textos vêm do i18n (`i18n/*`), em PT e EN; nenhum texto fixo em componente; chaves novas existem nos dois idiomas.
- [ ] Notificações traduzidas pelo código em `i18n/notifications.js`; código novo do backend tem texto em PT e EN (ou cai no texto padrão da ação).

## Design system (ADR 0006)

- [ ] Estilo só com as classes `st-*` de `styles/components.css` (cópia do `bundle.css` do design system, sem edição) e os tokens de `styles/tokens.css`. Cor, espaço, raio ou fonte fixos (`#F5E600`, `16px`, `style={{ color: "#fff" }}`) são achado 🟡: quebram o tema claro e divergem do design system. `style` só com token (`var(--brand)`) ou valor calculado (largura da barra do Concluído, posição no FileFlow).
- [ ] Ajuste que o design system não cobre vai em `styles/app.css`, com um comentário do porquê; nunca em `components.css`.
- [ ] Sem Tailwind, CSS-in-JS ou biblioteca de ícones: ícone novo é SVG inline em `components/icons.jsx` (traço 2px, `currentColor`, `aria-hidden`).
- [ ] Mudança visual confere nos dois temas (`data-theme="light"` no `<html>`) e nas larguras da janela padrão e mínima.
- [ ] A fonte Pixelify Sans vem embutida (`@fontsource`); `@import` remoto (Google Fonts) não funciona offline.

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
- [ ] Imagem decorativa ao lado de texto tem `alt=""`: um `alt` igual ao texto faz o leitor de tela repetir o rótulo ("PT-BR PT-BR").
- [ ] Botões de alternância (idioma, filtros) informam o estado com `aria-pressed`. Isso muda o padrão de UI Automation de Invoke para Toggle: confira o roteiro de paridade.
- [ ] Eventos de arrastar: o WebKit (macOS e Linux) costuma entregar `relatedTarget` vazio; para "saiu do painel", conte `dragenter`/`dragleave`.
- [ ] Inputs com `<label>` associado; checkbox controlado (`checked` + `onChange`).
- [ ] Foco visível e navegação por teclado funcionando nos painéis (configurações, notificações).
- [ ] Toast, painel ou outro elemento fixo não cobre Organizar/Desfazer na janela padrão (980×700) nem na mínima (820×600); o que some sozinho para com o mouse ou o foco em cima.
- [ ] Mudança visual não pedida é achado (a interface é congelada sem pedido explícito).

## JavaScript

- [ ] `===`; sem coerção implícita surpreendente (`'' || padrão` engole string vazia válida? `??` quando só null/undefined devem cair no padrão).
- [ ] `async` sem `try/catch` onde a rejeição precisa virar notificação.
- [ ] Datas/números formatados com `toLocale` do idioma escolhido, não do sistema.
- [ ] Sem código morto, `console.log` ou export não usado.

## Testes (Vitest + Testing Library)

- [ ] Consultas por papel/texto (`getByRole`, `getByText`), não por classe ou estrutura.
- [ ] Testa comportamento visível, não detalhe de implementação (estado interno, nome de função).
- [ ] Um arquivo de teste por módulo, ao lado dele; `frontend/tests/` só para integração (App inteiro).
- [ ] Teste do `useViewState` com `vi.mock` dos bindings e do runtime; só a integração usa `window.go`/`window.runtime` falsos.
- [ ] Casos ≥ CC também no JS: liste a CC com a API do ESLint (regra `complexity` com máximo 1; cada ternário, `&&`, `||` e `??` conta) e compare com os `it`/linhas de `it.each` que exercitam a função. Não estime de cabeça.
- [ ] `await`/`waitFor` em toda atualização assíncrona (sem warnings de `act`).
- [ ] Limites: último critério marcado, 80/81 notificações, toast em 3 999 / 4 000 ms, 7/8 pastas na prévia (o resto vira "outras").

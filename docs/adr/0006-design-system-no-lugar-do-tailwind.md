# ADR 0006 — CSS do design system "Sortly" no lugar do Tailwind

- **Status:** Aceita
- **Data:** 2026-10-08
- **Autores:** Caio Reis, Claude

## Contexto

O redesign (milestone "Redesign", #74–#84) parte de um design system feito com o Claude a partir da logo nova: tokens de cor para os temas escuro e claro, escala tipográfica com a fonte pixel Pixelify Sans e componentes prontos em CSS puro (`st-btn`, `st-drop`, `st-path`, `st-nav`, `st-notif`…), que usam só variáveis CSS. A interface anterior era escrita com classes utilitárias do Tailwind e cores fixas em cada componente (`bg-[#0F172A]`, `text-[#94A3B8]`), com a fonte Inter.

## Decisão

- O frontend usa o CSS do design system como está: `styles/tokens.css` (variáveis dos dois temas; o claro em `[data-theme="light"]` no `<html>`) e `styles/components.css` (cópia do `bundle.css`, sem o `@import` remoto da fonte). Ajustes do app ficam em `styles/app.css`.
- `components.css` não passa pelo Prettier, para continuar comparável com o original do design system.
- A Pixelify Sans vem embutida pelo `@fontsource/pixelify-sans` (OFL), como a Inter antes: o app abre offline e a fonte é a mesma nos três WebViews.
- O Tailwind, o PostCSS/Autoprefixer e a Inter saem do projeto.
- Ícones são SVG inline (`components/icons.jsx`), no traço do design system, sem biblioteca nova.

## Alternativas consideradas

| Alternativa | Por que não |
|---|---|
| Manter o Tailwind e mapear os tokens no `tailwind.config` | Duas fontes de verdade para o mesmo componente (classes `st-*` do design system e utilitários); cada atualização do design system exigiria traduzir o CSS para classes |
| CSS Modules por componente | Reescreveria o CSS do design system em pedaços; o `bundle.css` já é organizado por componente e só usa variáveis |
| Carregar a fonte do Google Fonts | O app precisa funcionar offline e não deve depender de rede para abrir |

## Consequências

- Atualizar o visual é trocar `tokens.css`/`components.css` pelas versões novas do design system e revisar o `app.css`.
- O tema claro (#76) é só trocar o atributo `data-theme`; os componentes não mudam.
- Os componentes React passam a usar classes semânticas (`st-btn st-btn--primary`), sem cores no JSX.
- Os testes do frontend verificam classes semânticas (`st-drop--active`, `st-btn--busy`) em vez de cores.

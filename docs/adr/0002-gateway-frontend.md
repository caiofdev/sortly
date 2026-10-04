# ADR 0002 — Gateway no frontend para acesso ao backend

- **Status:** Substituída pela [ADR 0005](0005-estado-da-tela-no-backend.md). O gateway continua sendo o único módulo que importa os bindings, mas o controller, a normalização de erros e o `describeError` saíram.
- **Data:** 2026-10-02
- **Autores:** Caio Reis, Claude

## Contexto

Hoje o `useFileOrganizerController` chama diretamente a global `window.electronAPI`, criada pelo preload do Electron. No Wails, o backend é exposto por bindings JavaScript gerados (`wailsjs/go/app/App`), com outros nomes e outra forma de reportar erros.

Precisamos trocar o backend sem alterar a interface e queremos poder testar a lógica do frontend sem o runtime do Wails.

## Decisão

Criar um **gateway** (padrão Adapter) em `frontend/src/services/sortlyGateway.js`:

- É o **único** módulo que importa os bindings do Wails.
- Expõe funções com o mesmo contrato usado hoje pela interface (`selectSourceFolder`, `organizeFiles`, …).
- Converte os **códigos de erro** do backend em textos traduzidos ([ADR 0004](0004-erros-com-codigo.md)).
- O controller recebe o gateway como dependência, o que permite substituí-lo por um mock no Vitest.
- O gateway normaliza os erros em `SortlyError { code, message }`, aceitando a rejeição do Wails como string ou como `Error`. A tradução do código para o idioma atual fica em `i18n/describeError.js`.
- Durante a migração, o gateway também reconhece `window.electronAPI`, para a versão Electron continuar funcionando até ser removida.

## Alternativas consideradas

| Alternativa | Por que não |
|---|---|
| Shim global (`window.electronAPI = {...}`) | Funciona sem tocar no controller, mas mantém uma dependência global implícita, difícil de mockar e de rastrear |
| Importar os bindings direto nos componentes | Espalha o acoplamento ao Wails pela interface |

## Consequências

- O controller passa a depender de uma abstração, não do runtime (DIP).
- Trocar ou atualizar o Wails afeta um único arquivo.
- Uma pequena alteração no controller é necessária; os componentes visuais não mudam.

# Sortly - Documentação Técnica

Aplicação desktop para organização de arquivos por extensão, construída com Electron no processo principal e React no renderer.

## 1) Stack utilizada

### Runtime e plataforma
- Node.js 18+
- Electron 31

### Frontend (Renderer)
- React 18
- Vite 5
- TailwindCSS 3

### Build e empacotamento
- electron-builder (geração do instalador Windows)
- concurrently, wait-on e cross-env (orquestração de ambiente de desenvolvimento)

## 2) Arquitetura MVC aplicada

O projeto adota uma separação de responsabilidades inspirada em MVC, respeitando o contexto Electron (main process + renderer).

### Main Process (Electron)
- Model: regras de domínio de arquivos e políticas de nomeação (ex.: evitar sobrescrita).
- Controller: handlers IPC finos que recebem requisições do renderer e delegam para serviços.
- View: não há view tradicional no processo principal; ele atua como backend local da aplicação.

### Renderer (React)
- View: componentes de apresentação da interface.
- Controller: hook de controle de fluxo da UI, estado e orquestração das chamadas IPC.
- Model: neste MVP, regras de negócio ficam concentradas no main process.

### Fluxo macro
1. Usuário interage com a View (renderer).
2. Controller do renderer chama APIs expostas no preload.
3. Controller IPC no main process delega para services.
4. Services aplicam regras de domínio e usam infraestrutura (`fs`, `dialog`).
5. Resultado retorna ao renderer para atualização de estado/feedback.

## 3) Estrutura de pastas e responsabilidades

```text
.
├─ electron/
│  ├─ main.js                    # Composition root do processo principal
│  ├─ preload.js                 # Bridge segura entre renderer e IPC
│  ├─ bootstrap/                 # Inicialização da aplicação e ciclo de vida da janela
│  ├─ controllers/               # Handlers IPC (camada Controller no main)
│  ├─ services/                  # Casos de uso (organizar e desfazer)
│  ├─ models/                    # Regras de domínio reutilizáveis
│  ├─ repositories/              # Estado em memória da última operação
│  └─ infrastructure/            # Acesso a dependências externas (fs, dialog)
├─ src/
│  ├─ main.jsx                   # Entry point do React
│  ├─ App.jsx                    # Composição View + Controller do renderer
│  ├─ controllers/               # Hooks/controllers da interface
│  ├─ views/                     # Componentes de apresentação
│  └─ index.css                  # Estilos globais (Tailwind)
├─ build/                        # Recursos de build (ex.: ícone)
├─ dist/                         # Build do renderer
└─ release/                      # Artefatos empacotados do Electron
```

## 4) Como executar o projeto

### Pré-requisitos
- Node.js 18 ou superior
- npm 9 ou superior

### Instalação

```bash
npm install
```

### Desenvolvimento (Electron + React com hot reload)

```bash
npm run dev
```

Esse comando:
- inicia o Vite em `http://localhost:5173`
- aguarda o renderer ficar disponível
- inicia o Electron apontando para o servidor de desenvolvimento

### Executar Electron diretamente (sem servidor dev)

```bash
npm run start
```

Observação: para esse modo funcionar corretamente, o build do renderer deve existir em `dist/`.

### Build de produção

```bash
npm run build
```

Resultado esperado:
- `dist/`: build do frontend
- `release/`: instalador e artefatos do app desktop

## 5) Comportamento funcional atual

- Seleção de pasta de origem
- Seleção opcional de pasta de destino
- Organização por extensão (somente arquivos da raiz da origem)
- Criação automática de subpastas por extensão
- Tratamento de colisão de nome (`arquivo (1).ext`, `arquivo (2).ext`, ...)
- Ação de desfazer última organização
- Feedback com contadores de processados, movidos e ignorados
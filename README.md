# Sortly - Organizador de Arquivos

Aplicativo desktop MVP para organizar arquivos por extensão usando **Electron + React + Vite + TailwindCSS**.

## Requisitos

- Node.js 18+
- npm 9+

## Instalação

```bash
npm install
```

## Desenvolvimento (React + Electron)

```bash
npm run dev
```

## Build

```bash
npm run build
```

O build do frontend é gerado em `dist/` e o instalador do Electron é gerado em `release/`.

## Funcionalidades

- Seleção de pasta com diálogo nativo
- Organização apenas dos arquivos da raiz da pasta selecionada
- Criação automática de subpastas por extensão (case-insensitive)
- Ignora arquivos sem extensão
- Evita sobrescrita renomeando automaticamente (`nome (1).ext`, etc.)
- Feedback de sucesso/erro com contagem de arquivos

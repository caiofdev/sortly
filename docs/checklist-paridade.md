# Checklist de paridade (Electron 1.0 → Wails 2.0)

> Autores: Caio Reis, Claude

Roteiro manual para confirmar, em cada plataforma, que a versão 2.0 faz tudo o que a 1.0 fazia (e corrige B1–B8). Marque cada item ao conferir.

Legenda: ✅ conferido · 🤖 coberto por teste automatizado (CI) · ⏳ pendente · — não se aplica

No Windows, os itens marcados ✅ sem referência foram conferidos pelo roteiro automatizado [`scripts/parity/windows.ps1`](../scripts/parity/windows.ps1) (interface via UI Automation, pasta de usuário temporária, destino em outro disco para o B3) na issue [#15](https://github.com/caiofdev/sortly/issues/15). Os marcados com uma issue foram conferidos nela. macOS e Linux ficam para quem tiver as máquinas.

```powershell
powershell -File scripts/parity/windows.ps1 -OtherVolume E:\
```

## Como preparar

- Instale o pacote da plataforma (veja [release.md](release.md)) ou rode `wails dev`.
- Monte uma pasta de teste com imagens, vídeos `.mp4` (com e sem áudio), PDFs, `.docx`, `.odt`, um arquivo sem extensão (`LEIAME`), um `.gitignore` e uma subpasta.

## 1. Critérios de organização

| Item | Windows | macOS | Linux |
|---|---|---|---|
| Extensão (padrão) | ✅ | ⏳ | ⏳ |
| Data | ✅ | ⏳ | ⏳ |
| Tamanho | ✅ | ⏳ | ⏳ |
| Resolução (imagens) | ✅ | ⏳ | ⏳ |
| Duração (mp4, inclusive sem áudio) | ✅ | ⏳ | ⏳ |
| Páginas (pdf, docx, odt; `.doc` → `pages-unknown`) | ✅ | ⏳ | ⏳ |
| Os 6 critérios juntos (ordem das pastas) | ✅ | ⏳ | ⏳ |
| Não dá para desmarcar o último critério | ✅ | ⏳ | ⏳ |

## 2. Nomes, desfazer e registro

| Item | Windows | macOS | Linux |
|---|---|---|---|
| Arquivo com o mesmo nome no destino vira `nome (1).ext` (B4) | ✅ | ⏳ | ⏳ |
| Desfazer logo após organizar devolve tudo e apaga as pastas criadas | ✅ | ⏳ | ⏳ |
| Desfazer **após fechar e abrir** o app | ✅ | ⏳ | ⏳ |
| Arquivo apagado depois de organizar → "Não encontrados: 1" | ✅ | ⏳ | ⏳ |
| Organizar sem nada para mover mantém o desfazer anterior (B2) | ✅ | ⏳ | ⏳ |
| Arquivo bloqueado/em uso: os demais são movidos e podem ser desfeitos (B1) | ✅ | ⏳ | ⏳ |
| Mover para outro disco/volume (B3) | ✅ | ⏳ | ⏳ |
| Organizar na própria pasta só com "Resolução" não renomeia outros arquivos (B8) | ✅ | ⏳ | ⏳ |
| Gravação atômica do registro (B5) | 🤖 | 🤖 | 🤖 |

## 3. Interface

| Item | Windows | macOS | Linux |
|---|---|---|---|
| Selecionar origem e destino pelo seletor de pastas | ⏳ manual | ⏳ | ⏳ |
| Arrastar um **arquivo** para o painel (origem = pasta dele) | ✅ evento (#12) · ⏳ arrastar do Explorer | ⏳ | ⏳ |
| Arrastar uma **pasta** para o painel | ✅ evento (#12) · ⏳ arrastar do Explorer | ⏳ | ⏳ |
| Soltar fora do painel não faz nada | ✅ evento (#12) | ⏳ | ⏳ |
| Trocar o idioma (PT/EN) | ✅ | ⏳ | ⏳ |
| Idioma e critérios continuam após reabrir o app | ✅ | ⏳ | ⏳ |
| Erros aparecem no idioma escolhido (B6) | ✅ | ⏳ | ⏳ |
| Fonte Inter, ícone, gradientes e desfoque do cartão | ✅ (#13) | ⏳ | ⏳ |

## 4. Instalação

| Item | Windows | macOS | Linux |
|---|---|---|---|
| Instalar o pacote e abrir pelo atalho/menu | ✅ (#14) | ⏳ | ⏳ |
| Atualizar a partir da 1.0 (remove a anterior, preserva o desfazer) | ✅ (#14) | — | — |
| Desinstalar (remove app e atalhos, preserva `~/.sortly`) | ✅ (#14) | ⏳ | ⏳ |

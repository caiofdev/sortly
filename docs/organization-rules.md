# Regras de organização

> Autores: Caio Reis, Claude

Este documento descreve o comportamento do Sortly. Ele nasceu como especificação de paridade da versão 1.0 (Electron) para a reescrita em Go: o comportamento é o mesmo, e as diferenças intencionais estão marcadas com 🔧 e detalhadas em [code-review.md](code-review.md).

Implementação: `internal/organizer`, `internal/metadata`, `internal/undo` e `internal/fsutil`. O código da 1.0 (`electron/services/organizeFilesService.js`, `undoOrganizationService.js`, `models/pathModel.js`) está no histórico do git.

## 1. Entrada

| Campo | Regra |
|---|---|
| Pasta de origem | Obrigatória. Sem ela: erro **"Pasta inválida."** |
| Pasta de destino | Opcional. Se vazia, usa a própria origem |
| Critérios | Pelo menos um deve estar ativo. Sem nenhum: erro **"Selecione ao menos um criterio de organizacao."** |

Os seis critérios são booleanos. Valores ausentes contam como desligados, **exceto `byExtension`, que é ligado por padrão** (só desliga se vier explicitamente `false`).

Padrão da interface (primeiro uso): apenas **Extensão** ligado.

## 2. Varredura

- Apenas o **nível superior** da pasta de origem é lido — não há recursão.
- Subpastas são ignoradas e contadas em `ignoredFolders`.
- Entradas que não são arquivo regular nem pasta (ex.: alguns tipos de link) são ignoradas silenciosamente.
- Cada arquivo regular conta em `processedFiles`.
- 🔧 Se o destino calculado é o próprio arquivo (destino = origem e nenhum critério gerou subpasta), a versão Electron o renomeava para `nome (1).ext`. Na versão Go ele fica onde está e conta em `unchangedFiles` (B8 ✅).

## 3. Extensão do arquivo

A extensão é obtida como no `path.extname` do Node, sem o ponto e em minúsculas:

| Nome | Extensão |
|---|---|
| `foto.JPG` | `jpg` |
| `backup.tar.gz` | `gz` |
| `README` | *(nenhuma)* |
| `arquivo.` | *(nenhuma)* |
| `.gitignore` | *(nenhuma)* — arquivo oculto sem extensão |

> ⚠️ Em Go, `filepath.Ext(".gitignore")` retorna `.gitignore`. A reescrita **deve** replicar o comportamento do Node.

## 4. Segmentos de pasta

Cada critério ativo gera **uma subpasta**, aninhada sempre nesta ordem:

| Ordem | Critério | Aplica-se a | Nome da pasta | Sem informação |
|---|---|---|---|---|
| 1 | Extensão (`byExtension`) | todos | `<ext>` (ex.: `pdf`) | arquivo **não é movido**; conta em `ignoredWithoutExtension` |
| 2 | Data (`byDate`) | todos | `date-YYYY-MM-DD` (data de modificação, fuso local) | — |
| 3 | Tamanho (`bySize`) | todos | `size-<N>mb` | — |
| 4 | Resolução (`byResolution`) | png, jpg, jpeg, gif, webp, bmp | `<largura>x<altura>` (ex.: `1920x1080`) | `unknown` |
| 5 | Duração (`byDuration`) | mp4 | `duration-HHhMMmSSs` | `duration-unknown` |
| 6 | Páginas (`byPages`) | pdf, docx, odt, doc | `pages-<N>` | `pages-unknown` |

Critérios que não se aplicam ao tipo do arquivo são simplesmente pulados (não geram pasta).

**Exemplo** — `relatorio.pdf` (2,3 MB, 12 páginas, modificado em 05/03/2026) com Extensão + Tamanho + Páginas:

```
<destino>/pdf/size-3mb/pages-12/relatorio.pdf
```

Com a Extensão desligada, arquivos sem extensão **são** movidos pelos demais critérios.

### 4.1 Tamanho

`N = max(1, teto(bytes / 1 048 576))`

| Bytes | Pasta |
|---|---|
| 0 | `size-1mb` |
| 1 | `size-1mb` |
| 1 048 576 | `size-1mb` |
| 1 048 577 | `size-2mb` |
| 2 097 152 | `size-2mb` |
| 2 097 153 | `size-3mb` |

### 4.2 Duração

**De onde vem a duração.** A versão Electron (`music-metadata`) usa a duração da **primeira faixa de áudio** do mp4 (box `mdhd`), não a do filme. Por isso, um vídeo sem áudio (gravação de tela, mp4 convertido de GIF) cai em `duration-unknown`.

🔧 Na versão Go (`internal/metadata`), a faixa de áudio continua sendo a primeira opção, então quem já organizou com áudio recebe a mesma pasta. Se não houver faixa de áudio com duração, usa a duração do filme (box `mvhd`), a mesma que o player mostra. A versão Go também lê o `mdhd`/`mvhd` versão 1 (campos de 64 bits), que a `music-metadata` ignora e trata como `unknown`.

Os segundos são arredondados para o inteiro mais próximo (`Math.round`, 0,5 arredonda para cima) e formatados com dois dígitos em cada parte. Duração ausente, zero ou negativa resulta em `duration-unknown`.

| Duração | Pasta |
|---|---|
| 0 s (ou ilegível) | `duration-unknown` |
| 0,4 s | `duration-00h00m00s` |
| 0,5 s | `duration-00h00m01s` |
| 59 s | `duration-00h00m59s` |
| 60 s | `duration-00h01m00s` |
| 3 599 s | `duration-00h59m59s` |
| 3 600 s | `duration-01h00m00s` |

### 4.3 Páginas

| Formato | Fonte da contagem |
|---|---|
| pdf | contagem real de páginas do documento |
| docx | `docProps/app.xml` → `<Pages>N</Pages>` |
| odt | `meta.xml` → `meta:page-count="N"` |
| doc | sem leitor — sempre `pages-unknown` |

Valores ausentes, zero ou inválidos resultam em `pages-unknown`.

### 4.4 Resolução

Lida do cabeçalho da imagem (sem considerar orientação EXIF). O formato é detectado pelo **conteúdo**, não pela extensão: um `.png` que na verdade é um JPEG tem a resolução lida normalmente. No BMP, altura negativa (imagem gravada de cima para baixo) conta pelo valor absoluto. Falha de leitura ou dimensão zero resulta em `unknown`.

## 5. Conflito de nomes

O Sortly **nunca sobrescreve** arquivos. Se o destino já existe, tenta `nome (1).ext`, `nome (2).ext`, … até encontrar um nome livre.

| Situação | Resultado |
|---|---|
| `foto.jpg` livre | `foto.jpg` |
| `foto.jpg` existe | `foto (1).jpg` |
| `foto.jpg` e `foto (1).jpg` existem | `foto (2).jpg` |
| `LEIAME` existe | `LEIAME (1)` |

🔧 Na versão Go (`internal/fsutil`, `Reserve`/`MoveUnique`) a busca vai até `nome (9999)` e o nome é reservado com criação exclusiva, evitando que dois movimentos usem o mesmo nome (B4 ✅). A extensão segue a mesma regra do §3, então `.gitignore` vira `.gitignore (1)` e `backup.tar.gz` vira `backup.tar (1).gz`.

## 6. Movimento

Os arquivos são movidos (não copiados), um de cada vez. As pastas de destino são criadas conforme necessário.

🔧 Na versão Go (`internal/fsutil.Move`), se o rename falha por serem volumes diferentes, o arquivo é copiado, a data de modificação é preservada e a origem é removida. Se qualquer etapa falhar, a cópia é apagada e a origem fica intacta: o arquivo nunca fica duplicado nem perdido (B3 ✅).

## 7. Registro da operação

Ao final, a operação é salva em `~/.sortly/last-operation.json`:

```json
{
  "sourceFolderPath": "C:\\Users\\ana\\Downloads",
  "destinationFolderPath": "C:\\Users\\ana\\Organizados",
  "movedItems": [
    { "from": "C:\\Users\\ana\\Downloads\\a.pdf", "to": "C:\\Users\\ana\\Organizados\\pdf\\a.pdf" }
  ],
  "createdFolders": ["C:\\Users\\ana\\Organizados\\pdf"]
}
```

Esse formato **deve ser mantido** para que um desfazer pendente da versão Electron continue funcionando na versão Wails.

🔧 Na versão Electron, o registro era sobrescrito mesmo quando nenhum arquivo era movido (B2) e não era salvo se um movimento falhasse no meio (B1). Na versão Go (`internal/organizer`), uma falha num arquivo não interrompe os demais: cada movimento concluído entra no registro e a falha é contada em `failedFiles` (B1 ✅). Se nada for movido, o registro anterior é preservado e o desfazer dele continua disponível (B2 ✅).

🔧 Na versão Go (`internal/store`), a gravação é atômica: o JSON vai para um arquivo temporário na mesma pasta, que depois substitui o atual. Um crash no meio da gravação deixa o registro anterior intacto, e falhas são registradas no log (B5 ✅). O arquivo gravado é idêntico, byte a byte, ao da versão Electron, e um arquivo vazio ou corrompido é tratado como "nada para desfazer".

## 8. Desfazer

- Apenas a **última** organização pode ser desfeita. Sem registro: erro **"Nenhuma separação recente para desfazer."**
- Os movimentos são revertidos em **ordem inversa**.
- Arquivo que não está mais no destino: pulado, conta em `skippedMissing`.
- Se o local original estiver ocupado, aplica a regra de conflito (§5) e conta em `renamedOnRestore`.
- Depois, remove pastas que ficaram **vazias**, subindo da pasta do arquivo até a raiz do destino (exclusive). Pastas com outros arquivos são preservadas.
- No Windows a comparação de caminhos ignora maiúsculas/minúsculas.
- Ao final o registro é apagado.

🔧 Na versão Go (`internal/undo`):

- Uma falha num arquivo não interrompe os demais. O registro é regravado só com os itens que falharam, e o botão de desfazer continua disponível para tentar de novo (`failedFiles`). Na versão Electron, a primeira falha interrompia o desfazer.
- A subida para remover pastas vazias para na raiz do destino mesmo que o caminho venha com maiúsculas/minúsculas diferentes no Windows; a raiz nunca é removida.
- Um registro gravado pela versão Electron é desfeito normalmente. Testado: organizar com a versão 1.0 e desfazer com a versão Go devolve a árvore idêntica.

O estado de desfazer sobrevive ao fechamento do app: ao abrir, a interface recupera origem e destino e avisa que é possível desfazer.

## 9. Arrastar e soltar

| Item solto | Resultado |
|---|---|
| Pasta | vira a pasta de origem |
| Arquivo | a pasta que contém o arquivo vira a origem |
| Item inexistente | erro |
| Outro tipo | erro |
| Vários itens | usa o primeiro |

O item precisa ser solto **sobre o painel** de arrastar e soltar; fora dele, nada acontece.

🔧 Na versão Go, o Wails entrega os caminhos dos arquivos soltos (DragAndDrop.EnableFileDrop) apenas quando o drop termina num elemento com o estilo --wails-drop-target: drop, que o painel define. O frontend recebe esses caminhos por gateway.subscribeFileDrop. A versão Electron lia File.path, que não existe no WebView do sistema.

## 10. Preferências

Idioma (`pt-BR` padrão, `en`) e critérios selecionados ficam salvos localmente na interface. A interface impede desmarcar o último critério ativo.

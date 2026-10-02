# Regras de organização

> Autores: Caio Reis, Claude

Este documento descreve **o comportamento atual** do Sortly (versão 1.0, Electron) e serve como especificação de paridade para a reescrita em Go. Divergências planejadas estão marcadas com 🔧 e detalhadas em [code-review.md](code-review.md).

Referência de código atual: `electron/services/organizeFilesService.js`, `electron/services/undoOrganizationService.js`, `electron/models/pathModel.js`.

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

Lida do cabeçalho da imagem (sem considerar orientação EXIF). Falha de leitura ou dimensão zero resulta em `unknown`.

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

🔧 Hoje o registro é sobrescrito mesmo quando nenhum arquivo é movido (B2) e não é salvo se um movimento falhar no meio (B1).

## 8. Desfazer

- Apenas a **última** organização pode ser desfeita. Sem registro: erro **"Nenhuma separação recente para desfazer."**
- Os movimentos são revertidos em **ordem inversa**.
- Arquivo que não está mais no destino: pulado, conta em `skippedMissing`.
- Se o local original estiver ocupado, aplica a regra de conflito (§5) e conta em `renamedOnRestore`.
- Depois, remove pastas que ficaram **vazias**, subindo da pasta do arquivo até a raiz do destino (exclusive). Pastas com outros arquivos são preservadas.
- No Windows a comparação de caminhos ignora maiúsculas/minúsculas.
- Ao final o registro é apagado.

O estado de desfazer sobrevive ao fechamento do app: ao abrir, a interface recupera origem e destino e avisa que é possível desfazer.

## 9. Arrastar e soltar

| Item solto | Resultado |
|---|---|
| Pasta | vira a pasta de origem |
| Arquivo | a pasta que contém o arquivo vira a origem |
| Item inexistente | erro |
| Outro tipo | erro |

## 10. Preferências

Idioma (`pt-BR` padrão, `en`) e critérios selecionados ficam salvos localmente na interface. A interface impede desmarcar o último critério ativo.

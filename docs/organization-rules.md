# Regras de organização

> Autores: Caio Reis, Claude

Comportamento do Sortly ao organizar e desfazer. É a especificação que os testes seguem: nomes de pastas, fronteiras e formato do registro não mudam sem uma decisão registrada, porque um desfazer pendente da versão 1.0 precisa continuar funcionando.

Implementação: `backend/organizer` (planejamento e execução), `backend/organizer/criteria` (um arquivo por critério), `backend/metadata` (resolução, duração e páginas), `backend/fs` (mover, nomes e caminhos), `backend/store` (registro) e `backend/undo`.

## 1. Entrada

| Campo | Regra | Erro |
|---|---|---|
| Pasta de origem | Obrigatória e precisa existir | `INVALID_SOURCE` — "Pasta inválida." |
| Pasta de destino | Opcional; se vazia, usa a própria origem | `INVALID_DESTINATION` — "Pasta de destino inválida." |
| Critérios | Pelo menos um ativo | `NO_CRITERIA` — "Selecione ao menos um criterio de organizacao." |

O backend devolve só o código; o texto, em português ou inglês, vem da interface.

Os critérios usados são os salvos nas preferências (§10); no primeiro uso, só **Extensão**.

## 2. Varredura

- Apenas o **nível superior** da pasta de origem é lido; não há recursão.
- Subpastas são ignoradas e contadas em `ignoredFolders`.
- Entradas que não são arquivo regular nem pasta (links, dispositivos) são ignoradas.
- Cada arquivo regular conta em `processedFiles`.
- Se o destino calculado é o próprio arquivo (destino = origem e nenhum critério gerou subpasta), ele fica onde está e conta em `unchangedFiles`.

## 3. Extensão do arquivo

A extensão segue o `path.extname` do Node (usado na versão 1.0), sem o ponto e em minúsculas — diferente do `filepath.Ext` do Go, que trata `.gitignore` como extensão:

| Nome | Extensão |
|---|---|
| `foto.JPG` | `jpg` |
| `backup.tar.gz` | `gz` |
| `README` | *(nenhuma)* |
| `arquivo.` | *(nenhuma)* |
| `.gitignore` | *(nenhuma)* — arquivo oculto sem extensão |

## 4. Segmentos de pasta

Cada critério ativo gera **uma subpasta**, aninhada sempre nesta ordem:

| Ordem | Critério | Aplica-se a | Nome da pasta | Sem informação |
|---|---|---|---|---|
| 1 | Tipo (`byType`, desligado por padrão) | todos | a categoria (§4.5): `images`, `documents`, `archives`, `installers`, `videos`, `audio` ou `other` | `other` (sem extensão ou extensão desconhecida) |
| 2 | Extensão (`byExtension`) | todos | `<ext>` (ex.: `pdf`) | arquivo **não é movido**; conta em `ignoredWithoutExtension` |
| 3 | Data (`byDate`) | todos | `date-YYYY-MM-DD` (data de modificação, fuso local) | — |
| 4 | Tamanho (`bySize`) | todos | `size-<N>mb` | — |
| 5 | Resolução (`byResolution`) | png, jpg, jpeg, gif, webp, bmp | `<largura>x<altura>` (ex.: `1920x1080`) | `unknown` |
| 6 | Duração (`byDuration`) | mp4 | `duration-HHhMMmSSs` | `duration-unknown` |
| 7 | Páginas (`byPages`) | pdf, docx, odt, doc | `pages-<N>` | `pages-unknown` |

Critérios que não se aplicam ao tipo do arquivo são pulados (não geram pasta).

**Exemplo** — `relatorio.pdf` (2,3 MB, 12 páginas) com Extensão + Tamanho + Páginas:

```
<destino>/pdf/size-3mb/pages-12/relatorio.pdf
```

Com a Extensão desligada, arquivos sem extensão **são** movidos pelos demais critérios. Com Tipo e Extensão ligados, o arquivo sem extensão continua sem ser movido, como na 1.0.

**Exemplo** — `foto.jpg` com Tipo + Extensão:

```
<destino>/images/jpg/foto.jpg
```

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

A duração vem do box `mdhd` da **primeira faixa de áudio** do mp4 (com pelo menos um canal), como na versão 1.0. Se não houver faixa de áudio com duração — gravação de tela, mp4 convertido de GIF —, usa a duração do filme (box `mvhd`), a mesma que o player mostra. Os boxes versão 0 (32 bits) e versão 1 (64 bits) são lidos.

Os segundos são arredondados para o inteiro mais próximo (0,5 arredonda para cima) e formatados com dois dígitos em cada parte. Duração ausente ou zero resulta em `duration-unknown`.

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

### 4.5 Tipo

As pastas têm nomes neutros, como as dos outros critérios; a interface mostra o rótulo traduzido (Imagens, Documentos, Compactados, Instaladores, Vídeos, Áudio, Outros). Para isso, a prévia e o resultado trazem `categoryFolders: true` quando o Tipo está ligado.

| Pasta | Extensões |
|---|---|
| `images` | jpg, jpeg, png, gif, bmp, webp, svg, tif, tiff, heic, heif, ico, raw, cr2, nef, arw, dng, psd |
| `documents` | pdf, doc, docx, odt, rtf, txt, md, xls, xlsx, ods, csv, ppt, pptx, odp, epub |
| `archives` | zip, rar, 7z, tar, gz, tgz, bz2, xz, zst |
| `installers` | exe, msi, msix, dmg, pkg, deb, rpm, appimage, apk |
| `videos` | mp4, mkv, avi, mov, wmv, webm, m4v, flv, mpg, mpeg, 3gp |
| `audio` | mp3, wav, flac, aac, ogg, oga, m4a, wma, opus, aiff |
| `other` | qualquer outra extensão e arquivos sem extensão |

A extensão é comparada em minúsculas e é só a última parte do nome: `backup.tar.gz` é `gz`, em `archives`.

## 5. Conflito de nomes

O Sortly **nunca sobrescreve** arquivos. Se o destino já existe, tenta `nome (1).ext`, `nome (2).ext`, … até `nome (9999).ext`. O nome escolhido é reservado com criação exclusiva, então dois movimentos nunca usam o mesmo nome.

| Situação | Resultado |
|---|---|
| `foto.jpg` livre | `foto.jpg` |
| `foto.jpg` existe | `foto (1).jpg` |
| `foto.jpg` e `foto (1).jpg` existem | `foto (2).jpg` |
| `LEIAME` existe | `LEIAME (1)` |
| `.gitignore` existe | `.gitignore (1)` |
| `backup.tar.gz` existe | `backup.tar (1).gz` |

## 6. Movimento

- Os arquivos são movidos (não copiados), um de cada vez. As pastas de destino são criadas conforme necessário.
- Entre volumes diferentes (outro disco, pendrive), o arquivo é copiado com a data de modificação preservada, a cópia é gravada em disco (`fsync` do arquivo e, no Linux e no macOS, da pasta de destino) e só então a origem é removida. Se qualquer etapa falhar, a cópia é apagada e a origem fica intacta: o arquivo nunca fica duplicado nem perdido.
- Uma falha num arquivo (bloqueado, sem permissão) não interrompe os demais; ela conta em `failedFiles`.
- No Windows, nomes que terminam em ponto ou espaço e caminhos longos funcionam (acesso com o prefixo `\\?\`).

## 7. Registro da operação

Ao final, a operação é salva em `~/.sortly/last-operation.json`, no mesmo formato da versão 1.0:

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

- Cada movimento concluído entra no registro, mesmo que outro falhe depois.
- `createdFolders` lista só as pastas que a organização **criou**, em todos os níveis (ex.: `pdf` e `pdfpages-3` quando nenhuma existia). Pastas que já existiam no destino não entram. A versão 1.0 listava a pasta de cada arquivo; o formato é o mesmo.
- Se nada for movido, o registro anterior é preservado e o desfazer dele continua disponível.
- A gravação é atômica (arquivo temporário na mesma pasta, depois rename): um crash no meio deixa o registro anterior intacto. Falhas de gravação vão para o log e viram o erro `RECORD_NOT_SAVED`; nesse caso o registro anterior é apagado, porque descreve outra organização e o desfazer não pode apontar para ela.
- Um arquivo vazio ou corrompido é tratado como "nada para desfazer".

## 8. Desfazer

- Apenas a **última** organização pode ser desfeita. Sem registro: `NOTHING_TO_UNDO` — "Nenhuma separação recente para desfazer."
- Os movimentos são revertidos em **ordem inversa**.
- Arquivo que não está mais no destino é pulado e conta em `skippedMissing`.
- Só voltam itens com `from` dentro da pasta de origem e `to` dentro da pasta de destino do registro (destino vazio = origem), sem ser a própria pasta. Um registro editado ou corrompido com caminhos de fora não move nada: o item é pulado, conta em `skippedMissing` e vai para o log. Registro sem pasta de origem não restaura nada.
- Se o local original estiver ocupado, aplica a regra de conflito (§5) e conta em `renamedOnRestore`.
- Uma falha num arquivo não interrompe os demais. O registro é regravado só com os itens que falharam, e o desfazer continua disponível para tentar de novo (`failedFiles`). Sem falhas, o registro é apagado.
- Depois, remove as pastas de `createdFolders` que ficaram **vazias**, da mais funda para a mais rasa. Pastas que já existiam antes da organização ficam, mesmo vazias. Em registros antigos, sem `createdFolders`, sobe da pasta de cada arquivo até a raiz do destino (exclusive). Nada fora dessa raiz é tocado. Pastas com outros arquivos são preservadas, e a raiz nunca é removida, mesmo com diferença de maiúsculas/minúsculas no caminho (Windows e macOS).

O estado de desfazer sobrevive ao fechamento do app: ao abrir, a interface recupera origem e destino e avisa que é possível desfazer.

## 9. Arrastar e soltar

| Item solto | Resultado | Erro |
|---|---|---|
| Pasta | vira a pasta de origem | — |
| Arquivo | a pasta que contém o arquivo vira a origem | — |
| Item inexistente | — | `DROPPED_MISSING` |
| Outro tipo | — | `DROPPED_UNSUPPORTED` |
| Caminho inválido | — | `DROPPED_INVALID` |
| Vários itens | usa o primeiro | — |

O item precisa ser solto **sobre o painel** de arrastar e soltar; fora dele, nada acontece. O Wails só entrega os caminhos quando o drop termina num elemento com o estilo `--wails-drop-target: drop`, que o painel define.

## 10. Preferências

Idioma e critérios ficam em `~/.sortly/settings.json`, gravado de forma atômica:

```json
{"language":"pt-BR","theme":"dark","organizationOptions":{"byDate":false,"byDuration":false,"byExtension":true,"byPages":false,"byResolution":false,"bySize":false,"byType":false}}
```

| Situação | Resultado |
|---|---|
| Arquivo ausente, vazio, corrompido ou `null` | Padrão: `pt-BR` e só Extensão |
| Idioma diferente de `pt-BR` e `en` | Idioma padrão; os critérios do arquivo valem |
| Nenhum critério ligado no arquivo | Critérios padrão; o idioma do arquivo vale |
| Chave de critério desconhecida | Ignorada |
| Trocar para um idioma não suportado | `INVALID_LANGUAGE` |
| Desligar o último critério ligado | `LAST_CRITERION` — a interface já mostra esse checkbox desabilitado |
| Chave de critério desconhecida ao alterar | `UNKNOWN_CRITERION` |
| Falha ao gravar | `SETTINGS_NOT_SAVED`; as preferências continuam as anteriores |

A interface mostra os critérios na ordem Tipo, Duração, Páginas, Resolução, Data, Tamanho, Extensão (diferente da ordem de aninhamento das pastas, §4). Um `settings.json` sem `byType` (de antes da #81) vale como Tipo desligado.

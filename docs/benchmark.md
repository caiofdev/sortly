# Benchmark: Electron 1.0 × Wails 2.0

> Autores: Caio Reis, Claude · Medido em 2026-10-03 e 2026-10-04 (issues [#15](https://github.com/caiofdev/sortly/issues/15) e [#34](https://github.com/caiofdev/sortly/issues/34))

## Resumo

Com a configuração atual (aceleração por GPU do WebView2 **desligada**, [#34](https://github.com/caiofdev/sortly/issues/34)), a versão Wails usa no Windows **32% menos memória privada em repouso** e **49% menos no pico** que o Electron 1.0, com instalador 9× menor, disco 15× menor e organização ~2× mais rápida. A interface é visualmente idêntica (conferido pixel a pixel na janela nativa).

Com a GPU ligada, como na primeira medição (#15), a memória ficava próxima da do Electron: o WebView2 também é Chromium e o processo de GPU respondia por ~80% da memória comprometida nas duas versões.

## Ambiente

- Windows 11 Pro 25H2, Intel Core i3-10100F, 8 GB de RAM, NVIDIA GeForce GTX 1050 Ti.
- Electron 31.7.7 (Chromium 126) × Wails 2.16.0 com WebView2 154.
- As duas versões rodaram com o **mesmo frontend** (esta branch) e com pasta de usuário e AppData temporárias.
- Conjunto de teste: 1000 arquivos montados a partir das fixtures (imagens, mp4, PDFs, docx, odt), organizados com os **6 critérios** e depois desfeitos.

## Resultados (média de 3 rodadas)

| Medida | Electron 1.0 | **Wails 2.0 (atual, sem GPU)** | Wails com GPU (#15) |
|---|---|---|---|
| Memória privada em repouso | 260 MB | **177 MB** | 277 MB |
| Memória privada no pico ao organizar | 388 MB | **200 MB** | 346 MB |
| Memória privada no pico ao desfazer | 393 MB | **200 MB** | 351 MB |
| Working set privado em repouso (≈ Gerenciador de Tarefas) | 99 MB | **92 MB** | 95 MB |
| Working set total em repouso | 331 MB | 391 MB | 367 MB |
| Processos | 4 | 7 | 7 |
| Organizar 1000 arquivos | 7,8 s | **4,0 s** | 4,1 s |
| Desfazer 1000 arquivos | 2,6 s | **2,1 s** | 2,1 s |
| Tempo até a janela | 1,3 s | **0,8 s** | 1,4 s |
| Instalador | 82 MB | **8,8 MB** | 8,8 MB |
| Espaço em disco | 259 MB | **17 MB** | 17 MB |

### Como ler as medidas

- **Memória privada** (commit): memória exclusiva do app, reservada no sistema. É a medida que mais cai sem GPU.
- **Working set privado:** memória exclusiva que está de fato na RAM, a coluna "Memória" do Gerenciador de Tarefas.
- **Working set total:** inclui páginas compartilhadas com outros programas (as DLLs do WebView2 são compartilhadas com o Edge e o Windows). Por isso ele é maior no Wails, sem significar mais consumo exclusivo.

A coluna "atual" é o build da [#34](https://github.com/caiofdev/sortly/issues/34) (`WebviewGpuIsDisabled: true` no Windows, `WebviewGpuPolicyNever` no Linux). Na #15, um experimento temporário com a mesma configuração mediu 174 MB em repouso, 199 MB no pico e 88 MB de working set privado, dentro da variação esperada entre rodadas. O macOS não tem opção equivalente no Wails.

### Visual sem GPU

A renderização passa a ser por software. Capturas da janela nativa com e sem GPU, nos estados inicial, configurações abertas e notificações abertas, diferem em no máximo **0,006% dos pixels** (cerca de 120 pixels de arredondamento de gradiente).

### Memória por processo, em repouso (GPU ligada, #15)

| Processo | Electron (privada) | Wails (privada) |
|---|---|---|
| GPU | 247,9 MB | 223,3 MB |
| Principal (Node / Go) | 75,0 MB | 55,4 MB |
| Navegador (WebView2) | (dentro do principal) | 36,5 MB |
| Renderer | 31,7 MB | 26,2 MB |
| Rede | 13,4 MB | 11,3 MB |
| Armazenamento + crashpad | — | 9,5 MB |

## Como reproduzir

```powershell
wails build                                   # build/bin/Sortly.exe
npm install                                   # runtime do Electron em node_modules/
cd frontend; npm run build; cd ..             # frontend usado pela versão Electron

powershell -File scripts/benchmark/memory.ps1 -App electron -IdleOnly   # repouso sem acessibilidade
powershell -File scripts/benchmark/memory.ps1 -App electron             # organizar + desfazer (1000 arquivos)
powershell -File scripts/benchmark/memory.ps1 -App wails -IdleOnly
powershell -File scripts/benchmark/memory.ps1 -App wails
```

O script soma a memória de **toda a árvore de processos** do app a cada 200 ms. As ações são feitas pela interface, com UI Automation (os mesmos cliques de um usuário), porque o loader de WebView2 do Wails bloqueia a depuração remota (`WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS`).

A UI Automation liga a árvore de acessibilidade, que aumenta a memória. Por isso o repouso é medido à parte (`-IdleOnly`), sem nenhuma consulta. No Electron, a automação exige `--force-renderer-accessibility --enable-features=UiaProvider`, ligadas só na execução das ações.

## Limitações

- Uma única máquina, com GPU dedicada NVIDIA. Em GPUs integradas, o processo de GPU tende a ocupar menos memória.
- macOS e Linux não foram medidos. O WKWebView e o WebKitGTK não usam a arquitetura multiprocesso do Chromium, e a diferença para o Electron tende a ser maior nessas plataformas.
- A versão Electron medida usa o frontend refatorado desta branch, com a mesma interface. A configuração do Electron e o backend Node são os da versão 1.0.

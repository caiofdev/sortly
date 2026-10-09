# Aprendizados

Regras que já falharam neste projeto e que a revisão deve conferir sempre. Cada item: regra generalizada — origem.

## Herdados da versão 1.0 (code review da migração)

- Falha no meio de um lote não pode impedir desfazer o que já foi feito: o journal é gravado mesmo com erro. — B1
- Operação que não fez nada não sobrescreve o registro anterior (não apaga o desfazer pendente). — B2
- `rename` não funciona entre volumes; sempre há fallback de cópia. — B3
- Laço de nome único tem limite e reserva atômica (`O_EXCL`); "existe? → cria" tem corrida. — B4
- Gravação de estado em disco é atômica (temporário + rename) e falhas vão para o log, nunca são engolidas. — B5
- Backend devolve código de erro; texto para o usuário só no frontend, no idioma escolhido. — B6/B7
- Destino calculado igual à origem é no-op, não conflito. — B8

## Da reescrita em Wails

- Nome de arquivo terminado em ponto não funciona no Windows sem o prefixo `\\?\`: todo acesso a disco passa por `paths.Native`. — #8
- Vídeo sem faixa de áudio tem duração no `mvhd`; não assuma que todo mp4 tem áudio. — #7
- `options.Linux` não nil desliga padrões do Wails (política de GPU): ao preencher uma struct de opções, confira o que o padrão fazia. — #34
- O loader de WebView2 do Wails apaga `WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS`; não planeje automação por CDP. — #15
- Desinstalador do Electron tinha `InstallLocation` vazio; não confie em campos opcionais do registro do Windows. — #14
- Scripts `.ps1` com acentos precisam de BOM no PowerShell 5.1. — #15

## Novos

- Slice nil em struct enviada ao frontend vira `null` no JSON, e o JS que faz `.map` quebra a tela inteira. Toda lista de um binding é inicializada (`[]T{}`) e há teste do JSON. — #45 (2026-10-04)

- Preferência do projeto: nada de `switch`/`case`; mapa, retorno antecipado ou Strategy. — Caio (2026-10-04)
- Comentário que repete o nome da função ("// Get devolve as preferências") é ruído, mesmo em identificador exportado; comente só o porquê. — Caio (2026-10-04)

- Cópia entre volumes só é segura depois de `Sync` do arquivo **e** da pasta de destino (Unix); só então a origem é apagada. — #52 (2026-10-05)
- Erro com resultado parcial ainda muda o estado: se arquivos foram movidos e o registro falhou, o desfazer anterior não pode continuar disponível (nem após reabrir). — #53 (2026-10-05)
- Dado lido do disco que vira caminho de destino (registro do desfazer) é validado contra as pastas registradas antes de mover. — #54 (2026-10-05)
- "Pasta criada" registrada tem de ser só a que não existia; pasta do usuário nunca é removida por limpeza. — #57 (2026-10-06)
- macOS (APFS) também não diferencia maiúsculas; plataforma insensível não é só Windows. — #58 (2026-10-06)
- Contagem de CC no JS estimada de cabeça errou (CC 8 anotada como 5); liste pela regra `complexity` do ESLint. — #60 (2026-10-06)
- Mudar atributo de acessibilidade (`aria-pressed`) muda o padrão de UI Automation e quebra o roteiro de paridade: rode a paridade em toda mudança de componente. — #60 (2026-10-06)

- Elemento fixo por cima da tela (toast, painel) que só some quando o usuário fecha não pode cobrir as ações principais: meça na janela padrão (980×700) e na mínima (820×600), com o máximo de itens empilhados. — #75 (2026-10-08)

- Atributo do <html> que muda a aparência (data-theme) vai em useLayoutEffect: com useEffect, a primeira tela pinta com os tokens errados por um quadro. — #76 (2026-10-09)

- Código que lê arquivos do usuário e passa a rodar numa goroutine própria (fora de um binding) precisa de recover: ali um panic encerra o app inteiro. — #77 (2026-10-09)

- Desenho com posições fixas em px (FileFlow, 520px) precisa caber na janela mínima: confira 820×600 e reduza por inteiro (container query + zoom) em vez de deixar cortar. — #78 (2026-10-09)
- Painel que substitui o botão focado (Organizar → Organizando) move o foco para a ação seguinte; senão o teclado cai no <body>. — #78 (2026-10-09)
- No roteiro de paridade, espere pelo efeito no disco (arquivo saiu da pasta), não por um texto da tela, antes de agir no meio de uma operação. — #78 (2026-10-09)

<!-- Acrescente aqui: - Regra generalizada. — PR/issue (AAAA-MM-DD) -->

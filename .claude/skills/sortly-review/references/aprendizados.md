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

- Nome de arquivo terminado em ponto não funciona no Windows sem o prefixo `\\?\`: todo acesso a disco passa por `nativePath`. — #8
- Vídeo sem faixa de áudio tem duração no `mvhd`; não assuma que todo mp4 tem áudio. — #7
- `options.Linux` não nil desliga padrões do Wails (política de GPU): ao preencher uma struct de opções, confira o que o padrão fazia. — #34
- O loader de WebView2 do Wails apaga `WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS`; não planeje automação por CDP. — #15
- Desinstalador do Electron tinha `InstallLocation` vazio; não confie em campos opcionais do registro do Windows. — #14
- Scripts `.ps1` com acentos precisam de BOM no PowerShell 5.1. — #15

## Novos

<!-- Acrescente aqui: - Regra generalizada. — PR/issue (AAAA-MM-DD) -->

# ADR 0007 — Histórico separado do registro do desfazer

- **Status:** Aceita
- **Data:** 2026-10-09
- **Autores:** Caio Reis, Claude

## Contexto

O redesign tem uma página Histórico, com as organizações feitas no computador: quando, a pasta, quantos arquivos e se foi concluída, desfeita ou interrompida (#80). Até aqui, o app guardava só a última organização, em `~/.sortly/last-operation.json`, no formato da versão 1.0. Esse registro é o que o desfazer usa, e uma invariante do projeto diz que ele não muda: um desfazer pendente da 1.0 precisa continuar funcionando na 2.x.

## Decisão

- O histórico fica num arquivo próprio, `~/.sortly/history.json`, no pacote `backend/history`: uma lista JSON com as últimas 50 organizações, da mais recente para a mais antiga. Cada entrada tem `at`, `sourceFolderPath`, `destinationFolderPath`, `movedFiles` e `status` (`done`, `undone` ou `canceled`).
- O `last-operation.json` continua igual, no formato da 1.0. O desfazer continua de um nível e só lê o registro.
- Só entram as organizações que moveram arquivos, as mesmas que gravam o registro do desfazer. Assim, a entrada mais recente é sempre a que o desfazer desfaz, e desfazer marca essa entrada como `undone`. Uma organização interrompida com arquivos movidos entra como `canceled` e também pode ser desfeita.
- O histórico é secundário: arquivo ausente, ilegível ou corrompido vale como lista vazia, com log; uma falha ao gravar vai só para o log, sem aviso na tela.
- A gravação usa `files.WriteAtomic`, como o registro e as preferências.

## Alternativas consideradas

- **Acrescentar o histórico ao `last-operation.json`.** Mudaria o arquivo que a 1.0 lê e que a invariante protege; um campo novo também obrigaria o golden do formato a mudar.
- **Desfazer de vários níveis a partir do histórico.** Exigiria guardar os movimentos de cada organização e validar cada desfazer contra as anteriores. Fica fora do redesign: o protótipo só pede o histórico para consulta.
- **Histórico no `localStorage` do WebView.** Some ao limpar os dados do navegador embutido, o mesmo problema que levou as preferências para `~/.sortly` (#44), e contraria a ADR 0005 (estado no backend).

## Consequências

- A página Histórico lê `ViewState.history`, sempre um array.
- Um registro da 1.0 recuperado ao abrir não tem entrada no histórico; desfazê-lo não marca nada, porque a lista está vazia ou a mais recente já foi desfeita.
- `~/.sortly` passa a ter três arquivos: `last-operation.json`, `settings.json` e `history.json`.

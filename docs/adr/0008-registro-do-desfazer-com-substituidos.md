# ADR 0008 — Registro do desfazer estendido com os substituídos

- **Status:** Aceita
- **Data:** 2026-10-09
- **Autores:** Caio Reis, Claude

## Contexto

Até aqui, o Sortly nunca sobrescrevia: um arquivo com o mesmo nome no destino virava `nome (1).ext`. O redesign deixa escolher o que fazer com duplicados: **Renomear** (o padrão, como na 1.0), **Ignorar** (o arquivo fica na origem) ou **Substituir** (#82).

Substituir apaga, na prática, um arquivo do usuário. Para que o desfazer continue devolvendo tudo como estava, o arquivo substituído precisa ser guardado e o registro do desfazer precisa saber onde. Uma invariante do projeto diz que o formato de `~/.sortly/last-operation.json` é o da versão 1.0: um desfazer pendente da 1.0 precisa funcionar na 2.x, e o golden de `store.encode` confere o formato byte a byte.

## Decisão

- **Antes de substituir, o arquivo existente vai para o backup**, em `~/.sortly/substituidos/<organização>/`, na mesma posição relativa ao destino (`txt/nota.txt`). Cada organização tem a sua pasta, nomeada pelo instante em nanossegundos. Se o arquivo novo não puder ocupar o lugar, o existente volta na hora.
- **O registro ganha dois campos opcionais**, com `omitempty`:
  - `replacedItems`: `[{ "path": onde o arquivo estava, "backup": onde ele ficou guardado }]`;
  - `backupFolder`: a pasta de backup desta organização.

  Sem substituições, o JSON gravado é exatamente o da 1.0, e o golden continua passando. A 1.0 ignora campos que não conhece, e a 2.x lê registros sem eles.
- **O desfazer devolve primeiro os movidos e depois os substituídos**, para o lugar estar livre. Um substituído só volta se `path` estiver dentro da pasta de destino do registro e `backup` dentro de `~/.sortly/substituidos` (a mesma proteção da #54 contra registros editados). Se o lugar estiver ocupado, aplica a regra de conflito. Um substituído que falhar continua no registro, e o desfazer continua disponível.
- **O backup é apagado quando o registro deixa de existir:** depois de um desfazer completo, quando uma nova organização grava outro registro ou quando o registro anterior é apagado porque o novo não pôde ser salvo (#53). Só se apaga o que estiver dentro de `~/.sortly/substituidos`.
- **Só conta como duplicado o arquivo que já estava no destino.** Dois arquivos da mesma organização com o mesmo nome indo para a mesma pasta são renomeados, para nunca ignorar nem substituir o próprio trabalho.
- A política é uma Strategy no `Executor` (ADR 0003), sem `switch`. Renomear não precisa de política: é o movimento de sempre. Política desconhecida (preferência de uma versão futura) vale como renomear.

## Alternativas consideradas

- **Lixeira do sistema.** Cada sistema tem uma API diferente (Shell no Windows, `NSFileManager` no macOS, a especificação freedesktop no Linux), e o desfazer precisaria achar o arquivo de volta na lixeira.
- **Guardar o substituído ao lado, com outro nome.** Deixaria arquivos estranhos na pasta do usuário depois da organização, justamente o que o Sortly tenta evitar.
- **Registro novo, separado do da 1.0.** Duplicaria a lógica de gravação atômica e de validação do desfazer. Os campos opcionais resolvem sem quebrar a compatibilidade.

## Consequências

- `~/.sortly` passa a ter `substituidos/`, que só cresce enquanto houver um desfazer pendente com substituições.
- `undo.Result` ganha `restoredReplaced`, e `organizer.Result` ganha `skippedDuplicates` e `replacedFiles`. A interface conta isso nos avisos.
- O formato completo do registro está em `docs/organization-rules.md` (§7).

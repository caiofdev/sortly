# Changelog

Todas as mudanças que você percebe ao usar o Sortly ficam registradas aqui, da mais recente para a mais antiga.

O formato segue o [Keep a Changelog](https://keepachangelog.com/pt-BR/1.1.0/) e as versões seguem o [Versionamento Semântico](https://semver.org/lang/pt-BR/).

Mantido por Caio Reis & Claude.

## [Não lançado]

### Adicionado
- Documentação do projeto na pasta `docs/`: como o app está organizado, as regras usadas para separar os arquivos e as decisões tomadas para a nova versão.
- README renovado em português e inglês, com espaço para imagens da interface.
- Este changelog.
- Arquivos podem ser organizados entre discos diferentes (por exemplo, do computador para um pendrive). Antes isso falhava.
- Vídeos sem faixa de áudio, como gravações de tela, agora entram na pasta da duração certa em vez de `duration-unknown`.
- A mensagem ao organizar informa quantos arquivos não puderam ser movidos e quantos já estavam no lugar certo, quando houver.
- O app guarda um registro de atividades (log) para ajudar a investigar problemas.
- A fonte da interface vem dentro do app e não depende mais de internet. Antes, sem conexão, a tela usava outra fonte.

### Alterado
- Instaladores bem menores (cerca de 9 MB no Windows, antes 82 MB). Também há versões para macOS (.dmg) e Linux (.deb).
- Ao instalar a nova versão, a versão anterior é removida automaticamente; um desfazer pendente continua disponível.
- Se um arquivo não puder ser movido ou restaurado, os demais continuam. Os que falharam ao desfazer ficam guardados, e você pode tentar desfazer de novo.

### Corrigido
- Se um arquivo falhava no meio da organização, não era mais possível desfazer o que já tinha sido movido.
- Organizar uma pasta sem nada para mover apagava a possibilidade de desfazer a organização anterior.
- Organizando na própria pasta, arquivos que nenhum critério separava eram renomeados para `nome (1)` sem motivo.
- Com o app em inglês, algumas mensagens de erro apareciam em português (e vice-versa).
- Arquivos com nome terminado em ponto não podiam ser organizados no Windows.

### Em andamento
- Nova versão do Sortly construída com Wails, com o objetivo de usar **muito menos memória** e manter exatamente a mesma interface. Acompanhe na [milestone](https://github.com/caiofdev/sortly/milestone/1).

## [1.0.0] - 2026-03-27

Primeira versão pública do Sortly. Por Caio Reis.

### Adicionado
- Organização de arquivos por extensão, data, tamanho, resolução de imagem, duração de vídeo e número de páginas, combináveis entre si.
- Escolha da pasta de origem e da pasta de destino.
- Arrastar e soltar uma pasta ou arquivo para definir a origem.
- Proteção contra sobrescrita: arquivos com o mesmo nome recebem um número, como `foto (1).jpg`.
- Desfazer a última organização, inclusive depois de reabrir o app.
- Central de notificações com o histórico das ações.
- Interface em português e inglês.
- Instalador para Windows.

[Não lançado]: https://github.com/caiofdev/sortly/compare/main...wails-rewrite
[1.0.0]: https://github.com/caiofdev/sortly/releases

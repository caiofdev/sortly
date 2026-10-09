# Changelog

Todas as mudanças que você percebe ao usar o Sortly ficam registradas aqui, da mais recente para a mais antiga.

O formato segue o [Keep a Changelog](https://keepachangelog.com/pt-BR/1.1.0/) e as versões seguem o [Versionamento Semântico](https://semver.org/lang/pt-BR/).

Mantido por Caio Reis & Claude.

## [Não lançado]

### Adicionado
- Visual novo, a partir da logo nova: barra lateral preta com as páginas Organizar e Configurações, amarelo do Sortly para a ação principal e fonte pixel no app todo.
- Os critérios de organização ficam na página Configurações, como interruptores, cada um com um exemplo da pasta que cria.
- Botões "Organizar" e "Desfazer" com texto; caminhos longos de pasta aparecem inteiros, quebrando a linha.
- Avisos no canto da tela ao organizar, ao desfazer e quando algo dá errado. Os de sucesso somem sozinhos em 4 segundos; os de erro ficam até você fechar.
- O sino mostra um ponto amarelo quando há aviso novo. O painel tem "Marcar como lidas" e fecha com Esc.
- Tema claro nas Configurações, com a barra lateral branca. O tema escolhido fica salvo, e a janela já abre na cor dele.
- Ao escolher a pasta de origem, o app mostra quantos arquivos encontrou e em quais pastas eles vão ficar, antes de organizar. A prévia se atualiza ao mudar os critérios.
- Enquanto organiza, o app mostra os arquivos indo de uma pasta para a outra, quantos faltam e qual arquivo está sendo movido. Dá para cancelar no meio: o que já foi movido pode ser desfeito.
- Ao terminar, uma tela de resumo mostra quantos arquivos foram organizados e em quais pastas, com os botões "Abrir pasta de destino", "Desfazer" e "Organizar outra pasta".

### Corrigido
- Ao mover para outro disco ou pendrive, o arquivo original só é apagado depois que a cópia está gravada de fato. Uma queda de energia ou um pendrive removido logo depois não perde mais o arquivo.
- Quando o registro para desfazer não pode ser salvo (disco cheio, sem permissão), o botão "Desfazer" fica desabilitado, em vez de desfazer a organização anterior por engano. Isso vale também depois de reabrir o app.
- O desfazer só devolve arquivos para dentro das pastas da última organização. Um registro danificado ou editado à mão não move mais arquivos para outros lugares.
- No macOS e no Linux, o log do app e a pasta `.sortly` só podem ser lidos pelo seu usuário. O log registra caminhos dos seus arquivos.
- No macOS, uma pasta digitada com maiúsculas diferentes (`~/downloads` em vez de `~/Downloads`) é reconhecida como a mesma: arquivos que já estão no lugar não são mais renomeados para `nome (1)`.
- Desfazer não remove mais pastas que já existiam no destino antes da organização (por exemplo, uma pasta `pdf` vazia criada por você). Só as pastas criadas pelo Sortly são apagadas.
- A tela não fica mais presa em "Organizando…" quando outra ação (como trocar o idioma) termina quase junto com a organização.
- Leitores de tela: os botões de idioma dizem qual está ativo e não repetem mais o nome ("PT-BR PT-BR"), e os botões de organizar e desfazer avisam quando a ação está em andamento.
- O destaque da área de arrastar e soltar não pisca mais ao passar o arquivo sobre o texto e o link dentro dela.

### Alterado
- Mensagens de organizar e desfazer mais curtas, como "Pronto! 12 arquivos organizados", com os detalhes (arquivos já no lugar, sem extensão ou que falharam) logo abaixo. Uma organização em que algum arquivo falhou aparece como erro.
- O idioma e os critérios escolhidos ficam salvos na pasta `.sortly` do usuário, junto do registro para desfazer. Eles não se perdem mais quando os dados internos da janela do app são limpos.
- O app abre direto no idioma escolhido, sem mostrar o português por um instante.

## [2.0.0] - 2026-10-04

Nova versão do Sortly, reconstruída por dentro com a mesma interface. Por Caio Reis & Claude.

### Destaques
- **Mais leve:** instalador de cerca de 9 MB no Windows (antes 82 MB), 15× menos espaço em disco e cerca de 1/3 menos memória (metade no pico ao organizar).
- **Mais rápido:** organizar 1000 arquivos leva cerca de metade do tempo.
- **Mais sistemas:** além do Windows, agora há versões para macOS (.dmg) e Linux (.deb).

### Adicionado
- Arquivos podem ser organizados entre discos diferentes (por exemplo, do computador para um pendrive). Antes isso falhava.
- Vídeos sem faixa de áudio, como gravações de tela, agora entram na pasta da duração certa em vez de `duration-unknown`.
- A mensagem ao organizar informa quantos arquivos não puderam ser movidos e quantos já estavam no lugar certo, quando houver.
- O app guarda um registro de atividades (log) para ajudar a investigar problemas.
- A fonte da interface vem dentro do app e não depende mais de internet. Antes, sem conexão, a tela usava outra fonte.
- README em português e inglês, com imagens da interface, e documentação do projeto na pasta `docs/`.

### Alterado
- Ao instalar a nova versão no Windows, a versão anterior é removida automaticamente; um desfazer pendente continua disponível.
- Se um arquivo não puder ser movido ou restaurado, os demais continuam. Os que falharam ao desfazer ficam guardados, e você pode tentar desfazer de novo.
- O idioma e os critérios escolhidos voltam ao padrão (português, só "Extensão") na primeira abertura da versão 2.0.

### Corrigido
- Se um arquivo falhava no meio da organização, não era mais possível desfazer o que já tinha sido movido.
- Organizar uma pasta sem nada para mover apagava a possibilidade de desfazer a organização anterior.
- Organizando na própria pasta, arquivos que nenhum critério separava eram renomeados para `nome (1)` sem motivo.
- Com o app em inglês, algumas mensagens de erro apareciam em português (e vice-versa).
- Arquivos com nome terminado em ponto não podiam ser organizados no Windows.

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

[Não lançado]: https://github.com/caiofdev/sortly/compare/v2.0.0...HEAD
[2.0.0]: https://github.com/caiofdev/sortly/releases/tag/v2.0.0
[1.0.0]: https://github.com/caiofdev/sortly/releases

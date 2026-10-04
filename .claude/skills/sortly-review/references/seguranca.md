# Segurança e robustez

O Sortly roda local e não tem rede, mas lê arquivos arbitrários do usuário e recebe caminhos do frontend. A ameaça realista é **arquivo malformado** (travar, estourar memória, crash) e **caminho inesperado** (mexer fora da pasta escolhida).

## Entrada do frontend

- [ ] Todo binding valida os argumentos no Go, mesmo que o frontend já valide: caminho vazio, relativo, inexistente, que é arquivo em vez de pasta.
- [ ] Nenhuma operação escreve ou remove fora das pastas de origem/destino escolhidas e de `~/.sortly`.
- [ ] Desfazer confia no registro em disco: valide que `from`/`to` estão dentro das pastas registradas antes de mover ou remover (um registro adulterado ou corrompido não pode apagar `C:\Windows`).
- [ ] Remoção de pasta só se estiver vazia (`os.Remove`), nunca `os.RemoveAll` em pasta do usuário.

## Arquivos malformados (metadata)

- [ ] mp4: tamanhos de box vindos do arquivo são limitados antes de alocar ou pular; box de tamanho 0/1 (até o fim / 64 bits) tratado; laço de leitura sempre avança (box de tamanho menor que o cabeçalho = laço infinito).
- [ ] zip (docx/odt): limite de tamanho descompactado ao ler a entrada (zip bomb); `io.LimitReader`.
- [ ] pdf: parser com limite de leitura e recuperação de `panic` da biblioteca, se ela puder entrar em pânico com entrada ruim.
- [ ] Imagem: ler só o cabeçalho (`image.DecodeConfig`), nunca decodificar a imagem inteira.
- [ ] Todo erro de leitura vira o segmento `*-unknown`, nunca derruba o lote.
- [ ] Fuzzing (`go test -fuzz`) é bem-vindo nos leitores de formato.

## Dados e logs

- [ ] Log registra caminhos (dado do usuário, só local) mas nunca conteúdo de arquivo.
- [ ] Arquivos do app (`~/.sortly/*`, logs) criados com permissão restrita.
- [ ] Rotação/limite de tamanho de log respeitado (5 MB).

## Dependências e build

- [ ] Dependência nova justificada (o que ela faz que a stdlib não faz), mantida e com licença compatível (MIT do projeto).
- [ ] `go.sum` e `package-lock.json` atualizados juntos com o manifesto.
- [ ] Nada de segredo, token ou caminho pessoal versionado.

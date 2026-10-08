# Sistema operacional e sistema de arquivos

O Sortly move arquivos do usuário: um erro aqui é perda de dados. Leia com atenção redobrada qualquer diff em `fs/files`, `fs/paths`, `organizer/executor`, `undo` e `store`.

## Mover e renomear

- [ ] `os.Rename` é atômico **só no mesmo volume**. Entre volumes falha (`EXDEV` no Unix, `ERROR_NOT_SAME_DEVICE` no Windows) → copiar + preservar mtime + `fsync` + apagar a origem só depois da cópia completa.
- [ ] Cópia interrompida (disco cheio, pendrive removido) não deixa o arquivo nos dois lugares nem em nenhum: destino parcial removido, origem intacta.
- [ ] No Windows, `Rename` **falha** se o destino existe; no Unix **sobrescreve** em silêncio. Nunca dependa disso: reserve o destino com `O_CREATE|O_EXCL` (veja `files.Reserve`).
- [ ] TOCTOU: "verificar se existe → depois criar/mover" tem corrida. A verificação e a ação precisam ser uma operação atômica, ou a falha da ação precisa ser tratada.
- [ ] Mover um arquivo para ele mesmo (destino = origem) é no-op, não renomeia para `(1)` (B8).

## Gravação durável

- [ ] Arquivo que não pode corromper (registro do desfazer, preferências) é gravado em temporário **na mesma pasta**, com `Sync()`, e então `Rename` sobre o original.
- [ ] Erro do `Close` do arquivo escrito é verificado (é onde aparecem erros de flush).
- [ ] Leitura tolera arquivo ausente, vazio, truncado ou com JSON inválido (vira "sem registro", com log), sem derrubar o app.

## Caminhos

- [ ] Windows e macOS (APFS padrão) **não diferenciam maiúsculas** de minúsculas: compare com `paths.Equal`/`paths.IsInside`, nunca `==` ou `strings.HasPrefix`.
- [ ] `strings.HasPrefix(path, root)` é bug: `C:\dados2` começa com `C:\dados`. Use `filepath.Rel` ou o helper.
- [ ] Use `filepath` (separador do SO), nunca `path` nem `"/"` concatenado, para caminhos de disco.
- [ ] Caminhos longos (> 260 caracteres) no Windows exigem o prefixo `\\?\` (veja `paths.Native`); nomes terminados em ponto ou espaço só funcionam com ele.
- [ ] Nomes reservados no Windows (`CON`, `NUL`, `COM1`…) e caracteres proibidos (`<>:"|?*`) ao **criar** pastas a partir de dados (extensão do arquivo vira nome de pasta!).
- [ ] Unicode: o macOS pode devolver nomes em NFD (`é` = `e` + acento combinante); comparar com uma string NFC falha. Extensões/nomes de pasta derivados do nome do arquivo podem divergir.
- [ ] `filepath.Ext` difere do `path.extname` do Node em `.gitignore` e `arquivo.` — o projeto usa `paths.Ext`.

## Links, permissões e tipos de arquivo

- [ ] Symlinks e junctions: `os.Stat` segue o link, `os.Lstat` não. Organizar um link move o link, não o alvo? Desfazer segue o link para fora da pasta?
- [ ] Ignorar diretórios, dispositivos, pipes e sockets na varredura (só arquivos regulares).
- [ ] Permissões: dados do app e log (contêm caminhos do usuário) só do dono, `0o700`/`0o600`; pastas criadas na área do usuário (organização) `0o755`. O modo do `MkdirAll`/`OpenFile` só vale na criação: arquivo que já existe precisa de `Chmod`.
- [ ] `os.SameFile` (com `Lstat`) reconhece o mesmo arquivo por outro caminho; comparação só por texto falha em volume que foge do padrão da plataforma.
- [ ] Arquivo sem permissão de leitura/escrita: falha só daquele arquivo, os demais continuam, e o resultado informa (B1).

## Windows

- [ ] Arquivo aberto por outro programa não pode ser movido nem apagado (`ERROR_SHARING_VIOLATION`): trate como falha do item, não do lote.
- [ ] Antivírus/indexador pode segurar o arquivo por instantes logo após criar; exclusão de pasta temporária com retentativa nos scripts.
- [ ] Atributos somente leitura e oculto.

## Datas

- [ ] `ModTime` no fuso **local** para `date-YYYY-MM-DD`; testes com fuso fixo injetado, nunca `time.Local` implícito.
- [ ] Preservar mtime ao copiar entre volumes (senão o critério Data muda depois de mover).
- [ ] Resolução de mtime varia (FAT32: 2 s; NTFS: 100 ns) — testes não comparam com precisão de nanossegundo em FAT.

## Processos e recursos

- [ ] Limite de arquivos abertos (`ulimit -n`, comum 256 no macOS): não mantenha arquivos abertos em laço sobre milhares de itens.
- [ ] Operação longa não bloqueia a janela (binding em goroutine do Wails) e pode ser cancelada pelo `context`.
- [ ] Pastas de dados por SO: `os.UserConfigDir` (logs), `os.UserHomeDir` (`~/.sortly`); nunca caminho fixo.

# Design e padrões de projeto

## Princípios

- **SRP:** cada função/tipo/pacote tem um motivo para mudar. Sinal de violação: nome com "e" (`validateAndMove`), função que valida, calcula, grava e formata.
- **OCP:** critério novo de organização = nova regra no registry, sem mexer no laço do Planner (ADR 0003).
- **LSP:** implementações de uma interface se comportam igual nos contratos (ex.: todo `PageCounter` devolve erro, nunca 0 silencioso, quando não consegue contar).
- **ISP:** interfaces pequenas, só com o que o consumidor usa.
- **DIP:** serviços dependem de interfaces (FileSystem, Store, MetadataReader); a composição concreta fica em `app/wire.go`.
- **DRY com a regra de três:** duplicação de conhecimento (uma regra de negócio em dois lugares) é achado já na segunda vez; duplicação de código parecido só vira abstração na terceira. Abstração precoce é achado também.
- **KISS / YAGNI:** sem parâmetro, interface ou camada "para o futuro" sem uso hoje.
- **Lei de Deméter:** evite `a.B().C().D()` atravessando estruturas de outros pacotes.
- **Coesão alta, acoplamento baixo:** pacote com coisas que mudam juntas; dependências apontando para o núcleo (organizer/undo não importam app; nenhum ciclo).

## Padrões já usados (mantenha a consistência)

| Padrão | Onde | Use quando |
|---|---|---|
| Strategy + Registry | `organizer` regras, `metadata` PageCounter | variação de comportamento por tipo/critério |
| Command | journal do desfazer (move ↔ inverso) | operação que precisa ser revertida |
| Adapter / Gateway | `frontend/services/sortlyGateway.js` | isolar a fronteira com o backend |
| Injeção por construtor | todos os serviços | sempre |
| Planejar × executar | `Planner` puro + `Executor` | cálculo testável sem disco |

- [ ] Padrão novo só quando reduz complexidade real; nomeie-o no PR.
- [ ] Não misture estilos: se o projeto usa construtor + interface, não introduza singleton ou service locator.

## Pacotes Go

- [ ] Nome curto, substantivo, sem `util`/`common`/`misc`.
- [ ] Sem import circular; pacote de baixo nível não importa o de alto nível.
- [ ] O que é usado só dentro do pacote fica não exportado.
- [ ] Arquivos agrupam um conceito (um tipo e seus métodos, uma regra), não "tudo que é pequeno".

## Decisões do projeto (ADRs)

- [ ] O diff respeita as ADRs em `docs/adr/`. Contrariar uma ADR exige uma ADR nova que a substitua — sem isso, é achado 🟠.
- [ ] Invariantes do `CLAUDE.md` (nomes de pastas, formato do registro, backend sem texto para o usuário, interface congelada).

## Legibilidade

- [ ] Nomes revelam intenção; sem abreviação obscura; booleanos como pergunta (`hasUndo`, `isDir`).
- [ ] Função cabe na tela e tem um nível de abstração.
- [ ] Retorno antecipado em vez de `else` aninhado.
- [ ] Comentário explica o porquê; comentário que repete o código é achado 🔵.
- [ ] Números mágicos viram constantes com nome (80 notificações, 10 000 tentativas, 5 MB de log).

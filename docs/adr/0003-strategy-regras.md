# ADR 0003 — Strategy + Registry para os critérios de organização

- **Status:** Aceita
- **Data:** 2026-10-02
- **Autores:** Caio Reis, Claude

## Contexto

Na versão Electron, `buildSegments` decide as subpastas com uma cadeia de seis `if`, cada um com sua própria checagem de tipo de arquivo. Adicionar um critério exige editar essa função, e a complexidade ciclomática cresce a cada critério. O projeto estabeleceu o limite de **CC ≤ 10 por função** e exige ao menos N casos de teste para uma função com CC = N.

## Decisão

Cada critério é uma implementação da interface `SegmentRule` (padrão Strategy):

```go
type SegmentRule interface {
    Key() string                          // ex.: "byExtension", igual à opção da interface
    Applies(f FileInfo) bool              // o critério vale para este arquivo?
    Segment(f FileInfo) (string, error)   // nome da subpasta
}
```

As regras ficam num **registry** com ordem fixa, que define a ordem de aninhamento:

```go
var rules = []SegmentRule{Extension{}, Date{}, Size{}, Resolution{...}, Duration{...}, Pages{...}}
```

O `Planner` percorre o registry e aplica as regras ativas que valem para cada arquivo. A regra de extensão sinaliza "não mover" com o erro sentinela `ErrSkipNoExtension`.

A contagem de páginas usa a mesma ideia: `map[string]PageCounter`, com um único `zipRegexCounter` reaproveitado por docx e odt.

## Alternativas consideradas

| Alternativa | Por que não |
|---|---|
| Portar a cadeia de `if` como está | CC alto, viola OCP e exige muitos casos de teste numa única função |
| Tabela de funções `map[string]func` | Perde a ordem e não separa `Applies` de `Segment` |

## Consequências

- Cada regra tem CC baixo e é testada isoladamente.
- Um critério novo exige só uma struct nova e uma linha no registry (OCP).
- A ordem das pastas fica explícita num único lugar.
- Há uma indireção a mais em relação ao código original.

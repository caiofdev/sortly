// Command cccases cruza a complexidade ciclomática (CC) de cada função Go com a
// quantidade de casos de teste que a chamam diretamente, para conferir a regra
// "casos ≥ CC" (docs/development.md).
//
// A contagem de casos é heurística, feita sobre a AST dos testes:
//   - cada linha de uma tabela percorrida com range conta como um caso;
//   - cada t.Run fora de laço conta como um caso;
//   - cada if fora de laço cujo corpo chama t.Error/t.Fatal conta como um caso;
//   - um teste sem nada disso conta como um caso.
//
// Um teste conta para uma função quando a chama. Para métodos, o teste também
// precisa citar o tipo do receptor. Funções exercitadas só por outras funções
// aparecem como "sem teste direto", o que nem sempre é um problema.
//
// Uso:
//
//	go run ./scripts/cccases [-min 2] [-strict] [dir...]
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	minCC := flag.Int("min", 2, "lista só funções com CC >= min")
	strict := flag.Bool("strict", false, "sai com código 1 se alguma função tiver casos < CC")
	flag.Parse()

	dirs := flag.Args()
	if len(dirs) == 0 {
		dirs = []string{"internal"}
	}
	os.Exit(run(os.Stdout, dirs, *minCC, *strict))
}

func run(w io.Writer, dirs []string, minCC int, strict bool) int {
	var rows []row
	for _, dir := range dirs {
		r, err := analyzeTree(dir)
		if err != nil {
			fmt.Fprintln(os.Stderr, "cccases:", err)
			return 2
		}
		rows = append(rows, r...)
	}
	rows = filterAndSort(rows, minCC)
	short, err := writeReport(w, rows)
	if err != nil {
		fmt.Fprintln(os.Stderr, "cccases:", err)
		return 2
	}
	if strict && short > 0 {
		return 1
	}
	return 0
}

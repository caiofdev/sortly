package main

import (
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"
)

type status int

const (
	statusShort    status = iota // casos < CC
	statusIndirect               // nenhum teste chama a função
	statusOK
)

var statusText = map[status]string{
	statusShort:    "⚠ casos < CC",
	statusIndirect: "↪ sem teste direto",
	statusOK:       "✓",
}

type row struct {
	fn    function
	cases int
	tests []string
}

func (r row) status() status {
	if len(r.tests) == 0 {
		return statusIndirect
	}
	if r.cases < r.fn.cc {
		return statusShort
	}
	return statusOK
}

func matchRows(funcs []function, tests []testFunc) []row {
	receivers := map[string]int{}
	for _, f := range funcs {
		if f.recv != "" {
			receivers[f.name]++
		}
	}
	rows := make([]row, 0, len(funcs))
	for _, f := range funcs {
		r := row{fn: f}
		for _, t := range tests {
			if f.testedBy(t, receivers[f.name] > 1) {
				r.cases += t.cases
				r.tests = append(r.tests, t.name)
			}
		}
		rows = append(rows, r)
	}
	return rows
}

// filterAndSort mantém as funções com CC >= minCC, com os problemas primeiro e,
// dentro de cada grupo, a maior CC primeiro.
func filterAndSort(rows []row, minCC int) []row {
	out := rows[:0]
	for _, r := range rows {
		if r.fn.cc >= minCC {
			out = append(out, r)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		si, sj := out[i].status(), out[j].status()
		if si != sj {
			return si < sj
		}
		return out[i].fn.cc > out[j].fn.cc
	})
	return out
}

// writeReport escreve a tabela em Markdown e devolve quantas funções têm casos < CC.
func writeReport(w io.Writer, rows []row) (int, error) {
	var b strings.Builder
	b.WriteString("| Status | Função | CC | Casos | Testes | Local |\n|---|---|---|---|---|---|\n")
	short := 0
	for _, r := range rows {
		if r.status() == statusShort {
			short++
		}
		fmt.Fprintf(&b, "| %s | `%s.%s` | %d | %d | %s | %s:%d |\n",
			statusText[r.status()], r.fn.pkg, r.fn.label(), r.fn.cc, r.cases,
			strings.Join(r.tests, ", "), filepath.ToSlash(r.fn.pos.Filename), r.fn.pos.Line)
	}
	fmt.Fprintf(&b, "\n%d funções listadas; %d com casos < CC.\n", len(rows), short)
	_, err := io.WriteString(w, b.String())
	return short, err
}

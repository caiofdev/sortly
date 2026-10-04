package main

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func parse(t *testing.T, src string) (*token.FileSet, *ast.File) {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "x.go", "package p\n"+src, 0)
	if err != nil {
		t.Fatal(err)
	}
	return fset, file
}

func firstFunc(t *testing.T, src string) *ast.FuncDecl {
	t.Helper()
	_, file := parse(t, src)
	return file.Decls[0].(*ast.FuncDecl)
}

func TestComplexity(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want int
	}{
		{"linear", `func f() { g() }`, 1},
		{"if e else", `func f(a bool) { if a { g() } else { h() } }`, 2},
		{"for e range", `func f(xs []int) { for i := 0; i < 1; i++ {}; for range xs {} }`, 3},
		{"switch sem contar default", `func f(x int) { switch x { case 1: case 2, 3: default: } }`, 3},
		{"select sem contar default", `func f(c chan int) { select { case <-c: default: } }`, 2},
		{"&& e ||", `func f(a, b, c bool) bool { return a && b || c }`, 3},
		{"função literal conta para a externa", `func f(a bool) { g := func() { if a {} }; g() }`, 2},
		{"type switch", `func f(x any) { switch x.(type) { case int: case string: default: } }`, 3},
		{"select com envio e recebimento", `func f(c chan int) { select { case c <- 1: case v := <-c: _ = v } }`, 3},
	}
	for _, c := range cases {
		if got := complexity(firstFunc(t, c.src)); got != c.want {
			t.Errorf("%s: complexity = %d, want %d", c.name, got, c.want)
		}
	}
}

func TestReceiverType(t *testing.T) {
	cases := map[string]string{
		`func f() {}`:             "",
		`func (s S) f() {}`:       "S",
		`func (s *S) f() {}`:      "S",
		`func (s *S[T]) f() {}`:   "S",
		`func (s S[K, V]) f() {}`: "S",
		`func (S) f() {}`:         "S",
		`func (s **S) f() {}`:     "S",
		`func (s pkg.S) f() {}`:   "",
	}
	for src, want := range cases {
		if got := receiverType(firstFunc(t, src)); got != want {
			t.Errorf("receiverType(%q) = %q, want %q", src, got, want)
		}
	}
}

func TestCountCases(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want int
	}{
		{"sem asserção conta 1", `func TestX(t *testing.T) { f() }`, 1},
		{"ifs de asserção fora de laço", `func TestX(t *testing.T) {
			if f() != 1 { t.Errorf("a") }
			if f() != 2 { t.Fatal("b") }
			if ok { g() }
		}`, 2},
		{"tabela por variável", `func TestX(t *testing.T) {
			cases := []int{1, 2, 3}
			for _, c := range cases { if f(c) { t.Error(c) } }
		}`, 3},
		{"tabela com var", `func TestX(t *testing.T) {
			var cases = map[string]int{"a": 1, "b": 2}
			for k := range cases { _ = k }
		}`, 2},
		{"range direto no literal", `func TestX(t *testing.T) { for _, c := range []int{1, 2} { _ = c } }`, 2},
		{"literal não percorrido não conta", `func TestX(t *testing.T) {
			want := []int{1, 2, 3}
			for _, g := range f() { _ = g }
			_ = want
		}`, 1},
		{"t.Run fora de laço, sem contar o corpo", `func TestX(t *testing.T) {
			t.Run("a", func(t *testing.T) { if f() { t.Error() } })
			t.Run("b", func(t *testing.T) {})
		}`, 2},
		{"t.Run em laço conta pela tabela", `func TestX(t *testing.T) {
			cases := []int{1, 2}
			for _, c := range cases { t.Run("x", func(t *testing.T) { _ = c }) }
		}`, 2},
	}
	for _, c := range cases {
		if got := countCases(firstFunc(t, c.src).Body); got != c.want {
			t.Errorf("%s: countCases = %d, want %d", c.name, got, c.want)
		}
	}
}

func TestTestedBy(t *testing.T) {
	test := testFunc{
		calls:  map[string]bool{"plain": true, ".Run": true, ".Apply": true},
		idents: map[string]bool{"NewExecutor": true},
	}
	cases := []struct {
		name      string
		fn        function
		ambiguous bool
		want      bool
	}{
		{"função chamada", function{name: "plain"}, false, true},
		{"função não chamada", function{name: "other"}, false, false},
		{"método único chamado", function{name: "Run", recv: "Service"}, false, true},
		{"método não chamado", function{name: "Stop", recv: "Service"}, false, false},
		{"método ambíguo sem citar o tipo", function{name: "Run", recv: "Service"}, true, false},
		{"método ambíguo com construtor citado", function{name: "Apply", recv: "Executor"}, true, true},
	}
	for _, c := range cases {
		if got := c.fn.testedBy(test, c.ambiguous); got != c.want {
			t.Errorf("%s: testedBy = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestExpandHelpers(t *testing.T) {
	helpers := map[string]testFunc{
		"setup":  {calls: map[string]bool{"build": true}, idents: map[string]bool{"Store": true}},
		"build":  {calls: map[string]bool{"NewStore": true, "setup": true}, idents: map[string]bool{}},
		"unused": {calls: map[string]bool{"Other": true}, idents: map[string]bool{}},
	}
	test := testFunc{calls: map[string]bool{"setup": true, "Direct": true}, idents: map[string]bool{"x": true}}
	expandHelpers(&test, helpers)

	if !test.calls["build"] {
		t.Error("chamada do helper direto não foi acrescentada")
	}
	if !test.calls["NewStore"] {
		t.Error("chamada do helper indireto não foi acrescentada")
	}
	if !test.idents["Store"] {
		t.Error("identificadores do helper não foram acrescentados")
	}
	if test.calls["Other"] {
		t.Error("helper não usado foi expandido")
	}
	if !test.calls["Direct"] || !test.idents["x"] {
		t.Error("referências próprias do teste se perderam")
	}
	if len(test.calls) != 4 {
		t.Errorf("ciclo setup ↔ build: calls = %v", test.calls)
	}
}

func TestStatus(t *testing.T) {
	cases := []struct {
		r    row
		want status
	}{
		{row{fn: function{cc: 3}}, statusIndirect},
		{row{fn: function{cc: 3}, cases: 2, tests: []string{"T"}}, statusShort},
		{row{fn: function{cc: 3}, cases: 3, tests: []string{"T"}}, statusOK},
		{row{fn: function{cc: 3}, cases: 4, tests: []string{"T"}}, statusOK},
	}
	for _, c := range cases {
		if got := c.r.status(); got != c.want {
			t.Errorf("status(cc=3, casos=%d) = %v, want %v", c.r.cases, got, c.want)
		}
	}
}

func TestFilterAndSort(t *testing.T) {
	rows := []row{
		{fn: function{name: "ok", cc: 5}, cases: 5, tests: []string{"T"}},
		{fn: function{name: "simples", cc: 1}},
		{fn: function{name: "indireta", cc: 2}},
		{fn: function{name: "curta", cc: 3}, cases: 1, tests: []string{"T"}},
		{fn: function{name: "indiretaMaior", cc: 4}},
	}
	got := filterAndSort(rows, 2)
	names := make([]string, len(got))
	for i, r := range got {
		names[i] = r.fn.name
	}

	if len(got) != 4 {
		t.Fatalf("CC abaixo do mínimo não foi filtrada: %v", names)
	}
	if names[0] != "curta" {
		t.Errorf("casos < CC deveria vir primeiro: %v", names)
	}
	if names[1] != "indiretaMaior" || names[2] != "indireta" {
		t.Errorf("dentro do grupo, a maior CC deveria vir primeiro: %v", names)
	}
	if names[3] != "ok" {
		t.Errorf("as funções ok deveriam vir por último: %v", names)
	}
}

func TestRunOnPackage(t *testing.T) {
	dir := t.TempDir()
	write := func(name, src string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(src), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write("p.go", `package p
func Sign(x int) int { if x < 0 { return -1 }; if x > 0 { return 1 }; return 0 }
type S struct{}
func (S) Twice(a bool) bool { if a { return true }; return false }
`)
	write("p_test.go", `package p
import "testing"
func TestSign(t *testing.T) {
	cases := map[int]int{-1: -1, 0: 0}
	for in, want := range cases { if Sign(in) != want { t.Error(in) } }
}
func TestTwice(t *testing.T) { if !(S{}).Twice(true) { t.Error() }; if (S{}).Twice(false) { t.Error() } }
`)
	if err := os.MkdirAll(filepath.Join(dir, "testdata"), 0o755); err != nil {
		t.Fatal(err)
	}
	write(filepath.Join("testdata", "broken.go"), "isso não é Go")

	var out bytes.Buffer
	if code := run(&out, []string{dir}, 2, true); code != 1 {
		t.Errorf("strict com casos < CC: código = %d, want 1\n%s", code, out.String())
	}
	report := out.String()
	for _, want := range []string{"⚠ casos < CC | `p.Sign` | 3 | 2", "✓ | `p.(S).Twice` | 2 | 2", "1 com casos < CC"} {
		if !strings.Contains(report, want) {
			t.Errorf("relatório sem %q:\n%s", want, report)
		}
	}
	if code := run(&out, []string{dir}, 2, false); code != 0 {
		t.Errorf("sem strict: código = %d, want 0", code)
	}
	if code := run(&out, []string{filepath.Join(dir, "nao-existe")}, 2, false); code != 2 {
		t.Errorf("pasta inexistente: código = %d, want 2", code)
	}
}

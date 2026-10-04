package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type function struct {
	pkg, name, recv string
	pos             token.Position
	cc              int
}

type testFunc struct {
	name   string
	calls  map[string]bool
	idents map[string]bool
	cases  int
}

// label devolve o nome no formato do gocyclo: F ou (T).M.
func (f function) label() string {
	if f.recv == "" {
		return f.name
	}
	return "(" + f.recv + ")." + f.name
}

// testedBy diz se o teste chama a função. Quando outro tipo do pacote tem um
// método homônimo, exige também que o teste cite o tipo ou o seu construtor.
func (f function) testedBy(t testFunc, ambiguous bool) bool {
	if f.recv == "" {
		return t.calls[f.name]
	}
	if !t.calls["."+f.name] {
		return false
	}
	return !ambiguous || t.idents[f.recv] || t.idents["New"+f.recv] || t.idents["new"+f.recv]
}

func analyzeTree(root string) ([]row, error) {
	var rows []row
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return err
		}
		if d.Name() == "testdata" {
			return filepath.SkipDir
		}
		r, err := analyzeDir(path)
		rows = append(rows, r...)
		return err
	})
	return rows, err
}

func analyzeDir(dir string) ([]row, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	fset := token.NewFileSet()
	var funcs []function
	var tests []testFunc
	helpers := map[string]testFunc{}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		file, err := parser.ParseFile(fset, filepath.Join(dir, e.Name()), nil, 0)
		if err != nil {
			return nil, err
		}
		if strings.HasSuffix(e.Name(), "_test.go") {
			tests = append(tests, collectTests(file, helpers)...)
		} else {
			funcs = append(funcs, collectFuncs(fset, file)...)
		}
	}
	for i := range tests {
		expandHelpers(&tests[i], helpers)
	}
	return matchRows(funcs, tests), nil
}

func collectFuncs(fset *token.FileSet, file *ast.File) []function {
	var out []function
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		out = append(out, function{
			pkg:  file.Name.Name,
			name: fn.Name.Name,
			recv: receiverType(fn),
			pos:  fset.Position(fn.Pos()),
			cc:   complexity(fn),
		})
	}
	return out
}

func receiverType(fn *ast.FuncDecl) string {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return ""
	}
	expr := fn.Recv.List[0].Type
	for {
		switch x := expr.(type) {
		case *ast.StarExpr:
			expr = x.X
		case *ast.IndexExpr:
			expr = x.X
		case *ast.IndexListExpr:
			expr = x.X
		case *ast.Ident:
			return x.Name
		default:
			return ""
		}
	}
}

// complexity segue a definição do gocyclo: 1 + if, for, range, case e comm
// clause não-default, && e ||. Funções literais contam para a função externa.
func complexity(fn ast.Node) int {
	cc := 1
	ast.Inspect(fn, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt:
			cc++
		case *ast.CaseClause:
			if x.List != nil {
				cc++
			}
		case *ast.CommClause:
			if x.Comm != nil {
				cc++
			}
		case *ast.BinaryExpr:
			if x.Op == token.LAND || x.Op == token.LOR {
				cc++
			}
		}
		return true
	})
	return cc
}

// collectTests devolve os testes do arquivo e guarda em helpers as demais
// funções de teste (montagem de fixtures, fakes), cujas chamadas contam para
// os testes que as usam.
func collectTests(file *ast.File, helpers map[string]testFunc) []testFunc {
	var out []testFunc
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv != nil || fn.Body == nil {
			continue
		}
		calls, idents := references(fn.Body)
		t := testFunc{name: fn.Name.Name, calls: calls, idents: idents}
		if !strings.HasPrefix(fn.Name.Name, "Test") {
			helpers[t.name] = t
			continue
		}
		t.cases = countCases(fn.Body)
		out = append(out, t)
	}
	return out
}

// expandHelpers acrescenta ao teste as chamadas e identificadores dos helpers
// que ele usa, direta ou indiretamente.
func expandHelpers(t *testFunc, helpers map[string]testFunc) {
	seen := map[string]bool{}
	queue := keys(t.calls)
	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]
		h, ok := helpers[name]
		if !ok || seen[name] {
			continue
		}
		seen[name] = true
		for c := range h.calls {
			t.calls[c] = true
			queue = append(queue, c)
		}
		for id := range h.idents {
			t.idents[id] = true
		}
	}
}

func keys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// references devolve os nomes chamados (F, .M e, para pkg.F, também F) e
// todos os identificadores citados no corpo do teste.
func references(body *ast.BlockStmt) (calls, idents map[string]bool) {
	calls, idents = map[string]bool{}, map[string]bool{}
	ast.Inspect(body, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.Ident:
			idents[x.Name] = true
		case *ast.CallExpr:
			switch fun := x.Fun.(type) {
			case *ast.Ident:
				calls[fun.Name] = true
			case *ast.SelectorExpr:
				calls["."+fun.Sel.Name] = true
				calls[fun.Sel.Name] = true
			}
		}
		return true
	})
	return calls, idents
}

package main

import "go/ast"

var assertFuncs = map[string]bool{"Error": true, "Errorf": true, "Fatal": true, "Fatalf": true}

// countCases estima os casos de um teste: linhas de tabelas percorridas com
// range, t.Run fora de laço e ifs de asserção fora de laço. Nunca devolve 0.
func countCases(body *ast.BlockStmt) int {
	n := tableRows(body) + topLevelChecks(body)
	if n == 0 {
		return 1
	}
	return n
}

// tableRows soma os elementos dos literais percorridos com range, seja direto
// (range []T{...}) ou por uma variável declarada no próprio teste.
func tableRows(body *ast.BlockStmt) int {
	literals := map[string]*ast.CompositeLit{}
	rows := 0
	ast.Inspect(body, func(n ast.Node) bool {
		if x, ok := n.(*ast.AssignStmt); ok {
			recordLiterals(literals, x.Lhs, x.Rhs)
		}
		if x, ok := n.(*ast.ValueSpec); ok {
			recordSpec(literals, x)
		}
		if x, ok := n.(*ast.RangeStmt); ok {
			rows += rangedRows(literals, x.X)
		}
		return true
	})
	return rows
}

func recordSpec(literals map[string]*ast.CompositeLit, spec *ast.ValueSpec) {
	for i, name := range spec.Names {
		if i < len(spec.Values) {
			recordLiterals(literals, []ast.Expr{name}, spec.Values[i:i+1])
		}
	}
}

func recordLiterals(literals map[string]*ast.CompositeLit, lhs, rhs []ast.Expr) {
	for i, l := range lhs {
		id, ok := l.(*ast.Ident)
		if !ok || i >= len(rhs) {
			continue
		}
		if lit, ok := rhs[i].(*ast.CompositeLit); ok {
			literals[id.Name] = lit
		}
	}
}

func rangedRows(literals map[string]*ast.CompositeLit, x ast.Expr) int {
	if lit, ok := x.(*ast.CompositeLit); ok {
		return len(lit.Elts)
	}
	id, ok := x.(*ast.Ident)
	if !ok {
		return 0
	}
	if lit, ok := literals[id.Name]; ok {
		return len(lit.Elts)
	}
	return 0
}

// topLevelChecks conta t.Run e ifs de asserção sem entrar em laços (o que está
// num laço já foi contado como linha de tabela) nem no corpo de um t.Run.
func topLevelChecks(body *ast.BlockStmt) int {
	n := 0
	ast.Inspect(body, func(node ast.Node) bool {
		if isLoop(node) {
			return false
		}
		if call, ok := node.(*ast.CallExpr); ok && isMethodCall(call, "Run") {
			n++
			return false
		}
		if stmt, ok := node.(*ast.IfStmt); ok && callsAssert(stmt.Body) {
			n++
		}
		return true
	})
	return n
}

func isMethodCall(call *ast.CallExpr, name string) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	return ok && sel.Sel.Name == name
}

func callsAssert(block *ast.BlockStmt) bool {
	for _, stmt := range block.List {
		expr, ok := stmt.(*ast.ExprStmt)
		if !ok {
			continue
		}
		if call, ok := expr.X.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok && assertFuncs[sel.Sel.Name] {
				return true
			}
		}
	}
	return false
}

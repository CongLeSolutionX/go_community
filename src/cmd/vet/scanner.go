package main

import (
	"go/ast"
	"log"
)

func init() {
	register("scanner",
		"check that bufio.Scanner's Err method is called",
		checkScannerErrCalled,
		assignStmt)
}

func checkScannerErrCalled(f *File, node ast.Node) {
	a := node.(*ast.AssignStmt)

	for i, e := range a.Rhs {
		c, ok := e.(*ast.CallExpr)
		if !ok {
			continue
		}
		s, ok := c.Fun.(*ast.SelectorExpr)
		if !ok {
			continue
		}
		pkg, ok := s.X.(*ast.Ident)
		if !ok {
			continue
		}
		if pkg.Name != "bufio" || pkg.Obj != nil || s.Sel.Name != "NewScanner" {
			continue
		}
		if !findScannerErr(f, a.Lhs[i]) {
			f.Badf(node.Pos(), "missing call to Err() for %v", a.Lhs[i])
		}
	}
}

func value(e ast.Expr) interface{} {
	switch t := e.(type) {
	case *ast.Ident:
		if t.Obj == nil {
			return nil
		}
		return t.Obj.Decl
	case *ast.SelectorExpr:
		return value(t.X)
	default:
		log.Fatalf("unsupported type %T", e)
	}
	panic("unreachable")
}

func findScannerErr(f *File, scanner ast.Expr) bool {
	found := false
	ast.Inspect(f.file, func(n ast.Node) bool {
		c, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		s, ok := c.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if value(s) != value(scanner) {
			return true
		}
		s, ok = c.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if s.Sel.Name == "Err" {
			found = true
		}
		return true
	})
	return found
}

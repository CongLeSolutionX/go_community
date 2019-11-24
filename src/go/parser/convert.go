// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file contains the exported entry points for invoking the parser.

package parser

import (
	"fmt"
	"go/ast"
	"go/token"
	"internal/syntax"
	"strconv"
	"strings"
	"unicode"
)

type converter struct {
	mode  Mode
	fset  *token.FileSet
	lines []int
	errh  func(error)

	astfile *ast.File
	tokfile *token.File

	// Ordinary identifier scopes
	pkgScope   *ast.Scope        // pkgScope.Outer == nil
	topScope   *ast.Scope        // top-most scope; may be pkgScope
	unresolved []*ast.Ident      // unresolved identifiers
	imports    []*ast.ImportSpec // list of imports

	// Label scopes
	// (maintained by open/close LabelScope)
	labelScope  *ast.Scope     // label scope for current function
	targetStack [][]*ast.Ident // stack of unresolved labels
}

func (c *converter) file(sf *syntax.File) *ast.File {
	if sf == nil {
		return nil
	}
	f := &ast.File{
		Package: c.pos(sf.Pos()),
		Name:    c.name(sf.PkgName),
	}
	c.astfile = f
	c.openScope()
	c.pkgScope = c.topScope
	f.Decls = c.decls(sf.DeclList)

	c.closeScope()
	assert(c.topScope == nil, "unbalanced scopes")
	assert(c.labelScope == nil, "unbalanced label scopes")

	// resolve global identifiers within the same file
	i := 0
	for _, ident := range c.unresolved {
		// i <= index for current ident
		assert(ident.Obj == unresolved, "object already resolved")
		ident.Obj = c.pkgScope.Lookup(ident.Name) // also removes unresolved sentinel
		if ident.Obj == nil {
			c.unresolved[i] = ident
			i++
		}
	}

	f.Scope = c.pkgScope
	f.Unresolved = c.unresolved[:i]
	return f
}

func (c *converter) pos(pos syntax.Pos) token.Pos {
	if pos.Line() == 0 {
		return token.NoPos
	}
	return token.Pos(c.lines[pos.Line()-1]) + token.Pos(pos.Col())
}

func (c *converter) name(n *syntax.Name) *ast.Ident {
	if n == nil {
		return nil
	}
	return &ast.Ident{NamePos: c.pos(n.Pos()), Name: n.Value}
}

func (c *converter) basic(l *syntax.BasicLit) *ast.BasicLit {
	var kind token.Token
	switch l.Kind {
	case syntax.IntLit:
		kind = token.INT
	case syntax.FloatLit:
		kind = token.FLOAT
	case syntax.ImagLit:
		kind = token.IMAG
	case syntax.RuneLit:
		kind = token.CHAR
	case syntax.StringLit:
		kind = token.STRING
	}
	return &ast.BasicLit{
		ValuePos: c.pos(l.Pos()),
		Value:    l.Value,
		Kind:     kind,
	}
}

func isValidImport(lit string) bool {
	const illegalChars = `!"#$%&'()*,:;<=>?[\]^{|}` + "`\uFFFD"
	s, _ := strconv.Unquote(lit) // go/scanner returns a legal string literal
	for _, r := range s {
		if !unicode.IsGraphic(r) || unicode.IsSpace(r) || strings.ContainsRune(illegalChars, r) {
			return false
		}
	}
	return s != ""
}

func (c *converter) decls(in []syntax.Decl) []ast.Decl {
	var decls []ast.Decl

	var curGroup *syntax.Group
	var curGen *ast.GenDecl
	for _, sd := range in {
		if fd, ok := sd.(*syntax.FuncDecl); ok {
			if curGen != nil {
				decls = append(decls, curGen)
			}
			curGroup = nil
			curGen = nil
			fd := &ast.FuncDecl{
				Name: c.name(fd.Name),
				Body: c.block(fd.Body),
				Type: c.funcType(fd.Type),
			}

			// Go spec: The scope of an identifier denoting a
			// constant, type, variable, or function (but not
			// method) declared at top level (outside any function)
			// is the package block.
			//
			// init() functions cannot be referred to and there may
			// be more than one - don't put them in the pkgScope
			if fd.Recv == nil && fd.Name.Name != "init" {
				c.declare(fd, nil, c.pkgScope, ast.Fun, fd.Name)
			}
			decls = append(decls, fd)
			continue
		}
		spec, tok, group := c.spec(sd)
		if curGroup == nil || group != curGroup {
			if curGen != nil {
				decls = append(decls, curGen)
			}
			curGroup = group
			curGen = nil
		}
		if curGen == nil {
			curGen = &ast.GenDecl{
				//TokPos: c.pos(sd.Pos()),
				Tok: tok,
			}
		}
		curGen.Specs = append(curGen.Specs, spec)
	}
	if curGen != nil {
		decls = append(decls, curGen)
	}
	return decls
}

func (c *converter) names(names []*syntax.Name) []*ast.Ident {
	var out []*ast.Ident
	for _, name := range names {
		out = append(out, c.name(name))
	}
	return out
}

func (c *converter) spec(sd syntax.Decl) (ast.Spec, token.Token, *syntax.Group) {
	switch x := sd.(type) {
	case *syntax.ImportDecl:
		if !isValidImport(x.Path.Value) {
			c.errh(fmt.Errorf("invalid import path: %s", x.Path.Value))
		}
		imp := &ast.ImportSpec{
			Name: c.name(x.LocalPkgName),
			Path: c.basic(x.Path),
		}
		c.astfile.Imports = append(c.astfile.Imports, imp)
		return imp, token.IMPORT, x.Group
	case *syntax.ConstDecl:
		vs := &ast.ValueSpec{
			Names:  c.names(x.NameList),
			Type:   c.expr(x.Type),
			Values: c.listExpr(x.Values),
		}
		c.declare(vs, 0, c.topScope, ast.Con, vs.Names...)
		return vs, token.CONST, x.Group
	case *syntax.VarDecl:
		vs := &ast.ValueSpec{
			Names:  c.names(x.NameList),
			Type:   c.expr(x.Type),
			Values: c.listExpr(x.Values),
		}
		c.declare(vs, 0, c.topScope, ast.Var, vs.Names...)
		return vs, token.VAR, x.Group
	case *syntax.TypeDecl:
		ts := &ast.TypeSpec{
			Name: c.name(x.Name),
			Type: c.expr(x.Type),
		}
		c.declare(ts, nil, c.topScope, ast.Typ, ts.Name)
		return ts, token.TYPE, x.Group
	default:
		panic(fmt.Sprintf("spec %T", x))
	}
}

func (c *converter) block(sb *syntax.BlockStmt) *ast.BlockStmt {
	if sb == nil {
		return nil
	}
	c.openScope()
	c.openLabelScope()
	b := &ast.BlockStmt{List: c.stmts(sb.List)}
	c.closeLabelScope()
	c.closeScope()
	return b
}

func (c *converter) stmts(sts []syntax.Stmt) []ast.Stmt {
	var out []ast.Stmt
	for _, st := range sts {
		out = append(out, c.stmt(st))
	}
	return out
}

func (c *converter) stmt(st syntax.Stmt) ast.Stmt {
	switch x := st.(type) {
	case nil:
		return nil
	case *syntax.EmptyStmt:
		return &ast.EmptyStmt{}
	case *syntax.BlockStmt:
		return c.block(x)
	case *syntax.LabeledStmt:
		ls := &ast.LabeledStmt{
			Label: c.name(x.Label),
			Stmt:  c.stmt(x.Stmt),
		}
		c.declare(ls, nil, c.labelScope, ast.Lbl, ls.Label)
		return ls

	case *syntax.ExprStmt:
		return &ast.ExprStmt{X: c.expr(x.X)}
	case *syntax.DeclStmt:
		gd := &ast.GenDecl{}
		for _, dc := range x.DeclList {
			spec, _, _ := c.spec(dc)
			gd.Specs = append(gd.Specs, spec)
		}
		return &ast.DeclStmt{
			Decl: gd,
		}
	case *syntax.ForStmt:
		return &ast.ForStmt{
			Body: c.block(x.Body),
		}
	case *syntax.AssignStmt:
		as := &ast.AssignStmt{
			Lhs: c.listExpr(x.Lhs),
			Rhs: c.listExpr(x.Rhs),
		}
		c.shortVarDecl(as)
		return as
	case *syntax.BranchStmt:
		return &ast.BranchStmt{
			Label: c.name(x.Label),
		}
	case *syntax.IfStmt:
		return &ast.IfStmt{
			Init: c.stmt(x.Init),
			Cond: c.expr(x.Cond),
			Body: c.block(x.Then),
			Else: c.stmt(x.Else),
		}
	case *syntax.SwitchStmt:
		var cases []ast.Stmt
		for _, cc := range x.Body {
			cases = append(cases, &ast.CaseClause{
				List: c.listExpr(cc.Cases),
				Body: c.stmts(cc.Body),
			})
		}
		if guard, ok := x.Tag.(*syntax.TypeSwitchGuard); ok {
			ts := &ast.TypeSwitchStmt{
				Init: c.stmt(x.Init),
				Body: &ast.BlockStmt{List: cases},
			}
			expr := c.expr(guard.X)
			if guard.Lhs != nil {
				ts.Assign = &ast.AssignStmt{
					Lhs: []ast.Expr{c.name(guard.Lhs)},
					Rhs: []ast.Expr{expr},
				}
			} else {
				ts.Assign = &ast.ExprStmt{X: expr}
			}
			return ts
		}
		return &ast.SwitchStmt{
			Init: c.stmt(x.Init),
			Tag:  c.expr(x.Tag),
			Body: &ast.BlockStmt{List: cases},
		}
	case *syntax.SelectStmt:
		var clauses []ast.Stmt
		for _, cc := range x.Body {
			clauses = append(clauses, &ast.CommClause{
				Comm: c.stmt(cc.Comm),
				Body: c.stmts(cc.Body),
			})
		}
		return &ast.SelectStmt{
			Body: &ast.BlockStmt{List: clauses},
		}
	case *syntax.SendStmt:
		return &ast.SendStmt{
			Chan:  c.expr(x.Chan),
			Value: c.expr(x.Value),
		}

	case *syntax.ReturnStmt:
		return &ast.ReturnStmt{Results: c.listExpr(x.Results)}
	case *syntax.CallStmt:
		switch x.Tok {
		case syntax.Go:
			return &ast.GoStmt{Call: c.call(x.Call)}
		default: // syntax.Defer
			return &ast.DeferStmt{Call: c.call(x.Call)}
		}
	default:
		panic(fmt.Sprintf("stmt %T", x))
	}
}

func (c *converter) listExpr(expr syntax.Expr) []ast.Expr {
	switch x := expr.(type) {
	case *syntax.ListExpr:
		return c.exprs(x.ElemList)
	case nil:
		return nil
	}
	return []ast.Expr{c.expr(expr)}
}

func (c *converter) exprs(exprs []syntax.Expr) []ast.Expr {
	var out []ast.Expr
	for _, expr := range exprs {
		out = append(out, c.expr(expr))
	}
	return out
}

func (c *converter) expr(expr syntax.Expr) ast.Expr {
	switch x := expr.(type) {
	case nil:
		return nil
	case *syntax.ParenExpr:
		return &ast.ParenExpr{X: c.expr(x.X)}
	case *syntax.BadExpr:
		return &ast.BadExpr{}

	case *syntax.Name:
		return c.name(x)
	case *syntax.BasicLit:
		return c.basic(x)
	case *syntax.CompositeLit:
		return &ast.CompositeLit{
			Type: c.expr(x.Type),
			Elts: c.exprs(x.ElemList),
		}
	case *syntax.FuncLit:
		return &ast.FuncLit{
			Type: c.funcType(x.Type),
			Body: c.block(x.Body),
		}

	case *syntax.SelectorExpr:
		return &ast.SelectorExpr{X: c.expr(x.X), Sel: c.name(x.Sel)}
	case *syntax.CallExpr:
		return c.call(x)
	case *syntax.Operation:
		if x.Y == nil {
			return &ast.UnaryExpr{
				X: c.expr(x.X),
			}
		}
		return &ast.BinaryExpr{
			X: c.expr(x.X),
			Y: c.expr(x.Y),
		}
	case *syntax.IndexExpr:
		return &ast.IndexExpr{
			X:     c.expr(x.X),
			Index: c.expr(x.Index),
		}
	case *syntax.SliceExpr:
		return &ast.SliceExpr{
			X:      c.expr(x.X),
			Low:    c.expr(x.Index[0]),
			High:   c.expr(x.Index[1]),
			Max:    c.expr(x.Index[2]),
			Slice3: x.Full,
		}
	case *syntax.AssertExpr:
		return &ast.TypeAssertExpr{
			X:    c.expr(x.X),
			Type: c.expr(x.Type),
		}

	case *syntax.SliceType:
		return &ast.ArrayType{Elt: c.expr(x.Elem)}
	case *syntax.ArrayType:
		len := c.expr(x.Len)
		if len == nil {
			len = &ast.Ellipsis{}
		}
		return &ast.ArrayType{Len: len, Elt: c.expr(x.Elem)}
	case *syntax.StructType:
		return &ast.StructType{Fields: c.fields(x.FieldList)}
	case *syntax.FuncType:
		return c.funcType(x)
	case *syntax.DotsType:
		return &ast.Ellipsis{Elt: c.expr(x.Elem)}
	case *syntax.InterfaceType:
		return &ast.InterfaceType{Methods: c.fields(x.MethodList)}
	case *syntax.MapType:
		return &ast.MapType{Key: c.expr(x.Key), Value: c.expr(x.Value)}
	case *syntax.KeyValueExpr:
		return &ast.KeyValueExpr{Key: c.expr(x.Key), Value: c.expr(x.Value)}
	case *syntax.ChanType:
		dir := ast.SEND | ast.RECV
		switch x.Dir {
		case syntax.SendOnly:
			dir = ast.SEND
		case syntax.RecvOnly:
			dir = ast.RECV
		}
		return &ast.ChanType{Dir: dir, Value: c.expr(x.Elem)}
	default:
		panic(fmt.Sprintf("expr %T", x))
	}
}

func (c *converter) call(ce *syntax.CallExpr) *ast.CallExpr {
	return &ast.CallExpr{
		Fun:  c.expr(ce.Fun),
		Args: c.exprs(ce.ArgList),
	}
}

func (c *converter) funcType(ft *syntax.FuncType) *ast.FuncType {
	return &ast.FuncType{
		Func:    c.pos(ft.Pos()),
		Params:  c.fields(ft.ParamList),
		Results: c.fields(ft.ResultList),
	}
}

func (c *converter) fields(fields []*syntax.Field) *ast.FieldList {
	fl := &ast.FieldList{}
	for _, f := range fields {
		fl.List = append(fl.List, &ast.Field{
			//Name: c.name(),
			Type: c.expr(f.Type),
		})
		// TODO: declare
	}
	return fl
}

// The unresolved object is a sentinel to mark identifiers that have been added
// to the list of unresolved identifiers. The sentinel is only used for verifying
// internal consistency.
var unresolved = new(ast.Object)

// If x is an identifier, tryResolve attempts to resolve x by looking up
// the object it denotes. If no object is found and collectUnresolved is
// set, x is marked as unresolved and collected in the list of unresolved
// identifiers.
//
func (c *converter) tryResolve(x ast.Expr, collectUnresolved bool) {
	// nothing to do if x is not an identifier or the blank identifier
	ident, _ := x.(*ast.Ident)
	if ident == nil {
		return
	}
	assert(ident.Obj == nil, "identifier already declared or resolved")
	if ident.Name == "_" {
		return
	}
	// try to resolve the identifier
	for s := c.topScope; s != nil; s = s.Outer {
		if obj := s.Lookup(ident.Name); obj != nil {
			ident.Obj = obj
			return
		}
	}
	// all local scopes are known, so any unresolved identifier
	// must be found either in the file scope, package scope
	// (perhaps in another file), or universe scope --- collect
	// them so that they can be resolved later
	if collectUnresolved {
		ident.Obj = unresolved
		c.unresolved = append(c.unresolved, ident)
	}
}

func (c *converter) resolve(x ast.Expr) {
	c.tryResolve(x, true)
}

func assert(cond bool, msg string) {
	if !cond {
		panic("go/parser internal error: " + msg)
	}
}

func (c *converter) openScope() {
	c.topScope = ast.NewScope(c.topScope)
}

func (c *converter) closeScope() {
	c.topScope = c.topScope.Outer
}

func (c *converter) openLabelScope() {
	c.labelScope = ast.NewScope(c.labelScope)
	c.targetStack = append(c.targetStack, nil)
}

func (c *converter) closeLabelScope() {
	// resolve labels
	n := len(c.targetStack) - 1
	scope := c.labelScope
	for _, ident := range c.targetStack[n] {
		ident.Obj = scope.Lookup(ident.Name)
		if ident.Obj == nil && c.mode&DeclarationErrors != 0 {
			//c.error(ident.Pos(), fmt.Sprintf("label %s undefined", ident.Name))
		}
	}
	// pop label scope
	c.targetStack = c.targetStack[0:n]
	c.labelScope = c.labelScope.Outer
}

func (c *converter) declare(decl, data interface{}, scope *ast.Scope, kind ast.ObjKind, idents ...*ast.Ident) {
	for _, ident := range idents {
		assert(ident.Obj == nil, "identifier already declared or resolved")
		obj := ast.NewObj(kind, ident.Name)
		// remember the corresponding declaration for redeclaration
		// errors and global variable resolution/typechecking phase
		obj.Decl = decl
		obj.Data = data
		ident.Obj = obj
		if ident.Name != "_" {
			if alt := scope.Insert(obj); alt != nil && c.mode&DeclarationErrors != 0 {
				prevDecl := ""
				if pos := alt.Pos(); pos.IsValid() {
					prevDecl = fmt.Sprintf("\n\tprevious declaration at %s", c.fset.Position(pos))
				}
				_ = prevDecl
				//c.error(ident.Pos(), fmt.Sprintf("%s redeclared in this block%s", ident.Name, prevDecl))
			}
		}
	}
}

func (c *converter) shortVarDecl(decl *ast.AssignStmt) {
	// Go spec: A short variable declaration may redeclare variables
	// provided they were originally declared in the same block with
	// the same type, and at least one of the non-blank variables is new.
	n := 0 // number of new variables
	for _, x := range decl.Lhs {
		if ident, isIdent := x.(*ast.Ident); isIdent {
			assert(ident.Obj == nil, "identifier already declared or resolved")
			obj := ast.NewObj(ast.Var, ident.Name)
			// remember corresponding assignment for other tools
			obj.Decl = decl
			ident.Obj = obj
			if ident.Name != "_" {
				if alt := c.topScope.Insert(obj); alt != nil {
					ident.Obj = alt // redeclaration
				} else {
					n++ // new declaration
				}
			}
		} else {
			//c.errorExpected(x.Pos(), "identifier on left side of :=")
		}
	}
	if n == 0 && c.mode&DeclarationErrors != 0 {
		//c.error(list[0].Pos(), "no new variables on left side of :=")
	}
}

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Parse input AST and prepare Prog structure.

package main

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/scanner"
	"go/token"
	"os"
	"strings"
)

func parse(name string, src []byte, flags parser.Mode) *ast.File {
	ast1, err := parser.ParseFile(fset, name, src, flags)
	if err != nil {
		if list, ok := err.(scanner.ErrorList); ok {
			// If err is a scanner.ErrorList, its String will print just
			// the first error and then (+n more errors).
			// Instead, turn it into a new Error that will return
			// details for all the errors.
			for _, e := range list {
				fmt.Fprintln(os.Stderr, e)
			}
			os.Exit(2)
		}
		fatalf("parsing %s: %s", name, err)
	}
	return ast1
}

func sourceLine(n ast.Node) int {
	return fset.Position(n.Pos()).Line
}

// ParseGo populates f with information learned from the Go source code
// which was read from the named file. It gathers the C preamble
// attached to the import "C" comment, a list of references to C.xxx,
// a list of exported functions, and the actual AST, to be rewritten and
// printed.
func (f *File) ParseGo(abspath string, src []byte) {
	// Two different parses: once with comments, once without.
	// The printer is not good enough at printing comments in the
	// right place when we start editing the AST behind its back,
	// so we use ast1 to look for the doc comments on import "C"
	// and on exported functions, and we use ast2 for translating
	// and reprinting.
	// In cgo mode, we ignore ast2 and just apply edits directly
	// the text behind ast1. In godefs mode we modify and print ast2.
	ast1 := parse(abspath, src, parser.SkipObjectResolution|parser.ParseComments)
	ast2 := parse(abspath, src, parser.SkipObjectResolution)

	f.Package = ast1.Name.Name
	f.Name = make(map[string]*Name)
	f.NamePos = make(map[*Name]token.Pos)

	// In ast1, find the import "C" line and get any extra C preamble.
	sawC := false
	for _, decl := range ast1.Decls {
		switch decl := decl.(type) {
		case *ast.GenDecl:
			for _, spec := range decl.Specs {
				s, ok := spec.(*ast.ImportSpec)
				if !ok || s.Path.Value != `"C"` {
					continue
				}
				sawC = true
				if s.Name != nil {
					error_(s.Path.Pos(), `cannot rename import "C"`)
				}
				cg := s.Doc
				if cg == nil && len(decl.Specs) == 1 {
					cg = decl.Doc
				}
				if cg != nil {
					if strings.ContainsAny(abspath, "\r\n") {
						// This should have been checked when the file path was first resolved,
						// but we double check here just to be sure.
						fatalf("internal error: ParseGo: abspath contains unexpected newline character: %q", abspath)
					}
					f.Preamble += fmt.Sprintf("#line %d %q\n", sourceLine(cg), abspath)
					f.Preamble += commentText(cg) + "\n"
					f.Preamble += "#line 1 \"cgo-generated-wrapper\"\n"
				}
			}

		case *ast.FuncDecl:
			// Also, reject attempts to declare methods on C.T or *C.T.
			// (The generated code would otherwise accept this
			// invalid input; see issue #57926.)
			if decl.Recv != nil && len(decl.Recv.List) > 0 {
				recvType := decl.Recv.List[0].Type
				if recvType != nil {
					t := recvType
					if star, ok := unparen(t).(*ast.StarExpr); ok {
						t = star.X
					}
					if sel, ok := unparen(t).(*ast.SelectorExpr); ok {
						var buf strings.Builder
						format.Node(&buf, fset, recvType)
						error_(sel.Pos(), `cannot define new methods on non-local type %s`, &buf)
					}
				}
			}
		}

	}
	if !sawC {
		error_(ast1.Package, `cannot find import "C"`)
	}

	// In ast2, strip the import "C" line.
	if *godefs {
		w := 0
		for _, decl := range ast2.Decls {
			d, ok := decl.(*ast.GenDecl)
			if !ok {
				ast2.Decls[w] = decl
				w++
				continue
			}
			ws := 0
			for _, spec := range d.Specs {
				s, ok := spec.(*ast.ImportSpec)
				if !ok || s.Path.Value != `"C"` {
					d.Specs[ws] = spec
					ws++
				}
			}
			if ws == 0 {
				continue
			}
			d.Specs = d.Specs[0:ws]
			ast2.Decls[w] = d
			w++
		}
		ast2.Decls = ast2.Decls[0:w]
	} else {
		for _, decl := range ast2.Decls {
			d, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range d.Specs {
				if s, ok := spec.(*ast.ImportSpec); ok && s.Path.Value == `"C"` {
					// Replace "C" with _ "unsafe", to keep program valid.
					// (Deleting import statement or clause is not safe if it is followed
					// in the source by an explicit semicolon.)
					f.Edit.Replace(f.offset(s.Path.Pos()), f.offset(s.Path.End()), `_ "unsafe"`)
				}
			}
		}
	}

	// Accumulate pointers to uses of C.x.
	if f.Ref == nil {
		f.Ref = make([]*Ref, 0, 8)
	}
	f.walk(ast2, ctxProg, nilTC, (*File).validateIdents)
	f.walk(ast2, ctxProg, nilTC, (*File).saveExprs)

	// Accumulate exported functions.
	// The comments are only on ast1 but we need to
	// save the function bodies from ast2.
	// The first walk fills in ExpFunc, and the
	// second walk changes the entries to
	// refer to ast2 instead.
	f.walk(ast1, ctxProg, nilTC, (*File).saveExport)
	f.walk(ast2, ctxProg, nilTC, (*File).saveExport2)

	f.Comments = ast1.Comments
	f.AST = ast2
}

// Like ast.CommentGroup's Text method but preserves
// leading blank lines, so that line numbers line up.
func commentText(g *ast.CommentGroup) string {
	var pieces []string
	for _, com := range g.List {
		c := com.Text
		// Remove comment markers.
		// The parser has given us exactly the comment text.
		switch c[1] {
		case '/':
			//-style comment (no newline at the end)
			c = c[2:] + "\n"
		case '*':
			/*-style comment */
			c = c[2 : len(c)-2]
		}
		pieces = append(pieces, c)
	}
	return strings.Join(pieces, "")
}

func (f *File) validateIdents(x any, context astContext, typeOf typeContext) {
	if x, ok := x.(*ast.Ident); ok {
		if f.isMangledName(x.Name) {
			error_(x.Pos(), "identifier %q may conflict with identifiers generated by cgo", x.Name)
		}
	}
}

// Save various references we are going to need later.
func (f *File) saveExprs(x any, context astContext, typeOf typeContext) {
	switch x := x.(type) {
	case *ast.Expr:
		switch (*x).(type) {
		case *ast.SelectorExpr:
			f.saveRef(x, context)
		}
	case *ast.CallExpr:
		f.saveCall(x, context)
	case *ast.CompositeLit:
		f.saveLiteral(x, context, typeOf)
	}
}

// Save references to C.xxx for later processing.
func (f *File) saveRef(n *ast.Expr, context astContext) {
	sel := (*n).(*ast.SelectorExpr)
	// For now, assume that the only instance of capital C is when
	// used as the imported package identifier.
	// The parser should take care of scoping in the future, so
	// that we will be able to distinguish a "top-level C" from a
	// local C.
	if l, ok := sel.X.(*ast.Ident); !ok || l.Name != "C" {
		return
	}
	if context == ctxAssign2 {
		context = ctxExpr
	}
	if context == ctxEmbedType {
		error_(sel.Pos(), "cannot embed C type")
	}
	goname := sel.Sel.Name
	if goname == "errno" {
		error_(sel.Pos(), "cannot refer to errno directly; see documentation")
		return
	}
	if goname == "_CMalloc" {
		error_(sel.Pos(), "cannot refer to C._CMalloc; use C.malloc")
		return
	}
	if goname == "malloc" {
		goname = "_CMalloc"
	}
	name := f.Name[goname]
	if name == nil {
		name = &Name{
			Go: goname,
		}
		f.Name[goname] = name
		f.NamePos[name] = sel.Pos()
	}
	f.Ref = append(f.Ref, &Ref{
		Name:    name,
		Expr:    n,
		Context: context,
	})
}

// saveCall saves calls to C.xxx for later processing.
func (f *File) saveCall(call *ast.CallExpr, context astContext) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}
	if l, ok := sel.X.(*ast.Ident); !ok || l.Name != "C" {
		return
	}
	c := &Call{Call: call, Deferred: context == ctxDefer}
	f.Calls = append(f.Calls, c)
}

// saveLiteral saves composite literals for later processing.
func (f *File) saveLiteral(lit *ast.CompositeLit, context astContext, typeOf typeContext) {
	if len(lit.Elts) == 0 {
		return
	}
	// if it's already in field:value form, no need to edit
	if _, ok := lit.Elts[0].(*ast.KeyValueExpr); ok {
		return
	}

	if lit.Type != nil {
		// there is an explicit type, hooray.
		sel, ok := lit.Type.(*ast.SelectorExpr)
		if !ok {
			return // Cannot be "C.whatever" // TODO what about "type CdotWhatever C.whatever"?
		}
		if l, ok := sel.X.(*ast.Ident); !ok || l.Name != "C" {
			return
		}
		c := &Lit{Lit: lit}
		f.Lits = append(f.Lits, c)
		f.LitMap[lit] = c
		return
	}

	// otherwise perhaps the type is implicit, e.g. the slice elements in []C.foo{ {1,2}, {3,4}, {5,6} }
	if typeOf.ty == nil {
		return
	}
	typeOf.path = append([]any{}, typeOf.path...) // must make a private copy
	c := &Lit{Lit: lit, TypeOf: typeOf}
	f.Lits = append(f.Lits, c)
	f.LitMap[lit] = c
}

// doneLiteral marks composite literals in an AST (fragment) as done.
// This is used when a call has been rewritten, which will also cause
// the literal to be processed (and it has to be processed as part of the
// call, otherwise it will cause an "overlapping rewrite" error).
func (f *File) doneLiteral(x any, context astContext, typeOf typeContext) {
	if lit, ok := x.(*ast.CompositeLit); ok {
		if c := f.LitMap[lit]; c != nil {
			c.Done = true
		}
	}
}

// If a function should be exported add it to ExpFunc.
func (f *File) saveExport(x any, context astContext, typeOf typeContext) {
	n, ok := x.(*ast.FuncDecl)
	if !ok {
		return
	}

	if n.Doc == nil {
		return
	}
	for _, c := range n.Doc.List {
		if !strings.HasPrefix(c.Text, "//export ") {
			continue
		}

		name := strings.TrimSpace(c.Text[9:])
		if name == "" {
			error_(c.Pos(), "export missing name")
		}

		if name != n.Name.Name {
			error_(c.Pos(), "export comment has wrong name %q, want %q", name, n.Name.Name)
		}

		doc := ""
		for _, c1 := range n.Doc.List {
			if c1 != c {
				doc += c1.Text + "\n"
			}
		}

		f.ExpFunc = append(f.ExpFunc, &ExpFunc{
			Func:    n,
			ExpName: name,
			Doc:     doc,
		})
		break
	}
}

// Make f.ExpFunc[i] point at the Func from this AST instead of the other one.
func (f *File) saveExport2(x any, context astContext, typeOf typeContext) {
	n, ok := x.(*ast.FuncDecl)
	if !ok {
		return
	}

	for _, exp := range f.ExpFunc {
		if exp.Func.Name.Name == n.Name.Name {
			exp.Func = n
			break
		}
	}
}

type astContext int

const (
	ctxProg astContext = iota
	ctxEmbedType
	ctxType
	ctxStmt
	ctxExpr
	ctxField
	ctxParam
	ctxAssign2 // assignment of a single expression to two variables
	ctxSwitch
	ctxTypeSwitch
	ctxFile
	ctxDecl
	ctxSpec
	ctxDefer
	ctxCall  // any function call other than ctxCall2
	ctxCall2 // function call whose result is assigned to two variables
	ctxSelector
)

// typeContext provides a path into an composite type, to later figure out the
// types of composite parts.  These must be entirely syntactic since there's no
// reliable type information at parsing time (and it's not great at generation time).
// e.g.,
//
//	for T{x, y} the type context for the Compound literal is {T,empty}
//	for T{{a,b}, {c,d}} the inner type contexts are {T,{0}} and {T,{1}} -- T could be slice or struct
//	for T{x:{a,b}, y:{c,d}}, x and y identifiers, the inner type contexts for {a,b} and {c,d} are {T,{"x"} and {T,{"y"}} -- T could be map or struct or slice/array
//	for T{x:{a,b}, y:{c,d}}, x and y NOT identifiers, then T could be a map or slice/array, and
//	         the inner type contexts for {a,b} and {c,d} are {T,{1} and {T,{1}};
//	for T{{u,v}:{a,b}, {x,y}:{c,d} }
//	         the inner type contexts for {u,v} and {x,y} are {T,{0}} and {T,{0}}
//
// When no type is provided (as is the case for the inner compound literals above)
// the existing context is extended in the specified way.
//
// When interpreting typeContexts, given types:
// if T is struct, then integers give positions, strings gave field names
// if T is slice, any value means the element type (could be integer, or `A` from `const A=1`, or `pkg.A` from another package)
// if T is map, 0 means key type, not zero means value type.
type typeContext struct {
	ty   ast.Expr
	path []any // string or integer.
}

var nilTC typeContext

// walk walks the AST x, calling visit(f, x, context) for each node.
func (f *File) walk(x any, context astContext, typeOf typeContext, visit func(*File, any, astContext, typeContext)) {
	visit(f, x, context, typeOf)
	switch n := x.(type) {
	case *ast.Expr:
		f.walk(*n, context, typeOf, visit)

	// everything else just recurs
	default:
		f.walkUnexpected(x, context, typeOf, visit)

	case nil:

	// These are ordered and grouped to match ../../go/ast/ast.go
	case *ast.Field:
		if len(n.Names) == 0 && context == ctxField {
			f.walk(&n.Type, ctxEmbedType, typeOf, visit)
		} else {
			f.walk(&n.Type, ctxType, typeOf, visit)
		}
	case *ast.FieldList:
		for _, field := range n.List {
			f.walk(field, context, typeOf, visit)
		}
	case *ast.BadExpr:
	case *ast.Ident:
	case *ast.Ellipsis:
		f.walk(&n.Elt, ctxType, typeOf, visit)
	case *ast.BasicLit:
	case *ast.FuncLit:
		f.walk(n.Type, ctxType, typeOf, visit)
		f.walk(n.Body, ctxStmt, typeOf, visit)
	case *ast.CompositeLit:
		f.walk(&n.Type, ctxType, typeOf, visit)

		if n.Type != nil {
			typeOf = typeContext{ty: n.Type, path: nil}
		}
		pathLen := len(typeOf.path)
		// open-code the elts visit to get the typeOf context right
		for i, e := range n.Elts {
			if e == nil {
				continue
			}

			if _, ok := e.(*ast.KeyValueExpr); !ok {
				// if not KeyValue, record position, otherwise KeyValue does its own path extension
				typeOf.path = append(typeOf.path, i)
			}

			f.walk(&n.Elts[i], context, typeOf, visit)
			typeOf.path = typeOf.path[:pathLen]
		}

	case *ast.ParenExpr:
		f.walk(&n.X, context, typeOf, visit)
	case *ast.SelectorExpr:
		f.walk(&n.X, ctxSelector, typeOf, visit)
	case *ast.IndexExpr:
		f.walk(&n.X, ctxExpr, typeOf, visit)
		f.walk(&n.Index, ctxExpr, typeOf, visit)
	case *ast.SliceExpr:
		f.walk(&n.X, ctxExpr, typeOf, visit)
		if n.Low != nil {
			f.walk(&n.Low, ctxExpr, typeOf, visit)
		}
		if n.High != nil {
			f.walk(&n.High, ctxExpr, typeOf, visit)
		}
		if n.Max != nil {
			f.walk(&n.Max, ctxExpr, typeOf, visit)
		}
	case *ast.TypeAssertExpr:
		f.walk(&n.X, ctxExpr, typeOf, visit)
		f.walk(&n.Type, ctxType, typeOf, visit)
	case *ast.CallExpr:
		if context == ctxAssign2 {
			f.walk(&n.Fun, ctxCall2, typeOf, visit)
		} else {
			f.walk(&n.Fun, ctxCall, typeOf, visit)
		}
		f.walk(n.Args, ctxExpr, typeOf, visit)
	case *ast.StarExpr:
		f.walk(&n.X, context, typeOf, visit)
	case *ast.UnaryExpr:
		f.walk(&n.X, ctxExpr, typeOf, visit)
	case *ast.BinaryExpr:
		f.walk(&n.X, ctxExpr, typeOf, visit)
		f.walk(&n.Y, ctxExpr, typeOf, visit)
	case *ast.KeyValueExpr:

		if typeOf.ty != nil {
			pathLen := len(typeOf.path)
			typeOf.path = append(typeOf.path, 0) // marks key, if a map.
			f.walk(&n.Key, ctxExpr, typeOf, visit)

			if ident, ok := n.Key.(*ast.Ident); ok {
				typeOf.path[pathLen] = ident.Name // could be map or struct
			} else {
				typeOf.path[pathLen] = 1 // must be map, record not-zero
			}
			f.walk(&n.Value, ctxExpr, typeOf, visit)
			typeOf.path = typeOf.path[:pathLen]
		} else {
			f.walk(&n.Key, ctxExpr, typeOf, visit)
			f.walk(&n.Value, ctxExpr, typeOf, visit)
		}

	case *ast.ArrayType:
		f.walk(&n.Len, ctxExpr, typeOf, visit)
		f.walk(&n.Elt, ctxType, typeOf, visit)
	case *ast.StructType:
		f.walk(n.Fields, ctxField, typeOf, visit)
	case *ast.FuncType:
		if tparams := funcTypeTypeParams(n); tparams != nil {
			f.walk(tparams, ctxParam, typeOf, visit)
		}
		f.walk(n.Params, ctxParam, typeOf, visit)
		if n.Results != nil {
			f.walk(n.Results, ctxParam, typeOf, visit)
		}
	case *ast.InterfaceType:
		f.walk(n.Methods, ctxField, typeOf, visit)
	case *ast.MapType:
		f.walk(&n.Key, ctxType, typeOf, visit)
		f.walk(&n.Value, ctxType, typeOf, visit)
	case *ast.ChanType:
		f.walk(&n.Value, ctxType, typeOf, visit)

	case *ast.BadStmt:
	case *ast.DeclStmt:
		f.walk(n.Decl, ctxDecl, typeOf, visit)
	case *ast.EmptyStmt:
	case *ast.LabeledStmt:
		f.walk(n.Stmt, ctxStmt, typeOf, visit)
	case *ast.ExprStmt:
		f.walk(&n.X, ctxExpr, typeOf, visit)
	case *ast.SendStmt:
		f.walk(&n.Chan, ctxExpr, typeOf, visit)
		f.walk(&n.Value, ctxExpr, typeOf, visit)
	case *ast.IncDecStmt:
		f.walk(&n.X, ctxExpr, typeOf, visit)
	case *ast.AssignStmt:
		f.walk(n.Lhs, ctxExpr, typeOf, visit)
		if len(n.Lhs) == 2 && len(n.Rhs) == 1 {
			f.walk(n.Rhs, ctxAssign2, typeOf, visit)
		} else {
			f.walk(n.Rhs, ctxExpr, typeOf, visit)
		}
	case *ast.GoStmt:
		f.walk(n.Call, ctxExpr, typeOf, visit)
	case *ast.DeferStmt:
		f.walk(n.Call, ctxDefer, typeOf, visit)
	case *ast.ReturnStmt:
		f.walk(n.Results, ctxExpr, typeOf, visit)
	case *ast.BranchStmt:
	case *ast.BlockStmt:
		f.walk(n.List, context, typeOf, visit)
	case *ast.IfStmt:
		f.walk(n.Init, ctxStmt, typeOf, visit)
		f.walk(&n.Cond, ctxExpr, typeOf, visit)
		f.walk(n.Body, ctxStmt, typeOf, visit)
		f.walk(n.Else, ctxStmt, typeOf, visit)
	case *ast.CaseClause:
		if context == ctxTypeSwitch {
			context = ctxType
		} else {
			context = ctxExpr
		}
		f.walk(n.List, context, typeOf, visit)
		f.walk(n.Body, ctxStmt, typeOf, visit)
	case *ast.SwitchStmt:
		f.walk(n.Init, ctxStmt, typeOf, visit)
		f.walk(&n.Tag, ctxExpr, typeOf, visit)
		f.walk(n.Body, ctxSwitch, typeOf, visit)
	case *ast.TypeSwitchStmt:
		f.walk(n.Init, ctxStmt, typeOf, visit)
		f.walk(n.Assign, ctxStmt, typeOf, visit)
		f.walk(n.Body, ctxTypeSwitch, typeOf, visit)
	case *ast.CommClause:
		f.walk(n.Comm, ctxStmt, typeOf, visit)
		f.walk(n.Body, ctxStmt, typeOf, visit)
	case *ast.SelectStmt:
		f.walk(n.Body, ctxStmt, typeOf, visit)
	case *ast.ForStmt:
		f.walk(n.Init, ctxStmt, typeOf, visit)
		f.walk(&n.Cond, ctxExpr, typeOf, visit)
		f.walk(n.Post, ctxStmt, typeOf, visit)
		f.walk(n.Body, ctxStmt, typeOf, visit)
	case *ast.RangeStmt:
		f.walk(&n.Key, ctxExpr, typeOf, visit)
		f.walk(&n.Value, ctxExpr, typeOf, visit)
		f.walk(&n.X, ctxExpr, typeOf, visit)
		f.walk(n.Body, ctxStmt, typeOf, visit)

	case *ast.ImportSpec:
	case *ast.ValueSpec:
		f.walk(&n.Type, ctxType, typeOf, visit)
		if len(n.Names) == 2 && len(n.Values) == 1 {
			f.walk(&n.Values[0], ctxAssign2, typeOf, visit)
		} else {
			f.walk(n.Values, ctxExpr, typeOf, visit)
		}
	case *ast.TypeSpec:
		if tparams := typeSpecTypeParams(n); tparams != nil {
			f.walk(tparams, ctxParam, typeOf, visit)
		}
		f.walk(&n.Type, ctxType, typeOf, visit)

	case *ast.BadDecl:
	case *ast.GenDecl:
		f.walk(n.Specs, ctxSpec, typeOf, visit)
	case *ast.FuncDecl:
		if n.Recv != nil {
			f.walk(n.Recv, ctxParam, typeOf, visit)
		}
		f.walk(n.Type, ctxType, typeOf, visit)
		if n.Body != nil {
			f.walk(n.Body, ctxStmt, typeOf, visit)
		}

	case *ast.File:
		f.walk(n.Decls, ctxDecl, typeOf, visit)

	case *ast.Package:
		for _, file := range n.Files {
			f.walk(file, ctxFile, typeOf, visit)
		}

	case []ast.Decl:
		for _, d := range n {
			f.walk(d, context, typeOf, visit)
		}
	case []ast.Expr:
		for i := range n {
			f.walk(&n[i], context, typeOf, visit)
		}
	case []ast.Stmt:
		for _, s := range n {
			f.walk(s, context, typeOf, visit)
		}
	case []ast.Spec:
		for _, s := range n {
			f.walk(s, context, typeOf, visit)
		}
	}
}

// If x is of the form (T), unparen returns unparen(T), otherwise it returns x.
func unparen(x ast.Expr) ast.Expr {
	if p, isParen := x.(*ast.ParenExpr); isParen {
		x = unparen(p.X)
	}
	return x
}

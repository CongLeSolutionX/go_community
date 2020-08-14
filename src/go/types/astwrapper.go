// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"fmt"
	"go/ast"
	"go/token"
)

type fileWrapper struct{ *ast.File }

func (w fileWrapper) Pos() astPosition     { return w.File.Pos() }
func (w fileWrapper) End() astPosition     { return w.File.End() }
func (w fileWrapper) Unwrap() ast.Node     { return w.File }
func (w fileWrapper) Decls() astDeclList   { return declListWrapper(w.File.Decls) }
func (w fileWrapper) Name() astIdent       { return wrapIdent(w.File.Name) }
func (w fileWrapper) Package() astPosition { return w.File.Package }

type declListWrapper []ast.Decl

func (w declListWrapper) Len() int           { return len(w) }
func (w declListWrapper) Decl(i int) astDecl { return wrapDecl(w[i]) }

type declWrapper struct {
	ast.Decl
}

func wrapDecl(d ast.Decl) astDecl {
	if d == nil {
		return nil
	}
	return declWrapper{d}
}

func (w declWrapper) Kind() declKind {
	switch w.Decl.(type) {
	case nil:
		return nilDeclKind
	case *ast.BadDecl:
		return badDeclKind
	case *ast.GenDecl:
		return genDeclKind
	case *ast.FuncDecl:
		return funcDeclKind
	}
	panic(fmt.Sprintf("decl %T", w.Decl))
}

func (w declWrapper) Pos() astPosition { return w.Decl.Pos() }
func (w declWrapper) End() astPosition { return w.Decl.End() }
func (w declWrapper) Unwrap() ast.Node { return w.Decl }

func (w declWrapper) BadDecl() astBadDecl {
	d, _ := w.Decl.(*ast.BadDecl)
	return wrapBadDecl(d)
}

func (w declWrapper) GenDecl() astGenDecl {
	d, _ := w.Decl.(*ast.GenDecl)
	return wrapGenDecl(d)
}

func (w declWrapper) FuncDecl() astFuncDecl {
	d, _ := w.Decl.(*ast.FuncDecl)
	return wrapFuncDecl(d)
}

type aDecl struct{}

func (aDecl) BadDecl() astBadDecl   { return nil }
func (aDecl) GenDecl() astGenDecl   { return nil }
func (aDecl) FuncDecl() astFuncDecl { return nil }

type badDeclWrapper struct {
	aDecl
	d *ast.BadDecl
}

func (badDeclWrapper) Kind() declKind        { return badDeclKind }
func (w badDeclWrapper) BadDecl() astBadDecl { return w }

func (w badDeclWrapper) Unwrap() ast.Node { return w.d }
func (w badDeclWrapper) Pos() astPosition { return w.d.Pos() }
func (w badDeclWrapper) End() astPosition { return w.d.End() }

func wrapBadDecl(d *ast.BadDecl) astBadDecl {
	if d == nil {
		return nil
	}
	return badDeclWrapper{d: d}
}

type genDeclWrapper struct {
	aDecl
	d *ast.GenDecl
}

func (genDeclWrapper) Kind() declKind        { return genDeclKind }
func (w genDeclWrapper) GenDecl() astGenDecl { return w }
func (w genDeclWrapper) Specs() astSpecList  { return specListWrapper(w.d.Specs) }
func (w genDeclWrapper) Tok() token.Token    { return w.d.Tok }

func (w genDeclWrapper) Unwrap() ast.Node { return w.d }
func (w genDeclWrapper) Pos() astPosition { return w.d.Pos() }
func (w genDeclWrapper) End() astPosition { return w.d.End() }

func wrapGenDecl(d *ast.GenDecl) astGenDecl {
	if d == nil {
		return nil
	}
	return genDeclWrapper{d: d}
}

type funcDeclWrapper struct {
	aDecl
	d *ast.FuncDecl
}

func (w funcDeclWrapper) Kind() declKind        { return funcDeclKind }
func (w funcDeclWrapper) FuncDecl() astFuncDecl { return w }
func (w funcDeclWrapper) Recv() astFieldList    { return wrapFieldList(w.d.Recv) }
func (w funcDeclWrapper) Name() astIdent        { return wrapIdent(w.d.Name) }
func (w funcDeclWrapper) Type() astFuncType     { return wrapFuncType(w.d.Type) }
func (w funcDeclWrapper) Body() astBlockStmt    { return wrapBlockStmt(w.d.Body) }

func (w funcDeclWrapper) Unwrap() ast.Node { return w.d }
func (w funcDeclWrapper) Pos() astPosition { return w.d.Pos() }
func (w funcDeclWrapper) End() astPosition { return w.d.End() }

func wrapFuncDecl(d *ast.FuncDecl) astFuncDecl {
	if d == nil {
		return nil
	}
	return funcDeclWrapper{d: d}
}

type specListWrapper []ast.Spec

func (w specListWrapper) Len() int           { return len(w) }
func (w specListWrapper) Spec(i int) astSpec { return wrapSpec(w[i]) }

type specWrapper struct{ ast.Spec }

func (w specWrapper) Pos() astPosition { return w.Spec.Pos() }
func (w specWrapper) End() astPosition { return w.Spec.End() }
func (w specWrapper) Unwrap() ast.Node { return w.Spec }

func (w specWrapper) Kind() specKind {
	switch w.Spec.(type) {
	case nil:
		return NilSpecKind
	case (*ast.ImportSpec):
		return ImportSpecKind
	case (*ast.ValueSpec):
		return ValueSpecKind
	case (*ast.TypeSpec):
		return TypeSpecKind
	}
	panic(fmt.Sprintf("spec %T", w.Spec))
}

func (w specWrapper) ValueSpec() astValueSpec {
	s, _ := w.Spec.(*ast.ValueSpec)
	return wrapValueSpec(s)
}

func (w specWrapper) ImportSpec() astImportSpec {
	s, _ := w.Spec.(*ast.ImportSpec)
	return wrapImportSpec(s)
}

func (w specWrapper) TypeSpec() astTypeSpec {
	s, _ := w.Spec.(*ast.TypeSpec)
	return wrapTypeSpec(s)
}

func wrapSpec(spec ast.Spec) astSpec {
	if spec == nil {
		return nil
	}
	return specWrapper{spec}
}

type aSpec struct{}

func (aSpec) ImportSpec() astImportSpec { return nil }
func (aSpec) ValueSpec() astValueSpec   { return nil }
func (aSpec) TypeSpec() astTypeSpec     { return nil }

type importSpecWrapper struct {
	aSpec
	s *ast.ImportSpec
}

func (w importSpecWrapper) Kind() specKind            { return ImportSpecKind }
func (w importSpecWrapper) ImportSpec() astImportSpec { return w }
func (w importSpecWrapper) Path() astBasicLit         { return wrapBasicLit(w.s.Path) }
func (w importSpecWrapper) Name() astIdent            { return wrapIdent(w.s.Name) }

func (w importSpecWrapper) Pos() astPosition { return w.s.Pos() }
func (w importSpecWrapper) End() astPosition { return w.s.End() }
func (w importSpecWrapper) Unwrap() ast.Node { return w.s }

func wrapImportSpec(s *ast.ImportSpec) astImportSpec {
	if s == nil {
		return nil
	}
	return importSpecWrapper{s: s}
}

type valueSpecWrapper struct {
	aSpec
	s *ast.ValueSpec
}

func (w valueSpecWrapper) Kind() specKind          { return ValueSpecKind }
func (w valueSpecWrapper) ValueSpec() astValueSpec { return w }
func (w valueSpecWrapper) Names() astIdentList     { return identListWrapper(w.s.Names) }
func (w valueSpecWrapper) Type() astExpr           { return wrapExpr(w.s.Type) }
func (w valueSpecWrapper) Values() astExprList     { return exprListWrapper(w.s.Values) }

func (w valueSpecWrapper) Pos() astPosition { return w.s.Pos() }
func (w valueSpecWrapper) End() astPosition { return w.s.End() }
func (w valueSpecWrapper) Unwrap() ast.Node { return w.s }

func wrapValueSpec(s *ast.ValueSpec) astValueSpec {
	if s == nil {
		return nil
	}
	return valueSpecWrapper{s: s}
}

type typeSpecWrapper struct {
	aSpec
	s *ast.TypeSpec
}

func (w typeSpecWrapper) Kind() specKind        { return TypeSpecKind }
func (w typeSpecWrapper) TypeSpec() astTypeSpec { return w }
func (w typeSpecWrapper) Name() astIdent        { return wrapIdent(w.s.Name) }
func (w typeSpecWrapper) Assign() astPosition   { return w.s.Assign }
func (w typeSpecWrapper) Type() astExpr         { return wrapExpr(w.s.Type) }

func (w typeSpecWrapper) Pos() astPosition { return w.s.Pos() }
func (w typeSpecWrapper) End() astPosition { return w.s.End() }
func (w typeSpecWrapper) Unwrap() ast.Node { return w.s }

func wrapTypeSpec(s *ast.TypeSpec) astTypeSpec {
	if s == nil {
		return nil
	}
	return typeSpecWrapper{s: s}
}

func astNewValueSpec() astValueSpec {
	return wrapValueSpec(new(ast.ValueSpec))
}

type exprWrapper struct {
	ast.Expr
}

func (w exprWrapper) Kind() expressionKind {
	switch w.Expr.(type) {
	case nil:
		return nilExprKind
	case *ast.BadExpr:
		return badExprKind
	case *ast.Ident:
		return identKind
	case *ast.Ellipsis:
		return ellipsisKind
	case *ast.BasicLit:
		return basicLitKind
	case *ast.FuncLit:
		return funcLitKind
	case *ast.CompositeLit:
		return compositeLitKind
	case *ast.ParenExpr:
		return parenExprKind
	case *ast.SelectorExpr:
		return selectorExprKind
	case *ast.IndexExpr:
		return indexExprKind
	case *ast.SliceExpr:
		return sliceExprKind
	case *ast.TypeAssertExpr:
		return typeAssertExprKind
	case *ast.CallExpr:
		return callExprKind
	case *ast.StarExpr:
		return starExprKind
	case *ast.UnaryExpr:
		return unaryExprKind
	case *ast.BinaryExpr:
		return binaryExprKind
	case *ast.KeyValueExpr:
		return keyValueExprKind
	case *ast.ArrayType:
		return arrayTypeKind
	case *ast.StructType:
		return structTypeKind
	case *ast.FuncType:
		return funcTypeKind
	case *ast.InterfaceType:
		return interfaceTypeKind
	case *ast.MapType:
		return mapTypeKind
	case *ast.ChanType:
		return chanTypeKind
	}
	panic(fmt.Sprintf("expr %T", w.Expr))
}

func (w exprWrapper) Reify() astExpr {
	switch w.Expr.(type) {
	case nil:
		return nil
	case *ast.BadExpr:
		return w.BadExpr()
	case *ast.Ident:
		return w.Ident()
	case *ast.Ellipsis:
		return w.Ellipsis()
	case *ast.BasicLit:
		return w.BasicLit()
	case *ast.FuncLit:
		return w.FuncLit()
	case *ast.CompositeLit:
		return w.CompositeLit()
	case *ast.ParenExpr:
		return w.ParenExpr()
	case *ast.SelectorExpr:
		return w.SelectorExpr()
	case *ast.IndexExpr:
		return w.IndexExpr()
	case *ast.SliceExpr:
		return w.SliceExpr()
	case *ast.TypeAssertExpr:
		return w.TypeAssertExpr()
	case *ast.CallExpr:
		return w.CallExpr()
	case *ast.StarExpr:
		return w.StarExpr()
	case *ast.UnaryExpr:
		return w.UnaryExpr()
	case *ast.BinaryExpr:
		return w.BinaryExpr()
	case *ast.KeyValueExpr:
		return w.KeyValueExpr()
	case *ast.ArrayType:
		return w.ArrayType()
	case *ast.StructType:
		return w.StructType()
	case *ast.FuncType:
		return w.FuncType()
	case *ast.InterfaceType:
		return w.InterfaceType()
	case *ast.MapType:
		return w.MapType()
	case *ast.ChanType:
		return w.ChanType()
	}
	panic(fmt.Sprintf("expr %T", w.Expr))
}

func (w exprWrapper) Pos() astPosition { return w.Expr.Pos() }
func (w exprWrapper) End() astPosition { return w.Expr.End() }
func (w exprWrapper) Unwrap() ast.Node { return w.Expr }

func (w exprWrapper) BadExpr() astBadExpr {
	x, _ := w.Expr.(*ast.BadExpr)
	return wrapBadExpr(x)
}
func (w exprWrapper) Ident() astIdent {
	ident, _ := w.Expr.(*ast.Ident)
	return wrapIdent(ident)
}
func (w exprWrapper) Ellipsis() astEllipsis {
	x, _ := w.Expr.(*ast.Ellipsis)
	return wrapEllipsis(x)
}
func (w exprWrapper) BasicLit() astBasicLit {
	x, _ := w.Expr.(*ast.BasicLit)
	return wrapBasicLit(x)
}
func (w exprWrapper) FuncLit() astFuncLit {
	x, _ := w.Expr.(*ast.FuncLit)
	return wrapFuncLit(x)
}
func (w exprWrapper) CompositeLit() astCompositeLit {
	x, _ := w.Expr.(*ast.CompositeLit)
	return wrapCompositeLit(x)
}
func (w exprWrapper) ParenExpr() astParenExpr {
	x, _ := w.Expr.(*ast.ParenExpr)
	return wrapParenExpr(x)
}
func (w exprWrapper) SelectorExpr() astSelectorExpr {
	x, _ := w.Expr.(*ast.SelectorExpr)
	return wrapSelectorExpr(x)
}
func (w exprWrapper) IndexExpr() astIndexExpr {
	x, _ := w.Expr.(*ast.IndexExpr)
	return wrapIndexExpr(x)
}
func (w exprWrapper) SliceExpr() astSliceExpr {
	x, _ := w.Expr.(*ast.SliceExpr)
	return wrapSliceExpr(x)
}
func (w exprWrapper) TypeAssertExpr() astTypeAssertExpr {
	x, _ := w.Expr.(*ast.TypeAssertExpr)
	return wrapTypeAssertExpr(x)
}
func (w exprWrapper) CallExpr() astCallExpr {
	x, _ := w.Expr.(*ast.CallExpr)
	return wrapCallExpr(x)
}
func (w exprWrapper) StarExpr() astStarExpr {
	x, _ := w.Expr.(*ast.StarExpr)
	return wrapStarExpr(x)
}
func (w exprWrapper) UnaryExpr() astUnaryExpr {
	x, _ := w.Expr.(*ast.UnaryExpr)
	return wrapUnaryExpr(x)
}
func (w exprWrapper) BinaryExpr() astBinaryExpr {
	x, _ := w.Expr.(*ast.BinaryExpr)
	return wrapBinaryExpr(x)
}
func (w exprWrapper) KeyValueExpr() astKeyValueExpr {
	x, _ := w.Expr.(*ast.KeyValueExpr)
	return wrapKeyValueExpr(x)
}
func (w exprWrapper) ArrayType() astArrayType {
	x, _ := w.Expr.(*ast.ArrayType)
	return wrapArrayType(x)
}
func (w exprWrapper) StructType() astStructType {
	x, _ := w.Expr.(*ast.StructType)
	return wrapStructType(x)
}
func (w exprWrapper) FuncType() astFuncType {
	x, _ := w.Expr.(*ast.FuncType)
	return wrapFuncType(x)
}
func (w exprWrapper) InterfaceType() astInterfaceType {
	x, _ := w.Expr.(*ast.InterfaceType)
	return wrapInterfaceType(x)
}
func (w exprWrapper) MapType() astMapType {
	x, _ := w.Expr.(*ast.MapType)
	return wrapMapType(x)
}
func (w exprWrapper) ChanType() astChanType {
	x, _ := w.Expr.(*ast.ChanType)
	return wrapChanType(x)
}

// TODO: perhaps we could eagerly type switch here, and eliminate exprWrapper?
// Need to do some investigation.
func wrapExpr(expr ast.Expr) astExpr {
	if expr == nil {
		return nil
	}
	return exprWrapper{expr}
}

type anExprNode struct{}

func (anExprNode) BadExpr() astBadExpr               { return nil }
func (anExprNode) Ident() astIdent                   { return nil }
func (anExprNode) Ellipsis() astEllipsis             { return nil }
func (anExprNode) BasicLit() astBasicLit             { return nil }
func (anExprNode) FuncLit() astFuncLit               { return nil }
func (anExprNode) CompositeLit() astCompositeLit     { return nil }
func (anExprNode) ParenExpr() astParenExpr           { return nil }
func (anExprNode) SelectorExpr() astSelectorExpr     { return nil }
func (anExprNode) IndexExpr() astIndexExpr           { return nil }
func (anExprNode) SliceExpr() astSliceExpr           { return nil }
func (anExprNode) TypeAssertExpr() astTypeAssertExpr { return nil }
func (anExprNode) CallExpr() astCallExpr             { return nil }
func (anExprNode) StarExpr() astStarExpr             { return nil }
func (anExprNode) UnaryExpr() astUnaryExpr           { return nil }
func (anExprNode) BinaryExpr() astBinaryExpr         { return nil }
func (anExprNode) KeyValueExpr() astKeyValueExpr     { return nil }
func (anExprNode) ArrayType() astArrayType           { return nil }
func (anExprNode) StructType() astStructType         { return nil }
func (anExprNode) FuncType() astFuncType             { return nil }
func (anExprNode) InterfaceType() astInterfaceType   { return nil }
func (anExprNode) MapType() astMapType               { return nil }
func (anExprNode) ChanType() astChanType             { return nil }

type badExprWrapper struct {
	anExprNode
	x *ast.BadExpr
}

func (badExprWrapper) Kind() expressionKind  { return badExprKind }
func (w badExprWrapper) BadExpr() astBadExpr { return w }

func (w badExprWrapper) Pos() astPosition { return w.x.Pos() }
func (w badExprWrapper) End() astPosition { return w.x.End() }
func (w badExprWrapper) Unwrap() ast.Node { return w.x }
func (w badExprWrapper) Reify() astExpr   { return w }

func wrapBadExpr(x *ast.BadExpr) astBadExpr {
	if x == nil {
		return nil
	}
	return badExprWrapper{x: x}
}

type identWrapper struct {
	anExprNode
	x *ast.Ident
}

func (identWrapper) Kind() expressionKind { return identKind }
func (w identWrapper) Ident() astIdent    { return w }
func (w identWrapper) Name() string       { return w.x.Name }

func (w identWrapper) Pos() astPosition { return w.x.Pos() }
func (w identWrapper) End() astPosition { return w.x.End() }
func (w identWrapper) Unwrap() ast.Node { return w.x }
func (w identWrapper) Reify() astExpr   { return w }
func (w identWrapper) String() string   { return w.x.String() }

func astNewIdent(name string, namePos astPosition) astIdent {
	ident := ast.NewIdent(name)
	ident.NamePos = namePos.(token.Pos)
	return wrapIdent(ident)
}

func wrapIdent(ident *ast.Ident) astIdent {
	if ident == nil {
		return nil
	}
	return identWrapper{x: ident}
}

type ellipsisWrapper struct {
	anExprNode
	x *ast.Ellipsis
}

func (ellipsisWrapper) Kind() expressionKind    { return ellipsisKind }
func (w ellipsisWrapper) Ellipsis() astEllipsis { return w }

func (w ellipsisWrapper) Pos() astPosition { return w.x.Pos() }
func (w ellipsisWrapper) End() astPosition { return w.x.End() }
func (w ellipsisWrapper) Unwrap() ast.Node { return w.x }
func (w ellipsisWrapper) Reify() astExpr   { return w }

func wrapEllipsis(x *ast.Ellipsis) astEllipsis {
	if x == nil {
		return nil
	}
	return ellipsisWrapper{x: x}
}

func (w ellipsisWrapper) Elt() astExpr { return wrapExpr(w.x.Elt) }

type basicLitWrapper struct {
	anExprNode
	x *ast.BasicLit
}

func astNewBasicLit(pos astPosition, kind token.Token, value string) astBasicLit {
	lit := &ast.BasicLit{ValuePos: pos.(token.Pos), Kind: kind, Value: value}
	return wrapBasicLit(lit)
}

func (basicLitWrapper) Kind() expressionKind    { return basicLitKind }
func (w basicLitWrapper) BasicLit() astBasicLit { return w }

func (w basicLitWrapper) Pos() astPosition { return w.x.Pos() }
func (w basicLitWrapper) End() astPosition { return w.x.End() }
func (w basicLitWrapper) Unwrap() ast.Node { return w.x }
func (w basicLitWrapper) Reify() astExpr   { return w }

func wrapBasicLit(x *ast.BasicLit) astBasicLit {
	if x == nil {
		return nil
	}
	return basicLitWrapper{x: x}
}

func (w basicLitWrapper) LitKind() token.Token { return w.x.Kind }
func (w basicLitWrapper) Value() string        { return w.x.Value }

type funcLitWrapper struct {
	anExprNode
	x *ast.FuncLit
}

func (funcLitWrapper) Kind() expressionKind  { return funcLitKind }
func (w funcLitWrapper) FuncLit() astFuncLit { return w }
func (w funcLitWrapper) Type() astExpr       { return wrapExpr(w.x.Type) }
func (w funcLitWrapper) Body() astBlockStmt  { return wrapBlockStmt(w.x.Body) }

func (w funcLitWrapper) Pos() astPosition { return w.x.Pos() }
func (w funcLitWrapper) End() astPosition { return w.x.End() }
func (w funcLitWrapper) Unwrap() ast.Node { return w.x }
func (w funcLitWrapper) Reify() astExpr   { return w }

func wrapFuncLit(x *ast.FuncLit) astFuncLit {
	if x == nil {
		return nil
	}
	return funcLitWrapper{x: x}
}

type compositeLitWrapper struct {
	anExprNode
	x *ast.CompositeLit
}

func (compositeLitWrapper) Kind() expressionKind            { return compositeLitKind }
func (w compositeLitWrapper) CompositeLit() astCompositeLit { return w }
func (w compositeLitWrapper) Type() astExpr                 { return wrapExpr(w.x.Type) }
func (w compositeLitWrapper) Elts() astExprList             { return exprListWrapper(w.x.Elts) }
func (w compositeLitWrapper) Rbrace() astPosition           { return w.x.Rbrace }

func (w compositeLitWrapper) Pos() astPosition { return w.x.Pos() }
func (w compositeLitWrapper) End() astPosition { return w.x.End() }
func (w compositeLitWrapper) Unwrap() ast.Node { return w.x }
func (w compositeLitWrapper) Reify() astExpr   { return w }

func wrapCompositeLit(x *ast.CompositeLit) astCompositeLit {
	if x == nil {
		return nil
	}
	return compositeLitWrapper{x: x}
}

type parenExprWrapper struct {
	anExprNode
	x *ast.ParenExpr
}

func (parenExprWrapper) Kind() expressionKind      { return parenExprKind }
func (w parenExprWrapper) ParenExpr() astParenExpr { return w }
func (w parenExprWrapper) X() astExpr              { return wrapExpr(w.x.X) }

func (w parenExprWrapper) Pos() astPosition { return w.x.Pos() }
func (w parenExprWrapper) End() astPosition { return w.x.End() }
func (w parenExprWrapper) Unwrap() ast.Node { return w.x }
func (w parenExprWrapper) Reify() astExpr   { return w }

func wrapParenExpr(x *ast.ParenExpr) astParenExpr {
	if x == nil {
		return nil
	}
	return parenExprWrapper{x: x}
}

type selectorExprWrapper struct {
	anExprNode
	x *ast.SelectorExpr
}

func (selectorExprWrapper) Kind() expressionKind            { return selectorExprKind }
func (w selectorExprWrapper) SelectorExpr() astSelectorExpr { return w }
func (w selectorExprWrapper) X() astExpr                    { return wrapExpr(w.x.X) }
func (w selectorExprWrapper) Sel() astIdent                 { return wrapIdent(w.x.Sel) }

func (w selectorExprWrapper) Pos() astPosition { return w.x.Pos() }
func (w selectorExprWrapper) End() astPosition { return w.x.End() }
func (w selectorExprWrapper) Unwrap() ast.Node { return w.x }
func (w selectorExprWrapper) Reify() astExpr   { return w }

func wrapSelectorExpr(x *ast.SelectorExpr) astSelectorExpr {
	if x == nil {
		return nil
	}
	return selectorExprWrapper{x: x}
}

type indexExprWrapper struct {
	anExprNode
	x *ast.IndexExpr
}

func (indexExprWrapper) Kind() expressionKind      { return indexExprKind }
func (w indexExprWrapper) IndexExpr() astIndexExpr { return w }
func (w indexExprWrapper) X() astExpr              { return wrapExpr(w.x.X) }
func (w indexExprWrapper) Index() astExpr          { return wrapExpr(w.x.Index) }

func (w indexExprWrapper) Pos() astPosition { return w.x.Pos() }
func (w indexExprWrapper) End() astPosition { return w.x.End() }
func (w indexExprWrapper) Unwrap() ast.Node { return w.x }
func (w indexExprWrapper) Reify() astExpr   { return w }

func wrapIndexExpr(x *ast.IndexExpr) astIndexExpr {
	if x == nil {
		return nil
	}
	return indexExprWrapper{x: x}
}

type sliceExprWrapper struct {
	anExprNode
	x *ast.SliceExpr
}

func (sliceExprWrapper) Kind() expressionKind      { return sliceExprKind }
func (w sliceExprWrapper) SliceExpr() astSliceExpr { return w }
func (w sliceExprWrapper) X() astExpr              { return wrapExpr(w.x.X) }
func (w sliceExprWrapper) Low() astExpr            { return wrapExpr(w.x.Low) }
func (w sliceExprWrapper) High() astExpr           { return wrapExpr(w.x.High) }
func (w sliceExprWrapper) Max() astExpr            { return wrapExpr(w.x.Max) }
func (w sliceExprWrapper) Slice3() bool            { return w.x.Slice3 }
func (w sliceExprWrapper) Rbrack() astPosition     { return w.x.Rbrack }

func (w sliceExprWrapper) Pos() astPosition { return w.x.Pos() }
func (w sliceExprWrapper) End() astPosition { return w.x.End() }
func (w sliceExprWrapper) Unwrap() ast.Node { return w.x }
func (w sliceExprWrapper) Reify() astExpr   { return w }

func wrapSliceExpr(x *ast.SliceExpr) astSliceExpr {
	if x == nil {
		return nil
	}
	return sliceExprWrapper{x: x}
}

type typeAssertExprWrapper struct {
	anExprNode
	x *ast.TypeAssertExpr
}

func (typeAssertExprWrapper) Kind() expressionKind                { return typeAssertExprKind }
func (w typeAssertExprWrapper) TypeAssertExpr() astTypeAssertExpr { return w }
func (w typeAssertExprWrapper) X() astExpr                        { return wrapExpr(w.x.X) }
func (w typeAssertExprWrapper) Type() astExpr                     { return wrapExpr(w.x.Type) }

func (w typeAssertExprWrapper) Pos() astPosition { return w.x.Pos() }
func (w typeAssertExprWrapper) End() astPosition { return w.x.End() }
func (w typeAssertExprWrapper) Unwrap() ast.Node { return w.x }
func (w typeAssertExprWrapper) Reify() astExpr   { return w }

func wrapTypeAssertExpr(x *ast.TypeAssertExpr) astTypeAssertExpr {
	if x == nil {
		return nil
	}
	return typeAssertExprWrapper{x: x}
}

type callExprWrapper struct {
	anExprNode
	x *ast.CallExpr
}

func (callExprWrapper) Kind() expressionKind       { return callExprKind }
func (w callExprWrapper) CallExpr() astCallExpr    { return w }
func (w callExprWrapper) Args() astExprList        { return exprListWrapper(w.x.Args) }
func (w callExprWrapper) Fun() astExpr             { return wrapExpr(w.x.Fun) }
func (w callExprWrapper) EllipsisPos() astPosition { return w.x.Ellipsis }
func (w callExprWrapper) Rparen() astPosition      { return w.x.Rparen }

func (w callExprWrapper) Pos() astPosition { return w.x.Pos() }
func (w callExprWrapper) End() astPosition { return w.x.End() }
func (w callExprWrapper) Unwrap() ast.Node { return w.x }
func (w callExprWrapper) Reify() astExpr   { return w }

func wrapCallExpr(x *ast.CallExpr) astCallExpr {
	if x == nil {
		return nil
	}
	return callExprWrapper{x: x}
}

type starExprWrapper struct {
	anExprNode
	x *ast.StarExpr
}

func (starExprWrapper) Kind() expressionKind    { return starExprKind }
func (w starExprWrapper) StarExpr() astStarExpr { return w }
func (w starExprWrapper) X() astExpr            { return wrapExpr(w.x.X) }

func (w starExprWrapper) Pos() astPosition { return w.x.Pos() }
func (w starExprWrapper) End() astPosition { return w.x.End() }
func (w starExprWrapper) Unwrap() ast.Node { return w.x }
func (w starExprWrapper) Reify() astExpr   { return w }

func wrapStarExpr(x *ast.StarExpr) astStarExpr {
	if x == nil {
		return nil
	}
	return starExprWrapper{x: x}
}

type unaryExprWrapper struct {
	anExprNode
	x *ast.UnaryExpr
}

func (unaryExprWrapper) Kind() expressionKind      { return unaryExprKind }
func (w unaryExprWrapper) UnaryExpr() astUnaryExpr { return w }
func (w unaryExprWrapper) X() astExpr              { return wrapExpr(w.x.X) }
func (w unaryExprWrapper) Op() token.Token         { return w.x.Op }

func (w unaryExprWrapper) Pos() astPosition { return w.x.Pos() }
func (w unaryExprWrapper) End() astPosition { return w.x.End() }
func (w unaryExprWrapper) Unwrap() ast.Node { return w.x }
func (w unaryExprWrapper) Reify() astExpr   { return w }

func wrapUnaryExpr(x *ast.UnaryExpr) astUnaryExpr {
	if x == nil {
		return nil
	}
	return unaryExprWrapper{x: x}
}

type binaryExprWrapper struct {
	anExprNode
	x *ast.BinaryExpr
}

func (binaryExprWrapper) Kind() expressionKind        { return binaryExprKind }
func (w binaryExprWrapper) BinaryExpr() astBinaryExpr { return w }
func (w binaryExprWrapper) Op() token.Token           { return w.x.Op }
func (w binaryExprWrapper) X() astExpr                { return wrapExpr(w.x.X) }
func (w binaryExprWrapper) Y() astExpr                { return wrapExpr(w.x.Y) }

func (w binaryExprWrapper) Pos() astPosition { return w.x.Pos() }
func (w binaryExprWrapper) End() astPosition { return w.x.End() }
func (w binaryExprWrapper) Unwrap() ast.Node { return w.x }
func (w binaryExprWrapper) Reify() astExpr   { return w }

func wrapBinaryExpr(x *ast.BinaryExpr) astBinaryExpr {
	if x == nil {
		return nil
	}
	return binaryExprWrapper{x: x}
}

type keyValueExprWrapper struct {
	anExprNode
	x *ast.KeyValueExpr
}

func (keyValueExprWrapper) Kind() expressionKind            { return keyValueExprKind }
func (w keyValueExprWrapper) KeyValueExpr() astKeyValueExpr { return w }
func (w keyValueExprWrapper) Key() astExpr                  { return wrapExpr(w.x.Key) }
func (w keyValueExprWrapper) Colon() astPosition            { return w.x.Colon }
func (w keyValueExprWrapper) Value() astExpr                { return wrapExpr(w.x.Value) }

func (w keyValueExprWrapper) Pos() astPosition { return w.x.Pos() }
func (w keyValueExprWrapper) End() astPosition { return w.x.End() }
func (w keyValueExprWrapper) Unwrap() ast.Node { return w.x }
func (w keyValueExprWrapper) Reify() astExpr   { return w }

func wrapKeyValueExpr(x *ast.KeyValueExpr) astKeyValueExpr {
	if x == nil {
		return nil
	}
	return keyValueExprWrapper{x: x}
}

type arrayTypeWrapper struct {
	anExprNode
	x *ast.ArrayType
}

func (arrayTypeWrapper) Kind() expressionKind      { return arrayTypeKind }
func (w arrayTypeWrapper) ArrayType() astArrayType { return w }
func (w arrayTypeWrapper) Len() astExpr            { return wrapExpr(w.x.Len) }
func (w arrayTypeWrapper) Elt() astExpr            { return wrapExpr(w.x.Elt) }

func (w arrayTypeWrapper) Pos() astPosition { return w.x.Pos() }
func (w arrayTypeWrapper) End() astPosition { return w.x.End() }
func (w arrayTypeWrapper) Unwrap() ast.Node { return w.x }
func (w arrayTypeWrapper) Reify() astExpr   { return w }

func wrapArrayType(x *ast.ArrayType) astArrayType {
	if x == nil {
		return nil
	}
	return arrayTypeWrapper{x: x}
}

type structTypeWrapper struct {
	anExprNode
	x *ast.StructType
}

func (structTypeWrapper) Kind() expressionKind        { return structTypeKind }
func (w structTypeWrapper) StructType() astStructType { return w }
func (w structTypeWrapper) Fields() astFieldList      { return wrapFieldList(w.x.Fields) }

func (w structTypeWrapper) Pos() astPosition { return w.x.Pos() }
func (w structTypeWrapper) End() astPosition { return w.x.End() }
func (w structTypeWrapper) Unwrap() ast.Node { return w.x }
func (w structTypeWrapper) Reify() astExpr   { return w }

func wrapStructType(x *ast.StructType) astStructType {
	if x == nil {
		return nil
	}
	return structTypeWrapper{x: x}
}

type funcTypeWrapper struct {
	anExprNode
	x *ast.FuncType
}

func (funcTypeWrapper) Kind() expressionKind    { return funcTypeKind }
func (w funcTypeWrapper) FuncType() astFuncType { return w }
func (w funcTypeWrapper) Params() astFieldList  { return wrapFieldList(w.x.Params) }
func (w funcTypeWrapper) Results() astFieldList { return wrapFieldList(w.x.Results) }

func (w funcTypeWrapper) Pos() astPosition { return w.x.Pos() }
func (w funcTypeWrapper) End() astPosition { return w.x.End() }
func (w funcTypeWrapper) Unwrap() ast.Node { return w.x }
func (w funcTypeWrapper) Reify() astExpr   { return w }

func wrapFuncType(x *ast.FuncType) astFuncType {
	if x == nil {
		return nil
	}
	return funcTypeWrapper{x: x}
}

type interfaceTypeWrapper struct {
	anExprNode
	x *ast.InterfaceType
}

func (interfaceTypeWrapper) Kind() expressionKind              { return interfaceTypeKind }
func (w interfaceTypeWrapper) InterfaceType() astInterfaceType { return w }
func (w interfaceTypeWrapper) Methods() astFieldList           { return wrapFieldList(w.x.Methods) }

func (w interfaceTypeWrapper) Pos() astPosition { return w.x.Pos() }
func (w interfaceTypeWrapper) End() astPosition { return w.x.End() }
func (w interfaceTypeWrapper) Unwrap() ast.Node { return w.x }
func (w interfaceTypeWrapper) Reify() astExpr   { return w }

func wrapInterfaceType(x *ast.InterfaceType) astInterfaceType {
	if x == nil {
		return nil
	}
	return interfaceTypeWrapper{x: x}
}

type fieldListWrapper struct{ *ast.FieldList }

func (w fieldListWrapper) Len() int             { return len(w.FieldList.List) }
func (w fieldListWrapper) Field(i int) astField { return wrapField(w.List[i]) }
func (w fieldListWrapper) Pos() astPosition     { return w.FieldList.Pos() }
func (w fieldListWrapper) End() astPosition     { return w.FieldList.End() }
func (w fieldListWrapper) Unwrap() ast.Node     { return w.FieldList }

func wrapFieldList(list *ast.FieldList) astFieldList {
	if list == nil {
		return nil
	}
	return fieldListWrapper{list}
}

type fieldWrapper struct{ *ast.Field }

func (w fieldWrapper) Names() astIdentList { return identListWrapper(w.Field.Names) }
func (w fieldWrapper) Pos() astPosition    { return w.Field.Pos() }
func (w fieldWrapper) End() astPosition    { return w.Field.End() }
func (w fieldWrapper) Unwrap() ast.Node    { return w.Field }
func (w fieldWrapper) Type() astExpr       { return wrapExpr(w.Field.Type) }
func (w fieldWrapper) Tag() astBasicLit    { return wrapBasicLit(w.Field.Tag) }

func wrapField(field *ast.Field) astField {
	if field == nil {
		return nil
	}
	return fieldWrapper{field}
}

type identListWrapper []*ast.Ident

func (w identListWrapper) Len() int             { return len(w) }
func (w identListWrapper) Ident(i int) astIdent { return wrapIdent(w[i]) }

type exprListWrapper []ast.Expr

func (w exprListWrapper) Len() int           { return len(w) }
func (w exprListWrapper) Expr(i int) astExpr { return wrapExpr(w[i]) }

type mapTypeWrapper struct {
	anExprNode
	x *ast.MapType
}

func (w mapTypeWrapper) Kind() expressionKind { return mapTypeKind }
func (w mapTypeWrapper) MapType() astMapType  { return w }
func (w mapTypeWrapper) Key() astExpr         { return wrapExpr(w.x.Key) }
func (w mapTypeWrapper) Value() astExpr       { return wrapExpr(w.x.Value) }

func (w mapTypeWrapper) Pos() astPosition { return w.x.Pos() }
func (w mapTypeWrapper) End() astPosition { return w.x.End() }
func (w mapTypeWrapper) Unwrap() ast.Node { return w.x }
func (w mapTypeWrapper) Reify() astExpr   { return w }

func wrapMapType(x *ast.MapType) astMapType {
	if x == nil {
		return nil
	}
	return mapTypeWrapper{x: x}
}

type chanTypeWrapper struct {
	anExprNode
	x *ast.ChanType
}

func (w chanTypeWrapper) Kind() expressionKind  { return chanTypeKind }
func (w chanTypeWrapper) ChanType() astChanType { return w }
func (w chanTypeWrapper) Dir() ast.ChanDir      { return w.x.Dir }
func (w chanTypeWrapper) Value() astExpr        { return wrapExpr(w.x.Value) }

func (w chanTypeWrapper) Pos() astPosition { return w.x.Pos() }
func (w chanTypeWrapper) End() astPosition { return w.x.End() }
func (w chanTypeWrapper) Unwrap() ast.Node { return w.x }
func (w chanTypeWrapper) Reify() astExpr   { return w }

func wrapChanType(x *ast.ChanType) astChanType {
	if x == nil {
		return nil
	}
	return chanTypeWrapper{x: x}
}

type stmtWrapper struct {
	ast.Stmt
}

func (w stmtWrapper) Kind() stmtKind {
	switch w.Stmt.(type) {
	case *ast.BadStmt:
		return badStmtKind
	case *ast.DeclStmt:
		return declStmtKind
	case *ast.EmptyStmt:
		return emptyStmtKind
	case *ast.LabeledStmt:
		return labeledStmtKind
	case *ast.ExprStmt:
		return exprStmtKind
	case *ast.SendStmt:
		return sendStmtKind
	case *ast.IncDecStmt:
		return incDecStmtKind
	case *ast.AssignStmt:
		return assignStmtKind
	case *ast.GoStmt:
		return goStmtKind
	case *ast.DeferStmt:
		return deferStmtKind
	case *ast.ReturnStmt:
		return returnStmtKind
	case *ast.BranchStmt:
		return branchStmtKind
	case *ast.BlockStmt:
		return blockStmtKind
	case *ast.IfStmt:
		return ifStmtKind
	case *ast.CaseClause:
		return caseClauseKind
	case *ast.SwitchStmt:
		return switchStmtKind
	case *ast.TypeSwitchStmt:
		return typeSwitchStmtKind
	case *ast.CommClause:
		return commClauseKind
	case *ast.SelectStmt:
		return selectStmtKind
	case *ast.ForStmt:
		return forStmtKind
	case *ast.RangeStmt:
		return rangeStmtKind
	}
	panic(fmt.Sprintf("stmt %T", w.Stmt))
}

func (w stmtWrapper) Unwrap() ast.Node { return w.Stmt }
func (w stmtWrapper) Pos() astPosition { return w.Stmt.Pos() }
func (w stmtWrapper) End() astPosition { return w.Stmt.End() }

func (w stmtWrapper) BadStmt() astBadStmt {
	s, _ := w.Stmt.(*ast.BadStmt)
	return wrapBadStmt(s)
}
func (w stmtWrapper) DeclStmt() astDeclStmt {
	s, _ := w.Stmt.(*ast.DeclStmt)
	return wrapDeclStmt(s)
}
func (w stmtWrapper) EmptyStmt() astEmptyStmt {
	s, _ := w.Stmt.(*ast.EmptyStmt)
	return wrapEmptyStmt(s)
}
func (w stmtWrapper) LabeledStmt() astLabeledStmt {
	s, _ := w.Stmt.(*ast.LabeledStmt)
	return wrapLabeledStmt(s)
}
func (w stmtWrapper) ExprStmt() astExprStmt {
	s, _ := w.Stmt.(*ast.ExprStmt)
	return wrapExprStmt(s)
}
func (w stmtWrapper) SendStmt() astSendStmt {
	s, _ := w.Stmt.(*ast.SendStmt)
	return wrapSendStmt(s)
}
func (w stmtWrapper) IncDecStmt() astIncDecStmt {
	s, _ := w.Stmt.(*ast.IncDecStmt)
	return wrapIncDecStmt(s)
}
func (w stmtWrapper) AssignStmt() astAssignStmt {
	s, _ := w.Stmt.(*ast.AssignStmt)
	return wrapAssignStmt(s)
}
func (w stmtWrapper) GoStmt() astGoStmt {
	s, _ := w.Stmt.(*ast.GoStmt)
	return wrapGoStmt(s)
}
func (w stmtWrapper) DeferStmt() astDeferStmt {
	s, _ := w.Stmt.(*ast.DeferStmt)
	return wrapDeferStmt(s)
}
func (w stmtWrapper) ReturnStmt() astReturnStmt {
	s, _ := w.Stmt.(*ast.ReturnStmt)
	return wrapReturnStmt(s)
}
func (w stmtWrapper) BranchStmt() astBranchStmt {
	s, _ := w.Stmt.(*ast.BranchStmt)
	return wrapBranchStmt(s)
}
func (w stmtWrapper) BlockStmt() astBlockStmt {
	s, _ := w.Stmt.(*ast.BlockStmt)
	return wrapBlockStmt(s)
}
func (w stmtWrapper) IfStmt() astIfStmt {
	s, _ := w.Stmt.(*ast.IfStmt)
	return wrapIfStmt(s)
}
func (w stmtWrapper) CaseClause() astCaseClause {
	s, _ := w.Stmt.(*ast.CaseClause)
	return wrapCaseClause(s)
}
func (w stmtWrapper) SwitchStmt() astSwitchStmt {
	s, _ := w.Stmt.(*ast.SwitchStmt)
	return wrapSwitchStmt(s)
}
func (w stmtWrapper) TypeSwitchStmt() astTypeSwitchStmt {
	s, _ := w.Stmt.(*ast.TypeSwitchStmt)
	return wrapTypeSwitchStmt(s)
}
func (w stmtWrapper) CommClause() astCommClause {
	s, _ := w.Stmt.(*ast.CommClause)
	return wrapCommClause(s)
}
func (w stmtWrapper) SelectStmt() astSelectStmt {
	s, _ := w.Stmt.(*ast.SelectStmt)
	return wrapSelectStmt(s)
}
func (w stmtWrapper) ForStmt() astForStmt {
	s, _ := w.Stmt.(*ast.ForStmt)
	return wrapForStmt(s)
}
func (w stmtWrapper) RangeStmt() astRangeStmt {
	s, _ := w.Stmt.(*ast.RangeStmt)
	return wrapRangeStmt(s)
}

func wrapStmt(s ast.Stmt) astStmt {
	if s == nil {
		return nil
	}
	return stmtWrapper{s}
}

type aStmtNode struct{}

func (aStmtNode) BadStmt() astBadStmt               { return nil }
func (aStmtNode) DeclStmt() astDeclStmt             { return nil }
func (aStmtNode) EmptyStmt() astEmptyStmt           { return nil }
func (aStmtNode) LabeledStmt() astLabeledStmt       { return nil }
func (aStmtNode) ExprStmt() astExprStmt             { return nil }
func (aStmtNode) SendStmt() astSendStmt             { return nil }
func (aStmtNode) IncDecStmt() astIncDecStmt         { return nil }
func (aStmtNode) AssignStmt() astAssignStmt         { return nil }
func (aStmtNode) GoStmt() astGoStmt                 { return nil }
func (aStmtNode) DeferStmt() astDeferStmt           { return nil }
func (aStmtNode) ReturnStmt() astReturnStmt         { return nil }
func (aStmtNode) BranchStmt() astBranchStmt         { return nil }
func (aStmtNode) BlockStmt() astBlockStmt           { return nil }
func (aStmtNode) IfStmt() astIfStmt                 { return nil }
func (aStmtNode) CaseClause() astCaseClause         { return nil }
func (aStmtNode) SwitchStmt() astSwitchStmt         { return nil }
func (aStmtNode) TypeSwitchStmt() astTypeSwitchStmt { return nil }
func (aStmtNode) CommClause() astCommClause         { return nil }
func (aStmtNode) SelectStmt() astSelectStmt         { return nil }
func (aStmtNode) ForStmt() astForStmt               { return nil }
func (aStmtNode) RangeStmt() astRangeStmt           { return nil }

type badStmtWrapper struct {
	aStmtNode
	s *ast.BadStmt
}

func (badStmtWrapper) Kind() stmtKind        { return badStmtKind }
func (s badStmtWrapper) BadStmt() astBadStmt { return s }

func (s badStmtWrapper) Pos() astPosition { return s.s.Pos() }
func (s badStmtWrapper) End() astPosition { return s.s.End() }
func (s badStmtWrapper) Unwrap() ast.Node { return s.s }

func wrapBadStmt(s *ast.BadStmt) astBadStmt {
	if s == nil {
		return nil
	}
	return badStmtWrapper{s: s}
}

type declStmtWrapper struct {
	aStmtNode
	s *ast.DeclStmt
}

func (declStmtWrapper) Kind() stmtKind          { return declStmtKind }
func (s declStmtWrapper) DeclStmt() astDeclStmt { return s }
func (s declStmtWrapper) Decl() astDecl         { return wrapDecl(s.s.Decl) }

func (s declStmtWrapper) Pos() astPosition { return s.s.Pos() }
func (s declStmtWrapper) End() astPosition { return s.s.End() }
func (s declStmtWrapper) Unwrap() ast.Node { return s.s }

func wrapDeclStmt(s *ast.DeclStmt) astDeclStmt {
	if s == nil {
		return nil
	}
	return declStmtWrapper{s: s}
}

type emptyStmtWrapper struct {
	aStmtNode
	s *ast.EmptyStmt
}

func (emptyStmtWrapper) Kind() stmtKind            { return emptyStmtKind }
func (s emptyStmtWrapper) EmptyStmt() astEmptyStmt { return s }

func (s emptyStmtWrapper) Pos() astPosition { return s.s.Pos() }
func (s emptyStmtWrapper) End() astPosition { return s.s.End() }
func (s emptyStmtWrapper) Unwrap() ast.Node { return s.s }

func wrapEmptyStmt(s *ast.EmptyStmt) astEmptyStmt {
	if s == nil {
		return nil
	}
	return emptyStmtWrapper{s: s}
}

type labeledStmtWrapper struct {
	aStmtNode
	s *ast.LabeledStmt
}

func (labeledStmtWrapper) Kind() stmtKind                { return labeledStmtKind }
func (s labeledStmtWrapper) LabeledStmt() astLabeledStmt { return s }
func (s labeledStmtWrapper) Label() astIdent             { return wrapIdent(s.s.Label) }
func (s labeledStmtWrapper) Stmt() astStmt               { return wrapStmt(s.s.Stmt) }

func (s labeledStmtWrapper) Pos() astPosition { return s.s.Pos() }
func (s labeledStmtWrapper) End() astPosition { return s.s.End() }
func (s labeledStmtWrapper) Unwrap() ast.Node { return s.s }

func wrapLabeledStmt(s *ast.LabeledStmt) astLabeledStmt {
	if s == nil {
		return nil
	}
	return labeledStmtWrapper{s: s}
}

type exprStmtWrapper struct {
	aStmtNode
	s *ast.ExprStmt
}

func (exprStmtWrapper) Kind() stmtKind          { return exprStmtKind }
func (s exprStmtWrapper) ExprStmt() astExprStmt { return s }
func (s exprStmtWrapper) X() astExpr            { return wrapExpr(s.s.X) }

func (s exprStmtWrapper) Pos() astPosition { return s.s.Pos() }
func (s exprStmtWrapper) End() astPosition { return s.s.End() }
func (s exprStmtWrapper) Unwrap() ast.Node { return s.s }

func wrapExprStmt(s *ast.ExprStmt) astExprStmt {
	if s == nil {
		return nil
	}
	return exprStmtWrapper{s: s}
}

type sendStmtWrapper struct {
	aStmtNode
	s *ast.SendStmt
}

func (sendStmtWrapper) Kind() stmtKind          { return sendStmtKind }
func (s sendStmtWrapper) SendStmt() astSendStmt { return s }
func (s sendStmtWrapper) Chan() astExpr         { return wrapExpr(s.s.Chan) }
func (s sendStmtWrapper) Value() astExpr        { return wrapExpr(s.s.Value) }
func (s sendStmtWrapper) Arrow() astPosition    { return s.s.Arrow }

func (s sendStmtWrapper) Pos() astPosition { return s.s.Pos() }
func (s sendStmtWrapper) End() astPosition { return s.s.End() }
func (s sendStmtWrapper) Unwrap() ast.Node { return s.s }

func wrapSendStmt(s *ast.SendStmt) astSendStmt {
	if s == nil {
		return nil
	}
	return sendStmtWrapper{s: s}
}

type incDecStmtWrapper struct {
	aStmtNode
	s *ast.IncDecStmt
}

func (incDecStmtWrapper) Kind() stmtKind              { return incDecStmtKind }
func (s incDecStmtWrapper) IncDecStmt() astIncDecStmt { return s }
func (s incDecStmtWrapper) Tok() token.Token          { return s.s.Tok }
func (s incDecStmtWrapper) TokPos() astPosition       { return s.s.TokPos }
func (s incDecStmtWrapper) X() astExpr                { return wrapExpr(s.s.X) }

func (s incDecStmtWrapper) Pos() astPosition { return s.s.Pos() }
func (s incDecStmtWrapper) End() astPosition { return s.s.End() }
func (s incDecStmtWrapper) Unwrap() ast.Node { return s.s }

func wrapIncDecStmt(s *ast.IncDecStmt) astIncDecStmt {
	if s == nil {
		return nil
	}
	return incDecStmtWrapper{s: s}
}

type assignStmtWrapper struct {
	aStmtNode
	s *ast.AssignStmt
}

func (assignStmtWrapper) Kind() stmtKind              { return assignStmtKind }
func (s assignStmtWrapper) AssignStmt() astAssignStmt { return s }
func (s assignStmtWrapper) Lhs() astExprList          { return exprListWrapper(s.s.Lhs) }
func (s assignStmtWrapper) Rhs() astExprList          { return exprListWrapper(s.s.Rhs) }
func (s assignStmtWrapper) Tok() token.Token          { return s.s.Tok }
func (s assignStmtWrapper) TokPos() astPosition       { return s.s.TokPos }

func (s assignStmtWrapper) Pos() astPosition { return s.s.Pos() }
func (s assignStmtWrapper) End() astPosition { return s.s.End() }
func (s assignStmtWrapper) Unwrap() ast.Node { return s.s }

func wrapAssignStmt(s *ast.AssignStmt) astAssignStmt {
	if s == nil {
		return nil
	}
	return assignStmtWrapper{s: s}
}

type goStmtWrapper struct {
	aStmtNode
	s *ast.GoStmt
}

func (goStmtWrapper) Kind() stmtKind      { return goStmtKind }
func (s goStmtWrapper) GoStmt() astGoStmt { return s }
func (s goStmtWrapper) Call() astCallExpr { return wrapCallExpr(s.s.Call) }

func (s goStmtWrapper) Pos() astPosition { return s.s.Pos() }
func (s goStmtWrapper) End() astPosition { return s.s.End() }
func (s goStmtWrapper) Unwrap() ast.Node { return s.s }

func wrapGoStmt(s *ast.GoStmt) astGoStmt {
	if s == nil {
		return nil
	}
	return goStmtWrapper{s: s}
}

type deferStmtWrapper struct {
	aStmtNode
	s *ast.DeferStmt
}

func (deferStmtWrapper) Kind() stmtKind            { return deferStmtKind }
func (s deferStmtWrapper) DeferStmt() astDeferStmt { return s }
func (s deferStmtWrapper) Call() astCallExpr       { return wrapCallExpr(s.s.Call) }

func (s deferStmtWrapper) Pos() astPosition { return s.s.Pos() }
func (s deferStmtWrapper) End() astPosition { return s.s.End() }
func (s deferStmtWrapper) Unwrap() ast.Node { return s.s }

func wrapDeferStmt(s *ast.DeferStmt) astDeferStmt {
	if s == nil {
		return nil
	}
	return deferStmtWrapper{s: s}
}

type returnStmtWrapper struct {
	aStmtNode
	s *ast.ReturnStmt
}

func (returnStmtWrapper) Kind() stmtKind              { return returnStmtKind }
func (s returnStmtWrapper) ReturnStmt() astReturnStmt { return s }
func (s returnStmtWrapper) Results() astExprList      { return exprListWrapper(s.s.Results) }
func (s returnStmtWrapper) Return() astPosition       { return s.s.Return }

func (s returnStmtWrapper) Pos() astPosition { return s.s.Pos() }
func (s returnStmtWrapper) End() astPosition { return s.s.End() }
func (s returnStmtWrapper) Unwrap() ast.Node { return s.s }

func wrapReturnStmt(s *ast.ReturnStmt) astReturnStmt {
	if s == nil {
		return nil
	}
	return returnStmtWrapper{s: s}
}

type branchStmtWrapper struct {
	aStmtNode
	s *ast.BranchStmt
}

func (branchStmtWrapper) Kind() stmtKind              { return branchStmtKind }
func (s branchStmtWrapper) BranchStmt() astBranchStmt { return s }
func (s branchStmtWrapper) Tok() token.Token          { return s.s.Tok }
func (s branchStmtWrapper) Label() astIdent           { return wrapIdent(s.s.Label) }

func (s branchStmtWrapper) Pos() astPosition { return s.s.Pos() }
func (s branchStmtWrapper) End() astPosition { return s.s.End() }
func (s branchStmtWrapper) Unwrap() ast.Node { return s.s }

func wrapBranchStmt(s *ast.BranchStmt) astBranchStmt {
	if s == nil {
		return nil
	}
	return branchStmtWrapper{s: s}
}

type blockStmtWrapper struct {
	aStmtNode
	s *ast.BlockStmt
}

func (blockStmtWrapper) Kind() stmtKind            { return blockStmtKind }
func (s blockStmtWrapper) BlockStmt() astBlockStmt { return s }
func (s blockStmtWrapper) List() astStmtList       { return stmtListWrapper(s.s.List) }
func (s blockStmtWrapper) Lbrace() astPosition     { return s.s.Lbrace }
func (s blockStmtWrapper) Rbrace() astPosition     { return s.s.Rbrace }

func (s blockStmtWrapper) Pos() astPosition { return s.s.Pos() }
func (s blockStmtWrapper) End() astPosition { return s.s.End() }
func (s blockStmtWrapper) Unwrap() ast.Node { return s.s }

func wrapBlockStmt(s *ast.BlockStmt) astBlockStmt {
	if s == nil {
		return nil
	}
	return blockStmtWrapper{s: s}
}

type ifStmtWrapper struct {
	aStmtNode
	s *ast.IfStmt
}

func (ifStmtWrapper) Kind() stmtKind       { return ifStmtKind }
func (s ifStmtWrapper) IfStmt() astIfStmt  { return s }
func (s ifStmtWrapper) Init() astStmt      { return wrapStmt(s.s.Init) }
func (s ifStmtWrapper) Cond() astExpr      { return wrapExpr(s.s.Cond) }
func (s ifStmtWrapper) Body() astBlockStmt { return wrapBlockStmt(s.s.Body) }
func (s ifStmtWrapper) Else() astStmt      { return wrapStmt(s.s.Else) }

func (s ifStmtWrapper) Pos() astPosition { return s.s.Pos() }
func (s ifStmtWrapper) End() astPosition { return s.s.End() }
func (s ifStmtWrapper) Unwrap() ast.Node { return s.s }

func wrapIfStmt(s *ast.IfStmt) astIfStmt {
	if s == nil {
		return nil
	}
	return ifStmtWrapper{s: s}
}

type caseClauseWrapper struct {
	aStmtNode
	s *ast.CaseClause
}

func (caseClauseWrapper) Kind() stmtKind              { return caseClauseKind }
func (s caseClauseWrapper) CaseClause() astCaseClause { return s }
func (s caseClauseWrapper) List() astExprList         { return exprListWrapper(s.s.List) }
func (s caseClauseWrapper) Body() astStmtList         { return stmtListWrapper(s.s.Body) }

func (s caseClauseWrapper) Pos() astPosition { return s.s.Pos() }
func (s caseClauseWrapper) End() astPosition { return s.s.End() }
func (s caseClauseWrapper) Unwrap() ast.Node { return s.s }

func wrapCaseClause(s *ast.CaseClause) astCaseClause {
	if s == nil {
		return nil
	}
	return caseClauseWrapper{s: s}
}

type switchStmtWrapper struct {
	aStmtNode
	s *ast.SwitchStmt
}

func (switchStmtWrapper) Kind() stmtKind              { return switchStmtKind }
func (s switchStmtWrapper) SwitchStmt() astSwitchStmt { return s }
func (s switchStmtWrapper) Init() astStmt             { return wrapStmt(s.s.Init) }
func (s switchStmtWrapper) Tag() astExpr              { return wrapExpr(s.s.Tag) }
func (s switchStmtWrapper) Body() astBlockStmt        { return wrapBlockStmt(s.s.Body) }

func (s switchStmtWrapper) Pos() astPosition { return s.s.Pos() }
func (s switchStmtWrapper) End() astPosition { return s.s.End() }
func (s switchStmtWrapper) Unwrap() ast.Node { return s.s }

func wrapSwitchStmt(s *ast.SwitchStmt) astSwitchStmt {
	if s == nil {
		return nil
	}
	return switchStmtWrapper{s: s}
}

type typeSwitchStmtWrapper struct {
	aStmtNode
	s *ast.TypeSwitchStmt
}

func (typeSwitchStmtWrapper) Kind() stmtKind                      { return typeSwitchStmtKind }
func (s typeSwitchStmtWrapper) TypeSwitchStmt() astTypeSwitchStmt { return s }
func (s typeSwitchStmtWrapper) Init() astStmt                     { return wrapStmt(s.s.Init) }
func (s typeSwitchStmtWrapper) Assign() astStmt                   { return wrapStmt(s.s.Assign) }
func (s typeSwitchStmtWrapper) Body() astBlockStmt                { return wrapBlockStmt(s.s.Body) }

func (s typeSwitchStmtWrapper) Pos() astPosition { return s.s.Pos() }
func (s typeSwitchStmtWrapper) End() astPosition { return s.s.End() }
func (s typeSwitchStmtWrapper) Unwrap() ast.Node { return s.s }

func wrapTypeSwitchStmt(s *ast.TypeSwitchStmt) astTypeSwitchStmt {
	if s == nil {
		return nil
	}
	return typeSwitchStmtWrapper{s: s}
}

type commClauseWrapper struct {
	aStmtNode
	s *ast.CommClause
}

func (commClauseWrapper) Kind() stmtKind              { return commClauseKind }
func (s commClauseWrapper) CommClause() astCommClause { return s }
func (s commClauseWrapper) Comm() astStmt             { return wrapStmt(s.s.Comm) }
func (s commClauseWrapper) Body() astStmtList         { return stmtListWrapper(s.s.Body) }

func (s commClauseWrapper) Pos() astPosition { return s.s.Pos() }
func (s commClauseWrapper) End() astPosition { return s.s.End() }
func (s commClauseWrapper) Unwrap() ast.Node { return s.s }

func wrapCommClause(s *ast.CommClause) astCommClause {
	if s == nil {
		return nil
	}
	return commClauseWrapper{s: s}
}

type selectStmtWrapper struct {
	aStmtNode
	s *ast.SelectStmt
}

func (selectStmtWrapper) Kind() stmtKind              { return selectStmtKind }
func (s selectStmtWrapper) SelectStmt() astSelectStmt { return s }
func (s selectStmtWrapper) Body() astBlockStmt        { return wrapBlockStmt(s.s.Body) }

func (s selectStmtWrapper) Pos() astPosition { return s.s.Pos() }
func (s selectStmtWrapper) End() astPosition { return s.s.End() }
func (s selectStmtWrapper) Unwrap() ast.Node { return s.s }

func wrapSelectStmt(s *ast.SelectStmt) astSelectStmt {
	if s == nil {
		return nil
	}
	return selectStmtWrapper{s: s}
}

type forStmtWrapper struct {
	aStmtNode
	s *ast.ForStmt
}

func (forStmtWrapper) Kind() stmtKind        { return forStmtKind }
func (s forStmtWrapper) ForStmt() astForStmt { return s }
func (s forStmtWrapper) Init() astStmt       { return wrapStmt(s.s.Init) }
func (s forStmtWrapper) Cond() astExpr       { return wrapExpr(s.s.Cond) }
func (s forStmtWrapper) Post() astStmt       { return wrapStmt(s.s.Post) }
func (s forStmtWrapper) Body() astBlockStmt  { return wrapBlockStmt(s.s.Body) }

func (s forStmtWrapper) Pos() astPosition { return s.s.Pos() }
func (s forStmtWrapper) End() astPosition { return s.s.End() }
func (s forStmtWrapper) Unwrap() ast.Node { return s.s }

func wrapForStmt(s *ast.ForStmt) astForStmt {
	if s == nil {
		return nil
	}
	return forStmtWrapper{s: s}
}

type rangeStmtWrapper struct {
	aStmtNode
	s *ast.RangeStmt
}

func (rangeStmtWrapper) Kind() stmtKind            { return rangeStmtKind }
func (s rangeStmtWrapper) RangeStmt() astRangeStmt { return s }
func (s rangeStmtWrapper) Key() astExpr            { return wrapExpr(s.s.Key) }
func (s rangeStmtWrapper) Value() astExpr          { return wrapExpr(s.s.Value) }
func (s rangeStmtWrapper) X() astExpr              { return wrapExpr(s.s.X) }
func (s rangeStmtWrapper) Body() astBlockStmt      { return wrapBlockStmt(s.s.Body) }
func (s rangeStmtWrapper) Tok() token.Token        { return s.s.Tok }
func (s rangeStmtWrapper) TokPos() astPosition     { return s.s.TokPos }

func (s rangeStmtWrapper) Pos() astPosition { return s.s.Pos() }
func (s rangeStmtWrapper) End() astPosition { return s.s.End() }
func (s rangeStmtWrapper) Unwrap() ast.Node { return s.s }

func wrapRangeStmt(s *ast.RangeStmt) astRangeStmt {
	if s == nil {
		return nil
	}
	return rangeStmtWrapper{s: s}
}

type stmtListWrapper []ast.Stmt

func (w stmtListWrapper) Len() int               { return len(w) }
func (w stmtListWrapper) Stmt(i int) astStmt     { return wrapStmt(w[i]) }
func (w stmtListWrapper) Head(i int) astStmtList { return stmtListWrapper(w[:i]) }

// Helpers for switching on Kind(), to simulate the nilness handling of
// type switches.

func kindOfExpr(x astExpr) expressionKind {
	if x == nil {
		return nilExprKind
	}
	return x.Kind()
}

func kindOfStmt(s astStmt) stmtKind {
	if s == nil {
		return nilStmtKind
	}
	return s.Kind()
}

func kindOfDecl(d astDecl) declKind {
	if d == nil {
		return nilDeclKind
	}
	return d.Kind()
}

// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"fmt"
	"go/ast"
	"go/token"

	itypes "internal/types"
)

type unwrapper interface {
	Unwrap() ast.Node
}

type fileWrapper struct{ *ast.File }

func (w fileWrapper) Pos() itypes.Pos        { return w.File.Pos() }
func (w fileWrapper) End() itypes.Pos        { return w.File.End() }
func (w fileWrapper) Unwrap() ast.Node       { return w.File }
func (w fileWrapper) DeclsLen() int          { return len(w.File.Decls) }
func (w fileWrapper) Decl(i int) itypes.Decl { return wrapDecl(w.File.Decls[i]) }
func (w fileWrapper) Name() itypes.Ident     { return wrapIdent(w.File.Name) }
func (w fileWrapper) Package() itypes.Pos    { return w.File.Package }

func wrapDecl(d ast.Decl) itypes.Decl {
	switch d := d.(type) {
	case nil:
		return nil
	case *ast.BadDecl:
		return badDeclWrapper{d}
	case *ast.GenDecl:
		return genDeclWrapper{d}
	case *ast.FuncDecl:
		return funcDeclWrapper{d}
	}
	panic(fmt.Sprintf("decl %T", d))
}

type badDeclWrapper struct{ d *ast.BadDecl }

func (w badDeclWrapper) ADecl()           {}
func (w badDeclWrapper) ABadDecl()        {}
func (w badDeclWrapper) Unwrap() ast.Node { return w.d }
func (w badDeclWrapper) Pos() itypes.Pos  { return w.d.Pos() }
func (w badDeclWrapper) End() itypes.Pos  { return w.d.End() }

type genDeclWrapper struct{ d *ast.GenDecl }

func (w genDeclWrapper) ADecl()                 {}
func (w genDeclWrapper) AGenDecl()              {}
func (w genDeclWrapper) SpecsLen() int          { return len(w.d.Specs) }
func (w genDeclWrapper) Spec(i int) itypes.Spec { return wrapSpec(w.d.Specs[i]) }
func (w genDeclWrapper) Tok() token.Token       { return w.d.Tok }
func (w genDeclWrapper) Unwrap() ast.Node       { return w.d }
func (w genDeclWrapper) Pos() itypes.Pos        { return w.d.Pos() }
func (w genDeclWrapper) End() itypes.Pos        { return w.d.End() }

type funcDeclWrapper struct{ d *ast.FuncDecl }

func (w funcDeclWrapper) ADecl()                 {}
func (w funcDeclWrapper) AFuncDecl()             {}
func (w funcDeclWrapper) Recv() itypes.FieldList { return wrapFieldList(w.d.Recv) }
func (w funcDeclWrapper) Name() itypes.Ident     { return wrapIdent(w.d.Name) }
func (w funcDeclWrapper) Type() itypes.FuncType  { return wrapFuncType(w.d.Type) }
func (w funcDeclWrapper) Body() itypes.BlockStmt { return wrapBlockStmt(w.d.Body) }
func (w funcDeclWrapper) Unwrap() ast.Node       { return w.d }
func (w funcDeclWrapper) Pos() itypes.Pos        { return w.d.Pos() }
func (w funcDeclWrapper) End() itypes.Pos        { return w.d.End() }

func wrapSpec(spec ast.Spec) itypes.Spec {
	switch s := spec.(type) {
	case nil:
		return nil
	case (*ast.ImportSpec):
		return importSpecWrapper{s}
	case (*ast.ValueSpec):
		return valueSpecWrapper{s}
	case (*ast.TypeSpec):
		return typeSpecWrapper{s}
	}
	panic(fmt.Sprintf("spec %T", spec))
}

type importSpecWrapper struct{ s *ast.ImportSpec }

func (w importSpecWrapper) ASpec()                {}
func (w importSpecWrapper) AnImportSpec()         {}
func (w importSpecWrapper) Path() itypes.BasicLit { return wrapBasicLit(w.s.Path) }
func (w importSpecWrapper) Name() itypes.Ident    { return wrapIdent(w.s.Name) }
func (w importSpecWrapper) Pos() itypes.Pos       { return w.s.Pos() }
func (w importSpecWrapper) End() itypes.Pos       { return w.s.End() }
func (w importSpecWrapper) Unwrap() ast.Node      { return w.s }

type valueSpecWrapper struct{ s *ast.ValueSpec }

func (valueSpecWrapper) ASpec()                    {}
func (valueSpecWrapper) AValueSpec()               {}
func (w valueSpecWrapper) NamesLen() int           { return len(w.s.Names) }
func (w valueSpecWrapper) Name(i int) itypes.Ident { return wrapIdent(w.s.Names[i]) }
func (w valueSpecWrapper) Type() itypes.Expr       { return wrapExpr(w.s.Type) }
func (w valueSpecWrapper) ValuesLen() int          { return len(w.s.Values) }
func (w valueSpecWrapper) Value(i int) itypes.Expr { return wrapExpr(w.s.Values[i]) }
func (w valueSpecWrapper) Pos() itypes.Pos         { return w.s.Pos() }
func (w valueSpecWrapper) End() itypes.Pos         { return w.s.End() }
func (w valueSpecWrapper) Unwrap() ast.Node        { return w.s }

func wrapValueSpec(s *ast.ValueSpec) itypes.ValueSpec {
	if s == nil {
		return nil
	}
	return valueSpecWrapper{s}
}

func astNewValueSpec() itypes.ValueSpec {
	return wrapValueSpec(new(ast.ValueSpec))
}

type typeSpecWrapper struct{ s *ast.TypeSpec }

func (w typeSpecWrapper) ASpec()             {}
func (w typeSpecWrapper) ATypeSpec()         {}
func (w typeSpecWrapper) Name() itypes.Ident { return wrapIdent(w.s.Name) }
func (w typeSpecWrapper) Assign() itypes.Pos { return w.s.Assign }
func (w typeSpecWrapper) Type() itypes.Expr  { return wrapExpr(w.s.Type) }
func (w typeSpecWrapper) Pos() itypes.Pos    { return w.s.Pos() }
func (w typeSpecWrapper) End() itypes.Pos    { return w.s.End() }
func (w typeSpecWrapper) Unwrap() ast.Node   { return w.s }

func wrapExpr(expr ast.Expr) itypes.Expr {
	switch x := expr.(type) {
	case nil:
		return nil
	case *ast.BadExpr:
		return badExprWrapper{x: x}
	case *ast.Ident:
		return identWrapper{x}
	case *ast.Ellipsis:
		return ellipsisWrapper{x: x}
	case *ast.BasicLit:
		return basicLitWrapper{x}
	case *ast.FuncLit:
		return funcLitWrapper{x: x}
	case *ast.CompositeLit:
		return compositeLitWrapper{x: x}
	case *ast.ParenExpr:
		return parenExprWrapper{x: x}
	case *ast.SelectorExpr:
		return selectorExprWrapper{x: x}
	case *ast.IndexExpr:
		return indexExprWrapper{x: x}
	case *ast.SliceExpr:
		return sliceExprWrapper{x: x}
	case *ast.TypeAssertExpr:
		return typeAssertExprWrapper{x: x}
	case *ast.CallExpr:
		return callExprWrapper{x}
	case *ast.StarExpr:
		return starExprWrapper{x}
	case *ast.UnaryExpr:
		return unaryExprWrapper{x}
	case *ast.BinaryExpr:
		return binaryExprWrapper{x}
	case *ast.KeyValueExpr:
		return keyValueExprWrapper{x}
	case *ast.ArrayType:
		return arrayTypeWrapper{x: x}
	case *ast.StructType:
		return structTypeWrapper{x: x}
	case *ast.FuncType:
		return funcTypeWrapper{x: x}
	case *ast.InterfaceType:
		return interfaceTypeWrapper{x: x}
	case *ast.MapType:
		return mapTypeWrapper{x: x}
	case *ast.ChanType:
		return chanTypeWrapper{x: x}
	}
	panic(fmt.Sprintf("expr %T", expr))
}

type badExprWrapper struct{ x *ast.BadExpr }

func (badExprWrapper) AnExpr()            {}
func (badExprWrapper) ABadExpr()          {}
func (w badExprWrapper) Pos() itypes.Pos  { return w.x.Pos() }
func (w badExprWrapper) End() itypes.Pos  { return w.x.End() }
func (w badExprWrapper) Unwrap() ast.Node { return w.x }

type identWrapper struct{ x *ast.Ident }

func (identWrapper) AnExpr()            {}
func (identWrapper) AnIdent()           {}
func (w identWrapper) Name() string     { return w.x.Name }
func (w identWrapper) Pos() itypes.Pos  { return w.x.Pos() }
func (w identWrapper) End() itypes.Pos  { return w.x.End() }
func (w identWrapper) Unwrap() ast.Node { return w.x }
func (w identWrapper) String() string   { return w.x.String() }

func astNewIdent(name string, namePos itypes.Pos) itypes.Ident {
	ident := ast.NewIdent(name)
	ident.NamePos = namePos.(token.Pos)
	return wrapIdent(ident)
}

func wrapIdent(ident *ast.Ident) itypes.Ident {
	if ident == nil {
		return nil
	}
	return identWrapper{ident}
}

type ellipsisWrapper struct{ x *ast.Ellipsis }

func (ellipsisWrapper) AnExpr()            {}
func (ellipsisWrapper) AnEllipsis()        {}
func (w ellipsisWrapper) Pos() itypes.Pos  { return w.x.Pos() }
func (w ellipsisWrapper) End() itypes.Pos  { return w.x.End() }
func (w ellipsisWrapper) Unwrap() ast.Node { return w.x }
func (w ellipsisWrapper) Elt() itypes.Expr { return wrapExpr(w.x.Elt) }

type basicLitWrapper struct{ x *ast.BasicLit }

func astNewBasicLit(pos itypes.Pos, kind token.Token, value string) itypes.BasicLit {
	lit := &ast.BasicLit{ValuePos: pos.(token.Pos), Kind: kind, Value: value}
	return wrapBasicLit(lit)
}

func (basicLitWrapper) AnExpr()             {}
func (basicLitWrapper) ABasicLit()          {}
func (w basicLitWrapper) Pos() itypes.Pos   { return w.x.Pos() }
func (w basicLitWrapper) End() itypes.Pos   { return w.x.End() }
func (w basicLitWrapper) Unwrap() ast.Node  { return w.x }
func (w basicLitWrapper) Kind() token.Token { return w.x.Kind }
func (w basicLitWrapper) Value() string     { return w.x.Value }

func wrapBasicLit(x *ast.BasicLit) itypes.BasicLit {
	if x == nil {
		return nil
	}
	return basicLitWrapper{x}
}

type funcLitWrapper struct{ x *ast.FuncLit }

func (funcLitWrapper) AnExpr()                  {}
func (funcLitWrapper) AFuncLit()                {}
func (w funcLitWrapper) Type() itypes.Expr      { return wrapExpr(w.x.Type) }
func (w funcLitWrapper) Body() itypes.BlockStmt { return wrapBlockStmt(w.x.Body) }
func (w funcLitWrapper) Pos() itypes.Pos        { return w.x.Pos() }
func (w funcLitWrapper) End() itypes.Pos        { return w.x.End() }
func (w funcLitWrapper) Unwrap() ast.Node       { return w.x }

type compositeLitWrapper struct{ x *ast.CompositeLit }

func (compositeLitWrapper) AnExpr()                 {}
func (compositeLitWrapper) ACompositeLit()          {}
func (w compositeLitWrapper) Type() itypes.Expr     { return wrapExpr(w.x.Type) }
func (w compositeLitWrapper) EltsLen() int          { return len(w.x.Elts) }
func (w compositeLitWrapper) Elt(i int) itypes.Expr { return wrapExpr(w.x.Elts[i]) }
func (w compositeLitWrapper) Rbrace() itypes.Pos    { return w.x.Rbrace }
func (w compositeLitWrapper) Pos() itypes.Pos       { return w.x.Pos() }
func (w compositeLitWrapper) End() itypes.Pos       { return w.x.End() }
func (w compositeLitWrapper) Unwrap() ast.Node      { return w.x }

type parenExprWrapper struct{ x *ast.ParenExpr }

func (parenExprWrapper) AnExpr()            {}
func (parenExprWrapper) AParenExpr()        {}
func (w parenExprWrapper) X() itypes.Expr   { return wrapExpr(w.x.X) }
func (w parenExprWrapper) Pos() itypes.Pos  { return w.x.Pos() }
func (w parenExprWrapper) End() itypes.Pos  { return w.x.End() }
func (w parenExprWrapper) Unwrap() ast.Node { return w.x }

type selectorExprWrapper struct{ x *ast.SelectorExpr }

func (selectorExprWrapper) AnExpr()             {}
func (selectorExprWrapper) ASelectorExpr()      {}
func (w selectorExprWrapper) X() itypes.Expr    { return wrapExpr(w.x.X) }
func (w selectorExprWrapper) Sel() itypes.Ident { return wrapIdent(w.x.Sel) }
func (w selectorExprWrapper) Pos() itypes.Pos   { return w.x.Pos() }
func (w selectorExprWrapper) End() itypes.Pos   { return w.x.End() }
func (w selectorExprWrapper) Unwrap() ast.Node  { return w.x }

type indexExprWrapper struct{ x *ast.IndexExpr }

func (indexExprWrapper) AnExpr()              {}
func (indexExprWrapper) AnIndexExpr()         {}
func (w indexExprWrapper) X() itypes.Expr     { return wrapExpr(w.x.X) }
func (w indexExprWrapper) Index() itypes.Expr { return wrapExpr(w.x.Index) }
func (w indexExprWrapper) Pos() itypes.Pos    { return w.x.Pos() }
func (w indexExprWrapper) End() itypes.Pos    { return w.x.End() }
func (w indexExprWrapper) Unwrap() ast.Node   { return w.x }

type sliceExprWrapper struct{ x *ast.SliceExpr }

func (sliceExprWrapper) AnExpr()              {}
func (sliceExprWrapper) ASliceExpr()          {}
func (w sliceExprWrapper) X() itypes.Expr     { return wrapExpr(w.x.X) }
func (w sliceExprWrapper) Low() itypes.Expr   { return wrapExpr(w.x.Low) }
func (w sliceExprWrapper) High() itypes.Expr  { return wrapExpr(w.x.High) }
func (w sliceExprWrapper) Max() itypes.Expr   { return wrapExpr(w.x.Max) }
func (w sliceExprWrapper) Slice3() bool       { return w.x.Slice3 }
func (w sliceExprWrapper) Rbrack() itypes.Pos { return w.x.Rbrack }
func (w sliceExprWrapper) Pos() itypes.Pos    { return w.x.Pos() }
func (w sliceExprWrapper) End() itypes.Pos    { return w.x.End() }
func (w sliceExprWrapper) Unwrap() ast.Node   { return w.x }

type typeAssertExprWrapper struct{ x *ast.TypeAssertExpr }

func (typeAssertExprWrapper) AnExpr()             {}
func (typeAssertExprWrapper) ATypeAssertExpr()    {}
func (w typeAssertExprWrapper) X() itypes.Expr    { return wrapExpr(w.x.X) }
func (w typeAssertExprWrapper) Type() itypes.Expr { return wrapExpr(w.x.Type) }
func (w typeAssertExprWrapper) Pos() itypes.Pos   { return w.x.Pos() }
func (w typeAssertExprWrapper) End() itypes.Pos   { return w.x.End() }
func (w typeAssertExprWrapper) Unwrap() ast.Node  { return w.x }

type callExprWrapper struct{ x *ast.CallExpr }

func (callExprWrapper) AnExpr()                 {}
func (callExprWrapper) ACallExpr()              {}
func (w callExprWrapper) ArgsLen() int          { return len(w.x.Args) }
func (w callExprWrapper) Arg(i int) itypes.Expr { return wrapExpr(w.x.Args[i]) }
func (w callExprWrapper) Fun() itypes.Expr      { return wrapExpr(w.x.Fun) }
func (w callExprWrapper) Ellipsis() itypes.Pos  { return w.x.Ellipsis }
func (w callExprWrapper) Rparen() itypes.Pos    { return w.x.Rparen }
func (w callExprWrapper) Pos() itypes.Pos       { return w.x.Pos() }
func (w callExprWrapper) End() itypes.Pos       { return w.x.End() }
func (w callExprWrapper) Unwrap() ast.Node      { return w.x }

func wrapCallExpr(x *ast.CallExpr) itypes.CallExpr {
	if x == nil {
		return nil
	}
	return callExprWrapper{x}
}

type starExprWrapper struct{ x *ast.StarExpr }

func (starExprWrapper) AnExpr()            {}
func (starExprWrapper) AStarExpr()         {}
func (w starExprWrapper) X() itypes.Expr   { return wrapExpr(w.x.X) }
func (w starExprWrapper) Pos() itypes.Pos  { return w.x.Pos() }
func (w starExprWrapper) End() itypes.Pos  { return w.x.End() }
func (w starExprWrapper) Unwrap() ast.Node { return w.x }

type unaryExprWrapper struct{ x *ast.UnaryExpr }

func (unaryExprWrapper) AnExpr()            {}
func (unaryExprWrapper) AUnaryExpr()        {}
func (w unaryExprWrapper) X() itypes.Expr   { return wrapExpr(w.x.X) }
func (w unaryExprWrapper) Op() token.Token  { return w.x.Op }
func (w unaryExprWrapper) Pos() itypes.Pos  { return w.x.Pos() }
func (w unaryExprWrapper) End() itypes.Pos  { return w.x.End() }
func (w unaryExprWrapper) Unwrap() ast.Node { return w.x }

type binaryExprWrapper struct{ x *ast.BinaryExpr }

func (binaryExprWrapper) AnExpr()            {}
func (binaryExprWrapper) ABinaryExpr()       {}
func (w binaryExprWrapper) Op() token.Token  { return w.x.Op }
func (w binaryExprWrapper) X() itypes.Expr   { return wrapExpr(w.x.X) }
func (w binaryExprWrapper) Y() itypes.Expr   { return wrapExpr(w.x.Y) }
func (w binaryExprWrapper) Pos() itypes.Pos  { return w.x.Pos() }
func (w binaryExprWrapper) End() itypes.Pos  { return w.x.End() }
func (w binaryExprWrapper) Unwrap() ast.Node { return w.x }

type keyValueExprWrapper struct{ x *ast.KeyValueExpr }

func (keyValueExprWrapper) AnExpr()              {}
func (keyValueExprWrapper) AKeyValueExpr()       {}
func (w keyValueExprWrapper) Key() itypes.Expr   { return wrapExpr(w.x.Key) }
func (w keyValueExprWrapper) Value() itypes.Expr { return wrapExpr(w.x.Value) }
func (w keyValueExprWrapper) Pos() itypes.Pos    { return w.x.Pos() }
func (w keyValueExprWrapper) End() itypes.Pos    { return w.x.End() }
func (w keyValueExprWrapper) Unwrap() ast.Node   { return w.x }

type arrayTypeWrapper struct{ x *ast.ArrayType }

func (arrayTypeWrapper) AnExpr()            {}
func (arrayTypeWrapper) AnArrayType()       {}
func (w arrayTypeWrapper) Len() itypes.Expr { return wrapExpr(w.x.Len) }
func (w arrayTypeWrapper) Elt() itypes.Expr { return wrapExpr(w.x.Elt) }
func (w arrayTypeWrapper) Pos() itypes.Pos  { return w.x.Pos() }
func (w arrayTypeWrapper) End() itypes.Pos  { return w.x.End() }
func (w arrayTypeWrapper) Unwrap() ast.Node { return w.x }

type structTypeWrapper struct{ x *ast.StructType }

func (structTypeWrapper) AnExpr()                    {}
func (structTypeWrapper) AStructType()               {}
func (w structTypeWrapper) Fields() itypes.FieldList { return wrapFieldList(w.x.Fields) }
func (w structTypeWrapper) Pos() itypes.Pos          { return w.x.Pos() }
func (w structTypeWrapper) End() itypes.Pos          { return w.x.End() }
func (w structTypeWrapper) Unwrap() ast.Node         { return w.x }

type funcTypeWrapper struct{ x *ast.FuncType }

func (funcTypeWrapper) AnExpr()                     {}
func (funcTypeWrapper) AFuncType()                  {}
func (w funcTypeWrapper) Params() itypes.FieldList  { return wrapFieldList(w.x.Params) }
func (w funcTypeWrapper) Results() itypes.FieldList { return wrapFieldList(w.x.Results) }
func (w funcTypeWrapper) Pos() itypes.Pos           { return w.x.Pos() }
func (w funcTypeWrapper) End() itypes.Pos           { return w.x.End() }
func (w funcTypeWrapper) Unwrap() ast.Node          { return w.x }

func wrapFuncType(x *ast.FuncType) itypes.FuncType {
	if x == nil {
		return nil
	}
	return funcTypeWrapper{x: x}
}

type interfaceTypeWrapper struct{ x *ast.InterfaceType }

func (interfaceTypeWrapper) AnExpr()                     {}
func (interfaceTypeWrapper) AnInterfaceType()            {}
func (w interfaceTypeWrapper) Methods() itypes.FieldList { return wrapFieldList(w.x.Methods) }
func (w interfaceTypeWrapper) Pos() itypes.Pos           { return w.x.Pos() }
func (w interfaceTypeWrapper) End() itypes.Pos           { return w.x.End() }
func (w interfaceTypeWrapper) Unwrap() ast.Node          { return w.x }

type fieldListWrapper struct{ *ast.FieldList }

func (w fieldListWrapper) Len() int                 { return len(w.FieldList.List) }
func (w fieldListWrapper) Field(i int) itypes.Field { return wrapField(w.List[i]) }
func (w fieldListWrapper) Pos() itypes.Pos          { return w.FieldList.Pos() }
func (w fieldListWrapper) End() itypes.Pos          { return w.FieldList.End() }
func (w fieldListWrapper) Unwrap() ast.Node         { return w.FieldList }

func wrapFieldList(list *ast.FieldList) itypes.FieldList {
	if list == nil {
		return nil
	}
	return fieldListWrapper{list}
}

type fieldWrapper struct{ *ast.Field }

func (w fieldWrapper) NamesLen() int           { return len(w.Field.Names) }
func (w fieldWrapper) Name(i int) itypes.Ident { return wrapIdent(w.Field.Names[i]) }
func (w fieldWrapper) Pos() itypes.Pos         { return w.Field.Pos() }
func (w fieldWrapper) End() itypes.Pos         { return w.Field.End() }
func (w fieldWrapper) Unwrap() ast.Node        { return w.Field }
func (w fieldWrapper) Type() itypes.Expr       { return wrapExpr(w.Field.Type) }
func (w fieldWrapper) Tag() itypes.BasicLit    { return wrapBasicLit(w.Field.Tag) }

func wrapField(field *ast.Field) itypes.Field {
	if field == nil {
		return nil
	}
	return fieldWrapper{field}
}

type exprListWrapper []ast.Expr

func (w exprListWrapper) Len() int               { return len(w) }
func (w exprListWrapper) Expr(i int) itypes.Expr { return wrapExpr(w[i]) }

type mapTypeWrapper struct {
	x *ast.MapType
}

func (mapTypeWrapper) AnExpr()              {}
func (mapTypeWrapper) AMapType()            {}
func (w mapTypeWrapper) Key() itypes.Expr   { return wrapExpr(w.x.Key) }
func (w mapTypeWrapper) Value() itypes.Expr { return wrapExpr(w.x.Value) }

func (w mapTypeWrapper) Pos() itypes.Pos  { return w.x.Pos() }
func (w mapTypeWrapper) End() itypes.Pos  { return w.x.End() }
func (w mapTypeWrapper) Unwrap() ast.Node { return w.x }

type chanTypeWrapper struct{ x *ast.ChanType }

func (w chanTypeWrapper) AnExpr()            {}
func (chanTypeWrapper) AChanType()           {}
func (w chanTypeWrapper) Dir() ast.ChanDir   { return w.x.Dir }
func (w chanTypeWrapper) Value() itypes.Expr { return wrapExpr(w.x.Value) }
func (w chanTypeWrapper) Pos() itypes.Pos    { return w.x.Pos() }
func (w chanTypeWrapper) End() itypes.Pos    { return w.x.End() }
func (w chanTypeWrapper) Unwrap() ast.Node   { return w.x }

func wrapStmt(stmt ast.Stmt) itypes.Stmt {
	switch s := stmt.(type) {
	case nil:
		return nil
	case *ast.BadStmt:
		return badStmtWrapper{s}
	case *ast.DeclStmt:
		return declStmtWrapper{s}
	case *ast.EmptyStmt:
		return emptyStmtWrapper{s}
	case *ast.LabeledStmt:
		return labeledStmtWrapper{s}
	case *ast.ExprStmt:
		return exprStmtWrapper{s}
	case *ast.SendStmt:
		return sendStmtWrapper{s}
	case *ast.IncDecStmt:
		return incDecStmtWrapper{s}
	case *ast.AssignStmt:
		return assignStmtWrapper{s}
	case *ast.GoStmt:
		return goStmtWrapper{s}
	case *ast.DeferStmt:
		return deferStmtWrapper{s}
	case *ast.ReturnStmt:
		return returnStmtWrapper{s}
	case *ast.BranchStmt:
		return branchStmtWrapper{s}
	case *ast.BlockStmt:
		return blockStmtWrapper{s}
	case *ast.IfStmt:
		return ifStmtWrapper{s}
	case *ast.CaseClause:
		return caseClauseWrapper{s}
	case *ast.SwitchStmt:
		return switchStmtWrapper{s}
	case *ast.TypeSwitchStmt:
		return typeSwitchStmtWrapper{s}
	case *ast.CommClause:
		return commClauseWrapper{s}
	case *ast.SelectStmt:
		return selectStmtWrapper{s}
	case *ast.ForStmt:
		return forStmtWrapper{s}
	case *ast.RangeStmt:
		return rangeStmtWrapper{s}
	}
	panic(fmt.Sprintf("stmt %T", stmt))
}

type badStmtWrapper struct{ s *ast.BadStmt }

func (badStmtWrapper) AStmt()             {}
func (badStmtWrapper) ABadStmt()          {}
func (s badStmtWrapper) Pos() itypes.Pos  { return s.s.Pos() }
func (s badStmtWrapper) End() itypes.Pos  { return s.s.End() }
func (s badStmtWrapper) Unwrap() ast.Node { return s.s }

type declStmtWrapper struct{ s *ast.DeclStmt }

func (declStmtWrapper) AStmt()              {}
func (declStmtWrapper) ADeclStmt()          {}
func (s declStmtWrapper) Decl() itypes.Decl { return wrapDecl(s.s.Decl) }
func (s declStmtWrapper) Pos() itypes.Pos   { return s.s.Pos() }
func (s declStmtWrapper) End() itypes.Pos   { return s.s.End() }
func (s declStmtWrapper) Unwrap() ast.Node  { return s.s }

type emptyStmtWrapper struct{ s *ast.EmptyStmt }

func (emptyStmtWrapper) AStmt()             {}
func (emptyStmtWrapper) AnEmptyStmt()       {}
func (s emptyStmtWrapper) Pos() itypes.Pos  { return s.s.Pos() }
func (s emptyStmtWrapper) End() itypes.Pos  { return s.s.End() }
func (s emptyStmtWrapper) Unwrap() ast.Node { return s.s }

type labeledStmtWrapper struct{ s *ast.LabeledStmt }

func (labeledStmtWrapper) AStmt()                {}
func (labeledStmtWrapper) ALabeledStmt()         {}
func (s labeledStmtWrapper) Label() itypes.Ident { return wrapIdent(s.s.Label) }
func (s labeledStmtWrapper) Stmt() itypes.Stmt   { return wrapStmt(s.s.Stmt) }
func (s labeledStmtWrapper) Pos() itypes.Pos     { return s.s.Pos() }
func (s labeledStmtWrapper) End() itypes.Pos     { return s.s.End() }
func (s labeledStmtWrapper) Unwrap() ast.Node    { return s.s }

type exprStmtWrapper struct{ s *ast.ExprStmt }

func (exprStmtWrapper) AStmt()             {}
func (exprStmtWrapper) AnExprStmt()        {}
func (s exprStmtWrapper) X() itypes.Expr   { return wrapExpr(s.s.X) }
func (s exprStmtWrapper) Pos() itypes.Pos  { return s.s.Pos() }
func (s exprStmtWrapper) End() itypes.Pos  { return s.s.End() }
func (s exprStmtWrapper) Unwrap() ast.Node { return s.s }

type sendStmtWrapper struct{ s *ast.SendStmt }

func (sendStmtWrapper) AStmt()               {}
func (sendStmtWrapper) ASendStmt()           {}
func (s sendStmtWrapper) Chan() itypes.Expr  { return wrapExpr(s.s.Chan) }
func (s sendStmtWrapper) Value() itypes.Expr { return wrapExpr(s.s.Value) }
func (s sendStmtWrapper) Arrow() itypes.Pos  { return s.s.Arrow }
func (s sendStmtWrapper) Pos() itypes.Pos    { return s.s.Pos() }
func (s sendStmtWrapper) End() itypes.Pos    { return s.s.End() }
func (s sendStmtWrapper) Unwrap() ast.Node   { return s.s }

type incDecStmtWrapper struct{ s *ast.IncDecStmt }

func (incDecStmtWrapper) AStmt()               {}
func (incDecStmtWrapper) AnIncDecStmt()        {}
func (s incDecStmtWrapper) Tok() token.Token   { return s.s.Tok }
func (s incDecStmtWrapper) TokPos() itypes.Pos { return s.s.TokPos }
func (s incDecStmtWrapper) X() itypes.Expr     { return wrapExpr(s.s.X) }
func (s incDecStmtWrapper) Pos() itypes.Pos    { return s.s.Pos() }
func (s incDecStmtWrapper) End() itypes.Pos    { return s.s.End() }
func (s incDecStmtWrapper) Unwrap() ast.Node   { return s.s }

type assignStmtWrapper struct{ s *ast.AssignStmt }

func (assignStmtWrapper) AStmt()                      {}
func (assignStmtWrapper) AnAssignStmt()               {}
func (s assignStmtWrapper) Lhs() itypes.ExprList      { return exprListWrapper(s.s.Lhs) }
func (s assignStmtWrapper) LhsLen() int               { return len(s.s.Lhs) }
func (s assignStmtWrapper) LhsExpr(i int) itypes.Expr { return wrapExpr(s.s.Lhs[i]) }
func (s assignStmtWrapper) RhsLen() int               { return len(s.s.Rhs) }
func (s assignStmtWrapper) RhsExpr(i int) itypes.Expr { return wrapExpr(s.s.Rhs[i]) }
func (s assignStmtWrapper) Rhs() itypes.ExprList      { return exprListWrapper(s.s.Rhs) }
func (s assignStmtWrapper) Tok() token.Token          { return s.s.Tok }
func (s assignStmtWrapper) TokPos() itypes.Pos        { return s.s.TokPos }

func (s assignStmtWrapper) Pos() itypes.Pos  { return s.s.Pos() }
func (s assignStmtWrapper) End() itypes.Pos  { return s.s.End() }
func (s assignStmtWrapper) Unwrap() ast.Node { return s.s }

type goStmtWrapper struct{ s *ast.GoStmt }

func (goStmtWrapper) AStmt()                  {}
func (goStmtWrapper) AGoStmt()                {}
func (s goStmtWrapper) Call() itypes.CallExpr { return wrapCallExpr(s.s.Call) }
func (s goStmtWrapper) Pos() itypes.Pos       { return s.s.Pos() }
func (s goStmtWrapper) End() itypes.Pos       { return s.s.End() }
func (s goStmtWrapper) Unwrap() ast.Node      { return s.s }

type deferStmtWrapper struct{ s *ast.DeferStmt }

func (deferStmtWrapper) AStmt()                  {}
func (deferStmtWrapper) ADeferStmt()             {}
func (s deferStmtWrapper) Call() itypes.CallExpr { return wrapCallExpr(s.s.Call) }
func (s deferStmtWrapper) Pos() itypes.Pos       { return s.s.Pos() }
func (s deferStmtWrapper) End() itypes.Pos       { return s.s.End() }
func (s deferStmtWrapper) Unwrap() ast.Node      { return s.s }

type returnStmtWrapper struct{ s *ast.ReturnStmt }

func (returnStmtWrapper) AStmt()                     {}
func (returnStmtWrapper) AReturnStmt()               {}
func (s returnStmtWrapper) Results() itypes.ExprList { return exprListWrapper(s.s.Results) }
func (s returnStmtWrapper) ResultsLen() int          { return len(s.s.Results) }
func (s returnStmtWrapper) Result(i int) itypes.Expr { return wrapExpr(s.s.Results[i]) }
func (s returnStmtWrapper) Return() itypes.Pos       { return s.s.Return }

func (s returnStmtWrapper) Pos() itypes.Pos  { return s.s.Pos() }
func (s returnStmtWrapper) End() itypes.Pos  { return s.s.End() }
func (s returnStmtWrapper) Unwrap() ast.Node { return s.s }

type branchStmtWrapper struct{ s *ast.BranchStmt }

func (branchStmtWrapper) AStmt()                {}
func (branchStmtWrapper) ABranchStmt()          {}
func (s branchStmtWrapper) Tok() token.Token    { return s.s.Tok }
func (s branchStmtWrapper) Label() itypes.Ident { return wrapIdent(s.s.Label) }

func (s branchStmtWrapper) Pos() itypes.Pos  { return s.s.Pos() }
func (s branchStmtWrapper) End() itypes.Pos  { return s.s.End() }
func (s branchStmtWrapper) Unwrap() ast.Node { return s.s }

type blockStmtWrapper struct{ s *ast.BlockStmt }

func (blockStmtWrapper) AStmt()                  {}
func (blockStmtWrapper) ABlockStmt()             {}
func (s blockStmtWrapper) List() itypes.StmtList { return stmtListWrapper(s.s.List) }
func (s blockStmtWrapper) Lbrace() itypes.Pos    { return s.s.Lbrace }
func (s blockStmtWrapper) Rbrace() itypes.Pos    { return s.s.Rbrace }
func (s blockStmtWrapper) Pos() itypes.Pos       { return s.s.Pos() }
func (s blockStmtWrapper) End() itypes.Pos       { return s.s.End() }
func (s blockStmtWrapper) Unwrap() ast.Node      { return s.s }

func wrapBlockStmt(s *ast.BlockStmt) itypes.BlockStmt {
	if s == nil {
		return nil
	}
	return blockStmtWrapper{s: s}
}

type ifStmtWrapper struct{ s *ast.IfStmt }

func (ifStmtWrapper) AStmt()                   {}
func (ifStmtWrapper) AnIfStmt()                {}
func (s ifStmtWrapper) Init() itypes.Stmt      { return wrapStmt(s.s.Init) }
func (s ifStmtWrapper) Cond() itypes.Expr      { return wrapExpr(s.s.Cond) }
func (s ifStmtWrapper) Body() itypes.BlockStmt { return wrapBlockStmt(s.s.Body) }
func (s ifStmtWrapper) Else() itypes.Stmt      { return wrapStmt(s.s.Else) }

func (s ifStmtWrapper) Pos() itypes.Pos  { return s.s.Pos() }
func (s ifStmtWrapper) End() itypes.Pos  { return s.s.End() }
func (s ifStmtWrapper) Unwrap() ast.Node { return s.s }

type caseClauseWrapper struct{ s *ast.CaseClause }

func (caseClauseWrapper) AStmt()                   {}
func (caseClauseWrapper) ACaseClause()             {}
func (s caseClauseWrapper) ListLen() int           { return len(s.s.List) }
func (s caseClauseWrapper) Item(i int) itypes.Expr { return wrapExpr(s.s.List[i]) }
func (s caseClauseWrapper) Body() itypes.StmtList  { return stmtListWrapper(s.s.Body) }
func (s caseClauseWrapper) Pos() itypes.Pos        { return s.s.Pos() }
func (s caseClauseWrapper) End() itypes.Pos        { return s.s.End() }
func (s caseClauseWrapper) Unwrap() ast.Node       { return s.s }

type switchStmtWrapper struct{ s *ast.SwitchStmt }

func (switchStmtWrapper) AStmt()                   {}
func (switchStmtWrapper) ASwitchStmt()             {}
func (s switchStmtWrapper) Init() itypes.Stmt      { return wrapStmt(s.s.Init) }
func (s switchStmtWrapper) Tag() itypes.Expr       { return wrapExpr(s.s.Tag) }
func (s switchStmtWrapper) Body() itypes.BlockStmt { return wrapBlockStmt(s.s.Body) }
func (s switchStmtWrapper) Pos() itypes.Pos        { return s.s.Pos() }
func (s switchStmtWrapper) End() itypes.Pos        { return s.s.End() }
func (s switchStmtWrapper) Unwrap() ast.Node       { return s.s }

type typeSwitchStmtWrapper struct{ s *ast.TypeSwitchStmt }

func (typeSwitchStmtWrapper) AStmt()                   {}
func (typeSwitchStmtWrapper) ATypeSwitchStmt()         {}
func (s typeSwitchStmtWrapper) Init() itypes.Stmt      { return wrapStmt(s.s.Init) }
func (s typeSwitchStmtWrapper) Assign() itypes.Stmt    { return wrapStmt(s.s.Assign) }
func (s typeSwitchStmtWrapper) Body() itypes.BlockStmt { return wrapBlockStmt(s.s.Body) }

func (s typeSwitchStmtWrapper) Pos() itypes.Pos  { return s.s.Pos() }
func (s typeSwitchStmtWrapper) End() itypes.Pos  { return s.s.End() }
func (s typeSwitchStmtWrapper) Unwrap() ast.Node { return s.s }

type commClauseWrapper struct{ s *ast.CommClause }

func (commClauseWrapper) AStmt()                  {}
func (commClauseWrapper) ACommClause()            {}
func (s commClauseWrapper) Comm() itypes.Stmt     { return wrapStmt(s.s.Comm) }
func (s commClauseWrapper) Body() itypes.StmtList { return stmtListWrapper(s.s.Body) }
func (s commClauseWrapper) Pos() itypes.Pos       { return s.s.Pos() }
func (s commClauseWrapper) End() itypes.Pos       { return s.s.End() }
func (s commClauseWrapper) Unwrap() ast.Node      { return s.s }

type selectStmtWrapper struct{ s *ast.SelectStmt }

func (selectStmtWrapper) AStmt()                   {}
func (selectStmtWrapper) ASelectStmt()             {}
func (s selectStmtWrapper) Body() itypes.BlockStmt { return wrapBlockStmt(s.s.Body) }
func (s selectStmtWrapper) Pos() itypes.Pos        { return s.s.Pos() }
func (s selectStmtWrapper) End() itypes.Pos        { return s.s.End() }
func (s selectStmtWrapper) Unwrap() ast.Node       { return s.s }

type forStmtWrapper struct{ s *ast.ForStmt }

func (forStmtWrapper) AStmt()                   {}
func (forStmtWrapper) AForStmt()                {}
func (s forStmtWrapper) Init() itypes.Stmt      { return wrapStmt(s.s.Init) }
func (s forStmtWrapper) Cond() itypes.Expr      { return wrapExpr(s.s.Cond) }
func (s forStmtWrapper) Post() itypes.Stmt      { return wrapStmt(s.s.Post) }
func (s forStmtWrapper) Body() itypes.BlockStmt { return wrapBlockStmt(s.s.Body) }

func (s forStmtWrapper) Pos() itypes.Pos  { return s.s.Pos() }
func (s forStmtWrapper) End() itypes.Pos  { return s.s.End() }
func (s forStmtWrapper) Unwrap() ast.Node { return s.s }

type rangeStmtWrapper struct{ s *ast.RangeStmt }

func (rangeStmtWrapper) AStmt()                   {}
func (rangeStmtWrapper) ARangeStmt()              {}
func (s rangeStmtWrapper) Key() itypes.Expr       { return wrapExpr(s.s.Key) }
func (s rangeStmtWrapper) Value() itypes.Expr     { return wrapExpr(s.s.Value) }
func (s rangeStmtWrapper) X() itypes.Expr         { return wrapExpr(s.s.X) }
func (s rangeStmtWrapper) Body() itypes.BlockStmt { return wrapBlockStmt(s.s.Body) }
func (s rangeStmtWrapper) Tok() token.Token       { return s.s.Tok }
func (s rangeStmtWrapper) TokPos() itypes.Pos     { return s.s.TokPos }
func (s rangeStmtWrapper) Pos() itypes.Pos        { return s.s.Pos() }
func (s rangeStmtWrapper) End() itypes.Pos        { return s.s.End() }
func (s rangeStmtWrapper) Unwrap() ast.Node       { return s.s }

type stmtListWrapper []ast.Stmt

func (w stmtListWrapper) Len() int                   { return len(w) }
func (w stmtListWrapper) Stmt(i int) itypes.Stmt     { return wrapStmt(w[i]) }
func (w stmtListWrapper) Head(i int) itypes.StmtList { return stmtListWrapper(w[:i]) }

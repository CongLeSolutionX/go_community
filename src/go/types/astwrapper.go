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
func (w fileWrapper) DeclsLen() int        { return len(w.File.Decls) }
func (w fileWrapper) Decl(i int) astDecl   { return wrapDecl(w.File.Decls[i]) }
func (w fileWrapper) Name() astIdent       { return wrapIdent(w.File.Name) }
func (w fileWrapper) Package() astPosition { return w.File.Package }

func wrapDecl(d ast.Decl) astDecl {
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

func (w badDeclWrapper) aDecl()           {}
func (w badDeclWrapper) aBadDecl()        {}
func (w badDeclWrapper) Unwrap() ast.Node { return w.d }
func (w badDeclWrapper) Pos() astPosition { return w.d.Pos() }
func (w badDeclWrapper) End() astPosition { return w.d.End() }

type genDeclWrapper struct{ d *ast.GenDecl }

func (w genDeclWrapper) aDecl()             {}
func (w genDeclWrapper) aGenDecl()          {}
func (w genDeclWrapper) SpecsLen() int      { return len(w.d.Specs) }
func (w genDeclWrapper) Spec(i int) astSpec { return wrapSpec(w.d.Specs[i]) }
func (w genDeclWrapper) Tok() token.Token   { return w.d.Tok }
func (w genDeclWrapper) Unwrap() ast.Node   { return w.d }
func (w genDeclWrapper) Pos() astPosition   { return w.d.Pos() }
func (w genDeclWrapper) End() astPosition   { return w.d.End() }

type funcDeclWrapper struct{ d *ast.FuncDecl }

func (w funcDeclWrapper) aDecl()             {}
func (w funcDeclWrapper) aFuncDecl()         {}
func (w funcDeclWrapper) Recv() astFieldList { return wrapFieldList(w.d.Recv) }
func (w funcDeclWrapper) Name() astIdent     { return wrapIdent(w.d.Name) }
func (w funcDeclWrapper) Type() astFuncType  { return wrapFuncType(w.d.Type) }
func (w funcDeclWrapper) Body() astBlockStmt { return wrapBlockStmt(w.d.Body) }
func (w funcDeclWrapper) Unwrap() ast.Node   { return w.d }
func (w funcDeclWrapper) Pos() astPosition   { return w.d.Pos() }
func (w funcDeclWrapper) End() astPosition   { return w.d.End() }

func wrapSpec(spec ast.Spec) astSpec {
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

func (w importSpecWrapper) aSpec()            {}
func (w importSpecWrapper) anImportSpec()     {}
func (w importSpecWrapper) Path() astBasicLit { return wrapBasicLit(w.s.Path) }
func (w importSpecWrapper) Name() astIdent    { return wrapIdent(w.s.Name) }
func (w importSpecWrapper) Pos() astPosition  { return w.s.Pos() }
func (w importSpecWrapper) End() astPosition  { return w.s.End() }
func (w importSpecWrapper) Unwrap() ast.Node  { return w.s }

type valueSpecWrapper struct{ s *ast.ValueSpec }

func (valueSpecWrapper) aSpec()                {}
func (valueSpecWrapper) aValueSpec()           {}
func (w valueSpecWrapper) NamesLen() int       { return len(w.s.Names) }
func (w valueSpecWrapper) Name(i int) astIdent { return wrapIdent(w.s.Names[i]) }
func (w valueSpecWrapper) Type() astExpr       { return wrapExpr(w.s.Type) }
func (w valueSpecWrapper) ValuesLen() int      { return len(w.s.Values) }
func (w valueSpecWrapper) Value(i int) astExpr { return wrapExpr(w.s.Values[i]) }
func (w valueSpecWrapper) Pos() astPosition    { return w.s.Pos() }
func (w valueSpecWrapper) End() astPosition    { return w.s.End() }
func (w valueSpecWrapper) Unwrap() ast.Node    { return w.s }

func wrapValueSpec(s *ast.ValueSpec) astValueSpec {
	if s == nil {
		return nil
	}
	return valueSpecWrapper{s}
}

func astNewValueSpec() astValueSpec {
	return wrapValueSpec(new(ast.ValueSpec))
}

type typeSpecWrapper struct{ s *ast.TypeSpec }

func (w typeSpecWrapper) aSpec()              {}
func (w typeSpecWrapper) aTypeSpec()          {}
func (w typeSpecWrapper) Name() astIdent      { return wrapIdent(w.s.Name) }
func (w typeSpecWrapper) Assign() astPosition { return w.s.Assign }
func (w typeSpecWrapper) Type() astExpr       { return wrapExpr(w.s.Type) }
func (w typeSpecWrapper) Pos() astPosition    { return w.s.Pos() }
func (w typeSpecWrapper) End() astPosition    { return w.s.End() }
func (w typeSpecWrapper) Unwrap() ast.Node    { return w.s }

func wrapExpr(expr ast.Expr) astExpr {
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
		return (*basicLitWrapper)(x)
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
		return (*callExprWrapper)(x)
	case *ast.StarExpr:
		return (*starExprWrapper)(x)
	case *ast.UnaryExpr:
		return (*unaryExprWrapper)(x)
	case *ast.BinaryExpr:
		return (*binaryExprWrapper)(x)
	case *ast.KeyValueExpr:
		return (*keyValueExprWrapper)(x)
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

func (badExprWrapper) anExpr()            {}
func (badExprWrapper) anBadExpr()         {}
func (w badExprWrapper) Pos() astPosition { return w.x.Pos() }
func (w badExprWrapper) End() astPosition { return w.x.End() }
func (w badExprWrapper) Unwrap() ast.Node { return w.x }

type identWrapper struct{ x *ast.Ident }

func (identWrapper) anExpr()             {}
func (identWrapper) anIdent()            {}
func (w identWrapper) IdentName() string { return w.x.Name }
func (w identWrapper) Pos() astPosition  { return w.x.Pos() }
func (w identWrapper) End() astPosition  { return w.x.End() }
func (w identWrapper) Unwrap() ast.Node  { return w.x }
func (w identWrapper) String() string    { return w.x.String() }

func astNewIdent(name string, namePos astPosition) astIdent {
	ident := ast.NewIdent(name)
	ident.NamePos = namePos.(token.Pos)
	return wrapIdent(ident)
}

func wrapIdent(ident *ast.Ident) astIdent {
	if ident == nil {
		return nil
	}
	return identWrapper{ident}
}

type ellipsisWrapper struct{ x *ast.Ellipsis }

func (ellipsisWrapper) anExpr()            {}
func (ellipsisWrapper) anEllipsis()        {}
func (w ellipsisWrapper) Pos() astPosition { return w.x.Pos() }
func (w ellipsisWrapper) End() astPosition { return w.x.End() }
func (w ellipsisWrapper) Unwrap() ast.Node { return w.x }
func (w ellipsisWrapper) Elt() astExpr     { return wrapExpr(w.x.Elt) }

type basicLitWrapper ast.BasicLit

func astNewBasicLit(pos astPosition, kind token.Token, value string) astBasicLit {
	lit := &ast.BasicLit{ValuePos: pos.(token.Pos), Kind: kind, Value: value}
	return wrapBasicLit(lit)
}

func (basicLitWrapper) anExpr()             {}
func (basicLitWrapper) anBasicLit()         {}
func (w *basicLitWrapper) Pos() astPosition { return (*ast.BasicLit)(w).Pos() }
func (w *basicLitWrapper) End() astPosition { return (*ast.BasicLit)(w).End() }
func (w *basicLitWrapper) Unwrap() ast.Node { return (*ast.BasicLit)(w) }

func wrapBasicLit(x *ast.BasicLit) astBasicLit {
	if x == nil {
		return nil
	}
	return (*basicLitWrapper)(x)
}

func (w basicLitWrapper) LitKind() token.Token { return w.Kind }
func (w basicLitWrapper) LitValue() string     { return w.Value }

type funcLitWrapper struct{ x *ast.FuncLit }

func (funcLitWrapper) anExpr()              {}
func (funcLitWrapper) anFuncLit()           {}
func (w funcLitWrapper) Type() astExpr      { return wrapExpr(w.x.Type) }
func (w funcLitWrapper) Body() astBlockStmt { return wrapBlockStmt(w.x.Body) }
func (w funcLitWrapper) Pos() astPosition   { return w.x.Pos() }
func (w funcLitWrapper) End() astPosition   { return w.x.End() }
func (w funcLitWrapper) Unwrap() ast.Node   { return w.x }

type compositeLitWrapper struct{ x *ast.CompositeLit }

func (compositeLitWrapper) anExpr()               {}
func (compositeLitWrapper) anCompositeLit()       {}
func (w compositeLitWrapper) Type() astExpr       { return wrapExpr(w.x.Type) }
func (w compositeLitWrapper) EltsLen() int        { return len(w.x.Elts) }
func (w compositeLitWrapper) Elt(i int) astExpr   { return wrapExpr(w.x.Elts[i]) }
func (w compositeLitWrapper) Rbrace() astPosition { return w.x.Rbrace }
func (w compositeLitWrapper) Pos() astPosition    { return w.x.Pos() }
func (w compositeLitWrapper) End() astPosition    { return w.x.End() }
func (w compositeLitWrapper) Unwrap() ast.Node    { return w.x }

type parenExprWrapper struct{ x *ast.ParenExpr }

func (parenExprWrapper) anExpr()            {}
func (parenExprWrapper) anParenExpr()       {}
func (w parenExprWrapper) X() astExpr       { return wrapExpr(w.x.X) }
func (w parenExprWrapper) Pos() astPosition { return w.x.Pos() }
func (w parenExprWrapper) End() astPosition { return w.x.End() }
func (w parenExprWrapper) Unwrap() ast.Node { return w.x }

type selectorExprWrapper struct{ x *ast.SelectorExpr }

func (selectorExprWrapper) anExpr()            {}
func (selectorExprWrapper) anSelectorExpr()    {}
func (w selectorExprWrapper) X() astExpr       { return wrapExpr(w.x.X) }
func (w selectorExprWrapper) Sel() astIdent    { return wrapIdent(w.x.Sel) }
func (w selectorExprWrapper) Pos() astPosition { return w.x.Pos() }
func (w selectorExprWrapper) End() astPosition { return w.x.End() }
func (w selectorExprWrapper) Unwrap() ast.Node { return w.x }

type indexExprWrapper struct{ x *ast.IndexExpr }

func (indexExprWrapper) anExpr()            {}
func (indexExprWrapper) anIndexExpr()       {}
func (w indexExprWrapper) X() astExpr       { return wrapExpr(w.x.X) }
func (w indexExprWrapper) Index() astExpr   { return wrapExpr(w.x.Index) }
func (w indexExprWrapper) Pos() astPosition { return w.x.Pos() }
func (w indexExprWrapper) End() astPosition { return w.x.End() }
func (w indexExprWrapper) Unwrap() ast.Node { return w.x }

type sliceExprWrapper struct{ x *ast.SliceExpr }

func (sliceExprWrapper) anExpr()               {}
func (sliceExprWrapper) anSliceExpr()          {}
func (w sliceExprWrapper) X() astExpr          { return wrapExpr(w.x.X) }
func (w sliceExprWrapper) Low() astExpr        { return wrapExpr(w.x.Low) }
func (w sliceExprWrapper) High() astExpr       { return wrapExpr(w.x.High) }
func (w sliceExprWrapper) Max() astExpr        { return wrapExpr(w.x.Max) }
func (w sliceExprWrapper) Slice3() bool        { return w.x.Slice3 }
func (w sliceExprWrapper) Rbrack() astPosition { return w.x.Rbrack }
func (w sliceExprWrapper) Pos() astPosition    { return w.x.Pos() }
func (w sliceExprWrapper) End() astPosition    { return w.x.End() }
func (w sliceExprWrapper) Unwrap() ast.Node    { return w.x }

type typeAssertExprWrapper struct{ x *ast.TypeAssertExpr }

func (typeAssertExprWrapper) anExpr()            {}
func (typeAssertExprWrapper) anTypeAssertExpr()  {}
func (w typeAssertExprWrapper) X() astExpr       { return wrapExpr(w.x.X) }
func (w typeAssertExprWrapper) Type() astExpr    { return wrapExpr(w.x.Type) }
func (w typeAssertExprWrapper) Pos() astPosition { return w.x.Pos() }
func (w typeAssertExprWrapper) End() astPosition { return w.x.End() }
func (w typeAssertExprWrapper) Unwrap() ast.Node { return w.x }

type callExprWrapper ast.CallExpr

func (*callExprWrapper) anExpr()                    {}
func (*callExprWrapper) anCallExpr()                {}
func (w *callExprWrapper) ArgsLen() int             { return len((*ast.CallExpr)(w).Args) }
func (w *callExprWrapper) Arg(i int) astExpr        { return wrapExpr((*ast.CallExpr)(w).Args[i]) }
func (w *callExprWrapper) FunExpr() astExpr         { return wrapExpr((*ast.CallExpr)(w).Fun) }
func (w *callExprWrapper) EllipsisPos() astPosition { return (*ast.CallExpr)(w).Ellipsis }
func (w *callExprWrapper) RparenPos() astPosition   { return (*ast.CallExpr)(w).Rparen }
func (w *callExprWrapper) Pos() astPosition         { return (*ast.CallExpr)(w).Pos() }
func (w *callExprWrapper) End() astPosition         { return (*ast.CallExpr)(w).End() }
func (w *callExprWrapper) Unwrap() ast.Node         { return (*ast.CallExpr)(w) }

func wrapCallExpr(x *ast.CallExpr) astCallExpr {
	if x == nil {
		return nil
	}
	return (*callExprWrapper)(x)
}

type starExprWrapper ast.StarExpr

func (starExprWrapper) anExpr()             {}
func (starExprWrapper) anStarExpr()         {}
func (w *starExprWrapper) InnerX() astExpr  { return wrapExpr((*ast.StarExpr)(w).X) }
func (w *starExprWrapper) Pos() astPosition { return (*ast.StarExpr)(w).Pos() }
func (w *starExprWrapper) End() astPosition { return (*ast.StarExpr)(w).End() }
func (w *starExprWrapper) Unwrap() ast.Node { return (*ast.StarExpr)(w) }

type unaryExprWrapper ast.UnaryExpr

func (*unaryExprWrapper) anExpr()              {}
func (*unaryExprWrapper) anUnaryExpr()         {}
func (w *unaryExprWrapper) InnerX() astExpr    { return wrapExpr((*ast.UnaryExpr)(w).X) }
func (w *unaryExprWrapper) OpTok() token.Token { return (*ast.UnaryExpr)(w).Op }
func (w *unaryExprWrapper) Pos() astPosition   { return (*ast.UnaryExpr)(w).Pos() }
func (w *unaryExprWrapper) End() astPosition   { return (*ast.UnaryExpr)(w).End() }
func (w *unaryExprWrapper) Unwrap() ast.Node   { return (*ast.UnaryExpr)(w) }

type binaryExprWrapper ast.BinaryExpr

func (*binaryExprWrapper) anExpr()              {}
func (*binaryExprWrapper) anBinaryExpr()        {}
func (w *binaryExprWrapper) OpTok() token.Token { return (*ast.BinaryExpr)(w).Op }
func (w *binaryExprWrapper) XExpr() astExpr     { return wrapExpr((*ast.BinaryExpr)(w).X) }
func (w *binaryExprWrapper) YExpr() astExpr     { return wrapExpr((*ast.BinaryExpr)(w).Y) }
func (w *binaryExprWrapper) Pos() astPosition   { return (*ast.BinaryExpr)(w).Pos() }
func (w *binaryExprWrapper) End() astPosition   { return (*ast.BinaryExpr)(w).End() }
func (w *binaryExprWrapper) Unwrap() ast.Node   { return (*ast.BinaryExpr)(w) }

type keyValueExprWrapper ast.KeyValueExpr

func (*keyValueExprWrapper) anExpr()                 {}
func (*keyValueExprWrapper) anKeyValueExpr()         {}
func (w *keyValueExprWrapper) KeyExpr() astExpr      { return wrapExpr((*ast.KeyValueExpr)(w).Key) }
func (w *keyValueExprWrapper) ColonPos() astPosition { return (*ast.KeyValueExpr)(w).Colon }
func (w *keyValueExprWrapper) ValueExpr() astExpr    { return wrapExpr((*ast.KeyValueExpr)(w).Value) }
func (w *keyValueExprWrapper) Pos() astPosition      { return (*ast.KeyValueExpr)(w).Pos() }
func (w *keyValueExprWrapper) End() astPosition      { return (*ast.KeyValueExpr)(w).End() }
func (w *keyValueExprWrapper) Unwrap() ast.Node      { return (*ast.KeyValueExpr)(w) }

type arrayTypeWrapper struct{ x *ast.ArrayType }

func (arrayTypeWrapper) anExpr()            {}
func (arrayTypeWrapper) anArrayType()       {}
func (w arrayTypeWrapper) Len() astExpr     { return wrapExpr(w.x.Len) }
func (w arrayTypeWrapper) Elt() astExpr     { return wrapExpr(w.x.Elt) }
func (w arrayTypeWrapper) Pos() astPosition { return w.x.Pos() }
func (w arrayTypeWrapper) End() astPosition { return w.x.End() }
func (w arrayTypeWrapper) Unwrap() ast.Node { return w.x }

type structTypeWrapper struct{ x *ast.StructType }

func (structTypeWrapper) anExpr()                {}
func (structTypeWrapper) anStructType()          {}
func (w structTypeWrapper) Fields() astFieldList { return wrapFieldList(w.x.Fields) }
func (w structTypeWrapper) Pos() astPosition     { return w.x.Pos() }
func (w structTypeWrapper) End() astPosition     { return w.x.End() }
func (w structTypeWrapper) Unwrap() ast.Node     { return w.x }

type funcTypeWrapper struct{ x *ast.FuncType }

func (funcTypeWrapper) anExpr()                 {}
func (funcTypeWrapper) anFuncType()             {}
func (w funcTypeWrapper) Params() astFieldList  { return wrapFieldList(w.x.Params) }
func (w funcTypeWrapper) Results() astFieldList { return wrapFieldList(w.x.Results) }
func (w funcTypeWrapper) Pos() astPosition      { return w.x.Pos() }
func (w funcTypeWrapper) End() astPosition      { return w.x.End() }
func (w funcTypeWrapper) Unwrap() ast.Node      { return w.x }

func wrapFuncType(x *ast.FuncType) astFuncType {
	if x == nil {
		return nil
	}
	return funcTypeWrapper{x: x}
}

type interfaceTypeWrapper struct{ x *ast.InterfaceType }

func (interfaceTypeWrapper) anExpr()                 {}
func (interfaceTypeWrapper) anInterfaceType()        {}
func (w interfaceTypeWrapper) Methods() astFieldList { return wrapFieldList(w.x.Methods) }
func (w interfaceTypeWrapper) Pos() astPosition      { return w.x.Pos() }
func (w interfaceTypeWrapper) End() astPosition      { return w.x.End() }
func (w interfaceTypeWrapper) Unwrap() ast.Node      { return w.x }

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

func (w fieldWrapper) NamesLen() int       { return len(w.Field.Names) }
func (w fieldWrapper) Name(i int) astIdent { return wrapIdent(w.Field.Names[i]) }
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

type exprListWrapper []ast.Expr

func (w exprListWrapper) Len() int           { return len(w) }
func (w exprListWrapper) Expr(i int) astExpr { return wrapExpr(w[i]) }

type mapTypeWrapper struct {
	x *ast.MapType
}

func (mapTypeWrapper) anExpr()          {}
func (mapTypeWrapper) anMapType()       {}
func (w mapTypeWrapper) Key() astExpr   { return wrapExpr(w.x.Key) }
func (w mapTypeWrapper) Value() astExpr { return wrapExpr(w.x.Value) }

func (w mapTypeWrapper) Pos() astPosition { return w.x.Pos() }
func (w mapTypeWrapper) End() astPosition { return w.x.End() }
func (w mapTypeWrapper) Unwrap() ast.Node { return w.x }

type chanTypeWrapper struct{ x *ast.ChanType }

func (w chanTypeWrapper) anExpr()          {}
func (chanTypeWrapper) anChanType()        {}
func (w chanTypeWrapper) Dir() ast.ChanDir { return w.x.Dir }
func (w chanTypeWrapper) Value() astExpr   { return wrapExpr(w.x.Value) }
func (w chanTypeWrapper) Pos() astPosition { return w.x.Pos() }
func (w chanTypeWrapper) End() astPosition { return w.x.End() }
func (w chanTypeWrapper) Unwrap() ast.Node { return w.x }

func wrapStmt(stmt ast.Stmt) astStmt {
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

func (badStmtWrapper) aStmt()             {}
func (badStmtWrapper) aBadStmt()          {}
func (s badStmtWrapper) Pos() astPosition { return s.s.Pos() }
func (s badStmtWrapper) End() astPosition { return s.s.End() }
func (s badStmtWrapper) Unwrap() ast.Node { return s.s }

type declStmtWrapper struct{ s *ast.DeclStmt }

func (declStmtWrapper) aStmt()             {}
func (declStmtWrapper) aDeclStmt()         {}
func (s declStmtWrapper) Decl() astDecl    { return wrapDecl(s.s.Decl) }
func (s declStmtWrapper) Pos() astPosition { return s.s.Pos() }
func (s declStmtWrapper) End() astPosition { return s.s.End() }
func (s declStmtWrapper) Unwrap() ast.Node { return s.s }

type emptyStmtWrapper struct{ s *ast.EmptyStmt }

func (emptyStmtWrapper) aStmt()             {}
func (emptyStmtWrapper) aEmptyStmt()        {}
func (s emptyStmtWrapper) Pos() astPosition { return s.s.Pos() }
func (s emptyStmtWrapper) End() astPosition { return s.s.End() }
func (s emptyStmtWrapper) Unwrap() ast.Node { return s.s }

type labeledStmtWrapper struct{ s *ast.LabeledStmt }

func (labeledStmtWrapper) aStmt()             {}
func (labeledStmtWrapper) aLabeledStmt()      {}
func (s labeledStmtWrapper) Label() astIdent  { return wrapIdent(s.s.Label) }
func (s labeledStmtWrapper) Stmt() astStmt    { return wrapStmt(s.s.Stmt) }
func (s labeledStmtWrapper) Pos() astPosition { return s.s.Pos() }
func (s labeledStmtWrapper) End() astPosition { return s.s.End() }
func (s labeledStmtWrapper) Unwrap() ast.Node { return s.s }

type exprStmtWrapper struct{ s *ast.ExprStmt }

func (exprStmtWrapper) aStmt()             {}
func (exprStmtWrapper) aExprStmt()         {}
func (s exprStmtWrapper) X() astExpr       { return wrapExpr(s.s.X) }
func (s exprStmtWrapper) Pos() astPosition { return s.s.Pos() }
func (s exprStmtWrapper) End() astPosition { return s.s.End() }
func (s exprStmtWrapper) Unwrap() ast.Node { return s.s }

type sendStmtWrapper struct{ s *ast.SendStmt }

func (sendStmtWrapper) aStmt()               {}
func (sendStmtWrapper) aSendStmt()           {}
func (s sendStmtWrapper) Chan() astExpr      { return wrapExpr(s.s.Chan) }
func (s sendStmtWrapper) Value() astExpr     { return wrapExpr(s.s.Value) }
func (s sendStmtWrapper) Arrow() astPosition { return s.s.Arrow }
func (s sendStmtWrapper) Pos() astPosition   { return s.s.Pos() }
func (s sendStmtWrapper) End() astPosition   { return s.s.End() }
func (s sendStmtWrapper) Unwrap() ast.Node   { return s.s }

type incDecStmtWrapper struct{ s *ast.IncDecStmt }

func (incDecStmtWrapper) aStmt()                {}
func (incDecStmtWrapper) aIncDecStmt()          {}
func (s incDecStmtWrapper) Tok() token.Token    { return s.s.Tok }
func (s incDecStmtWrapper) TokPos() astPosition { return s.s.TokPos }
func (s incDecStmtWrapper) X() astExpr          { return wrapExpr(s.s.X) }
func (s incDecStmtWrapper) Pos() astPosition    { return s.s.Pos() }
func (s incDecStmtWrapper) End() astPosition    { return s.s.End() }
func (s incDecStmtWrapper) Unwrap() ast.Node    { return s.s }

type assignStmtWrapper struct{ s *ast.AssignStmt }

func (assignStmtWrapper) aStmt()                  {}
func (assignStmtWrapper) aAssignStmt()            {}
func (s assignStmtWrapper) Lhs() astExprList      { return exprListWrapper(s.s.Lhs) }
func (s assignStmtWrapper) LhsLen() int           { return len(s.s.Lhs) }
func (s assignStmtWrapper) LhsExpr(i int) astExpr { return wrapExpr(s.s.Lhs[i]) }
func (s assignStmtWrapper) RhsLen() int           { return len(s.s.Rhs) }
func (s assignStmtWrapper) RhsExpr(i int) astExpr { return wrapExpr(s.s.Rhs[i]) }
func (s assignStmtWrapper) Rhs() astExprList      { return exprListWrapper(s.s.Rhs) }
func (s assignStmtWrapper) Tok() token.Token      { return s.s.Tok }
func (s assignStmtWrapper) TokPos() astPosition   { return s.s.TokPos }

func (s assignStmtWrapper) Pos() astPosition { return s.s.Pos() }
func (s assignStmtWrapper) End() astPosition { return s.s.End() }
func (s assignStmtWrapper) Unwrap() ast.Node { return s.s }

type goStmtWrapper struct{ s *ast.GoStmt }

func (goStmtWrapper) aStmt()              {}
func (goStmtWrapper) aGoStmt()            {}
func (s goStmtWrapper) Call() astCallExpr { return wrapCallExpr(s.s.Call) }
func (s goStmtWrapper) Pos() astPosition  { return s.s.Pos() }
func (s goStmtWrapper) End() astPosition  { return s.s.End() }
func (s goStmtWrapper) Unwrap() ast.Node  { return s.s }

type deferStmtWrapper struct{ s *ast.DeferStmt }

func (deferStmtWrapper) aStmt()              {}
func (deferStmtWrapper) aDeferStmt()         {}
func (s deferStmtWrapper) Call() astCallExpr { return wrapCallExpr(s.s.Call) }
func (s deferStmtWrapper) Pos() astPosition  { return s.s.Pos() }
func (s deferStmtWrapper) End() astPosition  { return s.s.End() }
func (s deferStmtWrapper) Unwrap() ast.Node  { return s.s }

type returnStmtWrapper struct{ s *ast.ReturnStmt }

func (returnStmtWrapper) aStmt()                 {}
func (returnStmtWrapper) aReturnStmt()           {}
func (s returnStmtWrapper) Results() astExprList { return exprListWrapper(s.s.Results) }
func (s returnStmtWrapper) ResultsLen() int      { return len(s.s.Results) }
func (s returnStmtWrapper) Result(i int) astExpr { return wrapExpr(s.s.Results[i]) }
func (s returnStmtWrapper) Return() astPosition  { return s.s.Return }

func (s returnStmtWrapper) Pos() astPosition { return s.s.Pos() }
func (s returnStmtWrapper) End() astPosition { return s.s.End() }
func (s returnStmtWrapper) Unwrap() ast.Node { return s.s }

type branchStmtWrapper struct{ s *ast.BranchStmt }

func (branchStmtWrapper) aStmt()             {}
func (branchStmtWrapper) aBranchStmt()       {}
func (s branchStmtWrapper) Tok() token.Token { return s.s.Tok }
func (s branchStmtWrapper) Label() astIdent  { return wrapIdent(s.s.Label) }

func (s branchStmtWrapper) Pos() astPosition { return s.s.Pos() }
func (s branchStmtWrapper) End() astPosition { return s.s.End() }
func (s branchStmtWrapper) Unwrap() ast.Node { return s.s }

type blockStmtWrapper struct{ s *ast.BlockStmt }

func (blockStmtWrapper) aStmt()                {}
func (blockStmtWrapper) aBlockStmt()           {}
func (s blockStmtWrapper) List() astStmtList   { return stmtListWrapper(s.s.List) }
func (s blockStmtWrapper) Lbrace() astPosition { return s.s.Lbrace }
func (s blockStmtWrapper) Rbrace() astPosition { return s.s.Rbrace }
func (s blockStmtWrapper) Pos() astPosition    { return s.s.Pos() }
func (s blockStmtWrapper) End() astPosition    { return s.s.End() }
func (s blockStmtWrapper) Unwrap() ast.Node    { return s.s }

func wrapBlockStmt(s *ast.BlockStmt) astBlockStmt {
	if s == nil {
		return nil
	}
	return blockStmtWrapper{s: s}
}

type ifStmtWrapper struct{ s *ast.IfStmt }

func (ifStmtWrapper) aStmt()               {}
func (ifStmtWrapper) aIfStmt()             {}
func (s ifStmtWrapper) Init() astStmt      { return wrapStmt(s.s.Init) }
func (s ifStmtWrapper) Cond() astExpr      { return wrapExpr(s.s.Cond) }
func (s ifStmtWrapper) Body() astBlockStmt { return wrapBlockStmt(s.s.Body) }
func (s ifStmtWrapper) Else() astStmt      { return wrapStmt(s.s.Else) }

func (s ifStmtWrapper) Pos() astPosition { return s.s.Pos() }
func (s ifStmtWrapper) End() astPosition { return s.s.End() }
func (s ifStmtWrapper) Unwrap() ast.Node { return s.s }

type caseClauseWrapper struct{ s *ast.CaseClause }

func (caseClauseWrapper) aStmt()               {}
func (caseClauseWrapper) aCaseClause()         {}
func (s caseClauseWrapper) ListLen() int       { return len(s.s.List) }
func (s caseClauseWrapper) Item(i int) astExpr { return wrapExpr(s.s.List[i]) }
func (s caseClauseWrapper) Body() astStmtList  { return stmtListWrapper(s.s.Body) }
func (s caseClauseWrapper) Pos() astPosition   { return s.s.Pos() }
func (s caseClauseWrapper) End() astPosition   { return s.s.End() }
func (s caseClauseWrapper) Unwrap() ast.Node   { return s.s }

type switchStmtWrapper struct{ s *ast.SwitchStmt }

func (switchStmtWrapper) aStmt()               {}
func (switchStmtWrapper) aSwitchStmt()         {}
func (s switchStmtWrapper) Init() astStmt      { return wrapStmt(s.s.Init) }
func (s switchStmtWrapper) Tag() astExpr       { return wrapExpr(s.s.Tag) }
func (s switchStmtWrapper) Body() astBlockStmt { return wrapBlockStmt(s.s.Body) }
func (s switchStmtWrapper) Pos() astPosition   { return s.s.Pos() }
func (s switchStmtWrapper) End() astPosition   { return s.s.End() }
func (s switchStmtWrapper) Unwrap() ast.Node   { return s.s }

type typeSwitchStmtWrapper struct{ s *ast.TypeSwitchStmt }

func (typeSwitchStmtWrapper) aStmt()               {}
func (typeSwitchStmtWrapper) aTypeSwitchStmt()     {}
func (s typeSwitchStmtWrapper) Init() astStmt      { return wrapStmt(s.s.Init) }
func (s typeSwitchStmtWrapper) Assign() astStmt    { return wrapStmt(s.s.Assign) }
func (s typeSwitchStmtWrapper) Body() astBlockStmt { return wrapBlockStmt(s.s.Body) }

func (s typeSwitchStmtWrapper) Pos() astPosition { return s.s.Pos() }
func (s typeSwitchStmtWrapper) End() astPosition { return s.s.End() }
func (s typeSwitchStmtWrapper) Unwrap() ast.Node { return s.s }

type commClauseWrapper struct{ s *ast.CommClause }

func (commClauseWrapper) aStmt()              {}
func (commClauseWrapper) aCommClause()        {}
func (s commClauseWrapper) Comm() astStmt     { return wrapStmt(s.s.Comm) }
func (s commClauseWrapper) Body() astStmtList { return stmtListWrapper(s.s.Body) }
func (s commClauseWrapper) Pos() astPosition  { return s.s.Pos() }
func (s commClauseWrapper) End() astPosition  { return s.s.End() }
func (s commClauseWrapper) Unwrap() ast.Node  { return s.s }

type selectStmtWrapper struct{ s *ast.SelectStmt }

func (selectStmtWrapper) aStmt()               {}
func (selectStmtWrapper) aSelectStmt()         {}
func (s selectStmtWrapper) Body() astBlockStmt { return wrapBlockStmt(s.s.Body) }
func (s selectStmtWrapper) Pos() astPosition   { return s.s.Pos() }
func (s selectStmtWrapper) End() astPosition   { return s.s.End() }
func (s selectStmtWrapper) Unwrap() ast.Node   { return s.s }

type forStmtWrapper struct{ s *ast.ForStmt }

func (forStmtWrapper) aStmt()               {}
func (forStmtWrapper) aForStmt()            {}
func (s forStmtWrapper) Init() astStmt      { return wrapStmt(s.s.Init) }
func (s forStmtWrapper) Cond() astExpr      { return wrapExpr(s.s.Cond) }
func (s forStmtWrapper) Post() astStmt      { return wrapStmt(s.s.Post) }
func (s forStmtWrapper) Body() astBlockStmt { return wrapBlockStmt(s.s.Body) }

func (s forStmtWrapper) Pos() astPosition { return s.s.Pos() }
func (s forStmtWrapper) End() astPosition { return s.s.End() }
func (s forStmtWrapper) Unwrap() ast.Node { return s.s }

type rangeStmtWrapper struct{ s *ast.RangeStmt }

func (rangeStmtWrapper) aStmt()                {}
func (rangeStmtWrapper) aRangeStmt()           {}
func (s rangeStmtWrapper) Key() astExpr        { return wrapExpr(s.s.Key) }
func (s rangeStmtWrapper) Value() astExpr      { return wrapExpr(s.s.Value) }
func (s rangeStmtWrapper) X() astExpr          { return wrapExpr(s.s.X) }
func (s rangeStmtWrapper) Body() astBlockStmt  { return wrapBlockStmt(s.s.Body) }
func (s rangeStmtWrapper) Tok() token.Token    { return s.s.Tok }
func (s rangeStmtWrapper) TokPos() astPosition { return s.s.TokPos }
func (s rangeStmtWrapper) Pos() astPosition    { return s.s.Pos() }
func (s rangeStmtWrapper) End() astPosition    { return s.s.End() }
func (s rangeStmtWrapper) Unwrap() ast.Node    { return s.s }

type stmtListWrapper []ast.Stmt

func (w stmtListWrapper) Len() int               { return len(w) }
func (w stmtListWrapper) Stmt(i int) astStmt     { return wrapStmt(w[i]) }
func (w stmtListWrapper) Head(i int) astStmtList { return stmtListWrapper(w[:i]) }

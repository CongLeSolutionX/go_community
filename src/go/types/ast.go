// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"go/ast"
	"go/token"
)

type astNode interface {
	Pos() astPosition
	End() astPosition

	// Unwrap returns the underlying node in the ast package. This is for
	// convenience: it wouldn't be part of the internal types interfaces.
	Unwrap() ast.Node
}

type astFile interface {
	astNode
	Package() astPosition
	Name() astIdent
	Decls() astDeclList
}

type astDeclList interface {
	Len() int
	Decl(i int) astDecl
}

type astPosition interface {
	IsValid() bool
}

type declKind int

const (
	nilDeclKind declKind = iota
	badDeclKind
	genDeclKind
	funcDeclKind

	// In a future iteration, it makes sense to split up GenDecl into the
	// following:
	/*
		importDeclKind
		constDeclKind
		typeDeclKind
		varDeclKind
	*/
)

type astDecl interface {
	astNode

	Kind() declKind

	BadDecl() astBadDecl
	GenDecl() astGenDecl
	FuncDecl() astFuncDecl
}

type astBadDecl interface {
	astDecl
}

type astGenDecl interface {
	astDecl
	Tok() token.Token
	Specs() astSpecList
}

type astSpec interface {
	astNode

	Kind() specKind

	ImportSpec() astImportSpec
	ValueSpec() astValueSpec
	TypeSpec() astTypeSpec
}

type specKind int

const (
	NilSpecKind specKind = iota
	ImportSpecKind
	ValueSpecKind
	TypeSpecKind
)

type astValueSpec interface {
	astSpec
	Names() astIdentList
	Type() astExpr
	Values() astExprList
}

type astTypeSpec interface {
	astSpec
	Name() astIdent
	Assign() astPosition
	Type() astExpr
}

type astImportSpec interface {
	astSpec
	Path() astBasicLit
	Name() astIdent
}

type astSpecList interface {
	Len() int
	Spec(i int) astSpec
}

type astFuncDecl interface {
	astDecl
	Recv() astFieldList
	Name() astIdent
	Type() astFuncType
	Body() astBlockStmt
}

type expressionKind int

const (
	nilExprKind expressionKind = iota
	badExprKind
	identKind
	ellipsisKind
	basicLitKind
	funcLitKind
	compositeLitKind
	parenExprKind
	selectorExprKind
	indexExprKind
	sliceExprKind
	typeAssertExprKind
	callExprKind
	starExprKind
	unaryExprKind
	binaryExprKind
	keyValueExprKind
	arrayTypeKind
	structTypeKind
	funcTypeKind
	interfaceTypeKind
	mapTypeKind
	chanTypeKind
)

type astExpr interface {
	astNode

	// Reify returns the concrete wrapper underlying a possible lazy
	// wrapper. This is necessary in situations where we rely on ast
	// comparison (for example, when using expressions as map keys).
	Reify() astExpr

	Kind() expressionKind

	BadExpr() astBadExpr
	Ident() astIdent
	Ellipsis() astEllipsis
	BasicLit() astBasicLit
	FuncLit() astFuncLit
	CompositeLit() astCompositeLit
	ParenExpr() astParenExpr
	SelectorExpr() astSelectorExpr
	IndexExpr() astIndexExpr
	SliceExpr() astSliceExpr
	TypeAssertExpr() astTypeAssertExpr
	CallExpr() astCallExpr
	StarExpr() astStarExpr
	UnaryExpr() astUnaryExpr
	BinaryExpr() astBinaryExpr
	KeyValueExpr() astKeyValueExpr
	ArrayType() astArrayType
	StructType() astStructType
	FuncType() astFuncType
	InterfaceType() astInterfaceType
	MapType() astMapType
	ChanType() astChanType
}

type astIdent interface {
	astExpr
	Name() string
}

type astSelectorExpr interface {
	astExpr
	X() astExpr
	Sel() astIdent
}

type astBadExpr interface {
	astExpr
}

type astEllipsis interface {
	astExpr
	Elt() astExpr
}

type astBasicLit interface {
	astExpr
	LitKind() token.Token
	Value() string
}

type astFuncLit interface {
	astExpr
	Type() astExpr
	Body() astBlockStmt
}

type astCompositeLit interface {
	astExpr
	Type() astExpr
	Elts() astExprList
	Rbrace() astPosition
}

type astParenExpr interface {
	astExpr
	X() astExpr
}

type astIndexExpr interface {
	astExpr
	X() astExpr
	Index() astExpr
}

type astSliceExpr interface {
	astExpr
	X() astExpr
	Low() astExpr
	High() astExpr
	Max() astExpr
	Slice3() bool
	Rbrack() astPosition
}

type astTypeAssertExpr interface {
	astExpr
	X() astExpr
	Type() astExpr
}

type astCallExpr interface {
	astExpr
	Args() astExprList
	Fun() astExpr
	EllipsisPos() astPosition
	Rparen() astPosition
}

type astStarExpr interface {
	astExpr

	X() astExpr
}

type astUnaryExpr interface {
	astExpr
	Op() token.Token
	X() astExpr
}

type astBinaryExpr interface {
	astExpr
	Op() token.Token
	X() astExpr
	Y() astExpr
}

type astKeyValueExpr interface {
	astExpr
	Key() astExpr
	Value() astExpr
	Colon() astPosition
}

type astArrayType interface {
	astExpr
	Len() astExpr
	Elt() astExpr
}

type astStructType interface {
	astExpr
	Fields() astFieldList
}

type astFieldList interface {
	astNode
	Len() int
	Field(i int) astField
}

type astField interface {
	astNode
	Names() astIdentList
	Type() astExpr
	Tag() astBasicLit
}

type astExprList interface {
	Len() int
	Expr(i int) astExpr
}

type astIdentList interface {
	Len() int
	Ident(i int) astIdent
}

type astFuncType interface {
	astExpr
	Params() astFieldList
	Results() astFieldList
}

type astInterfaceType interface {
	astExpr
	Methods() astFieldList
}

type astMapType interface {
	astExpr
	Key() astExpr
	Value() astExpr
}

type astChanType interface {
	astExpr
	// TODO: replace this return type
	Dir() ast.ChanDir
	Value() astExpr
}

type stmtKind int

const (
	nilStmtKind stmtKind = iota
	badStmtKind
	declStmtKind
	emptyStmtKind
	labeledStmtKind
	exprStmtKind
	sendStmtKind
	incDecStmtKind
	assignStmtKind
	goStmtKind
	deferStmtKind
	returnStmtKind
	branchStmtKind
	blockStmtKind
	ifStmtKind
	caseClauseKind
	switchStmtKind
	typeSwitchStmtKind
	commClauseKind
	selectStmtKind
	forStmtKind
	rangeStmtKind
)

type astStmt interface {
	astNode

	Kind() stmtKind
	BadStmt() astBadStmt
	DeclStmt() astDeclStmt
	EmptyStmt() astEmptyStmt
	LabeledStmt() astLabeledStmt
	ExprStmt() astExprStmt
	SendStmt() astSendStmt
	IncDecStmt() astIncDecStmt
	AssignStmt() astAssignStmt
	GoStmt() astGoStmt
	DeferStmt() astDeferStmt
	ReturnStmt() astReturnStmt
	BranchStmt() astBranchStmt
	BlockStmt() astBlockStmt
	IfStmt() astIfStmt
	CaseClause() astCaseClause
	SwitchStmt() astSwitchStmt
	TypeSwitchStmt() astTypeSwitchStmt
	CommClause() astCommClause
	SelectStmt() astSelectStmt
	ForStmt() astForStmt
	RangeStmt() astRangeStmt
}

type astBadStmt interface {
	astStmt
}

type astDeclStmt interface {
	astStmt
	Decl() astDecl
}

type astEmptyStmt interface {
	astStmt
}

type astLabeledStmt interface {
	astStmt
	Label() astIdent
	Stmt() astStmt
}

type astExprStmt interface {
	astStmt
	X() astExpr
}

type astSendStmt interface {
	astStmt
	Chan() astExpr
	Value() astExpr
	Arrow() astPosition
}

type astIncDecStmt interface {
	astStmt
	Tok() token.Token
	TokPos() astPosition
	X() astExpr
}

type astAssignStmt interface {
	astStmt
	Lhs() astExprList
	Rhs() astExprList
	Tok() token.Token
	TokPos() astPosition
}

type astGoStmt interface {
	astStmt
	Call() astCallExpr
}

type astDeferStmt interface {
	astStmt
	Call() astCallExpr
}

type astReturnStmt interface {
	astStmt
	Results() astExprList
	Return() astPosition
}

type astBranchStmt interface {
	astStmt
	Tok() token.Token
	Label() astIdent
}

type astBlockStmt interface {
	astStmt
	List() astStmtList
	Lbrace() astPosition
	Rbrace() astPosition
}

type astStmtList interface {
	Len() int
	Stmt(i int) astStmt
	Head(i int) astStmtList // [:i]
}

type astIfStmt interface {
	astStmt
	Init() astStmt
	Cond() astExpr
	Body() astBlockStmt
	Else() astStmt
}

type astCaseClause interface {
	astStmt
	List() astExprList
	Body() astStmtList
}

type astSwitchStmt interface {
	astStmt
	Init() astStmt
	Tag() astExpr
	Body() astBlockStmt
}

type astTypeSwitchStmt interface {
	astStmt
	Init() astStmt
	Assign() astStmt
	Body() astBlockStmt
}

type astCommClause interface {
	astStmt
	Comm() astStmt
	Body() astStmtList
}

type astSelectStmt interface {
	astStmt
	Body() astBlockStmt
}

type astForStmt interface {
	astStmt
	Init() astStmt
	Cond() astExpr
	Post() astStmt
	Body() astBlockStmt
}

type astRangeStmt interface {
	astStmt
	Key() astExpr
	Value() astExpr
	X() astExpr
	Body() astBlockStmt
	Tok() token.Token
	TokPos() astPosition
}

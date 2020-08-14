// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package types implements the go type checker.
package types

import (
	"go/ast"
	"go/token"
)

type Pos interface {
	IsValid() bool
}

// Nodes
type (
	Node interface {
		Pos() Pos
		End() Pos
	}

	File interface {
		Node
		Package() Pos
		Name() Ident
		DeclsLen() int
		Decl(i int) Decl
	}

	FieldList interface {
		Node
		Len() int
		Field(i int) Field
	}

	Field interface {
		Node
		NamesLen() int
		Name(i int) Ident
		Type() Expr
		Tag() BasicLit
	}
)

// Decls
type (
	Decl interface {
		Node
		ADecl()
	}

	BadDecl interface {
		Decl
		ABadDecl()
	}

	GenDecl interface {
		Decl
		AGenDecl()

		Tok() token.Token
		SpecsLen() int
		Spec(i int) Spec
	}

	FuncDecl interface {
		Decl
		AFuncDecl()

		Recv() FieldList
		Name() Ident
		Type() FuncType
		Body() BlockStmt
	}
)

// Specs
type (
	Spec interface {
		Node
		ASpec()
	}

	ValueSpec interface {
		Spec
		AValueSpec()
		NamesLen() int
		Name(i int) Ident
		Type() Expr
		ValuesLen() int
		Value(i int) Expr
	}

	TypeSpec interface {
		Spec
		ATypeSpec()
		Name() Ident
		Assign() Pos
		Type() Expr
	}

	ImportSpec interface {
		Spec
		AnImportSpec()
		Path() BasicLit
		Name() Ident
	}
)

// Exprs
type (
	Expr interface {
		Node
		AnExpr()
	}

	Ident interface {
		Expr
		AnIdent()

		Name() string
	}

	SelectorExpr interface {
		Expr
		ASelectorExpr()

		X() Expr
		Sel() Ident
	}

	BadExpr interface {
		Expr
		ABadExpr()
	}

	Ellipsis interface {
		Expr
		AnEllipsis()

		Elt() Expr
	}

	BasicLit interface {
		Expr
		ABasicLit()

		Kind() token.Token
		Value() string
	}

	FuncLit interface {
		Expr
		AFuncLit()

		Type() Expr
		Body() BlockStmt
	}

	CompositeLit interface {
		Expr
		ACompositeLit()

		Type() Expr
		EltsLen() int
		Elt(i int) Expr
		Rbrace() Pos
	}

	ParenExpr interface {
		Expr
		AParenExpr()

		X() Expr
	}

	IndexExpr interface {
		Expr
		AnIndexExpr()

		X() Expr
		Index() Expr
	}

	SliceExpr interface {
		Expr
		ASliceExpr()

		X() Expr
		Low() Expr
		High() Expr
		Max() Expr
		Slice3() bool
		Rbrack() Pos
	}

	TypeAssertExpr interface {
		Expr
		ATypeAssertExpr()

		X() Expr
		Type() Expr
	}

	CallExpr interface {
		Expr
		ACallExpr()

		ArgsLen() int
		Arg(i int) Expr
		Fun() Expr
		Ellipsis() Pos
		Rparen() Pos
	}

	StarExpr interface {
		Expr
		AStarExpr()

		X() Expr
	}

	UnaryExpr interface {
		Expr
		AUnaryExpr()

		Op() token.Token
		X() Expr
	}

	BinaryExpr interface {
		Expr
		ABinaryExpr()

		Op() token.Token
		X() Expr
		Y() Expr
	}

	KeyValueExpr interface {
		Expr
		AKeyValueExpr()

		Key() Expr
		Value() Expr
	}

	ArrayType interface {
		Expr
		AnArrayType()

		Len() Expr
		Elt() Expr
	}

	StructType interface {
		Expr
		AStructType()

		Fields() FieldList
	}

	FuncType interface {
		Expr
		AFuncType()

		Params() FieldList
		Results() FieldList
	}

	InterfaceType interface {
		Expr
		AnInterfaceType()

		Methods() FieldList
	}

	MapType interface {
		Expr
		AMapType()

		Key() Expr
		Value() Expr
	}

	ChanType interface {
		Expr
		AChanType()

		// TODO: replace this return type
		Dir() ast.ChanDir
		Value() Expr
	}
)

type ExprList interface {
	Len() int
	Expr(i int) Expr
}

// Stmts
type (
	Stmt interface {
		Node
		AStmt()
	}

	BadStmt interface {
		Stmt
		ABadStmt()
	}

	DeclStmt interface {
		Stmt
		ADeclStmt()

		Decl() Decl
	}

	EmptyStmt interface {
		Stmt
		AnEmptyStmt()
	}

	LabeledStmt interface {
		Stmt
		ALabeledStmt()

		Label() Ident
		Stmt() Stmt
	}

	ExprStmt interface {
		Stmt
		AnExprStmt()

		X() Expr
	}

	SendStmt interface {
		Stmt
		ASendStmt()

		Chan() Expr
		Value() Expr
		Arrow() Pos
	}

	IncDecStmt interface {
		Stmt
		AnIncDecStmt()

		Tok() token.Token
		TokPos() Pos
		X() Expr
	}

	AssignStmt interface {
		Stmt
		AnAssignStmt()

		Lhs() ExprList
		LhsLen() int
		LhsExpr(i int) Expr
		RhsLen() int
		RhsExpr(i int) Expr
		Rhs() ExprList
		Tok() token.Token
		TokPos() Pos
	}

	GoStmt interface {
		Stmt
		AGoStmt()

		Call() CallExpr
	}

	DeferStmt interface {
		Stmt
		ADeferStmt()

		Call() CallExpr
	}

	ReturnStmt interface {
		Stmt
		AReturnStmt()

		Results() ExprList
		ResultsLen() int
		Result(i int) Expr
		Return() Pos
	}

	BranchStmt interface {
		Stmt
		ABranchStmt()

		Tok() token.Token
		Label() Ident
	}

	BlockStmt interface {
		Stmt
		ABlockStmt()

		List() StmtList
		Lbrace() Pos
		Rbrace() Pos
	}

	IfStmt interface {
		Stmt
		AnIfStmt()

		Init() Stmt
		Cond() Expr
		Body() BlockStmt
		Else() Stmt
	}

	CaseClause interface {
		Stmt
		ACaseClause()

		ListLen() int
		Item(i int) Expr
		Body() StmtList
	}

	SwitchStmt interface {
		Stmt
		ASwitchStmt()

		Init() Stmt
		Tag() Expr
		Body() BlockStmt
	}

	TypeSwitchStmt interface {
		Stmt
		ATypeSwitchStmt()

		Init() Stmt
		Assign() Stmt
		Body() BlockStmt
	}

	CommClause interface {
		Stmt
		ACommClause()

		Comm() Stmt
		Body() StmtList
	}

	SelectStmt interface {
		Stmt
		ASelectStmt()

		Body() BlockStmt
	}

	ForStmt interface {
		Stmt
		AForStmt()

		Init() Stmt
		Cond() Expr
		Post() Stmt
		Body() BlockStmt
	}

	RangeStmt interface {
		Stmt
		ARangeStmt()

		Key() Expr
		Value() Expr
		X() Expr
		Body() BlockStmt
		Tok() token.Token
		TokPos() Pos
	}
)

type StmtList interface {
	Len() int
	Stmt(i int) Stmt
}

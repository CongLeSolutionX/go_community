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
	DeclsLen() int
	Decl(i int) astDecl
}

type astPosition interface {
	IsValid() bool
}

// Decls
type (
	astDecl interface {
		astNode
		aDecl()
	}

	astBadDecl interface {
		astDecl
		aBadDecl()
	}

	astGenDecl interface {
		astDecl
		aGenDecl()

		Tok() token.Token
		SpecsLen() int
		Spec(i int) astSpec
	}

	astFuncDecl interface {
		astDecl
		aFuncDecl()

		Recv() astFieldList
		Name() astIdent
		Type() astFuncType
		Body() astBlockStmt
	}
)

// Specs
type (
	astSpec interface {
		astNode
		aSpec()
	}

	astValueSpec interface {
		astSpec
		aValueSpec()
		NamesLen() int
		Name(i int) astIdent
		Type() astExpr
		ValuesLen() int
		Value(i int) astExpr
	}

	astTypeSpec interface {
		astSpec
		aTypeSpec()
		Name() astIdent
		Assign() astPosition
		Type() astExpr
	}

	astImportSpec interface {
		astSpec
		anImportSpec()
		Path() astBasicLit
		Name() astIdent
	}
)

// Exprs
type (
	astExpr interface {
		astNode
		anExpr()
	}

	astIdent interface {
		astExpr
		IdentName() string
		anIdent()
	}

	astSelectorExpr interface {
		astExpr
		X() astExpr
		Sel() astIdent
		anSelectorExpr()
	}

	astBadExpr interface {
		astExpr
		anBadExpr()
	}

	astEllipsis interface {
		astExpr
		Elt() astExpr
		anEllipsis()
	}

	astBasicLit interface {
		astExpr
		LitKind() token.Token
		LitValue() string
		anBasicLit()
	}

	astFuncLit interface {
		astExpr
		Type() astExpr
		Body() astBlockStmt
		anFuncLit()
	}

	astCompositeLit interface {
		astExpr
		Type() astExpr
		EltsLen() int
		Elt(i int) astExpr
		Rbrace() astPosition
		anCompositeLit()
	}

	astParenExpr interface {
		astExpr
		X() astExpr
		anParenExpr()
	}

	astIndexExpr interface {
		astExpr
		X() astExpr
		Index() astExpr
		anIndexExpr()
	}

	astSliceExpr interface {
		astExpr
		X() astExpr
		Low() astExpr
		High() astExpr
		Max() astExpr
		Slice3() bool
		Rbrack() astPosition
		anSliceExpr()
	}

	astTypeAssertExpr interface {
		astExpr
		X() astExpr
		Type() astExpr
		anTypeAssertExpr()
	}

	astCallExpr interface {
		astExpr
		ArgsLen() int
		Arg(i int) astExpr
		FunExpr() astExpr
		EllipsisPos() astPosition
		RparenPos() astPosition
		anCallExpr()
	}

	astStarExpr interface {
		astExpr

		InnerX() astExpr
		anStarExpr()
	}

	astUnaryExpr interface {
		astExpr
		OpTok() token.Token
		InnerX() astExpr
		anUnaryExpr()
	}

	astBinaryExpr interface {
		astExpr
		OpTok() token.Token
		XExpr() astExpr
		YExpr() astExpr
		anBinaryExpr()
	}

	astKeyValueExpr interface {
		astExpr
		KeyExpr() astExpr
		ValueExpr() astExpr
		anKeyValueExpr()
	}

	astArrayType interface {
		astExpr
		Len() astExpr
		Elt() astExpr
		anArrayType()
	}

	astStructType interface {
		astExpr
		Fields() astFieldList
		anStructType()
	}
	astFuncType interface {
		astExpr
		Params() astFieldList
		Results() astFieldList
		anFuncType()
	}

	astInterfaceType interface {
		astExpr
		Methods() astFieldList
		anInterfaceType()
	}

	astMapType interface {
		astExpr
		Key() astExpr
		Value() astExpr
		anMapType()
	}

	astChanType interface {
		astExpr
		// TODO: replace this return type
		Dir() ast.ChanDir
		Value() astExpr
		anChanType()
	}
)

type astFieldList interface {
	astNode
	Len() int
	Field(i int) astField
}

type astField interface {
	astNode
	NamesLen() int
	Name(i int) astIdent
	Type() astExpr
	Tag() astBasicLit
}

type astExprList interface {
	Len() int
	Expr(i int) astExpr
}

// Stmts
type (
	astStmt interface {
		astNode
		aStmt()
	}

	astBadStmt interface {
		astStmt
		aBadStmt()
	}

	astDeclStmt interface {
		astStmt
		aDeclStmt()
		Decl() astDecl
	}

	astEmptyStmt interface {
		astStmt
		aEmptyStmt()
	}

	astLabeledStmt interface {
		astStmt
		aLabeledStmt()
		Label() astIdent
		Stmt() astStmt
	}

	astExprStmt interface {
		astStmt
		aExprStmt()
		X() astExpr
	}

	astSendStmt interface {
		astStmt
		aSendStmt()
		Chan() astExpr
		Value() astExpr
		Arrow() astPosition
	}

	astIncDecStmt interface {
		astStmt
		aIncDecStmt()
		Tok() token.Token
		TokPos() astPosition
		X() astExpr
	}

	astAssignStmt interface {
		astStmt
		aAssignStmt()
		Lhs() astExprList
		LhsLen() int
		LhsExpr(i int) astExpr
		RhsLen() int
		RhsExpr(i int) astExpr
		Rhs() astExprList
		Tok() token.Token
		TokPos() astPosition
	}

	astGoStmt interface {
		astStmt
		aGoStmt()
		Call() astCallExpr
	}

	astDeferStmt interface {
		astStmt
		aDeferStmt()
		Call() astCallExpr
	}

	astReturnStmt interface {
		astStmt
		aReturnStmt()
		Results() astExprList
		ResultsLen() int
		Result(i int) astExpr
		Return() astPosition
	}

	astBranchStmt interface {
		astStmt
		aBranchStmt()
		Tok() token.Token
		Label() astIdent
	}

	astBlockStmt interface {
		astStmt
		aBlockStmt()
		List() astStmtList
		Lbrace() astPosition
		Rbrace() astPosition
	}

	astIfStmt interface {
		astStmt
		aIfStmt()
		Init() astStmt
		Cond() astExpr
		Body() astBlockStmt
		Else() astStmt
	}

	astCaseClause interface {
		astStmt
		aCaseClause()
		ListLen() int
		Item(i int) astExpr
		Body() astStmtList
	}

	astSwitchStmt interface {
		astStmt
		aSwitchStmt()
		Init() astStmt
		Tag() astExpr
		Body() astBlockStmt
	}

	astTypeSwitchStmt interface {
		astStmt
		aTypeSwitchStmt()
		Init() astStmt
		Assign() astStmt
		Body() astBlockStmt
	}

	astCommClause interface {
		astStmt
		aCommClause()
		Comm() astStmt
		Body() astStmtList
	}

	astSelectStmt interface {
		astStmt
		aSelectStmt()
		Body() astBlockStmt
	}

	astForStmt interface {
		astStmt
		aForStmt()
		Init() astStmt
		Cond() astExpr
		Post() astStmt
		Body() astBlockStmt
	}

	astRangeStmt interface {
		astStmt
		aRangeStmt()
		Key() astExpr
		Value() astExpr
		X() astExpr
		Body() astBlockStmt
		Tok() token.Token
		TokPos() astPosition
	}
)

type astStmtList interface {
	Len() int
	Stmt(i int) astStmt
}

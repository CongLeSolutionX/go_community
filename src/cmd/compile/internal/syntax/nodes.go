// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syntax

import "cmd/internal/src"

// ----------------------------------------------------------------------------
// Nodes

type Node interface {
	// Begin returns the position of the left-most character belonging
	// to the respective production in the source.
	Begin() src.Pos

	// End returns the position immediately following the right-most
	// character belonging to the respective production in the source.
	End() src.Pos

	// Pos() returns the position associated with the node as follows:
	// 1) The position of a node representing a terminal syntax production
	//    (Name, BasicLit, etc.) is the position of the respective production
	//    in the source (which is the same as what Begin() returns).
	// 2) The position of a node representing a non-terminal production
	//    (IndexExpr, IfStmt, etc.) is the position of a token uniquely
	//    associated with that production; usually the left-most one
	//    ('[' for IndexExpr, 'if' for IfStmt, etc.).
	// The following invariant is satisfied: Begin() <= Pos() <= End().
	Pos() src.Pos

	aNode()
}

type node struct {
	// commented out for now since not yet used
	// doc  *Comment // nil means no comment(s) attached
}

func (*node) aNode() {}

// ----------------------------------------------------------------------------
// Files

// package PkgName; DeclList[0], DeclList[1], ...
type File struct {
	Package  src.Pos
	PkgName  *Name
	DeclList []Decl
	node
}

// ----------------------------------------------------------------------------
// Declarations

type (
	Decl interface {
		Node
		aDecl()
	}

	//              Path
	// LocalPkgName Path
	ImportDecl struct {
		LocalPkgName *Name // including "."; nil means no rename present
		Path         *BasicLit
		Group        *Group // nil means not part of a group
		decl
	}

	// NameList
	// NameList      = Values
	// NameList Type = Values
	ConstDecl struct {
		NameList []*Name
		Type     Expr   // nil means no type
		Values   Expr   // nil means no values
		Group    *Group // nil means not part of a group
		decl
	}

	// Name Type
	TypeDecl struct {
		Name   *Name
		Alias  bool
		Type   Expr
		Group  *Group // nil means not part of a group
		Pragma Pragma
		decl
	}

	// NameList Type
	// NameList Type = Values
	// NameList      = Values
	VarDecl struct {
		NameList []*Name
		Type     Expr   // nil means no type
		Values   Expr   // nil means no values
		Group    *Group // nil means not part of a group
		decl
	}

	// func          Name Type { Body }
	// func          Name Type
	// func Receiver Name Type { Body }
	// func Receiver Name Type
	FuncDecl struct {
		Attr   map[string]bool // go:attr map
		Func   src.Pos
		Recv   *Field // nil means regular function
		Name   *Name
		Type   *FuncType
		Body   []Stmt // nil means no body (forward declaration)
		Pragma Pragma // TODO(mdempsky): Cleaner solution.
		Rbrace src.Pos
		decl
	}
)

type decl struct{ node }

func (*decl) aDecl() {}

// All declarations belonging to the same group point to the same Group node.
type Group struct {
	dummy int // not empty so we are guaranteed different Group instances
}

// ----------------------------------------------------------------------------
// Expressions

type (
	Expr interface {
		Node
		aExpr()
	}

	// Value
	Name struct {
		pos   src.Pos // Value position
		Value string
		expr
	}

	// Value
	BasicLit struct {
		pos   src.Pos // Value position
		Value string
		Kind  LitKind
		expr
	}

	// Type { ElemList[0], ElemList[1], ... }
	CompositeLit struct {
		Type     Expr // nil means no literal type
		Lbrace   src.Pos
		ElemList []Expr
		NKeys    int // number of elements with keys
		Rbrace   src.Pos
		expr
	}

	// Key: Value
	KeyValueExpr struct {
		Key   Expr
		Colon src.Pos
		Value Expr
		expr
	}

	// func Type { Body }
	FuncLit struct {
		Func   src.Pos
		Type   *FuncType
		Body   []Stmt
		Rbrace src.Pos
		expr
	}

	// (X)
	ParenExpr struct {
		Lparen src.Pos
		X      Expr
		Rparen src.Pos
		expr
	}

	// X.Sel
	SelectorExpr struct {
		X   Expr
		Dot src.Pos
		Sel *Name
		expr
	}

	// X[Index]
	IndexExpr struct {
		X      Expr
		Lbrack src.Pos
		Index  Expr
		Rbrack src.Pos
		expr
	}

	// X[Index[0] : Index[1] : Index[2]]
	SliceExpr struct {
		X      Expr
		Lbrack src.Pos
		Index  [3]Expr
		// Full indicates whether this is a simple or full slice expression.
		// In a valid AST, this is equivalent to Index[2] != nil.
		// TODO(mdempsky): This is only needed to report the "3-index
		// slice of string" error when Index[2] is missing.
		Full   bool
		Rbrack src.Pos
		expr
	}

	// X.(Type)
	AssertExpr struct {
		X   Expr
		Dot src.Pos
		// TODO(gri) consider using Name{"..."} instead of nil (permits attaching of comments)
		Type   Expr
		Rparen src.Pos
		expr
	}

	Operation struct {
		X   Expr
		pos src.Pos // operator position
		Op  Operator
		Y   Expr // Y == nil means unary expression
		expr
	}

	// Fun(ArgList[0], ArgList[1], ...)
	CallExpr struct {
		Fun     Expr
		Lparen  src.Pos
		ArgList []Expr
		HasDots bool // last argument is followed by ...
		Rparen  src.Pos
		expr
	}

	// ElemList[0], ElemList[1], ...
	ListExpr struct {
		ElemList []Expr
		expr
	}

	// [Len]Elem
	ArrayType struct {
		// TODO(gri) consider using Name{"..."} instead of nil (permits attaching of comments)
		Lbrack src.Pos
		Len    Expr // nil means Len is ...
		Elem   Expr
		expr
	}

	// []Elem
	SliceType struct {
		Lbrack src.Pos
		Elem   Expr
		expr
	}

	// ...Elem
	DotsType struct {
		Dots src.Pos
		Elem Expr
		expr
	}

	// struct { FieldList[0] TagList[0]; FieldList[1] TagList[1]; ... }
	StructType struct {
		Struct    src.Pos
		FieldList []*Field
		TagList   []*BasicLit // i >= len(TagList) || TagList[i] == nil means no tag for field i
		Rbrace    src.Pos
		expr
	}

	// Name Type
	//      Type
	Field struct {
		Name *Name // nil means anonymous field/parameter (structs/parameters), or embedded interface (interfaces)
		Type Expr  // field names declared in a list share the same Type (identical pointers)
		node
	}

	// interface { MethodList[0]; MethodList[1]; ... }
	InterfaceType struct {
		Interface  src.Pos
		MethodList []*Field
		Rbrace     src.Pos
		expr
	}

	FuncType struct {
		Func       src.Pos
		ParamList  []*Field
		ResultList []*Field
		expr
	}

	// map[Key]Value
	MapType struct {
		Map   src.Pos
		Key   Expr
		Value Expr
		expr
	}

	//   chan Elem
	// <-chan Elem
	// chan<- Elem
	ChanType struct {
		pos  src.Pos // chan or <- position
		Dir  ChanDir // 0 means no direction
		Elem Expr
		expr
	}
)

type expr struct{ node }

func (*expr) aExpr() {}

type ChanDir uint

const (
	_ ChanDir = iota
	SendOnly
	RecvOnly
)

// ----------------------------------------------------------------------------
// Statements

type (
	Stmt interface {
		Node
		aStmt()
	}

	SimpleStmt interface {
		Stmt
		aSimpleStmt()
	}

	EmptyStmt struct {
		pos src.Pos
		simpleStmt
	}

	LabeledStmt struct {
		Label *Name
		Colon src.Pos
		Stmt  Stmt
		stmt
	}

	BlockStmt struct {
		Lbrace src.Pos
		Body   []Stmt
		Rbrace src.Pos
		stmt
	}

	ExprStmt struct {
		X Expr
		simpleStmt
	}

	// Chan <- Value
	SendStmt struct {
		Chan  Expr
		Arrow src.Pos
		Value Expr
		simpleStmt
	}

	DeclStmt struct {
		pos      src.Pos // keyword position
		DeclList []Decl
		stmt
	}

	AssignStmt struct {
		Lhs Expr
		pos src.Pos  // operator position
		Op  Operator // 0 means no operation
		Rhs Expr     // Rhs == ImplicitOne means Lhs++ (Op == Add) or Lhs-- (Op == Sub)
		simpleStmt
	}

	BranchStmt struct {
		pos   src.Pos // token position
		Tok   token   // Break, Continue, Fallthrough, or Goto
		Label *Name   // nil means no explicit label
		stmt
	}

	CallStmt struct {
		pos  src.Pos // token position
		Tok  token   // Go or Defer
		Call *CallExpr
		stmt
	}

	ReturnStmt struct {
		Return  src.Pos
		Results Expr // nil means no explicit return values
		stmt
	}

	IfStmt struct {
		If     src.Pos
		Init   SimpleStmt
		Cond   Expr
		Then   []Stmt
		Rbrace src.Pos // position of Rbrace of Then part
		Else   Stmt    // either *IfStmt or *BlockStmt
		stmt
	}

	ForStmt struct {
		For    src.Pos
		Init   SimpleStmt // incl. *RangeClause
		Cond   Expr
		Post   SimpleStmt
		Body   []Stmt
		Rbrace src.Pos
		stmt
	}

	SwitchStmt struct {
		Switch src.Pos
		Init   SimpleStmt
		Tag    Expr
		Body   []*CaseClause
		Rbrace src.Pos
		stmt
	}

	SelectStmt struct {
		Select src.Pos
		Body   []*CommClause
		Rbrace src.Pos
		stmt
	}
)

type (
	RangeClause struct {
		Lhs   Expr // nil means no Lhs = or Lhs :=
		Def   bool // means :=
		Range src.Pos
		X     Expr // range X
		simpleStmt
	}

	TypeSwitchGuard struct {
		// TODO(gri) consider using Name{"..."} instead of nil (permits attaching of comments)
		Lhs *Name // nil means no Lhs :=
		Dot src.Pos
		X   Expr // X.(type)
		expr
	}

	CaseClause struct {
		Case  src.Pos
		Cases Expr // nil means default clause
		Colon src.Pos
		Body  []Stmt
		node
	}

	CommClause struct {
		Case  src.Pos
		Comm  SimpleStmt // send or receive stmt; nil means default clause
		Colon src.Pos
		Body  []Stmt
		node
	}
)

type stmt struct{ node }

func (stmt) aStmt() {}

type simpleStmt struct {
	stmt
}

func (simpleStmt) aSimpleStmt() {}

// ----------------------------------------------------------------------------
// Comments

// TODO(gri) Consider renaming to CommentPos, CommentPlacement, etc.
//           Kind = Above doesn't make much sense.
type CommentKind uint

const (
	Above CommentKind = iota
	Below
	Left
	Right
)

type Comment struct {
	Kind CommentKind
	Text string
	Next *Comment
}

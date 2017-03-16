// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file implements Begin(), End(), and Pos() for all nodes.

package syntax

import "cmd/internal/src"

func (n *File) Begin() src.Pos { return n.Package }
func (n *File) End() src.Pos {
	if i := len(n.DeclList); i > 0 {
		return n.DeclList[i-1].End()
	}
	return n.PkgName.End()
}
func (n *File) Pos() src.Pos { return n.Begin() }

func (n *ImportDecl) Begin() src.Pos {
	if n.LocalPkgName != nil {
		return n.LocalPkgName.Begin()
	}
	return n.Path.Begin()
}
func (n *ImportDecl) End() src.Pos { return n.Path.End() }
func (n *ImportDecl) Pos() src.Pos { return n.Begin() }

func (n *ConstDecl) Begin() src.Pos { return n.NameList[0].Begin() }
func (n *ConstDecl) End() src.Pos {
	if n.Values != nil {
		return n.Values.End()
	}
	return n.NameList[len(n.NameList)-1].End()
}
func (n *ConstDecl) Pos() src.Pos { return n.Begin() }

func (n *TypeDecl) Begin() src.Pos { return n.Name.Begin() }
func (n *TypeDecl) End() src.Pos   { return n.Type.End() }
func (n *TypeDecl) Pos() src.Pos   { return n.Begin() }

func (n *VarDecl) Begin() src.Pos { return n.NameList[0].Begin() }
func (n *VarDecl) End() src.Pos {
	if n.Values != nil {
		return n.Values.End()
	}
	return n.Type.End()
}
func (n *VarDecl) Pos() src.Pos { return n.Begin() }

func (n *FuncDecl) Begin() src.Pos { return n.Func }
func (n *FuncDecl) End() src.Pos   { return adj(n.Rbrace, len("}")) }
func (n *FuncDecl) Pos() src.Pos   { return n.Begin() }

func (n *Name) Begin() src.Pos { return n.pos }
func (n *Name) End() src.Pos   { return adj(n.pos, len(n.Value)) }
func (n *Name) Pos() src.Pos   { return n.Begin() }

func (n *BasicLit) Begin() src.Pos { return n.pos }
func (n *BasicLit) End() src.Pos   { return adj(n.pos, len(n.Value)) }
func (n *BasicLit) Pos() src.Pos   { return n.Begin() }

func (n *CompositeLit) Begin() src.Pos {
	if n.Type != nil {
		return n.Type.Begin()
	}
	return n.Lbrace
}
func (n *CompositeLit) End() src.Pos { return adj(n.Rbrace, len("}")) }
func (n *CompositeLit) Pos() src.Pos { return n.Lbrace }

func (n *KeyValueExpr) Begin() src.Pos { return n.Key.Begin() }
func (n *KeyValueExpr) End() src.Pos   { return n.Value.End() }
func (n *KeyValueExpr) Pos() src.Pos   { return n.Colon }

func (n *FuncLit) Begin() src.Pos { return n.Func }
func (n *FuncLit) End() src.Pos   { return adj(n.Rbrace, len("}")) }
func (n *FuncLit) Pos() src.Pos   { return n.Begin() }

func (n *ParenExpr) Begin() src.Pos { return n.Lparen }
func (n *ParenExpr) End() src.Pos   { return adj(n.Rparen, len(")")) }
func (n *ParenExpr) Pos() src.Pos   { return n.Begin() }

func (n *SelectorExpr) Begin() src.Pos { return n.X.Begin() }
func (n *SelectorExpr) End() src.Pos   { return n.Sel.End() }
func (n *SelectorExpr) Pos() src.Pos   { return n.Dot }

func (n *IndexExpr) Begin() src.Pos { return n.X.Begin() }
func (n *IndexExpr) End() src.Pos   { return adj(n.Rbrack, len("]")) }
func (n *IndexExpr) Pos() src.Pos   { return n.Lbrack }

func (n *SliceExpr) Begin() src.Pos { return n.X.Begin() }
func (n *SliceExpr) End() src.Pos   { return adj(n.Rbrack, len("]")) }
func (n *SliceExpr) Pos() src.Pos   { return n.Lbrack }

func (n *AssertExpr) Begin() src.Pos { return n.X.Begin() }
func (n *AssertExpr) End() src.Pos   { return adj(n.Rparen, len(")")) }
func (n *AssertExpr) Pos() src.Pos   { return n.Dot }

func (n *Operation) Begin() src.Pos {
	if n.Y != nil {
		return n.X.Begin()
	}
	return n.pos
}
func (n *Operation) End() src.Pos {
	if n.Y != nil {
		return n.Y.End()
	}
	return n.X.End()
}
func (n *Operation) Pos() src.Pos { return n.pos }

func (n *CallExpr) Begin() src.Pos { return n.Fun.Begin() }
func (n *CallExpr) End() src.Pos   { return adj(n.Rparen, len(")")) }
func (n *CallExpr) Pos() src.Pos   { return n.Lparen }

func (n *ListExpr) Begin() src.Pos { return n.ElemList[0].Begin() }
func (n *ListExpr) End() src.Pos   { return n.ElemList[len(n.ElemList)-1].End() }
func (n *ListExpr) Pos() src.Pos   { return n.Begin() }

func (n *ArrayType) Begin() src.Pos { return n.Lbrack }
func (n *ArrayType) End() src.Pos   { return n.Elem.End() }
func (n *ArrayType) Pos() src.Pos   { return n.Begin() }

func (n *SliceType) Begin() src.Pos { return n.Lbrack }
func (n *SliceType) End() src.Pos   { return n.Elem.End() }
func (n *SliceType) Pos() src.Pos   { return n.Begin() }

func (n *DotsType) Begin() src.Pos { return n.Dots }
func (n *DotsType) End() src.Pos   { return n.Elem.End() }
func (n *DotsType) Pos() src.Pos   { return n.Begin() }

func (n *StructType) Begin() src.Pos { return n.Struct }
func (n *StructType) End() src.Pos   { return adj(n.Rbrace, len("}")) }
func (n *StructType) Pos() src.Pos   { return n.Begin() }

func (n *Field) Begin() src.Pos {
	if n.Name != nil {
		return n.Name.Begin()
	}
	return n.Type.Begin()
}
func (n *Field) End() src.Pos {
	// a Field comprises the field name only
	// (we can't do better because multiple field names using a single
	// type are represented by multiple separate Field nodes in the
	// syntax tree)
	if n.Name != nil {
		return n.Name.End()
	}
	return n.Type.End()
}
func (n *Field) Pos() src.Pos { return n.Begin() }

func (n *InterfaceType) Begin() src.Pos { return n.Interface }
func (n *InterfaceType) End() src.Pos   { return adj(n.Rbrace, len("}")) }
func (n *InterfaceType) Pos() src.Pos   { return n.Begin() }

func (n *FuncType) Begin() src.Pos { return n.Func }
func (n *FuncType) End() src.Pos {
	// TODO(gri) need to track end correctly instead of this heuristic
	switch len(n.ResultList) {
	case 0:
		if len(n.ParamList) > 0 {
			return adj(n.ParamList[len(n.ParamList)-1].End(), len(")"))
		}
		return adj(n.Func, len("func ()"))
	case 1:
		return n.ResultList[0].End()
	default:
		return adj(n.ResultList[len(n.ResultList)-1].End(), len(")"))
	}
}
func (n *FuncType) Pos() src.Pos { return n.Begin() }

func (n *MapType) Begin() src.Pos { return n.Map }
func (n *MapType) End() src.Pos   { return n.Value.End() }
func (n *MapType) Pos() src.Pos   { return n.Begin() }

func (n *ChanType) Begin() src.Pos { return n.pos }
func (n *ChanType) End() src.Pos   { return n.Elem.End() }
func (n *ChanType) Pos() src.Pos   { return n.Begin() }

func (n *EmptyStmt) Begin() src.Pos { return n.pos }
func (n *EmptyStmt) End() src.Pos   { return n.pos }
func (n *EmptyStmt) Pos() src.Pos   { return n.Begin() }

func (n *LabeledStmt) Begin() src.Pos { return n.Label.Begin() }
func (n *LabeledStmt) End() src.Pos   { return n.Stmt.End() }
func (n *LabeledStmt) Pos() src.Pos   { return n.Colon }

func (n *BlockStmt) Begin() src.Pos { return n.Lbrace }
func (n *BlockStmt) End() src.Pos   { return adj(n.Rbrace, len("}")) }
func (n *BlockStmt) Pos() src.Pos   { return n.Begin() }

func (n *ExprStmt) Begin() src.Pos { return n.X.Begin() }
func (n *ExprStmt) End() src.Pos   { return n.X.End() }
func (n *ExprStmt) Pos() src.Pos   { return n.Begin() }

func (n *SendStmt) Begin() src.Pos { return n.Chan.Begin() }
func (n *SendStmt) End() src.Pos   { return n.Value.End() }
func (n *SendStmt) Pos() src.Pos   { return n.Arrow }

func (n *DeclStmt) Begin() src.Pos { return n.pos }
func (n *DeclStmt) End() src.Pos   { return n.DeclList[len(n.DeclList)-1].End() }
func (n *DeclStmt) Pos() src.Pos   { return n.Begin() }

func (n *AssignStmt) Begin() src.Pos { return n.Lhs.Begin() }
func (n *AssignStmt) End() src.Pos {
	if n.Rhs == ImplicitOne {
		return adj(n.Lhs.End(), len("++"))
	}
	return n.Rhs.End()
}
func (n *AssignStmt) Pos() src.Pos { return n.pos }

func (n *BranchStmt) Begin() src.Pos { return n.pos }
func (n *BranchStmt) End() src.Pos {
	if n.Label != nil {
		return n.Label.End()
	}
	return adj(n.pos, len(n.Tok.String()))
}
func (n *BranchStmt) Pos() src.Pos { return n.Begin() }

func (n *CallStmt) Begin() src.Pos { return n.pos }
func (n *CallStmt) End() src.Pos   { return n.Call.End() }
func (n *CallStmt) Pos() src.Pos   { return n.Begin() }

func (n *ReturnStmt) Begin() src.Pos { return n.Return }
func (n *ReturnStmt) End() src.Pos {
	if n.Results != nil {
		return n.Results.End()
	}
	return adj(n.Return, len("return"))
}
func (n *ReturnStmt) Pos() src.Pos { return n.Begin() }

func (n *IfStmt) Begin() src.Pos { return n.If }
func (n *IfStmt) End() src.Pos {
	if n.Else != nil {
		return n.Else.End()
	}
	return adj(n.Rbrace, len("}"))
}
func (n *IfStmt) Pos() src.Pos { return n.Begin() }

func (n *ForStmt) Begin() src.Pos { return n.For }
func (n *ForStmt) End() src.Pos   { return adj(n.Rbrace, len("}")) }
func (n *ForStmt) Pos() src.Pos   { return n.Begin() }

func (n *SwitchStmt) Begin() src.Pos { return n.Switch }
func (n *SwitchStmt) End() src.Pos   { return adj(n.Rbrace, len("}")) }
func (n *SwitchStmt) Pos() src.Pos   { return n.Begin() }

func (n *SelectStmt) Begin() src.Pos { return n.Select }
func (n *SelectStmt) End() src.Pos   { return adj(n.Rbrace, len("}")) }
func (n *SelectStmt) Pos() src.Pos   { return n.Begin() }

func (n *RangeClause) Begin() src.Pos {
	if n.Lhs != nil {
		return n.Lhs.Begin()
	}
	return n.Range
}
func (n *RangeClause) End() src.Pos { return n.X.End() }
func (n *RangeClause) Pos() src.Pos { return n.Range }

func (n *TypeSwitchGuard) Begin() src.Pos {
	if n.Lhs != nil {
		return n.Lhs.Begin()
	}
	return n.X.Begin()
}
func (n *TypeSwitchGuard) End() src.Pos { return adj(n.X.End(), len(".(type)")) }
func (n *TypeSwitchGuard) Pos() src.Pos { return n.Dot }

func (n *CaseClause) Begin() src.Pos { return n.Case }
func (n *CaseClause) End() src.Pos {
	if i := len(n.Body); i > 0 {
		return n.Body[i-1].End()
	}
	return adj(n.Colon, len(":"))
}
func (n *CaseClause) Pos() src.Pos { return n.Begin() }

func (n *CommClause) Begin() src.Pos { return n.Case }
func (n *CommClause) End() src.Pos {
	if i := len(n.Body); i > 0 {
		return n.Body[i-1].End()
	}
	return adj(n.Colon, len(":"))
}
func (n *CommClause) Pos() src.Pos { return n.Begin() }

// adj returns the position pos + delta, where delta is a number of bytes on the same line.
// TODO(gri) this can be done more efficiently in package src
func adj(pos src.Pos, delta int) src.Pos {
	col := int(pos.Col()) + delta
	if col < 1 {
		panic("invalid column")
	}
	return src.MakePos(pos.Base(), pos.Line(), uint(col))
}

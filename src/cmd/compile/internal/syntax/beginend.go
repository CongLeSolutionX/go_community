// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syntax

import "cmd/internal/src"

func (n *File) Pos() src.Pos {
	return n.Package
}

func (n *File) Begin() src.Pos {
	return n.Package
}

func (n *File) End() src.Pos {
	if i := len(n.DeclList); i > 0 {
		return n.DeclList[i-1].End()
	}
	return n.PkgName.End()
}

func (n *ImportDecl) Begin() src.Pos {
	if n.LocalPkgName != nil {
		return n.LocalPkgName.Begin()
	}
	return n.Path.Begin()
}

func (n *ImportDecl) End() src.Pos {
	return n.Path.End()
}

func (n *ConstDecl) Begin() src.Pos {
	return n.NameList[0].Begin()
}

func (n *ConstDecl) End() src.Pos {
	if n.Values != nil {
		return n.Values.End()
	}
	return n.NameList[len(n.NameList)-1].End()
}

func (n *TypeDecl) Begin() src.Pos {
	return n.Name.Begin()
}

func (n *TypeDecl) End() src.Pos {
	return n.Type.End()
}

func (n *VarDecl) Begin() src.Pos {
	return n.NameList[0].Begin()
}

func (n *VarDecl) End() src.Pos {
	if n.Values != nil {
		return n.Values.End()
	}
	return n.Type.End()
}

func (n *FuncDecl) Begin() src.Pos {
	return n.pos
}

func (n *FuncDecl) End() src.Pos {
	return adj(n.Rbrace, len("}"))
}

func (n *Name) Begin() src.Pos {
	return n.pos
}

func (n *Name) End() src.Pos {
	return adj(n.pos, len(n.Value))
}

func (n *BasicLit) Begin() src.Pos {
	return n.pos
}

func (n *BasicLit) End() src.Pos {
	return adj(n.pos, len(n.Value))
}

func (n *CompositeLit) Begin() src.Pos {
	if n.Type != nil {
		return n.Type.Begin()
	}
	return n.pos
}

func (n *CompositeLit) End() src.Pos {
	return adj(n.Rbrace, len("}"))
}

func (n *KeyValueExpr) Begin() src.Pos {
	return n.Key.Begin()
}

func (n *KeyValueExpr) End() src.Pos {
	return n.Value.End()
}

func (n *FuncLit) Begin() src.Pos {
	return n.pos
}

func (n *FuncLit) End() src.Pos {
	return adj(n.Rbrace, len("}"))
}

func (n *ParenExpr) Pos() src.Pos {
	return n.Lparen
}

func (n *ParenExpr) Begin() src.Pos {
	return n.Lparen
}

func (n *ParenExpr) End() src.Pos {
	return adj(n.Rparen, len(")"))
}

func (n *SelectorExpr) Begin() src.Pos {
	return n.X.Begin()
}

func (n *SelectorExpr) End() src.Pos {
	return n.Sel.End()
}

func (n *IndexExpr) Begin() src.Pos {
	return n.X.Begin()
}

func (n *IndexExpr) End() src.Pos {
	return adj(n.Rbrack, len("]"))
}

func (n *SliceExpr) Begin() src.Pos {
	return n.X.Begin()
}

func (n *SliceExpr) End() src.Pos {
	return adj(n.Rbrack, len("]"))
}

func (n *AssertExpr) Begin() src.Pos {
	return n.X.Begin()
}

func (n *AssertExpr) End() src.Pos {
	// TODO(gri) need to record actual ")" position
	if n.Type != nil {
		return adj(n.Type.End(), len(")"))
	}
	return adj(n.X.End(), len(".(type)"))
}

func (n *Operation) Begin() src.Pos {
	return n.X.Begin()
}

func (n *Operation) End() src.Pos {
	if n.Y != nil {
		return n.Y.End()
	}
	return n.X.End()
}

func (n *CallExpr) Begin() src.Pos {
	return n.Fun.Begin()
}

func (n *CallExpr) End() src.Pos {
	// TODO(gri) need to record actual ")" position
	if i := len(n.ArgList); i > 0 {
		return adj(n.ArgList[i-1].End(), len(")"))
	}
	return adj(n.Fun.End(), len("()"))
}

func (n *ListExpr) Begin() src.Pos {
	return n.ElemList[0].Begin()
}

func (n *ListExpr) End() src.Pos {
	return n.ElemList[len(n.ElemList)-1].End()
}

func (n *ArrayType) Begin() src.Pos {
	return n.pos
}

func (n *ArrayType) End() src.Pos {
	return n.Elem.End()
}

func (n *SliceType) Begin() src.Pos {
	return n.pos
}

func (n *SliceType) End() src.Pos {
	return n.Elem.End()
}

func (n *DotsType) Begin() src.Pos {
	return n.pos
}

func (n *DotsType) End() src.Pos {
	return n.Elem.End()
}

func (n *StructType) Begin() src.Pos {
	return n.pos
}

func (n *StructType) End() src.Pos {
	return adj(n.Rbrace, len("}"))
}

func (n *Field) Begin() src.Pos {
	if n.Name != nil {
		return n.Name.Begin()
	}
	return n.Type.Begin()
}

func (n *Field) End() src.Pos {
	// a Field comprises of the field name only
	if n.Name != nil {
		return n.Name.End()
	}
	return n.Type.End()
}

func (n *InterfaceType) Begin() src.Pos {
	return n.pos
}

func (n *InterfaceType) End() src.Pos {
	return adj(n.Rbrace, len("}"))
}

func (n *FuncType) Begin() src.Pos {
	return n.pos
}

func (n *FuncType) End() src.Pos {
	// TODO(gri) need to record actual end position
	panic("unimplemented")
}

func (n *MapType) Begin() src.Pos {
	return n.pos
}

func (n *MapType) End() src.Pos {
	return n.Value.End()
}

func (n *ChanType) Begin() src.Pos {
	return n.pos
}

func (n *ChanType) End() src.Pos {
	return n.Elem.End()
}

func (n *EmptyStmt) Begin() src.Pos {
	return n.pos
}

func (n *EmptyStmt) End() src.Pos {
	return n.pos
}

func (n *LabeledStmt) Pos() src.Pos {
	return n.Colon
}

func (n *LabeledStmt) Begin() src.Pos {
	return n.Label.Begin()
}

func (n *LabeledStmt) End() src.Pos {
	return n.Stmt.End()
}

func (n *BlockStmt) Pos() src.Pos {
	return n.Lbrace
}

func (n *BlockStmt) Begin() src.Pos {
	return n.Lbrace
}

func (n *BlockStmt) End() src.Pos {
	return adj(n.Rbrace, len("}"))
}

func (n *ExprStmt) Begin() src.Pos {
	return n.X.Begin()
}

func (n *ExprStmt) End() src.Pos {
	return n.X.End()
}

func (n *SendStmt) Begin() src.Pos {
	return n.Chan.Begin()
}

func (n *SendStmt) End() src.Pos {
	return n.Value.End()
}

func (n *DeclStmt) Begin() src.Pos {
	return n.pos
}

func (n *DeclStmt) End() src.Pos {
	return n.DeclList[len(n.DeclList)-1].End()
}

func (n *AssignStmt) Begin() src.Pos {
	return n.Lhs.Begin()
}

func (n *AssignStmt) End() src.Pos {
	return n.Rhs.End()
}

func (n *BranchStmt) Begin() src.Pos {
	return n.pos
}

func (n *BranchStmt) End() src.Pos {
	if n.Label != nil {
		return n.Label.End()
	}
	return adj(n.pos, len(n.Tok.String()))
}

func (n *CallStmt) Begin() src.Pos {
	return n.pos
}

func (n *CallStmt) End() src.Pos {
	return n.Call.End()
}

func (n *ReturnStmt) Begin() src.Pos {
	return n.pos
}

func (n *ReturnStmt) End() src.Pos {
	if n.Results != nil {
		return n.Results.End()
	}
	return adj(n.pos, len("return"))
}

func (n *IfStmt) Begin() src.Pos {
	return n.pos
}

func (n *IfStmt) End() src.Pos {
	if n.Else != nil {
		return n.Else.End()
	}
	// TODO(gri) need to use actual block here (probably)
	return n.Then[len(n.Then)-1].End()
}

func (n *ForStmt) Begin() src.Pos {
	return n.pos
}

func (n *ForStmt) End() src.Pos {
	// TODO(gri) need to use actual block here (probably)
	return n.Body[len(n.Body)-1].End()
}

func (n *SwitchStmt) Begin() src.Pos {
	return n.pos
}

func (n *SwitchStmt) End() src.Pos {
	// TODO(gri) need to use actual block here (probably)
	return n.Body[len(n.Body)-1].End()
}

func (n *SelectStmt) Begin() src.Pos {
	return n.pos
}

func (n *SelectStmt) End() src.Pos {
	// TODO(gri) need to use actual block here (probably)
	return n.Body[len(n.Body)-1].End()
}

func (n *RangeClause) Begin() src.Pos {
	if n.Lhs != nil {
		return n.Lhs.Begin()
	}
	return n.pos
}

func (n *RangeClause) End() src.Pos {
	return n.X.End()
}

func (n *TypeSwitchGuard) Begin() src.Pos {
	if n.Lhs != nil {
		return n.Lhs.Begin()
	}
	return n.pos
}

func (n *TypeSwitchGuard) End() src.Pos {
	return adj(n.X.End(), len(".(type)"))
}

func (n *CaseClause) Pos() src.Pos {
	return n.Case
}

func (n *CaseClause) Begin() src.Pos {
	return n.Case
}

func (n *CaseClause) End() src.Pos {
	if i := len(n.Body); i > 0 {
		return n.Body[i-1].End()
	}
	return adj(n.Colon, len(":"))
}

func (n *CommClause) Pos() src.Pos {
	return n.Case
}

func (n *CommClause) Begin() src.Pos {
	return n.Case
}

func (n *CommClause) End() src.Pos {
	if i := len(n.Body); i > 0 {
		return n.Body[i-1].End()
	}
	return adj(n.Colon, len(":"))
}

// adj returns the position pos + delta, where delta is a number of bytes on the same line.
// TODO(gri) this can be done more efficiently in package src
func adj(pos src.Pos, delta int) src.Pos {
	col := int(pos.Col()) + delta
	if col < 1 {
		panic("invalid column")
	}
	return src.MakePos(pos.Base(), pos.Line(), uint(col))
}

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package noder

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/ir"
	"cmd/compile/internal/syntax"
	"cmd/compile/internal/typecheck"
	"cmd/compile/internal/types"
	"cmd/internal/src"
)

func (p *irgen) funcBody(fn *ir.Func, block *syntax.BlockStmt) {
	// TODO(mdempsky): Shouldn't need this.
	oldfn, oldctxt := ir.CurFunc, typecheck.DeclContext
	ir.CurFunc, typecheck.DeclContext = fn, ir.PAUTO
	defer func() {
		ir.CurFunc, typecheck.DeclContext = oldfn, oldctxt
	}()

	if block != nil {
		body := p.stmts(block.List)
		if body == nil {
			body = []ir.Node{ir.NewBlockStmt(base.Pos, nil)}
		}
		fn.Body = body

		base.Pos = p.pos0(block.Rbrace)
		fn.Endlineno = base.Pos
	}

	typecheck.Target.Decls = append(typecheck.Target.Decls, fn)
}

func (p *irgen) stmts(stmts []syntax.Stmt) []ir.Node {
	var nodes []ir.Node
	for _, stmt := range stmts {
		switch s := p.stmt(stmt).(type) {
		case nil:
		case *ir.BlockStmt:
			nodes = append(nodes, s.List...)
		default:
			nodes = append(nodes, s)
		}
	}
	return nodes
}

func (p *irgen) stmt(stmt syntax.Stmt) ir.Node {
	switch stmt := stmt.(type) {
	case nil, *syntax.EmptyStmt:
		return nil
	case *syntax.LabeledStmt:
		return p.labeledStmt(stmt)
	case *syntax.BlockStmt:
		return ir.NewBlockStmt(p.pos(stmt), p.blockStmt(stmt))
	case *syntax.ExprStmt:
		return p.expr(stmt.X)
	case *syntax.SendStmt:
		return ir.NewSendStmt(p.pos(stmt), p.expr(stmt.Chan), p.expr(stmt.Value))
	case *syntax.DeclStmt:
		return ir.NewBlockStmt(p.pos(stmt), p.decls(nil, stmt.DeclList))

	case *syntax.AssignStmt:
		if stmt.Op != 0 && stmt.Op != syntax.Def {
			n := ir.NewAssignOpStmt(p.pos(stmt), p.op(stmt.Op, binOps[:]), p.expr(stmt.Lhs), p.expr(stmt.Rhs))
			n.IncDec = stmt.Rhs == syntax.ImplicitOne
			return n
		}

		rhs := p.exprList(stmt.Rhs)
		if list, ok := stmt.Lhs.(*syntax.ListExpr); ok && len(list.ElemList) != 1 || len(rhs) != 1 {
			n := ir.NewAssignListStmt(p.pos(stmt), ir.OAS2, nil, nil)
			n.Def = stmt.Op == syntax.Def
			n.Lhs = p.assignList(stmt.Lhs, n, n.Def)
			n.Rhs = rhs
			return n
		}

		n := ir.NewAssignStmt(p.pos(stmt), nil, nil)
		n.Def = stmt.Op == syntax.Def
		n.X = p.assignList(stmt.Lhs, n, n.Def)[0]
		n.Y = rhs[0]
		return n

	case *syntax.BranchStmt:
		return ir.NewBranchStmt(p.pos(stmt), p.tokOp(stmt.Tok, branchOps[:]), p.name(stmt.Label))
	case *syntax.CallStmt:
		return ir.NewGoDeferStmt(p.pos(stmt), p.tokOp(stmt.Tok, callOps[:]), p.expr(stmt.Call))
	case *syntax.ReturnStmt:
		return ir.NewReturnStmt(p.pos(stmt), p.exprList(stmt.Results))
	case *syntax.IfStmt:
		return p.ifStmt(stmt)
	case *syntax.ForStmt:
		return p.forStmt(stmt)
	case *syntax.SelectStmt:
		return p.selectStmt(stmt)
	case *syntax.SwitchStmt:
		return p.switchStmt(stmt)
	}

	base.FatalfAt(p.pos(stmt), "unhandled statement: %v (%T)", stmt, stmt)
	panic("unreachable")
}

var branchOps = [...]ir.Op{
	syntax.Break:       ir.OBREAK,
	syntax.Continue:    ir.OCONTINUE,
	syntax.Fallthrough: ir.OFALL,
	syntax.Goto:        ir.OGOTO,
}

var callOps = [...]ir.Op{
	syntax.Defer: ir.ODEFER,
	syntax.Go:    ir.OGO,
}

func (p *irgen) tokOp(tok syntax.Token, ops []ir.Op) ir.Op {
	// TODO(mdempsky): Validate.
	return ops[tok]
}

func (p *irgen) op(op syntax.Operator, ops []ir.Op) ir.Op {
	// TODO(mdempsky): Validate.
	return ops[op]
}

func (p *irgen) assignList(expr syntax.Expr, defn ir.InitNode, colas bool) []ir.Node {
	if !colas {
		return p.exprList(expr)
	}

	var exprs []syntax.Expr
	if list, ok := expr.(*syntax.ListExpr); ok {
		exprs = list.ElemList
	} else {
		exprs = []syntax.Expr{expr}
	}

	res := make([]ir.Node, len(exprs))
	for i, expr := range exprs {
		expr := expr.(*syntax.Name)
		if expr.Value == "_" {
			// TODO(mdempsky): Does info.Uses handle blanks?
			res[i] = ir.BlankNode
			continue
		}

		if obj, ok := p.info.Uses[expr]; ok {
			res[i] = p.obj(obj)
			continue
		}

		// TODO(mdempsky): Anything else to setup here?
		name := p.def(expr)
		name.Defn = defn
		res[i] = name
	}
	return res
}

func (p *irgen) blockStmt(stmt *syntax.BlockStmt) []ir.Node {
	return p.stmts(stmt.List)
}

func (p *irgen) ifStmt(stmt *syntax.IfStmt) ir.Node {
	init := p.stmt(stmt.Init)
	n := ir.NewIfStmt(p.pos(stmt), p.expr(stmt.Cond), p.blockStmt(stmt.Then), nil)
	if stmt.Else != nil {
		e := p.stmt(stmt.Else)
		if e.Op() == ir.OBLOCK {
			e := e.(*ir.BlockStmt)
			n.Else = e.List
		} else {
			n.Else = []ir.Node{e}
		}
	}
	return p.init(init, n)
}

func (p *irgen) forStmt(stmt *syntax.ForStmt) ir.Node {
	if r, ok := stmt.Init.(*syntax.RangeClause); ok {
		pos := p.pos(r)
		if stmt.Cond != nil || stmt.Post != nil {
			base.FatalfAt(pos, "invalid RangeClause")
		}

		n := ir.NewRangeStmt(pos, nil, nil, p.expr(r.X), nil)
		if r.Lhs != nil {
			n.Def = r.Def
			lhs := p.assignList(r.Lhs, n, n.Def)
			n.Key = lhs[0]
			if len(lhs) > 1 {
				n.Value = lhs[1]
			}
		}
		n.Body = p.blockStmt(stmt.Body)
		return n
	}

	return ir.NewForStmt(p.pos(stmt), p.stmt(stmt.Init), p.expr(stmt.Cond), p.stmt(stmt.Post), p.blockStmt(stmt.Body))
}

func (p *irgen) init(init ir.Node, stmt ir.InitNode) ir.InitNode {
	if init != nil {
		stmt.SetInit([]ir.Node{init})
	}
	return stmt
}

func (p *irgen) selectStmt(stmt *syntax.SelectStmt) ir.Node {
	body := make([]*ir.CommClause, len(stmt.Body))
	for i, clause := range stmt.Body {
		body[i] = ir.NewCommStmt(p.pos(clause), p.stmt(clause.Comm), p.stmts(clause.Body))
	}
	return ir.NewSelectStmt(p.pos(stmt), body)
}

func (p *irgen) switchStmt(stmt *syntax.SwitchStmt) ir.Node {
	pos := p.pos(stmt)
	init := p.stmt(stmt.Init)

	var expr ir.Node
	switch tag := stmt.Tag.(type) {
	case *syntax.TypeSwitchGuard:
		var ident *ir.Ident
		if tag.Lhs != nil {
			ident = ir.NewIdent(p.pos(tag.Lhs), p.name(tag.Lhs))
		}
		expr = ir.NewTypeSwitchGuard(pos, ident, p.expr(tag.X))
	default:
		expr = p.expr(tag)
	}

	body := make([]*ir.CaseClause, len(stmt.Body))
	for i, clause := range stmt.Body {
		body[i] = ir.NewCaseStmt(p.pos(clause), p.exprList(clause.Cases), p.stmts(clause.Body))
		if obj, ok := p.info.Implicits[clause]; ok {
			body[i].Var = p.obj(obj)
		}
	}

	return p.init(init, ir.NewSwitchStmt(pos, expr, body))
}

func (p *irgen) labeledStmt(label *syntax.LabeledStmt) ir.Node {
	sym := p.name(label.Label)
	lhs := ir.NewLabelStmt(p.pos(label), sym)

	var ls ir.Node
	if label.Stmt != nil { // TODO(mdempsky): Should always be present.
		ls = p.stmt(label.Stmt)
		// Attach label directly to control statement too.
		switch ls := ls.(type) {
		case *ir.ForStmt:
			ls.Label = sym
		case *ir.RangeStmt:
			ls.Label = sym
		case *ir.SelectStmt:
			ls.Label = sym
		case *ir.SwitchStmt:
			ls.Label = sym
		}
	}

	l := []ir.Node{lhs}
	if ls != nil {
		if ls.Op() == ir.OBLOCK {
			ls := ls.(*ir.BlockStmt)
			l = append(l, ls.List...)
		} else {
			l = append(l, ls)
		}
	}
	return ir.NewBlockStmt(src.NoXPos, l)
}

func (p *irgen) name(name *syntax.Name) *types.Sym {
	if name == nil {
		return nil
	}
	return typecheck.Lookup(name.Value)
}

func (p *irgen) funcLit(expr *syntax.FuncLit) ir.Node {
	typ := p.typeExpr(expr.Type)

	fn := ir.NewFunc(p.pos(expr))
	fn.SetIsHiddenClosure(ir.CurFunc != nil)

	fn.Nname = ir.NewNameAt(p.pos(expr), ir.BlankNode.Sym()) // filled in by typecheckclosure
	fn.Nname.Ntype = typ
	fn.Nname.Func = fn
	fn.Nname.Defn = fn

	fn.OClosure = ir.NewClosureExpr(p.pos(expr), fn)
	fn.ClosureType = typ

	p.funcBody(fn, expr.Body)

	return fn.OClosure
}

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package noder

import (
	"go/constant"

	"cmd/compile/internal/base"
	"cmd/compile/internal/ir"
	"cmd/compile/internal/syntax"
	"cmd/compile/internal/typecheck"
	"cmd/compile/internal/types"
	"cmd/compile/internal/types2"
	"cmd/internal/src"
)

func (p *nyan) funk(decl *syntax.FuncDecl) {
	fn := ir.NewFunc(p.pos(decl))
	fn.Nname = p.def(decl.Name)
	fn.Nname.Func = fn
	fn.Nname.Defn = fn

	if decl.Name.Value == "init" {
		typecheck.Target.Inits = append(typecheck.Target.Inits, fn)
	}

	p.funcBody(fn, decl.Body)
}

func (p *nyan) funcBody(fn *ir.Func, block *syntax.BlockStmt) {
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

func (p *nyan) def(name *syntax.Name) *ir.Name {
	if obj, ok := p.info.Defs[name]; ok {
		return p.obj(obj)
	}
	base.FatalfAt(p.pos(name), "unknown name %v", name)
	panic("unreachable")
}

func (p *nyan) defs(names []*syntax.Name) []*ir.Name {
	res := make([]*ir.Name, len(names))
	for i, name := range names {
		res[i] = p.def(name)
	}
	return res
}

func (p *nyan) use(name *syntax.Name) ir.Node {
	if obj, ok := p.info.Uses[name]; ok {
		// TODO(mdempsky): Create closure variables.
		return p.obj(obj)
	}
	base.FatalfAt(p.pos(name), "unknown name %v", name)
	panic("unreachable")
}

func (p *nyan) stmts(stmts []syntax.Stmt) []ir.Node {
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

func (p *nyan) decls(decls []syntax.Decl) []ir.Node {
	var res []ir.Node
	for _, decl := range decls {
		// TODO(mdempsky): Do we care about other declarations? Probably
		// eventually for handling pragmas.
		switch decl := decl.(type) {
		case *syntax.VarDecl:
			pos := p.pos(decl)
			names := p.defs(decl.NameList)
			values := p.exprList(decl.Values)

			// TODO(mdempsky): To avoid typecheck clobbering things later,
			// I'm not setting name.Defn. That is, I'm basically
			// representing "var x T = v" as separate "var x T; x = v"
			// statements. This works correctly, but it will interfere with
			// frontend optimizations that recognize never-reassigned
			// variables.

			var as *ir.AssignListStmt
			if len(values) != 0 && len(names) != len(values) {
				as = ir.NewAssignListStmt(pos, ir.OAS2, make([]ir.Node, len(names)), values)
			}

			for i, name := range names {
				res = append(res, ir.NewDecl(pos, ir.ODCL, name), ir.NewAssignStmt(pos, name, nil))
				if as != nil {
					as.Lhs[i] = name
				} else if len(values) != 0 {
					res = append(res, ir.NewAssignStmt(pos, name, values[i]))
				}
			}
		}
	}
	return res
}

func (p *nyan) stmt(stmt syntax.Stmt) ir.Node {
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
		return ir.NewBlockStmt(p.pos(stmt), p.decls(stmt.DeclList))

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

func (p *nyan) tokOp(tok syntax.Token, ops []ir.Op) ir.Op {
	// TODO(mdempsky): Validate.
	return ops[tok]
}

func (p *nyan) op(op syntax.Operator, ops []ir.Op) ir.Op {
	// TODO(mdempsky): Validate.
	return ops[op]
}

func (p *nyan) assignList(expr syntax.Expr, defn ir.InitNode, colas bool) []ir.Node {
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

func (p *nyan) blockStmt(stmt *syntax.BlockStmt) []ir.Node {
	return p.stmts(stmt.List)
}

func (p *nyan) ifStmt(stmt *syntax.IfStmt) ir.Node {
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

func (p *nyan) forStmt(stmt *syntax.ForStmt) ir.Node {
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

func (p *nyan) init(init ir.Node, stmt ir.InitNode) ir.InitNode {
	if init != nil {
		stmt.SetInit([]ir.Node{init})
	}
	return stmt
}

func (p *nyan) selectStmt(stmt *syntax.SelectStmt) ir.Node {
	body := make([]*ir.CommClause, len(stmt.Body))
	for i, clause := range stmt.Body {
		body[i] = ir.NewCommStmt(p.pos(clause), p.stmt(clause.Comm), p.stmts(clause.Body))
	}
	return ir.NewSelectStmt(p.pos(stmt), body)
}

func (p *nyan) switchStmt(stmt *syntax.SwitchStmt) ir.Node {
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

func (p *nyan) labeledStmt(label *syntax.LabeledStmt) ir.Node {
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

func (p *nyan) name(name *syntax.Name) *types.Sym {
	if name == nil {
		return nil
	}
	return typecheck.Lookup(name.Value)
}

func (p *nyan) funcLit(expr *syntax.FuncLit) ir.Node {
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

func (p *nyan) exprList(expr syntax.Expr) []ir.Node {
	switch expr := expr.(type) {
	case nil:
		return nil
	case *syntax.ListExpr:
		return p.exprs(expr.ElemList)
	default:
		return []ir.Node{p.expr(expr)}
	}
}

func (p *nyan) exprs(exprs []syntax.Expr) []ir.Node {
	nodes := make([]ir.Node, len(exprs))
	for i, expr := range exprs {
		nodes[i] = p.expr(expr)
	}
	return nodes
}

func (p *nyan) expr(expr syntax.Expr) ir.Node {
	if expr == nil {
		// TODO(mdempsky): Is this still needed?
		return nil
	}

	if expr == syntax.ImplicitOne {
		return ir.NewBasicLit(src.NoXPos, constant.MakeInt64(1))
	}

	if expr, ok := expr.(*syntax.Name); ok && expr.Value == "_" {
		return ir.BlankNode
	}

	// TODO(mdempsky): Is there a better way to handle qualified identifiers?
	if expr, ok := expr.(*syntax.SelectorExpr); ok {
		if name, ok := expr.X.(*syntax.Name); ok {
			if _, ok := p.info.Uses[name].(*types2.PkgName); ok {
				return p.use(expr.Sel)
			}
		}
	}

	pos := p.pos(expr)
	tv, ok := p.info.Types[expr]
	if !ok {
		base.FatalfAt(pos, "missing type for %v (%T)", expr, expr)
	}
	switch {
	case tv.IsType():
		return ir.TypeNode(p.typ(tv.Type))
	case tv.IsVoid():
		// ok
	case tv.IsBuiltin():
		return p.use(expr.(*syntax.Name))
	case !tv.IsValue():
		base.FatalfAt(pos, "TODO: %v %v", expr, tv)
	}

	// Note: tv.Type can be a tuple here, for map index and type
	// asserts; presumably also receives and function calls.

	if tv.Value != nil {
		n := ir.NewBasicLit(pos, tv.Value)
		n.SetType(p.typ(tv.Type))
		return n
	}

	switch expr := expr.(type) {
	case *syntax.Name:
		if _, isNil := p.info.Uses[expr].(*types2.Nil); isNil {
			n := ir.NewNilExpr(pos)
			n.SetType(p.typ(tv.Type))
			return n
		}

		n := p.use(expr)
		if typ := p.typ(tv.Type); !types.Identical(n.Type(), typ) {
			base.FatalfAt(pos, "huh: %v: %v != %v", n, n.Type(), typ)
		}
		return n

	case *syntax.CompositeLit:
		return p.compLit(tv.Type, expr)
	case *syntax.FuncLit:
		return p.funcLit(expr)
	case *syntax.ParenExpr:
		return ir.NewParenExpr(pos, p.expr(expr.X))
	case *syntax.SelectorExpr:
		return ir.NewSelectorExpr(pos, ir.OXDOT, p.expr(expr.X), p.name(expr.Sel))
	case *syntax.IndexExpr:
		return ir.NewIndexExpr(pos, p.expr(expr.X), p.expr(expr.Index))
	case *syntax.SliceExpr:
		op := ir.OSLICE
		if expr.Full {
			op = ir.OSLICE3
		}
		x := p.expr(expr.X)
		var index [3]ir.Node
		for i, n := range &expr.Index {
			if n != nil {
				index[i] = p.expr(n)
			}
		}
		return ir.NewSliceExpr(pos, op, x, index[0], index[1], index[2])
	case *syntax.AssertExpr:
		return ir.NewTypeAssertExpr(pos, p.expr(expr.X), p.typeExpr(expr.Type))

	case *syntax.Operation:
		// Binary operations.
		if expr.Y != nil {
			x, y := p.expr(expr.X), p.expr(expr.Y)
			switch op := p.op(expr.Op, binOps[:]); op {
			case ir.OANDAND, ir.OOROR:
				return ir.NewLogicalExpr(pos, op, x, y)
			default:
				return ir.NewBinaryExpr(pos, op, x, y)
			}
		}

		// Unary operations.
		x := p.expr(expr.X)
		switch op := p.op(expr.Op, unOps[:]); op {
		case ir.OADDR:
			return typecheck.NodAddrAt(pos, x)
		case ir.ODEREF:
			return ir.NewStarExpr(pos, x)
		default:
			return ir.NewUnaryExpr(pos, op, x)
		}

	case *syntax.CallExpr:
		n := ir.NewCallExpr(pos, ir.OCALL, p.expr(expr.Fun), p.exprs(expr.ArgList))
		n.IsDDD = expr.HasDots
		return n
	}
	panic("unhandled Expr")
}

func (p *nyan) compLit(typ types2.Type, lit *syntax.CompositeLit) *ir.CompLitExpr {
	_, isStruct := typ.Underlying().(*types2.Struct)

	exprs := make([]ir.Node, len(lit.ElemList))
	for i, elem := range lit.ElemList {
		switch elem := elem.(type) {
		case *syntax.KeyValueExpr:
			if isStruct {
				exprs[i] = ir.NewStructKeyExpr(p.pos(elem), p.name(elem.Key.(*syntax.Name)), p.expr(elem.Value))
			} else {
				exprs[i] = ir.NewKeyExpr(p.pos(elem), p.expr(elem.Key), p.expr(elem.Value))
			}
		default:
			exprs[i] = p.expr(elem)
		}
	}

	return ir.NewCompLitExpr(p.pos(lit), ir.OCOMPLIT, ir.TypeNode(p.typ(typ)), exprs)
}

// TODO(mdempsky): This shouldn't be necessary at all.
func (p *nyan) typeExpr(typ syntax.Expr) ir.Ntype {
	if typ == nil {
		return nil
	}
	return p.expr(typ).(ir.Ntype)
}

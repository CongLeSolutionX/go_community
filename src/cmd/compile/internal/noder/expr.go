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

func (g *irgen) expr(expr syntax.Expr) ir.Node {
	// TODO(mdempsky): Is this still needed?
	if expr == nil {
		return nil
	}

	// TODO(mdempsky): Push this responsibility to caller.
	if expr == syntax.ImplicitOne {
		return ir.NewBasicLit(src.NoXPos, constant.MakeInt64(1))
	}

	if expr, ok := expr.(*syntax.Name); ok && expr.Value == "_" {
		return ir.BlankNode
	}

	// TODO(mdempsky): Is there a better way to handle qualified identifiers?
	if expr, ok := expr.(*syntax.SelectorExpr); ok {
		if name, ok := expr.X.(*syntax.Name); ok {
			if _, ok := g.info.Uses[name].(*types2.PkgName); ok {
				return g.use(expr.Sel)
			}
		}
	}

	pos := g.pos(expr)
	tv, ok := g.info.Types[expr]
	if !ok {
		base.FatalfAt(pos, "missing type for %v (%T)", expr, expr)
	}
	switch {
	case tv.IsType():
		return ir.TypeNode(g.typ(tv.Type))
	case tv.IsVoid():
		// ok
	case tv.IsBuiltin():
		return g.use(expr.(*syntax.Name))
	case !tv.IsValue():
		base.FatalfAt(pos, "TODO: %v %v", expr, tv)
	}

	// Note: tv.Type can be a tuple here, for map index and type
	// asserts; presumably also receives and function calls.
	// TODO(mdempsky): Handle these when constructing assignments
	// instead.

	if tv.Value != nil {
		n := ir.NewBasicLit(pos, tv.Value)
		n.SetType(g.typ(tv.Type))
		return n
	}

	switch expr := expr.(type) {
	case *syntax.Name:
		typ := g.typ(tv.Type)
		if _, isNil := g.info.Uses[expr].(*types2.Nil); isNil {
			n := ir.NewNilExpr(pos)
			n.SetType(g.typ(tv.Type))
			return n
		}

		n := g.use(expr)
		if !types.Identical(n.Type(), typ) {
			base.FatalfAt(pos, "unexpected type for %v: %v != %v", n, n.Type(), typ)
		}
		return n

	case *syntax.CompositeLit:
		return g.compLit(tv.Type, expr)
	case *syntax.FuncLit:
		return g.funcLit(expr)
	case *syntax.ParenExpr:
		return ir.NewParenExpr(pos, g.expr(expr.X))
	case *syntax.SelectorExpr:
		return ir.NewSelectorExpr(pos, ir.OXDOT, g.expr(expr.X), g.name(expr.Sel))
	case *syntax.IndexExpr:
		return ir.NewIndexExpr(pos, g.expr(expr.X), g.expr(expr.Index))
	case *syntax.SliceExpr:
		op := ir.OSLICE
		if expr.Full {
			op = ir.OSLICE3
		}
		x := g.expr(expr.X)
		var index [3]ir.Node
		for i, n := range &expr.Index {
			if n != nil {
				index[i] = g.expr(n)
			}
		}
		return ir.NewSliceExpr(pos, op, x, index[0], index[1], index[2])
	case *syntax.AssertExpr:
		return ir.NewTypeAssertExpr(pos, g.expr(expr.X), g.typeExpr(expr.Type))

	case *syntax.Operation:
		// Binary operations.
		if expr.Y != nil {
			x, y := g.expr(expr.X), g.expr(expr.Y)
			switch op := g.op(expr.Op, binOps[:]); op {
			case ir.OANDAND, ir.OOROR:
				return ir.NewLogicalExpr(pos, op, x, y)
			default:
				return ir.NewBinaryExpr(pos, op, x, y)
			}
		}

		// Unary operations.
		x := g.expr(expr.X)
		switch op := g.op(expr.Op, unOps[:]); op {
		case ir.OADDR:
			return typecheck.NodAddrAt(pos, x)
		case ir.ODEREF:
			return ir.NewStarExpr(pos, x)
		default:
			return ir.NewUnaryExpr(pos, op, x)
		}

	case *syntax.CallExpr:
		n := ir.NewCallExpr(pos, ir.OCALL, g.expr(expr.Fun), g.exprs(expr.ArgList))
		n.IsDDD = expr.HasDots
		return n
	}

	g.unhandled("expression", expr)
	panic("unreachable")
}

func (g *irgen) exprList(expr syntax.Expr) []ir.Node {
	switch expr := expr.(type) {
	case nil:
		return nil
	case *syntax.ListExpr:
		return g.exprs(expr.ElemList)
	default:
		return []ir.Node{g.expr(expr)}
	}
}

func (g *irgen) exprs(exprs []syntax.Expr) []ir.Node {
	nodes := make([]ir.Node, len(exprs))
	for i, expr := range exprs {
		nodes[i] = g.expr(expr)
	}
	return nodes
}

func (g *irgen) compLit(typ types2.Type, lit *syntax.CompositeLit) ir.Node {
	// TODO(mdempsky): I don't think there can be composite
	// literals of named pointer types?
	if ptr, ok := typ.(*types2.Pointer); ok {
		return ir.NewAddrExpr(g.pos(lit), g.compLit(ptr.Elem(), lit))
	}

	_, isStruct := typ.Underlying().(*types2.Struct)

	exprs := make([]ir.Node, len(lit.ElemList))
	for i, elem := range lit.ElemList {
		switch elem := elem.(type) {
		case *syntax.KeyValueExpr:
			if isStruct {
				exprs[i] = ir.NewStructKeyExpr(g.pos(elem), g.name(elem.Key.(*syntax.Name)), g.expr(elem.Value))
			} else {
				exprs[i] = ir.NewKeyExpr(g.pos(elem), g.expr(elem.Key), g.expr(elem.Value))
			}
		default:
			exprs[i] = g.expr(elem)
		}
	}

	return ir.NewCompLitExpr(g.pos(lit), ir.OCOMPLIT, ir.TypeNode(g.typ(typ)), exprs)
}

func (g *irgen) funcLit(expr *syntax.FuncLit) ir.Node {
	typ := g.typeExpr(expr.Type)

	fn := ir.NewFunc(g.pos(expr))
	fn.SetIsHiddenClosure(ir.CurFunc != nil)

	fn.Nname = ir.NewNameAt(g.pos(expr), ir.BlankNode.Sym()) // filled in by typecheckclosure
	fn.Nname.Ntype = typ
	fn.Nname.Func = fn
	fn.Nname.Defn = fn

	fn.OClosure = ir.NewClosureExpr(g.pos(expr), fn)
	fn.ClosureType = typ

	g.funcBody(fn, expr.Body)

	return fn.OClosure
}

// TODO(mdempsky): This shouldn't be necessary at all.
func (g *irgen) typeExpr(typ syntax.Expr) ir.Ntype {
	if typ == nil {
		return nil
	}
	return g.expr(typ).(ir.Ntype)
}

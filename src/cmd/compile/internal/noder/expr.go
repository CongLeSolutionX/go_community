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
	"cmd/compile/internal/types2"
	"cmd/internal/src"
	"go/constant"
)

func (p *irgen) expr(expr syntax.Expr) ir.Node {
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

func (p *irgen) exprList(expr syntax.Expr) []ir.Node {
	switch expr := expr.(type) {
	case nil:
		return nil
	case *syntax.ListExpr:
		return p.exprs(expr.ElemList)
	default:
		return []ir.Node{p.expr(expr)}
	}
}

func (p *irgen) exprs(exprs []syntax.Expr) []ir.Node {
	nodes := make([]ir.Node, len(exprs))
	for i, expr := range exprs {
		nodes[i] = p.expr(expr)
	}
	return nodes
}

func (p *irgen) compLit(typ types2.Type, lit *syntax.CompositeLit) *ir.CompLitExpr {
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
func (p *irgen) typeExpr(typ syntax.Expr) ir.Ntype {
	if typ == nil {
		return nil
	}
	return p.expr(typ).(ir.Ntype)
}

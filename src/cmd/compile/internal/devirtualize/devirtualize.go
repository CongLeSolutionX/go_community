// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package devirtualize implements a simple "devirtualization"
// optimization pass, which replaces interface method calls with
// direct concrete-type method calls where possible.
package devirtualize

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/ir"
	"cmd/compile/internal/typecheck"
	"cmd/compile/internal/types"
)

// Func devirtualizes calls within fn where possible.
func Func(fn *ir.Func) {
	ir.CurFunc = fn
	ir.VisitList(fn.Body, func(n ir.Node) {
		if call, ok := n.(*ir.CallExpr); ok {
			Call(call)
		}
	})
}

// Call devirtualizes the given call if possible.
func Call(call *ir.CallExpr) {
	if call.Op() != ir.OCALLINTER {
		return
	}
	sel := call.X.(*ir.SelectorExpr)
	r := ir.StaticValue(sel.X)
	if r.Op() != ir.OCONVIFACE {
		return
	}
	recv := r.(*ir.ConvExpr)

	typ := recv.X.Type()
	if typ.IsInterface() {
		return
	}

	dt := ir.NewTypeAssertExpr(sel.Pos(), sel.X, nil)
	dt.SetType(typ)
	x := typecheck.Callee(ir.NewSelectorExpr(sel.Pos(), ir.OXDOT, dt, sel.Sel))
	switch x.Op() {
	case ir.ODOTMETH:
		x := x.(*ir.SelectorExpr)
		if base.Flag.LowerM > 0 {
			base.WarnfAt(call.Pos(), "devirtualizing %v to %v", sel, typ)
		}
		call.SetOp(ir.OCALLMETH)
		call.X = x
	case ir.ODOTINTER:
		// Promoted method from embedded interface-typed field (#42279).
		x := x.(*ir.SelectorExpr)
		if base.Flag.LowerM > 0 {
			base.WarnfAt(call.Pos(), "partially devirtualizing %v to %v", sel, typ)
		}
		call.SetOp(ir.OCALLINTER)
		call.X = x
	default:
		// TODO(mdempsky): Turn back into Fatalf after more testing.
		if base.Flag.LowerM > 0 {
			base.WarnfAt(call.Pos(), "failed to devirtualize %v (%v)", x, x.Op())
		}
		return
	}

	// Duplicated logic from typecheck for function call return
	// value types.
	//
	// Receiver parameter size may have changed; need to update
	// call.Type to get correct stack offsets for result
	// parameters.
	types.CheckSize(x.Type())
	switch ft := x.Type(); ft.NumResults() {
	case 0:
	case 1:
		call.SetType(ft.Results().Field(0).Type)
	default:
		call.SetType(ft.Results())
	}
}

// For transforms for and range loops that declare variables that might be
// captured by a closure or escaped to the heap.  It returns the list of names
// subject to this change, that may (once transformed) be heap allocated in the
// process.  (This allows checking after escape analysis to call out any such
// variables, in case it causes allocation/performance problems).

func ForCapture(fn *ir.Func) []*ir.Name {

	var transformed []*ir.Name
	dclFixups := make(map[*ir.Name]ir.Stmt)

	ir.CurFunc = fn
	possiblyLeaked := make(map[*ir.Name]bool)

	var do func(x ir.Node) bool
	do = func(n ir.Node) bool {
		switch x := n.(type) {
		case *ir.ClosureExpr:
			for _, cv := range x.Func.ClosureVars {
				v := cv.Canonical()
				if already, ok := possiblyLeaked[v]; ok && !already {
					// base.WarnfAt(x.Pos(), "Closure capture of not-addressed for/range init %v", v)
					possiblyLeaked[v] = true
				}
			}

		case *ir.RangeStmt:
			if !x.Def {
				break
			}
			if n, ok := x.Key.(*ir.Name); ok {
				if n.Addrtaken() {
					possiblyLeaked[n] = true
					// base.WarnfAt(x.Pos(), "Addrtaken range key %v", n)
				} else {
					possiblyLeaked[n] = false
				}
			}
			if n, ok := x.Value.(*ir.Name); ok {
				if n.Addrtaken() {
					possiblyLeaked[n] = true
					// base.WarnfAt(x.Pos(), "Addrtaken range value %v", n)
				} else {
					possiblyLeaked[n] = false
				}
			}
			ir.DoChildren(n, do)
			var k, v, tk, tv *ir.Name
			if n, ok := x.Key.(*ir.Name); ok && possiblyLeaked[n] {
				// Rename the loop key, prefix body with assignment from loop key
				transformed = append(transformed, n)
				tk = typecheck.Temp(n.Type())
				tk.SetTypecheck(1)
				x.Key = tk
				k = n
			}
			if n, ok := x.Value.(*ir.Name); ok && possiblyLeaked[n] {
				// Rename the loop value, prefix body with assignment from loop value
				transformed = append(transformed, n)
				tv = typecheck.Temp(n.Type())
				tv.SetTypecheck(1)
				x.Value = tv
				v = n
			}
			if k != nil {
				as := ir.NewAssignStmt(x.Pos(), k, tk)
				as.Def = true
				as.SetTypecheck(1)
				x.Body.Prepend(as)
				dclFixups[k] = as
			}
			if v != nil {
				as := ir.NewAssignStmt(x.Pos(), v, tv)
				as.Def = true
				as.SetTypecheck(1)
				x.Body.Prepend(as)
				dclFixups[v] = as
			}
			return false

		case *ir.ForStmt:
			for _, s := range x.Init() {
				switch y := s.(type) {
				case *ir.AssignListStmt:
					if !y.Def {
						continue
					}
					for _, z := range y.Lhs {
						if n, ok := z.(*ir.Name); ok {
							if n.Addrtaken() {
								possiblyLeaked[n] = true
							}
						} else {
							possiblyLeaked[n] = false
						}
					}
				case *ir.AssignStmt:
					if !y.Def {
						continue
					}
					if n, ok := y.X.(*ir.Name); ok {
						if n.Addrtaken() {
							possiblyLeaked[n] = true
						} else {
							possiblyLeaked[n] = false
						}
					}
				}
			}
			ir.DoChildren(n, do)
			leaked := []*ir.Name{}
			for _, s := range x.Init() {
				switch y := s.(type) {
				case *ir.AssignListStmt:
					if !y.Def {
						continue
					}
					for _, z := range y.Lhs {
						if n, ok := z.(*ir.Name); ok && possiblyLeaked[n] {
							leaked = append(leaked, n)
						}
					}
				case *ir.AssignStmt:
					if !y.Def {
						continue
					}
					if n, ok := y.X.(*ir.Name); ok && possiblyLeaked[n] {
						leaked = append(leaked, n)
					}
				}
			}
			if len(leaked) > 0 {
				if base.Flag.LowerM < -2 {
					for _, l := range leaked {
						// not transforming yet, just warn for now.
						base.WarnfAt(x.Pos(), "MAYBE %v loop variable leak %v", x.Op(), l)
					}
				}

				// need to transform the for loop just so.

				// 	type ForStmt struct {
				// 	init     Nodes
				// 	Label    *types.Sym
				// 	Cond     Node  // empty if OFORUNTIL
				// 	Late     Nodes // empty if OFOR
				// 	Post     Node
				// 	Body     Nodes
				// 	HasBreak bool
				// }

				// cond' = {}
				// late' = {}
				// prebody = {x := x' for x in leaked}
				// postbody = {x' = x for x in leaked}
				// body_continue = {body : s/continue/goto next}

				// OFOR: init; loop: if !Cond {break}; Body; Post; goto loop
				// init' = (init : s/x/x' for x in leaked) + tmp_first := true
				// body' = prebody +
				//         if tmp_first {tmp_first = false} else {Post} +
				//         if !cond {break} +
				//         body_continue +
				//         next: postbody
				// post' = {}

				// OFORUNTIL: init; loop: Body; Post; if Late {goto loop}
				// init' = (init : s/x/x' for x in leaked)
				// body' = prebody +
				//         body_continue +
				//         next: post +
				//         if !late { break }
				//         postbody

			}

			return false
		}

		ir.DoChildren(n, do)

		return false
	}
	do(fn)
	if len(transformed) > 0 {
		editNodes := func(x ir.Node, c ir.Nodes) ir.Nodes {
			j := 0
			for i, n := range c {
				if d, ok := n.(*ir.Decl); ok {
					if s := dclFixups[d.X]; s != nil {
						switch l := s.(type) {
						case *ir.AssignStmt:
							l.PtrInit().Prepend(d)
							delete(dclFixups, d.X) // can't be sure of visit order, wouldn't want to visit twice.
						default:
							base.Fatalf("not implemented yet for node type %v", s.Op)
						}
						continue // do not copy this node, and do not increment j
					}
				}
				if j != i {
					c[j] = c[i]
				}
				j++
			}
			for k := j; k < len(c); k++ {
				c[k] = nil
			}
			return c[:j]
		}
		rewriteNodes(fn, editNodes)
	}
	return transformed
}

func rewriteNodes(fn *ir.Func, editNodes func(n ir.Node, c ir.Nodes) ir.Nodes) {
	var forNodes func(x ir.Node) bool
	forNodes = func(n ir.Node) bool {
		if stmt, ok := n.(ir.InitNode); ok {
			// process init list
			stmt.SetInit(editNodes(n, stmt.Init()))
		}
		switch x := n.(type) {
		case *ir.Func:
			x.Body = editNodes(n, x.Body)
			x.Enter = editNodes(n, x.Enter)
			x.Exit = editNodes(n, x.Exit)
		case *ir.InlinedCallExpr:
			x.Body = editNodes(n, x.Body)

		case *ir.CaseClause:
			x.Body = editNodes(n, x.Body)
		case *ir.CommClause:
			x.Body = editNodes(n, x.Body)

		case *ir.BlockStmt:
			x.List = editNodes(n, x.List)

		case *ir.ForStmt:
			x.Body = editNodes(n, x.Body) // TODO Late?
		case *ir.RangeStmt:
			x.Body = editNodes(n, x.Body)
		case *ir.IfStmt:
			x.Body = editNodes(n, x.Body)
			x.Else = editNodes(n, x.Else)
		case *ir.SelectStmt:
			x.Compiled = editNodes(n, x.Compiled)
		case *ir.SwitchStmt:
			x.Compiled = editNodes(n, x.Compiled)
		}
		ir.DoChildren(n, forNodes)
		return false
	}
	forNodes(fn)
}

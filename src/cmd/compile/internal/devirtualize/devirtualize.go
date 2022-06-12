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
	"fmt"
)

// Func devirtualizes calls within fn where possible.
func Func(fn *ir.Func) {
	ir.CurFunc = fn

	// For promoted methods (including value-receiver methods promoted to pointer-receivers),
	// the interface method wrapper may contain expressions that can panic (e.g., ODEREF, ODOTPTR, ODOTINTER).
	// Devirtualization involves inlining these expressions (and possible panics) to the call site.
	// This normally isn't a problem, but for go/defer statements it can move the panic from when/where
	// the call executes to the go/defer statement itself, which is a visible change in semantics (e.g., #52072).
	// To prevent this, we skip devirtualizing calls within go/defer statements altogether.
	goDeferCall := make(map[*ir.CallExpr]bool)
	ir.VisitList(fn.Body, func(n ir.Node) {
		switch n := n.(type) {
		case *ir.GoDeferStmt:
			if call, ok := n.Call.(*ir.CallExpr); ok {
				goDeferCall[call] = true
			}
			return
		case *ir.CallExpr:
			if !goDeferCall[n] {
				Call(n)
			}
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

	if base.Debug.Unified != 0 {
		// N.B., stencil.go converts shape-typed values to interface type
		// using OEFACE instead of OCONVIFACE, so devirtualization fails
		// above instead. That's why this code is specific to unified IR.

		// If typ is a shape type, then it was a type argument originally
		// and we'd need an indirect call through the dictionary anyway.
		// We're unable to devirtualize this call.
		if typ.IsShape() {
			return
		}

		// If typ *has* a shape type, then it's an shaped, instantiated
		// type like T[go.shape.int], and its methods (may) have an extra
		// dictionary parameter. We could devirtualize this call if we
		// could derive an appropriate dictionary argument.
		//
		// TODO(mdempsky): If typ has has a promoted non-generic method,
		// then that method won't require a dictionary argument. We could
		// still devirtualize those calls.
		//
		// TODO(mdempsky): We have the *runtime.itab in recv.TypeWord. It
		// should be possible to compute the represented type's runtime
		// dictionary from this (e.g., by adding a pointer from T[int]'s
		// *runtime._type to .dict.T[int]; or by recognizing static
		// references to go:itab.T[int],iface and constructing a direct
		// reference to .dict.T[int]).
		if typ.HasShape() {
			if base.Flag.LowerM > 0 {
				base.WarnfAt(call.Pos(), "cannot devirtualize %v: shaped receiver %v", call, typ)
			}
			return
		}

		// Further, if sel.X's type has a shape type, then it's a shaped
		// interface type. In this case, the (non-dynamic) TypeAssertExpr
		// we construct below would attempt to create an itab
		// corresponding to this shaped interface type; but the actual
		// itab pointer in the interface value will correspond to the
		// original (non-shaped) interface type instead. These are
		// functionally equivalent, but they have distinct pointer
		// identities, which leads to the type assertion failing.
		//
		// TODO(mdempsky): We know the type assertion here is safe, so we
		// could instead set a flag so that walk skips the itab check. For
		// now, punting is easy and safe.
		if sel.X.Type().HasShape() {
			if base.Flag.LowerM > 0 {
				base.WarnfAt(call.Pos(), "cannot devirtualize %v: shaped interface %v", call, sel.X.Type())
			}
			return
		}
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

// ForCapture transforms for and range loops that declare variables that might be
// captured by a closure or escaped to the heap. It returns the list of names
// subject to this change, that may (once transformed) be heap allocated in the
// process. (This allows checking after escape analysis to call out any such
// variables, in case it causes allocation/performance problems).
// Debug flags:
// -m < 0 report transformed escapes
// -m < -1 report transformed non-escapes
// -m < -2 report 3-clause for loops that may leak variables
// -m == -4 disable 3-clause transformation
// -m == -5 disable all transformation and reporting
func ForCapture(fn *ir.Func) []*ir.Name {
	if base.Flag.LowerM == -5 {
		return nil
	}
	seq := 1

	var transformed []*ir.Name
	dclFixups := make(map[*ir.Name]ir.Stmt)

	ir.CurFunc = fn
	possiblyLeaked := make(map[*ir.Name]bool)

	noteMayLeak := func(x ir.Node) {
		if n, ok := x.(*ir.Name); ok {
			possiblyLeaked[n] = false
		}
	}

	maybeReplaceVar := func(k ir.Node, x *ir.RangeStmt) ir.Node {
		if n, ok := k.(*ir.Name); ok && possiblyLeaked[n] {
			// Rename the loop key, prefix body with assignment from loop key
			transformed = append(transformed, n)
			tk := typecheck.Temp(n.Type())
			tk.SetTypecheck(1)
			as := ir.NewAssignStmt(x.Pos(), n, tk)
			as.Def = true
			as.SetTypecheck(1)
			x.Body.Prepend(as)
			dclFixups[n] = as
			return tk
		}
		return k
	}

	var do func(x ir.Node) bool
	do = func(n ir.Node) bool {
		switch x := n.(type) {
		case *ir.ClosureExpr:
			for _, cv := range x.Func.ClosureVars {
				v := cv.Canonical()
				if already, ok := possiblyLeaked[v]; ok && !already {
					possiblyLeaked[v] = true
				}
			}

		case *ir.AddrExpr:
			// Explicitly note address-taken so that return-statements can be excluded
			y := ir.OuterValue(x.X)
			if y.Op() != ir.ONAME {
				break
			}
			z, ok := y.(*ir.Name)
			if !ok {
				break
			}
			if b, ok := possiblyLeaked[z]; ok && !b {
				possiblyLeaked[z] = true
			}

		case *ir.ReturnStmt:
			// Address-taking and closures under return statements do not count
			return false

		case *ir.RangeStmt:
			if !x.Def {
				break
			}
			noteMayLeak(x.Key)
			noteMayLeak(x.Value)
			ir.DoChildren(n, do)
			x.Key = maybeReplaceVar(x.Key, x)
			x.Value = maybeReplaceVar(x.Value, x)
			return false

		case *ir.ForStmt:
			for _, s := range x.Init() {
				switch y := s.(type) {
				case *ir.AssignListStmt:
					if !y.Def {
						continue
					}
					for _, z := range y.Lhs {
						noteMayLeak(z)
					}
				case *ir.AssignStmt:
					if !y.Def {
						continue
					}
					noteMayLeak(y.X)
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
						base.WarnfAt(x.Pos(), "MAYBE %v 3-clause loop variable leak %v", x.Op(), l)
					}
				}

				if base.Flag.LowerM != -4 && base.DebugHashMatch(ir.PkgFuncName(fn)) {
					// need to transform the for loop just so.

					// 	type ForStmt struct {
					// 	init     Nodes
					// 	Label    *types.Sym
					// 	Cond     Node  // empty if OFORUNTIL
					// 	Post     Node
					// 	Body     Nodes
					// 	HasBreak bool
					// }

					// OFOR: init; loop: if !Cond {break}; Body; Post; goto loop

					// (1) prebody = {z := z' for z in leaked}
					// (2) postbody = {z' = z for z in leaked}
					// (3) body_continue = {body : s/continue/goto next}
					// (4) init' = (init : s/z/z' for z in leaked) + tmp_first := true
					// (5) body' = prebody +        // appears out of order below
					// (6)         if tmp_first {tmp_first = false} else {Post} +
					// (7)         if !cond {break} +
					// (8)         body_continue +
					// (9)         next: postbody
					// (10) cond' = {}
					// (11) post' = {}

					// minor optimization -- if Post is empty, tmp_first and step 7 can be skipped.

					var preBody, postBody ir.Nodes

					tempFor := make(map[*ir.Name]*ir.Name)

					// (1,2) initialize preBody and postBody
					for _, z := range leaked {

						transformed = append(transformed, z)

						tz := typecheck.Temp(z.Type())
						tz.SetTypecheck(1)
						tempFor[z] = tz

						as := ir.NewAssignStmt(x.Pos(), z, tz)
						as.Def = true
						as.SetTypecheck(1)
						preBody.Append(as)
						dclFixups[z] = as

						as = ir.NewAssignStmt(x.Pos(), tz, z)
						as.SetTypecheck(1)
						postBody.Append(as)

					}

					// (3) rewrite continues in body -- rewrite is inplace, so works for top level visit, too.
					label := typecheck.Lookup(fmt.Sprintf(".3clNext_%d", seq))
					seq++
					labelStmt := ir.NewLabelStmt(x.Pos(), label)
					labelStmt.SetTypecheck(1)

					loopLabel := x.Label
					loopDepth := 0
					var editContinues func(x ir.Node) bool
					editContinues = func(x ir.Node) bool {

						switch c := x.(type) {
						case *ir.BranchStmt:
							if c.Op() == ir.OCONTINUE && (loopDepth == 0 || loopLabel != nil && c.Label == loopLabel) {
								c.Label = label
								c.SetOp(ir.OGOTO)
							}
						case *ir.RangeStmt, *ir.ForStmt:
							loopDepth++
							ir.DoChildren(x, editContinues)
							loopDepth--
							return false
						}
						ir.DoChildren(x, editContinues)
						return false
					}
					for _, y := range x.Body {
						editContinues(y)
					}
					body_continue := x.Body

					// (4) rewrite init
					for _, s := range x.Init() {
						switch y := s.(type) {
						case *ir.AssignListStmt:
							if !y.Def {
								continue
							}
							for i, z := range y.Lhs {
								if n, ok := z.(*ir.Name); ok && possiblyLeaked[n] {
									y.Lhs[i] = tempFor[n]
								}
							}
						case *ir.AssignStmt:
							if !y.Def {
								continue
							}
							if n, ok := y.X.(*ir.Name); ok && possiblyLeaked[n] {
								y.X = tempFor[n]
							}
						}
					}

					var tmpFirstDcl *ir.AssignStmt
					postNotNil := x.Post != nil
					if postNotNil {
						// body' = prebody +
						// (6)     if tmp_first {tmp_first = false} else {Post} +
						//         if !cond {break} + ...
						tmpFirst := typecheck.Temp(types.Types[types.TBOOL])
						tmpFirstDcl = ir.NewAssignStmt(x.Pos(), tmpFirst, typecheck.OrigBool(tmpFirst, true))
						tmpFirstDcl.Def = true
						tmpFirstDcl.SetTypecheck(1)

						tmpFirstGetsFalse := ir.NewAssignStmt(x.Pos(), tmpFirst, typecheck.OrigBool(tmpFirst, false))
						tmpFirstGetsFalse.SetTypecheck(1)
						ifTmpFirst := ir.NewIfStmt(x.Pos(), tmpFirst, ir.Nodes{tmpFirstGetsFalse}, ir.Nodes{x.Post})
						ifTmpFirst.SetTypecheck(1)
						preBody.Append(ifTmpFirst)
					}

					// body' = prebody +
					//         if tmp_first {tmp_first = false} else {Post} +
					// (7)     if !cond {break} + ...
					notCond := ir.NewUnaryExpr(x.Cond.Pos(), ir.ONOT, x.Cond)
					notCond.SetType(x.Cond.Type())
					notCond.SetTypecheck(1)
					newBreak := ir.NewBranchStmt(x.Pos(), ir.OBREAK, nil)
					newBreak.SetTypecheck(1)
					ifNotCond := ir.NewIfStmt(x.Pos(), notCond, ir.Nodes{newBreak}, nil)
					ifNotCond.SetTypecheck(1)
					preBody.Append(ifNotCond)

					if postNotNil {
						x.PtrInit().Append(tmpFirstDcl)
					}

					// (8)
					preBody.Append(body_continue...)
					// (9)
					preBody.Append(labelStmt)
					preBody.Append(postBody...)

					// (5) body' = prebody + ...
					x.Body = preBody

					// (10) cond' = {}
					x.Cond = nil

					// (11) post' = {}
					x.Post = nil
				}
			}

			return false
		}

		ir.DoChildren(n, do)

		return false
	}
	do(fn)
	if len(transformed) > 0 {
		editNodes := func(c ir.Nodes) ir.Nodes {
			j := 0
			for i, n := range c {
				if d, ok := n.(*ir.Decl); ok {
					if s := dclFixups[d.X]; s != nil {
						switch l := s.(type) {
						case *ir.AssignStmt:
							l.PtrInit().Prepend(d)
							delete(dclFixups, d.X) // can't be sure of visit order, wouldn't want to visit twice.
						default:
							base.Fatalf("not implemented yet for node type %v", s.Op())
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

func rewriteNodes(fn *ir.Func, editNodes func(c ir.Nodes) ir.Nodes) {
	var forNodes func(x ir.Node) bool
	forNodes = func(n ir.Node) bool {
		if stmt, ok := n.(ir.InitNode); ok {
			// process init list
			stmt.SetInit(editNodes(stmt.Init()))
		}
		switch x := n.(type) {
		case *ir.Func:
			x.Body = editNodes(x.Body)
			x.Enter = editNodes(x.Enter)
			x.Exit = editNodes(x.Exit)
		case *ir.InlinedCallExpr:
			x.Body = editNodes(x.Body)

		case *ir.CaseClause:
			x.Body = editNodes(x.Body)
		case *ir.CommClause:
			x.Body = editNodes(x.Body)

		case *ir.BlockStmt:
			x.List = editNodes(x.List)

		case *ir.ForStmt:
			x.Body = editNodes(x.Body) // TODO Late?
		case *ir.RangeStmt:
			x.Body = editNodes(x.Body)
		case *ir.IfStmt:
			x.Body = editNodes(x.Body)
			x.Else = editNodes(x.Else)
		case *ir.SelectStmt:
			x.Compiled = editNodes(x.Compiled)
		case *ir.SwitchStmt:
			x.Compiled = editNodes(x.Compiled)
		}
		ir.DoChildren(n, forNodes)
		return false
	}
	forNodes(fn)
}

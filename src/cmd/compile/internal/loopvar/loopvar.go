// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package loopvar applies the proper variable capture, according
// to experiment, flags, language version, etc.
package loopvar

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/ir"
	"cmd/compile/internal/typecheck"
	"cmd/compile/internal/types"
	"fmt"
)

// ForCapture transforms for and range loops that declare variables that might be
// captured by a closure or escaped to the heap. It returns the list of names
// subject to this change, that may (once transformed) be heap allocated in the
// process. (This allows checking after escape analysis to call out any such
// variables, in case it causes allocation/performance problems).

// Debug flags:
//
// base.LoopVarHash != nil => use hash setting to govern transformation.
// note that LoopVarHash != nil sets base.Debug.LoopVar to 1 (unless it is >= 11, for testing/debugging).
//
// base.Debug.LoopVar == 0 => do nothing unless base.LoopVarHash != nil
// base.Debug.LoopVar == 1 => transform (may be set by GOEXPERIMENT)
// base.Debug.LoopVar == 2 => transform and log results (can be in addition to GOEXPERIMENT)
// base.Debug.LoopVar == 11 => transform ALL loops ignoring syntactic/potential escape.  Do not log, can be in addition to GOEXPERIMENT.
// base.Debug.LoopVar == 12 => 11, but log results
// base.Debug.LoopVar == 13 => 12 plus internal debugging
// base.Debug.LoopVar == -1 => do not transform
//

type NameFn struct {
	Name *ir.Name
	Fn   *ir.Func
}

func ForCapture(fn *ir.Func) []NameFn {
	if base.Debug.LoopVar <= 0 { // code in base:flags.go ensures >= 1 if loopvarhash != ""
		// TODO remove this when the transformation is made sensitive to inlining; this is least-risk for 1.21
		return nil
	}
	seq := 1

	// if a loop variable is transformed it is appended to this slice for later logging
	var transformed []NameFn
	dclFixups := make(map[*ir.Name]ir.Stmt)

	ir.CurFunc = fn

	// possibly leaked includes names of declared loop variables that may be leaked;
	// the mapped value is true if the name is *syntactically* leaked, and those loops
	// will be transformed.
	possiblyLeaked := make(map[*ir.Name]bool)

	// noteMayLeak is called for candidate variables in for range/3-clause.
	noteMayLeak := func(x ir.Node) {
		if n, ok := x.(*ir.Name); ok {
			if n.Type().Kind() == types.TBLANK {
				return
			}
			// default is false (leak candidate, not yet known to leak), but flag can make all variables "leak"
			possiblyLeaked[n] = base.Debug.LoopVar >= 11
		}
	}

	maybeReplaceVar := func(k ir.Node, x *ir.RangeStmt) ir.Node {
		if n, ok := k.(*ir.Name); ok && possiblyLeaked[n] {
			if base.LoopVarHash == nil ||
				base.LoopVarHash.DebugHashMatchPos(base.Ctxt, n.Pos()) {
				// Rename the loop key, prefix body with assignment from loop key
				transformed = append(transformed, NameFn{n, fn})
				tk := typecheck.Temp(n.Type())
				tk.SetTypecheck(1)
				as := ir.NewAssignStmt(x.Pos(), n, tk)
				as.Def = true
				as.SetTypecheck(1)
				x.Body.Prepend(as)
				dclFixups[n] = as
				return tk
			}
		}
		return k
	}

	// forAllDefInInitUpdate applies "do" to all the defining assignemnts in the Init clause of a ForStmt.
	// This abstracts away some of the boilerplate from the already complex and verbose for-3-clause case.
	forAllDefInInitUpdate := func(x *ir.ForStmt, do func(z ir.Node, update *ir.Node)) {
		for _, s := range x.Init() {
			switch y := s.(type) {
			case *ir.AssignListStmt:
				if !y.Def {
					continue
				}
				for i, z := range y.Lhs {
					do(z, &y.Lhs[i])
				}
			case *ir.AssignStmt:
				if !y.Def {
					continue
				}
				do(y.X, &y.X)
			}
		}
	}

	// forAllDefInInit is forAllDefInInitUpdate without the update option.
	forAllDefInInit := func(x *ir.ForStmt, do func(z ir.Node)) {
		forAllDefInInitUpdate(x, func(z ir.Node, _ *ir.Node) { do(z) })
	}

	// scanChildrenThenTransform processes node x to:
	//  1. if x is a for/range, note declared iteration variables PL
	//  2. search all of x's children for syntactically escaping references to v in PL
	//  3. for all v in PL that had a syntactically escaping reference, transform the declaration
	//     and (in case of 3-clause loop) the loop to the unshared loop semantics.
	//  This is all much simpler for range loops; 3-clause loops can have an arbitrary number
	//  of iteration variables and the transformation is more involved, range loops have at most 2.
	var scanChildrenThenTransform func(x ir.Node) bool
	scanChildrenThenTransform = func(n ir.Node) bool {
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
			// this optimization shows up in benchmarks, that's why it's here.
			// Syntactic escapes in code post-dominated by break/continue-outer
			// are harder because those might apply to a containing loop.
			return false

		case *ir.RangeStmt:
			if !x.Def {
				break
			}
			noteMayLeak(x.Key)
			noteMayLeak(x.Value)
			ir.DoChildren(n, scanChildrenThenTransform)
			x.Key = maybeReplaceVar(x.Key, x)
			x.Value = maybeReplaceVar(x.Value, x)
			return false

		case *ir.ForStmt:
			forAllDefInInit(x, noteMayLeak)
			ir.DoChildren(n, scanChildrenThenTransform)
			leaked := []*ir.Name{}
			// Collect the leaking variables for the much-more-complex transformation.
			forAllDefInInit(x, func(z ir.Node) {
				if n, ok := z.(*ir.Name); ok && possiblyLeaked[n] {
					// Hash on n.Pos() for most precise failure location.
					if base.LoopVarHash == nil ||
						base.LoopVarHash.DebugHashMatchPos(base.Ctxt, n.Pos()) {
						leaked = append(leaked, n)
					}
				}
			})

			if len(leaked) > 0 {
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

				// minor optimizations:
				//   if Post is empty, tmp_first and step 7 can be skipped.
				//   if Cond is empty, that code can also be skipped.

				var preBody, postBody ir.Nodes

				tempFor := make(map[*ir.Name]*ir.Name)

				// (1,2) initialize preBody and postBody
				for _, z := range leaked {

					transformed = append(transformed, NameFn{z, fn})

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
						if c.Op() == ir.OCONTINUE && (loopDepth == 0 && c.Label == nil || loopLabel != nil && c.Label == loopLabel) {
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
				forAllDefInInitUpdate(x, func(z ir.Node, pz *ir.Node) {
					// note tempFor[n] can be nil if hash searching.
					if n, ok := z.(*ir.Name); ok && possiblyLeaked[n] && tempFor[n] != nil {
						*pz = tempFor[n]
					}
				})

				postNotNil := x.Post != nil
				var tmpFirstDcl *ir.AssignStmt
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
				if x.Cond != nil {
					notCond := ir.NewUnaryExpr(x.Cond.Pos(), ir.ONOT, x.Cond)
					notCond.SetType(x.Cond.Type())
					notCond.SetTypecheck(1)
					newBreak := ir.NewBranchStmt(x.Pos(), ir.OBREAK, nil)
					newBreak.SetTypecheck(1)
					ifNotCond := ir.NewIfStmt(x.Pos(), notCond, ir.Nodes{newBreak}, nil)
					ifNotCond.SetTypecheck(1)
					preBody.Append(ifNotCond)
				}

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

			return false
		}

		ir.DoChildren(n, scanChildrenThenTransform)

		return false
	}
	scanChildrenThenTransform(fn)
	if len(transformed) > 0 {
		// editNodes scans a slice C of ir.Node, looking for declarations that
		// appear in dclFixups.  Any declaration D whose "fixup" is an assignmnt
		// statement A is removed from the C and relocated to the Init
		// of A.  editNodes returns the modified slice of ir.Node.
		editNodes := func(c ir.Nodes) ir.Nodes {
			j := 0
			for i, n := range c {
				if d, ok := n.(*ir.Decl); ok {
					if s := dclFixups[d.X]; s != nil {
						switch a := s.(type) {
						case *ir.AssignStmt:
							a.PtrInit().Prepend(d)
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
		// fixup all tagged declarations in all the statements lists in fn.
		rewriteNodes(fn, editNodes)
	}
	return transformed
}

// rewriteNodes applies editNodes to all statement lists in fn.
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

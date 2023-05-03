// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// The code specialization uses profile information to optimize interface method calls.
// An interface call is converted into an "if-else" code block based on the hot dynamic type.
// We use IRGraph and Profile data structures to determine the sole hot dynamic type.
// The transformed code introduces an "if" condition check on the runtime type.
// The body of the "then" block converts the interface method call into a direct function call. This call should be inlined during Inlining  which follows this optimization.
// The body of the "else" block retains the original slow interface method call.
// More details can be found at: https://github.com/golang/proposal/blob/master/design/55022-pgo-implementation.md

package pgo

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/ir"
	"cmd/compile/internal/logopt"
	"cmd/compile/internal/typecheck"
	"cmd/compile/internal/types"
	"fmt"
)

// Specializer is the main driver for code specialization using profile data.
func Specializer(p *Profile) {
	if p.IfaceCallMap == nil || len(p.IfaceCallMap) == 0 {
		if base.Flag.LowerM > 1 {
			fmt.Printf("Specializer: nothing to specialize\n")
		}
		return
	}

	ir.VisitFuncsBottomUp(typecheck.Target.Decls, func(list []*ir.Func, recursive bool) {
		for _, f := range list {
			SpecializeIfaceCallSites(f, p)
		}
	})

}

// CanSpecialize checks if the target function can be specialized. The checks
// are similar to that of inlining since our end goal is that the specialized
// method can be inlined.
func CanSpecialize(fn *ir.Func) bool {
	if fn.Nname == nil {
		base.Fatalf("CanSpecialize no nname %+v", fn)
	}
	var reason string
	if base.Flag.LowerM > 1 || logopt.Enabled() {
		defer func() {
			if reason != "" {
				if base.Flag.LowerM > 1 {
					fmt.Printf("Line %v: cannot specialize %v: %s\n", ir.Line(fn), fn.Nname, reason)
				}
				if logopt.Enabled() {
					logopt.LogOpt(fn.Pos(), ": cannot specialize function", "inline", ir.FuncName(fn), reason)
				}
			}
		}()
	}

	reason = CanInlineOrSpecialize(fn)
	if reason != "" {
		return false
	}

	if fn.Typecheck() == 0 {
		base.Fatalf("Specializer: CanSpecialize on non-typechecked function %v", fn)
	}

	return true
}

// SpecializeIfaceCallSites specializes calls using the profile information by
// walking over the body of fn.
func SpecializeIfaceCallSites(fn *ir.Func, p *Profile) bool {
	changed := false
	lineMap := make(map[int]int)
	countIfaceMethodCallsPerLine(fn, lineMap)
	var enclosingStmtNode ir.Node
	var enclosingStmtLn int
	ir.CurFunc = fn
	ir.VisitList(fn.Body, func(n ir.Node) {
		if isEnclosingStmt(n) {
			enclosingStmtNode = n
			enclosingStmtLn = NodeLineOffset(n, fn)
		}
		if call, ok := n.(*ir.CallExpr); ok {
			if call.Op() == ir.OCALLINTER {
				// Ensure that the interface method call `call` is enclosed in a
				// statement node and that the line offsets match.
				if enclosingStmtNode == nil {
					return
				}
				line := NodeLineOffset(n, fn)
				if line != enclosingStmtLn {
					return
				}
				// Bail, if we do not have a hot callee.
				f := p.findHotConcreteCallee(fn, n)
				if f == nil {
					return
				}
				// Bail, if we do not have a Type node for the hot callee.
				ctyp := p.findConcreteType(ir.PkgFuncName(f))
				if ctyp == nil {
					return
				}
				if ctyp.IsInterface() {
					return
				}
				// Bail, if we can not inline the hot callee proactively.
				if !CanSpecialize(f) {
					return
				}
				// Prevent specialization on lines with more than one iface method specialization opportunities.
				if count, ok := lineMap[line]; ok {
					if count == 1 {
						ret := SpecializeACallSite(call, fn, ctyp, enclosingStmtNode)
						changed = changed || ret
					} else if count > 1 {
						if base.Flag.LowerM != 0 {
							fmt.Printf("%v: cannot specialize hot interface method call %v\n", ir.Line(n), call)
						}
					}
				}
			}
		}
	})
	return changed
}

// SpecializeACallSite specializes the given call using a direct method call to
// concretenode.
func SpecializeACallSite(call *ir.CallExpr, curfn *ir.Func, concretetyp *types.Type, enclosingStmtNode ir.Node) bool {
	newnode := rewriteASTNode(enclosingStmtNode, curfn, concretetyp)
	if newnode == nil {
		return false
	}
	tmpnode := typecheck.Temp(concretetyp)
	tmpcond := typecheck.Temp(types.Types[types.TBOOL])
	r := createIfStmt(call, enclosingStmtNode, newnode, concretetyp, tmpnode, tmpcond)
	ReplaceNode(curfn, r, enclosingStmtNode)
	if base.Flag.LowerM != 0 {
		fmt.Printf("%v: specializing %v: %v\n", ir.Line(enclosingStmtNode), enclosingStmtNode, r)
	}
	return true
}

// createIfStmt creates an if-stmt surrounding the direct method call.
func createIfStmt(call *ir.CallExpr, oldnode ir.Node, newnode ir.Node, concretetyp *types.Type, tmpnode *ir.Name, tmpcond *ir.Name) ir.Node {
	sel := call.X.(*ir.SelectorExpr)
	dt := ir.NewTypeAssertExpr(sel.Pos(), sel.X, nil)
	dt.SetType(concretetyp)
	dt.SetOp(ir.ODOTTYPE2)
	dt.SetTypecheck(1)
	r := ir.NewIfStmt(sel.Pos(), nil, nil, nil)
	aslist := ir.NewAssignListStmt(sel.Pos(), ir.OAS2DOTTYPE, []ir.Node{tmpnode, tmpcond}, []ir.Node{dt})
	aslist.Def = true
	aslist.SetTypecheck(1)
	r.PtrInit().Append(typecheck.Stmt(aslist))
	r.Else = []ir.Node{oldnode}
	r.Cond = typecheck.Expr(tmpcond)
	r.Body = []ir.Node{newnode}
	r.SetTypecheck(1)
	return r
}

// isEnclosingStmt checks if the statement surrounding the interface method is specializable.
func isEnclosingStmt(node ir.Node) bool {
	// These are a subset of nodes that we decide to specialize. This list can
	// be extended in the future.
	switch node.Op() {
	case ir.OAS2FUNC, ir.OAS2RECV, ir.OAS2MAPR, ir.OAS2DOTTYPE:
		as := node.(*ir.AssignListStmt)
		if as.Def {
			return false
		}
	case ir.OAS:
		as := node.(*ir.AssignStmt)
		if as.Def {
			return false
		}
	case ir.OINDEX, ir.OEFACE, ir.OAND, ir.OANDNOT, ir.OASOP,
		ir.OSUB, ir.OMUL, ir.OADD, ir.OOR, ir.OXOR, ir.OLSH, ir.ORSH, ir.OUNSAFEADD, ir.OCOMPLEX, ir.OEQ, ir.ONE,
		ir.OLT, ir.OLE, ir.OGT, ir.OGE, ir.ONOT, ir.ONEG, ir.OPLUS, ir.OBITNOT, ir.OREAL, ir.OIMAG,
		ir.OSPTR, ir.OITAB, ir.OIDATA, ir.OADDSTR, ir.OANDAND, ir.OOROR, ir.ORETURN:
		return true
	default:
		return false
	}
	return true
}

// rewriteASTNode takes an expression node that contains an interface call
// (ir.OCALLINTER) and devirtualizes that call while keeping the remaining AST
// nodes untouched.
func rewriteASTNode(newNode ir.Node, fn *ir.Func, concretetyp *types.Type) ir.Node {
	cn := ir.DeepCopy(newNode.Pos(), newNode)
	var edit func(ir.Node) ir.Node
	edit = func(n ir.Node) ir.Node {
		if n == nil {
			return n
		}
		ir.EditChildren(n, edit)
		if call, ok := n.(*ir.CallExpr); ok {
			if call.Op() == ir.OCALLINTER {
				return mkcallnode(call, fn, concretetyp)
			}
		}
		return n
	}
	ir.EditChildren(cn, edit)
	return cn
}

// mkcallnode creates the direct call node with arguments while retaining the
// original statement.
func mkcallnode(call *ir.CallExpr, curfn *ir.Func, concretetyp *types.Type) ir.Node {
	if call.Op() != ir.OCALLINTER {
		return call
	}
	sel := call.X.(*ir.SelectorExpr)
	dt := ir.NewTypeAssertExpr(sel.Pos(), sel.X, nil)
	dt.SetType(concretetyp)
	x := typecheck.Callee(ir.NewSelectorExpr(sel.Pos(), ir.OXDOT, dt, sel.Sel))
	call1 := ir.NewCallExpr(sel.Pos(), ir.OCALL, nil, call.Args)

	switch x.Op() {
	case ir.ODOTMETH:
		x := x.(*ir.SelectorExpr)
		if base.Flag.LowerM > 1 {
			base.WarnfAt(call.Pos(), "specializing %v to %v", sel, concretetyp)
		}
		call1.SetOp(ir.OCALLMETH)
		call1.X = x
	case ir.ODOTINTER:
		x := x.(*ir.SelectorExpr)
		if base.Flag.LowerM > 1 {
			base.WarnfAt(call.Pos(), "partially specializing %v to %v", sel, concretetyp)
		}
		call1.SetOp(ir.OCALLINTER)
		call1.X = x
	default:
		if base.Flag.LowerM > 1 {
			base.WarnfAt(call.Pos(), "Specializer:: failed to specialize %v (%v)", x, x.Op())
		}
		return call
	}
	call1.SetTypecheck(1)
	types.CheckSize(x.Type())
	switch ft := x.Type(); ft.NumResults() {
	case 0:
	case 1:
		call1.SetType(ft.Results().Field(0).Type)
	default:
		call1.SetType(ft.Results())
	}
	typecheck.FixMethodCall(call1)
	return call1
}

// ReplaceNode replaces the old node containing virtual method calls with a new
// specialized node.
func ReplaceNode(fn *ir.Func, rep ir.Node, node ir.Node) {
	ir.CurFunc = fn
	var edit func(ir.Node) ir.Node
	line := ir.Line(node)
	edit = func(n ir.Node) ir.Node {
		ir.EditChildren(n, edit)
		if ir.Line(n) == line {
			return rep
		}
		return n
	}
	ir.EditChildren(fn, edit)
}

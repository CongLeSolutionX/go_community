// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// TODO: update
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
func SpecializeIfaceCallSites(fn *ir.Func, p *Profile) {
	if base.Flag.LowerM > 2 {
		fmt.Printf("%s before specialization: %+v\n", ir.LinkFuncName(fn), fn)
	}

	ir.CurFunc = fn
	var edit func(n ir.Node) ir.Node
	edit = func(n ir.Node) ir.Node {
		//fmt.Printf("node: %+v\n", n)

		if n == nil {
			return n
		}

		ir.EditChildren(n, edit)

		call, ok := n.(*ir.CallExpr)
		if !ok {
			return n
		}
		if call.Op() != ir.OCALLINTER {
			return n
		}

		//fmt.Printf("Call: %+v\n", call)
		//fmt.Printf("Enclosing statement: %+v\n", enclosingStmtNode)
		// Bail, if we do not have a hot callee.
		f := p.findHotConcreteCallee(fn, n)
		if f == nil {
			//fmt.Printf("no callee\n")
			return n
		}
		// Bail, if we do not have a Type node for the hot callee.
		ctyp := typeOfMethodParent(f)
		if ctyp == nil {
			//fmt.Printf("no type\n")
			return n
		}
		if ctyp.IsInterface() {
			//fmt.Printf("concrete type is an interface? %+v\n", ctyp)
			return n
		}
		// Bail, if we can not inline the hot callee proactively.
		if !CanSpecialize(f) {
			//fmt.Printf("can't specialize %+v\n", f)
			return n
		}
		return SpecializeACallSite(n, call, fn, ctyp)
	}

	ir.EditChildren(fn, edit)

	if base.Flag.LowerM > 2 {
		fmt.Printf("%s after specialization: %+v\n", ir.LinkFuncName(fn), fn)
	}
}

// SpecializeACallSite specializes the given call using a direct method call to
// concretenode.
func SpecializeACallSite(n ir.Node, call *ir.CallExpr, curfn *ir.Func, concretetyp *types.Type) ir.Node {
	if base.Flag.LowerM > 2 {
		fmt.Printf("Specializing call to %+v. Before: %+v\n", concretetyp, n)
	}

	// TODO: move into package inline?

	// OINCALL of:
	//
	// var ret1 R1
	// var retN RN
	//
	// var arg1 A1 = arg1 expr
	// var argN AN = argN expr
	//
	// t, ok := sel.(Concrete)
	// if ok {
	//   ret1, retN = t.Method(arg1, ... argN)
	// } else {
	//   ret1, retN = sel.Method(arg1, ... argN)
	// }
	//
	// OINCALL retvars: ret1, ... retN

	var retvars []ir.Node

	sig := call.X.Type()
	for _, ret := range sig.Results().FieldSlice() {
		retvars = append(retvars, typecheck.Temp(ret.Type))
	}

	sel := call.X.(*ir.SelectorExpr)
	pos := call.Pos()
	init := ir.TakeInit(call)

	// Move arguments to assignments prior to the if statement. We cannot
	// simply copy the args' IR, as some IR constructs cannot be copied,
	// such as labels (possible in InlinedCall nodes).
	args := call.Args.Take()
	argvars := make([]ir.Node, 0, len(args))
	for _, arg := range args {
		argvar := typecheck.Temp(arg.Type())
		argvars = append(argvars, argvar)

		assign := ir.NewAssignStmt(pos, argvar, arg)
		assign.SetTypecheck(1)
		init.Append(assign)
	}
	call.Args = argvars

	assert := ir.NewTypeAssertExpr(pos, sel.X, concretetyp)
	assert.SetOp(ir.ODOTTYPE2)
	assert.SetTypecheck(1)

	tmpnode := typecheck.Temp(concretetyp)
	tmpok := typecheck.Temp(types.Types[types.TBOOL])

	assertAsList := ir.NewAssignListStmt(pos, ir.OAS2DOTTYPE, []ir.Node{tmpnode, tmpok}, []ir.Node{assert})
	assertAsList.Def = true
	assertAsList.SetTypecheck(1)
	init.Append(typecheck.Stmt(assertAsList))

	concreteCallee := typecheck.Callee(ir.NewSelectorExpr(pos, ir.OXDOT, tmpnode, sel.Sel))
	concreteCall := typecheck.Call(pos, concreteCallee, argvars, call.IsDDD)

	var trueNode, elseNode ir.Node
	switch len(retvars) {
	case 0:
		trueNode = concreteCall
		elseNode = call
	case 1:
		trueAs := ir.NewAssignStmt(pos, retvars[0], concreteCall)
		trueAs.SetTypecheck(1)
		trueNode = trueAs

		elseAs := ir.NewAssignStmt(pos, retvars[0], call)
		elseAs.SetTypecheck(1)
		elseNode = elseAs
	default:
		trueAsList := ir.NewAssignListStmt(pos, ir.OAS2FUNC, retvars, []ir.Node{concreteCall})
		trueAsList.SetTypecheck(1)
		trueNode = trueAsList

		elseAsList := ir.NewAssignListStmt(pos, ir.OAS2FUNC, retvars, []ir.Node{call})
		elseAsList.SetTypecheck(1)
		elseNode = elseAsList
	}

	cond := ir.NewIfStmt(pos, nil, nil, nil)
	cond.SetInit(init)
	cond.Cond = typecheck.Expr(tmpok)
	cond.Body = []ir.Node{trueNode}
	cond.Else = []ir.Node{elseNode}
	cond.Likely = true
	cond.SetTypecheck(1)

	body := []ir.Node{cond}

	// This isn't really an inlined call, but InlinedCallExpr makes
	// handling reassignment of return values easier.
	//
	// TODO: make sure this doesn't muck up the inline tree.
	res := ir.NewInlinedCallExpr(pos, body, retvars)
	res.SetType(call.Type())
	res.SetTypecheck(1)

	if base.Flag.LowerM > 2 {
		fmt.Printf("Specializing call to %+v. After: %+v\n", concretetyp, res)
	}

	return res
}

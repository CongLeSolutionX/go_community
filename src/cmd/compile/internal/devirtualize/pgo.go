// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package devirtualize

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/inline"
	"cmd/compile/internal/ir"
	"cmd/compile/internal/logopt"
	"cmd/compile/internal/pgo"
	"cmd/compile/internal/typecheck"
	"cmd/compile/internal/types"
	"encoding/json"
	"fmt"
	"os"
)

// CallStat summarizes a single call site, for debug logging.
type CallStat struct {
	Pkg string
	Pos string

	Caller string

	// Call type. Interface must not be Direct.
	Direct    bool
	Interface bool

	Weight int64

	// Hottest callee from this call site, regardless of type
	// compatibility.
	Hottest       string
	HottestWeight int64

	// Devirtualized callee if != "".
	//
	// Note that this may be different than Hottest because we apply
	// type-check restrictions, which helps distinguish multiple calls on
	// the same line.
	Devirtualized       string
	DevirtualizedWeight int64
}

// ProfileGuided performs call devirtualization of indirect calls based on
// profile information.
//
// Specifically, it performs conditional devirtualization of interface calls
// for the hottest callee. That is, it performs a transformation like:
//
//	type Iface interface {
//		Foo()
//	}
//
//	type Concrete struct{}
//
//	func (Concrete) Foo() {}
//
//	func foo(i Iface) {
//		i.Foo()
//	}
//
// to:
//
//	func foo(i Iface) {
//		if c, ok := i.(Concrete); ok {
//			c.Foo()
//		} else {
//			i.Foo()
//		}
//	}
//
// The primary benefit of this transformation is enabling inlining of the
// direct call.
func ProfileGuided(fn *ir.Func, p *pgo.Profile) {
	ir.CurFunc = fn

	name := ir.LinkFuncName(fn)

	// Can't devirtualize go/defer calls. See comment in Static.
	goDeferCall := make(map[*ir.CallExpr]bool)

	var jsonW *json.Encoder
	if base.Debug.PGODebug >= 3 {
		jsonW = json.NewEncoder(os.Stdout)
	}

	var edit func(n ir.Node) ir.Node
	edit = func(n ir.Node) ir.Node {
		if n == nil {
			return n
		}

		ir.EditChildren(n, edit)

		var call *ir.CallExpr
		switch n := n.(type) {
		case *ir.CallExpr:
			call = n
		case *ir.GoDeferStmt:
			if call, ok := n.Call.(*ir.CallExpr); ok {
				goDeferCall[call] = true
			}
			return n
		default:
			return n
		}

		var stat *CallStat
		if base.Debug.PGODebug >= 3 {
			// Statistics about every single call. Handy for external data analysis.
			//
			// TODO(prattmic): Log via logopt?
			stat = constructCallStat(p, fn, name, call)
			if stat != nil {
				defer func() {
					jsonW.Encode(&stat)
				}()
			}
		}

		if call.Op() != ir.OCALLINTER {
			return n
		}

		if base.Debug.PGODebug >= 2 {
			fmt.Printf("%v: PGO devirtualize considering call %v\n", ir.Line(call), call)
		}

		if goDeferCall[call] {
			if base.Debug.PGODebug >= 2 {
				fmt.Printf("%v: can't PGO devirtualize go/defer call %v\n", ir.Line(call), call)
			}
			return n
		}

		// Bail if we do not have a hot callee.
		callee, weight := findHotConcreteCallee(p, fn, n)
		if callee == nil {
			return n
		}
		// Bail if we do not have a Type node for the hot callee.
		ctyp := typeOfMethodParent(callee)
		if ctyp == nil {
			return n
		}
		// Bail if we can we know for sure it won't inline.
		if !shouldPGODevirt(callee) {
			return n
		}

		if stat != nil {
			stat.Devirtualized = ir.LinkFuncName(callee)
			stat.DevirtualizedWeight = weight
		}

		return rewriteCondCall(n, call, fn, callee, ctyp)
	}

	ir.EditChildren(fn, edit)
}

// shouldPGODevirt checks if we should perform PGO devirtualization to the
// target function.
//
// PGO devirtualization is most valuable when the callee is inlined, so if it
// won't inline we can skip devirtualizing.
func shouldPGODevirt(fn *ir.Func) bool {
	var reason string
	if base.Flag.LowerM > 1 || logopt.Enabled() {
		defer func() {
			if reason != "" {
				if base.Flag.LowerM > 1 {
					fmt.Printf("%v: should not PGO devirtualize %v: %s\n", ir.Line(fn), ir.FuncName(fn), reason)
				}
				if logopt.Enabled() {
					logopt.LogOpt(fn.Pos(), ": should not PGO devirtualize function", "pgo-devirtualize", ir.FuncName(fn), reason)
				}
			}
		}()
	}

	reason = inline.InlineImpossible(fn)
	if reason != "" {
		return false
	}

	// TODO(prattmic): checking only InlineImpossible is very conservative,
	// primarily excluding only functions with pragmas. We probably want to
	// move in either direction. Either:
	//
	// 1. Don't even bother to check InlineImpossible, as it affects so few
	// functions.
	//
	// 2. Or consider the function body (notably cost) to better determine
	// if the function will actually inline.

	return true
}

// constructCallStat builds an initial CallStat describing this call, for
// logging. If the call is devirtualized, the devirtualization fields should be
// updated.
func constructCallStat(p *pgo.Profile, fn *ir.Func, name string, call *ir.CallExpr) *CallStat {
	switch call.Op() {
	case ir.OCALLFUNC, ir.OCALLINTER, ir.OCALLMETH:
	default:
		// We don't care about logging builtin functions.
		return nil
	}

	stat := CallStat{
		Pkg:    base.Ctxt.Pkgpath,
		Pos:    base.FmtPos(call.Pos()),
		Caller: name,
	}

	offset := pgo.NodeLineOffset(call, fn)

	// Sum of all edges from this callsite, regardless of callee.
	// For direct calls, this should be the same as the single edge
	// weight (except for multiple calls on one line, which we
	// can't distinguish).
	callerNode := p.WeightedCG.IRNodes[name]
	for _, edge := range callerNode.OutEdges {
		if edge.CallSiteOffset != offset {
			continue
		}
		stat.Weight += edge.Weight
		if edge.Weight > stat.HottestWeight {
			stat.HottestWeight = edge.Weight
			stat.Hottest = edge.Dst.Name()
		}
	}

	switch call.Op() {
	case ir.OCALLFUNC:
		stat.Interface = false

		callee := pgo.DirectCallee(call.X)
		if callee != nil {
			stat.Direct = true
			if stat.Hottest == "" {
				stat.Hottest = ir.LinkFuncName(callee)
			}
		} else {
			stat.Direct = false
		}
	case ir.OCALLINTER:
		stat.Direct = false
		stat.Interface = true
	case ir.OCALLMETH:
		base.FatalfAt(call.Pos(), "OCALLMETH missed by typecheck")
	}

	return &stat
}

// rewriteCondCall devirtualizes the given call using a direct method call to
// concretetyp.
func rewriteCondCall(n ir.Node, call *ir.CallExpr, curfn, callee *ir.Func, concretetyp *types.Type) ir.Node {
	if base.Flag.LowerM != 0 {
		fmt.Printf("%v: PGO devirtualizing call to %v\n", ir.Line(call), callee)
	}

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
	res := ir.NewInlinedCallExpr(pos, body, retvars)
	res.SetType(call.Type())
	res.SetTypecheck(1)

	if base.Debug.PGODebug >= 3 {
		fmt.Printf("PGO devirtualizing call to %+v. After: %+v\n", concretetyp, res)
	}

	return res
}

// typeOfMethodParent returns the type containing method fn. Returns nil if fn
// is not a method.
func typeOfMethodParent(fn *ir.Func) *types.Type {
	recv := fn.Nname.Type().Recv()
	if recv == nil {
		return nil
	}
	return recv.Type
}

// interfaceCallType returns the type of the interface used in an interface
// call.
func interfaceCallType(n ir.Node) *types.Type {
	if n.Op() != ir.OCALLINTER {
		panic(fmt.Sprintf("Call isn't OCALLINTER: %+v", n))
	}

	call, ok := n.(*ir.CallExpr)
	if !ok {
		panic(fmt.Sprintf("OCALLINTER isn't CallExpr: %+v", n))
	}

	sel, ok := call.X.(*ir.SelectorExpr)
	if !ok {
		panic(fmt.Sprintf("OCALLINTER doesn't contain SelectorExpr: %+v", n))
	}

	return sel.X.Type()
}

// findHotConcreteCallee returns the *ir.Func of the hottest callee of an
// indirect, if available, and its edge weight.
func findHotConcreteCallee(p *pgo.Profile, caller *ir.Func, call ir.Node) (*ir.Func, int64) {
	if _, ok := call.(*ir.CallExpr); !ok {
		panic(fmt.Sprintf("call isn't a call: %+v", call))
		return nil, 0
	}

	callerName := ir.LinkFuncName(caller)
	callerNode := p.WeightedCG.IRNodes[callerName]
	callOffset := pgo.NodeLineOffset(call, caller)

	inter := interfaceCallType(call)

	var hottest *pgo.IREdge

	// Returns true if e is hotter than hottest.
	//
	// Naively this is just e.Weight > hottest.Weight, but because OutEdges
	// has arbitrary iteration order, we need to apply additional sort
	// criteria when e.Weight == hottest.Weight to ensure we have stable
	// selection.
	hotter := func(e *pgo.IREdge) bool {
		if hottest == nil {
			return true
		}
		if e.Weight > hottest.Weight {
			return true
		}
		if e.Weight < hottest.Weight {
			return false
		}

		// e.Weight == hottest.Weight

		if hottest.Dst.AST == nil && e.Dst.AST != nil {
			// Prefer the edge with IR available.
			return true
		}

		// Arbitrary, but the callee names must differ.
		return e.Dst.Name() > hottest.Dst.Name()
	}

	for _, e := range callerNode.OutEdges {
		if e.CallSiteOffset != callOffset {
			continue
		}

		if !hotter(e) {
			// TODO(prattmic): consider total caller weight? i.e.,
			// if the hottest callee is only 10% of the weight,
			// maybe don't devirtualize?
			if base.Debug.PGODebug >= 2 {
				fmt.Printf("%v: edge %s:%d -> %s (weight %d): too cold (hottest %d)\n", ir.Line(call), callerName, callOffset, e.Dst.Name(), e.Weight, hottest.Weight)
			}
			continue
		}

		if e.Dst.AST == nil {
			// Destination isn't visible from this package
			// compilation.
			//
			// We must assume it implements the interface.
			//
			// We still record this as the hottest callee so far
			// because we only want to return the #1 hottest
			// callee. If we skip this then we'd return the #2
			// hottest callee.
			if base.Debug.PGODebug >= 2 {
				fmt.Printf("%v: edge %s:%d -> %s (weight %d) (missing IR): hottest so far\n", ir.Line(call), callerName, callOffset, e.Dst.Name(), e.Weight)
			}
			hottest = e
			continue
		}

		ctyp := typeOfMethodParent(e.Dst.AST)
		if ctyp == nil {
			// Not a method.
			// TODO(prattmic): Support non-interface indirect calls.
			if base.Debug.PGODebug >= 2 {
				fmt.Printf("%v: edge %s:%d -> %s (weight %d): callee not a method\n", ir.Line(call), callerName, callOffset, e.Dst.Name(), e.Weight)
			}
			continue
		}

		// If ctyp doesn't implement inter it is most likely from a
		// different call on the same line
		ok, why := typecheck.Implements(ctyp, inter, base.Debug.PGODebug >= 2)
		if !ok {
			// TODO(prattmic): this is overly strict. Consider if
			// ctyp is a partial implementation of an interface
			// that gets embedded in types that complete the
			// interface. It would still be OK to devirtualize a
			// call to this method.
			//
			// What we'd need to do is check that the function
			// pointer in the itab matches the method we want,
			// rather than doing a full type assertion.
			if base.Debug.PGODebug >= 2 {
				fmt.Printf("%v: edge %s:%d -> %s (weight %d): %v doesn't implement %v (%s)\n", ir.Line(call), callerName, callOffset, e.Dst.Name(), e.Weight, ctyp, inter, why)
			}
			continue
		}

		if base.Debug.PGODebug >= 2 {
			fmt.Printf("%v: edge %s:%d -> %s (weight %d): hottest so far\n", ir.Line(call), callerName, callOffset, e.Dst.Name(), e.Weight)
		}
		hottest = e
	}

	if hottest == nil {
		if base.Debug.PGODebug >= 2 {
			fmt.Printf("%v: call %s:%d: no hot callee\n", ir.Line(call), callerName, callOffset)
		}
		return nil, 0
	}

	if base.Debug.PGODebug >= 2 {
		fmt.Printf("%v call %s:%d: hottest callee %s (weight %d)\n", ir.Line(call), callerName, callOffset, hottest.Dst.Name(), hottest.Weight)
	}
	return hottest.Dst.AST, hottest.Weight
}

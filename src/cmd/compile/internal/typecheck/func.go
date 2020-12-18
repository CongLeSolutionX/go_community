// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// The inlining facility makes 2 passes: first caninl determines which
// functions are suitable for inlining, and for those that are it
// saves a copy of the body. Then inlcalls walks each function body to
// expand calls to inlinable functions.
//
// The Debug.l flag controls the aggressiveness. Note that main() swaps level 0 and 1,
// making 1 the default and -l disable. Additional levels (beyond -l) may be buggy and
// are not supported.
//      0: disabled
//      1: 80-nodes leaf functions, oneliners, panic, lazy typechecking (default)
//      2: (unassigned)
//      3: (unassigned)
//      4: allow non-leaf functions
//
// At some point this may get another default and become switch-offable with -N.
//
// The -d typcheckinl flag enables early typechecking of all imported bodies,
// which is useful to flush out bugs.
//
// The Debug.m flag enables diagnostic output.  a single -m is useful for verifying
// which calls get inlined or not, more is for debugging, and may go away at any point.

package typecheck

import (
	"fmt"

	"cmd/compile/internal/base"
	"cmd/compile/internal/ir"
	"cmd/compile/internal/types"
)

// Lazy typechecking of imported bodies. For local functions, caninl will set ->typecheck
// because they're a copy of an already checked body.
func TypecheckImportedBody(fn *ir.Func) {
	lno := ir.SetPos(fn.Nname)

	ImportBody(fn)

	// typecheckinl is only for imported functions;
	// their bodies may refer to unsafe as long as the package
	// was marked safe during import (which was checked then).
	// the ->inl of a local function has been typechecked before caninl copied it.
	pkg := fnpkg(fn.Nname)

	if pkg == types.LocalPkg || pkg == nil {
		return // typecheckinl on local function
	}

	if base.Flag.LowerM > 2 || base.Debug.Export != 0 {
		fmt.Printf("typecheck import [%v] %L { %v }\n", fn.Sym(), fn, ir.Nodes(fn.Inl.Body))
	}

	savefn := ir.CurFunc
	ir.CurFunc = fn
	Stmts(fn.Inl.Body)
	ir.CurFunc = savefn

	// During expandInline (which imports fn.Func.Inl.Body),
	// declarations are added to fn.Func.Dcl by funcHdr(). Move them
	// to fn.Func.Inl.Dcl for consistency with how local functions
	// behave. (Append because typecheckinl may be called multiple
	// times.)
	fn.Inl.Dcl = append(fn.Inl.Dcl, fn.Dcl...)
	fn.Dcl = nil

	base.Pos = lno
}

// Get the function's package. For ordinary functions it's on the ->sym, but for imported methods
// the ->sym can be re-used in the local package, so peel it off the receiver's type.
func fnpkg(fn *ir.Name) *types.Pkg {
	if ir.IsMethod(fn) {
		// method
		rcvr := fn.Type().Recv().Type

		if rcvr.IsPtr() {
			rcvr = rcvr.Elem()
		}
		if rcvr.Sym() == nil {
			base.Fatalf("receiver with no sym: [%v] %L  (%v)", fn.Sym(), fn, rcvr)
		}
		return rcvr.Sym().Pkg
	}

	// non-method
	return fn.Sym().Pkg
}

// capturevarscomplete is set to true when the capturevars phase is done.
var CaptureVarsComplete bool

// closurename generates a new unique name for a closure within
// outerfunc.
func closurename(outerfunc *ir.Func) *types.Sym {
	outer := "glob."
	prefix := "func"
	gen := &globClosgen

	if outerfunc != nil {
		if outerfunc.OClosure != nil {
			prefix = ""
		}

		outer = ir.FuncName(outerfunc)

		// There may be multiple functions named "_". In those
		// cases, we can't use their individual Closgens as it
		// would lead to name clashes.
		if !ir.IsBlank(outerfunc.Nname) {
			gen = &outerfunc.Closgen
		}
	}

	*gen++
	return Lookup(fmt.Sprintf("%s.%s%d", outer, prefix, *gen))
}

// globClosgen is like Func.Closgen, but for the global scope.
var globClosgen int32

// makepartialcall returns a DCLFUNC node representing the wrapper function (*-fm) needed
// for partial calls.
func makepartialcall(dot *ir.SelectorExpr, t0 *types.Type, meth *types.Sym) *ir.Func {
	rcvrtype := dot.X.Type()
	sym := ir.MethodSymSuffix(rcvrtype, meth, "-fm")

	if sym.Uniq() {
		return sym.Def.(*ir.Func)
	}
	sym.SetUniq(true)

	savecurfn := ir.CurFunc
	saveLineNo := base.Pos
	ir.CurFunc = nil

	// Set line number equal to the line number where the method is declared.
	var m *types.Field
	if lookdot0(meth, rcvrtype, &m, false) == 1 && m.Pos.IsKnown() {
		base.Pos = m.Pos
	}
	// Note: !m.Pos.IsKnown() happens for method expressions where
	// the method is implicitly declared. The Error method of the
	// built-in error type is one such method.  We leave the line
	// number at the use of the method expression in this
	// case. See issue 29389.

	tfn := ir.NewFuncType(base.Pos, nil,
		NewFuncParams(t0.Params(), true),
		NewFuncParams(t0.Results(), false))

	fn := DeclFunc(sym, tfn)
	fn.SetDupok(true)
	fn.SetNeedctxt(true)

	// Declare and initialize variable holding receiver.
	cr := ir.NewClosureRead(rcvrtype, types.Rnd(int64(types.PtrSize), int64(rcvrtype.Align)))
	ptr := NewName(Lookup(".this"))
	Declare(ptr, ir.PAUTO)
	ptr.SetUsed(true)
	var body []ir.Node
	if rcvrtype.IsPtr() || rcvrtype.IsInterface() {
		ptr.SetType(rcvrtype)
		body = append(body, ir.NewAssignStmt(base.Pos, ptr, cr))
	} else {
		ptr.SetType(types.NewPtr(rcvrtype))
		body = append(body, ir.NewAssignStmt(base.Pos, ptr, NodAddr(cr)))
	}

	call := ir.NewCallExpr(base.Pos, ir.OCALL, ir.NewSelectorExpr(base.Pos, ir.OXDOT, ptr, meth), nil)
	call.Args.Set(ir.ParamNames(tfn.Type()))
	call.IsDDD = tfn.Type().IsVariadic()
	if t0.NumResults() != 0 {
		ret := ir.NewReturnStmt(base.Pos, nil)
		ret.Results = []ir.Node{call}
		body = append(body, ret)
	} else {
		body = append(body, call)
	}

	fn.Body.Set(body)
	FinishFuncBody()

	Func(fn)
	// Need to typecheck the body of the just-generated wrapper.
	// typecheckslice() requires that Curfn is set when processing an ORETURN.
	ir.CurFunc = fn
	Stmts(fn.Body)
	sym.Def = fn
	AddFunc(fn)
	ir.CurFunc = savecurfn
	base.Pos = saveLineNo

	return fn
}

// typecheckclosure typechecks an OCLOSURE node. It also creates the named
// function associated with the closure.
// TODO: This creation of the named function should probably really be done in a
// separate pass from type-checking.
func typecheckclosure(clo ir.Node, top int) {
	fn := clo.Func()
	// Set current associated iota value, so iota can be used inside
	// function in ConstSpec, see issue #22344
	if x := getIotaValue(); x >= 0 {
		fn.Iota = x
	}

	fn.ClosureType = Check(fn.ClosureType, ctxType)
	clo.SetType(fn.ClosureType.Type())
	fn.SetClosureCalled(top&ctxCallee != 0)

	// Do not typecheck fn twice, otherwise, we will end up pushing
	// fn to xtop multiple times, causing initLSym called twice.
	// See #30709
	if fn.Typecheck() == 1 {
		return
	}

	for _, ln := range fn.ClosureVars {
		n := ln.Defn
		if !n.Name().Captured() {
			n.Name().SetCaptured(true)
			if n.Name().Decldepth == 0 {
				base.Fatalf("typecheckclosure: var %v does not have decldepth assigned", n)
			}

			// Ignore assignments to the variable in straightline code
			// preceding the first capturing by a closure.
			if n.Name().Decldepth == decldepth {
				n.Name().SetAssigned(false)
			}
		}
	}

	fn.Nname.SetSym(closurename(ir.CurFunc))
	ir.MarkFunc(fn.Nname)
	Func(fn)

	// Type check the body now, but only if we're inside a function.
	// At top level (in a variable initialization: curfn==nil) we're not
	// ready to type check code yet; we'll check it later, because the
	// underlying closure function we create is added to xtop.
	if ir.CurFunc != nil && clo.Type() != nil {
		oldfn := ir.CurFunc
		ir.CurFunc = fn
		olddd := decldepth
		decldepth = 1
		Stmts(fn.Body)
		decldepth = olddd
		ir.CurFunc = oldfn
	}

	AddFunc(fn)
}

func typecheckpartialcall(n ir.Node, sym *types.Sym) *ir.CallPartExpr {
	switch n.Op() {
	case ir.ODOTINTER, ir.ODOTMETH:
		break

	default:
		base.Fatalf("invalid typecheckpartialcall")
	}
	dot := n.(*ir.SelectorExpr)

	// Create top-level function.
	fn := makepartialcall(dot, dot.Type(), sym)
	fn.SetWrapper(true)

	return ir.NewCallPartExpr(dot.Pos(), dot.X, dot.Selection, fn)
}

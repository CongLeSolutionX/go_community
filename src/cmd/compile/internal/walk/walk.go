// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package walk

import (
	"encoding/binary"
	"errors"
	"fmt"
	"go/constant"
	"strings"

	"cmd/compile/internal/base"
	"cmd/compile/internal/ir"
	"cmd/compile/internal/reflectdata"
	"cmd/compile/internal/ssagen"
	"cmd/compile/internal/typecheck"
	"cmd/compile/internal/types"
	"cmd/internal/obj"
	"cmd/internal/objabi"
	"cmd/internal/sys"
)

// The constant is known to runtime.
const tmpstringbufsize = 32
const zeroValSize = 1024 // must match value of runtime/map.go:maxZero

func Walk(fn *ir.Func) {
	ir.CurFunc = fn
	errorsBefore := base.Errors()
	order(fn)
	if base.Errors() > errorsBefore {
		return
	}

	if base.Flag.W != 0 {
		s := fmt.Sprintf("\nbefore walk %v", ir.CurFunc.Sym())
		ir.DumpList(s, ir.CurFunc.Body)
	}

	lno := base.Pos

	// Final typecheck for any unused variables.
	for i, ln := range fn.Dcl {
		if ln.Op() == ir.ONAME && (ln.Class_ == ir.PAUTO || ln.Class_ == ir.PAUTOHEAP) {
			ln = typecheck.AssignExpr(ln).(*ir.Name)
			fn.Dcl[i] = ln
		}
	}

	// Propagate the used flag for typeswitch variables up to the NONAME in its definition.
	for _, ln := range fn.Dcl {
		if ln.Op() == ir.ONAME && (ln.Class_ == ir.PAUTO || ln.Class_ == ir.PAUTOHEAP) && ln.Defn != nil && ln.Defn.Op() == ir.OTYPESW && ln.Used() {
			ln.Defn.(*ir.TypeSwitchGuard).Used = true
		}
	}

	for _, ln := range fn.Dcl {
		if ln.Op() != ir.ONAME || (ln.Class_ != ir.PAUTO && ln.Class_ != ir.PAUTOHEAP) || ln.Sym().Name[0] == '&' || ln.Used() {
			continue
		}
		if defn, ok := ln.Defn.(*ir.TypeSwitchGuard); ok {
			if defn.Used {
				continue
			}
			base.ErrorfAt(defn.Tag.Pos(), "%v declared but not used", ln.Sym())
			defn.Used = true // suppress repeats
		} else {
			base.ErrorfAt(ln.Pos(), "%v declared but not used", ln.Sym())
		}
	}

	base.Pos = lno
	if base.Errors() > errorsBefore {
		return
	}
	walkStmtList(ir.CurFunc.Body)
	if base.Flag.W != 0 {
		s := fmt.Sprintf("after walk %v", ir.CurFunc.Sym())
		ir.DumpList(s, ir.CurFunc.Body)
	}

	zeroResults()
	heapmoves()
	if base.Flag.W != 0 && len(ir.CurFunc.Enter) > 0 {
		s := fmt.Sprintf("enter %v", ir.CurFunc.Sym())
		ir.DumpList(s, ir.CurFunc.Enter)
	}

	if base.Flag.Cfg.Instrumenting {
		instrument(fn)
	}
}

func paramoutheap(fn *ir.Func) bool {
	for _, ln := range fn.Dcl {
		switch ln.Class_ {
		case ir.PPARAMOUT:
			if ir.IsParamStackCopy(ln) || ln.Addrtaken() {
				return true
			}

		case ir.PAUTO:
			// stop early - parameters are over
			return false
		}
	}

	return false
}

// convFuncName builds the runtime function name for interface conversion.
// It also reports whether the function expects the data by address.
// Not all names are possible. For example, we never generate convE2E or convE2I.
func convFuncName(from, to *types.Type) (fnname string, needsaddr bool) {
	tkind := to.Tie()
	switch from.Tie() {
	case 'I':
		if tkind == 'I' {
			return "convI2I", false
		}
	case 'T':
		switch {
		case from.Size() == 2 && from.Align == 2:
			return "convT16", false
		case from.Size() == 4 && from.Align == 4 && !from.HasPointers():
			return "convT32", false
		case from.Size() == 8 && from.Align == types.Types[types.TUINT64].Align && !from.HasPointers():
			return "convT64", false
		}
		if sc := from.SoleComponent(); sc != nil {
			switch {
			case sc.IsString():
				return "convTstring", false
			case sc.IsSlice():
				return "convTslice", false
			}
		}

		switch tkind {
		case 'E':
			if !from.HasPointers() {
				return "convT2Enoptr", true
			}
			return "convT2E", true
		case 'I':
			if !from.HasPointers() {
				return "convT2Inoptr", true
			}
			return "convT2I", true
		}
	}
	base.Fatalf("unknown conv func %c2%c", from.Tie(), to.Tie())
	panic("unreachable")
}

// markTypeUsedInInterface marks that type t is converted to an interface.
// This information is used in the linker in dead method elimination.
func markTypeUsedInInterface(t *types.Type, from *obj.LSym) {
	tsym := reflectdata.TypeSym(t).Linksym()
	// Emit a marker relocation. The linker will know the type is converted
	// to an interface if "from" is reachable.
	r := obj.Addrel(from)
	r.Sym = tsym
	r.Type = objabi.R_USEIFACE
}

// markUsedIfaceMethod marks that an interface method is used in the current
// function. n is OCALLINTER node.
func markUsedIfaceMethod(n *ir.CallExpr) {
	dot := n.X.(*ir.SelectorExpr)
	ityp := dot.X.Type()
	tsym := reflectdata.TypeSym(ityp).Linksym()
	r := obj.Addrel(ir.CurFunc.LSym)
	r.Sym = tsym
	// dot.Xoffset is the method index * Widthptr (the offset of code pointer
	// in itab).
	midx := dot.Offset / int64(types.PtrSize)
	r.Add = reflectdata.InterfaceMethodOffset(ityp, midx)
	r.Type = objabi.R_USEIFACEMETHOD
}

// rtconvfn returns the parameter and result types that will be used by a
// runtime function to convert from type src to type dst. The runtime function
// name can be derived from the names of the returned types.
//
// If no such function is necessary, it returns (Txxx, Txxx).
func rtconvfn(src, dst *types.Type) (param, result types.Kind) {
	if ssagen.Arch.SoftFloat {
		return types.Txxx, types.Txxx
	}

	switch ssagen.Arch.LinkArch.Family {
	case sys.ARM, sys.MIPS:
		if src.IsFloat() {
			switch dst.Kind() {
			case types.TINT64, types.TUINT64:
				return types.TFLOAT64, dst.Kind()
			}
		}
		if dst.IsFloat() {
			switch src.Kind() {
			case types.TINT64, types.TUINT64:
				return src.Kind(), types.TFLOAT64
			}
		}

	case sys.I386:
		if src.IsFloat() {
			switch dst.Kind() {
			case types.TINT64, types.TUINT64:
				return types.TFLOAT64, dst.Kind()
			case types.TUINT32, types.TUINT, types.TUINTPTR:
				return types.TFLOAT64, types.TUINT32
			}
		}
		if dst.IsFloat() {
			switch src.Kind() {
			case types.TINT64, types.TUINT64:
				return src.Kind(), types.TFLOAT64
			case types.TUINT32, types.TUINT, types.TUINTPTR:
				return types.TUINT32, types.TFLOAT64
			}
		}
	}
	return types.Txxx, types.Txxx
}

// TODO(josharian): combine this with its caller and simplify
func reduceSlice(n *ir.SliceExpr) ir.Node {
	low, high, max := n.SliceBounds()
	if high != nil && high.Op() == ir.OLEN && ir.SameSafeExpr(n.X, high.(*ir.UnaryExpr).X) {
		// Reduce x[i:len(x)] to x[i:].
		high = nil
	}
	n.SetSliceBounds(low, high, max)
	if (n.Op() == ir.OSLICE || n.Op() == ir.OSLICESTR) && low == nil && high == nil {
		// Reduce x[:] to x.
		if base.Debug.Slice > 0 {
			base.Warn("slice: omit slice operation")
		}
		return n.X
	}
	return n
}

func ascompatee1(l ir.Node, r ir.Node, init *ir.Nodes) *ir.AssignStmt {
	// convas will turn map assigns into function calls,
	// making it impossible for reorder3 to work.
	n := ir.NewAssignStmt(base.Pos, l, r)

	if l.Op() == ir.OINDEXMAP {
		return n
	}

	return convas(n, init)
}

func ascompatee(op ir.Op, nl, nr []ir.Node, init *ir.Nodes) []ir.Node {
	// check assign expression list to
	// an expression list. called in
	//	expr-list = expr-list

	// ensure order of evaluation for function calls
	for i := range nl {
		nl[i] = safeexpr(nl[i], init)
	}
	for i1 := range nr {
		nr[i1] = safeexpr(nr[i1], init)
	}

	var nn []*ir.AssignStmt
	i := 0
	for ; i < len(nl); i++ {
		if i >= len(nr) {
			break
		}
		// Do not generate 'x = x' during return. See issue 4014.
		if op == ir.ORETURN && ir.SameSafeExpr(nl[i], nr[i]) {
			continue
		}
		nn = append(nn, ascompatee1(nl[i], nr[i], init))
	}

	// cannot happen: caller checked that lists had same length
	if i < len(nl) || i < len(nr) {
		var nln, nrn ir.Nodes
		nln.Set(nl)
		nrn.Set(nr)
		base.Fatalf("error in shape across %+v %v %+v / %d %d [%s]", nln, op, nrn, len(nl), len(nr), ir.FuncName(ir.CurFunc))
	}
	return reorder3(nn)
}

// fncall reports whether assigning an rvalue of type rt to an lvalue l might involve a function call.
func fncall(l ir.Node, rt *types.Type) bool {
	if l.HasCall() || l.Op() == ir.OINDEXMAP {
		return true
	}
	if types.Identical(l.Type(), rt) {
		return false
	}
	// There might be a conversion required, which might involve a runtime call.
	return true
}

// check assign type list to
// an expression list. called in
//	expr-list = func()
func ascompatet(nl ir.Nodes, nr *types.Type) []ir.Node {
	if len(nl) != nr.NumFields() {
		base.Fatalf("ascompatet: assignment count mismatch: %d = %d", len(nl), nr.NumFields())
	}

	var nn, mm ir.Nodes
	for i, l := range nl {
		if ir.IsBlank(l) {
			continue
		}
		r := nr.Field(i)

		// Any assignment to an lvalue that might cause a function call must be
		// deferred until all the returned values have been read.
		if fncall(l, r.Type) {
			tmp := ir.Node(typecheck.Temp(r.Type))
			tmp = typecheck.Expr(tmp)
			a := convas(ir.NewAssignStmt(base.Pos, l, tmp), &mm)
			mm.Append(a)
			l = tmp
		}

		res := ir.NewResultExpr(base.Pos, nil, types.BADWIDTH)
		res.Offset = base.Ctxt.FixedFrameSize() + r.Offset
		res.SetType(r.Type)
		res.SetTypecheck(1)

		a := convas(ir.NewAssignStmt(base.Pos, l, res), &nn)
		updateHasCall(a)
		if a.HasCall() {
			ir.Dump("ascompatet ucount", a)
			base.Fatalf("ascompatet: too many function calls evaluating parameters")
		}

		nn.Append(a)
	}
	return append(nn, mm...)
}

func callnew(t *types.Type) ir.Node {
	types.CalcSize(t)
	n := ir.NewUnaryExpr(base.Pos, ir.ONEWOBJ, reflectdata.TypePtr(t))
	n.SetType(types.NewPtr(t))
	n.SetTypecheck(1)
	n.MarkNonNil()
	return n
}

func convas(n *ir.AssignStmt, init *ir.Nodes) *ir.AssignStmt {
	if n.Op() != ir.OAS {
		base.Fatalf("convas: not OAS %v", n.Op())
	}
	defer updateHasCall(n)

	n.SetTypecheck(1)

	if n.X == nil || n.Y == nil {
		return n
	}

	lt := n.X.Type()
	rt := n.Y.Type()
	if lt == nil || rt == nil {
		return n
	}

	if ir.IsBlank(n.X) {
		n.Y = typecheck.DefaultLit(n.Y, nil)
		return n
	}

	if !types.Identical(lt, rt) {
		n.Y = typecheck.AssignConv(n.Y, lt, "assignment")
		n.Y = walkExpr(n.Y, init)
	}
	types.CalcSize(n.Y.Type())

	return n
}

// reorder3
// from ascompatee
//	a,b = c,d
// simultaneous assignment. there cannot
// be later use of an earlier lvalue.
//
// function calls have been removed.
func reorder3(all []*ir.AssignStmt) []ir.Node {
	// If a needed expression may be affected by an
	// earlier assignment, make an early copy of that
	// expression and use the copy instead.
	var early []ir.Node

	var mapinit ir.Nodes
	for i, n := range all {
		l := n.X

		// Save subexpressions needed on left side.
		// Drill through non-dereferences.
		for {
			switch ll := l; ll.Op() {
			case ir.ODOT:
				ll := ll.(*ir.SelectorExpr)
				l = ll.X
				continue
			case ir.OPAREN:
				ll := ll.(*ir.ParenExpr)
				l = ll.X
				continue
			case ir.OINDEX:
				ll := ll.(*ir.IndexExpr)
				if ll.X.Type().IsArray() {
					ll.Index = reorder3save(ll.Index, all, i, &early)
					l = ll.X
					continue
				}
			}
			break
		}

		switch l.Op() {
		default:
			base.Fatalf("reorder3 unexpected lvalue %v", l.Op())

		case ir.ONAME:
			break

		case ir.OINDEX, ir.OINDEXMAP:
			l := l.(*ir.IndexExpr)
			l.X = reorder3save(l.X, all, i, &early)
			l.Index = reorder3save(l.Index, all, i, &early)
			if l.Op() == ir.OINDEXMAP {
				all[i] = convas(all[i], &mapinit)
			}

		case ir.ODEREF:
			l := l.(*ir.StarExpr)
			l.X = reorder3save(l.X, all, i, &early)
		case ir.ODOTPTR:
			l := l.(*ir.SelectorExpr)
			l.X = reorder3save(l.X, all, i, &early)
		}

		// Save expression on right side.
		all[i].Y = reorder3save(all[i].Y, all, i, &early)
	}

	early = append(mapinit, early...)
	for _, as := range all {
		early = append(early, as)
	}
	return early
}

// if the evaluation of *np would be affected by the
// assignments in all up to but not including the ith assignment,
// copy into a temporary during *early and
// replace *np with that temp.
// The result of reorder3save MUST be assigned back to n, e.g.
// 	n.Left = reorder3save(n.Left, all, i, early)
func reorder3save(n ir.Node, all []*ir.AssignStmt, i int, early *[]ir.Node) ir.Node {
	if !aliased(n, all[:i]) {
		return n
	}

	q := ir.Node(typecheck.Temp(n.Type()))
	as := typecheck.Stmt(ir.NewAssignStmt(base.Pos, q, n))
	*early = append(*early, as)
	return q
}

// Is it possible that the computation of r might be
// affected by assignments in all?
func aliased(r ir.Node, all []*ir.AssignStmt) bool {
	if r == nil {
		return false
	}

	// Treat all fields of a struct as referring to the whole struct.
	// We could do better but we would have to keep track of the fields.
	for r.Op() == ir.ODOT {
		r = r.(*ir.SelectorExpr).X
	}

	// Look for obvious aliasing: a variable being assigned
	// during the all list and appearing in n.
	// Also record whether there are any writes to addressable
	// memory (either main memory or variables whose addresses
	// have been taken).
	memwrite := false
	for _, as := range all {
		// We can ignore assignments to blank.
		if ir.IsBlank(as.X) {
			continue
		}

		lv := ir.OuterValue(as.X)
		if lv.Op() != ir.ONAME {
			memwrite = true
			continue
		}
		l := lv.(*ir.Name)

		switch l.Class_ {
		default:
			base.Fatalf("unexpected class: %v, %v", l, l.Class_)

		case ir.PAUTOHEAP, ir.PEXTERN:
			memwrite = true
			continue

		case ir.PAUTO, ir.PPARAM, ir.PPARAMOUT:
			if l.Name().Addrtaken() {
				memwrite = true
				continue
			}

			if refersToName(l, r) {
				// Direct hit: l appears in r.
				return true
			}
		}
	}

	// The variables being written do not appear in r.
	// However, r might refer to computed addresses
	// that are being written.

	// If no computed addresses are affected by the writes, no aliasing.
	if !memwrite {
		return false
	}

	// If r does not refer to any variables whose addresses have been taken,
	// then the only possible writes to r would be directly to the variables,
	// and we checked those above, so no aliasing problems.
	if !anyAddrTaken(r) {
		return false
	}

	// Otherwise, both the writes and r refer to computed memory addresses.
	// Assume that they might conflict.
	return true
}

// anyAddrTaken reports whether the evaluation n,
// which appears on the left side of an assignment,
// may refer to variables whose addresses have been taken.
func anyAddrTaken(n ir.Node) bool {
	return ir.Any(n, func(n ir.Node) bool {
		switch n.Op() {
		case ir.ONAME:
			n := n.(*ir.Name)
			return n.Class_ == ir.PEXTERN || n.Class_ == ir.PAUTOHEAP || n.Name().Addrtaken()

		case ir.ODOT: // but not ODOTPTR - should have been handled in aliased.
			base.Fatalf("anyAddrTaken unexpected ODOT")

		case ir.OADD,
			ir.OAND,
			ir.OANDAND,
			ir.OANDNOT,
			ir.OBITNOT,
			ir.OCONV,
			ir.OCONVIFACE,
			ir.OCONVNOP,
			ir.ODIV,
			ir.ODOTTYPE,
			ir.OLITERAL,
			ir.OLSH,
			ir.OMOD,
			ir.OMUL,
			ir.ONEG,
			ir.ONIL,
			ir.OOR,
			ir.OOROR,
			ir.OPAREN,
			ir.OPLUS,
			ir.ORSH,
			ir.OSUB,
			ir.OXOR:
			return false
		}
		// Be conservative.
		return true
	})
}

// refersToName reports whether r refers to name.
func refersToName(name *ir.Name, r ir.Node) bool {
	return ir.Any(r, func(r ir.Node) bool {
		return r.Op() == ir.ONAME && r == name
	})
}

var stop = errors.New("stop")

// refersToCommonName reports whether any name
// appears in common between l and r.
// This is called from sinit.go.
func refersToCommonName(l ir.Node, r ir.Node) bool {
	if l == nil || r == nil {
		return false
	}

	// This could be written elegantly as a Find nested inside a Find:
	//
	//	found := ir.Find(l, func(l ir.Node) interface{} {
	//		if l.Op() == ir.ONAME {
	//			return ir.Find(r, func(r ir.Node) interface{} {
	//				if r.Op() == ir.ONAME && l.Name() == r.Name() {
	//					return r
	//				}
	//				return nil
	//			})
	//		}
	//		return nil
	//	})
	//	return found != nil
	//
	// But that would allocate a new closure for the inner Find
	// for each name found on the left side.
	// It may not matter at all, but the below way of writing it
	// only allocates two closures, not O(|L|) closures.

	var doL, doR func(ir.Node) error
	var targetL *ir.Name
	doR = func(r ir.Node) error {
		if r.Op() == ir.ONAME && r.Name() == targetL {
			return stop
		}
		return ir.DoChildren(r, doR)
	}
	doL = func(l ir.Node) error {
		if l.Op() == ir.ONAME {
			l := l.(*ir.Name)
			targetL = l.Name()
			if doR(r) == stop {
				return stop
			}
		}
		return ir.DoChildren(l, doL)
	}
	return doL(l) == stop
}

// paramstoheap returns code to allocate memory for heap-escaped parameters
// and to copy non-result parameters' values from the stack.
func paramstoheap(params *types.Type) []ir.Node {
	var nn []ir.Node
	for _, t := range params.Fields().Slice() {
		v := ir.AsNode(t.Nname)
		if v != nil && v.Sym() != nil && strings.HasPrefix(v.Sym().Name, "~r") { // unnamed result
			v = nil
		}
		if v == nil {
			continue
		}

		if stackcopy := v.Name().Stackcopy; stackcopy != nil {
			nn = append(nn, walkStmt(ir.NewDecl(base.Pos, ir.ODCL, v)))
			if stackcopy.Class_ == ir.PPARAM {
				nn = append(nn, walkStmt(typecheck.Stmt(ir.NewAssignStmt(base.Pos, v, stackcopy))))
			}
		}
	}

	return nn
}

// zeroResults zeros the return values at the start of the function.
// We need to do this very early in the function.  Defer might stop a
// panic and show the return values as they exist at the time of
// panic.  For precise stacks, the garbage collector assumes results
// are always live, so we need to zero them before any allocations,
// even allocations to move params/results to the heap.
// The generated code is added to Curfn's Enter list.
func zeroResults() {
	for _, f := range ir.CurFunc.Type().Results().Fields().Slice() {
		v := ir.AsNode(f.Nname)
		if v != nil && v.Name().Heapaddr != nil {
			// The local which points to the return value is the
			// thing that needs zeroing. This is already handled
			// by a Needzero annotation in plive.go:livenessepilogue.
			continue
		}
		if ir.IsParamHeapCopy(v) {
			// TODO(josharian/khr): Investigate whether we can switch to "continue" here,
			// and document more in either case.
			// In the review of CL 114797, Keith wrote (roughly):
			// I don't think the zeroing below matters.
			// The stack return value will never be marked as live anywhere in the function.
			// It is not written to until deferreturn returns.
			v = v.Name().Stackcopy
		}
		// Zero the stack location containing f.
		ir.CurFunc.Enter.Append(ir.NewAssignStmt(ir.CurFunc.Pos(), v, nil))
	}
}

// returnsfromheap returns code to copy values for heap-escaped parameters
// back to the stack.
func returnsfromheap(params *types.Type) []ir.Node {
	var nn []ir.Node
	for _, t := range params.Fields().Slice() {
		v := ir.AsNode(t.Nname)
		if v == nil {
			continue
		}
		if stackcopy := v.Name().Stackcopy; stackcopy != nil && stackcopy.Class_ == ir.PPARAMOUT {
			nn = append(nn, walkStmt(typecheck.Stmt(ir.NewAssignStmt(base.Pos, stackcopy, v))))
		}
	}

	return nn
}

// heapmoves generates code to handle migrating heap-escaped parameters
// between the stack and the heap. The generated code is added to Curfn's
// Enter and Exit lists.
func heapmoves() {
	lno := base.Pos
	base.Pos = ir.CurFunc.Pos()
	nn := paramstoheap(ir.CurFunc.Type().Recvs())
	nn = append(nn, paramstoheap(ir.CurFunc.Type().Params())...)
	nn = append(nn, paramstoheap(ir.CurFunc.Type().Results())...)
	ir.CurFunc.Enter.Append(nn...)
	base.Pos = ir.CurFunc.Endlineno
	ir.CurFunc.Exit.Append(returnsfromheap(ir.CurFunc.Type().Results())...)
	base.Pos = lno
}

func vmkcall(fn ir.Node, t *types.Type, init *ir.Nodes, va []ir.Node) *ir.CallExpr {
	if fn.Type() == nil || fn.Type().Kind() != types.TFUNC {
		base.Fatalf("mkcall %v %v", fn, fn.Type())
	}

	n := fn.Type().NumParams()
	if n != len(va) {
		base.Fatalf("vmkcall %v needs %v args got %v", fn, n, len(va))
	}

	call := ir.NewCallExpr(base.Pos, ir.OCALL, fn, va)
	typecheck.Call(call)
	call.SetType(t)
	return walkExpr(call, init).(*ir.CallExpr)
}

func mkcall(name string, t *types.Type, init *ir.Nodes, args ...ir.Node) *ir.CallExpr {
	return vmkcall(typecheck.LookupRuntime(name), t, init, args)
}

func mkcall1(fn ir.Node, t *types.Type, init *ir.Nodes, args ...ir.Node) *ir.CallExpr {
	return vmkcall(fn, t, init, args)
}

// byteindex converts n, which is byte-sized, to an int used to index into an array.
// We cannot use conv, because we allow converting bool to int here,
// which is forbidden in user code.
func byteindex(n ir.Node) ir.Node {
	// We cannot convert from bool to int directly.
	// While converting from int8 to int is possible, it would yield
	// the wrong result for negative values.
	// Reinterpreting the value as an unsigned byte solves both cases.
	if !types.Identical(n.Type(), types.Types[types.TUINT8]) {
		n = ir.NewConvExpr(base.Pos, ir.OCONV, nil, n)
		n.SetType(types.Types[types.TUINT8])
		n.SetTypecheck(1)
	}
	n = ir.NewConvExpr(base.Pos, ir.OCONV, nil, n)
	n.SetType(types.Types[types.TINT])
	n.SetTypecheck(1)
	return n
}

func chanfn(name string, n int, t *types.Type) ir.Node {
	if !t.IsChan() {
		base.Fatalf("chanfn %v", t)
	}
	fn := typecheck.LookupRuntime(name)
	switch n {
	default:
		base.Fatalf("chanfn %d", n)
	case 1:
		fn = typecheck.SubstArgTypes(fn, t.Elem())
	case 2:
		fn = typecheck.SubstArgTypes(fn, t.Elem(), t.Elem())
	}
	return fn
}

func mapfn(name string, t *types.Type) ir.Node {
	if !t.IsMap() {
		base.Fatalf("mapfn %v", t)
	}
	fn := typecheck.LookupRuntime(name)
	fn = typecheck.SubstArgTypes(fn, t.Key(), t.Elem(), t.Key(), t.Elem())
	return fn
}

func mapfndel(name string, t *types.Type) ir.Node {
	if !t.IsMap() {
		base.Fatalf("mapfn %v", t)
	}
	fn := typecheck.LookupRuntime(name)
	fn = typecheck.SubstArgTypes(fn, t.Key(), t.Elem(), t.Key())
	return fn
}

const (
	mapslow = iota
	mapfast32
	mapfast32ptr
	mapfast64
	mapfast64ptr
	mapfaststr
	nmapfast
)

type mapnames [nmapfast]string

func mkmapnames(base string, ptr string) mapnames {
	return mapnames{base, base + "_fast32", base + "_fast32" + ptr, base + "_fast64", base + "_fast64" + ptr, base + "_faststr"}
}

var mapaccess1 = mkmapnames("mapaccess1", "")
var mapaccess2 = mkmapnames("mapaccess2", "")
var mapassign = mkmapnames("mapassign", "ptr")
var mapdelete = mkmapnames("mapdelete", "")

func mapfast(t *types.Type) int {
	// Check runtime/map.go:maxElemSize before changing.
	if t.Elem().Width > 128 {
		return mapslow
	}
	switch reflectdata.AlgType(t.Key()) {
	case types.AMEM32:
		if !t.Key().HasPointers() {
			return mapfast32
		}
		if types.PtrSize == 4 {
			return mapfast32ptr
		}
		base.Fatalf("small pointer %v", t.Key())
	case types.AMEM64:
		if !t.Key().HasPointers() {
			return mapfast64
		}
		if types.PtrSize == 8 {
			return mapfast64ptr
		}
		// Two-word object, at least one of which is a pointer.
		// Use the slow path.
	case types.ASTRING:
		return mapfaststr
	}
	return mapslow
}

func writebarrierfn(name string, l *types.Type, r *types.Type) ir.Node {
	fn := typecheck.LookupRuntime(name)
	fn = typecheck.SubstArgTypes(fn, l, r)
	return fn
}

func walkAppendArgs(n *ir.CallExpr, init *ir.Nodes) {
	walkExprListSafe(n.Args, init)

	// walkexprlistsafe will leave OINDEX (s[n]) alone if both s
	// and n are name or literal, but those may index the slice we're
	// modifying here. Fix explicitly.
	ls := n.Args
	for i1, n1 := range ls {
		ls[i1] = cheapexpr(n1, init)
	}
}

// expand append(l1, l2...) to
//   init {
//     s := l1
//     n := len(s) + len(l2)
//     // Compare as uint so growslice can panic on overflow.
//     if uint(n) > uint(cap(s)) {
//       s = growslice(s, n)
//     }
//     s = s[:n]
//     memmove(&s[len(l1)], &l2[0], len(l2)*sizeof(T))
//   }
//   s
//
// l2 is allowed to be a string.
func appendslice(n *ir.CallExpr, init *ir.Nodes) ir.Node {
	walkAppendArgs(n, init)

	l1 := n.Args[0]
	l2 := n.Args[1]
	l2 = cheapexpr(l2, init)
	n.Args[1] = l2

	var nodes ir.Nodes

	// var s []T
	s := typecheck.Temp(l1.Type())
	nodes.Append(ir.NewAssignStmt(base.Pos, s, l1)) // s = l1

	elemtype := s.Type().Elem()

	// n := len(s) + len(l2)
	nn := typecheck.Temp(types.Types[types.TINT])
	nodes.Append(ir.NewAssignStmt(base.Pos, nn, ir.NewBinaryExpr(base.Pos, ir.OADD, ir.NewUnaryExpr(base.Pos, ir.OLEN, s), ir.NewUnaryExpr(base.Pos, ir.OLEN, l2))))

	// if uint(n) > uint(cap(s))
	nif := ir.NewIfStmt(base.Pos, nil, nil, nil)
	nuint := typecheck.Conv(nn, types.Types[types.TUINT])
	scapuint := typecheck.Conv(ir.NewUnaryExpr(base.Pos, ir.OCAP, s), types.Types[types.TUINT])
	nif.Cond = ir.NewBinaryExpr(base.Pos, ir.OGT, nuint, scapuint)

	// instantiate growslice(typ *type, []any, int) []any
	fn := typecheck.LookupRuntime("growslice")
	fn = typecheck.SubstArgTypes(fn, elemtype, elemtype)

	// s = growslice(T, s, n)
	nif.Body = []ir.Node{ir.NewAssignStmt(base.Pos, s, mkcall1(fn, s.Type(), nif.PtrInit(), reflectdata.TypePtr(elemtype), s, nn))}
	nodes.Append(nif)

	// s = s[:n]
	nt := ir.NewSliceExpr(base.Pos, ir.OSLICE, s)
	nt.SetSliceBounds(nil, nn, nil)
	nt.SetBounded(true)
	nodes.Append(ir.NewAssignStmt(base.Pos, s, nt))

	var ncopy ir.Node
	if elemtype.HasPointers() {
		// copy(s[len(l1):], l2)
		slice := ir.NewSliceExpr(base.Pos, ir.OSLICE, s)
		slice.SetType(s.Type())
		slice.SetSliceBounds(ir.NewUnaryExpr(base.Pos, ir.OLEN, l1), nil, nil)

		ir.CurFunc.SetWBPos(n.Pos())

		// instantiate typedslicecopy(typ *type, dstPtr *any, dstLen int, srcPtr *any, srcLen int) int
		fn := typecheck.LookupRuntime("typedslicecopy")
		fn = typecheck.SubstArgTypes(fn, l1.Type().Elem(), l2.Type().Elem())
		ptr1, len1 := backingArrayPtrLen(cheapexpr(slice, &nodes))
		ptr2, len2 := backingArrayPtrLen(l2)
		ncopy = mkcall1(fn, types.Types[types.TINT], &nodes, reflectdata.TypePtr(elemtype), ptr1, len1, ptr2, len2)
	} else if base.Flag.Cfg.Instrumenting && !base.Flag.CompilingRuntime {
		// rely on runtime to instrument:
		//  copy(s[len(l1):], l2)
		// l2 can be a slice or string.
		slice := ir.NewSliceExpr(base.Pos, ir.OSLICE, s)
		slice.SetType(s.Type())
		slice.SetSliceBounds(ir.NewUnaryExpr(base.Pos, ir.OLEN, l1), nil, nil)

		ptr1, len1 := backingArrayPtrLen(cheapexpr(slice, &nodes))
		ptr2, len2 := backingArrayPtrLen(l2)

		fn := typecheck.LookupRuntime("slicecopy")
		fn = typecheck.SubstArgTypes(fn, ptr1.Type().Elem(), ptr2.Type().Elem())
		ncopy = mkcall1(fn, types.Types[types.TINT], &nodes, ptr1, len1, ptr2, len2, ir.NewInt(elemtype.Width))
	} else {
		// memmove(&s[len(l1)], &l2[0], len(l2)*sizeof(T))
		ix := ir.NewIndexExpr(base.Pos, s, ir.NewUnaryExpr(base.Pos, ir.OLEN, l1))
		ix.SetBounded(true)
		addr := typecheck.NodAddr(ix)

		sptr := ir.NewUnaryExpr(base.Pos, ir.OSPTR, l2)

		nwid := cheapexpr(typecheck.Conv(ir.NewUnaryExpr(base.Pos, ir.OLEN, l2), types.Types[types.TUINTPTR]), &nodes)
		nwid = ir.NewBinaryExpr(base.Pos, ir.OMUL, nwid, ir.NewInt(elemtype.Width))

		// instantiate func memmove(to *any, frm *any, length uintptr)
		fn := typecheck.LookupRuntime("memmove")
		fn = typecheck.SubstArgTypes(fn, elemtype, elemtype)
		ncopy = mkcall1(fn, nil, &nodes, addr, sptr, nwid)
	}
	ln := append(nodes, ncopy)

	typecheck.Stmts(ln)
	walkStmtList(ln)
	init.Append(ln...)
	return s
}

// isAppendOfMake reports whether n is of the form append(x , make([]T, y)...).
// isAppendOfMake assumes n has already been typechecked.
func isAppendOfMake(n ir.Node) bool {
	if base.Flag.N != 0 || base.Flag.Cfg.Instrumenting {
		return false
	}

	if n.Typecheck() == 0 {
		base.Fatalf("missing typecheck: %+v", n)
	}

	if n.Op() != ir.OAPPEND {
		return false
	}
	call := n.(*ir.CallExpr)
	if !call.IsDDD || len(call.Args) != 2 || call.Args[1].Op() != ir.OMAKESLICE {
		return false
	}

	mk := call.Args[1].(*ir.MakeExpr)
	if mk.Cap != nil {
		return false
	}

	// y must be either an integer constant or the largest possible positive value
	// of variable y needs to fit into an uint.

	// typecheck made sure that constant arguments to make are not negative and fit into an int.

	// The care of overflow of the len argument to make will be handled by an explicit check of int(len) < 0 during runtime.
	y := mk.Len
	if !ir.IsConst(y, constant.Int) && y.Type().Size() > types.Types[types.TUINT].Size() {
		return false
	}

	return true
}

// extendslice rewrites append(l1, make([]T, l2)...) to
//   init {
//     if l2 >= 0 { // Empty if block here for more meaningful node.SetLikely(true)
//     } else {
//       panicmakeslicelen()
//     }
//     s := l1
//     n := len(s) + l2
//     // Compare n and s as uint so growslice can panic on overflow of len(s) + l2.
//     // cap is a positive int and n can become negative when len(s) + l2
//     // overflows int. Interpreting n when negative as uint makes it larger
//     // than cap(s). growslice will check the int n arg and panic if n is
//     // negative. This prevents the overflow from being undetected.
//     if uint(n) > uint(cap(s)) {
//       s = growslice(T, s, n)
//     }
//     s = s[:n]
//     lptr := &l1[0]
//     sptr := &s[0]
//     if lptr == sptr || !T.HasPointers() {
//       // growslice did not clear the whole underlying array (or did not get called)
//       hp := &s[len(l1)]
//       hn := l2 * sizeof(T)
//       memclr(hp, hn)
//     }
//   }
//   s
func extendslice(n *ir.CallExpr, init *ir.Nodes) ir.Node {
	// isAppendOfMake made sure all possible positive values of l2 fit into an uint.
	// The case of l2 overflow when converting from e.g. uint to int is handled by an explicit
	// check of l2 < 0 at runtime which is generated below.
	l2 := typecheck.Conv(n.Args[1].(*ir.MakeExpr).Len, types.Types[types.TINT])
	l2 = typecheck.Expr(l2)
	n.Args[1] = l2 // walkAppendArgs expects l2 in n.List.Second().

	walkAppendArgs(n, init)

	l1 := n.Args[0]
	l2 = n.Args[1] // re-read l2, as it may have been updated by walkAppendArgs

	var nodes []ir.Node

	// if l2 >= 0 (likely happens), do nothing
	nifneg := ir.NewIfStmt(base.Pos, ir.NewBinaryExpr(base.Pos, ir.OGE, l2, ir.NewInt(0)), nil, nil)
	nifneg.Likely = true

	// else panicmakeslicelen()
	nifneg.Else = []ir.Node{mkcall("panicmakeslicelen", nil, init)}
	nodes = append(nodes, nifneg)

	// s := l1
	s := typecheck.Temp(l1.Type())
	nodes = append(nodes, ir.NewAssignStmt(base.Pos, s, l1))

	elemtype := s.Type().Elem()

	// n := len(s) + l2
	nn := typecheck.Temp(types.Types[types.TINT])
	nodes = append(nodes, ir.NewAssignStmt(base.Pos, nn, ir.NewBinaryExpr(base.Pos, ir.OADD, ir.NewUnaryExpr(base.Pos, ir.OLEN, s), l2)))

	// if uint(n) > uint(cap(s))
	nuint := typecheck.Conv(nn, types.Types[types.TUINT])
	capuint := typecheck.Conv(ir.NewUnaryExpr(base.Pos, ir.OCAP, s), types.Types[types.TUINT])
	nif := ir.NewIfStmt(base.Pos, ir.NewBinaryExpr(base.Pos, ir.OGT, nuint, capuint), nil, nil)

	// instantiate growslice(typ *type, old []any, newcap int) []any
	fn := typecheck.LookupRuntime("growslice")
	fn = typecheck.SubstArgTypes(fn, elemtype, elemtype)

	// s = growslice(T, s, n)
	nif.Body = []ir.Node{ir.NewAssignStmt(base.Pos, s, mkcall1(fn, s.Type(), nif.PtrInit(), reflectdata.TypePtr(elemtype), s, nn))}
	nodes = append(nodes, nif)

	// s = s[:n]
	nt := ir.NewSliceExpr(base.Pos, ir.OSLICE, s)
	nt.SetSliceBounds(nil, nn, nil)
	nt.SetBounded(true)
	nodes = append(nodes, ir.NewAssignStmt(base.Pos, s, nt))

	// lptr := &l1[0]
	l1ptr := typecheck.Temp(l1.Type().Elem().PtrTo())
	tmp := ir.NewUnaryExpr(base.Pos, ir.OSPTR, l1)
	nodes = append(nodes, ir.NewAssignStmt(base.Pos, l1ptr, tmp))

	// sptr := &s[0]
	sptr := typecheck.Temp(elemtype.PtrTo())
	tmp = ir.NewUnaryExpr(base.Pos, ir.OSPTR, s)
	nodes = append(nodes, ir.NewAssignStmt(base.Pos, sptr, tmp))

	// hp := &s[len(l1)]
	ix := ir.NewIndexExpr(base.Pos, s, ir.NewUnaryExpr(base.Pos, ir.OLEN, l1))
	ix.SetBounded(true)
	hp := typecheck.ConvNop(typecheck.NodAddr(ix), types.Types[types.TUNSAFEPTR])

	// hn := l2 * sizeof(elem(s))
	hn := typecheck.Conv(ir.NewBinaryExpr(base.Pos, ir.OMUL, l2, ir.NewInt(elemtype.Width)), types.Types[types.TUINTPTR])

	clrname := "memclrNoHeapPointers"
	hasPointers := elemtype.HasPointers()
	if hasPointers {
		clrname = "memclrHasPointers"
		ir.CurFunc.SetWBPos(n.Pos())
	}

	var clr ir.Nodes
	clrfn := mkcall(clrname, nil, &clr, hp, hn)
	clr.Append(clrfn)

	if hasPointers {
		// if l1ptr == sptr
		nifclr := ir.NewIfStmt(base.Pos, ir.NewBinaryExpr(base.Pos, ir.OEQ, l1ptr, sptr), nil, nil)
		nifclr.Body = clr
		nodes = append(nodes, nifclr)
	} else {
		nodes = append(nodes, clr...)
	}

	typecheck.Stmts(nodes)
	walkStmtList(nodes)
	init.Append(nodes...)
	return s
}

func eqfor(t *types.Type) (n ir.Node, needsize bool) {
	// Should only arrive here with large memory or
	// a struct/array containing a non-memory field/element.
	// Small memory is handled inline, and single non-memory
	// is handled by walkcompare.
	switch a, _ := types.AlgType(t); a {
	case types.AMEM:
		n := typecheck.LookupRuntime("memequal")
		n = typecheck.SubstArgTypes(n, t, t)
		return n, true
	case types.ASPECIAL:
		sym := reflectdata.TypeSymPrefix(".eq", t)
		n := typecheck.NewName(sym)
		ir.MarkFunc(n)
		n.SetType(typecheck.NewFuncType(nil, []*ir.Field{
			ir.NewField(base.Pos, nil, nil, types.NewPtr(t)),
			ir.NewField(base.Pos, nil, nil, types.NewPtr(t)),
		}, []*ir.Field{
			ir.NewField(base.Pos, nil, nil, types.Types[types.TBOOL]),
		}))
		return n, false
	}
	base.Fatalf("eqfor %v", t)
	return nil, false
}

func tracecmpArg(n ir.Node, t *types.Type, init *ir.Nodes) ir.Node {
	// Ugly hack to avoid "constant -1 overflows uintptr" errors, etc.
	if n.Op() == ir.OLITERAL && n.Type().IsSigned() && ir.Int64Val(n) < 0 {
		n = copyexpr(n, n.Type(), init)
	}

	return typecheck.Conv(n, t)
}

// The result of finishcompare MUST be assigned back to n, e.g.
// 	n.Left = finishcompare(n.Left, x, r, init)
func finishcompare(n *ir.BinaryExpr, r ir.Node, init *ir.Nodes) ir.Node {
	r = typecheck.Expr(r)
	r = typecheck.Conv(r, n.Type())
	r = walkExpr(r, init)
	return r
}

// return 1 if integer n must be in range [0, max), 0 otherwise
func bounded(n ir.Node, max int64) bool {
	if n.Type() == nil || !n.Type().IsInteger() {
		return false
	}

	sign := n.Type().IsSigned()
	bits := int32(8 * n.Type().Width)

	if ir.IsSmallIntConst(n) {
		v := ir.Int64Val(n)
		return 0 <= v && v < max
	}

	switch n.Op() {
	case ir.OAND, ir.OANDNOT:
		n := n.(*ir.BinaryExpr)
		v := int64(-1)
		switch {
		case ir.IsSmallIntConst(n.X):
			v = ir.Int64Val(n.X)
		case ir.IsSmallIntConst(n.Y):
			v = ir.Int64Val(n.Y)
			if n.Op() == ir.OANDNOT {
				v = ^v
				if !sign {
					v &= 1<<uint(bits) - 1
				}
			}
		}
		if 0 <= v && v < max {
			return true
		}

	case ir.OMOD:
		n := n.(*ir.BinaryExpr)
		if !sign && ir.IsSmallIntConst(n.Y) {
			v := ir.Int64Val(n.Y)
			if 0 <= v && v <= max {
				return true
			}
		}

	case ir.ODIV:
		n := n.(*ir.BinaryExpr)
		if !sign && ir.IsSmallIntConst(n.Y) {
			v := ir.Int64Val(n.Y)
			for bits > 0 && v >= 2 {
				bits--
				v >>= 1
			}
		}

	case ir.ORSH:
		n := n.(*ir.BinaryExpr)
		if !sign && ir.IsSmallIntConst(n.Y) {
			v := ir.Int64Val(n.Y)
			if v > int64(bits) {
				return true
			}
			bits -= int32(v)
		}
	}

	if !sign && bits <= 62 && 1<<uint(bits) <= max {
		return true
	}

	return false
}

// usemethod checks interface method calls for uses of reflect.Type.Method.
func usemethod(n *ir.CallExpr) {
	t := n.X.Type()

	// Looking for either of:
	//	Method(int) reflect.Method
	//	MethodByName(string) (reflect.Method, bool)
	//
	// TODO(crawshaw): improve precision of match by working out
	//                 how to check the method name.
	if n := t.NumParams(); n != 1 {
		return
	}
	if n := t.NumResults(); n != 1 && n != 2 {
		return
	}
	p0 := t.Params().Field(0)
	res0 := t.Results().Field(0)
	var res1 *types.Field
	if t.NumResults() == 2 {
		res1 = t.Results().Field(1)
	}

	if res1 == nil {
		if p0.Type.Kind() != types.TINT {
			return
		}
	} else {
		if !p0.Type.IsString() {
			return
		}
		if !res1.Type.IsBoolean() {
			return
		}
	}

	// Note: Don't rely on res0.Type.String() since its formatting depends on multiple factors
	//       (including global variables such as numImports - was issue #19028).
	// Also need to check for reflect package itself (see Issue #38515).
	if s := res0.Type.Sym(); s != nil && s.Name == "Method" && types.IsReflectPkg(s.Pkg) {
		ir.CurFunc.SetReflectMethod(true)
		// The LSym is initialized at this point. We need to set the attribute on the LSym.
		ir.CurFunc.LSym.Set(obj.AttrReflectMethod, true)
	}
}

func usefield(n *ir.SelectorExpr) {
	if objabi.Fieldtrack_enabled == 0 {
		return
	}

	switch n.Op() {
	default:
		base.Fatalf("usefield %v", n.Op())

	case ir.ODOT, ir.ODOTPTR:
		break
	}
	if n.Sel == nil {
		// No field name.  This DOTPTR was built by the compiler for access
		// to runtime data structures.  Ignore.
		return
	}

	t := n.X.Type()
	if t.IsPtr() {
		t = t.Elem()
	}
	field := n.Selection
	if field == nil {
		base.Fatalf("usefield %v %v without paramfld", n.X.Type(), n.Sel)
	}
	if field.Sym != n.Sel || field.Offset != n.Offset {
		base.Fatalf("field inconsistency: %v,%v != %v,%v", field.Sym, field.Offset, n.Sel, n.Offset)
	}
	if !strings.Contains(field.Note, "go:\"track\"") {
		return
	}

	outer := n.X.Type()
	if outer.IsPtr() {
		outer = outer.Elem()
	}
	if outer.Sym() == nil {
		base.Errorf("tracked field must be in named struct type")
	}
	if !types.IsExported(field.Sym.Name) {
		base.Errorf("tracked field must be exported (upper case)")
	}

	sym := reflectdata.TrackSym(outer, field)
	if ir.CurFunc.FieldTrack == nil {
		ir.CurFunc.FieldTrack = make(map[*types.Sym]struct{})
	}
	ir.CurFunc.FieldTrack[sym] = struct{}{}
}

// anySideEffects reports whether n contains any operations that could have observable side effects.
func anySideEffects(n ir.Node) bool {
	return ir.Any(n, func(n ir.Node) bool {
		switch n.Op() {
		// Assume side effects unless we know otherwise.
		default:
			return true

		// No side effects here (arguments are checked separately).
		case ir.ONAME,
			ir.ONONAME,
			ir.OTYPE,
			ir.OPACK,
			ir.OLITERAL,
			ir.ONIL,
			ir.OADD,
			ir.OSUB,
			ir.OOR,
			ir.OXOR,
			ir.OADDSTR,
			ir.OADDR,
			ir.OANDAND,
			ir.OBYTES2STR,
			ir.ORUNES2STR,
			ir.OSTR2BYTES,
			ir.OSTR2RUNES,
			ir.OCAP,
			ir.OCOMPLIT,
			ir.OMAPLIT,
			ir.OSTRUCTLIT,
			ir.OARRAYLIT,
			ir.OSLICELIT,
			ir.OPTRLIT,
			ir.OCONV,
			ir.OCONVIFACE,
			ir.OCONVNOP,
			ir.ODOT,
			ir.OEQ,
			ir.ONE,
			ir.OLT,
			ir.OLE,
			ir.OGT,
			ir.OGE,
			ir.OKEY,
			ir.OSTRUCTKEY,
			ir.OLEN,
			ir.OMUL,
			ir.OLSH,
			ir.ORSH,
			ir.OAND,
			ir.OANDNOT,
			ir.ONEW,
			ir.ONOT,
			ir.OBITNOT,
			ir.OPLUS,
			ir.ONEG,
			ir.OOROR,
			ir.OPAREN,
			ir.ORUNESTR,
			ir.OREAL,
			ir.OIMAG,
			ir.OCOMPLEX:
			return false

		// Only possible side effect is division by zero.
		case ir.ODIV, ir.OMOD:
			n := n.(*ir.BinaryExpr)
			if n.Y.Op() != ir.OLITERAL || constant.Sign(n.Y.Val()) == 0 {
				return true
			}

		// Only possible side effect is panic on invalid size,
		// but many makechan and makemap use size zero, which is definitely OK.
		case ir.OMAKECHAN, ir.OMAKEMAP:
			n := n.(*ir.MakeExpr)
			if !ir.IsConst(n.Len, constant.Int) || constant.Sign(n.Len.Val()) != 0 {
				return true
			}

		// Only possible side effect is panic on invalid size.
		// TODO(rsc): Merge with previous case (probably breaks toolstash -cmp).
		case ir.OMAKESLICE, ir.OMAKESLICECOPY:
			return true
		}
		return false
	})
}

// Rewrite
//	go builtin(x, y, z)
// into
//	go func(a1, a2, a3) {
//		builtin(a1, a2, a3)
//	}(x, y, z)
// for print, println, and delete.
//
// Rewrite
//	go f(x, y, uintptr(unsafe.Pointer(z)))
// into
//	go func(a1, a2, a3) {
//		builtin(a1, a2, uintptr(a3))
//	}(x, y, unsafe.Pointer(z))
// for function contains unsafe-uintptr arguments.

var wrapCall_prgen int

// The result of wrapCall MUST be assigned back to n, e.g.
// 	n.Left = wrapCall(n.Left, init)
func wrapCall(n *ir.CallExpr, init *ir.Nodes) ir.Node {
	if len(n.Init()) != 0 {
		walkStmtList(n.Init())
		init.Append(n.PtrInit().Take()...)
	}

	isBuiltinCall := n.Op() != ir.OCALLFUNC && n.Op() != ir.OCALLMETH && n.Op() != ir.OCALLINTER

	// Turn f(a, b, []T{c, d, e}...) back into f(a, b, c, d, e).
	if !isBuiltinCall && n.IsDDD {
		last := len(n.Args) - 1
		if va := n.Args[last]; va.Op() == ir.OSLICELIT {
			va := va.(*ir.CompLitExpr)
			n.Args.Set(append(n.Args[:last], va.List...))
			n.IsDDD = false
		}
	}

	// origArgs keeps track of what argument is uintptr-unsafe/unsafe-uintptr conversion.
	origArgs := make([]ir.Node, len(n.Args))
	var funcArgs []*ir.Field
	for i, arg := range n.Args {
		s := typecheck.LookupNum("a", i)
		if !isBuiltinCall && arg.Op() == ir.OCONVNOP && arg.Type().IsUintptr() && arg.(*ir.ConvExpr).X.Type().IsUnsafePtr() {
			origArgs[i] = arg
			arg = arg.(*ir.ConvExpr).X
			n.Args[i] = arg
		}
		funcArgs = append(funcArgs, ir.NewField(base.Pos, s, nil, arg.Type()))
	}
	t := ir.NewFuncType(base.Pos, nil, funcArgs, nil)

	wrapCall_prgen++
	sym := typecheck.LookupNum("wrap·", wrapCall_prgen)
	fn := typecheck.DeclFunc(sym, t)

	args := ir.ParamNames(t.Type())
	for i, origArg := range origArgs {
		if origArg == nil {
			continue
		}
		args[i] = ir.NewConvExpr(base.Pos, origArg.Op(), origArg.Type(), args[i])
	}
	call := ir.NewCallExpr(base.Pos, n.Op(), n.X, args)
	if !isBuiltinCall {
		call.SetOp(ir.OCALL)
		call.IsDDD = n.IsDDD
	}
	fn.Body = []ir.Node{call}

	typecheck.FinishFuncBody()

	typecheck.Func(fn)
	typecheck.Stmts(fn.Body)
	typecheck.Target.Decls = append(typecheck.Target.Decls, fn)

	call = ir.NewCallExpr(base.Pos, ir.OCALL, fn.Nname, n.Args)
	return walkExpr(typecheck.Stmt(call), init)
}

// canMergeLoads reports whether the backend optimization passes for
// the current architecture can combine adjacent loads into a single
// larger, possibly unaligned, load. Note that currently the
// optimizations must be able to handle little endian byte order.
func canMergeLoads() bool {
	switch ssagen.Arch.LinkArch.Family {
	case sys.ARM64, sys.AMD64, sys.I386, sys.S390X:
		return true
	case sys.PPC64:
		// Load combining only supported on ppc64le.
		return ssagen.Arch.LinkArch.ByteOrder == binary.LittleEndian
	}
	return false
}

// isRuneCount reports whether n is of the form len([]rune(string)).
// These are optimized into a call to runtime.countrunes.
func isRuneCount(n ir.Node) bool {
	return base.Flag.N == 0 && !base.Flag.Cfg.Instrumenting && n.Op() == ir.OLEN && n.(*ir.UnaryExpr).X.Op() == ir.OSTR2RUNES
}

func walkCheckPtrAlignment(n *ir.ConvExpr, init *ir.Nodes, count ir.Node) ir.Node {
	if !n.Type().IsPtr() {
		base.Fatalf("expected pointer type: %v", n.Type())
	}
	elem := n.Type().Elem()
	if count != nil {
		if !elem.IsArray() {
			base.Fatalf("expected array type: %v", elem)
		}
		elem = elem.Elem()
	}

	size := elem.Size()
	if elem.Alignment() == 1 && (size == 0 || size == 1 && count == nil) {
		return n
	}

	if count == nil {
		count = ir.NewInt(1)
	}

	n.X = cheapexpr(n.X, init)
	init.Append(mkcall("checkptrAlignment", nil, init, typecheck.ConvNop(n.X, types.Types[types.TUNSAFEPTR]), reflectdata.TypePtr(elem), typecheck.Conv(count, types.Types[types.TUINTPTR])))
	return n
}

var walkCheckPtrArithmeticMarker byte

func walkCheckPtrArithmetic(n *ir.ConvExpr, init *ir.Nodes) ir.Node {
	// Calling cheapexpr(n, init) below leads to a recursive call
	// to walkexpr, which leads us back here again. Use n.Opt to
	// prevent infinite loops.
	if opt := n.Opt(); opt == &walkCheckPtrArithmeticMarker {
		return n
	} else if opt != nil {
		// We use n.Opt() here because today it's not used for OCONVNOP. If that changes,
		// there's no guarantee that temporarily replacing it is safe, so just hard fail here.
		base.Fatalf("unexpected Opt: %v", opt)
	}
	n.SetOpt(&walkCheckPtrArithmeticMarker)
	defer n.SetOpt(nil)

	// TODO(mdempsky): Make stricter. We only need to exempt
	// reflect.Value.Pointer and reflect.Value.UnsafeAddr.
	switch n.X.Op() {
	case ir.OCALLFUNC, ir.OCALLMETH, ir.OCALLINTER:
		return n
	}

	if n.X.Op() == ir.ODOTPTR && ir.IsReflectHeaderDataField(n.X) {
		return n
	}

	// Find original unsafe.Pointer operands involved in this
	// arithmetic expression.
	//
	// "It is valid both to add and to subtract offsets from a
	// pointer in this way. It is also valid to use &^ to round
	// pointers, usually for alignment."
	var originals []ir.Node
	var walk func(n ir.Node)
	walk = func(n ir.Node) {
		switch n.Op() {
		case ir.OADD:
			n := n.(*ir.BinaryExpr)
			walk(n.X)
			walk(n.Y)
		case ir.OSUB, ir.OANDNOT:
			n := n.(*ir.BinaryExpr)
			walk(n.X)
		case ir.OCONVNOP:
			n := n.(*ir.ConvExpr)
			if n.X.Type().IsUnsafePtr() {
				n.X = cheapexpr(n.X, init)
				originals = append(originals, typecheck.ConvNop(n.X, types.Types[types.TUNSAFEPTR]))
			}
		}
	}
	walk(n.X)

	cheap := cheapexpr(n, init)

	slice := typecheck.MakeDotArgs(types.NewSlice(types.Types[types.TUNSAFEPTR]), originals)
	slice.SetEsc(ir.EscNone)

	init.Append(mkcall("checkptrArithmetic", nil, init, typecheck.ConvNop(cheap, types.Types[types.TUNSAFEPTR]), slice))
	// TODO(khr): Mark backing store of slice as dead. This will allow us to reuse
	// the backing store for multiple calls to checkptrArithmetic.

	return cheap
}

// appendWalkStmt typechecks and walks stmt and then appends it to init.
func appendWalkStmt(init *ir.Nodes, stmt ir.Node) {
	op := stmt.Op()
	n := typecheck.Stmt(stmt)
	if op == ir.OAS || op == ir.OAS2 {
		// If the assignment has side effects, walkexpr will append them
		// directly to init for us, while walkstmt will wrap it in an OBLOCK.
		// We need to append them directly.
		// TODO(rsc): Clean this up.
		n = walkExpr(n, init)
	} else {
		n = walkStmt(n)
	}
	init.Append(n)
}

// The max number of defers in a function using open-coded defers. We enforce this
// limit because the deferBits bitmask is currently a single byte (to minimize code size)
const maxOpenDefers = 8

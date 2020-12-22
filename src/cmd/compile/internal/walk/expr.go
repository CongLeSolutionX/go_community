// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package walk

import (
	"encoding/binary"
	"fmt"
	"go/constant"
	"go/token"
	"strings"

	"cmd/compile/internal/base"
	"cmd/compile/internal/escape"
	"cmd/compile/internal/ir"
	"cmd/compile/internal/reflectdata"
	"cmd/compile/internal/ssagen"
	"cmd/compile/internal/staticdata"
	"cmd/compile/internal/typecheck"
	"cmd/compile/internal/types"
	"cmd/internal/src"
)

// The result of walkexpr MUST be assigned back to n, e.g.
// 	n.Left = walkexpr(n.Left, init)
func walkExpr(n ir.Node, init *ir.Nodes) ir.Node {
	if n == nil {
		return n
	}

	// Eagerly checkwidth all expressions for the back end.
	if n.Type() != nil && !n.Type().WidthCalculated() {
		switch n.Type().Kind() {
		case types.TBLANK, types.TNIL, types.TIDEAL:
		default:
			types.CheckSize(n.Type())
		}
	}

	if init == n.PtrInit() {
		// not okay to use n->ninit when walking n,
		// because we might replace n with some other node
		// and would lose the init list.
		base.Fatalf("walkexpr init == &n->ninit")
	}

	if len(n.Init()) != 0 {
		walkStmtList(n.Init())
		init.Append(n.PtrInit().Take()...)
	}

	lno := ir.SetPos(n)

	if base.Flag.LowerW > 1 {
		ir.Dump("before walk expr", n)
	}

	if n.Typecheck() != 1 {
		base.Fatalf("missed typecheck: %+v", n)
	}

	if n.Type().IsUntyped() {
		base.Fatalf("expression has untyped type: %+v", n)
	}

	if n.Op() == ir.ONAME && n.(*ir.Name).Class_ == ir.PAUTOHEAP {
		n := n.(*ir.Name)
		nn := ir.NewStarExpr(base.Pos, n.Name().Heapaddr)
		nn.X.MarkNonNil()
		return walkExpr(typecheck.Expr(nn), init)
	}

	n = walkExpr1(n, init)

	// Expressions that are constant at run time but not
	// considered const by the language spec are not turned into
	// constants until walk. For example, if n is y%1 == 0, the
	// walk of y%1 may have replaced it by 0.
	// Check whether n with its updated args is itself now a constant.
	t := n.Type()
	n = typecheck.EvalConst(n)
	if n.Type() != t {
		base.Fatalf("evconst changed Type: %v had type %v, now %v", n, t, n.Type())
	}
	if n.Op() == ir.OLITERAL {
		n = typecheck.Expr(n)
		// Emit string symbol now to avoid emitting
		// any concurrently during the backend.
		if v := n.Val(); v.Kind() == constant.String {
			_ = staticdata.StringSym(n.Pos(), constant.StringVal(v))
		}
	}

	updateHasCall(n)

	if base.Flag.LowerW != 0 && n != nil {
		ir.Dump("after walk expr", n)
	}

	base.Pos = lno
	return n
}

func walkExpr1(n ir.Node, init *ir.Nodes) ir.Node {
	switch n.Op() {
	default:
		ir.Dump("walk", n)
		base.Fatalf("walkexpr: switch 1 unknown op %+v", n.Op())
		panic("unreachable")

	case ir.ONONAME, ir.OGETG, ir.ONEWOBJ, ir.OMETHEXPR:
		return n

	case ir.OTYPE, ir.ONAME, ir.OLITERAL, ir.ONIL, ir.ONAMEOFFSET:
		// TODO(mdempsky): Just return n; see discussion on CL 38655.
		// Perhaps refactor to use Node.mayBeShared for these instead.
		// If these return early, make sure to still call
		// stringsym for constant strings.
		return n

	case ir.ONOT, ir.ONEG, ir.OPLUS, ir.OBITNOT, ir.OREAL, ir.OIMAG, ir.OSPTR, ir.OITAB, ir.OIDATA:
		n := n.(*ir.UnaryExpr)
		n.X = walkExpr(n.X, init)
		return n

	case ir.ODOTMETH, ir.ODOTINTER:
		n := n.(*ir.SelectorExpr)
		n.X = walkExpr(n.X, init)
		return n

	case ir.OADDR:
		n := n.(*ir.AddrExpr)
		n.X = walkExpr(n.X, init)
		return n

	case ir.ODEREF:
		n := n.(*ir.StarExpr)
		n.X = walkExpr(n.X, init)
		return n

	case ir.OEFACE, ir.OAND, ir.OANDNOT, ir.OSUB, ir.OMUL, ir.OADD, ir.OOR, ir.OXOR, ir.OLSH, ir.ORSH:
		n := n.(*ir.BinaryExpr)
		n.X = walkExpr(n.X, init)
		n.Y = walkExpr(n.Y, init)
		return n

	case ir.ODOT, ir.ODOTPTR:
		n := n.(*ir.SelectorExpr)
		return walkDot(n, init)

	case ir.ODOTTYPE, ir.ODOTTYPE2:
		n := n.(*ir.TypeAssertExpr)
		return walkDotType(n, init)

	case ir.OLEN, ir.OCAP:
		n := n.(*ir.UnaryExpr)
		return walkLenCap(n, init)

	case ir.OCOMPLEX:
		n := n.(*ir.BinaryExpr)
		n.X = walkExpr(n.X, init)
		n.Y = walkExpr(n.Y, init)
		return n

	case ir.OEQ, ir.ONE, ir.OLT, ir.OLE, ir.OGT, ir.OGE:
		n := n.(*ir.BinaryExpr)
		return walkCompare(n, init)

	case ir.OANDAND, ir.OOROR:
		n := n.(*ir.LogicalExpr)
		return walkLogical(n, init)

	case ir.OPRINT, ir.OPRINTN:
		return walkPrint(n.(*ir.CallExpr), init)

	case ir.OPANIC:
		n := n.(*ir.UnaryExpr)
		return mkcall("gopanic", nil, init, n.X)

	case ir.ORECOVER:
		n := n.(*ir.CallExpr)
		return mkcall("gorecover", n.Type(), init, typecheck.NodAddr(ir.RegFP))

	case ir.OCLOSUREREAD, ir.OCFUNC:
		return n

	case ir.OCALLINTER, ir.OCALLFUNC, ir.OCALLMETH:
		n := n.(*ir.CallExpr)
		return walkCall(n, init)

	case ir.OAS, ir.OASOP:
		return walkAssign(init, n)

	case ir.OAS2:
		n := n.(*ir.AssignListStmt)
		return walkAssignList(init, n)

	// a,b,... = fn()
	case ir.OAS2FUNC:
		n := n.(*ir.AssignListStmt)
		return walkAssignFunc(init, n)

	// x, y = <-c
	// order.stmt made sure x is addressable or blank.
	case ir.OAS2RECV:
		n := n.(*ir.AssignListStmt)
		return walkAssignRecv(init, n)

	// a,b = m[i]
	case ir.OAS2MAPR:
		n := n.(*ir.AssignListStmt)
		return walkAssignMapRead(init, n)

	case ir.ODELETE:
		n := n.(*ir.CallExpr)
		return walkDelete(init, n)

	case ir.OAS2DOTTYPE:
		n := n.(*ir.AssignListStmt)
		return walkAssignDotType(n, init)

	case ir.OCONVIFACE:
		n := n.(*ir.ConvExpr)
		return walkConvInterface(n, init)

	case ir.OCONV, ir.OCONVNOP:
		n := n.(*ir.ConvExpr)
		return walkConv(n, init)

	case ir.ODIV, ir.OMOD:
		n := n.(*ir.BinaryExpr)
		return walkDivMod(n, init)

	case ir.OINDEX:
		n := n.(*ir.IndexExpr)
		return walkIndex(n, init)

	case ir.OINDEXMAP:
		n := n.(*ir.IndexExpr)
		return walkIndexMap(n, init)

	case ir.ORECV:
		base.Fatalf("walkexpr ORECV") // should see inside OAS only
		panic("unreachable")

	case ir.OSLICEHEADER:
		n := n.(*ir.SliceHeaderExpr)
		return walkSliceHeader(n, init)

	case ir.OSLICE, ir.OSLICEARR, ir.OSLICESTR, ir.OSLICE3, ir.OSLICE3ARR:
		n := n.(*ir.SliceExpr)
		return walkSlice(n, init)

	case ir.ONEW:
		n := n.(*ir.UnaryExpr)
		return walkNew(n, init)

	case ir.OADDSTR:
		return walkAddString(n.(*ir.AddStringExpr), init)

	case ir.OAPPEND:
		// order should make sure we only see OAS(node, OAPPEND), which we handle above.
		base.Fatalf("append outside assignment")
		panic("unreachable")

	case ir.OCOPY:
		return walkCopy(n.(*ir.BinaryExpr), init, base.Flag.Cfg.Instrumenting && !base.Flag.CompilingRuntime)

	case ir.OCLOSE:
		n := n.(*ir.UnaryExpr)
		return walkClose(n, init)

	case ir.OMAKECHAN:
		n := n.(*ir.MakeExpr)
		return walkMakeChan(n, init)

	case ir.OMAKEMAP:
		n := n.(*ir.MakeExpr)
		return walkMakeMap(n, init)

	case ir.OMAKESLICE:
		n := n.(*ir.MakeExpr)
		return walkMakeSlice(n, init)

	case ir.OMAKESLICECOPY:
		n := n.(*ir.MakeExpr)
		return walkMakeSliceCopy(n, init)

	case ir.ORUNESTR:
		n := n.(*ir.ConvExpr)
		return walkRuneToString(n, init)

	case ir.OBYTES2STR, ir.ORUNES2STR:
		n := n.(*ir.ConvExpr)
		return walkBytesRunesToString(n, init)

	case ir.OBYTES2STRTMP:
		n := n.(*ir.ConvExpr)
		return walkBytesToStringTemp(n, init)

	case ir.OSTR2BYTES:
		n := n.(*ir.ConvExpr)
		return walkStringToBytes(n, init)

	case ir.OSTR2BYTESTMP:
		n := n.(*ir.ConvExpr)
		return walkStringToBytesTemp(n, init)

	case ir.OSTR2RUNES:
		n := n.(*ir.ConvExpr)
		return walkStringToRunes(n, init)

	case ir.OARRAYLIT, ir.OSLICELIT, ir.OMAPLIT, ir.OSTRUCTLIT, ir.OPTRLIT:
		return walkCompLit(n, init)

	case ir.OSEND:
		n := n.(*ir.SendStmt)
		return walkSend(n, init)

	case ir.OCLOSURE:
		return walkClosure(n.(*ir.ClosureExpr), init)

	case ir.OCALLPART:
		return walkCallPart(n.(*ir.CallPartExpr), init)
	}

	// No return! Each case must return (or panic),
	// to avoid confusion about what gets returned
	// in the presence of type assertions.
}

// walk the whole tree of the body of an
// expression or simple statement.
// the types expressions are calculated.
// compile-time constants are evaluated.
// complex side effects like statements are appended to init
func walkExprList(s []ir.Node, init *ir.Nodes) {
	for i := range s {
		s[i] = walkExpr(s[i], init)
	}
}

func walkExprListCheap(s []ir.Node, init *ir.Nodes) {
	for i, n := range s {
		s[i] = cheapexpr(n, init)
		s[i] = walkExpr(s[i], init)
	}
}

func walkExprListSafe(s []ir.Node, init *ir.Nodes) {
	for i, n := range s {
		s[i] = safeexpr(n, init)
		s[i] = walkExpr(s[i], init)
	}
}

func walkAddString(n *ir.AddStringExpr, init *ir.Nodes) ir.Node {
	c := len(n.List)

	if c < 2 {
		base.Fatalf("addstr count %d too small", c)
	}

	buf := typecheck.NodNil()
	if n.Esc() == ir.EscNone {
		sz := int64(0)
		for _, n1 := range n.List {
			if n1.Op() == ir.OLITERAL {
				sz += int64(len(ir.StringVal(n1)))
			}
		}

		// Don't allocate the buffer if the result won't fit.
		if sz < tmpstringbufsize {
			// Create temporary buffer for result string on stack.
			t := types.NewArray(types.Types[types.TUINT8], tmpstringbufsize)
			buf = typecheck.NodAddr(typecheck.Temp(t))
		}
	}

	// build list of string arguments
	args := []ir.Node{buf}
	for _, n2 := range n.List {
		args = append(args, typecheck.Conv(n2, types.Types[types.TSTRING]))
	}

	var fn string
	if c <= 5 {
		// small numbers of strings use direct runtime helpers.
		// note: order.expr knows this cutoff too.
		fn = fmt.Sprintf("concatstring%d", c)
	} else {
		// large numbers of strings are passed to the runtime as a slice.
		fn = "concatstrings"

		t := types.NewSlice(types.Types[types.TSTRING])
		// args[1:] to skip buf arg
		slice := ir.NewCompLitExpr(base.Pos, ir.OCOMPLIT, ir.TypeNode(t), args[1:])
		slice.Prealloc = n.Prealloc
		args = []ir.Node{buf, slice}
		slice.SetEsc(ir.EscNone)
	}

	cat := typecheck.LookupRuntime(fn)
	r := ir.NewCallExpr(base.Pos, ir.OCALL, cat, nil)
	r.Args.Set(args)
	r1 := typecheck.Expr(r)
	r1 = walkExpr(r1, init)
	r1.SetType(n.Type())

	return r1
}

// Rewrite append(src, x, y, z) so that any side effects in
// x, y, z (including runtime panics) are evaluated in
// initialization statements before the append.
// For normal code generation, stop there and leave the
// rest to cgen_append.
//
// For race detector, expand append(src, a [, b]* ) to
//
//   init {
//     s := src
//     const argc = len(args) - 1
//     if cap(s) - len(s) < argc {
//	    s = growslice(s, len(s)+argc)
//     }
//     n := len(s)
//     s = s[:n+argc]
//     s[n] = a
//     s[n+1] = b
//     ...
//   }
//   s
func walkAppend(n *ir.CallExpr, init *ir.Nodes, dst ir.Node) ir.Node {
	if !ir.SameSafeExpr(dst, n.Args[0]) {
		n.Args[0] = safeexpr(n.Args[0], init)
		n.Args[0] = walkExpr(n.Args[0], init)
	}
	walkExprListSafe(n.Args[1:], init)

	nsrc := n.Args[0]

	// walkexprlistsafe will leave OINDEX (s[n]) alone if both s
	// and n are name or literal, but those may index the slice we're
	// modifying here. Fix explicitly.
	// Using cheapexpr also makes sure that the evaluation
	// of all arguments (and especially any panics) happen
	// before we begin to modify the slice in a visible way.
	ls := n.Args[1:]
	for i, n := range ls {
		n = cheapexpr(n, init)
		if !types.Identical(n.Type(), nsrc.Type().Elem()) {
			n = typecheck.AssignConv(n, nsrc.Type().Elem(), "append")
			n = walkExpr(n, init)
		}
		ls[i] = n
	}

	argc := len(n.Args) - 1
	if argc < 1 {
		return nsrc
	}

	// General case, with no function calls left as arguments.
	// Leave for gen, except that instrumentation requires old form.
	if !base.Flag.Cfg.Instrumenting || base.Flag.CompilingRuntime {
		return n
	}

	var l []ir.Node

	ns := typecheck.Temp(nsrc.Type())
	l = append(l, ir.NewAssignStmt(base.Pos, ns, nsrc)) // s = src

	na := ir.NewInt(int64(argc))                 // const argc
	nif := ir.NewIfStmt(base.Pos, nil, nil, nil) // if cap(s) - len(s) < argc
	nif.Cond = ir.NewBinaryExpr(base.Pos, ir.OLT, ir.NewBinaryExpr(base.Pos, ir.OSUB, ir.NewUnaryExpr(base.Pos, ir.OCAP, ns), ir.NewUnaryExpr(base.Pos, ir.OLEN, ns)), na)

	fn := typecheck.LookupRuntime("growslice") //   growslice(<type>, old []T, mincap int) (ret []T)
	fn = typecheck.SubstArgTypes(fn, ns.Type().Elem(), ns.Type().Elem())

	nif.Body = []ir.Node{ir.NewAssignStmt(base.Pos, ns, mkcall1(fn, ns.Type(), nif.PtrInit(), reflectdata.TypePtr(ns.Type().Elem()), ns,
		ir.NewBinaryExpr(base.Pos, ir.OADD, ir.NewUnaryExpr(base.Pos, ir.OLEN, ns), na)))}

	l = append(l, nif)

	nn := typecheck.Temp(types.Types[types.TINT])
	l = append(l, ir.NewAssignStmt(base.Pos, nn, ir.NewUnaryExpr(base.Pos, ir.OLEN, ns))) // n = len(s)

	slice := ir.NewSliceExpr(base.Pos, ir.OSLICE, ns) // ...s[:n+argc]
	slice.SetSliceBounds(nil, ir.NewBinaryExpr(base.Pos, ir.OADD, nn, na), nil)
	slice.SetBounded(true)
	l = append(l, ir.NewAssignStmt(base.Pos, ns, slice)) // s = s[:n+argc]

	ls = n.Args[1:]
	for i, n := range ls {
		ix := ir.NewIndexExpr(base.Pos, ns, nn) // s[n] ...
		ix.SetBounded(true)
		l = append(l, ir.NewAssignStmt(base.Pos, ix, n)) // s[n] = arg
		if i+1 < len(ls) {
			l = append(l, ir.NewAssignStmt(base.Pos, nn, ir.NewBinaryExpr(base.Pos, ir.OADD, nn, ir.NewInt(1)))) // n = n + 1
		}
	}

	typecheck.Stmts(l)
	walkStmtList(l)
	init.Append(l...)
	return ns
}

// walkAssign walks an OAS (AssignExpr) or OASOP (AssignOpExpr) node.
func walkAssign(init *ir.Nodes, n ir.Node) ir.Node {
	init.Append(n.PtrInit().Take()...)

	var left, right ir.Node
	switch n.Op() {
	case ir.OAS:
		n := n.(*ir.AssignStmt)
		left, right = n.X, n.Y
	case ir.OASOP:
		n := n.(*ir.AssignOpStmt)
		left, right = n.X, n.Y
	}

	// Recognize m[k] = append(m[k], ...) so we can reuse
	// the mapassign call.
	var mapAppend *ir.CallExpr
	if left.Op() == ir.OINDEXMAP && right.Op() == ir.OAPPEND {
		left := left.(*ir.IndexExpr)
		mapAppend = right.(*ir.CallExpr)
		if !ir.SameSafeExpr(left, mapAppend.Args[0]) {
			base.Fatalf("not same expressions: %v != %v", left, mapAppend.Args[0])
		}
	}

	left = walkExpr(left, init)
	left = safeexpr(left, init)
	if mapAppend != nil {
		mapAppend.Args[0] = left
	}

	if n.Op() == ir.OASOP {
		// Rewrite x op= y into x = x op y.
		n = ir.NewAssignStmt(base.Pos, left, typecheck.Expr(ir.NewBinaryExpr(base.Pos, n.(*ir.AssignOpStmt).AsOp, left, right)))
	} else {
		n.(*ir.AssignStmt).X = left
	}
	as := n.(*ir.AssignStmt)

	if oaslit(as, init) {
		return ir.NewBlockStmt(as.Pos(), nil)
	}

	if as.Y == nil {
		// TODO(austin): Check all "implicit zeroing"
		return as
	}

	if !base.Flag.Cfg.Instrumenting && ir.IsZero(as.Y) {
		return as
	}

	switch as.Y.Op() {
	default:
		as.Y = walkExpr(as.Y, init)

	case ir.ORECV:
		// x = <-c; as.Left is x, as.Right.Left is c.
		// order.stmt made sure x is addressable.
		recv := as.Y.(*ir.UnaryExpr)
		recv.X = walkExpr(recv.X, init)

		n1 := typecheck.NodAddr(as.X)
		r := recv.X // the channel
		return mkcall1(chanfn("chanrecv1", 2, r.Type()), nil, init, r, n1)

	case ir.OAPPEND:
		// x = append(...)
		call := as.Y.(*ir.CallExpr)
		if call.Type().Elem().NotInHeap() {
			base.Errorf("%v can't be allocated in Go; it is incomplete (or unallocatable)", call.Type().Elem())
		}
		var r ir.Node
		switch {
		case isAppendOfMake(call):
			// x = append(y, make([]T, y)...)
			r = extendslice(call, init)
		case call.IsDDD:
			r = appendslice(call, init) // also works for append(slice, string).
		default:
			r = walkAppend(call, init, as)
		}
		as.Y = r
		if r.Op() == ir.OAPPEND {
			// Left in place for back end.
			// Do not add a new write barrier.
			// Set up address of type for back end.
			r.(*ir.CallExpr).X = reflectdata.TypePtr(r.Type().Elem())
			return as
		}
		// Otherwise, lowered for race detector.
		// Treat as ordinary assignment.
	}

	if as.X != nil && as.Y != nil {
		return convas(as, init)
	}
	return as
}

// walkAssignDotType walks an OAS2DOTTYPE node.
func walkAssignDotType(n *ir.AssignListStmt, init *ir.Nodes) ir.Node {
	walkExprListSafe(n.Lhs, init)
	n.Rhs[0] = walkExpr(n.Rhs[0], init)
	return n
}

// walkAssignFunc walks an OAS2FUNC node.
func walkAssignFunc(init *ir.Nodes, n *ir.AssignListStmt) ir.Node {
	init.Append(n.PtrInit().Take()...)

	r := n.Rhs[0]
	walkExprListSafe(n.Lhs, init)
	r = walkExpr(r, init)

	if ir.IsIntrinsicCall(r.(*ir.CallExpr)) {
		n.Rhs = []ir.Node{r}
		return n
	}
	init.Append(r)

	ll := ascompatet(n.Lhs, r.Type())
	return ir.NewBlockStmt(src.NoXPos, ll)
}

// walkAssignList walks an OAS2 node.
func walkAssignList(init *ir.Nodes, n *ir.AssignListStmt) ir.Node {
	init.Append(n.PtrInit().Take()...)
	walkExprListSafe(n.Lhs, init)
	walkExprListSafe(n.Rhs, init)
	return ir.NewBlockStmt(src.NoXPos, ascompatee(ir.OAS, n.Lhs, n.Rhs, init))
}

// walkAssignMapRead walks an OAS2MAPR node.
func walkAssignMapRead(init *ir.Nodes, n *ir.AssignListStmt) ir.Node {
	init.Append(n.PtrInit().Take()...)

	r := n.Rhs[0].(*ir.IndexExpr)
	walkExprListSafe(n.Lhs, init)
	r.X = walkExpr(r.X, init)
	r.Index = walkExpr(r.Index, init)
	t := r.X.Type()

	fast := mapfast(t)
	var key ir.Node
	if fast != mapslow {
		// fast versions take key by value
		key = r.Index
	} else {
		// standard version takes key by reference
		// order.expr made sure key is addressable.
		key = typecheck.NodAddr(r.Index)
	}

	// from:
	//   a,b = m[i]
	// to:
	//   var,b = mapaccess2*(t, m, i)
	//   a = *var
	a := n.Lhs[0]

	var call *ir.CallExpr
	if w := t.Elem().Width; w <= zeroValSize {
		fn := mapfn(mapaccess2[fast], t)
		call = mkcall1(fn, fn.Type().Results(), init, reflectdata.TypePtr(t), r.X, key)
	} else {
		fn := mapfn("mapaccess2_fat", t)
		z := reflectdata.ZeroAddr(w)
		call = mkcall1(fn, fn.Type().Results(), init, reflectdata.TypePtr(t), r.X, key, z)
	}

	// mapaccess2* returns a typed bool, but due to spec changes,
	// the boolean result of i.(T) is now untyped so we make it the
	// same type as the variable on the lhs.
	if ok := n.Lhs[1]; !ir.IsBlank(ok) && ok.Type().IsBoolean() {
		call.Type().Field(1).Type = ok.Type()
	}
	n.Rhs = []ir.Node{call}
	n.SetOp(ir.OAS2FUNC)

	// don't generate a = *var if a is _
	if ir.IsBlank(a) {
		return walkExpr(typecheck.Stmt(n), init)
	}

	var_ := typecheck.Temp(types.NewPtr(t.Elem()))
	var_.SetTypecheck(1)
	var_.MarkNonNil() // mapaccess always returns a non-nil pointer

	n.Lhs[0] = var_
	init.Append(walkExpr(n, init))

	as := ir.NewAssignStmt(base.Pos, a, ir.NewStarExpr(base.Pos, var_))
	return walkExpr(typecheck.Stmt(as), init)
}

// walkAssignRecv walks an OAS2RECV node.
func walkAssignRecv(init *ir.Nodes, n *ir.AssignListStmt) ir.Node {
	init.Append(n.PtrInit().Take()...)

	r := n.Rhs[0].(*ir.UnaryExpr) // recv
	walkExprListSafe(n.Lhs, init)
	r.X = walkExpr(r.X, init)
	var n1 ir.Node
	if ir.IsBlank(n.Lhs[0]) {
		n1 = typecheck.NodNil()
	} else {
		n1 = typecheck.NodAddr(n.Lhs[0])
	}
	fn := chanfn("chanrecv2", 2, r.X.Type())
	ok := n.Lhs[1]
	call := mkcall1(fn, types.Types[types.TBOOL], init, r.X, n1)
	return typecheck.Stmt(ir.NewAssignStmt(base.Pos, ok, call))
}

// walkBytesRunesToString walks an OBYTES2STR or ORUNES2STR node.
func walkBytesRunesToString(n *ir.ConvExpr, init *ir.Nodes) ir.Node {
	a := typecheck.NodNil()
	if n.Esc() == ir.EscNone {
		// Create temporary buffer for string on stack.
		t := types.NewArray(types.Types[types.TUINT8], tmpstringbufsize)
		a = typecheck.NodAddr(typecheck.Temp(t))
	}
	if n.Op() == ir.ORUNES2STR {
		// slicerunetostring(*[32]byte, []rune) string
		return mkcall("slicerunetostring", n.Type(), init, a, n.X)
	}
	// slicebytetostring(*[32]byte, ptr *byte, n int) string
	n.X = cheapexpr(n.X, init)
	ptr, len := backingArrayPtrLen(n.X)
	return mkcall("slicebytetostring", n.Type(), init, a, ptr, len)
}

// walkBytesToStringTemp walks an OBYTES2STRTMP node.
func walkBytesToStringTemp(n *ir.ConvExpr, init *ir.Nodes) ir.Node {
	n.X = walkExpr(n.X, init)
	if !base.Flag.Cfg.Instrumenting {
		// Let the backend handle OBYTES2STRTMP directly
		// to avoid a function call to slicebytetostringtmp.
		return n
	}
	// slicebytetostringtmp(ptr *byte, n int) string
	n.X = cheapexpr(n.X, init)
	ptr, len := backingArrayPtrLen(n.X)
	return mkcall("slicebytetostringtmp", n.Type(), init, ptr, len)
}

// walkCall walks an OCALLFUNC, OCALLINTER, or OCALLMETH node.
func walkCall(n *ir.CallExpr, init *ir.Nodes) ir.Node {
	if n.Op() == ir.OCALLINTER {
		usemethod(n)
		markUsedIfaceMethod(n)
	}

	if n.Op() == ir.OCALLFUNC && n.X.Op() == ir.OCLOSURE {
		// Transform direct call of a closure to call of a normal function.
		// transformclosure already did all preparation work.

		// Prepend captured variables to argument list.
		clo := n.X.(*ir.ClosureExpr)
		n.Args.Prepend(clo.Func.ClosureEnter...)
		clo.Func.ClosureEnter.Set(nil)

		// Replace OCLOSURE with ONAME/PFUNC.
		n.X = clo.Func.Nname

		// Update type of OCALLFUNC node.
		// Output arguments had not changed, but their offsets could.
		if n.X.Type().NumResults() == 1 {
			n.SetType(n.X.Type().Results().Field(0).Type)
		} else {
			n.SetType(n.X.Type().Results())
		}
	}

	walkCall1(n, init)
	return n
}

func walkCall1(n *ir.CallExpr, init *ir.Nodes) {
	if len(n.Rargs) != 0 {
		return // already walked
	}

	params := n.X.Type().Params()
	args := n.Args

	n.X = walkExpr(n.X, init)
	walkExprList(args, init)

	// If this is a method call, add the receiver at the beginning of the args.
	if n.Op() == ir.OCALLMETH {
		withRecv := make([]ir.Node, len(args)+1)
		dot := n.X.(*ir.SelectorExpr)
		withRecv[0] = dot.X
		dot.X = nil
		copy(withRecv[1:], args)
		args = withRecv
	}

	// For any argument whose evaluation might require a function call,
	// store that argument into a temporary variable,
	// to prevent that calls from clobbering arguments already on the stack.
	// When instrumenting, all arguments might require function calls.
	var tempAssigns []ir.Node
	for i, arg := range args {
		updateHasCall(arg)
		// Determine param type.
		var t *types.Type
		if n.Op() == ir.OCALLMETH {
			if i == 0 {
				t = n.X.Type().Recv().Type
			} else {
				t = params.Field(i - 1).Type
			}
		} else {
			t = params.Field(i).Type
		}
		if base.Flag.Cfg.Instrumenting || fncall(arg, t) {
			// make assignment of fncall to tempAt
			tmp := typecheck.Temp(t)
			a := convas(ir.NewAssignStmt(base.Pos, tmp, arg), init)
			tempAssigns = append(tempAssigns, a)
			// replace arg with temp
			args[i] = tmp
		}
	}

	n.Args.Set(tempAssigns)
	n.Rargs.Set(args)
}

// walkClose walks an OCLOSE node.
func walkClose(n *ir.UnaryExpr, init *ir.Nodes) ir.Node {
	// cannot use chanfn - closechan takes any, not chan any
	fn := typecheck.LookupRuntime("closechan")
	fn = typecheck.SubstArgTypes(fn, n.X.Type())
	return mkcall1(fn, nil, init, n.X)
}

// walkCompLit walks a composite literal node:
// OARRAYLIT, OSLICELIT, OMAPLIT, OSTRUCTLIT (all CompLitExpr), or OPTRLIT (AddrExpr).
func walkCompLit(n ir.Node, init *ir.Nodes) ir.Node {
	if isStaticCompositeLiteral(n) && !ssagen.TypeOK(n.Type()) {
		n := n.(*ir.CompLitExpr) // not OPTRLIT
		// n can be directly represented in the read-only data section.
		// Make direct reference to the static data. See issue 12841.
		vstat := readonlystaticname(n.Type())
		fixedlit(inInitFunction, initKindStatic, n, vstat, init)
		return typecheck.Expr(vstat)
	}
	var_ := typecheck.Temp(n.Type())
	anylit(n, var_, init)
	return var_
}

// The result of walkcompare MUST be assigned back to n, e.g.
// 	n.Left = walkcompare(n.Left, init)
func walkCompare(n *ir.BinaryExpr, init *ir.Nodes) ir.Node {
	if n.X.Type().IsInterface() && n.Y.Type().IsInterface() && n.X.Op() != ir.ONIL && n.Y.Op() != ir.ONIL {
		return walkCompareInterface(n, init)
	}

	if n.X.Type().IsString() && n.Y.Type().IsString() {
		return walkCompareString(n, init)
	}

	n.X = walkExpr(n.X, init)
	n.Y = walkExpr(n.Y, init)

	// Given mixed interface/concrete comparison,
	// rewrite into types-equal && data-equal.
	// This is efficient, avoids allocations, and avoids runtime calls.
	if n.X.Type().IsInterface() != n.Y.Type().IsInterface() {
		// Preserve side-effects in case of short-circuiting; see #32187.
		l := cheapexpr(n.X, init)
		r := cheapexpr(n.Y, init)
		// Swap so that l is the interface value and r is the concrete value.
		if n.Y.Type().IsInterface() {
			l, r = r, l
		}

		// Handle both == and !=.
		eq := n.Op()
		andor := ir.OOROR
		if eq == ir.OEQ {
			andor = ir.OANDAND
		}
		// Check for types equal.
		// For empty interface, this is:
		//   l.tab == type(r)
		// For non-empty interface, this is:
		//   l.tab != nil && l.tab._type == type(r)
		var eqtype ir.Node
		tab := ir.NewUnaryExpr(base.Pos, ir.OITAB, l)
		rtyp := reflectdata.TypePtr(r.Type())
		if l.Type().IsEmptyInterface() {
			tab.SetType(types.NewPtr(types.Types[types.TUINT8]))
			tab.SetTypecheck(1)
			eqtype = ir.NewBinaryExpr(base.Pos, eq, tab, rtyp)
		} else {
			nonnil := ir.NewBinaryExpr(base.Pos, brcom(eq), typecheck.NodNil(), tab)
			match := ir.NewBinaryExpr(base.Pos, eq, itabType(tab), rtyp)
			eqtype = ir.NewLogicalExpr(base.Pos, andor, nonnil, match)
		}
		// Check for data equal.
		eqdata := ir.NewBinaryExpr(base.Pos, eq, ifaceData(n.Pos(), l, r.Type()), r)
		// Put it all together.
		expr := ir.NewLogicalExpr(base.Pos, andor, eqtype, eqdata)
		return finishcompare(n, expr, init)
	}

	// Must be comparison of array or struct.
	// Otherwise back end handles it.
	// While we're here, decide whether to
	// inline or call an eq alg.
	t := n.X.Type()
	var inline bool

	maxcmpsize := int64(4)
	unalignedLoad := canMergeLoads()
	if unalignedLoad {
		// Keep this low enough to generate less code than a function call.
		maxcmpsize = 2 * int64(ssagen.Arch.LinkArch.RegSize)
	}

	switch t.Kind() {
	default:
		if base.Debug.Libfuzzer != 0 && t.IsInteger() {
			n.X = cheapexpr(n.X, init)
			n.Y = cheapexpr(n.Y, init)

			// If exactly one comparison operand is
			// constant, invoke the constcmp functions
			// instead, and arrange for the constant
			// operand to be the first argument.
			l, r := n.X, n.Y
			if r.Op() == ir.OLITERAL {
				l, r = r, l
			}
			constcmp := l.Op() == ir.OLITERAL && r.Op() != ir.OLITERAL

			var fn string
			var paramType *types.Type
			switch t.Size() {
			case 1:
				fn = "libfuzzerTraceCmp1"
				if constcmp {
					fn = "libfuzzerTraceConstCmp1"
				}
				paramType = types.Types[types.TUINT8]
			case 2:
				fn = "libfuzzerTraceCmp2"
				if constcmp {
					fn = "libfuzzerTraceConstCmp2"
				}
				paramType = types.Types[types.TUINT16]
			case 4:
				fn = "libfuzzerTraceCmp4"
				if constcmp {
					fn = "libfuzzerTraceConstCmp4"
				}
				paramType = types.Types[types.TUINT32]
			case 8:
				fn = "libfuzzerTraceCmp8"
				if constcmp {
					fn = "libfuzzerTraceConstCmp8"
				}
				paramType = types.Types[types.TUINT64]
			default:
				base.Fatalf("unexpected integer size %d for %v", t.Size(), t)
			}
			init.Append(mkcall(fn, nil, init, tracecmpArg(l, paramType, init), tracecmpArg(r, paramType, init)))
		}
		return n
	case types.TARRAY:
		// We can compare several elements at once with 2/4/8 byte integer compares
		inline = t.NumElem() <= 1 || (types.IsSimple[t.Elem().Kind()] && (t.NumElem() <= 4 || t.Elem().Width*t.NumElem() <= maxcmpsize))
	case types.TSTRUCT:
		inline = t.NumComponents(types.IgnoreBlankFields) <= 4
	}

	cmpl := n.X
	for cmpl != nil && cmpl.Op() == ir.OCONVNOP {
		cmpl = cmpl.(*ir.ConvExpr).X
	}
	cmpr := n.Y
	for cmpr != nil && cmpr.Op() == ir.OCONVNOP {
		cmpr = cmpr.(*ir.ConvExpr).X
	}

	// Chose not to inline. Call equality function directly.
	if !inline {
		// eq algs take pointers; cmpl and cmpr must be addressable
		if !ir.IsAssignable(cmpl) || !ir.IsAssignable(cmpr) {
			base.Fatalf("arguments of comparison must be lvalues - %v %v", cmpl, cmpr)
		}

		fn, needsize := eqfor(t)
		call := ir.NewCallExpr(base.Pos, ir.OCALL, fn, nil)
		call.Args.Append(typecheck.NodAddr(cmpl))
		call.Args.Append(typecheck.NodAddr(cmpr))
		if needsize {
			call.Args.Append(ir.NewInt(t.Width))
		}
		res := ir.Node(call)
		if n.Op() != ir.OEQ {
			res = ir.NewUnaryExpr(base.Pos, ir.ONOT, res)
		}
		return finishcompare(n, res, init)
	}

	// inline: build boolean expression comparing element by element
	andor := ir.OANDAND
	if n.Op() == ir.ONE {
		andor = ir.OOROR
	}
	var expr ir.Node
	compare := func(el, er ir.Node) {
		a := ir.NewBinaryExpr(base.Pos, n.Op(), el, er)
		if expr == nil {
			expr = a
		} else {
			expr = ir.NewLogicalExpr(base.Pos, andor, expr, a)
		}
	}
	cmpl = safeexpr(cmpl, init)
	cmpr = safeexpr(cmpr, init)
	if t.IsStruct() {
		for _, f := range t.Fields().Slice() {
			sym := f.Sym
			if sym.IsBlank() {
				continue
			}
			compare(
				ir.NewSelectorExpr(base.Pos, ir.OXDOT, cmpl, sym),
				ir.NewSelectorExpr(base.Pos, ir.OXDOT, cmpr, sym),
			)
		}
	} else {
		step := int64(1)
		remains := t.NumElem() * t.Elem().Width
		combine64bit := unalignedLoad && types.RegSize == 8 && t.Elem().Width <= 4 && t.Elem().IsInteger()
		combine32bit := unalignedLoad && t.Elem().Width <= 2 && t.Elem().IsInteger()
		combine16bit := unalignedLoad && t.Elem().Width == 1 && t.Elem().IsInteger()
		for i := int64(0); remains > 0; {
			var convType *types.Type
			switch {
			case remains >= 8 && combine64bit:
				convType = types.Types[types.TINT64]
				step = 8 / t.Elem().Width
			case remains >= 4 && combine32bit:
				convType = types.Types[types.TUINT32]
				step = 4 / t.Elem().Width
			case remains >= 2 && combine16bit:
				convType = types.Types[types.TUINT16]
				step = 2 / t.Elem().Width
			default:
				step = 1
			}
			if step == 1 {
				compare(
					ir.NewIndexExpr(base.Pos, cmpl, ir.NewInt(i)),
					ir.NewIndexExpr(base.Pos, cmpr, ir.NewInt(i)),
				)
				i++
				remains -= t.Elem().Width
			} else {
				elemType := t.Elem().ToUnsigned()
				cmplw := ir.Node(ir.NewIndexExpr(base.Pos, cmpl, ir.NewInt(i)))
				cmplw = typecheck.Conv(cmplw, elemType) // convert to unsigned
				cmplw = typecheck.Conv(cmplw, convType) // widen
				cmprw := ir.Node(ir.NewIndexExpr(base.Pos, cmpr, ir.NewInt(i)))
				cmprw = typecheck.Conv(cmprw, elemType)
				cmprw = typecheck.Conv(cmprw, convType)
				// For code like this:  uint32(s[0]) | uint32(s[1])<<8 | uint32(s[2])<<16 ...
				// ssa will generate a single large load.
				for offset := int64(1); offset < step; offset++ {
					lb := ir.Node(ir.NewIndexExpr(base.Pos, cmpl, ir.NewInt(i+offset)))
					lb = typecheck.Conv(lb, elemType)
					lb = typecheck.Conv(lb, convType)
					lb = ir.NewBinaryExpr(base.Pos, ir.OLSH, lb, ir.NewInt(8*t.Elem().Width*offset))
					cmplw = ir.NewBinaryExpr(base.Pos, ir.OOR, cmplw, lb)
					rb := ir.Node(ir.NewIndexExpr(base.Pos, cmpr, ir.NewInt(i+offset)))
					rb = typecheck.Conv(rb, elemType)
					rb = typecheck.Conv(rb, convType)
					rb = ir.NewBinaryExpr(base.Pos, ir.OLSH, rb, ir.NewInt(8*t.Elem().Width*offset))
					cmprw = ir.NewBinaryExpr(base.Pos, ir.OOR, cmprw, rb)
				}
				compare(cmplw, cmprw)
				i += step
				remains -= step * t.Elem().Width
			}
		}
	}
	if expr == nil {
		expr = ir.NewBool(n.Op() == ir.OEQ)
		// We still need to use cmpl and cmpr, in case they contain
		// an expression which might panic. See issue 23837.
		t := typecheck.Temp(cmpl.Type())
		a1 := typecheck.Stmt(ir.NewAssignStmt(base.Pos, t, cmpl))
		a2 := typecheck.Stmt(ir.NewAssignStmt(base.Pos, t, cmpr))
		init.Append(a1, a2)
	}
	return finishcompare(n, expr, init)
}

func walkCompareInterface(n *ir.BinaryExpr, init *ir.Nodes) ir.Node {
	n.Y = cheapexpr(n.Y, init)
	n.X = cheapexpr(n.X, init)
	eqtab, eqdata := reflectdata.EqInterface(n.X, n.Y)
	var cmp ir.Node
	if n.Op() == ir.OEQ {
		cmp = ir.NewLogicalExpr(base.Pos, ir.OANDAND, eqtab, eqdata)
	} else {
		eqtab.SetOp(ir.ONE)
		cmp = ir.NewLogicalExpr(base.Pos, ir.OOROR, eqtab, ir.NewUnaryExpr(base.Pos, ir.ONOT, eqdata))
	}
	return finishcompare(n, cmp, init)
}

func walkCompareString(n *ir.BinaryExpr, init *ir.Nodes) ir.Node {
	// Rewrite comparisons to short constant strings as length+byte-wise comparisons.
	var cs, ncs ir.Node // const string, non-const string
	switch {
	case ir.IsConst(n.X, constant.String) && ir.IsConst(n.Y, constant.String):
		// ignore; will be constant evaluated
	case ir.IsConst(n.X, constant.String):
		cs = n.X
		ncs = n.Y
	case ir.IsConst(n.Y, constant.String):
		cs = n.Y
		ncs = n.X
	}
	if cs != nil {
		cmp := n.Op()
		// Our comparison below assumes that the non-constant string
		// is on the left hand side, so rewrite "" cmp x to x cmp "".
		// See issue 24817.
		if ir.IsConst(n.X, constant.String) {
			cmp = brrev(cmp)
		}

		// maxRewriteLen was chosen empirically.
		// It is the value that minimizes cmd/go file size
		// across most architectures.
		// See the commit description for CL 26758 for details.
		maxRewriteLen := 6
		// Some architectures can load unaligned byte sequence as 1 word.
		// So we can cover longer strings with the same amount of code.
		canCombineLoads := canMergeLoads()
		combine64bit := false
		if canCombineLoads {
			// Keep this low enough to generate less code than a function call.
			maxRewriteLen = 2 * ssagen.Arch.LinkArch.RegSize
			combine64bit = ssagen.Arch.LinkArch.RegSize >= 8
		}

		var and ir.Op
		switch cmp {
		case ir.OEQ:
			and = ir.OANDAND
		case ir.ONE:
			and = ir.OOROR
		default:
			// Don't do byte-wise comparisons for <, <=, etc.
			// They're fairly complicated.
			// Length-only checks are ok, though.
			maxRewriteLen = 0
		}
		if s := ir.StringVal(cs); len(s) <= maxRewriteLen {
			if len(s) > 0 {
				ncs = safeexpr(ncs, init)
			}
			r := ir.Node(ir.NewBinaryExpr(base.Pos, cmp, ir.NewUnaryExpr(base.Pos, ir.OLEN, ncs), ir.NewInt(int64(len(s)))))
			remains := len(s)
			for i := 0; remains > 0; {
				if remains == 1 || !canCombineLoads {
					cb := ir.NewInt(int64(s[i]))
					ncb := ir.NewIndexExpr(base.Pos, ncs, ir.NewInt(int64(i)))
					r = ir.NewLogicalExpr(base.Pos, and, r, ir.NewBinaryExpr(base.Pos, cmp, ncb, cb))
					remains--
					i++
					continue
				}
				var step int
				var convType *types.Type
				switch {
				case remains >= 8 && combine64bit:
					convType = types.Types[types.TINT64]
					step = 8
				case remains >= 4:
					convType = types.Types[types.TUINT32]
					step = 4
				case remains >= 2:
					convType = types.Types[types.TUINT16]
					step = 2
				}
				ncsubstr := typecheck.Conv(ir.NewIndexExpr(base.Pos, ncs, ir.NewInt(int64(i))), convType)
				csubstr := int64(s[i])
				// Calculate large constant from bytes as sequence of shifts and ors.
				// Like this:  uint32(s[0]) | uint32(s[1])<<8 | uint32(s[2])<<16 ...
				// ssa will combine this into a single large load.
				for offset := 1; offset < step; offset++ {
					b := typecheck.Conv(ir.NewIndexExpr(base.Pos, ncs, ir.NewInt(int64(i+offset))), convType)
					b = ir.NewBinaryExpr(base.Pos, ir.OLSH, b, ir.NewInt(int64(8*offset)))
					ncsubstr = ir.NewBinaryExpr(base.Pos, ir.OOR, ncsubstr, b)
					csubstr |= int64(s[i+offset]) << uint8(8*offset)
				}
				csubstrPart := ir.NewInt(csubstr)
				// Compare "step" bytes as once
				r = ir.NewLogicalExpr(base.Pos, and, r, ir.NewBinaryExpr(base.Pos, cmp, csubstrPart, ncsubstr))
				remains -= step
				i += step
			}
			return finishcompare(n, r, init)
		}
	}

	var r ir.Node
	if n.Op() == ir.OEQ || n.Op() == ir.ONE {
		// prepare for rewrite below
		n.X = cheapexpr(n.X, init)
		n.Y = cheapexpr(n.Y, init)
		eqlen, eqmem := reflectdata.EqString(n.X, n.Y)
		// quick check of len before full compare for == or !=.
		// memequal then tests equality up to length len.
		if n.Op() == ir.OEQ {
			// len(left) == len(right) && memequal(left, right, len)
			r = ir.NewLogicalExpr(base.Pos, ir.OANDAND, eqlen, eqmem)
		} else {
			// len(left) != len(right) || !memequal(left, right, len)
			eqlen.SetOp(ir.ONE)
			r = ir.NewLogicalExpr(base.Pos, ir.OOROR, eqlen, ir.NewUnaryExpr(base.Pos, ir.ONOT, eqmem))
		}
	} else {
		// sys_cmpstring(s1, s2) :: 0
		r = mkcall("cmpstring", types.Types[types.TINT], init, typecheck.Conv(n.X, types.Types[types.TSTRING]), typecheck.Conv(n.Y, types.Types[types.TSTRING]))
		r = ir.NewBinaryExpr(base.Pos, n.Op(), r, ir.NewInt(0))
	}

	return finishcompare(n, r, init)
}

// walkConv walks an OCONV or OCONVNOP (but not OCONVIFACE) node.
func walkConv(n *ir.ConvExpr, init *ir.Nodes) ir.Node {
	n.X = walkExpr(n.X, init)
	if n.Op() == ir.OCONVNOP && n.Type() == n.X.Type() {
		return n.X
	}
	if n.Op() == ir.OCONVNOP && ir.ShouldCheckPtr(ir.CurFunc, 1) {
		if n.Type().IsPtr() && n.X.Type().IsUnsafePtr() { // unsafe.Pointer to *T
			return walkCheckPtrAlignment(n, init, nil)
		}
		if n.Type().IsUnsafePtr() && n.X.Type().IsUintptr() { // uintptr to unsafe.Pointer
			return walkCheckPtrArithmetic(n, init)
		}
	}
	param, result := rtconvfn(n.X.Type(), n.Type())
	if param == types.Txxx {
		return n
	}
	fn := types.BasicTypeNames[param] + "to" + types.BasicTypeNames[result]
	return typecheck.Conv(mkcall(fn, types.Types[result], init, typecheck.Conv(n.X, types.Types[param])), n.Type())
}

// walkConvInterface walks an OCONVIFACE node.
func walkConvInterface(n *ir.ConvExpr, init *ir.Nodes) ir.Node {
	n.X = walkExpr(n.X, init)

	fromType := n.X.Type()
	toType := n.Type()

	if !fromType.IsInterface() && !ir.IsBlank(ir.CurFunc.Nname) { // skip unnamed functions (func _())
		markTypeUsedInInterface(fromType, ir.CurFunc.LSym)
	}

	// typeword generates the type word of the interface value.
	typeword := func() ir.Node {
		if toType.IsEmptyInterface() {
			return reflectdata.TypePtr(fromType)
		}
		return reflectdata.ITabAddr(fromType, toType)
	}

	// Optimize convT2E or convT2I as a two-word copy when T is pointer-shaped.
	if types.IsInterfaceWord(fromType) {
		l := ir.NewBinaryExpr(base.Pos, ir.OEFACE, typeword(), n.X)
		l.SetType(toType)
		l.SetTypecheck(n.Typecheck())
		return l
	}

	if ir.Names.Staticuint64s == nil {
		ir.Names.Staticuint64s = typecheck.NewName(ir.Pkgs.Runtime.Lookup("staticuint64s"))
		ir.Names.Staticuint64s.Class_ = ir.PEXTERN
		// The actual type is [256]uint64, but we use [256*8]uint8 so we can address
		// individual bytes.
		ir.Names.Staticuint64s.SetType(types.NewArray(types.Types[types.TUINT8], 256*8))
		ir.Names.Zerobase = typecheck.NewName(ir.Pkgs.Runtime.Lookup("zerobase"))
		ir.Names.Zerobase.Class_ = ir.PEXTERN
		ir.Names.Zerobase.SetType(types.Types[types.TUINTPTR])
	}

	// Optimize convT2{E,I} for many cases in which T is not pointer-shaped,
	// by using an existing addressable value identical to n.Left
	// or creating one on the stack.
	var value ir.Node
	switch {
	case fromType.Size() == 0:
		// n.Left is zero-sized. Use zerobase.
		cheapexpr(n.X, init) // Evaluate n.Left for side-effects. See issue 19246.
		value = ir.Names.Zerobase
	case fromType.IsBoolean() || (fromType.Size() == 1 && fromType.IsInteger()):
		// n.Left is a bool/byte. Use staticuint64s[n.Left * 8] on little-endian
		// and staticuint64s[n.Left * 8 + 7] on big-endian.
		n.X = cheapexpr(n.X, init)
		// byteindex widens n.Left so that the multiplication doesn't overflow.
		index := ir.NewBinaryExpr(base.Pos, ir.OLSH, byteindex(n.X), ir.NewInt(3))
		if ssagen.Arch.LinkArch.ByteOrder == binary.BigEndian {
			index = ir.NewBinaryExpr(base.Pos, ir.OADD, index, ir.NewInt(7))
		}
		xe := ir.NewIndexExpr(base.Pos, ir.Names.Staticuint64s, index)
		xe.SetBounded(true)
		value = xe
	case n.X.Op() == ir.ONAME && n.X.(*ir.Name).Class_ == ir.PEXTERN && n.X.(*ir.Name).Readonly():
		// n.Left is a readonly global; use it directly.
		value = n.X
	case !fromType.IsInterface() && n.Esc() == ir.EscNone && fromType.Width <= 1024:
		// n.Left does not escape. Use a stack temporary initialized to n.Left.
		value = typecheck.Temp(fromType)
		init.Append(typecheck.Stmt(ir.NewAssignStmt(base.Pos, value, n.X)))
	}

	if value != nil {
		// Value is identical to n.Left.
		// Construct the interface directly: {type/itab, &value}.
		l := ir.NewBinaryExpr(base.Pos, ir.OEFACE, typeword(), typecheck.Expr(typecheck.NodAddr(value)))
		l.SetType(toType)
		l.SetTypecheck(n.Typecheck())
		return l
	}

	// Implement interface to empty interface conversion.
	// tmp = i.itab
	// if tmp != nil {
	//    tmp = tmp.type
	// }
	// e = iface{tmp, i.data}
	if toType.IsEmptyInterface() && fromType.IsInterface() && !fromType.IsEmptyInterface() {
		// Evaluate the input interface.
		c := typecheck.Temp(fromType)
		init.Append(ir.NewAssignStmt(base.Pos, c, n.X))

		// Get the itab out of the interface.
		tmp := typecheck.Temp(types.NewPtr(types.Types[types.TUINT8]))
		init.Append(ir.NewAssignStmt(base.Pos, tmp, typecheck.Expr(ir.NewUnaryExpr(base.Pos, ir.OITAB, c))))

		// Get the type out of the itab.
		nif := ir.NewIfStmt(base.Pos, typecheck.Expr(ir.NewBinaryExpr(base.Pos, ir.ONE, tmp, typecheck.NodNil())), nil, nil)
		nif.Body = []ir.Node{ir.NewAssignStmt(base.Pos, tmp, itabType(tmp))}
		init.Append(nif)

		// Build the result.
		e := ir.NewBinaryExpr(base.Pos, ir.OEFACE, tmp, ifaceData(n.Pos(), c, types.NewPtr(types.Types[types.TUINT8])))
		e.SetType(toType) // assign type manually, typecheck doesn't understand OEFACE.
		e.SetTypecheck(1)
		return e
	}

	fnname, needsaddr := convFuncName(fromType, toType)

	if !needsaddr && !fromType.IsInterface() {
		// Use a specialized conversion routine that only returns a data pointer.
		// ptr = convT2X(val)
		// e = iface{typ/tab, ptr}
		fn := typecheck.LookupRuntime(fnname)
		types.CalcSize(fromType)
		fn = typecheck.SubstArgTypes(fn, fromType)
		types.CalcSize(fn.Type())
		call := ir.NewCallExpr(base.Pos, ir.OCALL, fn, nil)
		call.Args = []ir.Node{n.X}
		e := ir.NewBinaryExpr(base.Pos, ir.OEFACE, typeword(), safeexpr(walkExpr(typecheck.Expr(call), init), init))
		e.SetType(toType)
		e.SetTypecheck(1)
		return e
	}

	var tab ir.Node
	if fromType.IsInterface() {
		// convI2I
		tab = reflectdata.TypePtr(toType)
	} else {
		// convT2x
		tab = typeword()
	}

	v := n.X
	if needsaddr {
		// Types of large or unknown size are passed by reference.
		// Orderexpr arranged for n.Left to be a temporary for all
		// the conversions it could see. Comparison of an interface
		// with a non-interface, especially in a switch on interface value
		// with non-interface cases, is not visible to order.stmt, so we
		// have to fall back on allocating a temp here.
		if !ir.IsAssignable(v) {
			v = copyexpr(v, v.Type(), init)
		}
		v = typecheck.NodAddr(v)
	}

	types.CalcSize(fromType)
	fn := typecheck.LookupRuntime(fnname)
	fn = typecheck.SubstArgTypes(fn, fromType, toType)
	types.CalcSize(fn.Type())
	call := ir.NewCallExpr(base.Pos, ir.OCALL, fn, nil)
	call.Args = []ir.Node{tab, v}
	return walkExpr(typecheck.Expr(call), init)
}

// Lower copy(a, b) to a memmove call or a runtime call.
//
// init {
//   n := len(a)
//   if n > len(b) { n = len(b) }
//   if a.ptr != b.ptr { memmove(a.ptr, b.ptr, n*sizeof(elem(a))) }
// }
// n;
//
// Also works if b is a string.
//
func walkCopy(n *ir.BinaryExpr, init *ir.Nodes, runtimecall bool) ir.Node {
	if n.X.Type().Elem().HasPointers() {
		ir.CurFunc.SetWBPos(n.Pos())
		fn := writebarrierfn("typedslicecopy", n.X.Type().Elem(), n.Y.Type().Elem())
		n.X = cheapexpr(n.X, init)
		ptrL, lenL := backingArrayPtrLen(n.X)
		n.Y = cheapexpr(n.Y, init)
		ptrR, lenR := backingArrayPtrLen(n.Y)
		return mkcall1(fn, n.Type(), init, reflectdata.TypePtr(n.X.Type().Elem()), ptrL, lenL, ptrR, lenR)
	}

	if runtimecall {
		// rely on runtime to instrument:
		//  copy(n.Left, n.Right)
		// n.Right can be a slice or string.

		n.X = cheapexpr(n.X, init)
		ptrL, lenL := backingArrayPtrLen(n.X)
		n.Y = cheapexpr(n.Y, init)
		ptrR, lenR := backingArrayPtrLen(n.Y)

		fn := typecheck.LookupRuntime("slicecopy")
		fn = typecheck.SubstArgTypes(fn, ptrL.Type().Elem(), ptrR.Type().Elem())

		return mkcall1(fn, n.Type(), init, ptrL, lenL, ptrR, lenR, ir.NewInt(n.X.Type().Elem().Width))
	}

	n.X = walkExpr(n.X, init)
	n.Y = walkExpr(n.Y, init)
	nl := typecheck.Temp(n.X.Type())
	nr := typecheck.Temp(n.Y.Type())
	var l []ir.Node
	l = append(l, ir.NewAssignStmt(base.Pos, nl, n.X))
	l = append(l, ir.NewAssignStmt(base.Pos, nr, n.Y))

	nfrm := ir.NewUnaryExpr(base.Pos, ir.OSPTR, nr)
	nto := ir.NewUnaryExpr(base.Pos, ir.OSPTR, nl)

	nlen := typecheck.Temp(types.Types[types.TINT])

	// n = len(to)
	l = append(l, ir.NewAssignStmt(base.Pos, nlen, ir.NewUnaryExpr(base.Pos, ir.OLEN, nl)))

	// if n > len(frm) { n = len(frm) }
	nif := ir.NewIfStmt(base.Pos, nil, nil, nil)

	nif.Cond = ir.NewBinaryExpr(base.Pos, ir.OGT, nlen, ir.NewUnaryExpr(base.Pos, ir.OLEN, nr))
	nif.Body.Append(ir.NewAssignStmt(base.Pos, nlen, ir.NewUnaryExpr(base.Pos, ir.OLEN, nr)))
	l = append(l, nif)

	// if to.ptr != frm.ptr { memmove( ... ) }
	ne := ir.NewIfStmt(base.Pos, ir.NewBinaryExpr(base.Pos, ir.ONE, nto, nfrm), nil, nil)
	ne.Likely = true
	l = append(l, ne)

	fn := typecheck.LookupRuntime("memmove")
	fn = typecheck.SubstArgTypes(fn, nl.Type().Elem(), nl.Type().Elem())
	nwid := ir.Node(typecheck.Temp(types.Types[types.TUINTPTR]))
	setwid := ir.NewAssignStmt(base.Pos, nwid, typecheck.Conv(nlen, types.Types[types.TUINTPTR]))
	ne.Body.Append(setwid)
	nwid = ir.NewBinaryExpr(base.Pos, ir.OMUL, nwid, ir.NewInt(nl.Type().Elem().Width))
	call := mkcall1(fn, nil, init, nto, nfrm, nwid)
	ne.Body.Append(call)

	typecheck.Stmts(l)
	walkStmtList(l)
	init.Append(l...)
	return nlen
}

// walkDelete walks an ODELETE node.
func walkDelete(init *ir.Nodes, n *ir.CallExpr) ir.Node {
	init.Append(n.PtrInit().Take()...)
	map_ := n.Args[0]
	key := n.Args[1]
	map_ = walkExpr(map_, init)
	key = walkExpr(key, init)

	t := map_.Type()
	fast := mapfast(t)
	if fast == mapslow {
		// order.stmt made sure key is addressable.
		key = typecheck.NodAddr(key)
	}
	return mkcall1(mapfndel(mapdelete[fast], t), nil, init, reflectdata.TypePtr(t), map_, key)
}

// walkDivMod walks an ODIV or OMOD node.
func walkDivMod(n *ir.BinaryExpr, init *ir.Nodes) ir.Node {
	n.X = walkExpr(n.X, init)
	n.Y = walkExpr(n.Y, init)

	// rewrite complex div into function call.
	et := n.X.Type().Kind()

	if types.IsComplex[et] && n.Op() == ir.ODIV {
		t := n.Type()
		call := mkcall("complex128div", types.Types[types.TCOMPLEX128], init, typecheck.Conv(n.X, types.Types[types.TCOMPLEX128]), typecheck.Conv(n.Y, types.Types[types.TCOMPLEX128]))
		return typecheck.Conv(call, t)
	}

	// Nothing to do for float divisions.
	if types.IsFloat[et] {
		return n
	}

	// rewrite 64-bit div and mod on 32-bit architectures.
	// TODO: Remove this code once we can introduce
	// runtime calls late in SSA processing.
	if types.RegSize < 8 && (et == types.TINT64 || et == types.TUINT64) {
		if n.Y.Op() == ir.OLITERAL {
			// Leave div/mod by constant powers of 2 or small 16-bit constants.
			// The SSA backend will handle those.
			switch et {
			case types.TINT64:
				c := ir.Int64Val(n.Y)
				if c < 0 {
					c = -c
				}
				if c != 0 && c&(c-1) == 0 {
					return n
				}
			case types.TUINT64:
				c := ir.Uint64Val(n.Y)
				if c < 1<<16 {
					return n
				}
				if c != 0 && c&(c-1) == 0 {
					return n
				}
			}
		}
		var fn string
		if et == types.TINT64 {
			fn = "int64"
		} else {
			fn = "uint64"
		}
		if n.Op() == ir.ODIV {
			fn += "div"
		} else {
			fn += "mod"
		}
		return mkcall(fn, n.Type(), init, typecheck.Conv(n.X, types.Types[et]), typecheck.Conv(n.Y, types.Types[et]))
	}
	return n
}

// walkDot walks an ODOT or ODOTPTR node.
func walkDot(n *ir.SelectorExpr, init *ir.Nodes) ir.Node {
	usefield(n)
	n.X = walkExpr(n.X, init)
	return n
}

// walkDotType walks an ODOTTYPE or ODOTTYPE2 node.
func walkDotType(n *ir.TypeAssertExpr, init *ir.Nodes) ir.Node {
	n.X = walkExpr(n.X, init)
	// Set up interface type addresses for back end.
	n.Ntype = reflectdata.TypePtr(n.Type())
	if n.Op() == ir.ODOTTYPE {
		n.Ntype.(*ir.AddrExpr).Alloc = reflectdata.TypePtr(n.X.Type())
	}
	if !n.Type().IsInterface() && !n.X.Type().IsEmptyInterface() {
		n.Itab = []ir.Node{reflectdata.ITabAddr(n.Type(), n.X.Type())}
	}
	return n
}

// walkIndex walks an OINDEX node.
func walkIndex(n *ir.IndexExpr, init *ir.Nodes) ir.Node {
	n.X = walkExpr(n.X, init)

	// save the original node for bounds checking elision.
	// If it was a ODIV/OMOD walk might rewrite it.
	r := n.Index

	n.Index = walkExpr(n.Index, init)

	// if range of type cannot exceed static array bound,
	// disable bounds check.
	if n.Bounded() {
		return n
	}
	t := n.X.Type()
	if t != nil && t.IsPtr() {
		t = t.Elem()
	}
	if t.IsArray() {
		n.SetBounded(bounded(r, t.NumElem()))
		if base.Flag.LowerM != 0 && n.Bounded() && !ir.IsConst(n.Index, constant.Int) {
			base.Warn("index bounds check elided")
		}
		if ir.IsSmallIntConst(n.Index) && !n.Bounded() {
			base.Errorf("index out of bounds")
		}
	} else if ir.IsConst(n.X, constant.String) {
		n.SetBounded(bounded(r, int64(len(ir.StringVal(n.X)))))
		if base.Flag.LowerM != 0 && n.Bounded() && !ir.IsConst(n.Index, constant.Int) {
			base.Warn("index bounds check elided")
		}
		if ir.IsSmallIntConst(n.Index) && !n.Bounded() {
			base.Errorf("index out of bounds")
		}
	}

	if ir.IsConst(n.Index, constant.Int) {
		if v := n.Index.Val(); constant.Sign(v) < 0 || ir.ConstOverflow(v, types.Types[types.TINT]) {
			base.Errorf("index out of bounds")
		}
	}
	return n
}

// walkIndexMap walks an OINDEXMAP node.
func walkIndexMap(n *ir.IndexExpr, init *ir.Nodes) ir.Node {
	// Replace m[k] with *map{access1,assign}(maptype, m, &k)
	n.X = walkExpr(n.X, init)
	n.Index = walkExpr(n.Index, init)
	map_ := n.X
	key := n.Index
	t := map_.Type()
	var call *ir.CallExpr
	if n.Assigned {
		// This m[k] expression is on the left-hand side of an assignment.
		fast := mapfast(t)
		if fast == mapslow {
			// standard version takes key by reference.
			// order.expr made sure key is addressable.
			key = typecheck.NodAddr(key)
		}
		call = mkcall1(mapfn(mapassign[fast], t), nil, init, reflectdata.TypePtr(t), map_, key)
	} else {
		// m[k] is not the target of an assignment.
		fast := mapfast(t)
		if fast == mapslow {
			// standard version takes key by reference.
			// order.expr made sure key is addressable.
			key = typecheck.NodAddr(key)
		}

		if w := t.Elem().Width; w <= zeroValSize {
			call = mkcall1(mapfn(mapaccess1[fast], t), types.NewPtr(t.Elem()), init, reflectdata.TypePtr(t), map_, key)
		} else {
			z := reflectdata.ZeroAddr(w)
			call = mkcall1(mapfn("mapaccess1_fat", t), types.NewPtr(t.Elem()), init, reflectdata.TypePtr(t), map_, key, z)
		}
	}
	call.SetType(types.NewPtr(t.Elem()))
	call.MarkNonNil() // mapaccess1* and mapassign always return non-nil pointers.
	star := ir.NewStarExpr(base.Pos, call)
	star.SetType(t.Elem())
	star.SetTypecheck(1)
	return star
}

// walkLenCap walks an OLEN or OCAP node.
func walkLenCap(n *ir.UnaryExpr, init *ir.Nodes) ir.Node {
	if isRuneCount(n) {
		// Replace len([]rune(string)) with runtime.countrunes(string).
		return mkcall("countrunes", n.Type(), init, typecheck.Conv(n.X.(*ir.ConvExpr).X, types.Types[types.TSTRING]))
	}

	n.X = walkExpr(n.X, init)

	// replace len(*[10]int) with 10.
	// delayed until now to preserve side effects.
	t := n.X.Type()

	if t.IsPtr() {
		t = t.Elem()
	}
	if t.IsArray() {
		safeexpr(n.X, init)
		con := typecheck.OrigInt(n, t.NumElem())
		con.SetTypecheck(1)
		return con
	}
	return n
}

// walkLogical walks an OANDAND or OOROR node.
func walkLogical(n *ir.LogicalExpr, init *ir.Nodes) ir.Node {
	n.X = walkExpr(n.X, init)

	// cannot put side effects from n.Right on init,
	// because they cannot run before n.Left is checked.
	// save elsewhere and store on the eventual n.Right.
	var ll ir.Nodes

	n.Y = walkExpr(n.Y, &ll)
	n.Y = ir.InitExpr(ll, n.Y)
	return n
}

// walkMakeChan walks an OMAKECHAN node.
func walkMakeChan(n *ir.MakeExpr, init *ir.Nodes) ir.Node {
	// When size fits into int, use makechan instead of
	// makechan64, which is faster and shorter on 32 bit platforms.
	size := n.Len
	fnname := "makechan64"
	argtype := types.Types[types.TINT64]

	// Type checking guarantees that TIDEAL size is positive and fits in an int.
	// The case of size overflow when converting TUINT or TUINTPTR to TINT
	// will be handled by the negative range checks in makechan during runtime.
	if size.Type().IsKind(types.TIDEAL) || size.Type().Size() <= types.Types[types.TUINT].Size() {
		fnname = "makechan"
		argtype = types.Types[types.TINT]
	}

	return mkcall1(chanfn(fnname, 1, n.Type()), n.Type(), init, reflectdata.TypePtr(n.Type()), typecheck.Conv(size, argtype))
}

// walkMakeMap walks an OMAKEMAP node.
func walkMakeMap(n *ir.MakeExpr, init *ir.Nodes) ir.Node {
	t := n.Type()
	hmapType := reflectdata.MapType(t)
	hint := n.Len

	// var h *hmap
	var h ir.Node
	if n.Esc() == ir.EscNone {
		// Allocate hmap on stack.

		// var hv hmap
		hv := typecheck.Temp(hmapType)
		init.Append(typecheck.Stmt(ir.NewAssignStmt(base.Pos, hv, nil)))
		// h = &hv
		h = typecheck.NodAddr(hv)

		// Allocate one bucket pointed to by hmap.buckets on stack if hint
		// is not larger than BUCKETSIZE. In case hint is larger than
		// BUCKETSIZE runtime.makemap will allocate the buckets on the heap.
		// Maximum key and elem size is 128 bytes, larger objects
		// are stored with an indirection. So max bucket size is 2048+eps.
		if !ir.IsConst(hint, constant.Int) ||
			constant.Compare(hint.Val(), token.LEQ, constant.MakeInt64(reflectdata.BUCKETSIZE)) {

			// In case hint is larger than BUCKETSIZE runtime.makemap
			// will allocate the buckets on the heap, see #20184
			//
			// if hint <= BUCKETSIZE {
			//     var bv bmap
			//     b = &bv
			//     h.buckets = b
			// }

			nif := ir.NewIfStmt(base.Pos, ir.NewBinaryExpr(base.Pos, ir.OLE, hint, ir.NewInt(reflectdata.BUCKETSIZE)), nil, nil)
			nif.Likely = true

			// var bv bmap
			bv := typecheck.Temp(reflectdata.MapBucketType(t))
			nif.Body.Append(ir.NewAssignStmt(base.Pos, bv, nil))

			// b = &bv
			b := typecheck.NodAddr(bv)

			// h.buckets = b
			bsym := hmapType.Field(5).Sym // hmap.buckets see reflect.go:hmap
			na := ir.NewAssignStmt(base.Pos, ir.NewSelectorExpr(base.Pos, ir.ODOT, h, bsym), b)
			nif.Body.Append(na)
			appendWalkStmt(init, nif)
		}
	}

	if ir.IsConst(hint, constant.Int) && constant.Compare(hint.Val(), token.LEQ, constant.MakeInt64(reflectdata.BUCKETSIZE)) {
		// Handling make(map[any]any) and
		// make(map[any]any, hint) where hint <= BUCKETSIZE
		// special allows for faster map initialization and
		// improves binary size by using calls with fewer arguments.
		// For hint <= BUCKETSIZE overLoadFactor(hint, 0) is false
		// and no buckets will be allocated by makemap. Therefore,
		// no buckets need to be allocated in this code path.
		if n.Esc() == ir.EscNone {
			// Only need to initialize h.hash0 since
			// hmap h has been allocated on the stack already.
			// h.hash0 = fastrand()
			rand := mkcall("fastrand", types.Types[types.TUINT32], init)
			hashsym := hmapType.Field(4).Sym // hmap.hash0 see reflect.go:hmap
			appendWalkStmt(init, ir.NewAssignStmt(base.Pos, ir.NewSelectorExpr(base.Pos, ir.ODOT, h, hashsym), rand))
			return typecheck.ConvNop(h, t)
		}
		// Call runtime.makehmap to allocate an
		// hmap on the heap and initialize hmap's hash0 field.
		fn := typecheck.LookupRuntime("makemap_small")
		fn = typecheck.SubstArgTypes(fn, t.Key(), t.Elem())
		return mkcall1(fn, n.Type(), init)
	}

	if n.Esc() != ir.EscNone {
		h = typecheck.NodNil()
	}
	// Map initialization with a variable or large hint is
	// more complicated. We therefore generate a call to
	// runtime.makemap to initialize hmap and allocate the
	// map buckets.

	// When hint fits into int, use makemap instead of
	// makemap64, which is faster and shorter on 32 bit platforms.
	fnname := "makemap64"
	argtype := types.Types[types.TINT64]

	// Type checking guarantees that TIDEAL hint is positive and fits in an int.
	// See checkmake call in TMAP case of OMAKE case in OpSwitch in typecheck1 function.
	// The case of hint overflow when converting TUINT or TUINTPTR to TINT
	// will be handled by the negative range checks in makemap during runtime.
	if hint.Type().IsKind(types.TIDEAL) || hint.Type().Size() <= types.Types[types.TUINT].Size() {
		fnname = "makemap"
		argtype = types.Types[types.TINT]
	}

	fn := typecheck.LookupRuntime(fnname)
	fn = typecheck.SubstArgTypes(fn, hmapType, t.Key(), t.Elem())
	return mkcall1(fn, n.Type(), init, reflectdata.TypePtr(n.Type()), typecheck.Conv(hint, argtype), h)
}

// walkMakeSlice walks an OMAKESLICE node.
func walkMakeSlice(n *ir.MakeExpr, init *ir.Nodes) ir.Node {
	l := n.Len
	r := n.Cap
	if r == nil {
		r = safeexpr(l, init)
		l = r
	}
	t := n.Type()
	if t.Elem().NotInHeap() {
		base.Errorf("%v can't be allocated in Go; it is incomplete (or unallocatable)", t.Elem())
	}
	if n.Esc() == ir.EscNone {
		if why := escape.HeapAllocReason(n); why != "" {
			base.Fatalf("%v has EscNone, but %v", n, why)
		}
		// var arr [r]T
		// n = arr[:l]
		i := typecheck.IndexConst(r)
		if i < 0 {
			base.Fatalf("walkexpr: invalid index %v", r)
		}

		// cap is constrained to [0,2^31) or [0,2^63) depending on whether
		// we're in 32-bit or 64-bit systems. So it's safe to do:
		//
		// if uint64(len) > cap {
		//     if len < 0 { panicmakeslicelen() }
		//     panicmakeslicecap()
		// }
		nif := ir.NewIfStmt(base.Pos, ir.NewBinaryExpr(base.Pos, ir.OGT, typecheck.Conv(l, types.Types[types.TUINT64]), ir.NewInt(i)), nil, nil)
		niflen := ir.NewIfStmt(base.Pos, ir.NewBinaryExpr(base.Pos, ir.OLT, l, ir.NewInt(0)), nil, nil)
		niflen.Body = []ir.Node{mkcall("panicmakeslicelen", nil, init)}
		nif.Body.Append(niflen, mkcall("panicmakeslicecap", nil, init))
		init.Append(typecheck.Stmt(nif))

		t = types.NewArray(t.Elem(), i) // [r]T
		var_ := typecheck.Temp(t)
		appendWalkStmt(init, ir.NewAssignStmt(base.Pos, var_, nil)) // zero temp
		r := ir.NewSliceExpr(base.Pos, ir.OSLICE, var_)             // arr[:l]
		r.SetSliceBounds(nil, l, nil)
		// The conv is necessary in case n.Type is named.
		return walkExpr(typecheck.Expr(typecheck.Conv(r, n.Type())), init)
	}

	// n escapes; set up a call to makeslice.
	// When len and cap can fit into int, use makeslice instead of
	// makeslice64, which is faster and shorter on 32 bit platforms.

	len, cap := l, r

	fnname := "makeslice64"
	argtype := types.Types[types.TINT64]

	// Type checking guarantees that TIDEAL len/cap are positive and fit in an int.
	// The case of len or cap overflow when converting TUINT or TUINTPTR to TINT
	// will be handled by the negative range checks in makeslice during runtime.
	if (len.Type().IsKind(types.TIDEAL) || len.Type().Size() <= types.Types[types.TUINT].Size()) &&
		(cap.Type().IsKind(types.TIDEAL) || cap.Type().Size() <= types.Types[types.TUINT].Size()) {
		fnname = "makeslice"
		argtype = types.Types[types.TINT]
	}

	m := ir.NewSliceHeaderExpr(base.Pos, nil, nil, nil, nil)
	m.SetType(t)

	fn := typecheck.LookupRuntime(fnname)
	m.Ptr = mkcall1(fn, types.Types[types.TUNSAFEPTR], init, reflectdata.TypePtr(t.Elem()), typecheck.Conv(len, argtype), typecheck.Conv(cap, argtype))
	m.Ptr.MarkNonNil()
	m.LenCap = []ir.Node{typecheck.Conv(len, types.Types[types.TINT]), typecheck.Conv(cap, types.Types[types.TINT])}
	return walkExpr(typecheck.Expr(m), init)
}

// walkMakeSliceCopy walks an OMAKESLICECOPY node.
func walkMakeSliceCopy(n *ir.MakeExpr, init *ir.Nodes) ir.Node {
	if n.Esc() == ir.EscNone {
		base.Fatalf("OMAKESLICECOPY with EscNone: %v", n)
	}

	t := n.Type()
	if t.Elem().NotInHeap() {
		base.Errorf("%v can't be allocated in Go; it is incomplete (or unallocatable)", t.Elem())
	}

	length := typecheck.Conv(n.Len, types.Types[types.TINT])
	copylen := ir.NewUnaryExpr(base.Pos, ir.OLEN, n.Cap)
	copyptr := ir.NewUnaryExpr(base.Pos, ir.OSPTR, n.Cap)

	if !t.Elem().HasPointers() && n.Bounded() {
		// When len(to)==len(from) and elements have no pointers:
		// replace make+copy with runtime.mallocgc+runtime.memmove.

		// We do not check for overflow of len(to)*elem.Width here
		// since len(from) is an existing checked slice capacity
		// with same elem.Width for the from slice.
		size := ir.NewBinaryExpr(base.Pos, ir.OMUL, typecheck.Conv(length, types.Types[types.TUINTPTR]), typecheck.Conv(ir.NewInt(t.Elem().Width), types.Types[types.TUINTPTR]))

		// instantiate mallocgc(size uintptr, typ *byte, needszero bool) unsafe.Pointer
		fn := typecheck.LookupRuntime("mallocgc")
		sh := ir.NewSliceHeaderExpr(base.Pos, nil, nil, nil, nil)
		sh.Ptr = mkcall1(fn, types.Types[types.TUNSAFEPTR], init, size, typecheck.NodNil(), ir.NewBool(false))
		sh.Ptr.MarkNonNil()
		sh.LenCap = []ir.Node{length, length}
		sh.SetType(t)

		s := typecheck.Temp(t)
		r := typecheck.Stmt(ir.NewAssignStmt(base.Pos, s, sh))
		r = walkExpr(r, init)
		init.Append(r)

		// instantiate memmove(to *any, frm *any, size uintptr)
		fn = typecheck.LookupRuntime("memmove")
		fn = typecheck.SubstArgTypes(fn, t.Elem(), t.Elem())
		ncopy := mkcall1(fn, nil, init, ir.NewUnaryExpr(base.Pos, ir.OSPTR, s), copyptr, size)
		init.Append(walkExpr(typecheck.Stmt(ncopy), init))

		return s
	}
	// Replace make+copy with runtime.makeslicecopy.
	// instantiate makeslicecopy(typ *byte, tolen int, fromlen int, from unsafe.Pointer) unsafe.Pointer
	fn := typecheck.LookupRuntime("makeslicecopy")
	s := ir.NewSliceHeaderExpr(base.Pos, nil, nil, nil, nil)
	s.Ptr = mkcall1(fn, types.Types[types.TUNSAFEPTR], init, reflectdata.TypePtr(t.Elem()), length, copylen, typecheck.Conv(copyptr, types.Types[types.TUNSAFEPTR]))
	s.Ptr.MarkNonNil()
	s.LenCap = []ir.Node{length, length}
	s.SetType(t)
	return walkExpr(typecheck.Expr(s), init)
}

// walkNew walks an ONEW node.
func walkNew(n *ir.UnaryExpr, init *ir.Nodes) ir.Node {
	if n.Type().Elem().NotInHeap() {
		base.Errorf("%v can't be allocated in Go; it is incomplete (or unallocatable)", n.Type().Elem())
	}
	if n.Esc() == ir.EscNone {
		if n.Type().Elem().Width >= ir.MaxImplicitStackVarSize {
			base.Fatalf("large ONEW with EscNone: %v", n)
		}
		r := typecheck.Temp(n.Type().Elem())
		init.Append(typecheck.Stmt(ir.NewAssignStmt(base.Pos, r, nil))) // zero temp
		return typecheck.Expr(typecheck.NodAddr(r))
	}
	return callnew(n.Type().Elem())
}

// generate code for print
func walkPrint(nn *ir.CallExpr, init *ir.Nodes) ir.Node {
	// Hoist all the argument evaluation up before the lock.
	walkExprListCheap(nn.Args, init)

	// For println, add " " between elements and "\n" at the end.
	if nn.Op() == ir.OPRINTN {
		s := nn.Args
		t := make([]ir.Node, 0, len(s)*2)
		for i, n := range s {
			if i != 0 {
				t = append(t, ir.NewString(" "))
			}
			t = append(t, n)
		}
		t = append(t, ir.NewString("\n"))
		nn.Args.Set(t)
	}

	// Collapse runs of constant strings.
	s := nn.Args
	t := make([]ir.Node, 0, len(s))
	for i := 0; i < len(s); {
		var strs []string
		for i < len(s) && ir.IsConst(s[i], constant.String) {
			strs = append(strs, ir.StringVal(s[i]))
			i++
		}
		if len(strs) > 0 {
			t = append(t, ir.NewString(strings.Join(strs, "")))
		}
		if i < len(s) {
			t = append(t, s[i])
			i++
		}
	}
	nn.Args.Set(t)

	calls := []ir.Node{mkcall("printlock", nil, init)}
	for i, n := range nn.Args {
		if n.Op() == ir.OLITERAL {
			if n.Type() == types.UntypedRune {
				n = typecheck.DefaultLit(n, types.RuneType)
			}

			switch n.Val().Kind() {
			case constant.Int:
				n = typecheck.DefaultLit(n, types.Types[types.TINT64])

			case constant.Float:
				n = typecheck.DefaultLit(n, types.Types[types.TFLOAT64])
			}
		}

		if n.Op() != ir.OLITERAL && n.Type() != nil && n.Type().Kind() == types.TIDEAL {
			n = typecheck.DefaultLit(n, types.Types[types.TINT64])
		}
		n = typecheck.DefaultLit(n, nil)
		nn.Args[i] = n
		if n.Type() == nil || n.Type().Kind() == types.TFORW {
			continue
		}

		var on *ir.Name
		switch n.Type().Kind() {
		case types.TINTER:
			if n.Type().IsEmptyInterface() {
				on = typecheck.LookupRuntime("printeface")
			} else {
				on = typecheck.LookupRuntime("printiface")
			}
			on = typecheck.SubstArgTypes(on, n.Type()) // any-1
		case types.TPTR:
			if n.Type().Elem().NotInHeap() {
				on = typecheck.LookupRuntime("printuintptr")
				n = ir.NewConvExpr(base.Pos, ir.OCONV, nil, n)
				n.SetType(types.Types[types.TUNSAFEPTR])
				n = ir.NewConvExpr(base.Pos, ir.OCONV, nil, n)
				n.SetType(types.Types[types.TUINTPTR])
				break
			}
			fallthrough
		case types.TCHAN, types.TMAP, types.TFUNC, types.TUNSAFEPTR:
			on = typecheck.LookupRuntime("printpointer")
			on = typecheck.SubstArgTypes(on, n.Type()) // any-1
		case types.TSLICE:
			on = typecheck.LookupRuntime("printslice")
			on = typecheck.SubstArgTypes(on, n.Type()) // any-1
		case types.TUINT, types.TUINT8, types.TUINT16, types.TUINT32, types.TUINT64, types.TUINTPTR:
			if types.IsRuntimePkg(n.Type().Sym().Pkg) && n.Type().Sym().Name == "hex" {
				on = typecheck.LookupRuntime("printhex")
			} else {
				on = typecheck.LookupRuntime("printuint")
			}
		case types.TINT, types.TINT8, types.TINT16, types.TINT32, types.TINT64:
			on = typecheck.LookupRuntime("printint")
		case types.TFLOAT32, types.TFLOAT64:
			on = typecheck.LookupRuntime("printfloat")
		case types.TCOMPLEX64, types.TCOMPLEX128:
			on = typecheck.LookupRuntime("printcomplex")
		case types.TBOOL:
			on = typecheck.LookupRuntime("printbool")
		case types.TSTRING:
			cs := ""
			if ir.IsConst(n, constant.String) {
				cs = ir.StringVal(n)
			}
			switch cs {
			case " ":
				on = typecheck.LookupRuntime("printsp")
			case "\n":
				on = typecheck.LookupRuntime("printnl")
			default:
				on = typecheck.LookupRuntime("printstring")
			}
		default:
			badtype(ir.OPRINT, n.Type(), nil)
			continue
		}

		r := ir.NewCallExpr(base.Pos, ir.OCALL, on, nil)
		if params := on.Type().Params().FieldSlice(); len(params) > 0 {
			t := params[0].Type
			if !types.Identical(t, n.Type()) {
				n = ir.NewConvExpr(base.Pos, ir.OCONV, nil, n)
				n.SetType(t)
			}
			r.Args.Append(n)
		}
		calls = append(calls, r)
	}

	calls = append(calls, mkcall("printunlock", nil, init))

	typecheck.Stmts(calls)
	walkExprList(calls, init)

	r := ir.NewBlockStmt(base.Pos, nil)
	r.List.Set(calls)
	return walkStmt(typecheck.Stmt(r))
}

// walkRuneToString walks an ORUNESTR node.
func walkRuneToString(n *ir.ConvExpr, init *ir.Nodes) ir.Node {
	a := typecheck.NodNil()
	if n.Esc() == ir.EscNone {
		t := types.NewArray(types.Types[types.TUINT8], 4)
		a = typecheck.NodAddr(typecheck.Temp(t))
	}
	// intstring(*[4]byte, rune)
	return mkcall("intstring", n.Type(), init, a, typecheck.Conv(n.X, types.Types[types.TINT64]))
}

// walkSend walks an OSEND node.
func walkSend(n *ir.SendStmt, init *ir.Nodes) ir.Node {
	n1 := n.Value
	n1 = typecheck.AssignConv(n1, n.Chan.Type().Elem(), "chan send")
	n1 = walkExpr(n1, init)
	n1 = typecheck.NodAddr(n1)
	return mkcall1(chanfn("chansend1", 2, n.Chan.Type()), nil, init, n.Chan, n1)
}

// walkSlice walks an OSLICE, OSLICEARR, OSLICESTR, OSLICE3, or OSLICE3ARR node.
func walkSlice(n *ir.SliceExpr, init *ir.Nodes) ir.Node {

	checkSlice := ir.ShouldCheckPtr(ir.CurFunc, 1) && n.Op() == ir.OSLICE3ARR && n.X.Op() == ir.OCONVNOP && n.X.(*ir.ConvExpr).X.Type().IsUnsafePtr()
	if checkSlice {
		conv := n.X.(*ir.ConvExpr)
		conv.X = walkExpr(conv.X, init)
	} else {
		n.X = walkExpr(n.X, init)
	}

	low, high, max := n.SliceBounds()
	low = walkExpr(low, init)
	if low != nil && ir.IsZero(low) {
		// Reduce x[0:j] to x[:j] and x[0:j:k] to x[:j:k].
		low = nil
	}
	high = walkExpr(high, init)
	max = walkExpr(max, init)
	n.SetSliceBounds(low, high, max)
	if checkSlice {
		n.X = walkCheckPtrAlignment(n.X.(*ir.ConvExpr), init, max)
	}

	if n.Op().IsSlice3() {
		if max != nil && max.Op() == ir.OCAP && ir.SameSafeExpr(n.X, max.(*ir.UnaryExpr).X) {
			// Reduce x[i:j:cap(x)] to x[i:j].
			if n.Op() == ir.OSLICE3 {
				n.SetOp(ir.OSLICE)
			} else {
				n.SetOp(ir.OSLICEARR)
			}
			return reduceSlice(n)
		}
		return n
	}
	return reduceSlice(n)
}

// walkSliceHeader walks an OSLICEHEADER node.
func walkSliceHeader(n *ir.SliceHeaderExpr, init *ir.Nodes) ir.Node {
	n.Ptr = walkExpr(n.Ptr, init)
	n.LenCap[0] = walkExpr(n.LenCap[0], init)
	n.LenCap[1] = walkExpr(n.LenCap[1], init)
	return n
}

// walkStringToBytes walks an OSTR2BYTES node.
func walkStringToBytes(n *ir.ConvExpr, init *ir.Nodes) ir.Node {
	s := n.X
	if ir.IsConst(s, constant.String) {
		sc := ir.StringVal(s)

		// Allocate a [n]byte of the right size.
		t := types.NewArray(types.Types[types.TUINT8], int64(len(sc)))
		var a ir.Node
		if n.Esc() == ir.EscNone && len(sc) <= int(ir.MaxImplicitStackVarSize) {
			a = typecheck.NodAddr(typecheck.Temp(t))
		} else {
			a = callnew(t)
		}
		p := typecheck.Temp(t.PtrTo()) // *[n]byte
		init.Append(typecheck.Stmt(ir.NewAssignStmt(base.Pos, p, a)))

		// Copy from the static string data to the [n]byte.
		if len(sc) > 0 {
			as := ir.NewAssignStmt(base.Pos, ir.NewStarExpr(base.Pos, p), ir.NewStarExpr(base.Pos, typecheck.ConvNop(ir.NewUnaryExpr(base.Pos, ir.OSPTR, s), t.PtrTo())))
			appendWalkStmt(init, as)
		}

		// Slice the [n]byte to a []byte.
		slice := ir.NewSliceExpr(n.Pos(), ir.OSLICEARR, p)
		slice.SetType(n.Type())
		slice.SetTypecheck(1)
		return walkExpr(slice, init)
	}

	a := typecheck.NodNil()
	if n.Esc() == ir.EscNone {
		// Create temporary buffer for slice on stack.
		t := types.NewArray(types.Types[types.TUINT8], tmpstringbufsize)
		a = typecheck.NodAddr(typecheck.Temp(t))
	}
	// stringtoslicebyte(*32[byte], string) []byte
	return mkcall("stringtoslicebyte", n.Type(), init, a, typecheck.Conv(s, types.Types[types.TSTRING]))
}

// walkStringToBytesTemp walks an OSTR2BYTESTMP node.
func walkStringToBytesTemp(n *ir.ConvExpr, init *ir.Nodes) ir.Node {
	// []byte(string) conversion that creates a slice
	// referring to the actual string bytes.
	// This conversion is handled later by the backend and
	// is only for use by internal compiler optimizations
	// that know that the slice won't be mutated.
	// The only such case today is:
	// for i, c := range []byte(string)
	n.X = walkExpr(n.X, init)
	return n
}

// walkStringToRunes walks an OSTR2RUNES node.
func walkStringToRunes(n *ir.ConvExpr, init *ir.Nodes) ir.Node {
	a := typecheck.NodNil()
	if n.Esc() == ir.EscNone {
		// Create temporary buffer for slice on stack.
		t := types.NewArray(types.Types[types.TINT32], tmpstringbufsize)
		a = typecheck.NodAddr(typecheck.Temp(t))
	}
	// stringtoslicerune(*[32]rune, string) []rune
	return mkcall("stringtoslicerune", n.Type(), init, a, typecheck.Conv(n.X, types.Types[types.TSTRING]))
}

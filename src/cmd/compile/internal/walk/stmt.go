// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package walk

import (
	"go/constant"
	"go/token"
	"sort"
	"unicode/utf8"

	"cmd/compile/internal/base"
	"cmd/compile/internal/ir"
	"cmd/compile/internal/reflectdata"
	"cmd/compile/internal/ssagen"
	"cmd/compile/internal/typecheck"
	"cmd/compile/internal/types"
	"cmd/internal/src"
	"cmd/internal/sys"
)

func cheapComputableIndex(width int64) bool {
	switch ssagen.Arch.LinkArch.Family {
	// MIPS does not have R+R addressing
	// Arm64 may lack ability to generate this code in our assembler,
	// but the architecture supports it.
	case sys.PPC64, sys.S390X:
		return width == 1
	case sys.AMD64, sys.I386, sys.ARM64, sys.ARM:
		switch width {
		case 1, 2, 4, 8:
			return true
		}
	}
	return false
}

// walkrange transforms various forms of ORANGE into
// simpler forms.  The result must be assigned back to n.
// Node n may also be modified in place, and may also be
// the returned node.
func walkrange(nrange *ir.RangeStmt) ir.Node {
	if isMapClear(nrange) {
		m := nrange.X
		lno := ir.SetPos(m)
		n := mapClear(m)
		base.Pos = lno
		return n
	}

	nfor := ir.NewForStmt(nrange.Pos(), nil, nil, nil, nil)
	nfor.SetInit(nrange.Init())
	nfor.Label = nrange.Label

	// variable name conventions:
	//	ohv1, hv1, hv2: hidden (old) val 1, 2
	//	ha, hit: hidden aggregate, iterator
	//	hn, hp: hidden len, pointer
	//	hb: hidden bool
	//	a, v1, v2: not hidden aggregate, val 1, 2

	t := nrange.Type()

	a := nrange.X
	lno := ir.SetPos(a)

	var v1, v2 ir.Node
	l := len(nrange.Vars)
	if l > 0 {
		v1 = nrange.Vars[0]
	}

	if l > 1 {
		v2 = nrange.Vars[1]
	}

	if ir.IsBlank(v2) {
		v2 = nil
	}

	if ir.IsBlank(v1) && v2 == nil {
		v1 = nil
	}

	if v1 == nil && v2 != nil {
		base.Fatalf("walkrange: v2 != nil while v1 == nil")
	}

	var ifGuard *ir.IfStmt

	var body []ir.Node
	var init []ir.Node
	switch t.Kind() {
	default:
		base.Fatalf("walkrange")

	case types.TARRAY, types.TSLICE:
		if nn := arrayClear(nrange, v1, v2, a); nn != nil {
			base.Pos = lno
			return nn
		}

		// order.stmt arranged for a copy of the array/slice variable if needed.
		ha := a

		hv1 := typecheck.Temp(types.Types[types.TINT])
		hn := typecheck.Temp(types.Types[types.TINT])

		init = append(init, ir.NewAssignStmt(base.Pos, hv1, nil))
		init = append(init, ir.NewAssignStmt(base.Pos, hn, ir.NewUnaryExpr(base.Pos, ir.OLEN, ha)))

		nfor.Cond = ir.NewBinaryExpr(base.Pos, ir.OLT, hv1, hn)
		nfor.Post = ir.NewAssignStmt(base.Pos, hv1, ir.NewBinaryExpr(base.Pos, ir.OADD, hv1, ir.NewInt(1)))

		// for range ha { body }
		if v1 == nil {
			break
		}

		// for v1 := range ha { body }
		if v2 == nil {
			body = []ir.Node{ir.NewAssignStmt(base.Pos, v1, hv1)}
			break
		}

		// for v1, v2 := range ha { body }
		if cheapComputableIndex(nrange.Type().Elem().Width) {
			// v1, v2 = hv1, ha[hv1]
			tmp := ir.NewIndexExpr(base.Pos, ha, hv1)
			tmp.SetBounded(true)
			// Use OAS2 to correctly handle assignments
			// of the form "v1, a[v1] := range".
			a := ir.NewAssignListStmt(base.Pos, ir.OAS2, nil, nil)
			a.Lhs = []ir.Node{v1, v2}
			a.Rhs = []ir.Node{hv1, tmp}
			body = []ir.Node{a}
			break
		}

		// TODO(austin): OFORUNTIL is a strange beast, but is
		// necessary for expressing the control flow we need
		// while also making "break" and "continue" work. It
		// would be nice to just lower ORANGE during SSA, but
		// racewalk needs to see many of the operations
		// involved in ORANGE's implementation. If racewalk
		// moves into SSA, consider moving ORANGE into SSA and
		// eliminating OFORUNTIL.

		// TODO(austin): OFORUNTIL inhibits bounds-check
		// elimination on the index variable (see #20711).
		// Enhance the prove pass to understand this.
		ifGuard = ir.NewIfStmt(base.Pos, nil, nil, nil)
		ifGuard.Cond = ir.NewBinaryExpr(base.Pos, ir.OLT, hv1, hn)
		nfor.SetOp(ir.OFORUNTIL)

		hp := typecheck.Temp(types.NewPtr(nrange.Type().Elem()))
		tmp := ir.NewIndexExpr(base.Pos, ha, ir.NewInt(0))
		tmp.SetBounded(true)
		init = append(init, ir.NewAssignStmt(base.Pos, hp, typecheck.NodAddr(tmp)))

		// Use OAS2 to correctly handle assignments
		// of the form "v1, a[v1] := range".
		a := ir.NewAssignListStmt(base.Pos, ir.OAS2, nil, nil)
		a.Lhs = []ir.Node{v1, v2}
		a.Rhs = []ir.Node{hv1, ir.NewStarExpr(base.Pos, hp)}
		body = append(body, a)

		// Advance pointer as part of the late increment.
		//
		// This runs *after* the condition check, so we know
		// advancing the pointer is safe and won't go past the
		// end of the allocation.
		as := ir.NewAssignStmt(base.Pos, hp, addptr(hp, t.Elem().Width))
		nfor.Late = []ir.Node{typecheck.Stmt(as)}

	case types.TMAP:
		// order.stmt allocated the iterator for us.
		// we only use a once, so no copy needed.
		ha := a

		hit := nrange.Prealloc
		th := hit.Type()
		keysym := th.Field(0).Sym  // depends on layout of iterator struct.  See reflect.go:hiter
		elemsym := th.Field(1).Sym // ditto

		fn := typecheck.LookupRuntime("mapiterinit")

		fn = typecheck.SubstArgTypes(fn, t.Key(), t.Elem(), th)
		init = append(init, mkcall1(fn, nil, nil, reflectdata.TypePtr(t), ha, typecheck.NodAddr(hit)))
		nfor.Cond = ir.NewBinaryExpr(base.Pos, ir.ONE, ir.NewSelectorExpr(base.Pos, ir.ODOT, hit, keysym), typecheck.NodNil())

		fn = typecheck.LookupRuntime("mapiternext")
		fn = typecheck.SubstArgTypes(fn, th)
		nfor.Post = mkcall1(fn, nil, nil, typecheck.NodAddr(hit))

		key := ir.NewStarExpr(base.Pos, ir.NewSelectorExpr(base.Pos, ir.ODOT, hit, keysym))
		if v1 == nil {
			body = nil
		} else if v2 == nil {
			body = []ir.Node{ir.NewAssignStmt(base.Pos, v1, key)}
		} else {
			elem := ir.NewStarExpr(base.Pos, ir.NewSelectorExpr(base.Pos, ir.ODOT, hit, elemsym))
			a := ir.NewAssignListStmt(base.Pos, ir.OAS2, nil, nil)
			a.Lhs = []ir.Node{v1, v2}
			a.Rhs = []ir.Node{key, elem}
			body = []ir.Node{a}
		}

	case types.TCHAN:
		// order.stmt arranged for a copy of the channel variable.
		ha := a

		hv1 := typecheck.Temp(t.Elem())
		hv1.SetTypecheck(1)
		if t.Elem().HasPointers() {
			init = append(init, ir.NewAssignStmt(base.Pos, hv1, nil))
		}
		hb := typecheck.Temp(types.Types[types.TBOOL])

		nfor.Cond = ir.NewBinaryExpr(base.Pos, ir.ONE, hb, ir.NewBool(false))
		a := ir.NewAssignListStmt(base.Pos, ir.OAS2RECV, nil, nil)
		a.SetTypecheck(1)
		a.Lhs = []ir.Node{hv1, hb}
		a.Rhs = []ir.Node{ir.NewUnaryExpr(base.Pos, ir.ORECV, ha)}
		*nfor.Cond.PtrInit() = []ir.Node{a}
		if v1 == nil {
			body = nil
		} else {
			body = []ir.Node{ir.NewAssignStmt(base.Pos, v1, hv1)}
		}
		// Zero hv1. This prevents hv1 from being the sole, inaccessible
		// reference to an otherwise GC-able value during the next channel receive.
		// See issue 15281.
		body = append(body, ir.NewAssignStmt(base.Pos, hv1, nil))

	case types.TSTRING:
		// Transform string range statements like "for v1, v2 = range a" into
		//
		// ha := a
		// for hv1 := 0; hv1 < len(ha); {
		//   hv1t := hv1
		//   hv2 := rune(ha[hv1])
		//   if hv2 < utf8.RuneSelf {
		//      hv1++
		//   } else {
		//      hv2, hv1 = decoderune(ha, hv1)
		//   }
		//   v1, v2 = hv1t, hv2
		//   // original body
		// }

		// order.stmt arranged for a copy of the string variable.
		ha := a

		hv1 := typecheck.Temp(types.Types[types.TINT])
		hv1t := typecheck.Temp(types.Types[types.TINT])
		hv2 := typecheck.Temp(types.RuneType)

		// hv1 := 0
		init = append(init, ir.NewAssignStmt(base.Pos, hv1, nil))

		// hv1 < len(ha)
		nfor.Cond = ir.NewBinaryExpr(base.Pos, ir.OLT, hv1, ir.NewUnaryExpr(base.Pos, ir.OLEN, ha))

		if v1 != nil {
			// hv1t = hv1
			body = append(body, ir.NewAssignStmt(base.Pos, hv1t, hv1))
		}

		// hv2 := rune(ha[hv1])
		nind := ir.NewIndexExpr(base.Pos, ha, hv1)
		nind.SetBounded(true)
		body = append(body, ir.NewAssignStmt(base.Pos, hv2, typecheck.Conv(nind, types.RuneType)))

		// if hv2 < utf8.RuneSelf
		nif := ir.NewIfStmt(base.Pos, nil, nil, nil)
		nif.Cond = ir.NewBinaryExpr(base.Pos, ir.OLT, hv2, ir.NewInt(utf8.RuneSelf))

		// hv1++
		nif.Body = []ir.Node{ir.NewAssignStmt(base.Pos, hv1, ir.NewBinaryExpr(base.Pos, ir.OADD, hv1, ir.NewInt(1)))}

		// } else {
		eif := ir.NewAssignListStmt(base.Pos, ir.OAS2, nil, nil)
		nif.Else = []ir.Node{eif}

		// hv2, hv1 = decoderune(ha, hv1)
		eif.Lhs = []ir.Node{hv2, hv1}
		fn := typecheck.LookupRuntime("decoderune")
		eif.Rhs = []ir.Node{mkcall1(fn, fn.Type().Results(), nil, ha, hv1)}

		body = append(body, nif)

		if v1 != nil {
			if v2 != nil {
				// v1, v2 = hv1t, hv2
				a := ir.NewAssignListStmt(base.Pos, ir.OAS2, nil, nil)
				a.Lhs = []ir.Node{v1, v2}
				a.Rhs = []ir.Node{hv1t, hv2}
				body = append(body, a)
			} else {
				// v1 = hv1t
				body = append(body, ir.NewAssignStmt(base.Pos, v1, hv1t))
			}
		}
	}

	typecheck.Stmts(init)

	if ifGuard != nil {
		ifGuard.PtrInit().Append(init...)
		ifGuard = typecheck.Stmt(ifGuard).(*ir.IfStmt)
	} else {
		nfor.PtrInit().Append(init...)
	}

	typecheck.Stmts(nfor.Cond.Init())

	nfor.Cond = typecheck.Expr(nfor.Cond)
	nfor.Cond = typecheck.DefaultLit(nfor.Cond, nil)
	nfor.Post = typecheck.Stmt(nfor.Post)
	typecheck.Stmts(body)
	nfor.Body.Append(body...)
	nfor.Body.Append(nrange.Body...)

	var n ir.Node = nfor
	if ifGuard != nil {
		ifGuard.Body = []ir.Node{n}
		n = ifGuard
	}

	n = walkstmt(n)

	base.Pos = lno
	return n
}

// isMapClear checks if n is of the form:
//
// for k := range m {
//   delete(m, k)
// }
//
// where == for keys of map m is reflexive.
func isMapClear(n *ir.RangeStmt) bool {
	if base.Flag.N != 0 || base.Flag.Cfg.Instrumenting {
		return false
	}

	if n.Op() != ir.ORANGE || n.Type().Kind() != types.TMAP || len(n.Vars) != 1 {
		return false
	}

	k := n.Vars[0]
	if k == nil || ir.IsBlank(k) {
		return false
	}

	// Require k to be a new variable name.
	if !ir.DeclaredBy(k, n) {
		return false
	}

	if len(n.Body) != 1 {
		return false
	}

	stmt := n.Body[0] // only stmt in body
	if stmt == nil || stmt.Op() != ir.ODELETE {
		return false
	}

	m := n.X
	if delete := stmt.(*ir.CallExpr); !ir.SameSafeExpr(delete.Args[0], m) || !ir.SameSafeExpr(delete.Args[1], k) {
		return false
	}

	// Keys where equality is not reflexive can not be deleted from maps.
	if !types.IsReflexive(m.Type().Key()) {
		return false
	}

	return true
}

// mapClear constructs a call to runtime.mapclear for the map m.
func mapClear(m ir.Node) ir.Node {
	t := m.Type()

	// instantiate mapclear(typ *type, hmap map[any]any)
	fn := typecheck.LookupRuntime("mapclear")
	fn = typecheck.SubstArgTypes(fn, t.Key(), t.Elem())
	n := mkcall1(fn, nil, nil, reflectdata.TypePtr(t), m)
	return walkstmt(typecheck.Stmt(n))
}

// Lower n into runtime·memclr if possible, for
// fast zeroing of slices and arrays (issue 5373).
// Look for instances of
//
// for i := range a {
// 	a[i] = zero
// }
//
// in which the evaluation of a is side-effect-free.
//
// Parameters are as in walkrange: "for v1, v2 = range a".
func arrayClear(loop *ir.RangeStmt, v1, v2, a ir.Node) ir.Node {
	if base.Flag.N != 0 || base.Flag.Cfg.Instrumenting {
		return nil
	}

	if v1 == nil || v2 != nil {
		return nil
	}

	if len(loop.Body) != 1 || loop.Body[0] == nil {
		return nil
	}

	stmt1 := loop.Body[0] // only stmt in body
	if stmt1.Op() != ir.OAS {
		return nil
	}
	stmt := stmt1.(*ir.AssignStmt)
	if stmt.X.Op() != ir.OINDEX {
		return nil
	}
	lhs := stmt.X.(*ir.IndexExpr)

	if !ir.SameSafeExpr(lhs.X, a) || !ir.SameSafeExpr(lhs.Index, v1) {
		return nil
	}

	elemsize := loop.Type().Elem().Width
	if elemsize <= 0 || !ir.IsZero(stmt.Y) {
		return nil
	}

	// Convert to
	// if len(a) != 0 {
	// 	hp = &a[0]
	// 	hn = len(a)*sizeof(elem(a))
	// 	memclr{NoHeap,Has}Pointers(hp, hn)
	// 	i = len(a) - 1
	// }
	n := ir.NewIfStmt(base.Pos, nil, nil, nil)
	n.Body.Set(nil)
	n.Cond = ir.NewBinaryExpr(base.Pos, ir.ONE, ir.NewUnaryExpr(base.Pos, ir.OLEN, a), ir.NewInt(0))

	// hp = &a[0]
	hp := typecheck.Temp(types.Types[types.TUNSAFEPTR])

	ix := ir.NewIndexExpr(base.Pos, a, ir.NewInt(0))
	ix.SetBounded(true)
	addr := typecheck.ConvNop(typecheck.NodAddr(ix), types.Types[types.TUNSAFEPTR])
	n.Body.Append(ir.NewAssignStmt(base.Pos, hp, addr))

	// hn = len(a) * sizeof(elem(a))
	hn := typecheck.Temp(types.Types[types.TUINTPTR])
	mul := typecheck.Conv(ir.NewBinaryExpr(base.Pos, ir.OMUL, ir.NewUnaryExpr(base.Pos, ir.OLEN, a), ir.NewInt(elemsize)), types.Types[types.TUINTPTR])
	n.Body.Append(ir.NewAssignStmt(base.Pos, hn, mul))

	var fn ir.Node
	if a.Type().Elem().HasPointers() {
		// memclrHasPointers(hp, hn)
		ir.CurFunc.SetWBPos(stmt.Pos())
		fn = mkcall("memclrHasPointers", nil, nil, hp, hn)
	} else {
		// memclrNoHeapPointers(hp, hn)
		fn = mkcall("memclrNoHeapPointers", nil, nil, hp, hn)
	}

	n.Body.Append(fn)

	// i = len(a) - 1
	v1 = ir.NewAssignStmt(base.Pos, v1, ir.NewBinaryExpr(base.Pos, ir.OSUB, ir.NewUnaryExpr(base.Pos, ir.OLEN, a), ir.NewInt(1)))

	n.Body.Append(v1)

	n.Cond = typecheck.Expr(n.Cond)
	n.Cond = typecheck.DefaultLit(n.Cond, nil)
	typecheck.Stmts(n.Body)
	return walkstmt(n)
}

// addptr returns (*T)(uintptr(p) + n).
func addptr(p ir.Node, n int64) ir.Node {
	t := p.Type()

	p = ir.NewConvExpr(base.Pos, ir.OCONVNOP, nil, p)
	p.SetType(types.Types[types.TUINTPTR])

	p = ir.NewBinaryExpr(base.Pos, ir.OADD, p, ir.NewInt(n))

	p = ir.NewConvExpr(base.Pos, ir.OCONVNOP, nil, p)
	p.SetType(t)

	return p
}

func walkselect(sel *ir.SelectStmt) {
	lno := ir.SetPos(sel)
	if len(sel.Compiled) != 0 {
		base.Fatalf("double walkselect")
	}

	init := sel.Init()
	sel.PtrInit().Set(nil)

	init = append(init, walkselectcases(sel.Cases)...)
	sel.Cases = ir.Nodes{}

	sel.Compiled.Set(init)
	walkstmtlist(sel.Compiled)

	base.Pos = lno
}

func walkselectcases(cases ir.Nodes) []ir.Node {
	ncas := len(cases)
	sellineno := base.Pos

	// optimization: zero-case select
	if ncas == 0 {
		return []ir.Node{mkcall("block", nil, nil)}
	}

	// optimization: one-case select: single op.
	if ncas == 1 {
		cas := cases[0].(*ir.CaseStmt)
		ir.SetPos(cas)
		l := cas.Init()
		if cas.Comm != nil { // not default:
			n := cas.Comm
			l = append(l, n.Init()...)
			n.PtrInit().Set(nil)
			switch n.Op() {
			default:
				base.Fatalf("select %v", n.Op())

			case ir.OSEND:
				// already ok

			case ir.OSELRECV2:
				r := n.(*ir.AssignListStmt)
				if ir.IsBlank(r.Lhs[0]) && ir.IsBlank(r.Lhs[1]) {
					n = r.Rhs[0]
					break
				}
				r.SetOp(ir.OAS2RECV)
			}

			l = append(l, n)
		}

		l = append(l, cas.Body...)
		l = append(l, ir.NewBranchStmt(base.Pos, ir.OBREAK, nil))
		return l
	}

	// convert case value arguments to addresses.
	// this rewrite is used by both the general code and the next optimization.
	var dflt *ir.CaseStmt
	for _, cas := range cases {
		cas := cas.(*ir.CaseStmt)
		ir.SetPos(cas)
		n := cas.Comm
		if n == nil {
			dflt = cas
			continue
		}
		switch n.Op() {
		case ir.OSEND:
			n := n.(*ir.SendStmt)
			n.Value = typecheck.NodAddr(n.Value)
			n.Value = typecheck.Expr(n.Value)

		case ir.OSELRECV2:
			n := n.(*ir.AssignListStmt)
			if !ir.IsBlank(n.Lhs[0]) {
				n.Lhs[0] = typecheck.NodAddr(n.Lhs[0])
				n.Lhs[0] = typecheck.Expr(n.Lhs[0])
			}
		}
	}

	// optimization: two-case select but one is default: single non-blocking op.
	if ncas == 2 && dflt != nil {
		cas := cases[0].(*ir.CaseStmt)
		if cas == dflt {
			cas = cases[1].(*ir.CaseStmt)
		}

		n := cas.Comm
		ir.SetPos(n)
		r := ir.NewIfStmt(base.Pos, nil, nil, nil)
		r.PtrInit().Set(cas.Init())
		var call ir.Node
		switch n.Op() {
		default:
			base.Fatalf("select %v", n.Op())

		case ir.OSEND:
			// if selectnbsend(c, v) { body } else { default body }
			n := n.(*ir.SendStmt)
			ch := n.Chan
			call = mkcall1(chanfn("selectnbsend", 2, ch.Type()), types.Types[types.TBOOL], r.PtrInit(), ch, n.Value)

		case ir.OSELRECV2:
			n := n.(*ir.AssignListStmt)
			recv := n.Rhs[0].(*ir.UnaryExpr)
			ch := recv.X
			elem := n.Lhs[0]
			if ir.IsBlank(elem) {
				elem = typecheck.NodNil()
			}
			if ir.IsBlank(n.Lhs[1]) {
				// if selectnbrecv(&v, c) { body } else { default body }
				call = mkcall1(chanfn("selectnbrecv", 2, ch.Type()), types.Types[types.TBOOL], r.PtrInit(), elem, ch)
			} else {
				// TODO(cuonglm): make this use selectnbrecv()
				// if selectnbrecv2(&v, &received, c) { body } else { default body }
				receivedp := typecheck.Expr(typecheck.NodAddr(n.Lhs[1]))
				call = mkcall1(chanfn("selectnbrecv2", 2, ch.Type()), types.Types[types.TBOOL], r.PtrInit(), elem, receivedp, ch)
			}
		}

		r.Cond = typecheck.Expr(call)
		r.Body.Set(cas.Body)
		r.Else.Set(append(dflt.Init(), dflt.Body...))
		return []ir.Node{r, ir.NewBranchStmt(base.Pos, ir.OBREAK, nil)}
	}

	if dflt != nil {
		ncas--
	}
	casorder := make([]*ir.CaseStmt, ncas)
	nsends, nrecvs := 0, 0

	var init []ir.Node

	// generate sel-struct
	base.Pos = sellineno
	selv := typecheck.Temp(types.NewArray(scasetype(), int64(ncas)))
	init = append(init, typecheck.Stmt(ir.NewAssignStmt(base.Pos, selv, nil)))

	// No initialization for order; runtime.selectgo is responsible for that.
	order := typecheck.Temp(types.NewArray(types.Types[types.TUINT16], 2*int64(ncas)))

	var pc0, pcs ir.Node
	if base.Flag.Race {
		pcs = typecheck.Temp(types.NewArray(types.Types[types.TUINTPTR], int64(ncas)))
		pc0 = typecheck.Expr(typecheck.NodAddr(ir.NewIndexExpr(base.Pos, pcs, ir.NewInt(0))))
	} else {
		pc0 = typecheck.NodNil()
	}

	// register cases
	for _, cas := range cases {
		cas := cas.(*ir.CaseStmt)
		ir.SetPos(cas)

		init = append(init, cas.Init()...)
		cas.PtrInit().Set(nil)

		n := cas.Comm
		if n == nil { // default:
			continue
		}

		var i int
		var c, elem ir.Node
		switch n.Op() {
		default:
			base.Fatalf("select %v", n.Op())
		case ir.OSEND:
			n := n.(*ir.SendStmt)
			i = nsends
			nsends++
			c = n.Chan
			elem = n.Value
		case ir.OSELRECV2:
			n := n.(*ir.AssignListStmt)
			nrecvs++
			i = ncas - nrecvs
			recv := n.Rhs[0].(*ir.UnaryExpr)
			c = recv.X
			elem = n.Lhs[0]
		}

		casorder[i] = cas

		setField := func(f string, val ir.Node) {
			r := ir.NewAssignStmt(base.Pos, ir.NewSelectorExpr(base.Pos, ir.ODOT, ir.NewIndexExpr(base.Pos, selv, ir.NewInt(int64(i))), typecheck.Lookup(f)), val)
			init = append(init, typecheck.Stmt(r))
		}

		c = typecheck.ConvNop(c, types.Types[types.TUNSAFEPTR])
		setField("c", c)
		if !ir.IsBlank(elem) {
			elem = typecheck.ConvNop(elem, types.Types[types.TUNSAFEPTR])
			setField("elem", elem)
		}

		// TODO(mdempsky): There should be a cleaner way to
		// handle this.
		if base.Flag.Race {
			r := mkcall("selectsetpc", nil, nil, typecheck.NodAddr(ir.NewIndexExpr(base.Pos, pcs, ir.NewInt(int64(i)))))
			init = append(init, r)
		}
	}
	if nsends+nrecvs != ncas {
		base.Fatalf("walkselectcases: miscount: %v + %v != %v", nsends, nrecvs, ncas)
	}

	// run the select
	base.Pos = sellineno
	chosen := typecheck.Temp(types.Types[types.TINT])
	recvOK := typecheck.Temp(types.Types[types.TBOOL])
	r := ir.NewAssignListStmt(base.Pos, ir.OAS2, nil, nil)
	r.Lhs = []ir.Node{chosen, recvOK}
	fn := typecheck.LookupRuntime("selectgo")
	r.Rhs = []ir.Node{mkcall1(fn, fn.Type().Results(), nil, bytePtrToIndex(selv, 0), bytePtrToIndex(order, 0), pc0, ir.NewInt(int64(nsends)), ir.NewInt(int64(nrecvs)), ir.NewBool(dflt == nil))}
	init = append(init, typecheck.Stmt(r))

	// selv and order are no longer alive after selectgo.
	init = append(init, ir.NewUnaryExpr(base.Pos, ir.OVARKILL, selv))
	init = append(init, ir.NewUnaryExpr(base.Pos, ir.OVARKILL, order))
	if base.Flag.Race {
		init = append(init, ir.NewUnaryExpr(base.Pos, ir.OVARKILL, pcs))
	}

	// dispatch cases
	dispatch := func(cond ir.Node, cas *ir.CaseStmt) {
		cond = typecheck.Expr(cond)
		cond = typecheck.DefaultLit(cond, nil)

		r := ir.NewIfStmt(base.Pos, cond, nil, nil)

		if n := cas.Comm; n != nil && n.Op() == ir.OSELRECV2 {
			n := n.(*ir.AssignListStmt)
			if !ir.IsBlank(n.Lhs[1]) {
				x := ir.NewAssignStmt(base.Pos, n.Lhs[1], recvOK)
				r.Body.Append(typecheck.Stmt(x))
			}
		}

		r.Body.Append(cas.Body.Take()...)
		r.Body.Append(ir.NewBranchStmt(base.Pos, ir.OBREAK, nil))
		init = append(init, r)
	}

	if dflt != nil {
		ir.SetPos(dflt)
		dispatch(ir.NewBinaryExpr(base.Pos, ir.OLT, chosen, ir.NewInt(0)), dflt)
	}
	for i, cas := range casorder {
		ir.SetPos(cas)
		dispatch(ir.NewBinaryExpr(base.Pos, ir.OEQ, chosen, ir.NewInt(int64(i))), cas)
	}

	return init
}

// bytePtrToIndex returns a Node representing "(*byte)(&n[i])".
func bytePtrToIndex(n ir.Node, i int64) ir.Node {
	s := typecheck.NodAddr(ir.NewIndexExpr(base.Pos, n, ir.NewInt(i)))
	t := types.NewPtr(types.Types[types.TUINT8])
	return typecheck.ConvNop(s, t)
}

var scase *types.Type

// Keep in sync with src/runtime/select.go.
func scasetype() *types.Type {
	if scase == nil {
		scase = typecheck.NewStructType([]*ir.Field{
			ir.NewField(base.Pos, typecheck.Lookup("c"), nil, types.Types[types.TUNSAFEPTR]),
			ir.NewField(base.Pos, typecheck.Lookup("elem"), nil, types.Types[types.TUNSAFEPTR]),
		})
		scase.SetNoalg(true)
	}
	return scase
}

// walkswitch walks a switch statement.
func walkswitch(sw *ir.SwitchStmt) {
	// Guard against double walk, see #25776.
	if len(sw.Cases) == 0 && len(sw.Compiled) > 0 {
		return // Was fatal, but eliminating every possible source of double-walking is hard
	}

	if sw.Tag != nil && sw.Tag.Op() == ir.OTYPESW {
		walkTypeSwitch(sw)
	} else {
		walkExprSwitch(sw)
	}
}

// walkExprSwitch generates an AST implementing sw.  sw is an
// expression switch.
func walkExprSwitch(sw *ir.SwitchStmt) {
	lno := ir.SetPos(sw)

	cond := sw.Tag
	sw.Tag = nil

	// convert switch {...} to switch true {...}
	if cond == nil {
		cond = ir.NewBool(true)
		cond = typecheck.Expr(cond)
		cond = typecheck.DefaultLit(cond, nil)
	}

	// Given "switch string(byteslice)",
	// with all cases being side-effect free,
	// use a zero-cost alias of the byte slice.
	// Do this before calling walkexpr on cond,
	// because walkexpr will lower the string
	// conversion into a runtime call.
	// See issue 24937 for more discussion.
	if cond.Op() == ir.OBYTES2STR && allCaseExprsAreSideEffectFree(sw) {
		cond := cond.(*ir.ConvExpr)
		cond.SetOp(ir.OBYTES2STRTMP)
	}

	cond = walkexpr(cond, sw.PtrInit())
	if cond.Op() != ir.OLITERAL && cond.Op() != ir.ONIL {
		cond = copyexpr(cond, cond.Type(), &sw.Compiled)
	}

	base.Pos = lno

	s := exprSwitch{
		exprname: cond,
	}

	var defaultGoto ir.Node
	var body ir.Nodes
	for _, ncase := range sw.Cases {
		ncase := ncase.(*ir.CaseStmt)
		label := typecheck.AutoLabel(".s")
		jmp := ir.NewBranchStmt(ncase.Pos(), ir.OGOTO, label)

		// Process case dispatch.
		if len(ncase.List) == 0 {
			if defaultGoto != nil {
				base.Fatalf("duplicate default case not detected during typechecking")
			}
			defaultGoto = jmp
		}

		for _, n1 := range ncase.List {
			s.Add(ncase.Pos(), n1, jmp)
		}

		// Process body.
		body.Append(ir.NewLabelStmt(ncase.Pos(), label))
		body.Append(ncase.Body...)
		if fall, pos := endsInFallthrough(ncase.Body); !fall {
			br := ir.NewBranchStmt(base.Pos, ir.OBREAK, nil)
			br.SetPos(pos)
			body.Append(br)
		}
	}
	sw.Cases.Set(nil)

	if defaultGoto == nil {
		br := ir.NewBranchStmt(base.Pos, ir.OBREAK, nil)
		br.SetPos(br.Pos().WithNotStmt())
		defaultGoto = br
	}

	s.Emit(&sw.Compiled)
	sw.Compiled.Append(defaultGoto)
	sw.Compiled.Append(body.Take()...)
	walkstmtlist(sw.Compiled)
}

// An exprSwitch walks an expression switch.
type exprSwitch struct {
	exprname ir.Node // value being switched on

	done    ir.Nodes
	clauses []exprClause
}

type exprClause struct {
	pos    src.XPos
	lo, hi ir.Node
	jmp    ir.Node
}

func (s *exprSwitch) Add(pos src.XPos, expr, jmp ir.Node) {
	c := exprClause{pos: pos, lo: expr, hi: expr, jmp: jmp}
	if types.IsOrdered[s.exprname.Type().Kind()] && expr.Op() == ir.OLITERAL {
		s.clauses = append(s.clauses, c)
		return
	}

	s.flush()
	s.clauses = append(s.clauses, c)
	s.flush()
}

func (s *exprSwitch) Emit(out *ir.Nodes) {
	s.flush()
	out.Append(s.done.Take()...)
}

func (s *exprSwitch) flush() {
	cc := s.clauses
	s.clauses = nil
	if len(cc) == 0 {
		return
	}

	// Caution: If len(cc) == 1, then cc[0] might not an OLITERAL.
	// The code below is structured to implicitly handle this case
	// (e.g., sort.Slice doesn't need to invoke the less function
	// when there's only a single slice element).

	if s.exprname.Type().IsString() && len(cc) >= 2 {
		// Sort strings by length and then by value. It is
		// much cheaper to compare lengths than values, and
		// all we need here is consistency. We respect this
		// sorting below.
		sort.Slice(cc, func(i, j int) bool {
			si := ir.StringVal(cc[i].lo)
			sj := ir.StringVal(cc[j].lo)
			if len(si) != len(sj) {
				return len(si) < len(sj)
			}
			return si < sj
		})

		// runLen returns the string length associated with a
		// particular run of exprClauses.
		runLen := func(run []exprClause) int64 { return int64(len(ir.StringVal(run[0].lo))) }

		// Collapse runs of consecutive strings with the same length.
		var runs [][]exprClause
		start := 0
		for i := 1; i < len(cc); i++ {
			if runLen(cc[start:]) != runLen(cc[i:]) {
				runs = append(runs, cc[start:i])
				start = i
			}
		}
		runs = append(runs, cc[start:])

		// Perform two-level binary search.
		binarySearch(len(runs), &s.done,
			func(i int) ir.Node {
				return ir.NewBinaryExpr(base.Pos, ir.OLE, ir.NewUnaryExpr(base.Pos, ir.OLEN, s.exprname), ir.NewInt(runLen(runs[i-1])))
			},
			func(i int, nif *ir.IfStmt) {
				run := runs[i]
				nif.Cond = ir.NewBinaryExpr(base.Pos, ir.OEQ, ir.NewUnaryExpr(base.Pos, ir.OLEN, s.exprname), ir.NewInt(runLen(run)))
				s.search(run, &nif.Body)
			},
		)
		return
	}

	sort.Slice(cc, func(i, j int) bool {
		return constant.Compare(cc[i].lo.Val(), token.LSS, cc[j].lo.Val())
	})

	// Merge consecutive integer cases.
	if s.exprname.Type().IsInteger() {
		merged := cc[:1]
		for _, c := range cc[1:] {
			last := &merged[len(merged)-1]
			if last.jmp == c.jmp && ir.Int64Val(last.hi)+1 == ir.Int64Val(c.lo) {
				last.hi = c.lo
			} else {
				merged = append(merged, c)
			}
		}
		cc = merged
	}

	s.search(cc, &s.done)
}

func (s *exprSwitch) search(cc []exprClause, out *ir.Nodes) {
	binarySearch(len(cc), out,
		func(i int) ir.Node {
			return ir.NewBinaryExpr(base.Pos, ir.OLE, s.exprname, cc[i-1].hi)
		},
		func(i int, nif *ir.IfStmt) {
			c := &cc[i]
			nif.Cond = c.test(s.exprname)
			nif.Body = []ir.Node{c.jmp}
		},
	)
}

func (c *exprClause) test(exprname ir.Node) ir.Node {
	// Integer range.
	if c.hi != c.lo {
		low := ir.NewBinaryExpr(c.pos, ir.OGE, exprname, c.lo)
		high := ir.NewBinaryExpr(c.pos, ir.OLE, exprname, c.hi)
		return ir.NewLogicalExpr(c.pos, ir.OANDAND, low, high)
	}

	// Optimize "switch true { ...}" and "switch false { ... }".
	if ir.IsConst(exprname, constant.Bool) && !c.lo.Type().IsInterface() {
		if ir.BoolVal(exprname) {
			return c.lo
		} else {
			return ir.NewUnaryExpr(c.pos, ir.ONOT, c.lo)
		}
	}

	return ir.NewBinaryExpr(c.pos, ir.OEQ, exprname, c.lo)
}

func allCaseExprsAreSideEffectFree(sw *ir.SwitchStmt) bool {
	// In theory, we could be more aggressive, allowing any
	// side-effect-free expressions in cases, but it's a bit
	// tricky because some of that information is unavailable due
	// to the introduction of temporaries during order.
	// Restricting to constants is simple and probably powerful
	// enough.

	for _, ncase := range sw.Cases {
		ncase := ncase.(*ir.CaseStmt)
		for _, v := range ncase.List {
			if v.Op() != ir.OLITERAL {
				return false
			}
		}
	}
	return true
}

// endsInFallthrough reports whether stmts ends with a "fallthrough" statement.
func endsInFallthrough(stmts []ir.Node) (bool, src.XPos) {
	// Search backwards for the index of the fallthrough
	// statement. Do not assume it'll be in the last
	// position, since in some cases (e.g. when the statement
	// list contains autotmp_ variables), one or more OVARKILL
	// nodes will be at the end of the list.

	i := len(stmts) - 1
	for i >= 0 && stmts[i].Op() == ir.OVARKILL {
		i--
	}
	if i < 0 {
		return false, src.NoXPos
	}
	return stmts[i].Op() == ir.OFALL, stmts[i].Pos()
}

// walkTypeSwitch generates an AST that implements sw, where sw is a
// type switch.
func walkTypeSwitch(sw *ir.SwitchStmt) {
	var s typeSwitch
	s.facename = sw.Tag.(*ir.TypeSwitchGuard).X
	sw.Tag = nil

	s.facename = walkexpr(s.facename, sw.PtrInit())
	s.facename = copyexpr(s.facename, s.facename.Type(), &sw.Compiled)
	s.okname = typecheck.Temp(types.Types[types.TBOOL])

	// Get interface descriptor word.
	// For empty interfaces this will be the type.
	// For non-empty interfaces this will be the itab.
	itab := ir.NewUnaryExpr(base.Pos, ir.OITAB, s.facename)

	// For empty interfaces, do:
	//     if e._type == nil {
	//         do nil case if it exists, otherwise default
	//     }
	//     h := e._type.hash
	// Use a similar strategy for non-empty interfaces.
	ifNil := ir.NewIfStmt(base.Pos, nil, nil, nil)
	ifNil.Cond = ir.NewBinaryExpr(base.Pos, ir.OEQ, itab, typecheck.NodNil())
	base.Pos = base.Pos.WithNotStmt() // disable statement marks after the first check.
	ifNil.Cond = typecheck.Expr(ifNil.Cond)
	ifNil.Cond = typecheck.DefaultLit(ifNil.Cond, nil)
	// ifNil.Nbody assigned at end.
	sw.Compiled.Append(ifNil)

	// Load hash from type or itab.
	dotHash := ir.NewSelectorExpr(base.Pos, ir.ODOTPTR, itab, nil)
	dotHash.SetType(types.Types[types.TUINT32])
	dotHash.SetTypecheck(1)
	if s.facename.Type().IsEmptyInterface() {
		dotHash.Offset = int64(2 * types.PtrSize) // offset of hash in runtime._type
	} else {
		dotHash.Offset = int64(2 * types.PtrSize) // offset of hash in runtime.itab
	}
	dotHash.SetBounded(true) // guaranteed not to fault
	s.hashname = copyexpr(dotHash, dotHash.Type(), &sw.Compiled)

	br := ir.NewBranchStmt(base.Pos, ir.OBREAK, nil)
	var defaultGoto, nilGoto ir.Node
	var body ir.Nodes
	for _, ncase := range sw.Cases {
		ncase := ncase.(*ir.CaseStmt)
		var caseVar ir.Node
		if len(ncase.Vars) != 0 {
			caseVar = ncase.Vars[0]
		}

		// For single-type cases with an interface type,
		// we initialize the case variable as part of the type assertion.
		// In other cases, we initialize it in the body.
		var singleType *types.Type
		if len(ncase.List) == 1 && ncase.List[0].Op() == ir.OTYPE {
			singleType = ncase.List[0].Type()
		}
		caseVarInitialized := false

		label := typecheck.AutoLabel(".s")
		jmp := ir.NewBranchStmt(ncase.Pos(), ir.OGOTO, label)

		if len(ncase.List) == 0 { // default:
			if defaultGoto != nil {
				base.Fatalf("duplicate default case not detected during typechecking")
			}
			defaultGoto = jmp
		}

		for _, n1 := range ncase.List {
			if ir.IsNil(n1) { // case nil:
				if nilGoto != nil {
					base.Fatalf("duplicate nil case not detected during typechecking")
				}
				nilGoto = jmp
				continue
			}

			if singleType != nil && singleType.IsInterface() {
				s.Add(ncase.Pos(), n1.Type(), caseVar, jmp)
				caseVarInitialized = true
			} else {
				s.Add(ncase.Pos(), n1.Type(), nil, jmp)
			}
		}

		body.Append(ir.NewLabelStmt(ncase.Pos(), label))
		if caseVar != nil && !caseVarInitialized {
			val := s.facename
			if singleType != nil {
				// We have a single concrete type. Extract the data.
				if singleType.IsInterface() {
					base.Fatalf("singleType interface should have been handled in Add")
				}
				val = ifaceData(ncase.Pos(), s.facename, singleType)
			}
			l := []ir.Node{
				ir.NewDecl(ncase.Pos(), ir.ODCL, caseVar),
				ir.NewAssignStmt(ncase.Pos(), caseVar, val),
			}
			typecheck.Stmts(l)
			body.Append(l...)
		}
		body.Append(ncase.Body...)
		body.Append(br)
	}
	sw.Cases.Set(nil)

	if defaultGoto == nil {
		defaultGoto = br
	}
	if nilGoto == nil {
		nilGoto = defaultGoto
	}
	ifNil.Body = []ir.Node{nilGoto}

	s.Emit(&sw.Compiled)
	sw.Compiled.Append(defaultGoto)
	sw.Compiled.Append(body.Take()...)

	walkstmtlist(sw.Compiled)
}

// A typeSwitch walks a type switch.
type typeSwitch struct {
	// Temporary variables (i.e., ONAMEs) used by type switch dispatch logic:
	facename ir.Node // value being type-switched on
	hashname ir.Node // type hash of the value being type-switched on
	okname   ir.Node // boolean used for comma-ok type assertions

	done    ir.Nodes
	clauses []typeClause
}

type typeClause struct {
	hash uint32
	body ir.Nodes
}

func (s *typeSwitch) Add(pos src.XPos, typ *types.Type, caseVar, jmp ir.Node) {
	var body ir.Nodes
	if caseVar != nil {
		l := []ir.Node{
			ir.NewDecl(pos, ir.ODCL, caseVar),
			ir.NewAssignStmt(pos, caseVar, nil),
		}
		typecheck.Stmts(l)
		body.Append(l...)
	} else {
		caseVar = ir.BlankNode
	}

	// cv, ok = iface.(type)
	as := ir.NewAssignListStmt(pos, ir.OAS2, nil, nil)
	as.Lhs = []ir.Node{caseVar, s.okname} // cv, ok =
	dot := ir.NewTypeAssertExpr(pos, s.facename, nil)
	dot.SetType(typ) // iface.(type)
	as.Rhs = []ir.Node{dot}
	appendWalkStmt(&body, as)

	// if ok { goto label }
	nif := ir.NewIfStmt(pos, nil, nil, nil)
	nif.Cond = s.okname
	nif.Body = []ir.Node{jmp}
	body.Append(nif)

	if !typ.IsInterface() {
		s.clauses = append(s.clauses, typeClause{
			hash: types.TypeHash(typ),
			body: body,
		})
		return
	}

	s.flush()
	s.done.Append(body.Take()...)
}

func (s *typeSwitch) Emit(out *ir.Nodes) {
	s.flush()
	out.Append(s.done.Take()...)
}

func (s *typeSwitch) flush() {
	cc := s.clauses
	s.clauses = nil
	if len(cc) == 0 {
		return
	}

	sort.Slice(cc, func(i, j int) bool { return cc[i].hash < cc[j].hash })

	// Combine adjacent cases with the same hash.
	merged := cc[:1]
	for _, c := range cc[1:] {
		last := &merged[len(merged)-1]
		if last.hash == c.hash {
			last.body.Append(c.body.Take()...)
		} else {
			merged = append(merged, c)
		}
	}
	cc = merged

	binarySearch(len(cc), &s.done,
		func(i int) ir.Node {
			return ir.NewBinaryExpr(base.Pos, ir.OLE, s.hashname, ir.NewInt(int64(cc[i-1].hash)))
		},
		func(i int, nif *ir.IfStmt) {
			// TODO(mdempsky): Omit hash equality check if
			// there's only one type.
			c := cc[i]
			nif.Cond = ir.NewBinaryExpr(base.Pos, ir.OEQ, s.hashname, ir.NewInt(int64(c.hash)))
			nif.Body.Append(c.body.Take()...)
		},
	)
}

// binarySearch constructs a binary search tree for handling n cases,
// and appends it to out. It's used for efficiently implementing
// switch statements.
//
// less(i) should return a boolean expression. If it evaluates true,
// then cases before i will be tested; otherwise, cases i and later.
//
// leaf(i, nif) should setup nif (an OIF node) to test case i. In
// particular, it should set nif.Left and nif.Nbody.
func binarySearch(n int, out *ir.Nodes, less func(i int) ir.Node, leaf func(i int, nif *ir.IfStmt)) {
	const binarySearchMin = 4 // minimum number of cases for binary search

	var do func(lo, hi int, out *ir.Nodes)
	do = func(lo, hi int, out *ir.Nodes) {
		n := hi - lo
		if n < binarySearchMin {
			for i := lo; i < hi; i++ {
				nif := ir.NewIfStmt(base.Pos, nil, nil, nil)
				leaf(i, nif)
				base.Pos = base.Pos.WithNotStmt()
				nif.Cond = typecheck.Expr(nif.Cond)
				nif.Cond = typecheck.DefaultLit(nif.Cond, nil)
				out.Append(nif)
				out = &nif.Else
			}
			return
		}

		half := lo + n/2
		nif := ir.NewIfStmt(base.Pos, nil, nil, nil)
		nif.Cond = less(half)
		base.Pos = base.Pos.WithNotStmt()
		nif.Cond = typecheck.Expr(nif.Cond)
		nif.Cond = typecheck.DefaultLit(nif.Cond, nil)
		do(lo, half, &nif.Body)
		do(half, hi, &nif.Else)
		out.Append(nif)
	}

	do(0, n, out)
}

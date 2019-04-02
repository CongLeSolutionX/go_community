// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gc

import (
	"cmd/compile/internal/types"
	"fmt"
)

// Escape analysis.
//
// Here we analyze functions to determine whether Go variables can be
// allocated on the stack. The two key invariants we have to respect
// are: (1) objects allocated on the heap cannot point to objects on
// the stack, and (2) a pointer to a stack object cannot outlive that
// object (e.g., either because the function returned or its space is
// reused in a loop).
//
// We implement this with a simple data-flow analysis algorithm. For
// every Go variable, we create a "location." We then lower all Go
// statements into edges representing an assignment between the two,
// possibly with an addressing operation or an arbitrary number of
// dereference operations. For example:
//
//     p = &q    // -1
//     p = q     //  0
//     p = *q    //  1
//     p = **q   //  2
//
// Note that "p = &&q" is invalid, so the dereference counts can never
// go below -1.
//
// Assignments can also be directly to the heap.
//
// All Go language constructs are lowered into this graph
// representation, generally without sensitivity to flow, path, or
// context; and without distinguishing elements within a compound
// variable. For example:
//
//     var x struct { f, g *int }
//     var u []*int
//
//     x.f = u[0]
//
// is modeled simply as
//
//     x = *u
//
// We then define dist(p, q) as the shortest path distance from p to q
// on this graph, except that intermediate distance is bounded at 0.
// For example:
//
//     p = **q    //  2
//     q = &r     // -1
//     r = *s     //  1
//
// We have dist(p, r) == 1, dist(p, s) == 2, dist(q, s) == 1 (not 0!).
//
// Intuitively, if dist(p, q) == 0, then any value stored in q may be
// stored in p; if dist(p, q) == 1, then any value *pointed to* by q
// may be stored in p; and so on.
//
// Finally, if there exists p,q,r such that dist(p, q) == 0 and q =
// &r, then r's address leaks to p. If p outlives r, then r must be
// heap allocated.

type Escape struct {
	allLocs []*EscLocation

	curfn     *Node
	loopDepth int

	heapLoc  EscLocation
	blankLoc EscLocation
}

// escapeFuncs performs escape analysis on a minimal batch of
// functions.
func escapeFuncs(fns []*Node, recursive bool) {
	for _, fn := range fns {
		if fn.Op != ODCLFUNC {
			Fatalf("unexpected node: %v", fn)
		}
	}

	var e Escape

	// Construct data-flow graph from syntax trees.
	for _, fn := range fns {
		e.initFunc(fn)
	}
	for _, fn := range fns {
		e.walkFunc(fn)
	}
	e.curfn = nil

	e.flood()
	e.finish()

	// for all top level functions, tag the typenodes corresponding to the param nodes
	for _, fn := range fns {
		esctag(fn)
	}
}

func (e *Escape) initFunc(fn *Node) {
	if fn.Op != ODCLFUNC || fn.Esc != EscFuncUnknown {
		Fatalf("unexpected node: %v", fn)
	}
	fn.Esc = EscFuncPlanned
	if Debug['m'] > 3 {
		Dump("escAnalyze", fn)
	}

	e.curfn = fn
	e.loopDepth = 1

	// Allocate locations for local variables.
	for _, dcl := range fn.Func.Dcl {
		if dcl.Op == ONAME {
			loc := e.newLoc(dcl, false)

			// TODO(mdempsky): This should probably be handled elsewhere.
			if dcl.Class() == PPARAM && fn.Nbody.Len() == 0 && !fn.Noescape() {
				loc.paramEsc = EscHeap
			}
		}
	}
}

func (e *Escape) walkFunc(fn *Node) {
	fn.Esc = EscFuncStarted

	// Identify labels that mark the head of an unstructured loop.
	inspectList(fn.Nbody, func(n *Node) bool {
		switch n.Op {
		case OLABEL:
			n.Sym.Label = asTypesNode(&nonlooping)

		case OGOTO:
			// If we visited the label before the goto,
			// then this is a looping label.
			if n.Sym.Label == asTypesNode(&nonlooping) {
				n.Sym.Label = asTypesNode(&looping)
			}
		}

		return true
	})

	e.curfn = fn
	e.loopDepth = 1
	e.stmts(fn.Nbody)
}

// Below we implement the methods for walking the AST and recording
// data flow edges. Note that because a sub-expression might have
// side-effects, it's important to always visit the entire AST.
//
// For example, write either:
//
//     if x {
//         e.discard(n.Left)
//     } else {
//         e.value(k, n.Left)
//     }
//
// or
//
//     if x {
//         k = e.discardHole()
//     }
//     e.value(k, n.Left)
//
// Do NOT write:
//
//    // BAD: possibly loses side-effects within n.Left
//    if !x {
//        e.value(k, n.Left)
//    }

// stmt evaluates a single Go statement.
func (e *Escape) stmt(n *Node) {
	if n == nil {
		return
	}

	lno := setlineno(n)
	defer func() {
		lineno = lno
	}()

	if Debug['m'] > 2 {
		fmt.Printf("%v:[%d] %v stmt: %v\n", linestr(lineno), e.loopDepth, funcSym(e.curfn), n)
	}

	e.stmts(n.Ninit)

	switch n.Op {
	default:
		Fatalf("unexpected stmt: %v", n)

	case ODCLCONST, ODCLTYPE, OEMPTY, OFALL, OINLMARK:
		// nop

	case OBREAK, OCONTINUE, OGOTO:
		// TODO(mdempsky): Handle dead code?

	case OBLOCK:
		e.stmts(n.List)

	case ODCL:
		// Record loop depth at declaration.
		if !n.Left.isBlank() {
			e.dcl(n.Left)
		}

	case OLABEL:
		switch asNode(n.Sym.Label) {
		case &nonlooping:
			if Debug['m'] > 2 {
				fmt.Printf("%v:%v non-looping label\n", linestr(lineno), n)
			}
		case &looping:
			if Debug['m'] > 2 {
				fmt.Printf("%v: %v looping label\n", linestr(lineno), n)
			}
			e.loopDepth++
		default:
			Fatalf("label missing tag")
		}
		n.Sym.Label = nil

	case OIF:
		e.discard(n.Left)
		e.stmts(n.Nbody)
		e.stmts(n.Rlist)

	case OFOR, OFORUNTIL:
		e.loopDepth++
		e.discard(n.Left)
		e.stmt(n.Right)
		e.stmts(n.Nbody)
		e.loopDepth--

	case ORANGE:
		// for List = range Right { Nbody }

		// Right is evaluated outside the loop.
		tv := e.newLoc(n, false)
		e.value(tv.asHole(), n.Right)

		e.loopDepth++
		ks := e.addrs(n.List)
		if len(ks) >= 2 {
			if n.Right.Type.IsArray() {
				e.flow(ks[1].note(n, "range"), tv)
			} else {
				e.flow(ks[1].deref(n, "range-deref"), tv)
			}
		}

		e.stmts(n.Nbody)
		e.loopDepth--

	case OSWITCH:
		var tv *EscLocation
		if n.Left != nil {
			if n.Left.Op == OTYPESW {
				k := e.discardHole()
				if n.Left.Left != nil {
					tv = e.newLoc(n.Left, false)
					k = tv.asHole()
				}
				e.value(k, n.Left.Right)
			} else {
				e.discard(n.Left)
			}
		}

		for _, cas := range n.List.Slice() { // cases
			if tv != nil {
				// type switch variables have no ODCL.
				cv := cas.Rlist.First()
				k := e.dcl(cv)
				if types.Haspointers(cv.Type) {
					e.flow(k.dottype(cv.Type, n, "switch case"), tv)
				}
			}

			e.discards(cas.List)
			e.stmts(cas.Nbody)
		}

	case OSELECT:
		for _, cas := range n.List.Slice() {
			e.stmt(cas.Left)
			e.stmts(cas.Nbody)
		}
	case OSELRECV:
		e.assign(n.Left, n.Right, "selrecv", n)
	case OSELRECV2:
		e.assign(n.Left, n.Right, "selrecv", n)
		e.assign(n.List.First(), nil, "selrecv", n)
	case ORECV:
		// TODO(mdempsky): Consider e.discard(n.Left).
		e.valueSkipInit(e.discardHole(), n) // already visited n.Ninit
	case OSEND:
		e.discard(n.Left)
		e.assignHeap(n.Right, "send", n)

	case OAS, OASOP:
		e.assign(n.Left, n.Right, "assign", n)

	case OAS2:
		for i, nl := range n.List.Slice() {
			e.assign(nl, n.Rlist.Index(i), "assign-pair", n)
		}

	case OAS2DOTTYPE: // v, ok = x.(type)
		e.assign(n.List.First(), n.Rlist.First(), "assign-pair-dot-type", n)
		e.assign(n.List.Second(), nil, "assign-pair-dot-type", n)
	case OAS2MAPR: // v, ok = m[k]
		e.assign(n.List.First(), n.Rlist.First(), "assign-pair-mapr", n)
		e.assign(n.List.Second(), nil, "assign-pair-mapr", n)
	case OAS2RECV: // v, ok = <-ch
		e.assign(n.List.First(), n.Rlist.First(), "assign-pair-receive", n)
		e.assign(n.List.Second(), nil, "assign-pair-receive", n)

	case OAS2FUNC:
		e.stmts(n.Rlist.First().Ninit)
		e.call(e.addrs(n.List), n.Rlist.First(), nil)
	case ORETURN:
		ks := e.resultHoles()
		for i, v := range n.List.Slice() {
			e.value(ks[i], v)
		}
	case OCALLFUNC, OCALLMETH, OCALLINTER, OCLOSE, OCOPY, ODELETE, OPANIC, OPRINT, OPRINTN, ORECOVER:
		e.call(nil, n, nil)
	case OGO, ODEFER:
		e.stmts(n.Left.Ninit)
		e.call(nil, n.Left, n)

	case ORETJMP:
		// TODO(mdempsky): What do? esc.go just ignores it.
	}
}

func (e *Escape) stmts(l Nodes) {
	// TODO(mdempsky): Preserve and restore e.loopDepth? See also #22438.
	for _, n := range l.Slice() {
		e.stmt(n)
	}
}

// value evaluates an expression n in the context of k.
func (e *Escape) value(k EscHole, n *Node) {
	if n == nil {
		return
	}
	e.stmts(n.Ninit)
	e.valueSkipInit(k, n)
}

func (e *Escape) valueSkipInit(k EscHole, n *Node) {
	if n == nil {
		return
	}

	lno := setlineno(n)
	defer func() {
		lineno = lno
	}()

	if k.derefs >= 0 && !types.Haspointers(n.Type) {
		k = e.discardHole()
	}

	switch n.Op {
	default:
		Fatalf("unexpected expr: %v", n)

	case OLITERAL, OGETG, OCLOSUREVAR, OTYPE:
		// nop

	case ONAME:
		if n.Class() == PFUNC || n.Class() == PEXTERN {
			return
		}
		e.flow(k, e.oldLoc(n))

	case OPLUS, ONEG, OBITNOT, ONOT:
		e.discard(n.Left)
	case OADD, OSUB, OOR, OXOR, OMUL, ODIV, OMOD, OLSH, ORSH, OAND, OANDNOT, OEQ, ONE, OLT, OLE, OGT, OGE, OANDAND, OOROR:
		e.discard(n.Left)
		e.discard(n.Right)

	case OADDR:
		e.value(k.addr(n, "address-of"), n.Left) // "address-of"
	case ODEREF:
		e.value(k.deref(n, "indirection"), n.Left) // "indirection"
	case ODOT, ODOTMETH, ODOTINTER:
		e.value(k.note(n, "dot"), n.Left)
	case ODOTPTR:
		e.value(k.deref(n, "dot of pointer"), n.Left) // "dot of pointer"
	case ODOTTYPE, ODOTTYPE2:
		e.value(k.dottype(n.Type, n, "dot"), n.Left)
	case OINDEX:
		if n.Left.Type.IsArray() {
			e.value(k.note(n, "fixed-array-index-of"), n.Left)
		} else {
			// TODO(mdempsky): Fix why reason text.
			e.value(k.deref(n, "dot of pointer"), n.Left)
		}
		e.discard(n.Right)
	case OINDEXMAP:
		e.discard(n.Left)
		e.discard(n.Right)
	case OSLICE, OSLICEARR, OSLICE3, OSLICE3ARR, OSLICESTR:
		e.value(k.note(n, "slice"), n.Left)
		low, high, max := n.SliceBounds()
		e.discard(low)
		e.discard(high)
		e.discard(max)

	case OCONV, OCONVNOP:
		if n.Type.Etype == TUNSAFEPTR && n.Left.Type.Etype == TUINTPTR {
			e.unsafeValue(k, n.Left)
		} else {
			e.value(k, n.Left)
		}
	case OCONVIFACE:
		if !n.Left.Type.IsInterface() && !isdirectiface(n.Left.Type) {
			k = e.spill(k, n)
		} else {
			// esc.go prints "escapes to heap" / "does not
			// escape" messages for OCONVIFACE even when
			// they don't allocate.  Match that behavior
			// because it's easy.
			// TODO(mdempsky): Remove and cleanup test expectations.
			_ = e.spill(k, n)
		}
		e.value(k.note(n, "interface-converted"), n.Left)

	case ORECV:
		e.discard(n.Left)

	case OCALLMETH, OCALLFUNC, OCALLINTER, OLEN, OCAP, OCOMPLEX, OREAL, OIMAG, OAPPEND, OCOPY:
		e.call([]EscHole{k}, n, nil)

	case ONEW:
		e.spill(k, n)

	case OMAKESLICE:
		e.spill(k, n)
		e.discard(n.Left)
		e.discard(n.Right)
	case OMAKECHAN:
		e.discard(n.Left)
	case OMAKEMAP:
		e.spill(k, n)
		e.discard(n.Left)

	case ORECOVER:
		// nop

	case OCALLPART:
		e.spill(k, n)

		// esc.go says "Contents make it to memory, lose
		// track."  I think we can just flow n.Left to our
		// spilled location though.
		// TODO(mdempsky): Try that.
		e.assignHeap(n.Left, "call part", n)

	case OPTRLIT:
		e.value(e.spill(k, n), n.Left)

	case OARRAYLIT:
		for _, elt := range n.List.Slice() {
			if elt.Op == OKEY {
				elt = elt.Right
			}
			e.value(k.note(n, "array literal element"), elt)
		}

	case OSLICELIT:
		k = e.spill(k, n)

		for _, elt := range n.List.Slice() {
			if elt.Op == OKEY {
				elt = elt.Right
			}
			e.value(k.note(n, "slice-literal-element"), elt)
		}

	case OSTRUCTLIT:
		for _, elt := range n.List.Slice() {
			e.value(k.note(n, "struct literal element"), elt.Left)
		}

	case OMAPLIT:
		e.spill(k, n)

		// Map keys and values are always stored in the heap.
		for _, elt := range n.List.Slice() {
			e.assignHeap(elt.Left, "map literal key", n)
			e.assignHeap(elt.Right, "map literal value", n)
		}

	case OCLOSURE:
		k = e.spill(k, n)

		// Link addresses of captured variables to closure.
		for _, v := range n.Func.Closure.Func.Cvars.Slice() {
			if v.Op == OXXX { // unnamed out argument; see dcl.go:/^funcargs
				continue
			}

			k := k
			if !v.Name.Byval() {
				k = k.addr(v, "reference")
			}

			e.value(k.note(n, "captured by a closure"), v.Name.Defn)
		}

	case ORUNES2STR, OBYTES2STR, OSTR2RUNES, OSTR2BYTES, ORUNESTR:
		e.spill(k, n)
		e.discard(n.Left)

	case OADDSTR:
		e.spill(k, n)

		// Arguments of OADDSTR never escape;
		// runtime.concatstrings makes sure of that.
		e.discards(n.List)
	}
}

// unsafeValue evaluates a uintptr-typed arithmetic expression looking
// for conversions from an unsafe.Pointer.
func (e *Escape) unsafeValue(k EscHole, n *Node) {
	if n.Type.Etype != TUINTPTR {
		Fatalf("unexpected type %v for %v", n.Type, n)
	}

	e.stmts(n.Ninit)

	switch n.Op {
	case OCONV, OCONVNOP:
		if n.Left.Type.Etype == TUNSAFEPTR {
			e.value(k, n.Left)
		} else {
			e.discard(n.Left)
		}
	case ODOTPTR:
		if isReflectHeaderDataField(n) {
			e.value(k.deref(n, "reflect.Header.Data"), n.Left)
		} else {
			e.discard(n.Left)
		}
	case OPLUS, ONEG, OBITNOT:
		e.unsafeValue(k, n.Left)
	case OADD, OSUB, OOR, OXOR, OMUL, ODIV, OMOD, OLSH, ORSH, OAND, OANDNOT:
		e.unsafeValue(k, n.Left)
		e.unsafeValue(k, n.Right)
	default:
		e.valueSkipInit(e.discardHole(), n)
	}
}

// discard evaluates an expression n for side-effects, but discards
// its value.
func (e *Escape) discard(n *Node) {
	e.value(e.discardHole(), n)
}

func (e *Escape) discards(l Nodes) {
	for _, n := range l.Slice() {
		e.discard(n)
	}
}

// addr evaluates an addressable expression n and returns an EscHole
// that represents storing into the represented location.
func (e *Escape) addr(n *Node) EscHole {
	if n == nil || n.isBlank() {
		// Can happen at least in OSELRECV.
		// TODO(mdempsky): Anywhere else?
		return e.discardHole()
	}

	k := e.heapHole()

	switch n.Op {
	default:
		Fatalf("unexpected addr: %v", n)
	case ONAME:
		if n.Class() == PEXTERN {
			break
		}
		k = e.oldLoc(n).asHole()
	case ODOT:
		k = e.addr(n.Left)
	case OINDEX:
		e.discard(n.Right)
		if n.Left.Type.IsArray() {
			k = e.addr(n.Left)
		} else {
			e.discard(n.Left)
		}
	case ODEREF, ODOTPTR:
		e.discard(n)
	case OINDEXMAP:
		e.discard(n.Left)
		e.assignHeap(n.Right, "key of map put", n)
	}

	if !types.Haspointers(n.Type) {
		k = e.discardHole()
	}

	return k
}

func (e *Escape) addrs(l Nodes) []EscHole {
	var ks []EscHole
	for _, n := range l.Slice() {
		ks = append(ks, e.addr(n))
	}
	return ks
}

func (e *Escape) assign(dst, src *Node, why string, where *Node) {
	// Filter out some no-op assignments for escape analysis.
	ignore := dst != nil && src != nil && isSelfAssign(dst, src)
	if ignore && Debug['m'] != 0 {
		Warnl(where.Pos, "%v ignoring self-assignment in %S", funcSym(e.curfn), where)
	}

	k := e.addr(dst)
	if dst != nil && dst.Op == ODOTPTR && isReflectHeaderDataField(dst) {
		e.unsafeValue(e.heapHole(), src)
	} else {
		if ignore {
			k = e.discardHole()
		}
		e.value(k, src)
	}
}

func (e *Escape) assignHeap(src *Node, why string, where *Node) {
	e.value(e.heapHole().note(where, why), src)
}

// call evaluates a call expressions, including builtin calls. ks
// should be the contexts for the function callee's results; where is
// the OGO/ODEFER context of the call, if any.
func (e *Escape) call(ks []EscHole, call, where *Node) {
	// First, pick out the function callee, its type, and receiver
	// (if any) and normal arguments list.
	var fn, recv *Node
	var fntype *types.Type
	args := call.List.Slice()
	switch call.Op {
	case OCALLFUNC:
		fn = call.Left
		if fn.Op == OCLOSURE {
			fn = fn.Func.Closure.Func.Nname
		}
		fntype = fn.Type
	case OCALLMETH:
		fn = asNode(call.Left.Type.FuncType().Nname)
		fntype = fn.Type
		recv = call.Left.Left
	case OCALLINTER:
		fntype = call.Left.Type
		recv = call.Left.Left
	case OAPPEND, ODELETE, OPRINT, OPRINTN, ORECOVER:
		// ok
	case OLEN, OCAP, OREAL, OIMAG, OCLOSE, OPANIC:
		args = []*Node{call.Left}
	case OCOMPLEX, OCOPY:
		args = []*Node{call.Left, call.Right}
	default:
		Fatalf("unexpected call op: %v", call.Op)
	}

	direct := fn != nil && fn.Op == ONAME && fn.Class() == PFUNC

	// Setup evaluation contexts for each receiver/argument.
	var recvK EscHole
	var paramKs []EscHole

	if where != nil && !(where.Op == ODEFER && e.loopDepth == 1) {
		if recv != nil {
			recvK = e.heapHole()
		}
		for range args {
			paramKs = append(paramKs, e.heapHole())
		}
	} else if direct && fn.Name.Defn != nil && fn.Name.Defn.Esc < EscFuncTagged {
		// Direct call to function in same mutually recursive
		// group; incorporate into data flow graph.

		if fn.Name.Defn.Esc == EscFuncUnknown {
			Fatalf("graph inconsistency")
		}

		if ks != nil {
			for i, result := range fntype.Results().FieldSlice() {
				e.value(ks[i], asNode(result.Nname))
			}
		}

		if r := fntype.Recv(); r != nil {
			recvK = e.addr(asNode(r.Nname))
		}
		for _, param := range fntype.Params().FieldSlice() {
			paramKs = append(paramKs, e.addr(asNode(param.Nname)))
		}
	} else if call.Op == OCALLFUNC || call.Op == OCALLMETH || call.Op == OCALLINTER {
		// Indirect call, or call to previously tagged
		// function. Setup flows to heap and/or ks according
		// to parameter tags.
		if r := fntype.Recv(); r != nil {
			recvK = e.tagHole(ks, r, direct, where)
		}
		for _, param := range fntype.Params().FieldSlice() {
			paramKs = append(paramKs, e.tagHole(ks, param, direct, where))
		}
	} else {
		// Handle escape analysis for builtins.

		// By default, we just discard everything. However, if
		// we're in a top-level defer statement, we can't
		// allow transient values.
		k := e.discardHole()
		if where != nil {
			k = e.newLoc(where, false).asHole()
		}
		for range args {
			paramKs = append(paramKs, k)
		}

		switch call.Op {
		case OAPPEND:
			// Appendee slice may flow directly to the
			// result, if it has enough
			// capacity. Alternatively, a new heap slice
			// might be allocated, and all slice elements
			// might flow to heap.
			paramKs[0] = e.teeHole(paramKs[0], ks[0])
			if types.Haspointers(args[0].Type.Elem()) {
				paramKs[0] = e.teeHole(paramKs[0], e.heapHole().deref(call, "appendee slice"))
			}

			if call.IsDDD() {
				if args[1].Type.IsSlice() && types.Haspointers(args[1].Type.Elem()) {
					paramKs[1] = e.teeHole(paramKs[1], e.heapHole().deref(call, "appended slice..."))
				}
			} else {
				for i := 1; i < len(args); i++ {
					paramKs[i] = e.heapHole()
				}
			}

		case OCOPY:
			if call.Right.Type.IsSlice() && types.Haspointers(call.Right.Type.Elem()) {
				paramKs[1] = e.teeHole(paramKs[1], e.heapHole().deref(call, "copied slice"))
			}

		case OPANIC:
			paramKs[0] = e.heapHole()
		}
	}

	// TODO(mdempsky): Remove after early ddd-ification.
	if fntype != nil && fntype.IsVariadic() && !call.IsDDD() {
		vi := fntype.NumParams() - 1

		elt := fntype.Params().Field(vi).Type.Elem()
		nva := call.List.Len()
		nva -= vi

		// Introduce ODDDARG node to represent ... allocation.
		ddd := nodl(call.Pos, ODDDARG, nil, nil)
		ddd.Type = types.NewPtr(types.NewArray(elt, int64(nva)))
		call.Right = ddd

		dddK := e.spill(paramKs[vi], ddd)
		paramKs = paramKs[:vi]
		for i := 0; i < nva; i++ {
			paramKs = append(paramKs, dddK)
		}
	}

	if call.Op == OCALLFUNC {
		// Evaluate callee function expression.
		k := e.discardHole()
		if where != nil {
			if where.Op == ODEFER && e.loopDepth == 1 {
				k = e.newLoc(nil, false).asHole()
			} else {
				k = e.heapHole()
			}
		}
		e.value(k, call.Left)
	}

	if recv != nil {
		// TODO(mdempsky): Handle go:uintptrescapes here too?
		e.value(recvK, recv)
	}

	for i, arg := range args {
		// For arguments to go:uintptrescapes, peel
		// away an unsafe.Pointer->uintptr conversion,
		// if present.
		if direct && arg.Op == OCONVNOP && arg.Type.Etype == TUINTPTR && arg.Left.Type.Etype == TUNSAFEPTR {
			x := i
			if fntype.IsVariadic() && x >= fntype.NumParams() {
				x = fntype.NumParams() - 1
			}
			if fntype.Params().Field(x).Note == uintptrEscapesTag {
				arg = arg.Left
			}
		}

		e.value(paramKs[i], arg)
	}
}

// tagHole returns a context for evaluating an argument passed to
// param. ks should be the contexts for the function callee's results;
// direct indicates whether this is a direct call; where is the
// OGO/ODEFER context of the call, if any.
func (e *Escape) tagHole(ks []EscHole, param *types.Field, direct bool, where *Node) EscHole {
	// If this isn't a direct call, we can't rely on param.Note.
	if !direct {
		return e.heapHole()
	}

	esc := parsetag(param.Note)
	switch esc {
	case EscHeap, EscUnknown:
		return e.heapHole()
	}

	var tagKs []EscHole
	if where != nil {
		tagKs = append(tagKs, e.newLoc(nil, false).asHole())
	}

	if esc&EscContentEscapes != 0 {
		tagKs = append(tagKs, e.heapHole().shift(1))
	}

	if ks != nil {
		for i := 0; i < numEscReturns; i++ {
			if x := getEscReturn(esc, i); x >= 0 {
				tagKs = append(tagKs, ks[i].shift(x))
			}
		}
	}

	return e.teeHole(tagKs...)
}

// An EscLocation represents an abstract location that stores a Go
// variable.
type EscLocation struct {
	n         *Node
	curfn     *Node
	edges     []EscEdge
	loopDepth int

	distance  int
	walkgen   uint32
	escapes   bool
	transient bool
	paramEsc  uint16
}

func (l *EscLocation) isName(c Class) bool {
	return l.n != nil && l.n.Op == ONAME && l.n.Class() == c
}

type EscEdge struct {
	src    *EscLocation
	derefs int
}

// An EscHole represents a context for evaluation a Go
// expression. Intuitively, when evaluating x in "l = **x", we'd have
// a hole with dst==l and derefs==2.
type EscHole struct {
	dst    *EscLocation
	derefs int
}

func (k EscHole) note(where *Node, why string) EscHole {
	// TODO(mdempsky): Keep a record of where/why for diagnostics.
	return k
}

func (k EscHole) shift(delta int) EscHole {
	k.derefs += delta
	if k.derefs < -1 {
		Fatalf("derefs underflow: %v", k.derefs)
	}
	return k
}

func (k EscHole) deref(where *Node, why string) EscHole { return k.shift(1).note(where, why) }
func (k EscHole) addr(where *Node, why string) EscHole  { return k.shift(-1).note(where, why) }

func (k EscHole) dottype(t *types.Type, where *Node, why string) EscHole {
	if !t.IsInterface() && !isdirectiface(t) {
		k = k.shift(1)
	}
	return k.note(where, why)
}

// teeHole returns a new hole that flows into each hole of ks,
// similar to the Unix tee(1) command.
func (e *Escape) teeHole(ks ...EscHole) EscHole {
	if len(ks) == 0 {
		return e.discardHole()
	}
	if len(ks) == 1 {
		return ks[0]
	}
	// TODO(mdempsky): Optimize if there's only one non-discard hole?

	// Given holes "l1 = _", "l2 = **_", "l3 = *_", ..., create a
	// new temporary location ltmp, wire it into place, and return
	// a hole for "ltmp = _".
	loc := e.newLoc(nil, true)
	for _, k := range ks {
		// N.B., "p = &q" and "p = &tmp; tmp = q" are not
		// semantically equivalent. To combine holes like "l1
		// = _" and "l2 = &_", we'd need to wire them as "l1 =
		// *ltmp" and "l2 = ltmp" and return "ltmp = &_"
		// instead.
		if k.derefs < 0 {
			Fatalf("teeHole: negative derefs")
		}

		e.flow(k, loc)
	}
	return loc.asHole()
}

func (e *Escape) dcl(n *Node) EscHole {
	loc := e.oldLoc(n)
	loc.loopDepth = e.loopDepth
	return loc.asHole()
}

func (e *Escape) spill(k EscHole, n *Node) EscHole {
	// TODO(mdempsky): Optimize. E.g., if k is the heap or blank,
	// then we already know whether n leaks, and we can return a
	// more optimized hole.
	loc := e.newLoc(n, true)
	e.flow(k.addr(n, "spill"), loc)
	return loc.asHole()
}

// canonicalNode returns the canonical *Node that n logically
// represents.
func canonicalNode(n *Node) *Node {
	if n != nil && n.IsClosureVar() {
		n = n.Name.Defn
		if n.IsClosureVar() {
			Fatalf("still closure var")
		}
	}

	return n
}

func (e *Escape) newLoc(n *Node, transient bool) *EscLocation {
	if e.curfn == nil {
		Fatalf("Curfn isn't set")
	}

	n = canonicalNode(n)
	loc := &EscLocation{
		n:         n,
		curfn:     e.curfn,
		loopDepth: e.loopDepth,
		transient: transient,
	}
	e.allLocs = append(e.allLocs, loc)
	if n != nil {
		if n.Op == ONAME && n.Name.Curfn != e.curfn {
			Fatalf("Curfn mismatch: %v != %v", n.Name.Curfn, e.curfn)
		}

		if n.HasOpt() {
			Fatalf("%v already has a location", n)
		}
		n.SetOpt(loc)

		// TODO(mdempsky): Perhaps set n.Esc and then just return &HeapLoc?
		// TODO(mdempsky): Cleanup this mess.
		if n.Type != nil && !loc.isName(PPARAM) && !loc.isName(PPARAMOUT) &&
			(n.Type.Width > maxStackVarSize ||
				(n.Op == ONEW || n.Op == OPTRLIT) && n.Type.Elem().Width >= maxImplicitStackVarSize ||
				n.Op == OMAKESLICE && !isSmallMakeSlice(n)) {
			e.flow(e.heapHole().addr(nil, ""), loc)
		}
	}
	return loc
}

func (e *Escape) oldLoc(n *Node) *EscLocation {
	n = canonicalNode(n)
	return n.Opt().(*EscLocation)
}

func (l *EscLocation) asHole() EscHole {
	return EscHole{dst: l}
}

func (e *Escape) flow(k EscHole, src *EscLocation) {
	dst := k.dst
	if dst == &e.blankLoc {
		return
	}
	if dst == src && k.derefs >= 0 {
		return
	}
	// TODO(mdempsky): More optimizations?

	// TODO(mdempsky): Deduplicate edges?
	dst.edges = append(dst.edges, EscEdge{src: src, derefs: k.derefs})
}

func (e *Escape) heapHole() EscHole    { return e.heapLoc.asHole() }
func (e *Escape) discardHole() EscHole { return e.blankLoc.asHole() }

func (e *Escape) resultHoles() []EscHole {
	var ks []EscHole
	for _, f := range e.curfn.Type.Results().FieldSlice() {
		ks = append(ks, e.addr(asNode(f.Nname)))
	}
	return ks
}

func (e *Escape) flood() {
	var walkgen uint32

	for _, loc := range e.allLocs {
		walkgen++
		e.walk(loc, walkgen)
	}

	walkgen++
	e.walk(&e.heapLoc, walkgen)
}

func (e *Escape) walk(root *EscLocation, walkgen uint32) {
	// Find the shortest path from root to all other
	// locations. The data flow graph has negative edges, so we
	// have to use the Bellman-Ford algorithm. However, path
	// lengths are capped at 0, so we don't need to worry about
	// negative cycles.
	root.walkgen = walkgen
	root.distance = 0

	todo := []*EscLocation{root}
	for len(todo) > 0 {
		p := todo[len(todo)-1]
		todo = todo[:len(todo)-1]

		base := p.distance

		addressOf := base < 0
		if addressOf {
			base = 0

			// TODO(mdempsky): Explain this.
			if !root.transient {
				p.transient = false
			}
		}

		// p's address flows to root. If root outlives
		// p, then p needs to be heap allocated.
		if e.outlives(root, p) {
			if addressOf && !p.escapes {
				p.escapes = true

				// TODO(mdempsky): This is clumsy.
				if root != &e.heapLoc {
					e.flow(e.heapLoc.asHole(), p)
				}
			}

			// p's value flows to root. If p is a function
			// parameter and root is the heap or a corresponding
			// result parameter, then record that value flow for
			// tagging the function later.
			if p.isName(PPARAM) {
				p.leak(root, base)
			}
		}

		for _, edge := range p.edges {
			dist := base + edge.derefs
			if edge.src.walkgen != walkgen || edge.src.distance > dist {
				edge.src.walkgen = walkgen
				edge.src.distance = dist
				todo = append(todo, edge.src)
			}
		}
	}
}

// outlives reports whether values stored in l may survive beyond
// other's lifetime if stack allocated.
func (e *Escape) outlives(l, other *EscLocation) bool {
	// The heap outlives everything.
	if l == &e.heapLoc {
		return true
	}

	// We don't know what callers do with returned values, so
	// pessimistically we need to assume they flow to the heap and
	// outlive everything too.
	if l.isName(PPARAMOUT) {
		// Exception: Directly called closures can return
		// locations allocated outside of them without forcing
		// them to the heap. For example:
		//
		//    var u int  // okay to stack allocate
		//    *(func() *int { return &u }()) = 42
		if containsClosure(other.curfn, l.curfn) && l.curfn.Func.Closure.Func.Top&ctxCallee != 0 {
			return false
		}

		return true
	}

	// If l and other are within the same function, then l
	// outlives other if it was declared outside other's loop
	// scope. For example:
	//
	//    var l *int
	//    for {
	//        l = new(int)
	//    }
	if l.curfn == other.curfn && l.loopDepth < other.loopDepth {
		return true
	}

	// If other is declared within a child closure of where l is
	// declared, then l outlives it. For example:
	//
	//    var l *int
	//    func() {
	//        l = new(int)
	//    }
	if containsClosure(l.curfn, other.curfn) {
		return true
	}

	return false
}

// containsClosure reports whether c is a closure contained within f.
func containsClosure(f, c *Node) bool {
	if f.Op != ODCLFUNC || c.Op != ODCLFUNC {
		Fatalf("bad containsClosure: %v, %v", f, c)
	}

	// Common case.
	if f == c {
		return false
	}

	// Closures within function Foo are named like "Foo.funcN..."
	// TODO(mdempsky): Better way to recognize this.
	fn := f.Func.Nname.Sym.Name
	cn := c.Func.Nname.Sym.Name
	return len(cn) > len(fn) && cn[:len(fn)] == fn && cn[len(fn)] == '.'
}

// leak records that parameter l leaks to sink.
func (l *EscLocation) leak(sink *EscLocation, derefs int) {
	// Short circuit if l already leaks to heap.
	if l.paramEsc == EscHeap {
		return
	}

	// If sink is a result parameter and we can fit return bits
	// into the escape analysis tag, then record a return leak.
	if sink.isName(PPARAMOUT) && sink.curfn == l.curfn {
		// TODO(mdempsky): Eliminate dependency on Vargen here.
		ri := int(sink.n.Name.Vargen) - 1
		if ri < numEscReturns {
			// Leak to result parameter.
			if old := getEscReturn(l.paramEsc, ri); old < 0 || derefs < old {
				l.paramEsc = setEscReturn(l.paramEsc, ri, derefs)
			}
			return
		}
	}

	// Otherwise, record as heap leak.
	if derefs > 0 {
		l.paramEsc |= EscContentEscapes
	} else {
		l.paramEsc = EscHeap
	}
}

func (e *Escape) finish() {
	for _, loc := range e.allLocs {
		n := loc.n
		if n == nil {
			continue
		}
		n.SetOpt(nil)

		// TODO(mdempsky): Describe path when Debug['m'] >= 2.

		if loc.escapes {
			if Debug['m'] != 0 && n.Op != ONAME {
				Warnl(n.Pos, "%S escapes to heap", n)
			}
			n.Esc = EscHeap
			addrescapes(n)
		} else if loc.isName(PPARAM) {
			n.Esc = finalizeEsc(loc.paramEsc)

			if Debug['m'] != 0 && types.Haspointers(n.Type) {
				if n.Esc == EscNone {
					Warnl(n.Pos, "%S %S does not escape", funcSym(loc.curfn), n)
				} else if n.Esc == EscHeap {
					Warnl(n.Pos, "leaking param: %S", n)
				} else {
					if n.Esc&EscContentEscapes != 0 {
						Warnl(n.Pos, "leaking param content: %S", n)
					}
					for i := 0; i < numEscReturns; i++ {
						if x := getEscReturn(n.Esc, i); x >= 0 {
							res := n.Name.Curfn.Type.Results().Field(i).Sym
							Warnl(n.Pos, "leaking param: %S to result %v level=%d", n, res, x)
						}
					}
				}
			}
		} else {
			n.Esc = EscNone
			if loc.transient {
				switch n.Op {
				case OCALLPART, OCLOSURE, ODDDARG, OARRAYLIT, OSLICELIT, OPTRLIT, OSTRUCTLIT:
					n.SetNoescape(true)
				}
			}

			if Debug['m'] != 0 && n.Op != ONAME && n.Op != OTYPESW && n.Op != ORANGE && n.Op != ODEFER {
				Warnl(n.Pos, "%S %S does not escape", funcSym(loc.curfn), n)
			}
		}
	}
}

func finalizeEsc(esc uint16) uint16 {
	esc = optimizeReturns(esc)

	if esc>>EscReturnBits != 0 {
		esc |= EscReturn
	} else if esc&EscMask == 0 {
		esc |= EscNone
	}

	return esc
}

func optimizeReturns(esc uint16) uint16 {
	if esc&EscContentEscapes != 0 {
		// EscContentEscapes represents a path of length 1
		// from the heap. No point in keeping paths of equal
		// or longer length to result parameters.
		for i := 0; i < numEscReturns; i++ {
			if x := getEscReturn(esc, i); x >= 1 {
				esc = setEscReturn(esc, i, -1)
			}
		}
	}
	return esc
}

// The escape bits saved for each analyzed parameter record the
// shortest path (if any) from that parameter to the heap, or to any
// of its function's (first numEscReturns) result parameters.
//
// Paths to the heap are encoded via EscHeap (length 0) or
// EscContentEscapes (length 1); if neither of these are set, then
// there's no path to the heap.
//
// Paths to the result parameters are encoded in the upper
// bits.
//
// There are other values stored in the escape bits by esc.go for
// vestigial reasons, and other special tag values used (e.g.,
// uintptrEscapesTag and unsafeUintptrTag). These could be simplified
// once compatibility with esc.go is no longer a concern.

const numEscReturns = (16 - EscReturnBits) / bitsPerOutputInTag

func getEscReturn(esc uint16, i int) int {
	return int((esc>>escReturnShift(i))&bitsMaskForTag) - 1
}

func setEscReturn(esc uint16, i, v int) uint16 {
	if v < -1 {
		Fatalf("invalid esc return value: %v", v)
	}
	if v > maxEncodedLevel {
		v = maxEncodedLevel
	}

	shift := escReturnShift(i)
	esc &^= bitsMaskForTag << shift
	esc |= uint16(v+1) << shift
	return esc
}

func escReturnShift(i int) uint {
	if uint(i) >= numEscReturns {
		Fatalf("esc return index out of bounds: %v", i)
	}
	return uint(EscReturnBits + i*bitsPerOutputInTag)
}

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

import "fmt"

type edge_mem_ctr struct {
	e Edge
	m *Value // phi for memory at dest of e
	c *Value // phi for counter at dest of e
}

type rewrite struct {
	before, after *Value
}

func altinsertLoopReschedChecks(f *Func) {
	// The name of this phase "insert resched checks" is known
	// to deadstore.go, which must precede it to ensure that
	// f.lastMems is correctly initialized.

	if f.NoSplit { // nosplit functions don't reschedule.
		return
	}

	idom := f.Idom()
	backedges := f.backedges()
	po := f.postorder()
	// sdom requires order in children
	sdom := newOrderedSparseTree(f, po, idom)
	f.cachedSdom = sdom

	if f.pass.debug > 2 {
		fmt.Printf("%s = %s\n", f.Name, sdom.treestructure(f.Entry))
	}

	domstack := []*Block{}
	domlastcall := make([]*Block, f.NumBlocks()) // nearest dom-or-eq that contains call.
	ownlastcall := make([]bool, f.NumBlocks())   // true if contains call.

	ncainstack := func(x, y *Block) *Block {
		lo := 0 // lo is known-good, hi is unknown; domstack[0] is root of tree.
		for hi := len(domstack) - 1; lo < hi; {
			mid := (lo + hi + 1) / 2
			c := domstack[mid]
			if sdom.isAncestorEq(c, x) && sdom.isAncestorEq(c, y) {
				lo = mid // lo is known good; domstack[lo] is a common ancestor.
			} else {
				hi = mid - 1 // hi is unknown
			}
		}
		return domstack[lo]
	}

	// Recursive function to fill in domlastcall, and ownlastcall for every block.
	var visit func(b *Block)
	visit = func(b *Block) {
		domstack = append(domstack, b)
		l := len(domstack)

		if b.containsCall() { // TODO: check call for nosplit annotation
			if f.pass.debug > 2 {
				fmt.Printf("Lastcall %v is %v (contains call)\n", b, b)
			}
			domlastcall[b.ID] = b
			ownlastcall[b.ID] = true
		} else {
			var ca *Block // common ancestor of lastcalls of predecessors, or nil if none.
			for _, e := range b.Preds {
				p := e.b
				if sdom.isAncestorEq(b, p) {
					continue // reducible loop backedge, b ancestor of b's predecessor p.
				}
				lc := domlastcall[p.ID]
				if f.pass.debug > 2 {
					fmt.Printf("Lastcall %v pred %v lastcall %v\n", b, p, lc)
				}

				if lc == nil {
					// There is a path with no calls on it at all.
					ca = nil
					break
				}
				if ca == nil || sdom.isAncestor(ca, lc) {
					// lc is first predecessor or lower than ca
					ca = lc
					continue
				}
				if sdom.isAncestorEq(lc, ca) {
					// ca is already lower-eq than lc
					continue
				}
				// unordered: binary search for nearest common ancestor in work stack.
				ca = ncainstack(ca, lc)
			}
			if f.pass.debug > 2 {
				fmt.Printf("Lastcall %v is %v\n", b, ca)
			}
			domlastcall[b.ID] = ca
			// if ca is nil then there is a call-free path to b.
			// otherwise, ca is the lowest node in the dominator tree
			// guaranteeing all paths to b contain a call.
		}
		for c := sdom.Child(b); c != nil; c = sdom.Sibling(c) {
			visit(c)
		}
		domstack = domstack[0 : l-1]
	}

	visit(f.Entry)

	tofix_backedges := []edge_mem_ctr{}

	// filter backedges to discover those needing reschedule checks.
	for _, e := range backedges {
		// edge from bb to h.
		// if lc[bb] is not nil and is dominated by h, then any path
		// from h to h containing this backedge already contains a rescheduling check.
		h := e.b
		if ownlastcall[h.ID] {
			// header contains call
			continue
		}
		bb := h.Preds[e.i].b
		lc := domlastcall[bb.ID]
		if lc != nil && sdom.isAncestorEq(h, lc) {
			if f.pass.debug > 1 {
				f.Config.Warnl(bb.Line, "%s: Backedge from %v to %v has lastcall %v below header (no check needed)", f.Name, bb, h, lc)
			}
			continue
		}
		if f.pass.debug > 1 {
			f.Config.Warnl(bb.Line, "%s: Backedge from %v to %v has lastcall %v nil or not beloweq header", f.Name, bb, h, lc)
		}
		// need a reschedule check on bb -> h
		tofix_backedges = append(tofix_backedges, edge_mem_ctr{e, nil, nil})
	}

	if len(tofix_backedges) == 0 {
		return
	}

	// It's possible that there is no memory state (no global/pointer loads/stores or calls)
	if f.lastMems[f.Entry.ID] == nil {
		f.lastMems[f.Entry.ID] = f.Entry.NewValue0(f.Entry.Line, OpInitMem, TypeMem)
	}

	memDefsForUses := make([]*Value, f.NumBlocks()) // For each block, the mem def seen at its bottom.
	var visitmem func(b *Block, mem *Value)
	visitmem = func(b *Block, mem *Value) {
		if f.lastMems[b.ID] != nil {
			mem = f.lastMems[b.ID]
		}
		memDefsForUses[b.ID] = mem
		for c := sdom.Child(b); c != nil; c = sdom.Sibling(c) {
			visitmem(c, mem)
		}
	}

	// Propagate available memory operand.
	// This may be an overestimate of actual range, but "mem" is not register-allocated.
	visitmem(f.Entry, nil)

	// Set up counter.  There are no phis etc pre-existing for it.
	counter0 := f.Entry.NewValue0I(f.Entry.Line, OpConst32, f.Config.fe.TypeInt32(), 32707) // it's prime.
	ctrDefsForUses := make([]*Value, f.NumBlocks())                                         // For each block, the use-of/def-seen

	for _, emc := range tofix_backedges {
		e := emc.e
		// set initial uses of counter zero (note available-at-bottom and use are the same thing initially.)
		// each back-edge will be rewritten to include a reschedule check, and that will use the counter.
		src := e.b.Preds[e.i].b
		ctrDefsForUses[src.ID] = counter0
	}

	// Push uses towards root
	for _, b := range f.postorder() {
		bd := ctrDefsForUses[b.ID]
		if bd == nil {
			continue
		}
		for _, e := range b.Preds {
			p := e.b
			if ctrDefsForUses[p.ID] == nil {
				ctrDefsForUses[p.ID] = bd
			}
		}
	}

	// Map from block newly-inserted phi function in block.
	newmemphis := make(map[*Block]rewrite)
	newctrphis := make(map[*Block]rewrite)

	// Insert phi functions as necessary for future changes to flow graph.
	for i, emc := range tofix_backedges {
		e := emc.e
		h := e.b

		// find the phi function for the memory input at "h", if there is one.
		var headerMemPhi *Value // look for header mem phi

		for _, v := range h.Values {
			if v.Op == OpPhi && v.Type.IsMemory() {
				headerMemPhi = v
			}
		}

		if headerMemPhi == nil {
			// if the header is nil, make a trivial phi from the dominator
			mem0 := memDefsForUses[idom[h.ID].ID]
			headerMemPhi = newPhiFor(h, mem0)
			newmemphis[h] = rewrite{mem0, headerMemPhi}
			addDFphis(mem0, h, h, f, memDefsForUses, newmemphis)

		}
		tofix_backedges[i].m = headerMemPhi

		var headerCtrPhi *Value
		rw, ok := newctrphis[h]
		if !ok {
			headerCtrPhi = newPhiFor(h, counter0)
			newctrphis[h] = rewrite{counter0, headerCtrPhi}
			addDFphis(counter0, h, h, f, ctrDefsForUses, newctrphis)
		} else {
			headerCtrPhi = rw.after
		}
		tofix_backedges[i].c = headerCtrPhi
	}

	rewriteNewPhis(nil, nil, f.Entry, f.Entry, f, newmemphis)
	rewriteNewPhis(nil, nil, f.Entry, f.Entry, f, newctrphis)

	// Rewrite backedges to include reschedule checks.
	var invalidateCFG bool
	for _, emc := range tofix_backedges {
		e := emc.e
		headerMemPhi := emc.m
		headerCtrPhi := emc.c
		h := e.b
		i := e.i
		p := h.Preds[i]
		bb := p.b
		mem0 := headerMemPhi.Args[i]
		ctr0 := headerCtrPhi.Args[i]
		_ = ctr0
		// bb e->p h,
		// Because we're going to insert a rare-call, make sure the
		// looping edge still looks likely.
		likely := BranchLikely
		if p.i != 0 {
			likely = BranchUnlikely
		}
		bb.Likely = likely

		invalidateCFG = true

		// rewrite edge to include reschedile check
		// existing edges:
		//
		// bb.Succs[p.i] == Edge{h, i}
		// h.Preds[i] == p == Edge{bb,p.i}
		//
		// new block(s):
		// test:
		//    ctr1 := ctr0 - 1
		//    if ctr1 <= 0 { goto sched }
		//    goto join
		// sched:
		//    mem1 := call resched (mem0)
		//    goto join
		// join:
		//    ctr2 := phi(ctr1, counter0) // counter0 is the constant
		//    mem2 := phi(mem0, mem1)
		//    goto h
		//
		// and correct arg i of headerMemPhi and headerCtrPhi

		test := f.NewBlock(BlockIf)
		sched := f.NewBlock(BlockPlain)
		join := f.NewBlock(BlockPlain)

		test.Line = bb.Line
		sched.Line = bb.Line
		join.Line = bb.Line

		zero := test.NewValue0I(bb.Line, OpConst32, f.Config.fe.TypeInt32(), 0)
		one := test.NewValue0I(bb.Line, OpConst32, f.Config.fe.TypeInt32(), 1)

		//    ctr1 := ctr0 - 1
		//    if ctr1 <= 0 { goto sched }
		//    goto join
		ctr1 := test.NewValue2(bb.Line, OpSub32, f.Config.fe.TypeInt32(), ctr0, one)
		cmp := test.NewValue2(bb.Line, OpLeq32, f.Config.fe.TypeBool(), ctr1, zero)
		test.SetControl(cmp)
		test.AddEdgeTo(sched) // if true
		test.AddEdgeTo(join)  // if false -- 1st incoming edge to join
		test.Likely = BranchUnlikely

		// sched:
		//    mem1 := call resched (mem0)
		//    goto join
		resched := f.Config.fe.Syslook("goschedguarded")
		mem1 := sched.NewValue1A(bb.Line, OpStaticCall, TypeMem, resched, mem0)
		sched.AddEdgeTo(join) // 2nd incoming edge to join

		// join:
		//    ctr2 := phi(ctr1, counter0) // counter0 is the constant
		//    mem2 := phi(mem0, mem1)
		//    goto h
		mem2 := join.NewValue2(bb.Line, OpPhi, TypeMem, mem0, mem1)
		ctr2 := join.NewValue2(bb.Line, OpPhi, TypeMem, ctr1, counter0)
		// In-place edge change
		h.Preds[i] = Edge{join, 0}
		join.Succs = append(join.Succs, Edge{h, i})

		// adjust phi function in header.
		headerMemPhi.SetArg(i, mem2)
		headerCtrPhi.SetArg(i, ctr2)

		bb.Succs[p.i] = Edge{test, 0}
		test.Preds = append(test.Preds, Edge{bb, p.i})
	}

	if invalidateCFG {
		f.invalidateCFG()
	}
	// Reclaim space.
	f.lastMems = nil
	return
}

// newPhiFor inserts a new Phi function into b,
// with all inputs set to v.
func newPhiFor(b *Block, v *Value) *Value {
	phiV := b.NewValue0(b.Line, OpPhi, v.Type)

	for range b.Preds {
		phiV.AddArg(v)
	}
	return phiV
}

func rewriteNewPhis(x, y *Value, h, b *Block, f *Func, newphis map[*Block]rewrite) {
	// If b is a new phi, then a new rewrite applies
	if change, ok := newphis[b]; ok {
		h = b
		x = change.before
		y = change.after
	}

	sdom := f.sdom()

	// Apply rewrites to this block
	if x != nil { // don't waste time on this common case.
		for _, v := range b.Values {
			if v == y {
				continue
			}
			for i, w := range v.Args {
				if w == x {
					v.SetArg(i, y)
				}
			}
		}

		// Rewrite appropriate inputs of phis reached in dominance frontier (non-dominated successors)
		// Note this can include self in case of single-block loops.
		for _, e := range b.Succs {
			s := e.b
			if sdom.isAncestor(h, s) {
				continue
			}
			for _, v := range s.Values {
				if v.Op == OpPhi && v.Args[e.i] == x {
					// TODO might be more than one phi in general, though not for use here.
					v.SetArg(e.i, y)
					break
				}
			}
		}
	}

	for c := sdom[b.ID].child; c != nil; c = sdom[c.ID].sibling {
		rewriteNewPhis(x, y, h, c, f, newphis)
	}
}

func addDFphis(x *Value, h, b *Block, f *Func, defForUses []*Value, newphis map[*Block]rewrite) {
	oldv := defForUses[b.ID]
	if oldv != x { // either no uses, or a new definition replacing x.
		return
	}
	sdom := f.sdom()
	idom := f.Idom()
outer:
	for _, e := range b.Succs {
		s := e.b
		// check phi functions in the dominance frontier
		if sdom.isAncestor(h, s) {
			continue
		}
		if _, ok := newphis[s]; ok {
			continue
		}
		if x != nil {
			for _, v := range s.Values {
				if v.Op == OpPhi && v.Args[e.i] == x {
					continue outer
				}
			}
		}

		old := defForUses[idom[s.ID].ID]
		headerPhi := newPhiFor(s, old)
		// f.Config.Warnl(s.Line, "addDFphis: h=%v, b=%v, s=%v, oldv=%v, x=%v", h, b, s, oldv, x)
		newphis[s] = rewrite{old, headerPhi}
	}
	for c := sdom[b.ID].child; c != nil; c = sdom[c.ID].sibling {
		addDFphis(x, h, c, f, defForUses, newphis)
	}
}

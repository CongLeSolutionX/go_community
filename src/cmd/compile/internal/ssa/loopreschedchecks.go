// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

import "fmt"

type edgeMemCtr struct {
	e Edge
	m *Value // phi for memory at dest of e
	c *Value // phi for counter at dest of e
}

type rewriteTarget struct {
	v *Value
	i int
}

type rewrite struct {
	before, after *Value
	rewrites      *[]rewriteTarget
}

func (r *rewrite) String() string {
	s := "\n\tbefore=" + r.before.String() + ", after=" + r.after.String()
	for _, rw := range *r.rewrites {
		s += ", (i=" + fmt.Sprint(rw.i) + ", v=" + rw.v.LongString() + ")"
	}
	s += "\n"
	return s
}

const initialRescheduleCounterValue = 32749 // Largest 16-bit-signed prime.

// insertLoopReschedChecks inserts rescheduling checks on the loop
// backedges.  This version targets every loop backedge because we
// currently lack split information in export data.
func insertLoopReschedChecks(f *Func) {

	// Loop reschedule checks decrement a per-function counter
	// shared by all loops, and when the counter becomes non-positive
	// a call is made to a rescheduling check in the runtime.
	//
	// Steps:
	// 1. locate backedges.
	// 2. Record memory definitions at block end so that
	//    the SSA graph for mem can be prperly modified.
	// 3. Define a counter and record its future uses (at backedges)
	//    (Same process as 2, applied to a single definition of the counter.
	//     difference for mem is that there are zero-to-many existing mem
	//     definitions, versus exactly one for the new counter.)
	// 4. Ensure that phi functions that will-be-needed for mem and counter
	//    are present in the graph, initially with trivial inputs.
	// 5. Record all to-be-modified uses of mem and counter;
	//    apply modifications (split into two steps to simplify and
	//    avoided nagging order-dependences).
	// 6. Rewrite backedges to include counter check, reschedule check,
	//    and modify destination phi function appropriately with new
	//    definitions for mem and counter.

	lastMemUses(f)

	if f.NoSplit { // nosplit functions don't reschedule.
		return
	}

	idom := f.Idom()
	backedges := backedges(f)
	po := f.postorder()
	// sdom requires order in children
	// TODO: this may not be necessary anymore, delicate order issues dealt with differently.
	// however this order does yield a prettier flowgraph when printed for debugging.
	sdom := newSparseTree(f, idom)
	f.cachedSdom = sdom

	if f.pass.debug > 2 {
		fmt.Printf("before %s = %s\n", f.Name, sdom.treestructure(f.Entry))
	}

	tofixBackedges := []edgeMemCtr{}

	if len(backedges) == 0 {
		return
	}

	for _, e := range backedges { // TODO: could filter here by calls in loops, if declared and inferred nosplit are recorded in export data.
		tofixBackedges = append(tofixBackedges, edgeMemCtr{e, nil, nil})
	}

	// It's possible that there is no memory state (no global/pointer loads/stores or calls)
	if f.lastMems[f.Entry.ID] == nil {
		f.lastMems[f.Entry.ID] = f.Entry.NewValue0(f.Entry.Line, OpInitMem, TypeMem)
	}

	memDefsAtBlockEnds := make([]*Value, f.NumBlocks()) // For each block, the mem def seen at its bottom. Could be from earlier block.
	var visitmem func(b *Block, mem *Value)
	visitmem = func(b *Block, mem *Value) {
		if f.lastMems[b.ID] != nil {
			mem = f.lastMems[b.ID]
		}
		memDefsAtBlockEnds[b.ID] = mem
		for c := sdom.Child(b); c != nil; c = sdom.Sibling(c) {
			visitmem(c, mem)
		}
	}

	// Propagate available memory operand.
	// This may be an overestimate of actual range, but "mem" is not register-allocated.
	visitmem(f.Entry, nil)

	// Set up counter.  There are no phis etc pre-existing for it.
	counter0 := f.Entry.NewValue0I(f.Entry.Line, OpConst32, f.Config.fe.TypeInt32(), initialRescheduleCounterValue)
	ctrDefsAtBlockEnds := make([]*Value, f.NumBlocks()) // For each block, def visible at its end, if that def will be used.

	// There's a minor difference between memDefsAtBlockEnds and ctrDefsAtBlockEnds;
	// because the counter only matter for loops and code that reaches them, it is nil for blocks where the ctr is no
	// longer live.  This will avoid creation of dead phi functions.  This optimization is ignored for the mem variable
	// because it is harder and also less likely to be helpful, though dead code elimination ought to clean this out anyhow.

	for _, emc := range tofixBackedges {
		e := emc.e
		// set initial uses of counter zero (note available-at-bottom and use are the same thing initially.)
		// each back-edge will be rewritten to include a reschedule check, and that will use the counter.
		src := e.b.Preds[e.i].b
		ctrDefsAtBlockEnds[src.ID] = counter0
	}

	// Push uses towards root
	for _, b := range f.postorder() {
		bd := ctrDefsAtBlockEnds[b.ID]
		if bd == nil {
			continue
		}
		for _, e := range b.Preds {
			p := e.b
			if ctrDefsAtBlockEnds[p.ID] == nil {
				ctrDefsAtBlockEnds[p.ID] = bd
			}
		}
	}

	// Maps from block to newly-inserted phi function in block.
	newmemphis := make(map[*Block]rewrite)
	newctrphis := make(map[*Block]rewrite)

	// Insert phi functions as necessary for future changes to flow graph.
	for i, emc := range tofixBackedges {
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
			mem0 := memDefsAtBlockEnds[idom[h.ID].ID]
			headerMemPhi = newPhiFor(h, mem0)
			newmemphis[h] = rewrite{mem0, headerMemPhi, new([]rewriteTarget)}
			addDFphis(mem0, h, h, f, memDefsAtBlockEnds, newmemphis)

		}
		tofixBackedges[i].m = headerMemPhi

		var headerCtrPhi *Value
		rw, ok := newctrphis[h]
		if !ok {
			headerCtrPhi = newPhiFor(h, counter0)
			newctrphis[h] = rewrite{counter0, headerCtrPhi, new([]rewriteTarget)}
			addDFphis(counter0, h, h, f, ctrDefsAtBlockEnds, newctrphis)
		} else {
			headerCtrPhi = rw.after
		}
		tofixBackedges[i].c = headerCtrPhi
	}

	rewriteNewPhis(nil, nil, f.Entry, f.Entry, f, memDefsAtBlockEnds, newmemphis)
	rewriteNewPhis(nil, nil, f.Entry, f.Entry, f, ctrDefsAtBlockEnds, newctrphis)

	if f.pass.debug > 0 {
		for b, r := range newmemphis {
			fmt.Printf("b=%s, rewrite=%s\n", b, r.String())
		}

		for b, r := range newctrphis {
			fmt.Printf("b=%s, rewrite=%s\n", b, r.String())
		}
	}

	// Apply collected rewrites.
	for _, r := range newmemphis {
		for _, rw := range *r.rewrites {
			rw.v.SetArg(rw.i, r.after)
		}
	}

	for _, r := range newctrphis {
		for _, rw := range *r.rewrites {
			rw.v.SetArg(rw.i, r.after)
		}
	}

	// Rewrite backedges to include reschedule checks.
	var invalidateCFG bool
	for _, emc := range tofixBackedges {
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

		// rewrite edge to include reschedule check
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
		//
		// EXCEPT: block containing only phi functions is bad
		// for the register allocator.  Therefore, there is no
		// join, and instead branches targeting join instead target
		// the header, and the other phi functions within header are
		// adjusted for the additional input.

		test := f.NewBlock(BlockIf)
		sched := f.NewBlock(BlockPlain)

		test.Line = bb.Line
		sched.Line = bb.Line

		zero := test.NewValue0I(bb.Line, OpConst32, f.Config.fe.TypeInt32(), 0)
		one := test.NewValue0I(bb.Line, OpConst32, f.Config.fe.TypeInt32(), 1)

		//    ctr1 := ctr0 - 1
		//    if ctr1 <= 0 { goto sched }
		//    goto header
		ctr1 := test.NewValue2(bb.Line, OpSub32, f.Config.fe.TypeInt32(), ctr0, one)
		cmp := test.NewValue2(bb.Line, OpLeq32, f.Config.fe.TypeBool(), ctr1, zero)
		test.SetControl(cmp)
		test.AddEdgeTo(sched) // if true
		// if false -- rewrite edge to header.
		// do NOT remove+add, because that will perturb all the other phi functions
		// as well as messing up other edges to the header.
		test.Succs = append(test.Succs, Edge{h, i})
		h.Preds[i] = Edge{test, 1}
		headerMemPhi.SetArg(i, mem0)
		headerCtrPhi.SetArg(i, ctr1)

		test.Likely = BranchUnlikely

		// sched:
		//    mem1 := call resched (mem0)
		//    goto header
		resched := f.Config.fe.Syslook("goschedguarded")
		mem1 := sched.NewValue1A(bb.Line, OpStaticCall, TypeMem, resched, mem0)
		sched.AddEdgeTo(h)
		headerMemPhi.AddArg(mem1)
		headerCtrPhi.AddArg(counter0)

		bb.Succs[p.i] = Edge{test, 0}
		test.Preds = append(test.Preds, Edge{bb, p.i})

		// Must correct all the other phi functions in the header for new incoming edge.
		for _, v := range h.Values {
			if v.Op == OpPhi && v != headerMemPhi && v != headerCtrPhi {
				v.AddArg(v.Args[i])
			}
		}
	}

	if invalidateCFG {
		f.invalidateCFG()
	}

	if f.pass.debug > 2 {
		po = f.postorder()
		idom = f.Idom()
		// sdom requires order in children
		sdom = newOrderedSparseTree(f, po, idom)
		fmt.Printf("after %s = %s\n", f.Name, sdom.treestructure(f.Entry))
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

// rewriteNewPhis records all the rewrites that will need to be applied to account for the new phi functions
// added to this function.  The recorded rewrites are applied in a batch, later, to simplify sanity check
// the sets of rewrites and ensure that weird rewrite-order issues don't cause problems.
func rewriteNewPhis(x, y *Value, h, b *Block, f *Func, defsForUses []*Value, newphis map[*Block]rewrite) {
	// If b is a new phi, then a new rewrite applies
	if change, ok := newphis[b]; ok {
		h = b
		x = change.before
		y = change.after
	}

	sdom := f.sdom()

	// Apply rewrites to this block
	if x != nil { // don't waste time on the common case of no definition.
		p := newphis[h].rewrites
		for _, v := range b.Values {
			if v != y { // don't rewrite self -- phi inputs are handled below.
				for i, w := range v.Args {
					if w == x {
						*p = append(*p, rewriteTarget{v, i})
					}
				}
			}
		}

		// Rewrite appropriate inputs of phis reached in successors
		// in dominance frontier, self, and dominated.
		// If the variable def reaching uses in b is itself defined in b, then the new phi function
		// does not reach the successors of b.  (This assumes a bit about the structure of the
		// phi use-def graph, but it's true for memory and the inserted counter.)
		if dfu := defsForUses[b.ID]; dfu != nil && dfu.Block != b {
			for _, e := range b.Succs {
				s := e.b
				if sphi, ok := newphis[s]; ok { // saves time to find the phi this way.
					*p = append(*p, rewriteTarget{sphi.after, e.i})
					continue
				}
				for _, v := range s.Values {
					if v.Op == OpPhi && v.Args[e.i] == x {
						*p = append(*p, rewriteTarget{v, e.i})
						break
					}
				}
			}
		}
	}

	for c := sdom[b.ID].child; c != nil; c = sdom[c.ID].sibling {
		rewriteNewPhis(x, y, h, c, f, defsForUses, newphis)
	}
}

// addDFphis adds phis needed because "x" had a new definition inserted at h (usually but not necessarily a phi)
// these phis necessarily occur at the dominance fronter of h, and they require a recursive call to addDFphis.
// All phis added assume that the definitions are not yet added; the inserted phis merely collect the originally
// reaching value and propagate it under a new name.
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
			continue // h dominates s, successor of b.
		}
		if _, ok := newphis[s]; ok {
			continue // successor s of b already has a new phi function.
		}
		if x != nil {
			for _, v := range s.Values {
				if v.Op == OpPhi && v.Args[e.i] == x {
					continue outer // successor s of b has an old phi function
				}
			}
		}

		old := defForUses[idom[s.ID].ID] // new phi function is correct-but-redundant, joining dominator def on all inputs.
		headerPhi := newPhiFor(s, old)
		newphis[s] = rewrite{old, headerPhi, new([]rewriteTarget)} // record new phi, to have inputs labeled "old" rewritten to "headerPhi"
		addDFphis(old, s, s, f, defForUses, newphis)               // the new definition may also create new phi functions.
	}
	for c := sdom[b.ID].child; c != nil; c = sdom[c.ID].sibling {
		addDFphis(x, h, c, f, defForUses, newphis)
	}
}

func lastMemUses(f *Func) {
	var stores []*Value
	f.lastMems = make([]*Value, f.NumBlocks())
	storeUse := f.newSparseSet(f.NumValues())
	defer f.retSparseSet(storeUse)
	for _, b := range f.Blocks {
		// Find all the stores in this block. Categorize their uses:
		//  storeUse contains stores which are used by a subsequent store.
		storeUse.clear()
		stores = stores[:0]
		var memPhi *Value
		for _, v := range b.Values {
			if v.Op == OpPhi {
				if v.Type.IsMemory() {
					memPhi = v
				}
				// Ignore phis - they will always be first and can't be eliminated
				continue
			}
			if v.Type.IsMemory() {
				stores = append(stores, v)
				if v.Op == OpSelect1 {
					// Use the args of the tuple-generating op.
					v = v.Args[0]
				}
				for _, a := range v.Args {
					if a.Block == b && a.Type.IsMemory() {
						storeUse.add(a.ID)
					}
				}
			}
		}
		if len(stores) == 0 {
			if len(f.lastMems) > 0 {
				f.lastMems[b.ID] = memPhi
			}
			continue
		}

		// find last store in the block
		var last *Value
		for _, v := range stores {
			if storeUse.contains(v.ID) {
				continue
			}
			if last != nil {
				b.Fatalf("two final stores - simultaneous live stores %s %s", last, v)
			}
			last = v
		}
		if last == nil {
			b.Fatalf("no last store found - cycle?")
		}
		if len(f.lastMems) > 0 {
			f.lastMems[b.ID] = last
		}
	}
}

// backedges returns a pointer-to-slice (so it can be nil if not
// initialized in the Func cache) of successor edges that are back
// edges.  For reducible loops, edge.b is the header.
func backedges(f *Func) []Edge {
	edges := []Edge{}
	mark := make([]markKind, f.NumBlocks())
	backedges1(f.Entry, &edges, mark)
	return edges
}

func backedges1(b *Block, be *[]Edge, mark []markKind) {
	mark[b.ID] = notExplored
	for _, e := range b.Succs {
		s := e.b
		if mark[s.ID] == notFound {
			// TODO undo this recursion.
			backedges1(s, be, mark)
		} else if mark[s.ID] == notExplored {
			*be = append(*be, e)
		}
	}
	mark[b.ID] = done
}

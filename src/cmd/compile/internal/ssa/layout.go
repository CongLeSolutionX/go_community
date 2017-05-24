// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

import "fmt"

type segment struct {
	first, last *Block
	parent      *segment
}

// layout orders basic blocks in f with the goal of minimizing control flow instructions.
// After this phase returns, the order of f.Blocks matters and is the order
// in which those blocks will appear in the assembly output.
func layout(f *Func) {
	order := make([]*Block, 0, f.NumBlocks())
	scheduled := make([]bool, f.NumBlocks())
	idToBlock := make([]*Block, f.NumBlocks())
	indegree := make([]int32, f.NumBlocks())
	b2PO := make([]int32, f.NumBlocks()) // For each block, its PO#, or -1 if not reachable

	posdegree := f.newSparseSet(f.NumBlocks()) // blocks with positive remaining degree
	defer f.retSparseSet(posdegree)
	zerodegree := f.newSparseSet(f.NumBlocks()) // blocks with zero remaining degree
	defer f.retSparseSet(zerodegree)
	exit := f.newSparseSet(f.NumBlocks()) // exit blocks
	defer f.retSparseSet(exit)

	ln := f.loopnest()
	ln.findExits()
	ln.calculateDepths()
	ln.findBlocks()
	ln.assembleChildren()

	for i := range b2PO {
		b2PO[i] = -1
	}
	for i, b := range ln.po {
		b2PO[b.ID] = int32(i)
	}

	// loopSegments := make(map[*loop]*segment)

	// var visitLoop func(l *loop)
	// visitLoop = func(l *loop) {
	// 	for _, inner := range l.children {
	// 		visitLoop(inner)
	// 	}
	// 	var bestLastBlock *Block
	// 	// the "best" last block is conditional,
	// 	// with one branch to the header of the current block,
	// 	// and the other branch an exit.
	// 	// the "best" exit is one that is "nearest"
	// 	// meaning a good choice to directly follow.
	// 	// For example, that has the larger post-order number.
	// 	for _, b := range l.blocks {

	// 	}
	// }

	// // iterate over top level loops,
	// // recursively visited interior loops,
	// // and layout loops from inner to outer.
	// for _, l := range ln.loops {
	// 	if l.outer != nil {
	// 		continue
	// 	}

	// }

	// Initialize indegree of each block
	for _, b := range f.Blocks {
		idToBlock[b.ID] = b
		if b.Kind == BlockExit {
			// exit blocks are always scheduled last
			// TODO: also add blocks post-dominated by exit blocks
			exit.add(b.ID)
			continue
		}
		indegree[b.ID] = int32(len(b.Preds))
		if len(b.Preds) == 0 {
			zerodegree.add(b.ID)
		} else {
			posdegree.add(b.ID)
		}
	}

	bid := f.Entry.ID
blockloop:
	for {
		// add block to schedule
		b := idToBlock[bid]
		if len(order) > 0 {
			// Check for transition into new loop, if so, attempt to pre-rotate
			if bl, bendl := ln.b2l[b.ID], ln.b2l[order[len(order)-1].ID]; bl != nil &&
				(bendl == nil || bl != bendl && bl.depth >= bendl.depth) {

				bstart := ln.goodtop(bl, b2PO)
				// if f.pass.debug > 0 {
				// 	fmt.Printf("Goodtop %v returns %v\n", b, bstart)
				// }
				if bstart != nil && !scheduled[bstart.ID] {
					b = bstart
					bid = b.ID
				}
			}
		}
		order = append(order, b)
		scheduled[bid] = true
		if len(order) == len(f.Blocks) {
			break
		}

		for _, e := range b.Succs {
			c := e.b
			indegree[c.ID]--
			if indegree[c.ID] == 0 {
				posdegree.remove(c.ID)
				zerodegree.add(c.ID)
			}
		}

		// Pick the next block to schedule
		// Pick among the successor blocks that have not been scheduled yet.

		// Use likely direction if we have it.
		var likely *Block
		switch b.Likely {
		case BranchLikely:
			likely = b.Succs[0].b
		case BranchUnlikely:
			likely = b.Succs[1].b
		}
		if likely != nil && !scheduled[likely.ID] {
			bid = likely.ID
			continue
		}

		// Use degree for now.
		bid = 0
		mindegree := int32(f.NumBlocks())
		for _, e := range order[len(order)-1].Succs {
			c := e.b
			if scheduled[c.ID] || c.Kind == BlockExit {
				continue
			}
			if indegree[c.ID] < mindegree {
				mindegree = indegree[c.ID]
				bid = c.ID
			}
		}
		if bid != 0 {
			continue
		}
		// TODO: improve this part
		// No successor of the previously scheduled block works.
		// Pick a zero-degree block if we can.
		for zerodegree.size() > 0 {
			cid := zerodegree.pop()
			if !scheduled[cid] {
				bid = cid
				continue blockloop
			}
		}
		// Still nothing, pick any non-exit block.
		for posdegree.size() > 0 {
			cid := posdegree.pop()
			if !scheduled[cid] {
				bid = cid
				continue blockloop
			}
		}
		// Pick any exit block.
		// TODO: Order these to minimize jump distances?
		for {
			cid := exit.pop()
			if !scheduled[cid] {
				bid = cid
				continue blockloop
			}
		}
	}
	f.Blocks = order
}

// goodtop attempts to find a good first block for the linearized layout of loop l.
func (ln *loopnest) goodtop(l *loop, b2PO []int32) *Block {
	var dominatesOne, dominatesAll, dOneExit, dAllExit *Block
	f := ln.f
	dbg := ""

	// if f.pass.test == 0 {
	// 	return nil
	// }

	// if !f.DebugTest {
	// 	return nil
	// }

	if f.pass.debug > 0 {
		dbg += "Goodtop "
		dbg += l.header.String()
		dbg += " "
	}

	dom := f.Idom()
blocks:
	for _, b := range l.blocks {
		if len(b.Succs) <= 1 { // TODO have to capture that arch-dep Ifs but exclude defer, fault, etc.
			if f.pass.debug > 0 {
				dbg += "."
			}
			continue
		}
		b0 := b.Succs[0].Block()
		b1 := b.Succs[1].Block()
		l0 := ln.b2l[b0.ID]
		l1 := ln.b2l[b1.ID]
		if l0 == l1 {
			// note one successor must be in loop, so both are in loop if equal.
			if f.pass.debug > 0 {
				dbg += "="
			}
			continue
		}
		if l0 != nil && l1 != nil && (l0 == l || l0.depth > l.depth) && (l1 == l || l1.depth > l.depth) {
			// one or both successors are nested within l
			if f.pass.debug > 0 {
				dbg += ">"
			}
			continue
		}
		bExit := b0
		if l1 != l {
			bExit = b1
		}

		if f.pass.debug > 0 {
			dbg += " bexit="
			dbg += bExit.String()
		}

		// Therefore one successor is an exit.
		if dominatesOne == nil || b2PO[bExit.ID] > b2PO[dOneExit.ID] {
			dominatesOne = b
			dOneExit = bExit
		}
		for _, e := range l.header.Preds {
			be := e.Block()
			if ln.sdom.isAncestorEq(l.header, be) {
				// be is the source of a backedge, does b dominate it?
				if !ln.sdom.isAncestorEq(b, be) {
					continue blocks
				}
			}
		}
		if dominatesAll == nil || b2PO[bExit.ID] > b2PO[dAllExit.ID] {
			dominatesAll = b
			dAllExit = bExit
		}
	}
	if dominatesAll != nil {
		top := dominatesAll.Succs[ln.inLoopSuccessorIndex(l, dominatesAll)].Block()
		if f.pass.debug > 0 {
			fmt.Printf("%s; returns %s\n", dbg, top.String())
		}
		// pick the successor that is in the loop for top of loop.
		return top
	}
	if dominatesOne != nil {
		// look for BlockPlain predecessors of in-loop successor of dominatesOne,
		// iterate on their dominators as long as they do not dominate B,
		// iterate up dominators to B.  Ideally find A such that one successor
		// dominates B, the other dominates the in-loop successor of B (e.g.,
		// the root of a diamond).
		inloop := ln.inLoopSuccessorIndex(l, dominatesOne)
		inloopBlock := dominatesOne.Succs[inloop].Block()
		for _, p := range inloopBlock.Preds {
			pb := p.Block()
			if pb.Kind == BlockPlain {
				// Should always terminate on loop header
				for ib := pb; ib != nil && !ln.sdom.isAncestorEq(ib, dominatesOne); ib = dom[ib.ID] {
					pb = ib
				}
				// This is a little uncertain -- could make branches towards dominatesOne be less likely.
				if f.pass.debug > 0 {
					fmt.Printf("%s; diamond-ish case returns %s\n", dbg, pb.String())
				}
				return pb
			}
		}
		// Could come here for case of A -> (B,C); B -> (X, C); triangle, not diamond.
		if f.pass.debug > 0 {
			fmt.Printf("%s; triangle case returns %s\n", dbg, inloopBlock.String())
		}
		return inloopBlock
	}
	if f.pass.debug > 0 {
		fmt.Printf("%s; nil case\n", dbg)
	}
	return nil
}

func (ln *loopnest) inLoopSuccessorIndex(l *loop, b *Block) int {
	s0 := b.Succs[0].Block()
	l0 := ln.b2l[s0.ID]
	if l0 == nil || l0.depth > l.depth || l0 != l && l0.depth == l.depth {
		return 1
	}
	return 0
}

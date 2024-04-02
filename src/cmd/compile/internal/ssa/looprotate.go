// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

// loopRotate converts loops with a check-loop-condition-at-beginning
// to loops with a check-loop-condition-at-end.
// This helps loops avoid extra unnecessary jumps.
//
//	 loop:
//	   CMPQ ...
//	   JGE exit
//	   ...
//	   JMP loop
//	 exit:
//
//	  JMP entry
//	loop:
//	  ...
//	entry:
//	  CMPQ ...
//	  JLT loop
func loopRotate(f *Func) {
	loopnest := f.loopnest()
	if loopnest.hasIrreducible {
		return
	}
	if len(loopnest.loops) == 0 {
		return
	}

	idToIdx := f.Cache.allocIntSlice(f.NumBlocks())
	defer f.Cache.freeIntSlice(idToIdx)
	for i, b := range f.Blocks {
		idToIdx[b.ID] = i
	}

	// Set of blocks we're moving, by ID.
	move := map[ID]struct{}{}

	// Map from block ID to the moving blocks that should
	// come before before it.
	before := map[ID][]*Block{}

	// Check each loop header and decide if we want to move it.
	for _, loop := range loopnest.loops {
		b := loop.header
		var p *Block // b's in-loop predecessor
		for _, e := range b.Preds {
			if loopnest.b2l[e.b.ID] != loop {
				continue
			}
			if e.b.Kind != BlockPlain {
				continue
			}
			if p == nil || idToIdx[e.b.ID] > idToIdx[p.ID] {
				p = e.b
			}
		}
		if p == nil || p == b {
			continue
		}
		b0 := b
		before[b.ID] = []*Block{}
		for {
			nextIdx := idToIdx[b.ID] + 1
			if nextIdx >= len(f.Blocks) { // reached end of function (maybe impossible?)
				break
			}
			nextb := f.Blocks[nextIdx]
			if loopnest.b2l[nextb.ID] == loop {
				before[b0.ID] = append(before[b0.ID], nextb)
			}
			if nextb == p { // original b will follow its predecessor.
				break
			}
			b = nextb
		}

		for _, b := range before[b0.ID] {
			move[b.ID] = struct{}{}
		}
	}

	// Move blocks to their destinations in a single pass.
	// We rely here on the fact that loop headers must come
	// before the rest of the loop.  And that relies on the
	// fact that we only identify reducible loops.
	j := 0
	// Some blocks that are not part of a loop may be placed
	// between loop blocks. In order to avoid these blocks from
	// being overwritten, use a temporary slice.
	oldOrder := f.Cache.allocBlockSlice(len(f.Blocks))
	defer f.Cache.freeBlockSlice(oldOrder)
	copy(oldOrder, f.Blocks)
	for _, b := range oldOrder {
		if _, ok := move[b.ID]; ok {
			continue
		}
		for _, a := range before[b.ID] {
			f.Blocks[j] = a
			j++
		}
		f.Blocks[j] = b
		j++

	}
	if j != len(oldOrder) {
		f.Fatalf("bad reordering in looprotate")
	}
}

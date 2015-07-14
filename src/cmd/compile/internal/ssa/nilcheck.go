// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

func nilcheckelim(f *Func) {
	idom := dominators(f)
	maxBlockID := f.NumBlocks()
	domTree := make([][]*Block, maxBlockID)

	// Create a block ID -> [dominees] mapping
	for _, b := range f.Blocks {
		if dom := idom[b.ID]; dom != nil {
			domTree[dom.ID] = append(domTree[dom.ID], b)
		}
	}

	// TODO: Eliminate more nil checks.
	// For example, pointers to function arguments
	// and pointers to static values cannot be nil.
	// We could also track pointers constructed by
	// taking the address of another value.
	// We can also recursively remove any chain of
	// fixed offset calculations,
	// i.e. struct fields and array elements,
	// even with non-constant indices:
	// x is non-nil iff x.a.b[i].c is.

	type bp struct {
		block *Block
		ptr   *Value // if non-nil, ptr that is nilcheck'd in the block
	}

	work := make([]*bp, 0, 256)
	work = append(work, &bp{block: f.Entry, ptr: checkedptr(f.Entry)})

	// map from value ID to bool indicating if there is a nil check in the
	// current dominator path being walked
	nilchecks := make([]bool, f.NumValues())
	n := 0

	// perform a depth first walk of the dominator tree
	for len(work) > 0 {
		node := work[len(work)-1]
		work = work[:len(work)-1]

		if node.ptr != nil {
			// this is a marker block to indicate we are traversing
			// back up the dominator tree and need to clear our
			// nilcheck for this pointer
			if node.block == nil {
				nilchecks[node.ptr.ID] = false
				continue
			}

			// otherwise its a nil check we haven't seen before so
			// record it, and add a marker block that will be used
			// to remove it later when we traverse up
			nilchecks[node.ptr.ID] = true
			work = append(work, &bp{block: nil, ptr: node.ptr})
		}

		n++
		for _, w := range domTree[node.block.ID] {
			nb := &bp{block: w, ptr: checkedptr(w)}
			if nb.ptr != nil && nilchecks[nb.ptr.ID] {
				// Eliminate the nil check.
				// The deadcode pass will remove vestigial values,
				// and the fuse pass will join this block with its successor.
				w.Kind = BlockPlain
				w.Control = nil
				f.removePredecessor(w, w.Succs[1])
				w.Succs = w.Succs[:1]

				nb.ptr = nil
			}
			work = append(work, nb)
		}
	}
}

// checkedptr returns the Value, if any,
// that is used in a nil check in b's Control op.
func checkedptr(b *Block) *Value {
	if b.Kind == BlockIf && b.Control.Op == OpIsNonNil {
		return b.Control.Args[0]
	}
	return nil
}

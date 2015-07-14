// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

// Eliminate redundant nil checks.
// A nil check is redundant if the same
// nil check has been performed by a
// dominating block.
// The efficacy of this pass depends
// heavily on the efficacy of the cse pass
func nilcheckelim(f *Func) {
	idom := dominators(f)
	domTree := make([][]*Block, f.NumBlocks())

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

	type blockOp int
	const (
		Work   blockOp = iota // regular work node
		AddPtr                // register the pointer as being nil checked
		DelPtr                // unregister the pointer
	)

	type bp struct {
		block *Block
		ptr   *Value // if non-nil, ptr that is nilcheck'd in the block
		op    blockOp
	}

	work := make([]bp, 0, 256)
	work = append(work, bp{block: f.Entry, ptr: checkedptr(f.Entry)})

	// map from value ID to bool indicating if there is a nil check in the
	// current dominator path being walked
	nilchecks := make([]bool, f.NumValues())

	// perform a depth first walk of the dominator tree
	for len(work) > 0 {
		node := work[len(work)-1]
		work = work[:len(work)-1]

		switch node.op {
		case Work:
			if node.ptr != nil {
				nilchecks[node.ptr.ID] = true
				// register a DelPtr block to clear the ptr from the map
				// of nil checks once we traverse back up the tree
				work = append(work, bp{op: DelPtr, ptr: node.ptr})
			}
		case AddPtr:
			nilchecks[node.ptr.ID] = true
			continue
		case DelPtr:
			nilchecks[node.ptr.ID] = false
			continue
		}

		for _, w := range domTree[node.block.ID] {
			// We are about to traverse down the 'ptr is nil' side
			// of a nilcheck block.
			if node.block.Kind == BlockIf && node.block.Control.Op == OpIsNonNil {
				if w == node.block.Succs[1] {
					work = append(work, bp{op: AddPtr, ptr: node.ptr})
					nilchecks[node.ptr.ID] = false
				}
			}
			nb := bp{block: w, ptr: checkedptr(w)}
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

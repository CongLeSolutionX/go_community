// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

// nilcheckelim eliminates unnecessary nil checks.
func nilcheckelim(f *Func) {
	// A nil check is redundant if the same nil check was successful in a
	// dominating block or the value being tested is constructed from
	// OpAddr.  The efficacy of this pass depends heavily on the efficacy
	// of the cse pass
	idom := dominators(f)
	domTree := make([][]*Block, f.NumBlocks())

	// Create a block ID -> [dominees] mapping
	for _, b := range f.Blocks {
		if dom := idom[b.ID]; dom != nil {
			domTree[dom.ID] = append(domTree[dom.ID], b)
		}
	}

	// TODO: Eliminate more nil checks.
	// We can recursively remove any chain of fixed offset calculations,
	// i.e. struct fields and array elements, even with non-constant
	// indices: x is non-nil iff x.a.b[i].c is.

	type blockOp int
	const (
		Work   blockOp = iota // clear nil check if we should and traverse to dominees regardless
		RecPtr                // record the pointer as being nil checked
		ClearPtr
	)

	type bp struct {
		block *Block // block, or nil in SetPtr/ClearPtr nodes
		ptr   *Value // if non-nil, ptr that is to bet set/cleared in SetPtr/ClearPtr nodes
		op    blockOp
	}

	work := make([]bp, 0, 256)
	work = append(work, bp{block: f.Entry, ptr: checkedptr(f.Entry)})

	// map from value ID to bool indicating if value is known to be non-nil
	// in the current dominator path being walked
	nonNilValues := make([]bool, f.NumValues())

	// perform a depth first walk of the dominator tree
	for len(work) > 0 {
		node := work[len(work)-1]
		work = work[:len(work)-1]

		var pushRecPtr bool
		switch node.op {
		case Work:
			// a value resulting from taking the address of a value
			// implies it is non-nil
			for _, v := range node.block.Values {
				if v.Op == OpAddr {
					// set this immediately instead of
					// using SetPtr so we can potentially
					// remove an OpIsNonNil check in the
					// current work block
					nonNilValues[v.ID] = true
				}
			}

			if node.ptr != nil {
				// already have a nilcheck in the dominator path
				if nonNilValues[node.ptr.ID] {
					// Eliminate the nil check.
					// The deadcode pass will remove vestigial values,
					// and the fuse pass will join this block with its successor.
					node.block.Kind = BlockPlain
					node.block.Control = nil
					f.removePredecessor(node.block, node.block.Succs[1])
					node.block.Succs = node.block.Succs[:1]
				} else {
					// new nilcheck so add a ClearPtr node to clear the
					// ptr from the map of nil checks once we traverse
					// back up the tree
					work = append(work, bp{op: ClearPtr, ptr: node.ptr})
					// and cause a new setPtr to be appended after the
					// nodes dominees
					pushRecPtr = true
				}
			}
		case RecPtr:
			nonNilValues[node.ptr.ID] = true
			continue
		case ClearPtr:
			nonNilValues[node.ptr.ID] = false
			continue
		}

		var falseBranch *Block
		for _, w := range domTree[node.block.ID] {
			// TODO: Since we handle the false side of OpIsNonNil
			// correctly, look into rewriting user nil checks into
			// OpIsNonNil so they can be eliminated also

			// we are about to traverse down the 'ptr is nil' side
			// of a nilcheck block, so save it for later
			if node.block.Kind == BlockIf && node.block.Control.Op == OpIsNonNil &&
				w == node.block.Succs[1] {
				falseBranch = w
				continue
			}
			work = append(work, bp{block: w, ptr: checkedptr(w)})
		}

		if falseBranch != nil {
			// we pop from the back of the work slice, so this sets
			// up the false branch to be operated on before the
			// node.ptr is recorded
			work = append(work, bp{op: RecPtr, ptr: node.ptr})
			work = append(work, bp{block: falseBranch, ptr: checkedptr(falseBranch)})
		} else if pushRecPtr {
			work = append(work, bp{op: RecPtr, ptr: node.ptr})
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

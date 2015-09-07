// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

type sparseTreeBlock struct {
	block                  *Block
	child, sibling, parent *Block

	// Every block has 6 numbers associated with it:
	// entry-1, entry, entry+1, exit-1, and exit, exit+1.
	// entry and exit are conceptually the top of the block (phi functions)
	// entry+1 and exit-1 are conceptually the bottom of the block (ordinary defs)
	// entry-1 and exit+1 are conceptually "just before" the block (conditions flowing in)
	//
	// This simplifies life if we wish to query information about x
	// when x is both an input to and output of a block.
	entry, exit int32
}

// A sparseTree is a tree of Blocks.
// It allows rapid ancestor queries,
// such as whether one block dominates another.
type sparseTree []sparseTreeBlock

// newSparseTree creates a sparseTree from a block-to-parent tree.
func newSparseTree(f *Func, tree []*Block) sparseTree {
	t := make(sparseTree, f.NumBlocks())
	for _, b := range f.Blocks {
		n := t[b.ID]
		n.block = b
		if p := tree[b.ID]; p != nil {
			n.parent = p
			n.sibling = t[p.ID].child
			t[p.ID].child = b
		}
		t[b.ID] = n
	}
	t.numberBlock(f.Entry, 1)
	return t
}

// numberBlock numbers b given entry number n.
// It returns b's child's exit number.
func (t sparseTree) numberBlock(b *Block, n int32) int32 {
	n++ // go from -1 to 0
	t[b.ID].entry = n
	n++
	for c := t[b.ID].child; c != nil; c = t[c.ID].sibling {
		// child entry is one larger than entry+1
		n = t.numberBlock(c, n+1) // n = child exit
	}
	n += 2 // exit-1, exit
	t[b.ID].exit = n
	n++ // exit+1
	return n
}

// isAncestorEq reports whether x is an ancestor of or equal to y.
func (t sparseTree) isAncestorEq(x, y *Block) bool {
	xx := t[x.ID]
	yy := t[y.ID]
	return xx.entry <= yy.entry && yy.exit <= xx.exit
}

// isAncestor reports whether x is a strict ancestor of y.
func (t sparseTree) isAncestor(x, y *Block) bool {
	xx := t[x.ID]
	yy := t[y.ID]
	return xx.entry < yy.entry && yy.exit < xx.exit
}

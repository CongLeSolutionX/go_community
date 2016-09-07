// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

// trim removes blocks with no code in them.
// These blocks were inserted to remove critical edges.
func trim(f *Func) {
	n := 0
	for _, b := range f.Blocks {
		if !trimmableBlock(b) {
			f.Blocks[n] = b
			n++
			continue
		}

		// Splice b out of the graph.
		p, i := b.Preds[0].b, b.Preds[0].i
		s, j := b.Succs[0].b, b.Succs[0].i
		p.Succs[i] = Edge{s, j}
		s.Preds[j] = Edge{p, i}

		for _, e := range b.Preds[1:] {
			p, i := e.b, e.i
			p.Succs[i] = Edge{s, len(s.Preds)}
			s.Preds = append(s.Preds, Edge{p, i})
		}

		// Merge the values into the successor block. The values
		// either correspond to no code (e.g. PHI ops) or `b` is
		// the only predecessor of `s`, thus it does not change
		// program semantics to merge them.
		s.Values = append(b.Values, s.Values...)
	}
	if n < len(f.Blocks) {
		f.invalidateCFG()
		tail := f.Blocks[n:]
		for i := range tail {
			tail[i] = nil
		}
		f.Blocks = f.Blocks[:n]
	}
}

// emptyBlock returns true if the block does not contain actual
// instructions
func emptyBlock(b *Block) bool {
	for _, v := range b.Values {
		if v.Op != OpPhi {
			return false
		}
	}
	return true
}

// trimmableBlock returns true if the block can be trimmed from the CFG,
// subject to the following criteria:
//  - it should not be the first block
//  - it should be BlockPlain
//  - it should not loop back to itself
//  - it either is the single predecessor of the successor block or
//    contains no actual instructions
func trimmableBlock(b *Block) bool {
	if b.Kind != BlockPlain || len(b.Preds) == 0 {
		return false
	}
	s := b.Succs[0].b
	return s != b && (len(s.Preds) == 1 || emptyBlock(b))
}

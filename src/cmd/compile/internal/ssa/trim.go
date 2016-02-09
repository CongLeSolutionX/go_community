// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

// trim removes blocks with no code in them.
// These blocks were inserted to remove critical edges.
func trim(f *Func) {
	for ok := true; ok; {
		ok = false

		i := 0
		for _, b := range f.Blocks {
			// If a control block has the same successors then transform this
			// block into a plain block and remove the jumps.
			if b.Control != nil && b.Control.Type.IsFlags() &&
				len(b.Succs) == 2 && b.Succs[0] == b.Succs[1] {
				b.Kind = BlockPlain
				b.Control = nil
				b.Succs = b.Succs[:1]
				b.Succs[0].removePred(b)
			}
			if b.Kind != BlockPlain || len(b.Values) != 0 || len(b.Preds) != 1 {
				f.Blocks[i] = b
				i++
				continue
			}
			// TODO: handle len(b.Preds)>1 case.

			// Splice b out of the graph.
			ok = true
			pred := b.Preds[0]
			succ := b.Succs[0]
			for j, s := range pred.Succs {
				if s == b {
					pred.Succs[j] = succ
				}
			}
			for j, p := range succ.Preds {
				if p == b {
					succ.Preds[j] = pred
				}
			}
		}
		for j := i; j < len(f.Blocks); j++ {
			f.Blocks[j] = nil
		}
		f.Blocks = f.Blocks[:i]
	}
}

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

const (
	lt uint = 1 << iota
	eq
	gt
)

const (
	signed uint = 1 << iota
	unsigned
)

var (
	comp = map[BlockKind]uint{
		BlockAMD64EQ:  eq,
		BlockAMD64NE:  lt | gt,
		BlockAMD64LT:  lt,
		BlockAMD64LE:  lt | eq,
		BlockAMD64GT:  gt,
		BlockAMD64GE:  eq | gt,
		BlockAMD64ULT: lt,
		BlockAMD64ULE: lt | eq,
		BlockAMD64UGT: gt,
		BlockAMD64UGE: eq | gt,
	}
	// compatibility mask. we can put include the arch here if need.
	sign = map[BlockKind]uint{
		BlockAMD64EQ:  signed | unsigned,
		BlockAMD64NE:  signed | unsigned,
		BlockAMD64LT:  signed,
		BlockAMD64LE:  signed,
		BlockAMD64GT:  signed,
		BlockAMD64GE:  signed,
		BlockAMD64ULT: unsigned,
		BlockAMD64ULE: unsigned,
		BlockAMD64UGT: unsigned,
		BlockAMD64UGE: unsigned,
	}
)

func prove(f *Func) {
	idom := dominators(f)
	sdom := newSparseTree(f, idom)

	for _, b := range f.Blocks {
		if idom[b.ID] == nil {
			continue // unreachable
		}
		maskb := comp[b.Kind]
		if maskb == 0 {
			continue // not a control block
		}

		succ := -1 // which successor is always taken
		all := lt | eq | gt
		mask := all // restrictions comming from ancestors
		stop := false

		for p := idom[b.ID]; p != nil && !stop; p = idom[p.ID] {
			if p == b.Control.Block {
				// The control is defined in this block, no ancestor
				// can branch on it. We still have to run the logic below.
				stop = true
			}
			if p.Control != b.Control {
				continue
			}
			if sign[b.Kind]&sign[p.Kind] == 0 {
				continue
			}
			maskp := comp[p.Kind]
			if maskp == 0 {
				continue
			}

			// If p and p.Succs[0] are dominators it means that every path
			// from entry to b passes through p and p.Succs[0]. We care that
			// no path from entry to b passes through p.Succs[1]. If p.Succs[0]
			// has one predecessor then (apart from the degenerate case),
			// there is no path from entry that can reach b through p.Succs[1].
			// TODO: how about p->yes->b->yes, i.e. a loop in yes.
			yes := sdom.isAncestorEq(p.Succs[0], b) && len(p.Succs[0].Preds) == 1
			if yes {
				mask &= maskp
			}
			// The above comment applies here, too.
			no := sdom.isAncestorEq(p.Succs[1], b) && len(p.Succs[1].Preds) == 1
			if no {
				mask &= all ^ maskp
			}
			if yes && no {
				b.Fatalf("block %s came from both branches of %s", b, p)
			}

			// p and b have the same control block and test compatible signs
			if maskb&mask == mask {
				b.Logf("(%s.%s) proved positive branch of %s from %s in %s\n", b.Kind, p.Kind, b, p, f.Name)
				succ = 0
				break
			}
			if (all^maskb)&mask == mask {
				b.Logf("(%s.%s) proved negative branch of %s from %s in %s\n", b.Kind, p.Kind, b, p, f.Name)
				succ = 1
				break
			}
		}

		if succ == 0 {
			b.Kind = BlockFirst
			b.Control = nil
		} else if succ == 1 {
			b.Kind = BlockFirst
			b.Control = nil
			b.Succs[0], b.Succs[1] = b.Succs[1], b.Succs[0]
		}
	}
}

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
	pointer
)

var (
	rangeMask = map[Op]uint{
		OpEq8:   eq,
		OpEq16:  eq,
		OpEq32:  eq,
		OpEq64:  eq,
		OpEqPtr: eq,

		OpNeq8:   lt | gt,
		OpNeq16:  lt | gt,
		OpNeq32:  lt | gt,
		OpNeq64:  lt | gt,
		OpNeqPtr: lt | gt,

		OpLess8:   lt,
		OpLess8U:  lt,
		OpLess16:  lt,
		OpLess16U: lt,
		OpLess32:  lt,
		OpLess32U: lt,
		OpLess64:  lt,
		OpLess64U: lt,

		OpLeq8:   lt | eq,
		OpLeq8U:  lt | eq,
		OpLeq16:  lt | eq,
		OpLeq16U: lt | eq,
		OpLeq32:  lt | eq,
		OpLeq32U: lt | eq,
		OpLeq64:  lt | eq,
		OpLeq64U: lt | eq,

		OpGeq8:   eq | gt,
		OpGeq8U:  eq | gt,
		OpGeq16:  eq | gt,
		OpGeq16U: eq | gt,
		OpGeq32:  eq | gt,
		OpGeq32U: eq | gt,
		OpGeq64:  eq | gt,
		OpGeq64U: eq | gt,

		OpGreater8:   gt,
		OpGreater8U:  gt,
		OpGreater16:  gt,
		OpGreater16U: gt,
		OpGreater32:  gt,
		OpGreater32U: gt,
		OpGreater64:  gt,
		OpGreater64U: gt,

		OpIsInBounds:      lt,
		OpIsSliceInBounds: lt | eq,
	}

	// type compatibility mask.
	typeMask = map[Op]uint{
		OpEq8:   signed | unsigned,
		OpEq16:  signed | unsigned,
		OpEq32:  signed | unsigned,
		OpEq64:  signed | unsigned,
		OpEqPtr: pointer,

		OpNeq8:   signed | unsigned,
		OpNeq16:  signed | unsigned,
		OpNeq32:  signed | unsigned,
		OpNeq64:  signed | unsigned,
		OpNeqPtr: pointer,

		OpLess8:   signed,
		OpLess8U:  unsigned,
		OpLess16:  signed,
		OpLess16U: unsigned,
		OpLess32:  signed,
		OpLess32U: unsigned,
		OpLess64:  signed,
		OpLess64U: unsigned,

		OpLeq8:   signed,
		OpLeq8U:  unsigned,
		OpLeq16:  signed,
		OpLeq16U: unsigned,
		OpLeq32:  signed,
		OpLeq32U: unsigned,
		OpLeq64:  signed,
		OpLeq64U: unsigned,

		OpGeq8:   signed,
		OpGeq8U:  unsigned,
		OpGeq16:  signed,
		OpGeq16U: unsigned,
		OpGeq32:  signed,
		OpGeq32U: unsigned,
		OpGeq64:  signed,
		OpGeq64U: unsigned,

		OpGreater8:   signed,
		OpGreater8U:  unsigned,
		OpGreater16:  signed,
		OpGreater16U: unsigned,
		OpGreater32:  signed,
		OpGreater32U: unsigned,
		OpGreater64:  signed,
		OpGreater64U: unsigned,

		OpIsInBounds:      unsigned,
		OpIsSliceInBounds: unsigned,
	}
)

func prove(f *Func) {
	idom := dominators(f)
	sdom := newSparseTree(f, idom)

	for _, b := range f.Blocks {
		if b.Kind != BlockIf {
			continue
		}
		if idom[b.ID] == nil { // entry block
			continue
		}

		maskb := rangeMask[b.Control.Op]
		if maskb == 0 {
			continue // not a control block
		}

		succ := -1 // which successor is always taken
		all := lt | eq | gt
		mask := all // restrictions comming from ancestors

		for p := idom[b.ID]; p != nil; p = idom[p.ID] {
			if p.Kind != BlockIf { // not a branch
				continue
			}
			maskp := rangeMask[p.Control.Op]
			if maskp == 0 {
				continue
			}
			if p.Control.Args[0] != b.Control.Args[0] || p.Control.Args[1] != b.Control.Args[1] {
				continue
			}
			if typeMask[p.Control.Op]&typeMask[b.Control.Op] == 0 {
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

		if succ != -1 {
			b.Kind = BlockFirst
			b.Control = nil
			b.Succs[0], b.Succs[1] = b.Succs[succ], b.Succs[1-succ]
		}
	}
}

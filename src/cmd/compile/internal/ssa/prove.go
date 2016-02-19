// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

// import "fmt"

type rangeMask uint
type typeMask uint

const (
	lt rangeMask = 1 << iota
	eq
	gt
)

const (
	signed typeMask = 1 << iota
	unsigned
	pointer
)

var (
	reverseBits = [...]rangeMask{0, 4, 2, 6, 1, 5, 3, 7}

	// maps what we learn when the positive branch is taken.
	rangeMaskTable = map[Op]rangeMask{
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

		// TODO: OpIsInBounds actually test 0 <= a < b. This means
		// that the positive branch learns signed/LT and unsigned/LT
		// but the negative branch only learns unsigned/GE.
		OpIsInBounds:      lt,
		OpIsSliceInBounds: lt | eq,
	}

	// type compatibility mask.
	typeMaskTable = map[Op]typeMask{
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
	const full = lt | eq | gt

	idom := dominators(f)
	sdom := newSparseTree(f, idom)
	domTree := make([][]*Block, f.NumBlocks())

	// Create a block ID -> [dominees] mapping
	for _, b := range f.Blocks {
		if dom := idom[b.ID]; dom != nil {
			domTree[dom.ID] = append(domTree[dom.ID], b)
		}
	}

	// current node state
	type walkState int
	const (
		descend walkState = iota
		simplify
	)
	type typeRange struct {
		t typeMask
		r rangeMask
	}
	// work maintains the DFS stack.
	type bp struct {
		block *Block      // current handled block
		state walkState   // what's to do
		old   []typeRange // save previous map entries modified by node
	}
	work := make([]bp, 0, 256)
	work = append(work, bp{
		block: f.Entry,
		state: descend,
	})

	// mask keep tracks of restrictions for each pair of values in
	// the dominators for the the current node.
	// Invariant: a0.ID <= a1.ID
	type control struct {
		tm     typeMask
		a0, a1 *Value
	}
	mask := make(map[control]rangeMask)

	// DSF on the dominator tree.
	for len(work) > 0 {
		node := work[len(work)-1]
		work = work[:len(work)-1]

		switch node.state {
		case descend:
			parent := idom[node.block.ID]
			if parent != nil && parent.Kind == BlockIf && rangeMaskTable[parent.Control.Op] != 0 {
				op := parent.Control.Op
				rm := rangeMask(0)
				ok := false

				// If p and p.Succs[0] are dominators it means that every path
				// from entry to b passes through p and p.Succs[0]. We care that
				// no path from entry to b passes through p.Succs[1]. If p.Succs[0]
				// has one predecessor then (apart from the degenerate case),
				// there is no path from entry that can reach b through p.Succs[1].
				// TODO: how about p->yes->b->yes, i.e. a loop in yes.
				if sdom.isAncestorEq(parent.Succs[0], node.block) && len(parent.Succs[0].Preds) == 1 {
					ok, rm = true, rangeMaskTable[op]
				} else if sdom.isAncestorEq(parent.Succs[1], node.block) && len(parent.Succs[1].Preds) == 1 {
					ok, rm = true, full^rangeMaskTable[op]
				}

				if ok { // comes from a restricting branch
					// parent modifies the restrictions for (a0, a1).
					// saves the previous state.
					a0 := parent.Control.Args[0]
					a1 := parent.Control.Args[1]
					if a0.ID > a1.ID {
						rm = reverseBits[rm]
						a0, a1 = a1, a0
					}

					/*
						if a0.ID == a1.ID && op != OpIsSliceInBounds{
							fmt.Println(a0, a1, op, parent, f.Name)
							panic("equal")
						}
					*/

					tm := typeMaskTable[op]
					for t := typeMask(1); t <= tm; t <<= 1 {
						if t&tm == 0 {
							continue
						}

						i := control{t, a0, a1}
						oldRange, ok := mask[i]
						if !ok {
							if a1 != a0 {
								oldRange = full
							} else { // sometimes happens after cse
								oldRange = eq
							}
						}
						// if i was not already in the map we save the full range
						// so that when we restore it we properly keep track of it.
						node.old = append(node.old, typeRange{t, oldRange})
						mask[i] = oldRange & rm
					}

				}
			}

			// Simplify node after children have been updated.
			// TODO: simplify before? and save some work
			work = append(work, bp{
				block: node.block,
				state: simplify,
				old:   node.old,
			})

			for _, s := range domTree[node.block.ID] {
				work = append(work, bp{
					block: s,
					state: descend,
				})
			}

		case simplify:
			if node.block.Kind == BlockIf {
				op := node.block.Control.Op
				rm := rangeMaskTable[op]
				if rm != 0 {
					succ := -1
					a0 := node.block.Control.Args[0]
					a1 := node.block.Control.Args[1]
					if a0.ID > a1.ID {
						rm = reverseBits[rm]
						a0, a1 = a1, a0
					}

					tm := typeMaskTable[op]
					for t := typeMask(1); t <= tm; t <<= 1 {
						if t&tm == 0 {
							continue
						}

						i := control{t, a0, a1}
						m, has := mask[i]
						if has && rm&m == m {
							f.Logf("proved positive branch of %s, block %s in %s\n",
								node.block.Control, node.block, f.Name)
							succ = 0
							break
						}
						if has && (full^rm)&m == m {
							f.Logf("proved negative branch of %s, %s block in %s\n",
								node.block.Control, node.block, f.Name)
							succ = 1
							break
						}
					}

					if succ != -1 {
						b := node.block
						b.Kind = BlockFirst
						b.Control = nil
						b.Succs[0], b.Succs[1] = b.Succs[succ], b.Succs[1-succ]
					}
				}
			}

			// restores the previous type mask before we ascend to the parent.
			parent := idom[node.block.ID]
			if parent != nil && parent.Kind == BlockIf && rangeMaskTable[parent.Control.Op] != 0 {
				a0 := parent.Control.Args[0]
				a1 := parent.Control.Args[1]
				if a0.ID > a1.ID {
					a0, a1 = a1, a0
				}

				for _, tr := range node.old {
					i := control{tr.t, a0, a1}
					if tr.r != full {
						mask[i] = tr.r
					} else {
						delete(mask, i)
					}
				}
			}
		}
	}
}

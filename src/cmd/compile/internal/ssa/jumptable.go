// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

import (
	"cmd/compile/internal/base"
	"cmd/internal/src"
	"fmt"
	"math"
	"sort"
)

// We look for a group of blocks like this:
//
// b1 ----------> b2 --> t1
// |              |
// v              v
// b3 --> t2      t3
// |
// v
// t4
//
// Where all the b's control values are comparisons of the same value to
// different constants. Also:
//  - None of the b's, except b1, can have any side effects.
//    For now, we forbid b2+ from having any code at all, except the comparison test.
//  - The constants must be fairly dense.
//
// This pattern matching handles both linear and binary search code generated
// by the switch rewriting done in ../walk/switch.go.
//
// We convert that structure to a jump table.
//
//    idx := swiched_on_value - min(constants)
//    if idx < 0 || idx > max(constants)-min(constants) { goto default }
//    jump to table[idx]
//
// table[i] contains the t that is branched to when the switched-on value
// is equal to i+min(constants). Unmatched table entries are filled with default.
func jumpTable(f *Func) {
	switch f.Config.arch {
	default:
		// Most architectures can't do this (yet).
		return
	case "amd64":
	}

	// Find all blocks that do constant comparisons against the same value
	// as their parent.
	parent := map[*Block]*Block{}
	for _, b := range f.Blocks {
		if b.Kind != BlockIf {
			continue
		}
		c := b.ControlValue(0)

		// The only op in the block can be the control value.
		if len(b.Values) != 1 {
			continue
		}
		if b.Values[0] != c {
			continue
		}
		if c.Op != OpEq64 && c.Op != OpNeq64 && c.Op != OpLeq64 && c.Op != OpLess64 { // TODO: Unsigned, {32,16,8}, and maybe String.
			continue
		}
		if c.Uses != 1 {
			continue
		}
		x, y := c.Args[0], c.Args[1]
		if x.Op == OpConst64 {
			if y.Op == OpConst64 {
				// Can't handle const/const comparisons because later
				// parts of this pass can't tell which is the variable
				// and which is the constant.
				continue
			}
			x, y = y, x
		} else if y.Op != OpConst64 {
			continue
		}
		value := x

		// Look at predecessor block, see if it is branching on the same value.
		if len(b.Preds) != 1 {
			continue
		}
		p := b.Preds[0].b
		if p.Kind != BlockIf {
			continue
		}
		// Note: don't need len(p.Values)==1.
		c = p.ControlValue(0)
		if c.Op != OpEq64 && c.Op != OpNeq64 && c.Op != OpLeq64 && c.Op != OpLess64 {
			continue
		}
		// Note: don't need c.Uses == 1
		x, y = c.Args[0], c.Args[1]
		if x.Op == OpConst64 {
			if y.Op == OpConst64 {
				continue
			}
			x, y = y, x
		} else if y.Op != OpConst64 {
			continue
		}
		if x != value {
			// Both blocks must be comparing the same value against a constant.
			continue
		}
		if p.Likely != BranchUnknown {
			// Don't hide likely branch info by using a jump table.
			continue
		}

		parent[b] = p
	}

	// Compress parent links to point directly to the root.
	for {
		changed := false
		for b, p := range parent {
			gp := parent[p]
			if gp != nil {
				parent[b] = gp
				changed = true
			}
		}
		if !changed {
			break
		}
	}

	// Make map from root to all the blocks in that group.
	groups := map[*Block][]*Block{}
	for b, p := range parent {
		groups[p] = append(groups[p], b)
	}

	roots := make([]*Block, 0, len(groups))
	for root, group := range groups {
		groups[root] = append(groups[root], root) // add root to its own group
		sort.Slice(group, func(i, j int) bool {   // groups sort for determinism
			return group[i].ID < group[j].ID
		})
		roots = append(roots, root)
	}
	// Sort groups by root ID for determinism.
	sort.Slice(roots, func(i, j int) bool { // sort roots for determinism
		return roots[i].ID < roots[j].ID
	})

	// getConst returns the constant compared against by a block.
	getConst := func(b *Block) int64 {
		v := b.ControlValue(0)
		if v.Args[0].Op == OpConst64 {
			return v.Args[0].AuxInt
		}
		return v.Args[1].AuxInt
	}
	// getVal returns the value that we're switching on.
	getVal := func(b *Block) *Value {
		v := b.ControlValue(0)
		if v.Args[0].Op == OpConst64 {
			return v.Args[1]
		}
		return v.Args[0]
	}

	// nextBlock returns the edge outgoing from b if the choice
	// variable is equal to c.
	nextBlock := func(b *Block, c int64) Edge {
		d := getConst(b)
		var r bool
		switch b.ControlValue(0).Op {
		case OpEq64:
			r = c == d
		case OpNeq64:
			r = c != d
		case OpLess64:
			r = c < d
		case OpLeq64:
			r = c <= d
		default:
			// TODO: unsigned, <64 bit
			f.Fatalf("unknown op %s", b.ControlValue(0).Op)
		}
		if r {
			return b.Succs[0]
		}
		return b.Succs[1]
	}

	// This is the main loop, processing a group at a time.
grouploop:
	for _, root := range roots {
		group := groups[root]
		if f.pass.debug > 0 {
			fmt.Printf("%s: processing root=%s group=%v\n", f.Name, root, group)
		}

		// TODO: keep track of more than min/max.
		// For now, just min/max in signed domain. Later we can do
		// unsigned domain, mask low bits, etc.
		// Note: the constants are always int64, even if in the original
		// source code they were uint64, or int16, or whatever.

		// Figure out the range of constants we compare against.
		var min, max int64
		var cnt int
		for _, b := range group {
			c := getConst(b)
			if cnt == 0 {
				min = c
				max = c
			} else {
				if c < min {
					min = c
				}
				if c > max {
					max = c
				}
			}
			cnt++ // TODO: just include ==,!= or should we also count <,<= here?
		}
		if min == math.MinInt64 || max == math.MaxInt64 {
			// We use these as sentinel values later. Abort if we compare against them.
			continue
		}
		width := uint64(max - min + 1) // number of jump table entries

		// Check to see if a jump table is appropriate.
		if cnt < 4 {
			// Not wide enough to make this worthwhile.
			if f.pass.debug > 0 {
				fmt.Printf("  abort: only %d equality tests\n", cnt)
			}
			continue
		}
		if width/4 > uint64(cnt) { // <25% full. TODO: what's the right number here?
			if f.pass.debug > 0 {
				fmt.Printf("  abort: density too small, %d out of %d\n", cnt, width)
			}
			continue
		}

		// groupExit returns the edge that exits the group if the
		// choice variable is c.
		groupExit := func(c int64) Edge {
			b := root
			for {
				e := nextBlock(b, c)
				if parent[e.b] != root {
					// Note: branching back to the root is treated
					// as leaving the group.
					return e
				}
				b = e.b
			}

		}

		// Find the places where we'll exit the group if the choice variable is
		// not in [min,max]. We can do that with two probes, minint and maxint.
		default_ := groupExit(math.MinInt64)
		default2 := groupExit(math.MaxInt64)

		// Skip ahead, to help detect when default_ and default2 are identical.
		for default_.b.Kind == BlockPlain && len(default_.b.Values) == 0 {
			default_ = default_.b.Succs[0]
		}
		for default2.b.Kind == BlockPlain && len(default2.b.Values) == 0 {
			default2 = default2.b.Succs[0]
		}

		// Check that <min and >max defaults are the same.
		if default_.b != default2.b {
			if f.pass.debug > 0 {
				fmt.Printf("  abort: default exits %s and %s are different\n", default_.b, default2.b)
			}
			continue
		}
		// Check for phi compatibility between the two default edges.
		for _, v := range default_.b.Values {
			if v.Op != OpPhi {
				continue
			}
			if v.Args[default_.i] != v.Args[default2.i] {
				// The two edges to the default block induce a different value
				// of some phi op. Abandon the group.
				if f.pass.debug > 0 {
					fmt.Printf("  abort: default exits have different phi args: %v %d and %d\n", v, default_.i, default2.i)
				}
				continue grouploop
			}
		}
		// TODO: we could use 2 different compare/branch in the bcb block below
		// if the <min and >max destinations are different.
		// (In that case we could do CMP X, $min; BLT default1; CMP X $max; BGT default2; LEAQ ?, JT; JMP (-8*min)(JT)(X*8))

		//////////////////////////////////////////////////////////////
		// At this point, we're committed to making the jump table. //
		//////////////////////////////////////////////////////////////

		// Build the jump table block itself.
		jump := f.NewBlock(BlockJumpTable)
		jump.Pos = root.Pos
		// Add outgoing edges for each value in the table.
		for c := min; c <= max; c++ {
			e := groupExit(c)
			if f.pass.debug > 0 {
				fmt.Printf("  %d -> %s[%d]\n", c, e.b, e.i)
			}
			jump.Succs = append(jump.Succs, Edge{b: e.b, i: len(e.b.Preds)})
			e.b.Preds = append(e.b.Preds, Edge{b: jump, i: len(jump.Succs) - 1})
			for _, v := range e.b.Values {
				if v.Op == OpPhi {
					// Use the same phi argument for this edge as the
					// original edge from the group to this block.
					v.AddArg(v.Args[e.i])
				}
			}
		}
		if f.pass.debug > 0 {
			fmt.Printf("  default_ -> %s[%d]\n", default_.b, default_.i)
		}

		// Build bounds check block.
		// TODO: this pass is after prove, so if this comparison is obviously satisifiable (e.g. switch (x&3) { case 0: ... case 3: ... }).
		// We might want to squash this bounds check. Or move this pass before prove.
		bcb := f.NewBlock(BlockIf)
		val := getVal(root)
		minVal := f.Entry.NewValue0I(src.NoXPos, OpConst64, f.Config.Types.UInt64, min)
		widthVal := f.Entry.NewValue0I(src.NoXPos, OpConst64, f.Config.Types.UInt64, int64(width))
		idx := bcb.NewValue2(root.Pos, OpSub64, f.Config.Types.UInt64, val, minVal)
		cmp := bcb.NewValue2(root.Pos, OpLess64U, f.Config.Types.Bool, idx, widthVal)
		bcb.SetControl(cmp)
		// bcb's true branch goes to the jump block.
		bcb.Succs = append(bcb.Succs, Edge{b: jump, i: 0})
		jump.Preds = append(jump.Preds, Edge{b: bcb, i: 0})
		bcb.Likely = BranchLikely // TODO: assumes missing the table entirely is unlikely. True?
		// bcb's false branch goes to the default block.
		bcb.Succs = append(bcb.Succs, Edge{b: default_.b, i: len(default_.b.Preds)})
		default_.b.Preds = append(default_.b.Preds, Edge{b: bcb, i: 1})
		for _, v := range default_.b.Values {
			if v.Op == OpPhi {
				v.AddArg(v.Args[default_.i])
			}
		}

		// The jump block uses the in-bounds index as its control value.
		if base.Flag.Cfg.SpectreIndex {
			idx = jump.NewValue2(root.Pos, OpSpectreIndex, f.Config.Types.UInt64, idx, widthVal)
		}
		jump.SetControl(idx)

		// Modify the original root to unconditionally branch to the bounds check block.
		// One of root's successors is guaranteed to be in the group. That successor block
		// is easy to remove an edge from, because we know it has exactly 1 predecessor.
		if parent[root.Succs[0].b] != root {
			root.swapSuccessors()
		}
		// Always go to the bcb block.
		root.Succs[0].b.Preds = root.Succs[0].b.Preds[:0]  // remove incoming edge to root.Succs[0]
		root.Succs[0] = Edge{b: bcb, i: 0}                 // add outgoing edge to bcb
		bcb.Preds = append(bcb.Preds, Edge{b: root, i: 0}) // add incoming edge from root
		root.Reset(BlockFirst)                             // we don't go to root.Succs[1] either

		// At this point, the whole group except the root should be dead code, and the next deadcode
		// pass will remove it all.

		f.invalidateCFG()
	}
}

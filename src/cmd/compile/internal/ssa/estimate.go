// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

import "fmt"

type cost struct {
	spills, reloads, copies,
	unconditionalBranches int32
}

func (c *cost) isZero() bool {
	return c.spills == 0 && c.reloads == 0 && c.copies == 0 && c.unconditionalBranches == 0
}

// estimate attempts to estimate the (excess) costs of f and
// logs the estimate.  Format is a sequence of 4-tuples,
// #spills, #reloads, #copies, #unconditional
// The first tuple counts reschedule blocks (very rarely executed),
// tuple N aftewards shows the counts for loop nesting depth N-1.
func estimate(f *Func) {
	ln := loopnestfor(f)
	ln.calculateDepths()

	var costs [64]cost

blocks:
	for i, b := range f.Blocks {
		d := 0
		if !b.RarelyRun {
			l := ln.b2l[b.ID]
			d = 1
			if l != nil {
				d = int(l.depth) + 1
				if d >= len(costs) {
					d = len(costs) - 1
				}
			}
		}
		c := &costs[d]
		for _, v := range b.Values {
			switch v.Op {
			case OpStoreReg:
				c.spills++
			case OpLoadReg:
				c.reloads++
			case OpCopy:
				c.copies++
			}
		}
		if len(b.Succs) == 0 {
			continue
		}
		if i < len(f.Blocks)-1 {
			// check that next block is some successor of this block
			n := f.Blocks[i+1]
			for _, s := range b.Succs {
				if s.Block() == n {
					continue blocks
				}
			}
			c.unconditionalBranches++
		}
	}
	last := -1
	for i, c := range costs {
		if !c.isZero() {
			last = i
		}
	}
	s := "ESTIMATE,\"" + f.Name + "\""
	for i := 0; i <= last; i++ {
		c := &costs[i]
		s += fmt.Sprintf(",%d,%d,%d,%d", c.spills, c.reloads, c.copies, c.unconditionalBranches)
	}
	f.Warnl(f.Entry.Pos, s)
}

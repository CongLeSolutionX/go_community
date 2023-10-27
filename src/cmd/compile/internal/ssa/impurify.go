// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

import (
	"cmd/compile/internal/types"
	"cmd/internal/src"
	"fmt"
)

// impurifyCalls converts "const" calls that take/produce no memory operand
// and "pure" calls that produe no memory operand to static calls that do.
func impurifyCalls(f *Func) {
	debug := f.pass.debug
	po := f.postorder()

	// Convert const/pure functions to regular functions, rewriting mem as necessary.
	lastMems := make([]*Value, f.NumBlocks())  // for a block, what is the last visible mem?
	firstMems := make([]*Value, f.NumBlocks()) // for a block, what is the Preds[0] mem?
	memPhi := make([]*Value, f.NumBlocks())    // for a block, what is its mem phi function (if any)
	visited := make([]bool, f.NumBlocks())
	anyChanged := false

	// Topological sort stuff
	dependsOn := make([]int, f.NumValues())
	dependents := make([][]*Value, f.NumValues())
	var zeroDeps []*Value
	var zeroMemDeps []*Value
	var zeroMemDepsCall []*Value
	var lastMem *Value

	// Sort values into dependence/store order
	// Keep track of mem start and end values, and memory phi functions for blocks.
	// TODO optimize later if any of this is a performance problem.
	for j := len(po) - 1; j >= 0; j-- {
		b := po[j]
		anyChanged = false
		lastMem = nil

		if debug > 0 {
			fmt.Printf("PO visit b%d\n", b.ID)
		}

		// Figure out mem exposed at top of block, if any.
		if l := len(b.Preds); l > 0 {
			// nil if there is a difference
			for _, p := range b.Preds {
				pbid := p.Block().ID
				if debug > 0 {
					fmt.Printf("\tpbid=%d, visited=%v, lastMems[]=%s\n", pbid, visited[pbid], lastMems[pbid].LongString())
				}
				if visited[pbid] {
					if lastMem == nil {
						lastMem = lastMems[pbid]
						firstMems[b.ID] = lastMem // Could be nil, only matters for single predecessor case.
					} else if lastMems[pbid] != lastMem {
						lastMem = nil
						break
					}
				}
			}
		}

		if debug > 0 {
			fmt.Printf("\tlastMem=%s\n", lastMem.LongString())
		}

		visited[b.ID] = true

		zeroDeps = zeroDeps[:0]
		zeroMemDeps = zeroMemDeps[:0]
		zeroMemDepsCall = zeroMemDepsCall[:0]

		isCallOp := func(o Op) bool {
			return o == OpStaticLECall || o == OpClosureLECall || o == OpInterLECall || o == OpTailLECall || o == OpConstLECall || o == OpPureLECall
		}

		// Append v to the proper list of depends-on-nothing.
		appendZeroDeps := func(v *Value) {
			if v.Type != types.TypeMem && v.Op != OpConstLECall && v.Op != OpPureLECall {
				zeroDeps = append(zeroDeps, v)
				return
			}

			if isCallOp(v.Op) {
				zeroMemDepsCall = append(zeroMemDepsCall, v)
				return
			}
			zeroMemDeps = append(zeroMemDeps, v)

		}

		// Set up for topological order, note memory-typed Phi and InitMem
		for _, v := range b.Values {
			switch v.Op {
			case OpPhi:
				if v.Type == types.TypeMem {
					memPhi[b.ID] = v
					lastMem = v
				}
				zeroDeps = append(zeroDeps, v) // Phis come first, even the memory phis.
				continue

			case OpInitMem:
				lastMem = v
				zeroMemDeps = append(zeroMemDeps, v)
				continue
			}
			d := 0
			for _, a := range v.Args {
				if a.Block != b {
					continue
				}
				dependents[a.ID] = append(dependents[a.ID], v)
				d++
			}
			if d == 0 {
				appendZeroDeps(v)
			} else {
				dependsOn[v.ID] = d
			}
		}

		changed := false
		b.Values = b.Values[:0]

		if lastMem == nil { // no mem phi function, no InitMem, predecessors have varying mems, a phi is necessary.
			// Create a new memory phi; if there are any backedge inputs, use self for their value; that is both a marker and a best-guess.
			changed = true // if there is not a later mem output, this one will appear at the end and it is new.
			newPhi := b.NewValue0(src.NoXPos, OpPhi, types.TypeMem)
			for _, p := range b.Preds {
				pbid := p.Block().ID
				m := lastMems[pbid]
				if m == nil {
					m = newPhi
				}
				newPhi.AddArg(m)
			}
			if debug > 0 {
				fmt.Printf("\tb%d, lastMem == nil adding mem phi %s\n", b.ID, newPhi.LongString())
			}

			memPhi[b.ID] = newPhi
			lastMem = newPhi
		}

		// TODO use slices.Reverse(zeroDeps) when the bootstrap compiler supports it.
		reverse := func(zd []*Value) {
			l := len(zd)
			for i := 0; i < l/2; i++ {
				zd[i], zd[l-1-i] = zd[l-1-i], zd[i]
			}
		}
		reverse(zeroDeps)
		reverse(zeroMemDeps)
		reverse(zeroMemDepsCall)

		getOneZeroDep := func(p *[]*Value) *Value {
			s := *p
			if l := len(s); l > 0 {
				z := s[l-1]
				*p = s[:l-1]
				b.Values = append(b.Values, z)
				return z
			}
			return nil
		}

		getZeroDep := func() *Value {
			// Don't release another mem until everything depending on the previous mem has been scheduled.
			// Because of updates to other mems, getting this wrong will change memory operands and order.
			if v := getOneZeroDep(&zeroDeps); v != nil {
				return v
			}
			if v := getOneZeroDep(&zeroMemDeps); v != nil {
				return v
			}
			return getOneZeroDep(&zeroMemDepsCall)
		}

		// run a topological order on the values, and rewrite const/pure calls as the sort goes by.
		for z := getZeroDep(); z != nil; z = getZeroDep() {

			if z.Op != OpPhi {
				// If z has a memory operand and it is not lastMem, change it.
				if m := z.MemoryArg(); m != nil && m != lastMem && m.Type == types.TypeMem {
					z.SetArg(len(z.Args)-1, lastMem)
				}
			}

			if z.Type == types.TypeMem {
				changed = false // this output hides any new ones preceding it.
				lastMem = z
			}

			for _, v := range dependents[z.ID] {
				i := v.ID
				dependsOn[i]--
				if dependsOn[i] == 0 {
					appendZeroDeps(v)
				}
			}
			// Do this rewrite after updating dependsOn/zeroDeps
			if isConst := z.Op == OpConstLECall; isConst || z.Op == OpPureLECall {
				// Translate into OpStaticLECall and insert an OpSelectN to extract the new memory.
				z.Op = OpStaticLECall
				if isConst { // pure calls have the arg already.
					z.AddArg(lastMem)
				}
				auxCall := z.Aux.(*AuxCall)
				nresults := auxCall.NResults()
				z.Type = auxCall.LateExpansionResultType()
				lastMem = b.NewValue1I(z.Pos.WithNotStmt(), OpSelectN, types.TypeMem, nresults, z)
				changed = true
			}

			// TODO slice reuse for dependents is a likely optimization
			// possible hack -- use block value index instead of *Value/v.ID
			dependents[z.ID] = nil
		}

		if debug > 0 {
			fmt.Printf("\tb%d, lastMem=%s, changed=%v\n", b.ID, lastMem.LongString(), changed)
		}
		lastMems[b.ID] = lastMem
		if changed {
			anyChanged = true
		}
	}

	// Perhaps we could be clever-er about propagating change, but first, need to see if it is useful.
	for anyChanged {
		if debug > 0 {
			fmt.Printf("anyChanged loop\n")
		}
		anyChanged = false
		for j := len(po) - 1; j >= 0; j-- {
			b := po[j]
			l := len(b.Preds)
			if l == 0 {
				continue
			}

			lastMem := lastMems[b.Preds[0].Block().ID]

			// First figure out if the memory flowing into the non-phi part of the block has changed,
			// and/or if a new phi function is required.
			if l == 1 {
				if lastMem == firstMems[b.ID] {
					// no change
					continue
				}
				firstMems[b.ID] = lastMem
				if lastMems[b.ID].Block != b { // it is a flow-through block, successors will see the change.
					anyChanged = true
					lastMems[b.ID] = lastMem
				}
			} else if l > 1 {
				if p := memPhi[b.ID]; p != nil {
					for i, v := range p.Args {
						lm := lastMems[b.Preds[i].Block().ID]
						if v != lm {
							p.SetArg(i, lm)
						}
					}
					continue
				}

				// there's no phi function here, be sure that is the right choice.
				// nil if there is a difference
				for _, p := range b.Preds[1:] {
					pbid := p.Block().ID
					if lastMems[pbid] != lastMem {
						lastMem = nil
						break
					}
				}
				if lastMem != nil {
					continue
				}
				// need to insert a new phi function and correct memory inputs.
				newPhi := b.NewValue0(src.NoXPos, OpPhi, types.TypeMem)
				if lastMems[b.ID].Block != b { // it was a flow-through block and successors will see the change.
					anyChanged = true
					lastMems[b.ID] = newPhi
				}
				for _, p := range b.Preds {
					newPhi.AddArg(lastMems[p.Block().ID])
				}
				lastMem = newPhi
				if debug > 0 {
					fmt.Printf("\tb%d, changed, adding mem phi %s\n", b.ID, newPhi.LongString())
				}
				memPhi[b.ID] = newPhi
			}

			// If we've not continued the loop, that means a new memory
			// is present at the top of this block and needs to be copied
			// into any uses.
			for _, v := range b.Values {
				if v.Op == OpPhi {
					continue
				}
				// If z has a true memory operand (not a tuple) and it is not lastMem, change it.
				if m := v.MemoryArg(); m != nil && m.Type == types.TypeMem && m != lastMem {
					v.SetArg(len(v.Args)-1, lastMem)
				}
				if v.Type == types.TypeMem {
					// because they're in dependence order, can quit early
					break
				}
			}
		}
	}
}

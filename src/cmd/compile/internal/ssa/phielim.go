// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

// phielim eliminates redundant phi values from f.
// A phi is redundant if its arguments are all equal. For
// purposes of counting, ignore the phi itself. Both of
// these phis are redundant:
//   v = phi(x,x,x)
//   v = phi(x,v,x,v)
// We repeat this process to also catch situations like:
//   v = phi(x, phi(x, x), phi(x, v))
// For phis with duplicate args:
//   v = phi(x,x,w)
//  the duplicate arguments are removed, and a new predecessor
//  block is inserted.
// TODO: Can we also simplify cases like:
//   v = phi(v, w, x)
//   w = phi(v, w, x)
// and would that be useful?
func phielim(f *Func) {
	phiMap := newSparseMap(f.NumValues())
	for {
		change := false
		for _, b := range f.Blocks {
			phiCnt := 0
			for _, v := range b.Values {
				copyelimValue(v)
				change = phielimValue(v) || change
				if v.Op == OpPhi {
					phiCnt++
				}
			}

			// handling the single phi case as it gets more difficult
			// if there are multiple phis
			if phiCnt == 1 {
				for _, v := range b.Values {
					phiMap.clear()
					change = phielimSplit(f, v, phiMap) || change
				}
			}
		}
		if !change {
			break
		}
	}
}

func phielimValue(v *Value) bool {
	if v.Op != OpPhi {
		return false
	}

	// If there are two distinct args of v which
	// are not v itself, then the phi must remain.
	// Otherwise, we can replace it with a copy.
	var w *Value
	for i, x := range v.Args {
		if b := v.Block.Preds[i]; b.Kind == BlockFirst && b.Succs[1] == v.Block {
			// This branch is never taken so we can just eliminate it.
			continue
		}
		if x == v {
			continue
		}
		if x == w {
			continue
		}
		if w != nil {
			return false
		}
		w = x
	}

	if w == nil {
		// v references only itself. It must be in
		// a dead code loop. Don't bother modifying it.
		return false
	}
	v.Op = OpCopy
	v.SetArgs1(w)
	f := v.Block.Func
	if f.pass.debug > 0 {
		f.Config.Warnl(v.Line, "eliminated phi")
	}
	return true
}

type phiArg struct {
	phi  *Value
	idxs []int
}

// phielimSplit splits phis that contain multiple arguments of the same
// value.
func phielimSplit(f *Func, phi *Value, argMap *sparseMap) bool {
	if phi.Op != OpPhi {
		return false
	}

	// using sparseMap and a slice here avoids the random iteration order
	// associated with using a map

	// argMap maps from argument id to index in phiArgs
	phiArgs := make([]phiArg, 0, 5)
	// phiArgs is contains entries, one per unique argument, that
	// store the indices in the phi of that argument
	for i, x := range phi.Args {
		if !argMap.contains(x.ID) {
			phiArgs = append(phiArgs, phiArg{x, nil})
			argMap.set(x.ID, int32(len(phiArgs)-1))
		}
		idx := argMap.get(x.ID)
		phiArgs[idx].idxs = append(phiArgs[idx].idxs, i)
	}

	for _, se := range argMap.contents() {
		phiArg := phiArgs[se.val]
		// check if we have more then a single copy of this
		// argument value
		if len(phiArg.idxs) < 2 {
			continue
		}

		// construct a new block with a copy of the phi arg
		nb := f.NewBlock(BlockPlain)
		if f.pass.debug > 0 {
			f.Config.Warnl(phi.Line, "split phi")
		}
		nb.Succs = append(nb.Succs, phi.Block)

		for _, bidx := range phiArg.idxs {
			// copy predecessor from phi block to the new block
			nb.Preds = append(nb.Preds, phi.Block.Preds[bidx])
			// set those successors to the new block
			for pidx, bv := range phi.Block.Preds[bidx].Succs {
				if bv == phi.Block {
					phi.Block.Preds[bidx].Succs[pidx] = nb
					// set the predecessor and phi argument to nil so
					// they can be cleared later
					phi.Block.Preds[bidx] = nil
					phi.Args[bidx] = nil
				}
			}
		}

		// fix up the phi Preds/Args by filtering out the nil values
		prevPreds := phi.Block.Preds
		phi.Block.Preds = prevPreds[:0]
		prevArgs := phi.Args
		phi.Args = prevArgs[:0]
		if len(phi.Args) != len(phi.Block.Preds) {
			f.Fatalf("expected preds length == phi arg count")
		}

		for i := range prevPreds {
			if prevPreds[i] != nil {
				phi.Block.Preds = append(phi.Block.Preds, prevPreds[i])
			}
			if prevArgs[i] != nil {
				phi.Args = append(phi.Args, prevArgs[i])
			}
		}

		// add a single copy of the value to the phi, and a predecessor of our new block
		phi.Args = append(phi.Args, phiArg.phi)
		phi.Block.Preds = append(phi.Block.Preds, nb)
		return true
	}
	return false
}

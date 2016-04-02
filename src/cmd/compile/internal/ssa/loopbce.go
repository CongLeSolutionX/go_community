package ssa

import "fmt"

type indVar struct {
	ind    *Value // induction variable
	inc    *Value // increment, a constant
	nxt    *Value // ind+inc variable
	min    *Value // minimum value. inclusive,
	max    *Value // maximum value. exclusive.
	maxOff int64  // maximum value offset
	entry  *Block // entry block in the loop.
	// Invariants: for all blocks dominated by entry:
	//	min <= ind < max+maxOff
	//	min <= nxt <= max+maxOff

	slice *Value // points to SliceMake or StringMake if known
	ptr   *Value // points to the parallel vector iterating over slice
}

// findIndVar finds induction variables in a function.
//
// Look for variables and blocks that satisfy the following
//
// loop:
//   ind = (Phi min nxt),
//   if ind < max
//     then goto enter_loop
//     else goto exit_loop
//
//   enter_loop:
//	do something
//      nxt = inc + ind
//	goto loop
//
// exit_loop:
//
//
// TODO: handle 32 bit operations
func findIndVar(f *Func, sdom sparseTree) []indVar {
	var iv []indVar

nextb:
	for _, b := range f.Blocks {
		if b.Kind != BlockIf || len(b.Preds) != 2 {
			continue
		}

		var ind, max *Value // induction, and maximum
		entry := -1         // which successor of b enters the loop

		// Check thet the control if it either ind < max or max > ind.
		// TODO: Handle Leq64, Geq64.
		switch b.Control.Op {
		case OpLess64:
			entry = 0
			ind, max = b.Control.Args[0], b.Control.Args[1]
		case OpGreater64:
			entry = 0
			ind, max = b.Control.Args[1], b.Control.Args[0]
		default:
			continue nextb
		}

		// Check that the induction variable is a phi that depends on itself.
		if ind.Op != OpPhi {
			continue
		}

		// Extract min and nxt knowing that nxt is an addition (e.g. Add64).
		var min, nxt *Value // minimum, and next value
		if n := ind.Args[0]; n.Op == OpAdd64 && (n.Args[0] == ind || n.Args[1] == ind) {
			min, nxt = ind.Args[1], n
		} else if n := ind.Args[1]; n.Op == OpAdd64 && (n.Args[0] == ind || n.Args[1] == ind) {
			min, nxt = ind.Args[0], n
		} else {
			// Not a recognized induction variable.
			continue
		}

		var inc *Value
		if nxt.Args[0] == ind { // nxt = ind + inc
			inc = nxt.Args[1]
		} else if nxt.Args[1] == ind { // nxt = inc + ind
			inc = nxt.Args[0]
		} else {
			panic("unreachable") // one of the cases must be true from the above.
		}

		// Expect the increment to be a positive constant.
		// TODO: handle negative increment.
		if inc.Op != OpConst64 || inc.AuxInt <= 0 {
			continue
		}

		// Up to now we extracted the induction variable (ind),
		// the increment delta (inc), the temporary sum (nxt),
		// the mininum value (min) and the maximum value (max).
		//
		// We also know that ind has the form (Phi min nxt) where
		// nxt is (Add inc nxt) which means: 1) inc dominates nxt
		// and 2) there is a loop starting at inc and containing nxt.
		//
		// We need to prove that the induction variable is incremented
		// only when it's smaller than the maximum value.
		// Two conditions must happen listed below to accept ind
		// as an induction variable.

		// First condition: loop entry has a single predecessor, which
		// is the header block.  This implies that b.Succs[entry] is
		// reached iff ind < max.
		if len(b.Succs[entry].Preds) != 1 {
			// b.Succs[1-entry] must exit the loop.
			continue
		}

		// Second condition: b.Succs[entry] dominates nxt so that
		// nxt is computed when inc < max, meaning nxt <= max.
		if !sdom.isAncestorEq(b.Succs[entry], nxt.Block) {
			// inc+ind can only be reached through the branch that enters the loop.
			continue
		}

		// If max is c + SliceLen with c <= 0 then we drop c.
		// Makes sure c + SliceLen doesn't overflow when SliceLen == 0.
		maxOff := int64(0)
		if w, c := dropAdd64(max); (w.Op == OpStringLen || w.Op == OpSliceLen) && 0 >= c && -c >= 0 {
			max, maxOff = w, c
		}

		// We can only guarantee that the loops runs withing limits of induction variable
		// if the increment is 1 or when the limits are constants.
		if inc.AuxInt != 1 {
			ok := false
			if min.Op == OpConst64 && max.Op == OpConst64 {
				if max.AuxInt > min.AuxInt && max.AuxInt%inc.AuxInt == min.AuxInt%inc.AuxInt { // handle overflow
					ok = true
				}
			}
			if !ok {
				continue
			}
		}

		if f.pass.debug > 1 {
			if min.Op == OpConst64 {
				b.Func.Config.Warnl(b.Line, "Induction variable with minimum %d and increment %d", min.AuxInt, inc.AuxInt)
			} else {
				b.Func.Config.Warnl(b.Line, "Induction variable with non-const minimum and increment %d", inc.AuxInt)
			}
		}

		iv = append(iv, indVar{
			ind:    ind,
			inc:    inc,
			nxt:    nxt,
			min:    min,
			max:    max,
			maxOff: maxOff,
			entry:  b.Succs[entry],
		})
		b.Logf("found induction variable %v (inc = %v, min = %v, max = %v)\n", ind, inc, min, max)
	}

	return iv
}

func findSlice(sdom sparseTree, iv *indVar) {
	b := iv.ind.Block
	if iv.max.Op != OpSliceLen || iv.maxOff != 0 || iv.min.Op != OpConst64 || iv.min.AuxInt != 0 {
		// TODO: handle non-0 max offset
		// TODO: handle non-0 min.
		b.Logf("skipping %v %v %d %v %d\n", *iv, iv.max.Op, iv.maxOff, iv.min.Op, iv.min.AuxInt)
		return
	}

	iv.slice = iv.max.Args[0]
	b.Logf("iv %v iterates over %v\n", iv.ind, iv.slice)

	for _, ptr := range b.Values {
		if ptr.Op != OpPhi {
			continue
		}
		// Checks ptr is of form (+ permutations)
		// ptr = (Phi slice (Add64 size ptr))

		var inc *Value
		if ptr.Args[0].Op == OpSlicePtr && ptr.Args[0].Args[0] == iv.slice {
			inc = ptr.Args[1]
		} else {
			// TODO
			continue
		}
		if inc.Op != OpAdd64 {
			continue
		}
		if !sdom.isAncestorEq(iv.entry, inc.Block) {
			continue
		}

		size := iv.slice.Type.ElemType().Size()
		if !(inc.Args[0].Op == OpConst64 && inc.Args[0].AuxInt == size && inc.Args[1] == ptr) &&
			!(inc.Args[1].Op == OpConst64 && inc.Args[1].AuxInt == size && inc.Args[0] == ptr) {
			continue
		}

		iv.ptr = ptr
		b.Logf("ptr %v iterates over %v\n", iv.ptr, iv.slice)
		// fmt.Printf("ptr %v iterates over %v, iv.ind.Uses %d in %s\n", iv.ptr, iv.slice, iv.ind.Uses, iv.ind.Block.Func.Name)
		_ = fmt.Print
	}
}

// loopbce performs loop based bounds check elimination.
func loopbce(f *Func) {
	idom := dominators(f)
	sdom := newSparseTree(f, idom)

	ivList := findIndVar(f, sdom)
	ivMap := make(map[*Value]indVar)
	for i := range ivList {
		iv := &ivList[i]
		findSlice(sdom, iv)
		ivMap[iv.ind] = *iv
	}

	removeIndVar(f, ivList)
	removeBoundsChecks(f, sdom, ivMap)
}

func removeIndVar(f *Func, ivList []indVar) {
	for _, iv := range ivList {
		if iv.ptr == nil || iv.ind.Uses > 2 {
			continue
		}

		// Replaces ind < max by ptr < slice + size * max
		size := iv.slice.Type.ElemType().Size()
		maxptr := iv.ind.Block.NewValue2(iv.ptr.Line, OpAdd64, iv.ptr.Type,
			iv.ptr.Args[0], // TODO: what if not minimum
			iv.ind.Block.NewValue2(iv.ptr.Line, OpMul64, f.Config.Frontend().TypeInt(),
				f.ConstInt64(iv.ind.Line, f.Config.Frontend().TypeInt(), size),
				iv.max))
		iv.ind.Block.SetControl(iv.ind.Block.NewValue2(iv.ind.Line, OpLess64, f.Config.Frontend().TypeBool(), iv.ptr, maxptr))

		f.Logf("dropped induction variable %v in favor of %v\n", iv.ind, iv.ptr)
	}
}

// removesBoundsChecks remove IsInBounds and IsSliceInBounds based on the induction variables.
func removeBoundsChecks(f *Func, sdom sparseTree, m map[*Value]indVar) {
	for _, b := range f.Blocks {
		if b.Kind != BlockIf {
			continue
		}

		v := b.Control

		// Simplify:
		// (IsInBounds ind max) where 0 <= const == min <= ind < max.
		// (IsSliceInBounds ind max) where 0 <= const == min <= ind < max.
		// Found in:
		//	for i := range a {
		//		use a[i]
		//		use a[i:]
		//		use a[:i]
		//	}
		if v.Op == OpIsInBounds || v.Op == OpIsSliceInBounds {
			ind, add := dropAdd64(v.Args[0])
			if ind.Op != OpPhi {
				goto skip1
			}
			if v.Op == OpIsInBounds && add != 0 {
				goto skip1
			}
			if v.Op == OpIsSliceInBounds && (0 > add || add > 1) {
				goto skip1
			}

			if iv, has := m[ind]; has && sdom.isAncestorEq(iv.entry, b) && isNonNegative(iv.min) {
				if v.Args[1] == iv.max {
					if f.pass.debug > 0 {
						f.Config.Warnl(b.Line, "Found redundant %s", v.Op)
					}
					goto simplify
				}
			}
		}
	skip1:

		// Simplify:
		// (IsSliceInBounds ind (SliceCap a)) where 0 <= min <= ind < max == (SliceLen a)
		// Found in:
		//	for i := range a {
		//		use a[:i]
		//		use a[:i+1]
		//	}
		if v.Op == OpIsSliceInBounds {
			ind, add := dropAdd64(v.Args[0])
			if ind.Op != OpPhi {
				goto skip2
			}
			if 0 > add || add > 1 {
				goto skip2
			}

			if iv, has := m[ind]; has && sdom.isAncestorEq(iv.entry, b) && isNonNegative(iv.min) {
				if v.Args[1].Op == OpSliceCap && iv.max.Op == OpSliceLen && v.Args[1].Args[0] == iv.max.Args[0] {
					if f.pass.debug > 0 {
						f.Config.Warnl(b.Line, "Found redundant %s (len promoted to cap)", v.Op)
					}
					goto simplify
				}
			}
		}
	skip2:

		continue

	simplify:
		f.Logf("removing bounds check %v at %v in %s\n", b.Control, b, f.Name)
		b.Kind = BlockFirst
		b.SetControl(nil)
	}
}

func dropAdd64(v *Value) (*Value, int64) {
	if v.Op == OpAdd64 && v.Args[0].Op == OpConst64 {
		return v.Args[1], v.Args[0].AuxInt
	}
	if v.Op == OpAdd64 && v.Args[1].Op == OpConst64 {
		return v.Args[0], v.Args[1].AuxInt
	}
	return v, 0
}

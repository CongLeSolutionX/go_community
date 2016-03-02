package ssa

type indVar struct {
	ind   *Value // induction variable
	inc   *Value // increment, a constant
	nxt   *Value // ind+inc variable
	min   *Value // minimum value. inclusive,
	max   *Value // maximum value. exclusive.
	entry *Block // entry block in the loop.
	// Invariants: for all blocks dominated by entry:
	//	min <= ind < max
	//	min <= nxt <= max
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
func findIndVar(f *Func, sdom sparseTree) []indVar {
	var iv []indVar

	for _, b := range f.Blocks {
		if b.Kind != BlockIf || len(b.Preds) != 2 {
			continue
		}
		if b.Control.Op != OpLess64 {
			continue
		}
		if len(b.Succs[0].Preds) != 1 {
			// b.Succs[1] must exit the loop.
			continue
		}

		ind := b.Control.Args[0]
		if ind.Op != OpPhi {
			continue
		}
		if ind.Args[0].Op != OpConst64 {
			// TODO: handle non-const minimum
			continue
		}

		if ind.Args[1].Op != OpAdd64 {
			continue
		}

		var inc *Value
		if ind.Args[1].Args[0] == ind {
			// ind = (Phi min (Add64 ind inc))
			inc = ind.Args[1].Args[1]
		} else if ind.Args[1].Args[1] == ind {
			// ind = (Phi min (Add64 inc ind))
			inc = ind.Args[1].Args[0]
		} else {
			continue
		}

		if inc.Op != OpConst64 || inc.AuxInt <= 0 {
			// TODO: handle negative increment
			continue
		}
		if !sdom.isAncestorEq(b.Succs[0], ind.Args[1].Block) {
			// inc+ind can only be reached through the branch that enters the loop.
			continue
		}

		min := ind.Args[0]
		max := b.Control.Args[1]

		// If max is c + SliceLen with c <= 0 then we drop c.
		// TODO: save c as an offset from max.
		if w, c := dropAdd64(max); (w.Op == OpStringLen || w.Op == OpSliceLen) && c <= 0 {
			max = w
		}

		if inc.AuxInt != 1 {
			ok := false
			if min.Op == OpConst64 && max.Op == OpConst64 {
				if (max.AuxInt-min.AuxInt)%inc.AuxInt == 0 {
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
			ind:   ind,
			inc:   inc,
			nxt:   ind.Args[1],
			min:   min,
			max:   max,
			entry: b.Succs[0],
		})
		b.Logf("found induction variable %v (inc = %v, min = %v, max = %v)\n", ind, inc, min, max)
	}

	return iv
}

// loopbce performs loop based bounds check elimination.
func loopbce(f *Func) {
	idom := dominators(f)
	sdom := newSparseTree(f, idom)
	iv := findIndVar(f, sdom)

	m := make(map[*Value]indVar)
	for _, iv := range iv {
		m[iv.ind] = iv
	}

	removeBoundsChecks(f, sdom, m)
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

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

func findIndVar(f *Func, sdom sparseTree) []indVar {
	var iv []indVar

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
	// TODO: handle ind = (Phi min (inc+ind)).

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
		if ind.Args[1].Op != OpAdd64 || ind.Args[1].Args[1] != ind {
			continue
		}

		// cse sorts Add64 arguments such that the constant is first.
		inc := ind.Args[1].Args[0]
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
				b.Func.Config.Warnl(int(b.Line), "Induction variable with minimum %d and increment %d", min.AuxInt, inc.AuxInt)
			} else {
				b.Func.Config.Warnl(int(b.Line), "Induction variable with non-const minimum and increment %d", inc.AuxInt)
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

func strength(f *Func) {
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
		// Found in for i := range a { do something with a[i] }
		if v.Op == OpIsInBounds || v.Op == OpIsSliceInBounds {
			if iv, has := m[v.Args[0]]; has && sdom.isAncestorEq(iv.entry, b) {
				if iv.min.Op == OpConst64 && iv.min.AuxInt >= 0 && v.Args[1] == iv.max {
					if f.pass.debug > 0 {
						f.Config.Warnl(int(b.Line), "Found redundant %s", v.Op)
					}
					goto simplify
				}
			}
		}

		// Simplify
		// (IsSliceInBounds c ind) with c <= min.
		// Found in for i := range a { do something with a[:i] }
		if v.Op == OpIsSliceInBounds {
			ind := v.Args[1]
			if iv, has := m[ind]; has && sdom.isAncestorEq(iv.entry, b) {
				inner := v.Args[0]
				if inner.Op == OpConst64 && iv.min.Op == OpConst64 && inner.AuxInt <= iv.min.AuxInt {
					if f.pass.debug > 0 {
						f.Config.Warnl(int(b.Line), "Found redundant %s (%d <= %d)", v.Op, inner.AuxInt, iv.min.AuxInt)
					}
					goto simplify
				}
			}
		}

		// Simplify
		// (IsSliceInBounds (ind+d) max)
		// Found in for i := 0; i < c*C; i += c { do something with a[:i+d] where d <= c }
		// Most commonly c == 1 and C == len(string)
		if v.Op == OpIsSliceInBounds {
			ind, add := dropAdd64(v.Args[0])

			if iv, has := m[ind]; has && sdom.isAncestorEq(iv.entry, b) && iv.inc.AuxInt >= add && add >= 0 {
				bound := v.Args[1]
				if iv.min.Op == OpConst64 && iv.min.AuxInt >= 0 && bound == iv.max {
					if f.pass.debug > 0 {
						f.Config.Warnl(int(b.Line), "Found redundant %s (bound is %s, max is %s)", v.Op, bound.Op, iv.max.Op)
					}
					goto simplify
				}
				if iv.min.Op == OpConst64 && iv.min.AuxInt >= 0 && bound.Op == OpSliceCap && iv.max.Op == OpSliceLen && bound.Args[0] == iv.max.Args[0] {
					if f.pass.debug > 0 {
						f.Config.Warnl(int(b.Line), "Found redundant %s (bound is %s, max is %s)", v.Op, bound.Op, iv.max.Op)
					}
					goto simplify
				}
			}
		}

		continue

	simplify:
		f.Logf("removing bounds check %v at %v in %s\n", b.Control, b, f.Name)
		b.Kind = BlockFirst
		b.Control = nil
	}
}

func dropAdd64(v *Value) (*Value, int64) {
	if v.Op == OpAdd64 && v.Args[0].Op == OpConst64 {
		return v.Args[1], v.Args[0].AuxInt
	}
	return nil, 0
}

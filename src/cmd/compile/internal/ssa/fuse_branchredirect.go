// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

// fuseBranchRedirect checks for a CFG in which the outbound branch
// of an If block can be derived from its predecessor If block, in
// some such cases, we can redirect the predecessor If block to the
// corresponding successor block. For example:
// p:
//   v11 = Less64 <bool> v10 v8
//   If v11 goto b else u
// b: <- p ...
//   v17 = Leq64 <bool> v10 v8
//   If v17 goto s else o
// We can redirect p to s directly.
//
// TODO: Currently only the operators listed above are covered, and the
// target needs to meet strict conditions, such as the arguments must be
// the same (the order can be different). Thre are many operators of bool
// type, and some may not be covered.
// In addition, the following situation can theoretically be optimized.
// 		var a, b int64
// 		if a > 0 && b > 0 || a >= 1 && b < 0
// But the processing will be more complicated.
func fuseBranchRedirect(b *Block) bool {
	if b.Kind != BlockIf {
		return false
	}
	bCtl := b.Controls[0]
	// bCtl may be a phi value.
	if !bCtl.Type.IsBoolean() {
		return false
	}
	// Look for control values of the form Copy(Not(Copy(Less64(v1, v2)))).
	// Track the negations so that we can swap successors as needed later.
	var bswap, nval int
	for bCtl.Op == OpCopy || bCtl.Op == OpNot {
		if bCtl.Op == OpNot {
			bswap = 1 ^ bswap
		}
		if bCtl.Block == b && bCtl.Uses == 1 {
			nval++
		}
		bCtl = bCtl.Args[0]
	}
	if bCtl.Block == b && bCtl.Uses == 1 {
		nval++
	}
	// No other values can be included in b, otherwise, if these values are used
	// by other blocks (such as the taken branch of b), it may cause errors.
	if nval != len(b.Values) {
		return false
	}
	bCtlOp := bCtl.Op
	bWidth := flagOpWidth(bCtlOp)
	if bWidth == WdtInvalid {
		return false
	}
	changed := false
	for k := 0; k < len(b.Preds); k++ {
		pr := b.Preds[k]
		// p is the predecessor block of b.
		p := pr.b
		if p == nil || p.Kind != BlockIf || p == b {
			continue
		}
		pCtl := p.Controls[0]
		// pCtl may be a phi value.
		if !pCtl.Type.IsBoolean() {
			continue
		}
		pswap := bswap
		for pCtl.Op == OpCopy || pCtl.Op == OpNot {
			if pCtl.Op == OpNot {
				pswap = 1 ^ pswap
			}
			pCtl = pCtl.Args[0]
		}
		// Check if the arguments of bCtl and pCtl are the same.
		argsR := argsRelation(bCtl, pCtl)
		// If the arguments of bCtl and pCtl are not the same or
		// the width of the operands are inconsistent, continue.
		pCtlOp := pCtl.Op
		if argsR == NoRelation || bWidth != flagOpWidth(pCtlOp) {
			continue
		}
		// If both bCtro and pCtl are Phi or Arg value, only when their IDs
		// are the same can we determine that their judgments are the same.
		if (bCtl.Op == OpPhi || bCtl.Op == OpArg) && bCtl.ID != pCtl.ID {
			continue
		}
		// So far we have determined that the arguments of bCtl and pCtl are the same. But we still need to
		// determine whether the specific outbound branch of b is known through the Ops, branch information
		// (by which branch (0 or 1) p is connected to b) and the order (same or reverse) of their arguments.
		// Example:
		//     -----------------------
		//     | p:                  |
		//     |   ...               |
		//     |   v5 = Less64 v3 v4 |
		//     | If v5 -> b, s       |
		//     -----------------------
		//          /         \ 0
		//                     \
		//           -----------------------
		//           | b:                  |
		//           |   ...               |
		//           |   v6 = Leq64 v3 v4  |
		//           | If v6 -> tb, u      |
		//           -----------------------
		// p is connected to b through the "0" branch, then in b we can infer that v6 is true if the control
		// flow is come from p, so we can redirect p.Succs[0] to tb directly. But if p is connected to b
		// through the "1" branch or the Op of v6 is not Leq64 or the order of the arguments are reverse, then we
		// may not be able determine whether v6 is true. So we need to synthesize the above information to see if
		// we can determine a known successor branch.
		pi := pr.i ^ pswap
		bsi := -1
		bTyp := flagOpType(bCtlOp)
		pTyp := flagOpType(pCtlOp)
		// Find possible known successor branch.
		bsi = inferSuccs(pTyp, bTyp, argsR, pi)
		if bsi == -1 {
			continue
		}
		// Redirect b's k th predecessor to b's bsi th successor.
		tb := b.Succs[bsi].b
		if tb == b || tb == p {
			continue
		}
		b.removePred(k)
		p.Succs[pr.i] = Edge{tb, len(tb.Preds)}
		// Fix up Phi value in b to have one less argument.
		// Fixme: Is this possible ?
		if bCtlOp == OpPhi && bCtl.Block == b {
			bCtl.RemoveArg(k)
			phielimValue(bCtl)
		}
		// Fix up tb to have one more predecessor.
		tb.Preds = append(tb.Preds, Edge{p, pr.i})
		tbPi := b.Succs[bsi].i
		for _, v := range tb.Values {
			if v.Op != OpPhi {
				continue
			}
			v.AddArg(v.Args[tbPi])
		}
		if b.Func.pass.debug > 0 {
			b.Func.Warnl(bCtl.Pos, "Redirect %s based on %s", bCtlOp, pCtlOp)
		}
		changed = true
		k--
	}
	if len(b.Preds) == 0 {
		// Block is now dead.
		b.Kind = BlockInvalid
	}
	return changed
}

// flag Op type representations. Different operators operate on
// different data types, and some operators have the same behavior
// here, such as OpEq8, OpEq16, OpEq32, OpEq64, OpEq32F and OpEq64F.
// So classify them as a type, then we can do different processing
// according to the type rather than the specific operator, which
// can simplify the code.
const (
	TypEq int = iota
	TypNeq
	TypLess
	TypLeq
	TypLessU
	TypLeqU
	TypEqB
	TypNeqB
	TypEqPtr
	TypNeqPtr
	TypEqInter
	TypNeqInter
	TypEqSlice
	TypNeqSlice
	TypPhi
	TypArg
	TypIsNonNil
	TypInvalid
)

// flagOpType categorizes Ops with consistent processing logic
// into the same type representation.
func flagOpType(o Op) int {
	switch o {
	case OpEq8, OpEq16, OpEq32, OpEq64, OpEq32F, OpEq64F:
		return TypEq
	case OpNeq8, OpNeq16, OpNeq32, OpNeq64, OpNeq32F, OpNeq64F:
		return TypNeq
	case OpLess8, OpLess16, OpLess32, OpLess64, OpLess32F, OpLess64F:
		return TypLess
	case OpLeq8, OpLeq16, OpLeq32, OpLeq64, OpLeq32F, OpLeq64F:
		return TypLeq
	case OpLess8U, OpLess16U, OpLess32U, OpLess64U:
		return TypLessU
	case OpLeq8U, OpLeq16U, OpLeq32U, OpLeq64U:
		return TypLeqU
	case OpEqB:
		return TypEqB
	case OpNeqB:
		return TypNeqB
	case OpEqPtr:
		return TypEqPtr
	case OpNeqPtr:
		return TypNeqPtr
	case OpEqInter:
		return TypEqInter
	case OpNeqInter:
		return TypNeqInter
	case OpEqSlice:
		return TypEqSlice
	case OpNeqSlice:
		return TypNeqSlice
	case OpPhi:
		return TypPhi
	case OpArg:
		return TypArg
	case OpIsNonNil:
		return TypIsNonNil
	default:
		return TypInvalid
	}
}

// flag Op width representations.
const (
	WdtB int = iota
	Wdt8
	Wdt16
	Wdt32
	Wdt64
	Wdt32f
	Wdt64f
	WdtPtr
	WdtInter
	WdtSlice
	WdtPhi
	WdtArg
	WdtIsNonNil
	WdtInvalid
)

// flagOpWidth returns the width representation of the Op o.
func flagOpWidth(o Op) int {
	switch o {
	case OpEqB, OpNeqB:
		return WdtB
	case OpEq8, OpNeq8, OpLess8, OpLess8U, OpLeq8, OpLeq8U:
		return Wdt8
	case OpEq16, OpNeq16, OpLess16, OpLess16U, OpLeq16, OpLeq16U:
		return Wdt16
	case OpEq32, OpNeq32, OpLess32, OpLess32U, OpLeq32, OpLeq32U:
		return Wdt32
	case OpEq64, OpNeq64, OpLess64, OpLess64U, OpLeq64, OpLeq64U:
		return Wdt64
	case OpEq32F, OpNeq32F, OpLess32F, OpLeq32F:
		return Wdt32f
	case OpEq64F, OpNeq64F, OpLess64F, OpLeq64F:
		return Wdt64f
	case OpEqPtr, OpNeqPtr:
		return WdtPtr
	case OpEqInter, OpNeqInter:
		return WdtInter
	case OpEqSlice, OpNeqSlice:
		return WdtSlice
	case OpPhi:
		return WdtPhi
	case OpArg:
		return WdtArg
	case OpIsNonNil:
		return WdtIsNonNil
	default:
		return WdtInvalid
	}
}

// Representation of the argument relationship between two values.
const (
	NoRelation int = iota
	SameArgsDiffOrder
	SameArgsSameOrder
)

// argsRelation checks if v1 and v2 have the same arguments.
// If the arguments and the order are the same, return 2;
// if the arguments are the same but the order is reverse, return 1;
// otherwise return 0.
func argsRelation(v1, v2 *Value) int {
	if v1.ID == v2.ID {
		return SameArgsSameOrder
	}
	if len(v1.Args) != len(v2.Args) {
		return NoRelation
	}
	switch len(v1.Args) {
	case 2:
		if v1.Args[0] == v2.Args[1] && v1.Args[1] == v2.Args[0] {
			return SameArgsDiffOrder
		}
		if v1.Args[0] == v2.Args[0] && v1.Args[1] == v2.Args[1] {
			return SameArgsSameOrder
		}
	case 1:
		if v1.Args[0] == v2.Args[0] {
			return SameArgsSameOrder
		}
	default:
	}
	return NoRelation
}

// inferSuccs determines whether a known successor branch can be derived
// based on the following four types of parameter information. Suppose p
// is the predecessor of b.
// pTyp is the type representation of the Op of p's control value.
// bTyp is the type representation of the Op of b's control value.
// argRel is the relationship representation of the arguments of p's control
// value and b's control value
// pi indicates which branch (0 or 1) p connects to b.
func inferSuccs(pTyp, bTyp, argRel, pi int) int {
	bsi := -1
	if argRel == SameArgsSameOrder {
		switch {
		case pTyp == bTyp:
			// TypPhi, TypArg and TypIsNonNil only exist here.
			bsi = pi
		case pTyp == TypEq && bTyp == TypNeq,
			pTyp == TypNeq && bTyp == TypEq,
			pTyp == TypEqB && bTyp == TypNeqB,
			pTyp == TypNeqB && bTyp == TypEqB,
			pTyp == TypEqPtr && bTyp == TypNeqPtr,
			pTyp == TypNeqPtr && bTyp == TypEqPtr,
			pTyp == TypEqInter && bTyp == TypNeqInter,
			pTyp == TypNeqInter && bTyp == TypEqInter,
			pTyp == TypEqSlice && bTyp == TypNeqSlice,
			pTyp == TypNeqSlice && bTyp == TypEqSlice:
			bsi = pi ^ 1
		case pTyp == TypEq && bTyp == TypLess,
			pTyp == TypEq && bTyp == TypLessU,
			pTyp == TypLess && bTyp == TypEq,
			pTyp == TypLessU && bTyp == TypEq:
			if pi == 0 {
				bsi = 1
			}
		case pTyp == TypEq && bTyp == TypLeq,
			pTyp == TypEq && bTyp == TypLeqU,
			pTyp == TypLess && bTyp == TypNeq,
			pTyp == TypLess && bTyp == TypLeq,
			pTyp == TypLessU && bTyp == TypNeq,
			pTyp == TypLessU && bTyp == TypLeqU:
			if pi == 0 {
				bsi = 0
			}
		case pTyp == TypNeq && bTyp == TypLess,
			pTyp == TypNeq && bTyp == TypLessU,
			pTyp == TypLeq && bTyp == TypEq,
			pTyp == TypLeq && bTyp == TypLess,
			pTyp == TypLeqU && bTyp == TypEq,
			pTyp == TypLeqU && bTyp == TypLessU:
			if pi == 1 {
				bsi = 1
			}
		case pTyp == TypNeq && bTyp == TypLeq,
			pTyp == TypNeq && bTyp == TypLeqU,
			pTyp == TypLeq && bTyp == TypNeq,
			pTyp == TypLeqU && bTyp == TypNeq:
			if pi == 1 {
				bsi = 0
			}
		default:
		}
	} else if argRel == SameArgsDiffOrder {
		switch {
		case pTyp == TypEq && bTyp == TypEq,
			pTyp == TypNeq && bTyp == TypNeq,
			pTyp == TypNeqB && bTyp == TypNeqB,
			pTyp == TypEqB && bTyp == TypEqB,
			pTyp == TypEqPtr && bTyp == TypEqPtr,
			pTyp == TypNeqPtr && bTyp == TypNeqPtr,
			pTyp == TypEqInter && bTyp == TypEqInter,
			pTyp == TypNeqInter && bTyp == TypNeqInter,
			pTyp == TypEqSlice && bTyp == TypEqSlice,
			pTyp == TypNeqSlice && bTyp == TypNeqSlice:
			bsi = pi
		case pTyp == TypEq && bTyp == TypNeq,
			pTyp == TypNeq && bTyp == TypEq,
			pTyp == TypLess && bTyp == TypLeq,
			pTyp == TypLeq && bTyp == TypLess,
			pTyp == TypLessU && bTyp == TypLeqU,
			pTyp == TypLeqU && bTyp == TypLessU,
			pTyp == TypEqB && bTyp == TypNeqB,
			pTyp == TypNeqB && bTyp == TypEqB,
			pTyp == TypEqPtr && bTyp == TypNeqPtr,
			pTyp == TypNeqPtr && bTyp == TypEqPtr,
			pTyp == TypEqInter && bTyp == TypNeqInter,
			pTyp == TypNeqInter && bTyp == TypEqInter,
			pTyp == TypEqSlice && bTyp == TypNeqSlice,
			pTyp == TypNeqSlice && bTyp == TypEqSlice:
			bsi = pi ^ 1
		case pTyp == TypEq && bTyp == TypLess,
			pTyp == TypEq && bTyp == TypLessU,
			pTyp == TypLess && bTyp == TypEq,
			pTyp == TypLess && bTyp == TypLess,
			pTyp == TypLessU && bTyp == TypEq,
			pTyp == TypLessU && bTyp == TypLessU:
			if pi == 0 {
				bsi = 1
			}
		case pTyp == TypEq && bTyp == TypLeq,
			pTyp == TypEq && bTyp == TypLeqU,
			pTyp == TypLess && bTyp == TypNeq,
			pTyp == TypLessU && bTyp == TypNeq:
			if pi == 0 {
				bsi = 0
			}
		case pTyp == TypNeq && bTyp == TypLess,
			pTyp == TypNeq && bTyp == TypLessU,
			pTyp == TypLeq && bTyp == TypEq,
			pTyp == TypLeqU && bTyp == TypEq:
			if pi == 1 {
				bsi = 1
			}
		case pTyp == TypNeq && bTyp == TypLeq,
			pTyp == TypNeq && bTyp == TypLeqU,
			pTyp == TypLeq && bTyp == TypNeq,
			pTyp == TypLeq && bTyp == TypLeq,
			pTyp == TypLeqU && bTyp == TypNeq,
			pTyp == TypLeqU && bTyp == TypLeqU:
			if pi == 1 {
				bsi = 0
			}
		default:
		}
	}
	return bsi
}

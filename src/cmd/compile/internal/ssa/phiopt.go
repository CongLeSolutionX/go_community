// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

// phiopt eliminates boolean Phis based on the previous if.
//
// Main use case is to transform:
//   x := false
//   if b {
//     x = true
//   }
// into x = b.
//
// In SSA code this appears as
//
// b0
//   If b -> b1 b2
// b1
//   Plain -> b2
// b2
//   x = (OpPhi (ConstBool [true]) (ConstBool [false]))
//
// In this case we can replace x with a copy of b.
func phiopt(f *Func) {
	sdom := f.Sdom()
	for _, b := range f.Blocks {
		if len(b.Preds) != 2 || len(b.Values) == 0 {
			// TODO: handle more than 2 predecessors, e.g. a || b || c.
			continue
		}

		pb0, b0 := b, b.Preds[0].b
		for len(b0.Succs) == 1 && len(b0.Preds) == 1 {
			pb0, b0 = b0, b0.Preds[0].b
		}
		if b0.Kind != BlockIf {
			continue
		}
		pb1, b1 := b, b.Preds[1].b
		for len(b1.Succs) == 1 && len(b1.Preds) == 1 {
			pb1, b1 = b1, b1.Preds[0].b
		}
		if b1 != b0 {
			continue
		}
		// b0 is the if block giving the boolean value.
		// reverse is the predecessor from which the truth value comes.
		var reverse int
		if b0.Succs[0].b == pb0 && b0.Succs[1].b == pb1 {
			reverse = 0
		} else if b0.Succs[0].b == pb1 && b0.Succs[1].b == pb0 {
			reverse = 1
		} else {
			b.Fatalf("invalid predecessors\n")
		}

		for _, v := range b.Values {
			if v.Op != OpPhi {
				continue
			}

			// Look for conversions from bool to 0/1.
			if v.Type.IsInteger() {
				phioptint(v, b0, reverse)
			}

			if !v.Type.IsBoolean() {
				continue
			}

			// Replaces
			//   if a { x = true } else { x = false } with x = a
			// and
			//   if a { x = false } else { x = true } with x = !a
			if v.Args[0].Op == OpConstBool && v.Args[1].Op == OpConstBool {
				if v.Args[reverse].AuxInt != v.Args[1-reverse].AuxInt {
					ops := [2]Op{OpNot, OpCopy}
					v.reset(ops[v.Args[reverse].AuxInt])
					v.AddArg(b0.Controls[0])
					if f.pass.debug > 0 {
						f.Warnl(b.Pos, "converted OpPhi to %v", v.Op)
					}
					continue
				}
			}

			// Replaces
			//   if a { x = true } else { x = value } with x = a || value.
			// Requires that value dominates x, meaning that regardless of a,
			// value is always computed. This guarantees that the side effects
			// of value are not seen if a is false.
			if v.Args[reverse].Op == OpConstBool && v.Args[reverse].AuxInt == 1 {
				if tmp := v.Args[1-reverse]; sdom.IsAncestorEq(tmp.Block, b) {
					v.reset(OpOrB)
					v.SetArgs2(b0.Controls[0], tmp)
					if f.pass.debug > 0 {
						f.Warnl(b.Pos, "converted OpPhi to %v", v.Op)
					}
					continue
				}
			}

			// Replaces
			//   if a { x = value } else { x = false } with x = a && value.
			// Requires that value dominates x, meaning that regardless of a,
			// value is always computed. This guarantees that the side effects
			// of value are not seen if a is false.
			if v.Args[1-reverse].Op == OpConstBool && v.Args[1-reverse].AuxInt == 0 {
				if tmp := v.Args[reverse]; sdom.IsAncestorEq(tmp.Block, b) {
					v.reset(OpAndB)
					v.SetArgs2(b0.Controls[0], tmp)
					if f.pass.debug > 0 {
						f.Warnl(b.Pos, "converted OpPhi to %v", v.Op)
					}
					continue
				}
			}
		}
	}
	// strengthen phi optimization.
	// Main use case is to transform:
	//   x := false
	//   if c {
	//     x = true
	//     ...
	//   }
	// into
	//   x := c
	//   if x { ... }
	//
	// For example, in SSA code a case appears as
	// b0
	//   If c -> b, sb0
	// sb0
	//   If d -> sd0, sd1
	// sd1
	//   ...
	// sd0
	//   Plain -> b
	// b
	//   x = (OpPhi (ConstBool [true]) (ConstBool [false]))
	//
	// In this case we can also replace x with a copy of c.
	//
	// The optimization idea:
	// 1. block b has a phi value x, x = OpPhi (ConstBool [true]) (ConstBool [false]),
	//    and len(b.Preds) is equal to 2.
	// 2. find the common dominator(b0) of the predecessors(pb0, pb1) of block b,
	//    except when pb0 or pb1 is the dominator(b0), and the dominator(b0) is a If block.
	// 3. the successors(sb0, sb1) of the dominator need to dominate the predecessors(pb0, pb1)
	//    of block b respectively.
	// 4. replace this boolean Phi based on dominator block.
	//
	//     b0  ...       ...   b0              b0
	//    |  \ /           \  / |             /  \
	//    |  sb1           sb0  |           sb0  sb1
	//    |  ...           ...  |           ...   ...
	//    |  pb1           pb0  |           pb0  pb1
	//    |  /               \  |            \   /
	//     b                   b               b

	for _, b := range f.Blocks {
		if len(b.Preds) != 2 || len(b.Values) == 0 {
			// TODO: handle more than 2 predecessors, e.g. a || b || c.
			continue
		}
		find := false
		var v *Value
		for _, v = range b.Values {
			// find a phi value v = OpPhi (ConstBool [true]) (ConstBool [false]).
			// TODO: v = OpPhi (ConstBool [true]) (Arg <bool> {value})
			if v.Op == OpPhi && v.Args[0].Op == OpConstBool && v.Args[1].Op == OpConstBool {
				find = true
				break
			}
		}
		if !find {
			continue
		}
		pb0 := b.Preds[0].b
		pb1 := b.Preds[1].b
		if pb0.Kind == BlockIf && pb0 == sdom.Parent(b) {
			// special case: pb0 is the dominator block b0.
			//     b0  ...
			//    |  \ /
			//    |  sb1
			//    |  ...
			//    |  pb1
			//    |  /
			//     b
			b0 := pb0
			pb0 := b
			p := sdom.Parent(pb1)
			for ; p != nil; p = sdom.Parent(pb1) {
				if p == b0 {
					replace(b0, pb0, pb1, b, v)
					break
				}
				pb1 = p
			}
		} else if pb1.Kind == BlockIf && pb1 == sdom.Parent(b) {
			// special case: pb1 is the dominator block b0.
			//  ...   b0
			//    \ /  |
			//    sb0  |
			//    ...  |
			//    pb0  |
			//      \  |
			//        b
			b0 := pb1
			pb1 := b
			p := sdom.Parent(pb0)
			for ; p != nil; p = sdom.Parent(pb0) {
				if p == b0 {
					replace(b0, pb0, pb1, b, v)
					break
				}
				pb0 = p
			}
		} else {
			//      b0
			//     /   \
			//    sb0  sb1
			//    ...  ...
			//    pb0  pb1
			//      \   /
			//        b
			p0 := sdom.Parent(pb0)
			b1 := pb1
		outer:
			for ; p0 != nil; p0 = sdom.Parent(pb0) {
				if p0.Kind == BlockIf {
					p1 := sdom.Parent(pb1)
					for ; p1 != nil; p1 = sdom.Parent(pb1) {
						if p0 == p1 {
							sb0 := p0.Succs[0].b
							sb1 := p0.Succs[1].b
							// we can not replace phi value x in the following case.
							//   if gp == nil || sp < lo { x = true}
							//   if a || b { x = true }
							// so the if statement can only have one condition.
							if len(sb0.Preds) == 1 && len(sb1.Preds) == 1 {
								replace(p0, pb0, pb1, b, v)
								break outer
							}
						}
						pb1 = p1
					}
					pb1 = b1
				}
				pb0 = p0
			}
		}
	}
}

func phioptint(v *Value, b0 *Block, reverse int) {
	a0 := v.Args[0]
	a1 := v.Args[1]
	if a0.Op != a1.Op {
		return
	}

	switch a0.Op {
	case OpConst8, OpConst16, OpConst32, OpConst64:
	default:
		return
	}

	negate := false
	switch {
	case a0.AuxInt == 0 && a1.AuxInt == 1:
		negate = true
	case a0.AuxInt == 1 && a1.AuxInt == 0:
	default:
		return
	}

	if reverse == 1 {
		negate = !negate
	}

	a := b0.Controls[0]
	if negate {
		a = v.Block.NewValue1(v.Pos, OpNot, a.Type, a)
	}
	v.AddArg(a)

	cvt := v.Block.NewValue1(v.Pos, OpCvtBoolToUint8, v.Block.Func.Config.Types.UInt8, a)
	switch v.Type.Size() {
	case 1:
		v.reset(OpCopy)
	case 2:
		v.reset(OpZeroExt8to16)
	case 4:
		v.reset(OpZeroExt8to32)
	case 8:
		v.reset(OpZeroExt8to64)
	default:
		v.Fatalf("bad int size %d", v.Type.Size())
	}
	v.AddArg(cvt)

	f := b0.Func
	if f.pass.debug > 0 {
		f.Warnl(v.Block.Pos, "converted OpPhi bool -> int%d", v.Type.Size()*8)
	}
}

func replace(b0, pb0, pb1, b *Block, v *Value) {
	// b0 is the if block giving the boolean value.
	// b has the phi value v = (OpPhi (ConstBool [true]) (ConstBool [false])).
	// reverse is the predecessor from which the truth value comes.
	var reverse int
	f := b.Func
	if b0.Succs[0].b == pb0 && b0.Succs[1].b == pb1 {
		reverse = 0
	} else if b0.Succs[0].b == pb1 && b0.Succs[1].b == pb0 {
		reverse = 1
	} else {
		return
	}
	if v.Args[reverse].AuxInt != v.Args[1-reverse].AuxInt {
		ops := [2]Op{OpNot, OpCopy}
		v.reset(ops[v.Args[reverse].AuxInt])
		v.AddArg(b0.Controls[0])
		if f.pass.debug > 0 {
			f.Warnl(b.Pos, "converted OpPhi to %v", v.Op)
		}
	}
}

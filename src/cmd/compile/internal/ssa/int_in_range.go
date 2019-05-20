package ssa

import (
	"cmd/compile/internal/types"
)

// mergeBlockIntoPredecessor applies this CFG transformation:
//
//   p            p*
//   |\           |\
//   | b     ->   | \
//   |/ \         |  \
//   x   y        x*  y*
//
// All phis in x must have the same argument for the edges from p and b.
// It is the caller's responsibility to call invalidateCFG.
func mergeBlockIntoPredecessor(b *Block) {
	if len(b.Preds) != 1 {
		panic("block must have exactly 1 predecessor")
	}
	if len(b.Succs) != 2 {
		panic("block must have exactly 2 successors")
	}
	p := b.Preds[0].b
	if len(p.Succs) != 2 {
		panic("predecessor must have exactly 2 successors")
	}

	// Step 1: remove the edge from p to x and update phis in x:
	//
	//   p            p*
	//   |\            \
	//   | b     ->     b
	//   |/ \          / \
	//   x   y        x*  y
	//
	ix := b.Preds[0].i ^ 1 // index of common successor x
	px := p.Succs[ix]      // edge from p to x
	bx := b.Succs[ix]      // edge from b to x (for assertions)
	x := px.b
	if bx.b != x {
		panic("no common successor or common successor is not in same direction")
	}
	for _, v := range x.Values {
		if v.Op != OpPhi {
			continue
		}
		if v.Args[bx.i] != v.Args[px.i] {
			panic("phi invalidated by merge")
		}
		v.Args[px.i].Uses -= 1
		v.Args[px.i] = v.Args[len(v.Args)-1]
		v.Args = v.Args[:len(v.Args)-1]
	}
	x.removePred(px.i)

	// Step 2: replace the edge leaving p with the edges leaving b:
	//
	//   p*
	//   |\
	//   | \   b*
	//   |  \
	//   x*  y*
	//
	p.Succs = append(p.Succs[:0], b.Succs...)
	for i := range p.Succs {
		e := p.Succs[i]
		e.b.Preds[e.i].b = p
	}

	// Step 3: move all values from b to p.
	// TODO: fuseBlockPlain also does this, with important optimizations.
	// Share code with fuseBlockPlain to do this.
	for _, v := range b.Values {
		v.Block = p
	}
	p.Values = append(p.Values, b.Values...)

	// Step 4: clobber b.
	clobberBlock(b)
}

// fuseBlockIntInRange optimizes 1 <= x && x < 5 to unsigned(x-1) < 4.
//
// Look for branch structure like:
//
//   p
//   |\
//   | b
//   |/ \
//   s0 s1
//
// In our example, p has control '1 <= x', b has control 'x < 5',
// and s0 and s1 are the if and else results of the comparison.
//
// This will be optimized into:
//
//   c
//   | \
//  s0 s1
//
// where c has the combined control value.
func fuseBlockIntInRange(b *Block) bool {
	if len(b.Preds) != 1 || b.Kind != BlockIf {
		return false
	}
	p := b.Preds[0].b
	if p.Kind != BlockIf {
		return false
	}

	var negp, negb bool // whether to negate the control value of p/b during analysis

	// Find the non-b successor edge from p.
	pe := p.Succs[1]
	if pe.b == b {
		pe = p.Succs[0]
		negp = true
	}
	// Try to find a successor edge of b to the same block (s0 in the diagram).
	be := b.Succs[1]
	if be.b != pe.b {
		be = b.Succs[0]
		negb = true
	}
	if be.b != pe.b {
		return false
	}
	s0 := be.b
	// Check that all phi values in s0 match when reached via p/b.
	for _, v := range s0.Values {
		if v.Op == OpPhi && v.Args[pe.i] != v.Args[be.i] {
			return false
		}
	}

	// To get to s0, negp(p.Controls[0]) && negb(b.Controls[0]) must be true.

	// This is cheaper than a map lookup and lets us quickly rule out
	// many non-comparison values.
	if len(p.Controls[0].Args) != 2 || len(b.Controls[0].Args) != 2 {
		return false
	}

	// Check whether both control values are comparisons.
	opl := cmpOp[p.Controls[0].Op]
	if opl.op == 0 {
		return false
	}
	opr := cmpOp[b.Controls[0].Op]
	if opr.op == 0 {
		return false
	}

	if opl.bitwidth != opr.bitwidth {
		return false
	}
	bw := opl.bitwidth

	// Optimize integer-in-range checks, such as 4 <= x && x < 10.
	p0, p1 := p.Controls[0].Args[0], p.Controls[0].Args[1]
	b0, b1 := b.Controls[0].Args[0], b.Controls[0].Args[1]

	// We are looking for something equivalent to lo op1 x && x op2 hi, where op1 and op2 are < or ≤.

	// Find x, if it exists, and rename appropriately.
	// TODO Input is: p0 p.Controls[0].Op p1 && b0 r.Op r.Right
	// TODO Output is: a opl b(==x) ANDAND/OROR b(==x) opr c
	w, x := p0, p1
	x1, z := b0, b1
	for i := 0; ; i++ {
		if x == x1 {
			break
		}
		if i == 3 {
			// Tried all permutations and couldn't find an appropriate b == x.
			return false
		}
		if i&1 == 0 {
			w, opl, x = x, opl.Reverse(), w
		} else {
			x1, opr, z = z, opr.Reverse(), x1
		}
	}

	// We must be able to ensure that c-a is non-negative.
	// For now, require a and c to be constants.
	if w.Op != bw.constOp || z.Op != bw.constOp {
		return false
	}

	if negp {
		opl = opl.Negate()
	}
	if negb {
		opr = opr.Negate()
	}

	if opl.direction != opr.direction {
		// Not a range check; something like b < a && b < c.
		return false
	}

	if opl.direction == 1 {
		// We have something like a > b && b ≥ c.
		// Switch and reverse ops and rename constants,
		// to make it look like a ≤ b && b < c.
		w, z = z, w
		opl, opr = opr.Reverse(), opl.Reverse()
	}

	lo := bw.trunc(w.AuxInt)
	hi := bw.trunc(z.AuxInt)
	if opl.op == bw.lessOp {
		// We have a < b && ...
		// We need a ≤ b && ... to safely use unsigned comparison tricks.
		// If a is not the maximum constant for b's type,
		// we can increment a and assume ≤.
		if (bw.signed && lo >= bw.maxint) || (!bw.signed && uint64(lo) >= bw.umaxint) {
			return false
		}
		lo++
	}

	if (bw.signed && hi > lo) || (!bw.signed && uint64(hi) > uint64(lo)) {
		// Bad news. Something like 5 <= x && x < 3.
		// Probably a bug, and thus rare...just leave it alone.
		return false
	}
	bound := hi - lo

	// We have w ≤ x && x < z (or w ≤ x && x ≤ z).
	// This is equivalent to (w-w) ≤ (x-w) && (x-w) < (z-w),
	// which is equivalent to 0 ≤ (x-w) && (x-w) < (z-w),
	// which is equivalent to uint(x-w) < uint(z-w).

	mergeBlockIntoPredecessor(b)

	ut := x.Type.ToUnsigned()
	delta := x
	if lo != 0 {
		delta = p.NewValue2(x.Pos, bw.subOp, ut, x, bw.constInt(b, ut, lo))
	}
	if negb {
		opr = opr.Negate()
	}
	c := p.NewValue2(x.Pos, opr.unsigned, b.Func.Config.Types.Bool, delta, bw.constInt(b, ut, bound))
	p.SetControl(c)

	return true
}

// cmpInfo holds information about a comparison Op.
type cmpInfo struct {
	op        Op // the op this info is about
	reverse   Op // reversed direction of op, e.g. ≥ for ≤
	negate    Op // op with negated semantics, e.g. > for ≤
	unsigned  Op // unsigned flavor of op
	bitwidth  *bitwidthInfo
	direction int // direction the op "points", e.g. -1 for < and 1 for >
}

// TODO: unsigned ops

var cmpOp = map[Op]cmpInfo{
	OpLess8:    {op: OpLess8, reverse: OpGreater8, negate: OpGeq8, unsigned: OpLess8U, bitwidth: bitwidth8, direction: -1},
	OpLeq8:     {op: OpLeq8, reverse: OpGeq8, negate: OpGreater8, unsigned: OpLeq8U, bitwidth: bitwidth8, direction: -1},
	OpGreater8: {op: OpGreater8, reverse: OpLess8, negate: OpLeq8, unsigned: OpGreater8U, bitwidth: bitwidth8, direction: 1},
	OpGeq8:     {op: OpGeq8, reverse: OpLeq8, negate: OpLess8, unsigned: OpGeq8U, bitwidth: bitwidth8, direction: 1},

	OpLess16:    {op: OpLess16, reverse: OpGreater16, negate: OpGeq16, unsigned: OpLess16U, bitwidth: bitwidth16, direction: -1},
	OpLeq16:     {op: OpLeq16, reverse: OpGeq16, negate: OpGreater16, unsigned: OpLeq16U, bitwidth: bitwidth16, direction: -1},
	OpGreater16: {op: OpGreater16, reverse: OpLess16, negate: OpLeq16, unsigned: OpGreater16U, bitwidth: bitwidth16, direction: 1},
	OpGeq16:     {op: OpGeq16, reverse: OpLeq16, negate: OpLess16, unsigned: OpGeq16U, bitwidth: bitwidth16, direction: 1},

	OpLess32:    {op: OpLess32, reverse: OpGreater32, negate: OpGeq32, unsigned: OpLess32U, bitwidth: bitwidth32, direction: -1},
	OpLeq32:     {op: OpLeq32, reverse: OpGeq32, negate: OpGreater32, unsigned: OpLeq32U, bitwidth: bitwidth32, direction: -1},
	OpGreater32: {op: OpGreater32, reverse: OpLess32, negate: OpLeq32, unsigned: OpGreater32U, bitwidth: bitwidth32, direction: 1},
	OpGeq32:     {op: OpGeq32, reverse: OpLeq32, negate: OpLess32, unsigned: OpGeq32U, bitwidth: bitwidth32, direction: 1},

	OpLess64:    {op: OpLess64, reverse: OpGreater64, negate: OpGeq64, unsigned: OpLess64U, bitwidth: bitwidth64, direction: -1},
	OpLeq64:     {op: OpLeq64, reverse: OpGeq64, negate: OpGreater64, unsigned: OpLeq64U, bitwidth: bitwidth64, direction: -1},
	OpGreater64: {op: OpGreater64, reverse: OpLess64, negate: OpLeq64, unsigned: OpGreater64U, bitwidth: bitwidth64, direction: 1},
	OpGeq64:     {op: OpGeq64, reverse: OpLeq64, negate: OpLess64, unsigned: OpGeq64U, bitwidth: bitwidth64, direction: 1},

	OpLess8U:    {op: OpLess8U, reverse: OpGreater8U, negate: OpGeq8U, unsigned: OpLess8U, bitwidth: bitwidth8U, direction: -1},
	OpLeq8U:     {op: OpLeq8U, reverse: OpGeq8U, negate: OpGreater8U, unsigned: OpLeq8U, bitwidth: bitwidth8U, direction: -1},
	OpGreater8U: {op: OpGreater8U, reverse: OpLess8U, negate: OpLeq8U, unsigned: OpGreater8U, bitwidth: bitwidth8U, direction: 1},
	OpGeq8U:     {op: OpGeq8U, reverse: OpLeq8U, negate: OpLess8U, unsigned: OpGeq8U, bitwidth: bitwidth8U, direction: 1},

	OpLess16U:    {op: OpLess16U, reverse: OpGreater16U, negate: OpGeq16U, unsigned: OpLess16U, bitwidth: bitwidth16U, direction: -1},
	OpLeq16U:     {op: OpLeq16U, reverse: OpGeq16U, negate: OpGreater16U, unsigned: OpLeq16U, bitwidth: bitwidth16U, direction: -1},
	OpGreater16U: {op: OpGreater16U, reverse: OpLess16U, negate: OpLeq16U, unsigned: OpGreater16U, bitwidth: bitwidth16U, direction: 1},
	OpGeq16U:     {op: OpGeq16U, reverse: OpLeq16U, negate: OpLess16U, unsigned: OpGeq16U, bitwidth: bitwidth16U, direction: 1},

	OpLess32U:    {op: OpLess32U, reverse: OpGreater32U, negate: OpGeq32U, unsigned: OpLess32U, bitwidth: bitwidth32U, direction: -1},
	OpLeq32U:     {op: OpLeq32U, reverse: OpGeq32U, negate: OpGreater32U, unsigned: OpLeq32U, bitwidth: bitwidth32U, direction: -1},
	OpGreater32U: {op: OpGreater32U, reverse: OpLess32U, negate: OpLeq32U, unsigned: OpGreater32U, bitwidth: bitwidth32U, direction: 1},
	OpGeq32U:     {op: OpGeq32U, reverse: OpLeq32U, negate: OpLess32U, unsigned: OpGeq32U, bitwidth: bitwidth32U, direction: 1},

	OpLess64U:    {op: OpLess64U, reverse: OpGreater64U, negate: OpGeq64U, unsigned: OpLess64U, bitwidth: bitwidth64U, direction: -1},
	OpLeq64U:     {op: OpLeq64U, reverse: OpGeq64U, negate: OpGreater64U, unsigned: OpLeq64U, bitwidth: bitwidth64U, direction: -1},
	OpGreater64U: {op: OpGreater64U, reverse: OpLess64U, negate: OpLeq64U, unsigned: OpGreater64U, bitwidth: bitwidth64U, direction: 1},
	OpGeq64U:     {op: OpGeq64U, reverse: OpLeq64U, negate: OpLess64U, unsigned: OpGeq64U, bitwidth: bitwidth64U, direction: 1},
}

func (c cmpInfo) Reverse() cmpInfo {
	return cmpOp[c.reverse]
}

func (c cmpInfo) Negate() cmpInfo {
	return cmpOp[c.negate]
}

type bitwidthInfo struct {
	width    int    // width in bits
	signed   bool   // whether this bitwidth is signed
	maxint   int64  // maximum integer that fits in this width, for signed bitwidths
	umaxint  uint64 // maximum unsigned integer that fits in this width, for unsigned bitwidths
	subOp    Op     // subtraction op
	constOp  Op     // constant op
	lessOp   Op     // < op
	constInt func(b *Block, t *types.Type, n int64) *Value
	trunc    func(x int64) int64 // appropriately truncates AuxInt x
}

var (
	bitwidth8 = &bitwidthInfo{
		width:    8,
		signed:   false,
		maxint:   0x7f,
		subOp:    OpSub8,
		constOp:  OpConst8,
		lessOp:   OpLess8,
		constInt: func(b *Block, t *types.Type, n int64) *Value { return b.Func.ConstInt8(t, int8(n)) },
		trunc:    func(x int64) int64 { return int64(int8(x)) },
	}

	bitwidth16 = &bitwidthInfo{
		width:    16,
		signed:   false,
		maxint:   0x7fff,
		subOp:    OpSub16,
		constOp:  OpConst16,
		lessOp:   OpLess16,
		constInt: func(b *Block, t *types.Type, n int64) *Value { return b.Func.ConstInt16(t, int16(n)) },
		trunc:    func(x int64) int64 { return int64(int16(x)) },
	}

	bitwidth32 = &bitwidthInfo{
		width:    32,
		signed:   false,
		maxint:   0x7fffffff,
		subOp:    OpSub32,
		constOp:  OpConst32,
		lessOp:   OpLess32,
		constInt: func(b *Block, t *types.Type, n int64) *Value { return b.Func.ConstInt32(t, int32(n)) },
		trunc:    func(x int64) int64 { return int64(int32(x)) },
	}

	bitwidth64 = &bitwidthInfo{
		width:    64,
		signed:   false,
		maxint:   0x7fffffffffffffff,
		subOp:    OpSub64,
		constOp:  OpConst64,
		lessOp:   OpLess64,
		constInt: func(b *Block, t *types.Type, n int64) *Value { return b.Func.ConstInt64(t, n) },
		trunc:    func(x int64) int64 { return int64(x) },
	}

	bitwidth8U = &bitwidthInfo{
		width:    8,
		signed:   true,
		umaxint:  0xff,
		subOp:    OpSub8,
		constOp:  OpConst8,
		lessOp:   OpLess8U,
		constInt: func(b *Block, t *types.Type, n int64) *Value { return b.Func.ConstInt8(t, int8(uint8(n))) },
		trunc:    func(x int64) int64 { return int64(uint8(x)) },
	}

	bitwidth16U = &bitwidthInfo{
		width:    16,
		signed:   true,
		umaxint:  0xffff,
		subOp:    OpSub16,
		constOp:  OpConst16,
		lessOp:   OpLess16U,
		constInt: func(b *Block, t *types.Type, n int64) *Value { return b.Func.ConstInt16(t, int16(uint16(n))) },
		trunc:    func(x int64) int64 { return int64(uint16(x)) },
	}

	bitwidth32U = &bitwidthInfo{
		width:    32,
		signed:   true,
		umaxint:  0xffffffff,
		subOp:    OpSub32,
		constOp:  OpConst32,
		lessOp:   OpLess32U,
		constInt: func(b *Block, t *types.Type, n int64) *Value { return b.Func.ConstInt32(t, int32(uint32(n))) },
		trunc:    func(x int64) int64 { return int64(uint32(x)) },
	}

	bitwidth64U = &bitwidthInfo{
		width:    64,
		signed:   true,
		umaxint:  0xffffffffffffffff,
		subOp:    OpSub64,
		constOp:  OpConst64,
		lessOp:   OpLess64U,
		constInt: func(b *Block, t *types.Type, n int64) *Value { return b.Func.ConstInt64(t, n) },
		trunc:    func(x int64) int64 { return x },
	}
)

func init() {
	// Sanity check cmpOp.
	for op, info := range cmpOp {
		if op != cmpOp[info.reverse].reverse {
			panic("bad reverse")
		}
		if op != cmpOp[info.negate].negate {
			panic("bad reverse")
		}
	}
}

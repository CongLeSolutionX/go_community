// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gc

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/ir"
	"cmd/compile/internal/types"
	"cmd/internal/src"
	"fmt"
	"math/big"
	"strings"
)

// truncate float literal fv to 32-bit or 64-bit precision
// according to type; return truncated value.
func truncfltlit(oldv *ir.Float, t *types.Type) *ir.Float {
	if t == nil {
		return oldv
	}

	if overflow(ir.Val{U: oldv}, t) {
		// If there was overflow, simply continuing would set the
		// value to Inf which in turn would lead to spurious follow-on
		// errors. Avoid this by returning the existing value.
		return oldv
	}

	fv := ir.NewFloat()

	// convert large precision literal floating
	// into limited precision (float64 or float32)
	switch t.Etype {
	case types.TFLOAT32:
		fv.SetFloat64(oldv.Float32())
	case types.TFLOAT64:
		fv.SetFloat64(oldv.Float64())
	default:
		base.Fatal("truncfltlit: unexpected Etype %v", t.Etype)
	}

	return fv
}

// truncate Real and Imag parts of Mpcplx to 32-bit or 64-bit
// precision, according to type; return truncated value. In case of
// overflow, calls yyerror but does not truncate the input value.
func trunccmplxlit(oldv *ir.Complex, t *types.Type) *ir.Complex {
	if t == nil {
		return oldv
	}

	if overflow(ir.Val{U: oldv}, t) {
		// If there was overflow, simply continuing would set the
		// value to Inf which in turn would lead to spurious follow-on
		// errors. Avoid this by returning the existing value.
		return oldv
	}

	cv := ir.NewComplex()

	switch t.Etype {
	case types.TCOMPLEX64:
		cv.Real.SetFloat64(oldv.Real.Float32())
		cv.Imag.SetFloat64(oldv.Imag.Float32())
	case types.TCOMPLEX128:
		cv.Real.SetFloat64(oldv.Real.Float64())
		cv.Imag.SetFloat64(oldv.Imag.Float64())
	default:
		base.Fatal("trunccplxlit: unexpected Etype %v", t.Etype)
	}

	return cv
}

// TODO(mdempsky): Replace these with better APIs.
func convlit(n *ir.Node, t *types.Type) *ir.Node    { return convlit1(n, t, false, nil) }
func defaultlit(n *ir.Node, t *types.Type) *ir.Node { return convlit1(n, t, false, nil) }

// convlit1 converts an untyped expression n to type t. If n already
// has a type, convlit1 has no effect.
//
// For explicit conversions, t must be non-nil, and integer-to-string
// conversions are allowed.
//
// For implicit conversions (e.g., assignments), t may be nil; if so,
// n is converted to its default type.
//
// If there's an error converting n to t, context is used in the error
// message.
func convlit1(n *ir.Node, t *types.Type, explicit bool, context func() string) *ir.Node {
	if explicit && t == nil {
		base.Fatal("explicit conversion missing type")
	}
	if t != nil && t.IsUntyped() {
		base.Fatal("bad conversion to untyped: %v", t)
	}

	if n == nil || n.Type == nil {
		// Allow sloppy callers.
		return n
	}
	if !n.Type.IsUntyped() {
		// Already typed; nothing to do.
		return n
	}

	if n.Op == ir.OLITERAL {
		// Can't always set n.Type directly on OLITERAL nodes.
		// See discussion on CL 20813.
		n = n.RawCopy()
	}

	// Nil is technically not a constant, so handle it specially.
	if n.Type.Etype == types.TNIL {
		if t == nil {
			base.Error("use of untyped nil")
			n.SetDiag(true)
			n.Type = nil
			return n
		}

		if !t.HasNil() {
			// Leave for caller to handle.
			return n
		}

		n.Type = t
		return n
	}

	if t == nil || !okforconst[t.Etype] {
		t = defaultType(n.Type)
	}

	switch n.Op {
	default:
		base.Fatal("unexpected untyped expression: %v", n)

	case ir.OLITERAL:
		v := convertVal(n.Val(), t, explicit)
		if v.U == nil {
			break
		}
		n.SetVal(v)
		n.Type = t
		return n

	case ir.OPLUS, ir.ONEG, ir.OBITNOT, ir.ONOT, ir.OREAL, ir.OIMAG:
		ot := operandType(n.Op, t)
		if ot == nil {
			n = defaultlit(n, nil)
			break
		}

		n.Left = convlit(n.Left, ot)
		if n.Left.Type == nil {
			n.Type = nil
			return n
		}
		n.Type = t
		return n

	case ir.OADD, ir.OSUB, ir.OMUL, ir.ODIV, ir.OMOD, ir.OOR, ir.OXOR, ir.OAND, ir.OANDNOT, ir.OOROR, ir.OANDAND, ir.OCOMPLEX:
		ot := operandType(n.Op, t)
		if ot == nil {
			n = defaultlit(n, nil)
			break
		}

		n.Left = convlit(n.Left, ot)
		n.Right = convlit(n.Right, ot)
		if n.Left.Type == nil || n.Right.Type == nil {
			n.Type = nil
			return n
		}
		if !types.Identical(n.Left.Type, n.Right.Type) {
			base.Error("invalid operation: %v (mismatched types %v and %v)", n, n.Left.Type, n.Right.Type)
			n.Type = nil
			return n
		}

		n.Type = t
		return n

	case ir.OEQ, ir.ONE, ir.OLT, ir.OLE, ir.OGT, ir.OGE:
		if !t.IsBoolean() {
			break
		}
		n.Type = t
		return n

	case ir.OLSH, ir.ORSH:
		n.Left = convlit1(n.Left, t, explicit, nil)
		n.Type = n.Left.Type
		if n.Type != nil && !n.Type.IsInteger() {
			base.Error("invalid operation: %v (shift of type %v)", n, n.Type)
			n.Type = nil
		}
		return n
	}

	if !n.Diag() {
		if !t.Broke() {
			if explicit {
				base.Error("cannot convert %L to type %v", n, t)
			} else if context != nil {
				base.Error("cannot use %L as type %v in %s", n, t, context())
			} else {
				base.Error("cannot use %L as type %v", n, t)
			}
		}
		n.SetDiag(true)
	}
	n.Type = nil
	return n
}

func operandType(op ir.Op, t *types.Type) *types.Type {
	switch op {
	case ir.OCOMPLEX:
		if t.IsComplex() {
			return floatForComplex(t)
		}
	case ir.OREAL, ir.OIMAG:
		if t.IsFloat() {
			return complexForFloat(t)
		}
	default:
		if okfor[op][t.Etype] {
			return t
		}
	}
	return nil
}

// convertVal converts v into a representation appropriate for t. If
// no such representation exists, it returns Val{} instead.
//
// If explicit is true, then conversions from integer to string are
// also allowed.
func convertVal(v ir.Val, t *types.Type, explicit bool) ir.Val {
	switch ct := v.Ctype(); ct {
	case ir.CTBOOL:
		if t.IsBoolean() {
			return v
		}

	case ir.CTSTR:
		if t.IsString() {
			return v
		}

	case ir.CTINT, ir.CTRUNE:
		if explicit && t.IsString() {
			return tostr(v)
		}
		fallthrough
	case ir.CTFLT, ir.CTCPLX:
		switch {
		case t.IsInteger():
			v = toint(v)
			overflow(v, t)
			return v
		case t.IsFloat():
			v = toflt(v)
			v = ir.Val{U: truncfltlit(v.U.(*ir.Float), t)}
			return v
		case t.IsComplex():
			v = tocplx(v)
			v = ir.Val{U: trunccmplxlit(v.U.(*ir.Complex), t)}
			return v
		}
	}

	return ir.Val{}
}

func tocplx(v ir.Val) ir.Val {
	switch u := v.U.(type) {
	case *ir.Int:
		c := ir.NewComplex()
		c.Real.SetInt(u)
		c.Imag.SetFloat64(0.0)
		v.U = c

	case *ir.Float:
		c := ir.NewComplex()
		c.Real.Set(u)
		c.Imag.SetFloat64(0.0)
		v.U = c
	}

	return v
}

func toflt(v ir.Val) ir.Val {
	switch u := v.U.(type) {
	case *ir.Int:
		f := ir.NewFloat()
		f.SetInt(u)
		v.U = f

	case *ir.Complex:
		f := ir.NewFloat()
		f.Set(&u.Real)
		if u.Imag.CmpFloat64(0) != 0 {
			base.Error("constant %v truncated to real", u.GoString())
		}
		v.U = f
	}

	return v
}

func toint(v ir.Val) ir.Val {
	switch u := v.U.(type) {
	case *ir.Int:
		if u.Rune {
			i := new(ir.Int)
			i.Set(u)
			v.U = i
		}

	case *ir.Float:
		i := new(ir.Int)
		if !i.SetFloat(u) {
			if i.CheckOverflow(0) {
				base.Error("integer too large")
			} else {
				// The value of u cannot be represented as an integer;
				// so we need to print an error message.
				// Unfortunately some float values cannot be
				// reasonably formatted for inclusion in an error
				// message (example: 1 + 1e-100), so first we try to
				// format the float; if the truncation resulted in
				// something that looks like an integer we omit the
				// value from the error message.
				// (See issue #11371).
				var t big.Float
				t.Parse(u.GoString(), 10)
				if t.IsInt() {
					base.Error("constant truncated to integer")
				} else {
					base.Error("constant %v truncated to integer", u.GoString())
				}
			}
		}
		v.U = i

	case *ir.Complex:
		i := new(ir.Int)
		if !i.SetFloat(&u.Real) || u.Imag.CmpFloat64(0) != 0 {
			base.Error("constant %v truncated to integer", u.GoString())
		}

		v.U = i
	}

	return v
}

func doesoverflow(v ir.Val, t *types.Type) bool {
	switch u := v.U.(type) {
	case *ir.Int:
		if !t.IsInteger() {
			base.Fatal("overflow: %v integer constant", t)
		}
		return u.Cmp(minintval[t.Etype]) < 0 || u.Cmp(maxintval[t.Etype]) > 0

	case *ir.Float:
		if !t.IsFloat() {
			base.Fatal("overflow: %v floating-point constant", t)
		}
		return u.Cmp(minfltval[t.Etype]) <= 0 || u.Cmp(maxfltval[t.Etype]) >= 0

	case *ir.Complex:
		if !t.IsComplex() {
			base.Fatal("overflow: %v complex constant", t)
		}
		return u.Real.Cmp(minfltval[t.Etype]) <= 0 || u.Real.Cmp(maxfltval[t.Etype]) >= 0 ||
			u.Imag.Cmp(minfltval[t.Etype]) <= 0 || u.Imag.Cmp(maxfltval[t.Etype]) >= 0
	}

	return false
}

func overflow(v ir.Val, t *types.Type) bool {
	// v has already been converted
	// to appropriate form for t.
	if t == nil || t.Etype == types.TIDEAL {
		return false
	}

	// Only uintptrs may be converted to pointers, which cannot overflow.
	if t.IsPtr() || t.IsUnsafePtr() {
		return false
	}

	if doesoverflow(v, t) {
		base.Error("constant %v overflows %v", v, t)
		return true
	}

	return false

}

func tostr(v ir.Val) ir.Val {
	switch u := v.U.(type) {
	case *ir.Int:
		var r rune = 0xFFFD
		if u.Cmp(minintval[types.TINT32]) >= 0 && u.Cmp(maxintval[types.TINT32]) <= 0 {
			r = rune(u.Int64())
		}
		v.U = string(r)
	}

	return v
}

// evconst rewrites constant expressions into OLITERAL nodes.
func evconst(n *ir.Node) {
	nl, nr := n.Left, n.Right

	// Pick off just the opcodes that can be constant evaluated.
	switch op := n.Op; op {
	case ir.OPLUS, ir.ONEG, ir.OBITNOT, ir.ONOT:
		if nl.Op == ir.OLITERAL {
			setconst(n, unaryOp(op, nl.Val(), n.Type))
		}

	case ir.OADD, ir.OSUB, ir.OMUL, ir.ODIV, ir.OMOD, ir.OOR, ir.OXOR, ir.OAND, ir.OANDNOT, ir.OOROR, ir.OANDAND:
		if nl.Op == ir.OLITERAL && nr.Op == ir.OLITERAL {
			setconst(n, binaryOp(nl.Val(), op, nr.Val()))
		}

	case ir.OEQ, ir.ONE, ir.OLT, ir.OLE, ir.OGT, ir.OGE:
		if nl.Op == ir.OLITERAL && nr.Op == ir.OLITERAL {
			setboolconst(n, compareOp(nl.Val(), op, nr.Val()))
		}

	case ir.OLSH, ir.ORSH:
		if nl.Op == ir.OLITERAL && nr.Op == ir.OLITERAL {
			setconst(n, shiftOp(nl.Val(), op, nr.Val()))
		}

	case ir.OCONV, ir.ORUNESTR:
		if okforconst[n.Type.Etype] && nl.Op == ir.OLITERAL {
			setconst(n, convertVal(nl.Val(), n.Type, true))
		}

	case ir.OCONVNOP:
		if okforconst[n.Type.Etype] && nl.Op == ir.OLITERAL {
			// set so n.Orig gets OCONV instead of OCONVNOP
			n.Op = ir.OCONV
			setconst(n, nl.Val())
		}

	case ir.OADDSTR:
		// Merge adjacent constants in the argument list.
		s := n.List.Slice()
		for i1 := 0; i1 < len(s); i1++ {
			if ir.IsConst(s[i1], ir.CTSTR) && i1+1 < len(s) && ir.IsConst(s[i1+1], ir.CTSTR) {
				// merge from i1 up to but not including i2
				var strs []string
				i2 := i1
				for i2 < len(s) && ir.IsConst(s[i2], ir.CTSTR) {
					strs = append(strs, s[i2].StringVal())
					i2++
				}

				nl := *s[i1]
				nl.Orig = &nl
				nl.SetVal(ir.Val{U: strings.Join(strs, "")})
				s[i1] = &nl
				s = append(s[:i1+1], s[i2:]...)
			}
		}

		if len(s) == 1 && ir.IsConst(s[0], ir.CTSTR) {
			n.Op = ir.OLITERAL
			n.SetVal(s[0].Val())
		} else {
			n.List.Set(s)
		}

	case ir.OCAP, ir.OLEN:
		switch nl.Type.Etype {
		case types.TSTRING:
			if ir.IsConst(nl, ir.CTSTR) {
				setintconst(n, int64(len(nl.StringVal())))
			}
		case types.TARRAY:
			if !hascallchan(nl) {
				setintconst(n, nl.Type.NumElem())
			}
		}

	case ir.OALIGNOF, ir.OOFFSETOF, ir.OSIZEOF:
		setintconst(n, evalunsafe(n))

	case ir.OREAL, ir.OIMAG:
		if nl.Op == ir.OLITERAL {
			var re, im *ir.Float
			switch u := nl.Val().U.(type) {
			case *ir.Int:
				re = ir.NewFloat()
				re.SetInt(u)
				// im = 0
			case *ir.Float:
				re = u
				// im = 0
			case *ir.Complex:
				re = &u.Real
				im = &u.Imag
			default:
				base.Fatal("impossible")
			}
			if n.Op == ir.OIMAG {
				if im == nil {
					im = ir.NewFloat()
				}
				re = im
			}
			setconst(n, ir.Val{U: re})
		}

	case ir.OCOMPLEX:
		if nl.Op == ir.OLITERAL && nr.Op == ir.OLITERAL {
			// make it a complex literal
			c := ir.NewComplex()
			c.Real.Set(toflt(nl.Val()).U.(*ir.Float))
			c.Imag.Set(toflt(nr.Val()).U.(*ir.Float))
			setconst(n, ir.Val{U: c})
		}
	}
}

func match(x, y ir.Val) (ir.Val, ir.Val) {
	switch {
	case x.Ctype() == ir.CTCPLX || y.Ctype() == ir.CTCPLX:
		return tocplx(x), tocplx(y)
	case x.Ctype() == ir.CTFLT || y.Ctype() == ir.CTFLT:
		return toflt(x), toflt(y)
	}

	// Mixed int/rune are fine.
	return x, y
}

func compareOp(x ir.Val, op ir.Op, y ir.Val) bool {
	x, y = match(x, y)

	switch x.Ctype() {
	case ir.CTBOOL:
		x, y := x.U.(bool), y.U.(bool)
		switch op {
		case ir.OEQ:
			return x == y
		case ir.ONE:
			return x != y
		}

	case ir.CTINT, ir.CTRUNE:
		x, y := x.U.(*ir.Int), y.U.(*ir.Int)
		return cmpZero(x.Cmp(y), op)

	case ir.CTFLT:
		x, y := x.U.(*ir.Float), y.U.(*ir.Float)
		return cmpZero(x.Cmp(y), op)

	case ir.CTCPLX:
		x, y := x.U.(*ir.Complex), y.U.(*ir.Complex)
		eq := x.Real.Cmp(&y.Real) == 0 && x.Imag.Cmp(&y.Imag) == 0
		switch op {
		case ir.OEQ:
			return eq
		case ir.ONE:
			return !eq
		}

	case ir.CTSTR:
		x, y := x.U.(string), y.U.(string)
		switch op {
		case ir.OEQ:
			return x == y
		case ir.ONE:
			return x != y
		case ir.OLT:
			return x < y
		case ir.OLE:
			return x <= y
		case ir.OGT:
			return x > y
		case ir.OGE:
			return x >= y
		}
	}

	base.Fatal("compareOp: bad comparison: %v %v %v", x, op, y)
	panic("unreachable")
}

func cmpZero(x int, op ir.Op) bool {
	switch op {
	case ir.OEQ:
		return x == 0
	case ir.ONE:
		return x != 0
	case ir.OLT:
		return x < 0
	case ir.OLE:
		return x <= 0
	case ir.OGT:
		return x > 0
	case ir.OGE:
		return x >= 0
	}

	base.Fatal("cmpZero: want comparison operator, got %v", op)
	panic("unreachable")
}

func binaryOp(x ir.Val, op ir.Op, y ir.Val) ir.Val {
	x, y = match(x, y)

Outer:
	switch x.Ctype() {
	case ir.CTBOOL:
		x, y := x.U.(bool), y.U.(bool)
		switch op {
		case ir.OANDAND:
			return ir.Val{U: x && y}
		case ir.OOROR:
			return ir.Val{U: x || y}
		}

	case ir.CTINT, ir.CTRUNE:
		x, y := x.U.(*ir.Int), y.U.(*ir.Int)

		u := new(ir.Int)
		u.Rune = x.Rune || y.Rune
		u.Set(x)
		switch op {
		case ir.OADD:
			u.Add(y)
		case ir.OSUB:
			u.Sub(y)
		case ir.OMUL:
			u.Mul(y)
		case ir.ODIV:
			if y.CmpInt64(0) == 0 {
				base.Error("division by zero")
				return ir.Val{}
			}
			u.Quo(y)
		case ir.OMOD:
			if y.CmpInt64(0) == 0 {
				base.Error("division by zero")
				return ir.Val{}
			}
			u.Rem(y)
		case ir.OOR:
			u.Or(y)
		case ir.OAND:
			u.And(y)
		case ir.OANDNOT:
			u.AndNot(y)
		case ir.OXOR:
			u.Xor(y)
		default:
			break Outer
		}
		return ir.Val{U: u}

	case ir.CTFLT:
		x, y := x.U.(*ir.Float), y.U.(*ir.Float)

		u := ir.NewFloat()
		u.Set(x)
		switch op {
		case ir.OADD:
			u.Add(y)
		case ir.OSUB:
			u.Sub(y)
		case ir.OMUL:
			u.Mul(y)
		case ir.ODIV:
			if y.CmpFloat64(0) == 0 {
				base.Error("division by zero")
				return ir.Val{}
			}
			u.Quo(y)
		default:
			break Outer
		}
		return ir.Val{U: u}

	case ir.CTCPLX:
		x, y := x.U.(*ir.Complex), y.U.(*ir.Complex)

		u := ir.NewComplex()
		u.Real.Set(&x.Real)
		u.Imag.Set(&x.Imag)
		switch op {
		case ir.OADD:
			u.Real.Add(&y.Real)
			u.Imag.Add(&y.Imag)
		case ir.OSUB:
			u.Real.Sub(&y.Real)
			u.Imag.Sub(&y.Imag)
		case ir.OMUL:
			u.Mul(y)
		case ir.ODIV:
			if !u.Div(y) {
				base.Error("complex division by zero")
				return ir.Val{}
			}
		default:
			break Outer
		}
		return ir.Val{U: u}
	}

	base.Fatal("binaryOp: bad operation: %v %v %v", x, op, y)
	panic("unreachable")
}

func unaryOp(op ir.Op, x ir.Val, t *types.Type) ir.Val {
	switch op {
	case ir.OPLUS:
		switch x.Ctype() {
		case ir.CTINT, ir.CTRUNE, ir.CTFLT, ir.CTCPLX:
			return x
		}

	case ir.ONEG:
		switch x.Ctype() {
		case ir.CTINT, ir.CTRUNE:
			x := x.U.(*ir.Int)
			u := new(ir.Int)
			u.Rune = x.Rune
			u.Set(x)
			u.Neg()
			return ir.Val{U: u}

		case ir.CTFLT:
			x := x.U.(*ir.Float)
			u := ir.NewFloat()
			u.Set(x)
			u.Neg()
			return ir.Val{U: u}

		case ir.CTCPLX:
			x := x.U.(*ir.Complex)
			u := ir.NewComplex()
			u.Real.Set(&x.Real)
			u.Imag.Set(&x.Imag)
			u.Real.Neg()
			u.Imag.Neg()
			return ir.Val{U: u}
		}

	case ir.OBITNOT:
		switch x.Ctype() {
		case ir.CTINT, ir.CTRUNE:
			x := x.U.(*ir.Int)

			u := new(ir.Int)
			u.Rune = x.Rune
			if t.IsSigned() || t.IsUntyped() {
				// Signed values change sign.
				u.SetInt64(-1)
			} else {
				// Unsigned values invert their bits.
				u.Set(maxintval[t.Etype])
			}
			u.Xor(x)
			return ir.Val{U: u}
		}

	case ir.ONOT:
		return ir.Val{U: !x.U.(bool)}
	}

	base.Fatal("unaryOp: bad operation: %v %v", op, x)
	panic("unreachable")
}

func shiftOp(x ir.Val, op ir.Op, y ir.Val) ir.Val {
	if x.Ctype() != ir.CTRUNE {
		x = toint(x)
	}
	y = toint(y)

	u := new(ir.Int)
	u.Set(x.U.(*ir.Int))
	u.Rune = x.U.(*ir.Int).Rune
	switch op {
	case ir.OLSH:
		u.Lsh(y.U.(*ir.Int))
	case ir.ORSH:
		u.Rsh(y.U.(*ir.Int))
	default:
		base.Fatal("shiftOp: bad operator: %v", op)
		panic("unreachable")
	}
	return ir.Val{U: u}
}

// setconst rewrites n as an OLITERAL with value v.
func setconst(n *ir.Node, v ir.Val) {
	// If constant folding failed, mark n as broken and give up.
	if v.U == nil {
		n.Type = nil
		return
	}

	// Ensure n.Orig still points to a semantically-equivalent
	// expression after we rewrite n into a constant.
	if n.Orig == n {
		n.Orig = n.SepCopy()
	}

	*n = ir.Node{
		Op:      ir.OLITERAL,
		Pos:     n.Pos,
		Orig:    n.Orig,
		Type:    n.Type,
		Xoffset: types.BADWIDTH,
	}
	n.SetVal(v)
	if vt := idealType(v.Ctype()); n.Type.IsUntyped() && n.Type != vt {
		base.Fatal("untyped type mismatch, have: %v, want: %v", n.Type, vt)
	}

	// Check range.
	lno := setlineno(n)
	overflow(v, n.Type)
	base.Pos = lno

	if !n.Type.IsUntyped() {
		switch v.Ctype() {
		// Truncate precision for non-ideal float.
		case ir.CTFLT:
			n.SetVal(ir.Val{U: truncfltlit(v.U.(*ir.Float), n.Type)})
		// Truncate precision for non-ideal complex.
		case ir.CTCPLX:
			n.SetVal(ir.Val{U: trunccmplxlit(v.U.(*ir.Complex), n.Type)})
		}
	}
}

func setboolconst(n *ir.Node, v bool) {
	setconst(n, ir.Val{U: v})
}

func setintconst(n *ir.Node, v int64) {
	u := new(ir.Int)
	u.SetInt64(v)
	setconst(n, ir.Val{U: u})
}

// nodlit returns a new untyped constant with value v.
func nodlit(v ir.Val) *ir.Node {
	n := nod(ir.OLITERAL, nil, nil)
	n.SetVal(v)
	n.Type = idealType(v.Ctype())
	return n
}

func idealType(ct ir.Ctype) *types.Type {
	switch ct {
	case ir.CTSTR:
		return types.UntypedString
	case ir.CTBOOL:
		return types.UntypedBool
	case ir.CTINT:
		return types.UntypedInt
	case ir.CTRUNE:
		return types.UntypedRune
	case ir.CTFLT:
		return types.UntypedFloat
	case ir.CTCPLX:
		return types.UntypedComplex
	case ir.CTNIL:
		return types.Types[types.TNIL]
	}
	base.Fatal("unexpected Ctype: %v", ct)
	return nil
}

// defaultlit on both nodes simultaneously;
// if they're both ideal going in they better
// get the same type going out.
// force means must assign concrete (non-ideal) type.
// The results of defaultlit2 MUST be assigned back to l and r, e.g.
// 	n.Left, n.Right = defaultlit2(n.Left, n.Right, force)
func defaultlit2(l *ir.Node, r *ir.Node, force bool) (*ir.Node, *ir.Node) {
	if l.Type == nil || r.Type == nil {
		return l, r
	}
	if !l.Type.IsUntyped() {
		r = convlit(r, l.Type)
		return l, r
	}

	if !r.Type.IsUntyped() {
		l = convlit(l, r.Type)
		return l, r
	}

	if !force {
		return l, r
	}

	// Can't mix bool with non-bool, string with non-string, or nil with anything (untyped).
	if l.Type.IsBoolean() != r.Type.IsBoolean() {
		return l, r
	}
	if l.Type.IsString() != r.Type.IsString() {
		return l, r
	}
	if l.IsNil() || r.IsNil() {
		return l, r
	}

	t := defaultType(mixUntyped(l.Type, r.Type))
	l = convlit(l, t)
	r = convlit(r, t)
	return l, r
}

func ctype(t *types.Type) ir.Ctype {
	switch t {
	case types.UntypedBool:
		return ir.CTBOOL
	case types.UntypedString:
		return ir.CTSTR
	case types.UntypedInt:
		return ir.CTINT
	case types.UntypedRune:
		return ir.CTRUNE
	case types.UntypedFloat:
		return ir.CTFLT
	case types.UntypedComplex:
		return ir.CTCPLX
	}
	base.Fatal("bad type %v", t)
	panic("unreachable")
}

func mixUntyped(t1, t2 *types.Type) *types.Type {
	t := t1
	if ctype(t2) > ctype(t1) {
		t = t2
	}
	return t
}

func defaultType(t *types.Type) *types.Type {
	if !t.IsUntyped() || t.Etype == types.TNIL {
		return t
	}

	switch t {
	case types.UntypedBool:
		return types.Types[types.TBOOL]
	case types.UntypedString:
		return types.Types[types.TSTRING]
	case types.UntypedInt:
		return types.Types[types.TINT]
	case types.UntypedRune:
		return types.Runetype
	case types.UntypedFloat:
		return types.Types[types.TFLOAT64]
	case types.UntypedComplex:
		return types.Types[types.TCOMPLEX128]
	}

	base.Fatal("bad type %v", t)
	return nil
}

func smallintconst(n *ir.Node) bool {
	if n.Op == ir.OLITERAL && ir.IsConst(n, ir.CTINT) && n.Type != nil {
		switch simtype[n.Type.Etype] {
		case types.TINT8,
			types.TUINT8,
			types.TINT16,
			types.TUINT16,
			types.TINT32,
			types.TUINT32,
			types.TBOOL:
			return true

		case types.TIDEAL, types.TINT64, types.TUINT64, types.TPTR:
			v, ok := n.Val().U.(*ir.Int)
			if ok && v.Cmp(minintval[types.TINT32]) >= 0 && v.Cmp(maxintval[types.TINT32]) <= 0 {
				return true
			}
		}
	}

	return false
}

// indexconst checks if Node n contains a constant expression
// representable as a non-negative int and returns its value.
// If n is not a constant expression, not representable as an
// integer, or negative, it returns -1. If n is too large, it
// returns -2.
func indexconst(n *ir.Node) int64 {
	if n.Op != ir.OLITERAL {
		return -1
	}

	v := toint(n.Val()) // toint returns argument unchanged if not representable as an *Mpint
	vi, ok := v.U.(*ir.Int)
	if !ok || vi.CmpInt64(0) < 0 {
		return -1
	}
	if vi.Cmp(maxintval[types.TINT]) > 0 {
		return -2
	}

	return vi.Int64()
}

// isGoConst reports whether n is a Go language constant (as opposed to a
// compile-time constant).
//
// Expressions derived from nil, like string([]byte(nil)), while they
// may be known at compile time, are not Go language constants.
func isGoConst(n *ir.Node) bool {
	return n.Op == ir.OLITERAL && n.Val().Ctype() != ir.CTNIL
}

func hascallchan(n *ir.Node) bool {
	if n == nil {
		return false
	}
	switch n.Op {
	case ir.OAPPEND,
		ir.OCALL,
		ir.OCALLFUNC,
		ir.OCALLINTER,
		ir.OCALLMETH,
		ir.OCAP,
		ir.OCLOSE,
		ir.OCOMPLEX,
		ir.OCOPY,
		ir.ODELETE,
		ir.OIMAG,
		ir.OLEN,
		ir.OMAKE,
		ir.ONEW,
		ir.OPANIC,
		ir.OPRINT,
		ir.OPRINTN,
		ir.OREAL,
		ir.ORECOVER,
		ir.ORECV:
		return true
	}

	if hascallchan(n.Left) || hascallchan(n.Right) {
		return true
	}
	for _, n1 := range n.List.Slice() {
		if hascallchan(n1) {
			return true
		}
	}
	for _, n2 := range n.Rlist.Slice() {
		if hascallchan(n2) {
			return true
		}
	}

	return false
}

// A constSet represents a set of Go constant expressions.
type constSet struct {
	m map[constSetKey]src.XPos
}

type constSetKey struct {
	typ *types.Type
	val interface{}
}

// add adds constant expression n to s. If a constant expression of
// equal value and identical type has already been added, then add
// reports an error about the duplicate value.
//
// pos provides position information for where expression n occurred
// (in case n does not have its own position information). what and
// where are used in the error message.
//
// n must not be an untyped constant.
func (s *constSet) add(pos src.XPos, n *ir.Node, what, where string) {
	if n.Op == ir.OCONVIFACE && n.Implicit() {
		n = n.Left
	}

	if !isGoConst(n) {
		return
	}
	if n.Type.IsUntyped() {
		base.Fatal("%v is untyped", n)
	}

	// Consts are only duplicates if they have the same value and
	// identical types.
	//
	// In general, we have to use types.Identical to test type
	// identity, because == gives false negatives for anonymous
	// types and the byte/uint8 and rune/int32 builtin type
	// aliases.  However, this is not a problem here, because
	// constant expressions are always untyped or have a named
	// type, and we explicitly handle the builtin type aliases
	// below.
	//
	// This approach may need to be revisited though if we fix
	// #21866 by treating all type aliases like byte/uint8 and
	// rune/int32.

	typ := n.Type
	switch typ {
	case types.Bytetype:
		typ = types.Types[types.TUINT8]
	case types.Runetype:
		typ = types.Types[types.TINT32]
	}
	k := constSetKey{typ, n.Val().Interface()}

	if hasUniquePos(n) {
		pos = n.Pos
	}

	if s.m == nil {
		s.m = make(map[constSetKey]src.XPos)
	}

	if prevPos, isDup := s.m[k]; isDup {
		base.ErrorAt(pos, "duplicate %s %s in %s\n\tprevious %s at %v",
			what, nodeAndVal(n), where,
			what, base.FmtPos(prevPos))
	} else {
		s.m[k] = pos
	}
}

// nodeAndVal reports both an expression and its constant value, if
// the latter is non-obvious.
//
// TODO(mdempsky): This could probably be a fmt.go flag.
func nodeAndVal(n *ir.Node) string {
	show := n.String()
	val := n.Val().Interface()
	if s := fmt.Sprintf("%#v", val); show != s {
		show += " (value " + s + ")"
	}
	return show
}

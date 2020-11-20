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
	"go/constant"
	"math/big"
	"strings"
)

func Eqval(a, b ir.Val) bool {
	if a.Kind() != b.Kind() {
		return false
	}
	switch x := a.U.(type) {
	default:
		base.Fatalf("unexpected Ctype for %T", a.U)
		panic("unreachable")
	case bool:
		y := b.U.(bool)
		return x == y
	case *ir.Int:
		y := b.U.(*ir.Int)
		return x.Cmp(y) == 0
	case *ir.Float:
		y := b.U.(*ir.Float)
		return x.Cmp(y) == 0
	case *ir.Complex:
		y := b.U.(*ir.Complex)
		return x.Real.Cmp(&y.Real) == 0 && x.Imag.Cmp(&y.Imag) == 0
	case string:
		y := b.U.(string)
		return x == y
	}
}

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
		base.Fatalf("truncfltlit: unexpected Etype %v", t.Etype)
	}

	return fv
}

// truncate Real and Imag parts of Mpcplx to 32-bit or 64-bit
// precision, according to type; return truncated value. In case of
// overflow, calls Errorf but does not truncate the input value.
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
		base.Fatalf("trunccplxlit: unexpected Etype %v", t.Etype)
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
		base.Fatalf("explicit conversion missing type")
	}
	if t != nil && t.IsUntyped() {
		base.Fatalf("bad conversion to untyped: %v", t)
	}

	if n == nil || n.Type == nil {
		// Allow sloppy callers.
		return n
	}
	if !n.Type.IsUntyped() {
		// Already typed; nothing to do.
		return n
	}

	if n.Op == ir.OLITERAL || n.Op == ir.ONIL {
		// Can't always set n.Type directly on OLITERAL nodes.
		// See discussion on CL 20813.
		n = n.RawCopy()
	}

	// Nil is technically not a constant, so handle it specially.
	if n.Type.Etype == types.TNIL {
		if n.Op != ir.ONIL {
			base.Fatalf("unexpected op: %v (%v)", n, n.Op)
		}
		if t == nil {
			base.Errorf("use of untyped nil")
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
		base.Fatalf("unexpected untyped expression: %v", n)

	case ir.OLITERAL:
		v := convertVal(n.Val(), t, explicit)
		if v.U == nil {
			break
		}
		n.Type = t
		n.SetVal(v)
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
			base.Errorf("invalid operation: %v (mismatched types %v and %v)", n, n.Left.Type, n.Right.Type)
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
			base.Errorf("invalid operation: %v (shift of type %v)", n, n.Type)
			n.Type = nil
		}
		return n
	}

	if !n.Diag() {
		if !t.Broke() {
			if explicit {
				base.Errorf("cannot convert %L to type %v", n, t)
			} else if context != nil {
				base.Errorf("cannot use %L as type %v in %s", n, t, context())
			} else {
				base.Errorf("cannot use %L as type %v", n, t)
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
	switch ct := v.Kind(); ct {
	case constant.Bool:
		if t.IsBoolean() {
			return v
		}

	case constant.String:
		if t.IsString() {
			return v
		}

	case constant.Int:
		if explicit && t.IsString() {
			return tostr(v)
		}
		fallthrough
	case constant.Float, constant.Complex:
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
			base.Errorf("constant %v truncated to real", u.GoString())
		}
		v.U = f
	}

	return v
}

func toint(v ir.Val) ir.Val {
	switch u := v.U.(type) {
	case *ir.Int:

	case *ir.Float:
		i := new(ir.Int)
		if !i.SetFloat(u) {
			if i.CheckOverflow(0) {
				base.Errorf("integer too large")
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
					base.Errorf("constant truncated to integer")
				} else {
					base.Errorf("constant %v truncated to integer", u.GoString())
				}
			}
		}
		v.U = i

	case *ir.Complex:
		i := new(ir.Int)
		if !i.SetFloat(&u.Real) || u.Imag.CmpFloat64(0) != 0 {
			base.Errorf("constant %v truncated to integer", u.GoString())
		}

		v.U = i
	}

	return v
}

func doesoverflow(v ir.Val, t *types.Type) bool {
	switch u := v.U.(type) {
	case *ir.Int:
		if !t.IsInteger() {
			base.Fatalf("overflow: %v integer constant", t)
		}
		return u.Cmp(minintval[t.Etype]) < 0 || u.Cmp(maxintval[t.Etype]) > 0

	case *ir.Float:
		if !t.IsFloat() {
			base.Fatalf("overflow: %v floating-point constant", t)
		}
		return u.Cmp(minfltval[t.Etype]) <= 0 || u.Cmp(maxfltval[t.Etype]) >= 0

	case *ir.Complex:
		if !t.IsComplex() {
			base.Fatalf("overflow: %v complex constant", t)
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
		base.Errorf("constant %v overflows %v", v, t)
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

// evalConst returns a constant-evaluated expression equivalent to n.
// If n is not a constant, evalConst returns n.
// Otherwise, evalConst returns a new OLITERAL with the same value as n,
// and with .Orig pointing back to n.
func evalConst(n *ir.Node) *ir.Node {
	nl, nr := n.Left, n.Right

	// Pick off just the opcodes that can be constant evaluated.
	switch op := n.Op; op {
	case ir.OPLUS, ir.ONEG, ir.OBITNOT, ir.ONOT:
		if nl.Op == ir.OLITERAL {
			return origConst(n, unaryOp(op, nl.Val(), n.Type))
		}

	case ir.OADD, ir.OSUB, ir.OMUL, ir.ODIV, ir.OMOD, ir.OOR, ir.OXOR, ir.OAND, ir.OANDNOT, ir.OOROR, ir.OANDAND:
		if nl.Op == ir.OLITERAL && nr.Op == ir.OLITERAL {
			return origConst(n, binaryOp(nl.Val(), op, nr.Val()))
		}

	case ir.OEQ, ir.ONE, ir.OLT, ir.OLE, ir.OGT, ir.OGE:
		if nl.Op == ir.OLITERAL && nr.Op == ir.OLITERAL {
			return origBoolConst(n, compareOp(nl.Val(), op, nr.Val()))
		}

	case ir.OLSH, ir.ORSH:
		if nl.Op == ir.OLITERAL && nr.Op == ir.OLITERAL {
			return origConst(n, shiftOp(nl.Val(), op, nr.Val()))
		}

	case ir.OCONV, ir.ORUNESTR:
		if okforconst[n.Type.Etype] && nl.Op == ir.OLITERAL {
			return origConst(n, convertVal(nl.Val(), n.Type, true))
		}

	case ir.OCONVNOP:
		if okforconst[n.Type.Etype] && nl.Op == ir.OLITERAL {
			// set so n.Orig gets OCONV instead of OCONVNOP
			n.Op = ir.OCONV
			return origConst(n, nl.Val())
		}

	case ir.OADDSTR:
		// Merge adjacent constants in the argument list.
		s := n.List.Slice()
		need := 0
		for i := 0; i < len(s); i++ {
			if i == 0 || !ir.IsConst(s[i-1], constant.String) || !ir.IsConst(s[i], constant.String) {
				// Can't merge s[i] into s[i-1]; need a slot in the list.
				need++
			}
		}
		if need == len(s) {
			return n
		}
		if need == 1 {
			var strs []string
			for _, c := range s {
				strs = append(strs, c.StringVal())
			}
			return origConst(n, ir.Val{U: strings.Join(strs, "")})
		}
		newList := make([]*ir.Node, 0, need)
		for i := 0; i < len(s); i++ {
			if ir.IsConst(s[i], constant.String) && i+1 < len(s) && ir.IsConst(s[i+1], constant.String) {
				// merge from i up to but not including i2
				var strs []string
				i2 := i
				for i2 < len(s) && ir.IsConst(s[i2], constant.String) {
					strs = append(strs, s[i2].StringVal())
					i2++
				}

				nl := origConst(s[i], ir.Val{U: strings.Join(strs, "")})
				nl.Orig = nl // it's bigger than just s[i]
				newList = append(newList, nl)
				i = i2 - 1
			} else {
				newList = append(newList, s[i])
			}
		}

		n = ir.Copy(n)
		n.List.Set(newList)
		return n

	case ir.OCAP, ir.OLEN:
		switch nl.Type.Etype {
		case types.TSTRING:
			if ir.IsConst(nl, constant.String) {
				return origIntConst(n, int64(len(nl.StringVal())))
			}
		case types.TARRAY:
			if !hascallchan(nl) {
				return origIntConst(n, nl.Type.NumElem())
			}
		}

	case ir.OALIGNOF, ir.OOFFSETOF, ir.OSIZEOF:
		return origIntConst(n, evalunsafe(n))

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
				base.Fatalf("impossible")
			}
			if n.Op == ir.OIMAG {
				if im == nil {
					im = ir.NewFloat()
				}
				re = im
			}
			return origConst(n, ir.Val{U: re})
		}

	case ir.OCOMPLEX:
		if nl.Op == ir.OLITERAL && nr.Op == ir.OLITERAL {
			// make it a complex literal
			c := ir.NewComplex()
			c.Real.Set(toflt(nl.Val()).U.(*ir.Float))
			c.Imag.Set(toflt(nr.Val()).U.(*ir.Float))
			return origConst(n, ir.Val{U: c})
		}
	}

	return n
}

func match(x, y ir.Val) (ir.Val, ir.Val) {
	switch {
	case x.Kind() == constant.Complex || y.Kind() == constant.Complex:
		return tocplx(x), tocplx(y)
	case x.Kind() == constant.Float || y.Kind() == constant.Float:
		return toflt(x), toflt(y)
	}

	// Mixed int/rune are fine.
	return x, y
}

func compareOp(x ir.Val, op ir.Op, y ir.Val) bool {
	x, y = match(x, y)

	switch x.Kind() {
	case constant.Bool:
		x, y := x.U.(bool), y.U.(bool)
		switch op {
		case ir.OEQ:
			return x == y
		case ir.ONE:
			return x != y
		}

	case constant.Int:
		x, y := x.U.(*ir.Int), y.U.(*ir.Int)
		return cmpZero(x.Cmp(y), op)

	case constant.Float:
		x, y := x.U.(*ir.Float), y.U.(*ir.Float)
		return cmpZero(x.Cmp(y), op)

	case constant.Complex:
		x, y := x.U.(*ir.Complex), y.U.(*ir.Complex)
		eq := x.Real.Cmp(&y.Real) == 0 && x.Imag.Cmp(&y.Imag) == 0
		switch op {
		case ir.OEQ:
			return eq
		case ir.ONE:
			return !eq
		}

	case constant.String:
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

	base.Fatalf("compareOp: bad comparison: %v %v %v", x, op, y)
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

	base.Fatalf("cmpZero: want comparison operator, got %v", op)
	panic("unreachable")
}

func binaryOp(x ir.Val, op ir.Op, y ir.Val) ir.Val {
	x, y = match(x, y)

Outer:
	switch x.Kind() {
	case constant.Bool:
		x, y := x.U.(bool), y.U.(bool)
		switch op {
		case ir.OANDAND:
			return ir.Val{U: x && y}
		case ir.OOROR:
			return ir.Val{U: x || y}
		}

	case constant.Int:
		x, y := x.U.(*ir.Int), y.U.(*ir.Int)

		u := new(ir.Int)
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
				base.Errorf("division by zero")
				return ir.Val{}
			}
			u.Quo(y)
		case ir.OMOD:
			if y.CmpInt64(0) == 0 {
				base.Errorf("division by zero")
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

	case constant.Float:
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
				base.Errorf("division by zero")
				return ir.Val{}
			}
			u.Quo(y)
		default:
			break Outer
		}
		return ir.Val{U: u}

	case constant.Complex:
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
				base.Errorf("complex division by zero")
				return ir.Val{}
			}
		default:
			break Outer
		}
		return ir.Val{U: u}
	}

	base.Fatalf("binaryOp: bad operation: %v %v %v", x, op, y)
	panic("unreachable")
}

func unaryOp(op ir.Op, x ir.Val, t *types.Type) ir.Val {
	switch op {
	case ir.OPLUS:
		switch x.Kind() {
		case constant.Int, constant.Float, constant.Complex:
			return x
		}

	case ir.ONEG:
		switch x.Kind() {
		case constant.Int:
			x := x.U.(*ir.Int)
			u := new(ir.Int)
			u.Set(x)
			u.Neg()
			return ir.Val{U: u}

		case constant.Float:
			x := x.U.(*ir.Float)
			u := ir.NewFloat()
			u.Set(x)
			u.Neg()
			return ir.Val{U: u}

		case constant.Complex:
			x := x.U.(*ir.Complex)
			u := ir.NewComplex()
			u.Real.Set(&x.Real)
			u.Imag.Set(&x.Imag)
			u.Real.Neg()
			u.Imag.Neg()
			return ir.Val{U: u}
		}

	case ir.OBITNOT:
		switch x.Kind() {
		case constant.Int:
			x := x.U.(*ir.Int)

			u := new(ir.Int)
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

	base.Fatalf("unaryOp: bad operation: %v %v", op, x)
	panic("unreachable")
}

func shiftOp(x ir.Val, op ir.Op, y ir.Val) ir.Val {
	x = toint(x)
	y = toint(y)

	u := new(ir.Int)
	u.Set(x.U.(*ir.Int))
	switch op {
	case ir.OLSH:
		u.Lsh(y.U.(*ir.Int))
	case ir.ORSH:
		u.Rsh(y.U.(*ir.Int))
	default:
		base.Fatalf("shiftOp: bad operator: %v", op)
		panic("unreachable")
	}
	return ir.Val{U: u}
}

// origConst returns an OLITERAL with orig n and value v.
func origConst(n *ir.Node, v ir.Val) *ir.Node {
	// If constant folding was attempted (we were called)
	// but it produced an invalid constant value,
	// mark n as broken and give up.
	if v.U == nil {
		n.Type = nil
		return n
	}

	orig := n
	n = ir.Nod(ir.OLITERAL, nil, nil)
	n.Orig = orig
	n.Pos = orig.Pos
	n.Type = orig.Type
	n.SetVal(v)

	// Check range.
	lno := setlineno(n)
	overflow(v, n.Type)
	base.Pos = lno

	if !n.Type.IsUntyped() {
		switch v.Kind() {
		// Truncate precision for non-ideal float.
		case constant.Float:
			n.SetVal(ir.Val{U: truncfltlit(v.U.(*ir.Float), n.Type)})
		// Truncate precision for non-ideal complex.
		case constant.Complex:
			n.SetVal(ir.Val{U: trunccmplxlit(v.U.(*ir.Complex), n.Type)})
		}
	}
	return n
}

func origBoolConst(n *ir.Node, v bool) *ir.Node {
	return origConst(n, ir.Val{U: v})
}

func origIntConst(n *ir.Node, v int64) *ir.Node {
	u := new(ir.Int)
	u.SetInt64(v)
	return origConst(n, ir.Val{U: u})
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
	if ir.IsNil(l) || ir.IsNil(r) {
		return l, r
	}

	t := defaultType(mixUntyped(l.Type, r.Type))
	l = convlit(l, t)
	r = convlit(r, t)
	return l, r
}

func mixUntyped(t1, t2 *types.Type) *types.Type {
	if t1 == t2 {
		return t1
	}

	rank := func(t *types.Type) int {
		switch t {
		case types.UntypedInt:
			return 0
		case types.UntypedRune:
			return 1
		case types.UntypedFloat:
			return 2
		case types.UntypedComplex:
			return 3
		}
		base.Fatalf("bad type %v", t)
		panic("unreachable")
	}

	if rank(t2) > rank(t1) {
		return t2
	}
	return t1
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

	base.Fatalf("bad type %v", t)
	return nil
}

func smallintconst(n *ir.Node) bool {
	if n.Op == ir.OLITERAL && ir.IsConst(n, constant.Int) && n.Type != nil {
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
	return n.Op == ir.OLITERAL
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
		base.Fatalf("%v is untyped", n)
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
		base.ErrorfAt(pos, "duplicate %s %s in %s\n\tprevious %s at %v",
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

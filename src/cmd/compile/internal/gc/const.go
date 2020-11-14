// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gc

import (
	"cmd/compile/internal/types"
	"cmd/internal/src"
	"fmt"
	"go/constant"
	"go/token"
	"math"
	"math/big"
	"unicode"
)

const (
	// Maximum size in bits for big.Ints before signalling
	// overflow and also mantissa precision for big.Floats.
	Mpprec = 512
)

// Legacy constant kind names.
// TODO(mdempsky): Replace.
const (
	CTINT  = constant.Int
	CTFLT  = constant.Float
	CTCPLX = constant.Complex
	CTSTR  = constant.String
	CTBOOL = constant.Bool
)

// Interface returns the constant value stored in v as an interface{}.
// It returns int64s for ints and runes, float64s for floats,
// and complex128s for complex values.
func (n *Node) ValueInterface() interface{} {
	floatVal := func(v constant.Value) float64 {
		x, _ := constant.Float64Val(v)
		// check for overflow
		if math.IsInf(x, 0) && nsavederrors+nerrors == 0 {
			Fatalf("ovf in Mpflt Float64")
		}
		return x + 0 // avoid -0 (should not be needed, but be conservative)
	}

	switch v := n.Val(); v.Kind() {
	default:
		Fatalf("unexpected constant: %v", v)
		panic("unreachable")

	case constant.Bool:
		return constant.BoolVal(v)
	case constant.String:
		return constant.StringVal(v)
	case constant.Int:
		if n.Type.IsUnsigned() {
			i, ok := constant.Uint64Val(v)
			if !ok {
				Fatalf("unsigned value overflows uint64: %v", n)
			}
			return int64(i)
		}
		i, ok := constant.Int64Val(v)
		if !ok {
			Fatalf("signed value overflows int64: %v", n)
		}
		return i
	case constant.Float:
		return floatVal(v)
	case constant.Complex:
		return complex(floatVal(constant.Real(v)), floatVal(constant.Imag(v)))
	}
}

// Int64Val returns n as an int64.
// n must be an integer or rune constant.
func (n *Node) Int64Val() int64 {
	if !Isconst(n, CTINT) {
		Fatalf("Int64Val(%v)", n)
	}
	x, ok := constant.Int64Val(n.Val())
	if !ok {
		Fatalf("Int64Val(%v)", n)
	}
	return x
}

// CanInt64 reports whether it is safe to call Int64Val() on n.
func (n *Node) CanInt64() bool {
	if !Isconst(n, CTINT) {
		return false
	}

	// if the value inside n cannot be represented as an int64, the
	// return value of Int64 is undefined
	_, ok := constant.Int64Val(n.Val())
	return ok
}

// BoolVal returns n as a bool.
// n must be a boolean constant.
func (n *Node) BoolVal() bool {
	if !Isconst(n, CTBOOL) {
		Fatalf("BoolVal(%v)", n)
	}
	return constant.BoolVal(n.Val())
}

// StringVal returns the value of a literal string Node as a string.
// n must be a string constant.
func (n *Node) StringVal() string {
	if !Isconst(n, CTSTR) {
		Fatalf("StringVal(%v)", n)
	}
	return constant.StringVal(n.Val())
}

// truncate float literal fv to 32-bit or 64-bit precision
// according to type; return truncated value.
func truncfltlit(v constant.Value, t *types.Type) constant.Value {
	if t == nil || overflow(v, t) {
		// If there was overflow, simply continuing would set the
		// value to Inf which in turn would lead to spurious follow-on
		// errors. Avoid this by returning the existing value.
		return v
	}

	switch t.Etype {
	case TFLOAT32:
		f, _ := constant.Float32Val(v)
		v = constant.MakeFloat64(float64(f))
	case TFLOAT64:
		f, _ := constant.Float64Val(v)
		v = constant.MakeFloat64(f)
	}
	v = constant.ToFloat(v) // ugh, constant.MakeFloat64 truncates 0 to Int
	return v
}

// truncate Real and Imag parts of Mpcplx to 32-bit or 64-bit
// precision, according to type; return truncated value. In case of
// overflow, calls yyerror but does not truncate the input value.
func trunccmplxlit(v constant.Value, t *types.Type) constant.Value {
	if t == nil || overflow(v, t) {
		// If there was overflow, simply continuing would set the
		// value to Inf which in turn would lead to spurious follow-on
		// errors. Avoid this by returning the existing value.
		return v
	}

	switch t.Etype {
	case TCOMPLEX64:
		re, _ := constant.Float32Val(constant.Real(v))
		im, _ := constant.Float32Val(constant.Imag(v))
		v = complexVal(constant.MakeFloat64(float64(re)), constant.MakeFloat64(float64(im)))
	case TCOMPLEX128:
		re, _ := constant.Float64Val(constant.Real(v))
		im, _ := constant.Float64Val(constant.Imag(v))
		v = complexVal(constant.MakeFloat64(re), constant.MakeFloat64(im))
	}
	v = constant.ToComplex(v)
	return v
}

// TODO(mdempsky): Replace these with better APIs.
func convlit(n *Node, t *types.Type) *Node    { return convlit1(n, t, false, nil) }
func defaultlit(n *Node, t *types.Type) *Node { return convlit1(n, t, false, nil) }

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
func convlit1(n *Node, t *types.Type, explicit bool, context func() string) *Node {
	if explicit && t == nil {
		Fatalf("explicit conversion missing type")
	}
	if t != nil && t.IsUntyped() {
		Fatalf("bad conversion to untyped: %v", t)
	}

	if n == nil || n.Type == nil {
		// Allow sloppy callers.
		return n
	}
	if !n.Type.IsUntyped() {
		// Already typed; nothing to do.
		return n
	}

	if n.Op == OLITERAL || n.Op == ONIL {
		// Can't always set n.Type directly on OLITERAL nodes.
		// See discussion on CL 20813.
		n = n.rawcopy()
	}

	// Nil is technically not a constant, so handle it specially.
	if n.Type.Etype == TNIL {
		if n.Op != ONIL {
			Fatalf("unexpected op: %v (%v)", n, n.Op)
		}
		if t == nil {
			yyerror("use of untyped nil")
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
		Fatalf("unexpected untyped expression: %v", n)

	case OLITERAL:
		v := convertVal(n.Val(), t, explicit)
		if v.Kind() == constant.Unknown {
			break
		}
		n.Type = t
		n.SetVal(v)
		return n

	case OPLUS, ONEG, OBITNOT, ONOT, OREAL, OIMAG:
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

	case OADD, OSUB, OMUL, ODIV, OMOD, OOR, OXOR, OAND, OANDNOT, OOROR, OANDAND, OCOMPLEX:
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
			yyerror("invalid operation: %v (mismatched types %v and %v)", n, n.Left.Type, n.Right.Type)
			n.Type = nil
			return n
		}

		n.Type = t
		return n

	case OEQ, ONE, OLT, OLE, OGT, OGE:
		if !t.IsBoolean() {
			break
		}
		n.Type = t
		return n

	case OLSH, ORSH:
		n.Left = convlit1(n.Left, t, explicit, nil)
		n.Type = n.Left.Type
		if n.Type != nil && !n.Type.IsInteger() {
			yyerror("invalid operation: %v (shift of type %v)", n, n.Type)
			n.Type = nil
		}
		return n
	}

	if !n.Diag() {
		if !t.Broke() {
			if explicit {
				yyerror("cannot convert %L to type %v", n, t)
			} else if context != nil {
				yyerror("cannot use %L as type %v in %s", n, t, context())
			} else {
				yyerror("cannot use %L as type %v", n, t)
			}
		}
		n.SetDiag(true)
	}
	n.Type = nil
	return n
}

func operandType(op Op, t *types.Type) *types.Type {
	switch op {
	case OCOMPLEX:
		if t.IsComplex() {
			return floatForComplex(t)
		}
	case OREAL, OIMAG:
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
func convertVal(v constant.Value, t *types.Type, explicit bool) constant.Value {
	switch ct := v.Kind(); ct {
	case CTBOOL:
		if t.IsBoolean() {
			return v
		}

	case CTSTR:
		if t.IsString() {
			return v
		}

	case CTINT:
		if explicit && t.IsString() {
			return tostr(v)
		}
		fallthrough
	case CTFLT, CTCPLX:
		// TODO(mdempsky): Combine toint/overflow, toflt/truncfltlit, tocplx/trunccmplxlit?
		switch {
		case t.IsInteger():
			v = toint(v)
			overflow(v, t)
			return v
		case t.IsFloat():
			v = toflt(v)
			v = truncfltlit(v, t)
			return v
		case t.IsComplex():
			v = tocplx(v)
			v = trunccmplxlit(v, t)
			return v
		}
	}

	return constant.MakeUnknown()
}

func tocplx(v constant.Value) constant.Value {
	return constant.ToComplex(v)
}

func toflt(v constant.Value) constant.Value {
	if v.Kind() == constant.Complex {
		if constant.Sign(constant.Imag(v)) != 0 {
			yyerror("constant %v truncated to real", v)
		}
		v = constant.Real(v)
	}

	return constant.ToFloat(v)
}

func asBigFloat(v constant.Value) *big.Float {
	// TODO(mdempsky): Avoid heap allocation?
	f := new(big.Float)
	f.SetPrec(Mpprec)
	switch u := constant.Val(v).(type) {
	case int64:
		f.SetInt64(u)
	case *big.Int:
		f.SetInt(u)
	case *big.Float:
		f.Set(u)
	case *big.Rat:
		f.SetRat(u)
	default:
		Fatalf("unexpected: %v", u)
	}
	return f
}

// TODO(mdempsky): Replace callers with convertVal?
func toint(v constant.Value) constant.Value {
	// TODO(mdempsky): Simplify. Code salvaged from
	// mpint.go/mpfloat.go, so it should be okay; but this is way
	// too complicated.

	switch v.Kind() {
	case constant.Complex:
		if constant.Sign(constant.Imag(v)) != 0 {
			yyerror("constant %v truncated to integer", v)
		}

		v = constant.Real(v)
		fallthrough
	case constant.Float:
		f := asBigFloat(v)

		var i big.Int
		var ovf bool

		setOverflow := func() {
			i.SetUint64(1) // avoid spurious div-zero errors
			ovf = true
		}

		setFloat := func() bool {
			// avoid converting huge floating-point numbers to integers
			// (2*Mpprec is large enough to permit all tests to pass)
			if f.MantExp(nil) > 2*Mpprec {
				setOverflow()
				return false
			}

			if _, acc := f.Int(&i); acc == big.Exact {
				return true
			}

			const delta = 16 // a reasonably small number of bits > 0
			var t big.Float
			t.SetPrec(Mpprec - delta)

			// try rounding down a little
			t.SetMode(big.ToZero)
			t.Set(f)
			if _, acc := t.Int(&i); acc == big.Exact {
				return true
			}

			// try rounding up a little
			t.SetMode(big.AwayFromZero)
			t.Set(f)
			if _, acc := t.Int(&i); acc == big.Exact {
				return true
			}

			ovf = false
			if i.BitLen() > Mpprec {
				setOverflow()
			}
			return false
		}

		goString := func() string {
			// determine sign
			sign := ""
			if f.Sign() < 0 {
				sign = "-"
				f = new(big.Float).Abs(f)
			}

			// Don't try to convert infinities (will not terminate).
			if f.IsInf() {
				return sign + "Inf"
			}

			// Use exact fmt formatting if in float64 range (common case):
			// proceed if f doesn't underflow to 0 or overflow to inf.
			if x, _ := f.Float64(); f.Sign() == 0 == (x == 0) && !math.IsInf(x, 0) {
				return fmt.Sprintf("%s%.6g", sign, x)
			}

			// Out of float64 range. Do approximate manual to decimal
			// conversion to avoid precise but possibly slow Float
			// formatting.
			// f = mant * 2**exp
			var mant big.Float
			exp := f.MantExp(&mant) // 0.5 <= mant < 1.0

			// approximate float64 mantissa m and decimal exponent d
			// f ~ m * 10**d
			m, _ := mant.Float64()                     // 0.5 <= m < 1.0
			d := float64(exp) * (math.Ln2 / math.Ln10) // log_10(2)

			// adjust m for truncated (integer) decimal exponent e
			e := int64(d)
			m *= math.Pow(10, d-float64(e))

			// ensure 1 <= m < 10
			switch {
			case m < 1-0.5e-6:
				// The %.6g format below rounds m to 5 digits after the
				// decimal point. Make sure that m*10 < 10 even after
				// rounding up: m*10 + 0.5e-5 < 10 => m < 1 - 0.5e6.
				m *= 10
				e--
			case m >= 10:
				m /= 10
				e++
			}

			return fmt.Sprintf("%s%.6ge%+d", sign, m, e)
		}

		if !setFloat() {
			if ovf {
				yyerror("integer too large")
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
				t.Parse(goString(), 0)
				if t.IsInt() {
					yyerror("constant truncated to integer")
				} else {
					yyerror("constant %v truncated to integer", v)
				}
			}
		}
		v = constant.Make(&i)
	}

	return v
}

func doesoverflow(v constant.Value, t *types.Type) bool {
	switch {
	case t.IsBoolean():
		_ = constant.BoolVal(v)
		return false
	case t.IsString():
		_ = constant.StringVal(v)
		return false
	case t.IsInteger():
		return constant.Compare(v, token.LSS, minval[t.Etype]) ||
			constant.Compare(v, token.GTR, maxval[t.Etype])
	case t.IsFloat():
		return constant.Compare(v, token.LEQ, minval[t.Etype]) ||
			constant.Compare(v, token.GEQ, maxval[t.Etype])
	case t.IsComplex():
		re := constant.Real(v)
		im := constant.Imag(v)
		return constant.Compare(re, token.LEQ, minval[t.Etype]) ||
			constant.Compare(re, token.GEQ, maxval[t.Etype]) ||
			constant.Compare(im, token.LEQ, minval[t.Etype]) ||
			constant.Compare(im, token.GEQ, maxval[t.Etype])
	}
	Fatalf("doesoverflow: %v, %v", v, t)
	panic("unreachable")
}

func overflow(v constant.Value, t *types.Type) bool {
	// v has already been converted
	// to appropriate form for t.
	if t == nil || t.IsUntyped() {
		return false
	}

	if doesoverflow(v, t) {
		yyerror("constant %v overflows %v", vconv(v, 0), t)
		return true
	}

	return false
}

func tostr(v constant.Value) constant.Value {
	if v.Kind() == constant.Int {
		var r rune = unicode.ReplacementChar
		if x, ok := constant.Uint64Val(v); ok && x <= unicode.MaxRune {
			r = rune(x)
		}
		v = constant.MakeString(string(r))
	}
	return v
}

func consttype(n *Node) constant.Kind {
	if n == nil || n.Op != OLITERAL {
		return constant.Unknown
	}
	return n.Val().Kind()
}

// TODO(mdempsky): Double check if this is still safe to use.
func Isconst(n *Node, ct constant.Kind) bool {
	return consttype(n) == ct
}

var tokenForOp = [...]token.Token{
	OPLUS:   token.ADD,
	ONEG:    token.SUB,
	ONOT:    token.NOT,
	OBITNOT: token.XOR,

	OADD:    token.ADD,
	OSUB:    token.SUB,
	OMUL:    token.MUL,
	ODIV:    token.QUO,
	OMOD:    token.REM,
	OOR:     token.OR,
	OXOR:    token.XOR,
	OAND:    token.AND,
	OANDNOT: token.AND_NOT,
	OOROR:   token.LOR,
	OANDAND: token.LAND,

	OEQ: token.EQL,
	ONE: token.NEQ,
	OLT: token.LSS,
	OLE: token.LEQ,
	OGT: token.GTR,
	OGE: token.GEQ,

	OLSH: token.SHL,
	ORSH: token.SHR,
}

// evconst rewrites constant expressions into OLITERAL nodes.
func evconst(n *Node) {
	nl, nr := n.Left, n.Right

	// Pick off just the opcodes that can be constant evaluated.
	switch op := n.Op; op {
	case OPLUS, ONEG, OBITNOT, ONOT:
		if nl.Op == OLITERAL {
			var prec uint
			if n.Type.IsUnsigned() {
				prec = uint(n.Type.Size() * 8)
			}
			setconst(n, constant.UnaryOp(tokenForOp[op], nl.Val(), prec))
		}

	case OADD, OSUB, OMUL, ODIV, OMOD, OOR, OXOR, OAND, OANDNOT, OOROR, OANDAND:
		if nl.Op == OLITERAL && nr.Op == OLITERAL {
			rval := nr.Val()

			// check for divisor underflow in complex division (see issue 20227)
			if op == ODIV && n.Type.IsComplex() && constant.Sign(square(constant.Real(rval))) == 0 && constant.Sign(square(constant.Imag(rval))) == 0 {
				yyerror("complex division by zero")
				break
			}
			if (op == ODIV || op == OMOD) && constant.Sign(rval) == 0 {
				yyerror("division by zero")
				break
			}

			tok := tokenForOp[op]
			if op == ODIV && n.Type.IsInteger() {
				tok = token.QUO_ASSIGN // integer division
			}
			setconst(n, constant.BinaryOp(nl.Val(), tok, rval))
		}

	case OEQ, ONE, OLT, OLE, OGT, OGE:
		if nl.Op == OLITERAL && nr.Op == OLITERAL {
			setboolconst(n, constant.Compare(nl.Val(), tokenForOp[op], nr.Val()))
		}

	case OLSH, ORSH:
		if nl.Op == OLITERAL && nr.Op == OLITERAL {
			const shiftBound = 1023 - 1 + 52 // so we can express smallestFloat64
			if s, ok := constant.Uint64Val(nr.Val()); ok && s < shiftBound {
				setconst(n, constant.Shift(toint(nl.Val()), tokenForOp[op], uint(s)))
			} else {
				yyerror("invalid shift count %v", nr)
			}
		}

	case OCONV, ORUNESTR:
		if okforconst[n.Type.Etype] && nl.Op == OLITERAL {
			setconst(n, convertVal(nl.Val(), n.Type, true))
		}

	case OCONVNOP:
		if okforconst[n.Type.Etype] && nl.Op == OLITERAL {
			// set so n.Orig gets OCONV instead of OCONVNOP
			n.Op = OCONV
			setconst(n, nl.Val())
		}

	case OADDSTR:
		// Merge adjacent constants in the argument list.
		s := n.List.Slice()
		for i1 := 0; i1 < len(s); i1++ {
			if Isconst(s[i1], CTSTR) && i1+1 < len(s) && Isconst(s[i1+1], CTSTR) {
				// merge from i1 up to but not including i2
				val := s[i1].Val()
				i2 := i1 + 1
				for i2 < len(s) && Isconst(s[i2], CTSTR) {
					val = constant.BinaryOp(val, token.ADD, s[i2].Val())
					i2++
				}

				nl := *s[i1]
				nl.Orig = &nl
				nl.SetVal(val)
				s[i1] = &nl
				s = append(s[:i1+1], s[i2:]...)
			}
		}

		if len(s) == 1 && Isconst(s[0], CTSTR) {
			n.Op = OLITERAL
			n.SetVal(s[0].Val())
		} else {
			n.List.Set(s)
		}

	case OCAP, OLEN:
		switch nl.Type.Etype {
		case TSTRING:
			if Isconst(nl, CTSTR) {
				setintconst(n, int64(len(nl.StringVal())))
			}
		case TARRAY:
			if !hascallchan(nl) {
				setintconst(n, nl.Type.NumElem())
			}
		}

	case OALIGNOF, OOFFSETOF, OSIZEOF:
		setintconst(n, evalunsafe(n))

	case OREAL:
		if nl.Op == OLITERAL {
			setconst(n, constant.Real(nl.Val()))
		}

	case OIMAG:
		if nl.Op == OLITERAL {
			setconst(n, constant.Imag(nl.Val()))
		}

	case OCOMPLEX:
		if nl.Op == OLITERAL && nr.Op == OLITERAL {
			setconst(n, complexVal(nl.Val(), nr.Val()))
		}
	}
}

func complexVal(real, imag constant.Value) constant.Value {
	return constant.BinaryOp(constant.ToFloat(real), token.ADD, constant.MakeImag(constant.ToFloat(imag)))
}

func square(x constant.Value) constant.Value {
	return constant.BinaryOp(x, token.MUL, x)
}

func constMatch(t *types.Type, v constant.Value) {
	switch v.Kind() {
	case constant.Unknown:
		if okforconst[t.Etype] {
			return
		}
	case constant.Bool:
		if t.IsBoolean() {
			return
		}
	case constant.String:
		if t.IsString() {
			return
		}
	case constant.Int:
		if t.IsInteger() {
			return
		}
	case constant.Float:
		if t.IsFloat() {
			return
		}
	case constant.Complex:
		if t.IsComplex() {
			return
		}
	}

	Fatalf("waka waka: %v, %v (%v)", t, v, v.Kind())
}

// setconst rewrites n as an OLITERAL with value v.
func setconst(n *Node, v constant.Value) {
	lno := setlineno(n)
	v = convertVal(v, n.Type, false)
	lineno = lno

	constMatch(n.Type, v)

	// If constant folding failed, mark n as broken and give up.
	if v.Kind() == constant.Unknown {
		if nsavederrors+nerrors == 0 {
			Fatalf("should have reported an error")
		}
		n.Type = nil
		return
	}

	// TODO(mdempsky): This is still terrible.
	if v.Kind() == constant.Int {
		if i, ok := constant.Val(v).(*big.Int); ok && i.BitLen() > Mpprec {
			names := [...]string{
				OADD: "addition",
				OSUB: "subtraction",
				OMUL: "multiplication",
				OLSH: "shift",
			}
			what := names[n.Op]
			if what == "" {
				Fatalf("unexpected overflow: %v", n.Op)
			}

			yyerror("constant %v overflow", what)
			n.Type = nil
			return
		}
	}

	// Ensure n.Orig still points to a semantically-equivalent
	// expression after we rewrite n into a constant.
	if n.Orig == n {
		n.Orig = n.sepcopy()
	}

	*n = Node{
		Op:      OLITERAL,
		Pos:     n.Pos,
		Orig:    n.Orig,
		Type:    n.Type,
		Xoffset: BADWIDTH,
	}
	n.SetVal(v)
}

func setboolconst(n *Node, v bool) {
	setconst(n, constant.MakeBool(v))
}

func setintconst(n *Node, v int64) {
	setconst(n, constant.MakeInt64(v))
}

// nodlit returns a new untyped constant with value v.
func nodlit(v constant.Value) *Node {
	n := nod(OLITERAL, nil, nil)
	if v.Kind() != constant.Unknown {
		n.Type = idealType(v.Kind())
	}
	n.SetVal(v)
	return n
}

func idealType(ct constant.Kind) *types.Type {
	switch ct {
	case CTSTR:
		return types.UntypedString
	case CTBOOL:
		return types.UntypedBool
	case CTINT:
		return types.UntypedInt
	case CTFLT:
		return types.UntypedFloat
	case CTCPLX:
		return types.UntypedComplex
	}
	Fatalf("unexpected Ctype: %v", ct)
	return nil
}

// defaultlit on both nodes simultaneously;
// if they're both ideal going in they better
// get the same type going out.
// force means must assign concrete (non-ideal) type.
// The results of defaultlit2 MUST be assigned back to l and r, e.g.
// 	n.Left, n.Right = defaultlit2(n.Left, n.Right, force)
func defaultlit2(l *Node, r *Node, force bool) (*Node, *Node) {
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
	if l.isNil() || r.isNil() {
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
		Fatalf("bad type %v", t)
		panic("unreachable")
	}

	if rank(t2) > rank(t1) {
		return t2
	}
	return t1
}

func defaultType(t *types.Type) *types.Type {
	if !t.IsUntyped() || t.Etype == TNIL {
		return t
	}

	switch t {
	case types.UntypedBool:
		return types.Types[TBOOL]
	case types.UntypedString:
		return types.Types[TSTRING]
	case types.UntypedInt:
		return types.Types[TINT]
	case types.UntypedRune:
		return types.Runetype
	case types.UntypedFloat:
		return types.Types[TFLOAT64]
	case types.UntypedComplex:
		return types.Types[TCOMPLEX128]
	}

	Fatalf("bad type %v", t)
	return nil
}

// TODO(mdempsky): Revisit and remove.
func smallintconst(n *Node) bool {
	if n.Op == OLITERAL {
		// TODO(mdempsky): uint32??
		v, ok := constant.Int64Val(n.Val())
		return ok && int64(int32(v)) == v
	}
	return false
}

// indexconst checks if Node n contains a constant expression
// representable as a non-negative int and returns its value.
// If n is not a constant expression, not representable as an
// integer, or negative, it returns -1. If n is too large, it
// returns -2.
func indexconst(n *Node) int64 {
	if n.Op != OLITERAL {
		return -1
	}

	v := toint(n.Val()) // toint returns argument unchanged if not representable as an *Mpint
	if v.Kind() != constant.Int || constant.Sign(v) < 0 {
		return -1
	}
	if doesoverflow(v, types.Types[TINT]) {
		return -2
	}

	vi, ok := constant.Int64Val(v)
	if !ok {
		Fatalf("index doesn't fit int64: %v, %v", n, v)
	}
	return vi
}

// isGoConst reports whether n is a Go language constant (as opposed to a
// compile-time constant).
//
// Expressions derived from nil, like string([]byte(nil)), while they
// may be known at compile time, are not Go language constants.
func (n *Node) isGoConst() bool {
	return n.Op == OLITERAL
}

func hascallchan(n *Node) bool {
	if n == nil {
		return false
	}
	switch n.Op {
	case OAPPEND,
		OCALL,
		OCALLFUNC,
		OCALLINTER,
		OCALLMETH,
		OCAP,
		OCLOSE,
		OCOMPLEX,
		OCOPY,
		ODELETE,
		OIMAG,
		OLEN,
		OMAKE,
		ONEW,
		OPANIC,
		OPRINT,
		OPRINTN,
		OREAL,
		ORECOVER,
		ORECV:
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
func (s *constSet) add(pos src.XPos, n *Node, what, where string) {
	if n.Op == OCONVIFACE && n.Implicit() {
		n = n.Left
	}

	if !n.isGoConst() {
		return
	}
	if n.Type.IsUntyped() {
		Fatalf("%v is untyped", n)
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
		typ = types.Types[TUINT8]
	case types.Runetype:
		typ = types.Types[TINT32]
	}
	k := constSetKey{typ, n.ValueInterface()}

	if hasUniquePos(n) {
		pos = n.Pos
	}

	if s.m == nil {
		s.m = make(map[constSetKey]src.XPos)
	}

	if prevPos, isDup := s.m[k]; isDup {
		yyerrorl(pos, "duplicate %s %s in %s\n\tprevious %s at %v",
			what, nodeAndVal(n), where,
			what, linestr(prevPos))
	} else {
		s.m[k] = pos
	}
}

// nodeAndVal reports both an expression and its constant value, if
// the latter is non-obvious.
//
// TODO(mdempsky): This could probably be a fmt.go flag.
func nodeAndVal(n *Node) string {
	show := n.String()
	val := n.ValueInterface()
	if s := fmt.Sprintf("%#v", val); show != s {
		show += " (value " + s + ")"
	}
	return show
}

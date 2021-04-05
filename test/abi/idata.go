// run

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Excerpted from go/constant/value.go to capture a bug from there.

package main

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
)

// Kind specifies the kind of value represented by a Value.
type Kind int

const (
	// unknown values
	Unknown Kind = iota

	// non-numeric values
	Bool
	String

	// numeric values
	Int
	Float
	Complex
)


type (
	unknownVal struct{}
	boolVal    bool
	int64Val   int64                    // Int values representable as an int64
	intVal     struct{ val *big.Int }   // Int values not representable as an int64
	ratVal     struct{ val *big.Rat }   // Float values representable as a fraction
	floatVal   struct{ val *big.Float } // Float values not representable as a fraction
	complexVal struct{ re, im Value }
)

const prec = 512

func (unknownVal) Kind() Kind { return Unknown }
func (boolVal) Kind() Kind    { return Bool }
func (int64Val) Kind() Kind   { return Int }
func (intVal) Kind() Kind     { return Int }
func (ratVal) Kind() Kind     { return Float }
func (floatVal) Kind() Kind   { return Float }
func (complexVal) Kind() Kind { return Complex }

func (unknownVal) String() string { return "unknown" }
func (x boolVal) String() string  { return strconv.FormatBool(bool(x)) }

func (x int64Val) String() string { return strconv.FormatInt(int64(x), 10) }
func (x intVal) String() string   { return x.val.String() }
func (x ratVal) String() string   { return rtof(x).String() }

// String returns a decimal approximation of the Float value.
func (x floatVal) String() string {
	f := x.val

	// Don't try to convert infinities (will not terminate).
	if f.IsInf() {
		return f.String()
	}

	// Use exact fmt formatting if in float64 range (common case):
	// proceed if f doesn't underflow to 0 or overflow to inf.
	if x, _ := f.Float64(); f.Sign() == 0 == (x == 0) && !math.IsInf(x, 0) {
		return fmt.Sprintf("%.6g", x)
	}

	// Out of float64 range. Do approximate manual to decimal
	// conversion to avoid precise but possibly slow Float
	// formatting.
	// f = mant * 2**exp
	var mant big.Float
	exp := f.MantExp(&mant) // 0.5 <= |mant| < 1.0

	// approximate float64 mantissa m and decimal exponent d
	// f ~ m * 10**d
	m, _ := mant.Float64()                     // 0.5 <= |m| < 1.0
	d := float64(exp) * (math.Ln2 / math.Ln10) // log_10(2)

	// adjust m for truncated (integer) decimal exponent e
	e := int64(d)
	m *= math.Pow(10, d-float64(e))

	// ensure 1 <= |m| < 10
	switch am := math.Abs(m); {
	case am < 1-0.5e-6:
		// The %.6g format below rounds m to 5 digits after the
		// decimal point. Make sure that m*10 < 10 even after
		// rounding up: m*10 + 0.5e-5 < 10 => m < 1 - 0.5e6.
		m *= 10
		e--
	case am >= 10:
		m /= 10
		e++
	}

	return fmt.Sprintf("%.6ge%+d", m, e)
}

func (x complexVal) String() string { return fmt.Sprintf("(%s + %si)", x.re, x.im) }

func (unknownVal) implementsValue() {}
func (boolVal) implementsValue()    {}
func (int64Val) implementsValue()   {}
func (ratVal) implementsValue()     {}
func (intVal) implementsValue()     {}
func (floatVal) implementsValue()   {}
func (complexVal) implementsValue() {}

func newInt() *big.Int     { return new(big.Int) }
func newRat() *big.Rat     { return new(big.Rat) }
func newFloat() *big.Float { return new(big.Float).SetPrec(prec) }

func i64toi(x int64Val) intVal   { return intVal{newInt().SetInt64(int64(x))} }
//go:noinline
//go:registerparams
func i64tor(x int64Val) ratVal   { return ratVal{newRat().SetInt64(int64(x))} }
func i64tof(x int64Val) floatVal { return floatVal{newFloat().SetInt64(int64(x))} }
//go:noinline
//go:registerparams
func itor(x intVal) ratVal       { return ratVal{newRat().SetInt(x.val)} }
//go:noinline
//go:registerparams
func itof(x intVal) floatVal     { return floatVal{newFloat().SetInt(x.val)} }
func rtof(x ratVal) floatVal     { return floatVal{newFloat().SetRat(x.val)} }
func vtoc(x Value) complexVal    { return complexVal{x, int64Val(0)} }

// A Value represents the value of a Go constant.
type Value interface {
	// Kind returns the value kind.
	Kind() Kind

	// String returns a short, quoted (human-readable) form of the value.
	// For numeric values, the result may be an approximation;
	// for String values the result may be a shortened string.
	String() string
}

// ToFloat converts x to a Float value if x is representable as a Float.
// Otherwise it returns an Unknown.
//go:noinline
//go:registerparams
func ToFloat(x Value) Value {
	switch x := x.(type) {
	case int64Val:
		return i64tor(x) // x is always a small int
	case intVal:
		if smallInt(x.val) {
			return itor(x)
		}
		return itof(x)
	case ratVal, floatVal:
		return x
	case complexVal:
		if Sign(x.im) == 0 {
			return ToFloat(x.re)
		}
	}
	return unknownVal{}
}

// Permit fractions with component sizes up to maxExp
// before switching to using floating-point numbers.
const maxExp = 4 << 10

// smallInt reports whether x would lead to "reasonably"-sized fraction
// if converted to a *big.Rat.
//go:noinline
//go:registerparams
func smallInt(x *big.Int) bool {
	return x.BitLen() < maxExp
}

// Sign returns -1, 0, or 1 depending on whether x < 0, x == 0, or x > 0;
// x must be numeric or Unknown. For complex values x, the sign is 0 if x == 0,
// otherwise it is != 0. If x is Unknown, the result is 1.
//go:noinline
//go:registerparams
func Sign(x Value) int {
	switch x := x.(type) {
	case int64Val:
		switch {
		case x < 0:
			return -1
		case x > 0:
			return 1
		}
		return 0
	case intVal:
		return x.val.Sign()
	case ratVal:
		return x.val.Sign()
	case floatVal:
		return x.val.Sign()
	case complexVal:
		return Sign(x.re) | Sign(x.im)
	case unknownVal:
		return 1 // avoid spurious division by zero errors
	default:
		panic(fmt.Sprintf("%v not numeric", x))
	}
}


func main() {
	v := ratVal{big.NewRat(22,7)}
	s := ToFloat(v).String()
	fmt.Printf("s=%s\n", s)
}

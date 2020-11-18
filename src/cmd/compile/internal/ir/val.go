// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ir

import "cmd/compile/internal/base"

// Ctype describes the constant kind of an "ideal" (untyped) constant.
type Ctype uint8

const (
	CTxxx Ctype = iota

	CTINT
	CTRUNE
	CTFLT
	CTCPLX
	CTSTR
	CTBOOL
	CTNIL
)

type Val struct {
	// U contains one of:
	// bool     bool when Ctype() == CTBOOL
	// *Mpint   int when Ctype() == CTINT, rune when Ctype() == CTRUNE
	// *Mpflt   float when Ctype() == CTFLT
	// *Mpcplx  pair of floats when Ctype() == CTCPLX
	// string   string when Ctype() == CTSTR
	// *Nilval  when Ctype() == CTNIL
	U interface{}
}

func (v Val) Ctype() Ctype {
	switch x := v.U.(type) {
	default:
		base.Fatal("unexpected Ctype for %T", v.U)
		panic("unreachable")
	case nil:
		return CTxxx
	case *NilVal:
		return CTNIL
	case bool:
		return CTBOOL
	case *Int:
		if x.Rune {
			return CTRUNE
		}
		return CTINT
	case *Float:
		return CTFLT
	case *Complex:
		return CTCPLX
	case string:
		return CTSTR
	}
}

func Eqval(a, b Val) bool {
	if a.Ctype() != b.Ctype() {
		return false
	}
	switch x := a.U.(type) {
	default:
		base.Fatal("unexpected Ctype for %T", a.U)
		panic("unreachable")
	case *NilVal:
		return true
	case bool:
		y := b.U.(bool)
		return x == y
	case *Int:
		y := b.U.(*Int)
		return x.Cmp(y) == 0
	case *Float:
		y := b.U.(*Float)
		return x.Cmp(y) == 0
	case *Complex:
		y := b.U.(*Complex)
		return x.Real.Cmp(&y.Real) == 0 && x.Imag.Cmp(&y.Imag) == 0
	case string:
		y := b.U.(string)
		return x == y
	}
}

// Interface returns the constant value stored in v as an interface{}.
// It returns int64s for ints and runes, float64s for floats,
// complex128s for complex values, and nil for constant nils.
func (v Val) Interface() interface{} {
	switch x := v.U.(type) {
	default:
		base.Fatal("unexpected Interface for %T", v.U)
		panic("unreachable")
	case *NilVal:
		return nil
	case bool, string:
		return x
	case *Int:
		return x.Int64()
	case *Float:
		return x.Float64()
	case *Complex:
		return complex(x.Real.Float64(), x.Imag.Float64())
	}
}

type NilVal struct{}

func ConstType(n INode) Ctype {
	if n == nil || n.Op() != OLITERAL {
		return CTxxx
	}
	return n.Val().Ctype()
}

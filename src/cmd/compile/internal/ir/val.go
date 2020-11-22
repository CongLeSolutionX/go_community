// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ir

import (
	"go/constant"

	"cmd/compile/internal/base"
	"cmd/compile/internal/types"
)

type Val struct {
	// U contains one of:
	// bool     bool when Ctype() == CTBOOL
	// *Mpint   int when Ctype() == CTINT
	// *Mpflt   float when Ctype() == CTFLT
	// *Mpcplx  pair of floats when Ctype() == CTCPLX
	// string   string when Ctype() == CTSTR
	U interface{}
}

func (v Val) Kind() constant.Kind {
	switch v.U.(type) {
	default:
		base.Fatalf("unexpected Ctype for %T", v.U)
		panic("unreachable")
	case nil:
		return constant.Unknown
	case bool:
		return constant.Bool
	case *Int:
		return constant.Int
	case *Float:
		return constant.Float
	case *Complex:
		return constant.Complex
	case string:
		return constant.String
	}
}

// Interface returns the constant value stored in v as an interface{}.
// It returns int64s for ints and runes, float64s for floats,
// complex128s for complex values, and nil for constant nils.
func (v Val) Interface() interface{} {
	switch x := v.U.(type) {
	default:
		base.Fatalf("unexpected Interface for %T", v.U)
		panic("unreachable")
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

func ConstType(n *Node) constant.Kind {
	if n == nil || n.Op() != OLITERAL {
		return constant.Unknown
	}
	return n.Val().Kind()
}

func AssertValidTypeForConst(t *types.Type, v Val) {
	if !ValidTypeForConst(t, v) {
		base.Fatalf("%v does not represent %v", t, v)
	}
}

func ValidTypeForConst(t *types.Type, v Val) bool {
	if !t.IsUntyped() {
		// TODO(mdempsky): Stricter handling of typed types.
		return true
	}

	vt := idealType(v.Kind())
	return t == vt || (t == types.UntypedRune && vt == types.UntypedInt)
}

// nodlit returns a new untyped constant with value v.
func NewLiteral(v Val) *Node {
	n := Nod(OLITERAL, nil, nil)
	n.SetType(idealType(v.Kind()))
	n.SetVal(v)
	return n
}

func idealType(ct constant.Kind) *types.Type {
	switch ct {
	case constant.String:
		return types.UntypedString
	case constant.Bool:
		return types.UntypedBool
	case constant.Int:
		return types.UntypedInt
	case constant.Float:
		return types.UntypedFloat
	case constant.Complex:
		return types.UntypedComplex
	}
	base.Fatalf("unexpected Ctype: %v", ct)
	return nil
}

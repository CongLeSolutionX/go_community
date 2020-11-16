// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ir

import (
	"fmt"
	"math/big"

	"cmd/compile/internal/base"
)

// implements integer arithmetic

// Mpint represents an integer constant.
type Int struct {
	Val  big.Int
	Ovf  bool // set if Val overflowed compiler limit (sticky)
	Rune bool // set if syntax indicates default type rune
}

func (a *Int) SetOverflow() {
	a.Val.SetUint64(1) // avoid spurious div-zero errors
	a.Ovf = true
}

func (a *Int) CheckOverflow(extra int) bool {
	// We don't need to be precise here, any reasonable upper limit would do.
	// For now, use existing limit so we pass all the tests unchanged.
	if a.Val.BitLen()+extra > Mpprec {
		a.SetOverflow()
	}
	return a.Ovf
}

func (a *Int) Set(b *Int) {
	a.Val.Set(&b.Val)
}

func (a *Int) SetFloat(b *Float) bool {
	// avoid converting huge floating-point numbers to integers
	// (2*Mpprec is large enough to permit all tests to pass)
	if b.Val.MantExp(nil) > 2*Mpprec {
		a.SetOverflow()
		return false
	}

	if _, acc := b.Val.Int(&a.Val); acc == big.Exact {
		return true
	}

	const delta = 16 // a reasonably small number of bits > 0
	var t big.Float
	t.SetPrec(Mpprec - delta)

	// try rounding down a little
	t.SetMode(big.ToZero)
	t.Set(&b.Val)
	if _, acc := t.Int(&a.Val); acc == big.Exact {
		return true
	}

	// try rounding up a little
	t.SetMode(big.AwayFromZero)
	t.Set(&b.Val)
	if _, acc := t.Int(&a.Val); acc == big.Exact {
		return true
	}

	a.Ovf = false
	return false
}

func (a *Int) Add(b *Int) {
	if a.Ovf || b.Ovf {
		if base.Errors() == 0 {
			base.Fatal("ovf in Mpint Add")
		}
		a.SetOverflow()
		return
	}

	a.Val.Add(&a.Val, &b.Val)

	if a.CheckOverflow(0) {
		base.Error("constant addition overflow")
	}
}

func (a *Int) Sub(b *Int) {
	if a.Ovf || b.Ovf {
		if base.Errors() == 0 {
			base.Fatal("ovf in Mpint Sub")
		}
		a.SetOverflow()
		return
	}

	a.Val.Sub(&a.Val, &b.Val)

	if a.CheckOverflow(0) {
		base.Error("constant subtraction overflow")
	}
}

func (a *Int) Mul(b *Int) {
	if a.Ovf || b.Ovf {
		if base.Errors() == 0 {
			base.Fatal("ovf in Mpint Mul")
		}
		a.SetOverflow()
		return
	}

	a.Val.Mul(&a.Val, &b.Val)

	if a.CheckOverflow(0) {
		base.Error("constant multiplication overflow")
	}
}

func (a *Int) Quo(b *Int) {
	if a.Ovf || b.Ovf {
		if base.Errors() == 0 {
			base.Fatal("ovf in Mpint Quo")
		}
		a.SetOverflow()
		return
	}

	a.Val.Quo(&a.Val, &b.Val)

	if a.CheckOverflow(0) {
		// can only happen for div-0 which should be checked elsewhere
		base.Error("constant division overflow")
	}
}

func (a *Int) Rem(b *Int) {
	if a.Ovf || b.Ovf {
		if base.Errors() == 0 {
			base.Fatal("ovf in Mpint Rem")
		}
		a.SetOverflow()
		return
	}

	a.Val.Rem(&a.Val, &b.Val)

	if a.CheckOverflow(0) {
		// should never happen
		base.Error("constant modulo overflow")
	}
}

func (a *Int) Or(b *Int) {
	if a.Ovf || b.Ovf {
		if base.Errors() == 0 {
			base.Fatal("ovf in Mpint Or")
		}
		a.SetOverflow()
		return
	}

	a.Val.Or(&a.Val, &b.Val)
}

func (a *Int) And(b *Int) {
	if a.Ovf || b.Ovf {
		if base.Errors() == 0 {
			base.Fatal("ovf in Mpint And")
		}
		a.SetOverflow()
		return
	}

	a.Val.And(&a.Val, &b.Val)
}

func (a *Int) AndNot(b *Int) {
	if a.Ovf || b.Ovf {
		if base.Errors() == 0 {
			base.Fatal("ovf in Mpint AndNot")
		}
		a.SetOverflow()
		return
	}

	a.Val.AndNot(&a.Val, &b.Val)
}

func (a *Int) Xor(b *Int) {
	if a.Ovf || b.Ovf {
		if base.Errors() == 0 {
			base.Fatal("ovf in Mpint Xor")
		}
		a.SetOverflow()
		return
	}

	a.Val.Xor(&a.Val, &b.Val)
}

func (a *Int) Lsh(b *Int) {
	if a.Ovf || b.Ovf {
		if base.Errors() == 0 {
			base.Fatal("ovf in Mpint Lsh")
		}
		a.SetOverflow()
		return
	}

	s := b.Int64()
	if s < 0 || s >= Mpprec {
		msg := "shift count too large"
		if s < 0 {
			msg = "invalid negative shift count"
		}
		base.Error("%s: %d", msg, s)
		a.SetInt64(0)
		return
	}

	if a.CheckOverflow(int(s)) {
		base.Error("constant shift overflow")
		return
	}
	a.Val.Lsh(&a.Val, uint(s))
}

func (a *Int) Rsh(b *Int) {
	if a.Ovf || b.Ovf {
		if base.Errors() == 0 {
			base.Fatal("ovf in Mpint Rsh")
		}
		a.SetOverflow()
		return
	}

	s := b.Int64()
	if s < 0 {
		base.Error("invalid negative shift count: %d", s)
		if a.Val.Sign() < 0 {
			a.SetInt64(-1)
		} else {
			a.SetInt64(0)
		}
		return
	}

	a.Val.Rsh(&a.Val, uint(s))
}

func (a *Int) Cmp(b *Int) int {
	return a.Val.Cmp(&b.Val)
}

func (a *Int) CmpInt64(c int64) int {
	if c == 0 {
		return a.Val.Sign() // common case shortcut
	}
	return a.Val.Cmp(big.NewInt(c))
}

func (a *Int) Neg() {
	a.Val.Neg(&a.Val)
}

func (a *Int) Int64() int64 {
	if a.Ovf {
		if base.Errors() == 0 {
			base.Fatal("constant overflow")
		}
		return 0
	}

	return a.Val.Int64()
}

func (a *Int) SetInt64(c int64) {
	a.Val.SetInt64(c)
}

func (a *Int) SetString(as string) {
	_, ok := a.Val.SetString(as, 0)
	if !ok {
		// The lexer checks for correct syntax of the literal
		// and reports detailed errors. Thus SetString should
		// never fail (in theory it might run out of memory,
		// but that wouldn't be reported as an error here).
		base.Fatal("malformed integer constant: %s", as)
		return
	}
	if a.CheckOverflow(0) {
		base.Error("constant too large: %s", as)
	}
}

func (a *Int) GoString() string {
	return a.Val.String()
}

func (a *Int) String() string {
	return fmt.Sprintf("%#x", &a.Val)
}

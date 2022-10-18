// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// bbmod is a compatibility layer between math/big and bigmod, which allows
// bigmod not to depend on math/big.
package bbmod

import (
	"crypto/internal/bigmod"
	"errors"
	"math/big"
	"unsafe"
)

// NewModulusFromBig returns a new [bigmod.Modulus] from a [big.Int]. The input
// must be an odd positive number.
func NewModulusFromBig(n *big.Int) (*bigmod.Modulus, error) {
	if n.Sign() <= 0 {
		return nil, errors.New("bigmod: Modulus can't be negative")
	}
	x := n.Bits()
	limbs := unsafe.Slice((*uint)(&x[0]), len(x))
	return bigmod.NewModulusFromBits(limbs)
}

func SetBig(i *bigmod.Int, n *big.Int) (*bigmod.Int, error) {
	if n.Sign() <= 0 {
		return nil, errors.New("bigmod: Modulus can't be negative")
	}
	x := n.Bits()
	limbs := unsafe.Slice((*uint)(&x[0]), len(x))
	return i.SetBits(limbs)
}

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package fiat implements prime order fields using formally verified algorithms
// from the Fiat Cryptography project.
package fiat

import (
	"crypto/subtle"
	"errors"
)

// P521Element is an integer modulo 2^521 - 1.
//
// The zero value is a valid zero element.
type P521Element struct {
	// This element has the following bounds, which are tighter than
	// the output bounds of some operations. Those operations must be
	// followed by a carry.
	//
	// [0x0 ~> 0x400000000000000], [0x0 ~> 0x400000000000000], [0x0 ~> 0x400000000000000],
	// [0x0 ~> 0x400000000000000], [0x0 ~> 0x400000000000000], [0x0 ~> 0x400000000000000],
	// [0x0 ~> 0x400000000000000], [0x0 ~> 0x400000000000000], [0x0 ~> 0x200000000000000]
	x p521TightFieldElement
}

// xl returns a reference to the underlying element as a loose element. The
// return value can be used as an input to a fiat function because the bounds of
// p521TightFieldElement are strictly tighter than p521LooseFieldElement. If it
// is used as an output, though, it MUST be followed by a p521Carry call.
func (e *P521Element) xl() *p521LooseFieldElement {
	return (*p521LooseFieldElement)(&e.x)
}

// One sets e = 1, and returns e.
func (e *P521Element) One() *P521Element {
	*e = P521Element{}
	e.x[0] = 1
	return e
}

// Equal returns 1 if e == t, and zero otherwise.
func (e *P521Element) Equal(t *P521Element) int {
	eBytes := e.Bytes()
	tBytes := t.Bytes()
	return subtle.ConstantTimeCompare(eBytes, tBytes)
}

var p521ZeroEncoding = new(P521Element).Bytes()

// IsZero returns 1 if e == 0, and zero otherwise.
func (e *P521Element) IsZero() int {
	eBytes := e.Bytes()
	return subtle.ConstantTimeCompare(eBytes, p521ZeroEncoding)
}

// Set sets e = t, and returns e.
func (e *P521Element) Set(t *P521Element) *P521Element {
	e.x = t.x
	return e
}

// Bytes returns the 66-byte big-endian encoding of e.
func (e *P521Element) Bytes() []byte {
	// This function is outlined to make the allocations inline in the caller
	// rather than happen on the heap.
	var out [66]byte
	return e.bytes(&out)
}

func (e *P521Element) bytes(out *[66]byte) []byte {
	p521ToBytes(out, &e.x)
	invertEndianness(out[:])
	return out[:]
}

// SetBytes sets e = v, where v is a big-endian 66-byte encoding, and returns
// e. If v is not 66 bytes or it encodes a value higher than 2^521 - 1, SetBytes
// returns nil and an error, and e is unchanged.
func (e *P521Element) SetBytes(v []byte) (*P521Element, error) {
	if len(v) != 66 || v[0] > 1 {
		return nil, errors.New("invalid P-521 field encoding")
	}
	var in [66]byte
	copy(in[:], v)
	invertEndianness(in[:])
	p521FromBytes(&e.x, &in)
	return e, nil
}

func invertEndianness(v []byte) {
	for i := 0; i < len(v)/2; i++ {
		v[i], v[len(v)-1-i] = v[len(v)-1-i], v[i]
	}
}

// Add sets e = t1 + t2, and returns e.
func (e *P521Element) Add(t1, t2 *P521Element) *P521Element {
	p521Add(e.xl(), &t1.x, &t2.x)
	p521Carry(&e.x, e.xl())
	return e
}

// Sub sets e = t1 - t2, and returns e.
func (e *P521Element) Sub(t1, t2 *P521Element) *P521Element {
	p521Sub(e.xl(), &t1.x, &t2.x)
	p521Carry(&e.x, e.xl())
	return e
}

// Mul sets e = t1 * t2, and returns e.
func (e *P521Element) Mul(t1, t2 *P521Element) *P521Element {
	p521CarryMul(&e.x, t1.xl(), t2.xl())
	return e
}

// Square sets e = t * t, and returns e.
func (e *P521Element) Square(t *P521Element) *P521Element {
	p521CarrySquare(&e.x, t.xl())
	return e
}

// Select sets v to a if cond == 1, and to b if cond == 0.
func (v *P521Element) Select(a, b *P521Element, cond int) *P521Element {
	p521Selectznz((*[9]uint64)(&v.x), p521Uint1(cond),
		(*[9]uint64)(&b.x), (*[9]uint64)(&a.x))
	return v
}

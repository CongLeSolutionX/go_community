// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fiat

import (
	"crypto/subtle"
	"errors"
)

// P224Element is an integer modulo 2^224 - 2^96 + 1.
//
// The zero value is a valid zero element.
type P224Element struct {
	// Values are represented internally always in the Montgomery domain, and
	// converted in Bytes and SetBytes.
	x p224MontgomeryDomainFieldElement
}

// One sets e = 1, and returns e.
func (e *P224Element) One() *P224Element {
	p224SetOne(&e.x)
	return e
}

// Equal returns 1 if e == t, and zero otherwise.
func (e *P224Element) Equal(t *P224Element) int {
	eBytes := e.Bytes()
	tBytes := t.Bytes()
	return subtle.ConstantTimeCompare(eBytes, tBytes)
}

var p224ZeroEncoding = new(P224Element).Bytes()

// IsZero returns 1 if e == 0, and zero otherwise.
func (e *P224Element) IsZero() int {
	eBytes := e.Bytes()
	return subtle.ConstantTimeCompare(eBytes, p224ZeroEncoding)
}

// Set sets e = t, and returns e.
func (e *P224Element) Set(t *P224Element) *P224Element {
	e.x = t.x
	return e
}

// Bytes returns the 28-byte big-endian encoding of e.
func (e *P224Element) Bytes() []byte {
	// This function is outlined to make the allocations inline in the caller
	// rather than happen on the heap.
	var out [28]byte
	return e.bytes(&out)
}

func (e *P224Element) bytes(out *[28]byte) []byte {
	var tmp p224NonMontgomeryDomainFieldElement
	p224FromMontgomery(&tmp, &e.x)
	p224ToBytes(out, (*[4]uint64)(&tmp))
	invertEndianness(out[:])
	return out[:]
}

var p224MinusOneEncoding = new(P224Element).Sub(
	new(P224Element), new(P224Element).One()).Bytes()

// SetBytes sets e = v, where v is a big-endian 28-byte encoding, and returns e.
// If v is not 28 bytes or it encodes a value higher than 2^224 - 2^96 + 1,
// SetBytes returns nil and an error, and e is unchanged.
func (e *P224Element) SetBytes(v []byte) (*P224Element, error) {
	if len(v) != 28 {
		return nil, errors.New("invalid P-224 field encoding")
	}
	for i := range v {
		if v[i] < p224MinusOneEncoding[i] {
			break
		}
		if v[i] > p224MinusOneEncoding[i] {
			return nil, errors.New("invalid P-224 field encoding")
		}
	}
	var in [28]byte
	copy(in[:], v)
	invertEndianness(in[:])
	var tmp p224NonMontgomeryDomainFieldElement
	p224FromBytes((*[4]uint64)(&tmp), &in)
	p224ToMontgomery(&e.x, &tmp)
	return e, nil
}

// Add sets e = t1 + t2, and returns e.
func (e *P224Element) Add(t1, t2 *P224Element) *P224Element {
	p224Add(&e.x, &t1.x, &t2.x)
	return e
}

// Sub sets e = t1 - t2, and returns e.
func (e *P224Element) Sub(t1, t2 *P224Element) *P224Element {
	p224Sub(&e.x, &t1.x, &t2.x)
	return e
}

// Mul sets e = t1 * t2, and returns e.
func (e *P224Element) Mul(t1, t2 *P224Element) *P224Element {
	p224Mul(&e.x, &t1.x, &t2.x)
	return e
}

// Square sets e = t * t, and returns e.
func (e *P224Element) Square(t *P224Element) *P224Element {
	p224Square(&e.x, &t.x)
	return e
}

// Select sets v to a if cond == 1, and to b if cond == 0.
func (v *P224Element) Select(a, b *P224Element, cond int) *P224Element {
	p224Selectznz((*[4]uint64)(&v.x), p224Uint1(cond),
		(*[4]uint64)(&b.x), (*[4]uint64)(&a.x))
	return v
}

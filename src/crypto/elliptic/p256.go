// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !amd64 && !arm64

package elliptic

import (
	"crypto/elliptic/internal/nistec"
	"crypto/rand"
	"math/big"
)

var (
	p256Params *CurveParams

	// RInverse contains 1/R mod p - the inverse of the Montgomery constant
	// (2**257).
	p256RInverse *big.Int
)

func initP256() {
	// See FIPS 186-3, section D.2.3
	p256Params = &CurveParams{Name: "P-256"}
	p256Params.P, _ = new(big.Int).SetString("115792089210356248762697446949407573530086143415290314195533631308867097853951", 10)
	p256Params.N, _ = new(big.Int).SetString("115792089210356248762697446949407573529996955224135760342422259061068512044369", 10)
	p256Params.B, _ = new(big.Int).SetString("5ac635d8aa3a93e7b3ebbd55769886bc651d06b0cc53b0f63bce3c3e27d2604b", 16)
	p256Params.Gx, _ = new(big.Int).SetString("6b17d1f2e12c4247f8bce6e563a440f277037d812deb33a0f4a13945d898c296", 16)
	p256Params.Gy, _ = new(big.Int).SetString("4fe342e2fe1a7f9b8ee7eb4a7c0f9e162bce33576b315ececbb6406837bf51f5", 16)
	p256Params.BitSize = 256

	p256RInverse, _ = new(big.Int).SetString("7fffffff00000001fffffffe8000000100000000ffffffff0000000180000000", 16)

	// Arch-specific initialization, i.e. let a platform dynamically pick a P256 implementation
	initP256Arch()
}

// p256Curve is a pure-Go Curve implementation based on nistec.P256Point, used
// when there is no assembly implementation available.
//
// It's a wrapper that exposes the big.Int-based Curve interface and encodes the
// legacy idiosyncrasies it requires, such as invalid and infinity point
// handling.
//
// To interact with the nistec package, points are encoded into and decoded from
// properly formatted byte slices. All big.Int use is limited to this package.
// Encoding and decoding is 1/1000th of the runtime of a scalar multiplication,
// so the overhead is acceptable.
type p256Curve struct {
	params *CurveParams
}

func (curve p256Curve) Params() *CurveParams {
	return curve.params
}

func (curve p256Curve) IsOnCurve(x, y *big.Int) bool {
	// IsOnCurve is documented to reject (0, 0), so we don't use
	// p256PointFromAffine, but let SetBytes reject the invalid Marshal output.
	_, err := nistec.NewP256Point().SetBytes(Marshal(curve, x, y))
	return err == nil
}

func p256PointFromAffine(x, y *big.Int) (p *nistec.P256Point, ok bool) {
	// (0, 0) is by convention the point at infinity, which can't be represented
	// in affine coordinates. Marshal incorrectly encodes it as an uncompressed
	// point, which SetBytes correctly rejects. See Issue 37294.
	if x.Sign() == 0 && y.Sign() == 0 {
		return nistec.NewP256Point(), true
	}
	p, err := nistec.NewP256Point().SetBytes(Marshal(P256(), x, y))
	if err != nil {
		return nil, false
	}
	return p, true
}

func p256PointToAffine(p *nistec.P256Point) (x, y *big.Int) {
	out := p.Bytes()
	if len(out) == 1 && out[0] == 0 {
		// This is the correct encoding of the point at infinity, which
		// Unmarshal does not support. See Issue 37294.
		return new(big.Int), new(big.Int)
	}
	x, y = Unmarshal(P256(), out)
	if x == nil {
		panic("crypto/elliptic: internal error: Unmarshal rejected a valid point encoding")
	}
	return x, y
}

// p256RandomPoint returns a random point on the curve. It's used when Add,
// Double, or ScalarMult are fed a point not on the curve, which is undefined
// behavior. Originally, we used to do the math on it anyway (which allows
// invalid curve attacks) and relied on the caller and Unmarshal to avoid this
// happening in the first place. Now, we just can't construct a nistec.P256Point
// for an invalid pair of coordinates, because that API is safer. If we panic,
// we risk introducing a DoS. If we return nil, we risk a panic. If we return
// the input, ecdsa.Verify might fail open. The safest course seems to be to
// return a valid, random point, which hopefully won't help the attacker.
func p256RandomPoint() (x, y *big.Int) {
	_, x, y, err := GenerateKey(P256(), rand.Reader)
	if err != nil {
		panic("crypto/elliptic: failed to generate random point")
	}
	return x, y
}

func (p256Curve) Add(x1, y1, x2, y2 *big.Int) (*big.Int, *big.Int) {
	p1, ok := p256PointFromAffine(x1, y1)
	if !ok {
		return p256RandomPoint()
	}
	p2, ok := p256PointFromAffine(x2, y2)
	if !ok {
		return p256RandomPoint()
	}
	return p256PointToAffine(p1.Add(p1, p2))
}

func (p256Curve) Double(x1, y1 *big.Int) (*big.Int, *big.Int) {
	p, ok := p256PointFromAffine(x1, y1)
	if !ok {
		return p256RandomPoint()
	}
	return p256PointToAffine(p.Double(p))
}

func (p256Curve) ScalarMult(Bx, By *big.Int, scalar []byte) (*big.Int, *big.Int) {
	p, ok := p256PointFromAffine(Bx, By)
	if !ok {
		return p256RandomPoint()
	}
	return p256PointToAffine(p.ScalarMult(p, scalar))
}

func (p256Curve) ScalarBaseMult(scalar []byte) (*big.Int, *big.Int) {
	p := nistec.NewP256Generator()
	return p256PointToAffine(p.ScalarMult(p, scalar))
}

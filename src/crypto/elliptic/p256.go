// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package elliptic

// P-256 is implemented by various different backends, including a generic
// 32-bit constant-time one at p256_generic.go, which is used when assembly
// implementations are not available, or not appropriate for the hardware.

import "math/big"

var p256Params *CurveParams

func initP256() {
	// See FIPS 186-3, section D.2.3
	p256Params = &CurveParams{Name: "P-256"}
	p256Params.P, _ = new(big.Int).SetString("115792089210356248762697446949407573530086143415290314195533631308867097853951", 10)
	p256Params.N, _ = new(big.Int).SetString("115792089210356248762697446949407573529996955224135760342422259061068512044369", 10)
	p256Params.B, _ = new(big.Int).SetString("5ac635d8aa3a93e7b3ebbd55769886bc651d06b0cc53b0f63bce3c3e27d2604b", 16)
	p256Params.Gx, _ = new(big.Int).SetString("6b17d1f2e12c4247f8bce6e563a440f277037d812deb33a0f4a13945d898c296", 16)
	p256Params.Gy, _ = new(big.Int).SetString("4fe342e2fe1a7f9b8ee7eb4a7c0f9e162bce33576b315ececbb6406837bf51f5", 16)
	p256Params.BitSize = 256

	// Arch-specific initialization, i.e. let a platform dynamically pick a P256 implementation
	initP256Arch()
}

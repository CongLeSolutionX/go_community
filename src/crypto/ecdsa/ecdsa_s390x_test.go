// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build s390x,!gccgo

package ecdsa

import (
	"crypto/elliptic"
	"testing"
)

// CurveParams contains the parameters of an elliptic curve and also provides
// a generic, non-constant time implementation of Curve.
type CurveParamss390x struct {
	elliptic.CurveParams
}

func (curve *CurveParamss390x) init(cp *elliptic.CurveParams) {
	curve.CurveParams = *cp
}

var p256s390x *CurveParamss390x
var p384s390x *CurveParamss390x
var p521s390x *CurveParamss390x

func init() {
	initP256s390x()
	initP384s390x()
	initP521s390x()
}

func initP256s390x() {
	p256s390x = new(CurveParamss390x)
	p256s390x.init(elliptic.P256().Params())
	p256s390x.Name = "P-256_GENERIC_OVERRIDE"
}

func initP384s390x() {
	p384s390x = new(CurveParamss390x)
	p384s390x.init(elliptic.P384().Params())
	p384s390x.Name = "P-384_GENERIC_OVERRIDE"
}

func initP521s390x() {
	p521s390x = new(CurveParamss390x)
	p521s390x.init(elliptic.P521().Params())
	p521s390x.Name = "P-521_GENERIC_OVERRIDE"
}

func TestS390xKeyGeneration(t *testing.T) {
	testKeyGeneration(t, p256s390x, "p256")
	testKeyGeneration(t, p384s390x, "p384")
	testKeyGeneration(t, p521s390x, "p521")
}

func TestS390xSignAndVerify(t *testing.T) {
	testSignAndVerify(t, p256s390x, "p256")
	testSignAndVerify(t, p384s390x, "p384")
	testSignAndVerify(t, p521s390x, "p521")
}

func TestS390xNonceSafety(t *testing.T) {
	testNonceSafety(t, p256s390x, "p256")
	testNonceSafety(t, p384s390x, "p384")
	testNonceSafety(t, p521s390x, "p521")
}

func TestS390xINDCCA(t *testing.T) {
	testINDCCA(t, p256s390x, "p256")
	testINDCCA(t, p384s390x, "p384")
	testINDCCA(t, p521s390x, "p521")
}

func TestS390NegativeInputs(t *testing.T) {
	testNegativeInputs(t, p256s390x, "p256")
	testNegativeInputs(t, p384s390x, "p384")
	testNegativeInputs(t, p521s390x, "p521")
}

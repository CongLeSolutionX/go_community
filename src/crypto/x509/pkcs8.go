// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package x509

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509/pkix"
	"encoding/asn1"
	"errors"
	"fmt"
)

// pkcs8 reflects an ASN.1, PKCS#8 PrivateKey. See
// ftp://ftp.rsasecurity.com/pub/pkcs/pkcs-8/pkcs-8v1_2.asn
// and RFC 5208.
type pkcs8 struct {
	Version    int
	Algo       pkix.AlgorithmIdentifier
	PrivateKey []byte
	// optional attributes omitted.
}

// ParsePKCS8PrivateKey parses an unencrypted, PKCS#8 private key.
// See RFC 5208.
func ParsePKCS8PrivateKey(der []byte) (key interface{}, err error) {
	var privKey pkcs8
	if _, err := asn1.Unmarshal(der, &privKey); err != nil {
		return nil, err
	}
	switch {
	case privKey.Algo.Algorithm.Equal(oidPublicKeyRSA):
		key, err = ParsePKCS1PrivateKey(privKey.PrivateKey)
		if err != nil {
			return nil, errors.New("x509: failed to parse RSA private key embedded in PKCS#8: " + err.Error())
		}
		return key, nil

	case privKey.Algo.Algorithm.Equal(oidPublicKeyECDSA):
		bytes := privKey.Algo.Parameters.FullBytes
		namedCurveOID := new(asn1.ObjectIdentifier)
		if _, err := asn1.Unmarshal(bytes, namedCurveOID); err != nil {
			namedCurveOID = nil
		}
		key, err = parseECPrivateKey(namedCurveOID, privKey.PrivateKey)
		if err != nil {
			return nil, errors.New("x509: failed to parse EC private key embedded in PKCS#8: " + err.Error())
		}
		return key, nil

	default:
		return nil, fmt.Errorf("x509: PKCS#8 wrapping contained private key with unknown algorithm: %v", privKey.Algo.Algorithm)
	}
}

// MarshalPKCS8PrivateKey converts a private key to PKCS#8 encoded form. All
// keys types that are implemented via crypto.Signer are supported (This
// includes *rsa.PublicKey and *ecdsa.PublicKey.). Unknown key types result in
// an error.
//
// See RFC 5208.
func MarshalPKCS8PrivateKey(key interface{}) ([]byte, error) {
	var privKey pkcs8
	switch k := key.(type) {
	case *rsa.PrivateKey:
		// See RFC 3279 2.2.1 RSA Signature Algorithm
		// the parameters component of that type SHALL be the ASN.1 type NULL
		privKey.Algo = pkix.AlgorithmIdentifier{
			Algorithm: oidPublicKeyRSA,
			Parameters: asn1.RawValue{
				Tag: 5, /* NULL */
			},
		}
		privKey.PrivateKey = MarshalPKCS1PrivateKey(k)
	case *ecdsa.PrivateKey:
		oid, ok := oidFromNamedCurve(k.Curve)
		if !ok {
			return nil, errors.New("x509: unknown curve when marshalling PKCS#8")
		}
		par, err := asn1.Marshal(oid)
		if err != nil {
			return nil, errors.New("x509: failed to marshal curve object id in PKCS#8: " + err.Error())
		}
		privKey.Algo = pkix.AlgorithmIdentifier{
			Algorithm: oidPublicKeyECDSA,
			Parameters: asn1.RawValue{
				FullBytes: par,
			},
		}
		ePrivKey, err := MarshalECPrivateKey(k)
		if err != nil {
			return nil, errors.New("x509: failed to marshal EC private key embedded in PKCS#8: " + err.Error())
		}
		privKey.PrivateKey = ePrivKey
	default:
		return nil, fmt.Errorf("x509: unknown key type to marshal to PKCS#8: %T", key)
	}
	return asn1.Marshal(privKey)
}

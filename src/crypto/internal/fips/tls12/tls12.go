// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tls12

import (
	"crypto/internal/fips"
	"crypto/internal/fips/hmac"
	"crypto/internal/fips/sha256"
	"crypto/internal/fips/sha512"
)

// PRF implements the TLS 1.2 pseudo-random function, as defined in RFC 5246,
// Section 5 and allowed by SP 800-135, Revision 1, Section 4.2.2.
func PRF[H fips.Hash](hash func() H, secret []byte, label string, seed []byte, keyLen int) []byte {
	labelAndSeed := make([]byte, len(label)+len(seed))
	copy(labelAndSeed, label)
	copy(labelAndSeed[len(label):], seed)

	result := make([]byte, keyLen)
	pHash(hash, result, secret, labelAndSeed)
	return result
}

// pHash implements the P_hash function, as defined in RFC 5246, Section 5.
func pHash[H fips.Hash](hash func() H, result, secret, seed []byte) {
	h := hmac.New(hash, secret)
	h.Write(seed)
	a := h.Sum(nil)

	for len(result) > 0 {
		h.Reset()
		h.Write(a)
		h.Write(seed)
		b := h.Sum(nil)
		n := copy(result, b)
		result = result[n:]

		h.Reset()
		h.Write(a)
		a = h.Sum(nil)
	}
}

const masterSecretLength = 48
const extendedMasterSecretLabel = "extended master secret"

// MasterSecret implements the TLS 1.2 extended master secret derivation, as
// defined in RFC 7627.
func MasterSecret[H fips.Hash](hash func() H, preMasterSecret, transcript []byte) []byte {
	switch any(hash()).(type) {
	case *sha256.Digest, *sha512.Digest:
		// Approved.
	default:
		fips.RecordNonApproved()
	}
	return PRF(hash, preMasterSecret, extendedMasterSecretLabel, transcript, masterSecretLength)
}

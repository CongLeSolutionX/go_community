// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cipher

import "crypto/internal/fips/gcm"

const (
	gcmStandardTagSize   = 16
	gcmStandardNonceSize = 12
)

// NewGCM returns the given 128-bit, block cipher wrapped in Galois Counter Mode
// with the standard nonce length.
//
// In general, the GHASH operation performed by this implementation of GCM is not constant-time.
// An exception is when the underlying [Block] was created by aes.NewCipher
// on systems with hardware support for AES. See the [crypto/aes] package documentation for details.
func NewGCM(cipher Block) (AEAD, error) {
	// We don't return gcm.New directly, because it would always return a non-nil
	// AEAD interface value with type *gcm.GCM even if the *gcm.GCM is nil.
	g, err := gcm.New(cipher, gcmStandardNonceSize, gcmStandardTagSize)
	if err != nil {
		return nil, err
	}
	return g, nil
}

// NewGCMWithNonceSize returns the given 128-bit, block cipher wrapped in Galois
// Counter Mode, which accepts nonces of the given length. The length must not
// be zero.
//
// Only use this function if you require compatibility with an existing
// cryptosystem that uses non-standard nonce lengths. All other users should use
// [NewGCM], which is faster and more resistant to misuse.
func NewGCMWithNonceSize(cipher Block, size int) (AEAD, error) {
	g, err := gcm.New(cipher, size, gcmStandardTagSize)
	if err != nil {
		return nil, err
	}
	return g, nil
}

// NewGCMWithTagSize returns the given 128-bit, block cipher wrapped in Galois
// Counter Mode, which generates tags with the given length.
//
// Tag sizes between 12 and 16 bytes are allowed.
//
// Only use this function if you require compatibility with an existing
// cryptosystem that uses non-standard tag lengths. All other users should use
// [NewGCM], which is more resistant to misuse.
func NewGCMWithTagSize(cipher Block, tagSize int) (AEAD, error) {
	g, err := gcm.New(cipher, gcmStandardNonceSize, tagSize)
	if err != nil {
		return nil, err
	}
	return g, nil
}

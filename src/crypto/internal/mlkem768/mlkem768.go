// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package mlkem768 implements the quantum-resistant key encapsulation method
// ML-KEM (formerly known as Kyber), as specified in [NIST FIPS 203].
//
// Only the recommended ML-KEM-768 parameter set is provided.
//
// [NIST FIPS 203]: https://doi.org/10.6028/NIST.FIPS.203
package mlkem768

// This package targets security, correctness, simplicity, readability, and
// reviewability as its primary goals. All critical operations are performed in
// constant time.
//
// Variable and function names, as well as code layout, are selected to
// facilitate reviewing the implementation against the NIST FIPS 203 document.
//
// Reviewers unfamiliar with polynomials or linear algebra might find the
// background at https://words.filippo.io/kyber-math/ useful.

import (
	"crypto/rand"
	"crypto/subtle"
	"errors"

	"golang.org/x/crypto/sha3"
)

const (
	// ML-KEM global constants.
	n = 256
	q = 3329

	log2q = 12

	// ML-KEM-768 parameters. The code makes assumptions based on these values,
	// they can't be changed blindly.
	k  = 3
	η  = 2
	du = 10
	dv = 4

	// encodingSizeX is the byte size of a ringElement or nttElement encoded
	// by ByteEncode_X (FIPS 203, Algorithm 5).
	encodingSize12 = n * log2q / 8
	encodingSize10 = n * du / 8
	encodingSize4  = n * dv / 8
	encodingSize1  = n * 1 / 8

	messageSize       = encodingSize1
	decryptionKeySize = k * encodingSize12
	encryptionKeySize = k*encodingSize12 + 32

	CiphertextSize       = k*encodingSize10 + encodingSize4
	EncapsulationKeySize = encryptionKeySize
	SharedKeySize        = 32
	SeedSize             = 32 + 32
)

// A DecapsulationKey is the secret key used to decapsulate a shared key from a
// ciphertext. It includes various precomputed values.
type DecapsulationKey struct {
	d, z [32]byte // decapsulation key seed

	ρ [32]byte // sampleNTT seed for A, stored for the encapsulation key
	h [32]byte // H(ek), stored for ML-KEM.Decaps_internal

	encryptionKey
	decryptionKey
}

// Bytes returns the decapsulation key as a 64-byte seed in the "d || z" form.
func (dk *DecapsulationKey) Bytes() []byte {
	var b [SeedSize]byte
	copy(b[:], dk.d[:])
	copy(b[32:], dk.z[:])
	return b[:]
}

// EncapsulationKey returns the public encapsulation key necessary to produce
// ciphertexts.
func (dk *DecapsulationKey) EncapsulationKey() []byte {
	// The actual logic is in a separate function to outline this allocation.
	b := make([]byte, 0, EncapsulationKeySize)
	return dk.encapsulationKey(b)
}

func (dk *DecapsulationKey) encapsulationKey(b []byte) []byte {
	for i := range dk.t {
		b = polyByteEncode(b, dk.t[i])
	}
	b = append(b, dk.ρ[:]...)
	return b
}

// encryptionKey is the parsed and expanded form of a PKE encryption key.
type encryptionKey struct {
	t [k]nttElement     // ByteDecode₁₂(ek[:384k])
	a [k * k]nttElement // A[i*k+j] = sampleNTT(ρ, j, i)
}

// decryptionKey is the parsed and expanded form of a PKE decryption key.
type decryptionKey struct {
	s [k]nttElement // ByteDecode₁₂(dk[:decryptionKeySize])
}

// GenerateKey generates a new decapsulation key, drawing random bytes from
// crypto/rand. The decapsulation key must be kept secret.
func GenerateKey() (*DecapsulationKey, error) {
	// The actual logic is in a separate function to outline this allocation.
	dk := &DecapsulationKey{}
	return generateKey(dk)
}

func generateKey(dk *DecapsulationKey) (*DecapsulationKey, error) {
	var d [32]byte
	if _, err := rand.Read(d[:]); err != nil {
		return nil, errors.New("mlkem768: crypto/rand Read failed: " + err.Error())
	}
	var z [32]byte
	if _, err := rand.Read(z[:]); err != nil {
		return nil, errors.New("mlkem768: crypto/rand Read failed: " + err.Error())
	}
	return kemKeyGen(dk, &d, &z), nil
}

// NewKeyFromSeed deterministically generates a decapsulation key from a 64-byte
// seed in the "d || z" form. The seed must be uniformly random.
func NewKeyFromSeed(seed []byte) (*DecapsulationKey, error) {
	// The actual logic is in a separate function to outline this allocation.
	dk := &DecapsulationKey{}
	return newKeyFromSeed(dk, seed)
}

func newKeyFromSeed(dk *DecapsulationKey, seed []byte) (*DecapsulationKey, error) {
	if len(seed) != SeedSize {
		return nil, errors.New("mlkem768: invalid seed length")
	}
	d := (*[32]byte)(seed[:32])
	z := (*[32]byte)(seed[32:])
	return kemKeyGen(dk, d, z), nil
}

// kemKeyGen generates a decapsulation key.
//
// It implements ML-KEM.KeyGen_internal according to FIPS 203, Algorithm 16, and
// K-PKE.KeyGen according to FIPS 203, Algorithm 13. The two are merged to save
// copies and allocations.
func kemKeyGen(dk *DecapsulationKey, d, z *[32]byte) *DecapsulationKey {
	if dk == nil {
		dk = &DecapsulationKey{}
	}
	dk.d = *d
	dk.z = *z

	g := sha3.New512()
	g.Write(d[:])
	g.Write([]byte{k}) // Module dimension as a domain separator.
	G := g.Sum(make([]byte, 0, 64))
	ρ, σ := G[:32], G[32:]
	dk.ρ = [32]byte(ρ)

	A := &dk.a
	for i := byte(0); i < k; i++ {
		for j := byte(0); j < k; j++ {
			A[i*k+j] = sampleNTT(ρ, j, i)
		}
	}

	var N byte
	s := &dk.s
	for i := range s {
		s[i] = ntt(samplePolyCBD(σ, N))
		N++
	}
	e := make([]nttElement, k)
	for i := range e {
		e[i] = ntt(samplePolyCBD(σ, N))
		N++
	}

	t := &dk.t
	for i := range t { // t = A ◦ s + e
		t[i] = e[i]
		for j := range s {
			t[i] = polyAdd(t[i], nttMul(A[i*k+j], s[j]))
		}
	}

	H := sha3.New256()
	ek := dk.EncapsulationKey()
	H.Write(ek)
	H.Sum(dk.h[:0])

	return dk
}

// Encapsulate generates a shared key and an associated ciphertext from an
// encapsulation key, drawing random bytes from crypto/rand.
// If the encapsulation key is not valid, Encapsulate returns an error.
//
// The shared key must be kept secret.
func Encapsulate(encapsulationKey []byte) (ciphertext, sharedKey []byte, err error) {
	// The actual logic is in a separate function to outline this allocation.
	var cc [CiphertextSize]byte
	return encapsulate(&cc, encapsulationKey)
}

func encapsulate(cc *[CiphertextSize]byte, encapsulationKey []byte) (ciphertext, sharedKey []byte, err error) {
	if len(encapsulationKey) != EncapsulationKeySize {
		return nil, nil, errors.New("mlkem768: invalid encapsulation key length")
	}
	var m [messageSize]byte
	if _, err := rand.Read(m[:]); err != nil {
		return nil, nil, errors.New("mlkem768: crypto/rand Read failed: " + err.Error())
	}
	// Note that the modulus check (step 2 of the encapsulation key check from
	// FIPS 203, Section 7.2) is performed by polyByteDecode in parseEK.
	return kemEncaps(cc, encapsulationKey, &m)
}

// kemEncaps generates a shared key and an associated ciphertext.
//
// It implements ML-KEM.Encaps_internal according to FIPS 203, Algorithm 17.
func kemEncaps(cc *[CiphertextSize]byte, ek []byte, m *[messageSize]byte) (c, K []byte, err error) {
	if cc == nil {
		cc = &[CiphertextSize]byte{}
	}

	H := sha3.Sum256(ek[:])
	g := sha3.New512()
	g.Write(m[:])
	g.Write(H[:])
	G := g.Sum(nil)
	K, r := G[:SharedKeySize], G[SharedKeySize:]
	var ex encryptionKey
	if err := parseEK(&ex, ek[:]); err != nil {
		return nil, nil, err
	}
	c = pkeEncrypt(cc, &ex, m, r)
	return c, K, nil
}

// parseEK parses an encryption key from its encoded form.
//
// It implements the initial stages of K-PKE.Encrypt according to FIPS 203,
// Algorithm 14.
func parseEK(ex *encryptionKey, ekPKE []byte) error {
	if len(ekPKE) != encryptionKeySize {
		return errors.New("mlkem768: invalid encryption key length")
	}

	for i := range ex.t {
		var err error
		ex.t[i], err = polyByteDecode[nttElement](ekPKE[:encodingSize12])
		if err != nil {
			return err
		}
		ekPKE = ekPKE[encodingSize12:]
	}
	ρ := ekPKE

	for i := byte(0); i < k; i++ {
		for j := byte(0); j < k; j++ {
			ex.a[i*k+j] = sampleNTT(ρ, j, i)
		}
	}

	return nil
}

// pkeEncrypt encrypt a plaintext message.
//
// It implements K-PKE.Encrypt according to FIPS 203, Algorithm 14, although the
// computation of t and AT is done in parseEK.
func pkeEncrypt(cc *[CiphertextSize]byte, ex *encryptionKey, m *[messageSize]byte, rnd []byte) []byte {
	var N byte
	r, e1 := make([]nttElement, k), make([]ringElement, k)
	for i := range r {
		r[i] = ntt(samplePolyCBD(rnd, N))
		N++
	}
	for i := range e1 {
		e1[i] = samplePolyCBD(rnd, N)
		N++
	}
	e2 := samplePolyCBD(rnd, N)

	u := make([]ringElement, k) // NTT⁻¹(AT ◦ r) + e1
	for i := range u {
		u[i] = e1[i]
		for j := range r {
			// Note that i and j are inverted, as we need the transposed of A.
			u[i] = polyAdd(u[i], inverseNTT(nttMul(ex.a[j*k+i], r[j])))
		}
	}

	μ := ringDecodeAndDecompress1(m)

	var vNTT nttElement // t⊺ ◦ r
	for i := range ex.t {
		vNTT = polyAdd(vNTT, nttMul(ex.t[i], r[i]))
	}
	v := polyAdd(polyAdd(inverseNTT(vNTT), e2), μ)

	c := cc[:0]
	for _, f := range u {
		c = ringCompressAndEncode10(c, f)
	}
	c = ringCompressAndEncode4(c, v)

	return c
}

// Decapsulate generates a shared key from a ciphertext and a decapsulation key.
// If the ciphertext is not valid, Decapsulate returns an error.
//
// The shared key must be kept secret.
func (dk *DecapsulationKey) Decapsulate(ciphertext []byte) (sharedKey []byte, err error) {
	if len(ciphertext) != CiphertextSize {
		return nil, errors.New("mlkem768: invalid ciphertext length")
	}
	c := (*[CiphertextSize]byte)(ciphertext)
	// Note that the hash check (step 3 of the decapsulation input check from
	// FIPS 203, Section 7.3) is foregone as a DecapsulationKey is always
	// validly generated by ML-KEM.KeyGen_internal.
	return kemDecaps(dk, c), nil
}

// kemDecaps produces a shared key from a ciphertext.
//
// It implements ML-KEM.Decaps_internal according to FIPS 203, Algorithm 18.
func kemDecaps(dk *DecapsulationKey, c *[CiphertextSize]byte) (K []byte) {
	m := pkeDecrypt(&dk.decryptionKey, c)
	g := sha3.New512()
	g.Write(m[:])
	g.Write(dk.h[:])
	G := g.Sum(make([]byte, 0, 64))
	Kprime, r := G[:SharedKeySize], G[SharedKeySize:]
	J := sha3.NewShake256()
	J.Write(dk.z[:])
	J.Write(c[:])
	Kout := make([]byte, SharedKeySize)
	J.Read(Kout)
	var cc [CiphertextSize]byte
	c1 := pkeEncrypt(&cc, &dk.encryptionKey, (*[32]byte)(m), r)

	subtle.ConstantTimeCopy(subtle.ConstantTimeCompare(c[:], c1), Kout, Kprime)
	return Kout
}

// pkeDecrypt decrypts a ciphertext.
//
// It implements K-PKE.Decrypt according to FIPS 203, Algorithm 15,
// although s is retained from kemKeyGen.
func pkeDecrypt(dx *decryptionKey, c *[CiphertextSize]byte) []byte {
	u := make([]ringElement, k)
	for i := range u {
		b := (*[encodingSize10]byte)(c[encodingSize10*i : encodingSize10*(i+1)])
		u[i] = ringDecodeAndDecompress10(b)
	}

	b := (*[encodingSize4]byte)(c[encodingSize10*k:])
	v := ringDecodeAndDecompress4(b)

	var mask nttElement // s⊺ ◦ NTT(u)
	for i := range dx.s {
		mask = polyAdd(mask, nttMul(dx.s[i], ntt(u[i])))
	}
	w := polySub(v, inverseNTT(mask))

	return ringCompressAndEncode1(nil, w)
}

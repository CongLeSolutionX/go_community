// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rand_test

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"math/big"
)

// This example reads 10 cryptographically secure random numbers from
// rand.Reader and writes them to a byte slice.
func Example() {
	c := 10
	b := make([]byte, c)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	// The slice should now contain random bytes instead of only zeroes.
	fmt.Println(bytes.Equal(b, make([]byte, c)))

	// Output:
	// false
}

// This example demonstrates how to use rand.Read to implement
// a One-Time Pad cipher (https://wikipedia.org/wiki/One-time_pad).
func ExampleRead_oneTimePad() {
	plaintext := []byte("A secret message.")
	n := len(plaintext)
	key := make([]byte, n)
	ciphertext := make([]byte, n)

	// Read n random bytes into key. A one-time pad requires the key
	// to be at least as long as the plaintext.
	if _, err := rand.Read(key); err != nil {
		panic(err)
	}
	fmt.Printf("Original: %x\n", plaintext)

	// Encrypt message by XOR'ing the plaintext and the key.
	for i := 0; i < n; i++ {
		ciphertext[i] = plaintext[i] ^ key[i]
	}
	fmt.Printf("Encrypted: %x\n", ciphertext)

	// Decrypt message by XOR'ing the ciphertext and the key.
	for i := 0; i < n; i++ {
		ciphertext[i] = ciphertext[i] ^ key[i]
	}
	fmt.Printf("Decrypted: %x\n", ciphertext)
}

// This example demonstrates how to use rand.Prime to implement plain-old
// RSA key generation (https://wikipedia.org/wiki/RSA_(cryptosystem)#Key_generation).
// Obviously don't use this example code for anything real.
func ExamplePrime_rSA() {
	// Choose two random primes p and q
	bits := 100
	p, err := rand.Prime(rand.Reader, bits)
	if err != nil {
		panic(err)
	}
	q, err := rand.Prime(rand.Reader, bits)
	if err != nil {
		panic(err)
	}

	// Let n = p * q
	var n big.Int
	n.Mul(p, q)

	// Let totient = (p - 1) * (q - 1)
	one := big.NewInt(1)
	p.Sub(p, one)
	q.Sub(q, one)
	var totient big.Int
	totient.Mul(p, q)

	// Let e be coprime to the totient
	e := big.NewInt(3)

	// Let d satisfy d*e ≡ 1 mod totient
	var d big.Int
	d.ModInverse(e, &totient)

	fmt.Printf("Public Key-Pair: (%s, %s)\n", n.Text(10), e.Text(10))
	fmt.Printf("Private Key-Pair: (%s, %s)\n", n.Text(10), d.Text(10))
}

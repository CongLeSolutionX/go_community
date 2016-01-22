// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package aes_test

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	"log"
)

func ExampleDecrypt(cipherstring string, keystring string) string {
	// Byte array of the string
	ciphertext := []byte(cipherstring)

	// Key
	key := []byte(keystring)

	// Create the AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Before even testing the decryption,
	// if the text is too small, then it is incorrect
	if len(ciphertext) < aes.BlockSize {
		log.Fatal("Text is shorter than minium AES block.")
	}

	// Get the 16 byte IV
	iv := ciphertext[:aes.BlockSize]

	// Remove the IV from the ciphertext
	ciphertext = ciphertext[aes.BlockSize:]

	// Return a decrypted stream
	stream := cipher.NewCFBDecrypter(block, iv)

	// Decrypt bytes from ciphertext
	stream.XORKeyStream(ciphertext, ciphertext)

	return string(ciphertext)
}

func ExampleEncrypt(plainstring, keystring string) string {
	// Byte array of the string
	plaintext := []byte(plainstring)

	// Key
	key := []byte(keystring)

	// Create the AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Empty array of 16 + plaintext length
	// Include the IV at the beginning
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))

	// Slice of first 16 bytes
	iv := ciphertext[:aes.BlockSize]

	// Write 16 rand bytes to fill iv
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		log.Fatal(err.Error())
	}

	// Return an encrypted stream
	stream := cipher.NewCFBEncrypter(block, iv)

	// Encrypt bytes from plaintext to ciphertext
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return string(ciphertext)
}

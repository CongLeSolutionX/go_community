// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package aes

import (
	"crypto/cipher"
	"internal/cpu"
)

// define in asm_arm64.s
func encryptBlockAsm(nr int, xk *uint32, dst, src *byte)
func decryptBlockAsm(nr int, xk *uint32, dst, src *byte)

type aesCipherAsm struct {
	aesCipher
}

var hasAES = cpu.ARM64.HasAES

func newCipher(key []byte) (cipher.Block, error) {
	if !hasAES {
		return newCipherGeneric(key)
	}
	n := len(key) + 28
	c := aesCipherAsm{aesCipher{make([]uint32, n), make([]uint32, n)}}
	armExpandKey(key, c.enc, c.dec)
	return &c, nil
}

func (c *aesCipherAsm) BlockSize() int { return BlockSize }

func (c *aesCipherAsm) Encrypt(dst, src []byte) {
	if len(src) < BlockSize {
		panic("crypto/aes: input not full block")
	}
	if len(dst) < BlockSize {
		panic("crypto/aes: output not full block")
	}
	encryptBlockAsm(len(c.enc)/4-1, &c.enc[0], &dst[0], &src[0])
}

func (c *aesCipherAsm) Decrypt(dst, src []byte) {
        if len(src) < BlockSize {
                panic("crypto/aes: input not full block")
        }
        if len(dst) < BlockSize {
                panic("crypto/aes: output not full block")
        }
        decryptBlockAsm(len(c.dec)/4-1, &c.dec[0], &dst[0], &src[0])
}

func armExpandKey(key []byte, enc, dec []uint32) {
	var i int

	expandKeyGo(key, enc, dec)
	nk := len(enc)
	for ; i < nk; i++ {
		enc[i] = uint32(byte(enc[i] >> 24)) | uint32(byte(enc[i] >> 16)) << 8 | uint32(byte(enc[i] >> 8)) << 16 | uint32(byte(enc[i]))  << 24
		dec[i] = uint32(byte(dec[i] >> 24)) | uint32(byte(dec[i] >> 16)) << 8 | uint32(byte(dec[i] >> 8)) << 16 | uint32(byte(dec[i]))  << 24
	}
}


// expandKey is used by BenchmarkExpand to ensure that the asm implementation
// of key expansion is used for the benchmark when it is available.
func expandKey(key []byte, enc, dec []uint32) {
    expandKeyGo(key, enc, dec)
}

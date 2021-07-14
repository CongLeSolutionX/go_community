// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build amd64 || arm64
// +build amd64 arm64

package aes

import (
	"crypto/cipher"
	subtleoverlap "crypto/internal/subtle"
	"crypto/subtle"
	"errors"
	"golang.org/x/sys/cpu"
	"unsafe"
)

// aes_gcm.go contains aes-gcm encryption/decryption functions
// that export two asm-optimized implementations:
// 1. scalar aes-ni
// 2. vector vaes
// NewGCM() detects cpu instruction set support
// and selects the best version for the host cpu

// The following functions are defined in gcm_*.s.
// for the aes-ni scalar implementation

//go:noescape
func gcmAesInit(productTable *[256]byte, ks []uint32)

//go:noescape
func gcmAesData(productTable *[256]byte, data []byte, T *[16]byte)

//go:noescape
func gcmAesEnc(productTable *[256]byte, dst, src []byte, ctr, T *[16]byte, ks []uint32)

//go:noescape
func gcmAesDec(productTable *[256]byte, dst, src []byte, ctr, T *[16]byte, ks []uint32)

//go:noescape
func gcmAesFinish(productTable *[256]byte, tagMask, T *[16]byte, pLen, dLen uint64)

// The following functions are defined in gcmv_amd64.s.
// for the vaes vector implementation

// key expansion and GHASH table pre-computation, 128/192/256-bit
//go:noescape
func aesKeyExp128_avx2(key, keyData []byte)
func aesKeyExp192_avx2(key, keyData []byte)
func aesKeyExp256_avx2(key, keyData []byte)
func gcmAesGhashPrecomp128_vaes(keyData []byte)
func gcmAesGhashPrecomp192_vaes(keyData []byte)
func gcmAesGhashPrecomp256_vaes(keyData []byte)

// encrypt/decrypt single call api, 128/192/256-bit
//go:noescape
func gcmAesEnc128_vaes(keyData []byte, ctx *gcmContext, out, in, iv, aad, authTag []byte)
func gcmAesEnc192_vaes(keyData []byte, ctx *gcmContext, out, in, iv, aad, authTag []byte)
func gcmAesEnc256_vaes(keyData []byte, ctx *gcmContext, out, in, iv, aad, authTag []byte)
func gcmAesDec128_vaes(keyData []byte, ctx *gcmContext, out, in, iv, aad, authTag []byte)
func gcmAesDec192_vaes(keyData []byte, ctx *gcmContext, out, in, iv, aad, authTag []byte)
func gcmAesDec256_vaes(keyData []byte, ctx *gcmContext, out, in, iv, aad, authTag []byte)

// encrypt/decrypt multi call api, 128/192/256-bit
//go:noescape
func gcmAesInitVarIv_vaes(keyData []byte, ctx *gcmContext, iv, aad []byte)
func gcmAesEncUpdate128_vaes(keyData []byte, ctx *gcmContext, out, in []byte)
func gcmAesEncUpdate192_vaes(keyData []byte, ctx *gcmContext, out, in []byte)
func gcmAesEncUpdate256_vaes(keyData []byte, ctx *gcmContext, out, in []byte)
func gcmAesDecUpdate128_vaes(keyData []byte, ctx *gcmContext, out, in []byte)
func gcmAesDecUpdate192_vaes(keyData []byte, ctx *gcmContext, out, in []byte)
func gcmAesDecUpdate256_vaes(keyData []byte, ctx *gcmContext, out, in []byte)
func gcmAesFinish128_vaes(keyData []byte, ctx *gcmContext, authTag []byte)
func gcmAesFinish192_vaes(keyData []byte, ctx *gcmContext, authTag []byte)
func gcmAesFinish256_vaes(keyData []byte, ctx *gcmContext, authTag []byte)

const (
	align                = 64 // buffer alignment, in bytes, requried for AVX-512 buffers
	gcmBlockSize         = 16
	gcmTagSize           = 16
	gcmMinimumTagSize    = 12 // NIST SP 800-38D recommends tags with 12 or more bytes.
	gcmStandardNonceSize = 12
)

// vaes AVX-512 constants for which 64-byte alignment is required
var shufMaskStr = "0F0E0D0C0B0A09080706050403020100" +
	"0F0E0D0C0B0A09080706050403020100" +
	"0F0E0D0C0B0A09080706050403020100" +
	"0F0E0D0C0B0A09080706050403020100"

var ddqAddBE4444Str = "00000000000000000000000000000004" +
	"00000000000000000000000000000004" +
	"00000000000000000000000000000004" +
	"00000000000000000000000000000004"

var ddqAddBE1234Str = "00000000000000000000000000000001" +
	"00000000000000000000000000000002" +
	"00000000000000000000000000000003" +
	"00000000000000000000000000000004"

var byte64LenToMaskTableStr = "0000000000000000" + "0100000000000000" +
	"0300000000000000" + "0700000000000000" +
	"0f00000000000000" + "1f00000000000000" +
	"3f00000000000000" + "7f00000000000000" +
	"ff00000000000000" + "ff01000000000000" +
	"ff03000000000000" + "ff07000000000000" +
	"ff0f000000000000" + "ff1f000000000000" +
	"ff3f000000000000" + "ff7f000000000000" +
	"ffff000000000000" + "ffff010000000000" +
	"ffff030000000000" + "ffff070000000000" +
	"ffff0f0000000000" + "ffff1f0000000000" +
	"ffff3f0000000000" + "ffff7f0000000000" +
	"ffffff0000000000" + "ffffff0100000000" +
	"ffffff0300000000" + "ffffff0700000000" +
	"ffffff0f00000000" + "ffffff1f00000000" +
	"ffffff3f00000000" + "ffffff7f00000000" +
	"ffffffff00000000" + "ffffffff01000000" +
	"ffffffff03000000" + "ffffffff07000000" +
	"ffffffff0f000000" + "ffffffff1f000000" +
	"ffffffff3f000000" + "ffffffff7f000000" +
	"ffffffffff000000" + "ffffffffff010000" +
	"ffffffffff030000" + "ffffffffff070000" +
	"ffffffffff0f0000" + "ffffffffff1f0000" +
	"ffffffffff3f0000" + "ffffffffff7f0000" +
	"ffffffffffff0000" + "ffffffffffff0100" +
	"ffffffffffff0300" + "ffffffffffff0700" +
	"ffffffffffff0f00" + "ffffffffffff1f00" +
	"ffffffffffff3f00" + "ffffffffffff7f00" +
	"ffffffffffffff00" + "ffffffffffffff01" +
	"ffffffffffffff03" + "ffffffffffffff07" +
	"ffffffffffffff0f" + "ffffffffffffff1f" +
	"ffffffffffffff3f" + "ffffffffffffff7f" +
	"ffffffffffffffff"

// vaes encrypt/decrypt function pointer for either single or multi-call
type gcmEncDecFunc func([]byte, *gcmContext, []byte, []byte, []byte, []byte, []byte)

type gcmAsmVaes struct {
	// keyData contains the key schedule and binary-field product table
	keyData []byte
	// ctx contains cipher state variables and 64-byte aligned constants for AVX-512 SIMD
	ctx gcmContext
	// nonceSize contains the expected size of the nonce, in bytes.
	nonceSize int
	// tag contains the authenication tag
	// per NIST 800-38D: 128, 120, 112, 104, or 96, 64, or 32 bits
	tag []byte
	// tagSize contains the size of the tag, in bytes.
	tagSize int
	// encrypt/decrypt point to cipher implementations that match key and nonce lengths
	encrypt gcmEncDecFunc
	decrypt gcmEncDecFunc
}

type gcmContext struct {
	aad_hash              [gcmBlockSize]byte
	aad_length            uint64
	in_length             uint64
	partial_block_enc_key [gcmBlockSize]byte
	orig_IV               [gcmBlockSize]byte
	current_counter       [gcmBlockSize]byte
	partial_block_length  uint64
	// below are fields needed to compensate for lack of
	// 64-byte alignment primitives in Go asm db tables;
	// each constant is brute-force algined to 64-bytes
	// at the Go level to support the underlying asm function
	ddqAddBE4444   *byte
	ddqAddBE1234   *byte
	shuffleMask    *byte
	byte64Len2Mask *byte
}

// vaes encrypt, multi-call (variable-length iv), 128-bit key
func gcmAesEncVarIv128_vaes(keyData []byte, ctx *gcmContext, out, plaintext, nonce, data []byte, tag []byte) {
	gcmAesInitVarIv_vaes(keyData, ctx, nonce, data)
	gcmAesEncUpdate128_vaes(keyData, ctx, out, plaintext)
	gcmAesFinish128_vaes(keyData, ctx, tag)
}

// vaes encrypt, multi-call (variable-length iv), 192-bit key
func gcmAesEncVarIv192_vaes(keyData []byte, ctx *gcmContext, out, plaintext, nonce, data []byte, tag []byte) {
	gcmAesInitVarIv_vaes(keyData, ctx, nonce, data)
	gcmAesEncUpdate192_vaes(keyData, ctx, out, plaintext)
	gcmAesFinish192_vaes(keyData, ctx, tag)
}

// vaes encrypt, multi-call (variable-length iv), 256-bit key
func gcmAesEncVarIv256_vaes(keyData []byte, ctx *gcmContext, out, plaintext, nonce, data []byte, tag []byte) {
	gcmAesInitVarIv_vaes(keyData, ctx, nonce, data)
	gcmAesEncUpdate256_vaes(keyData, ctx, out, plaintext)
	gcmAesFinish256_vaes(keyData, ctx, tag)
}

// vaes decrypt, multi-call (variable-length iv), 128-bit key
func gcmAesDecVarIv128_vaes(keyData []byte, ctx *gcmContext, out, plaintext, nonce, data []byte, tag []byte) {
	gcmAesInitVarIv_vaes(keyData, ctx, nonce, data)
	gcmAesDecUpdate128_vaes(keyData, ctx, out, plaintext)
	gcmAesFinish128_vaes(keyData, ctx, tag)
}

// vaes decrypt, multi-call (variable-length iv), 192-bit key
func gcmAesDecVarIv192_vaes(keyData []byte, ctx *gcmContext, out, plaintext, nonce, data []byte, tag []byte) {
	gcmAesInitVarIv_vaes(keyData, ctx, nonce, data)
	gcmAesDecUpdate192_vaes(keyData, ctx, out, plaintext)
	gcmAesFinish192_vaes(keyData, ctx, tag)
}

// vaes decrypt, multi-call (variable-length iv), 256-bit key
func gcmAesDecVarIv256_vaes(keyData []byte, ctx *gcmContext, out, plaintext, nonce, data []byte, tag []byte) {
	gcmAesInitVarIv_vaes(keyData, ctx, nonce, data)
	gcmAesDecUpdate256_vaes(keyData, ctx, out, plaintext)
	gcmAesFinish256_vaes(keyData, ctx, tag)
}

// vaes key expansion and GHASH table init, 128-bit
func gcmAesInit128_vaes(key []byte, key_data []byte) {
	aesKeyExp128_avx2(key, key_data)
	gcmAesGhashPrecomp128_vaes(key_data)
}

// vaes key expansion and GHASH table init, 192-bit
func gcmAesInit192_vaes(key []byte, key_data []byte) {
	aesKeyExp192_avx2(key, key_data)
	gcmAesGhashPrecomp192_vaes(key_data)
}

// vaes key expansion and GHASH table init, 256-bit
func gcmAesInit256_vaes(key []byte, key_data []byte) {
	aesKeyExp256_avx2(key, key_data)
	gcmAesGhashPrecomp256_vaes(key_data)
}

// get buffer pointer offset to force alignment on align boundary
func getAlignmentOffset(ptr uintptr, align uint) uint {
	alignment := uintptr(align)
	unaligned := (ptr & (alignment - uintptr(1))) != 0
	var offset uintptr
	if unaligned {
		offset = alignment - ptr%uintptr(alignment)
	}
	return uint(offset)
}

// initialize an aligned buffer with string contents;
// used for avx-512 gcm constants
func initVectAlign(srcData string, dstVect **byte) []byte {
	bufLen := len(srcData) / 2
	buf := make([]byte, bufLen+align)
	ptr := uintptr(unsafe.Pointer(&buf[0]))
	offset := getAlignmentOffset(ptr, align)
	if nil != dstVect {
		*dstVect = &(buf[offset])
	}
	dataBytes := decodeString(srcData)
	for i := 0; i < bufLen; i++ {
		buf[int(offset)+i] = dataBytes[i]
	}
	return (buf[offset:])
}

// allocate an aligned slice
func allocVectAlign(numBytes int64) []byte {
	buf := make([]byte, numBytes+align)
	ptr := uintptr(unsafe.Pointer(&buf[0]))
	offset := getAlignmentOffset(ptr, align)
	return (buf[offset:])
}

// vaes gcmInit
func gcmInit(key []byte, nonceSize, tagSize int) (keyData, tag []byte, ctx gcmContext, encryptFunc, decryptFunc gcmEncDecFunc) {

	// alloc key expansion buffer
	keyData = allocVectAlign(1024)

	// alloc authentication tag buffer
	tag = make([]byte, tagSize)

	// init context with gcm constants
	initVectAlign(shufMaskStr, &(ctx.shuffleMask))
	initVectAlign(byte64LenToMaskTableStr, &(ctx.byte64Len2Mask))
	initVectAlign(ddqAddBE4444Str, &(ctx.ddqAddBE4444))
	initVectAlign(ddqAddBE1234Str, &(ctx.ddqAddBE1234))

	// compute key expansions and init gcm function pointers given iv and key lengths
	switch len(key) {
	case 16:
		gcmAesInit128_vaes(key, keyData)
		if gcmStandardNonceSize == nonceSize {
			encryptFunc = gcmAesEnc128_vaes
			decryptFunc = gcmAesDec128_vaes
		} else {
			encryptFunc = gcmAesEncVarIv128_vaes
			decryptFunc = gcmAesDecVarIv128_vaes
		}
	case 24:
		gcmAesInit192_vaes(key, keyData)
		if gcmStandardNonceSize == nonceSize {
			encryptFunc = gcmAesEnc192_vaes
			decryptFunc = gcmAesDec192_vaes
		} else {
			encryptFunc = gcmAesEncVarIv192_vaes
			decryptFunc = gcmAesDecVarIv192_vaes
		}
	case 32:
		gcmAesInit256_vaes(key, keyData)
		if gcmStandardNonceSize == nonceSize {
			encryptFunc = gcmAesEnc256_vaes
			decryptFunc = gcmAesDec256_vaes
		} else {
			encryptFunc = gcmAesEncVarIv256_vaes
			decryptFunc = gcmAesDecVarIv256_vaes
		}
	}
	return keyData, tag, ctx, encryptFunc, decryptFunc
}

var errOpen = errors.New("cipher: message authentication failed")

// aesCipherGCM implements crypto/cipher.gcmAble so that crypto/cipher.NewGCM
// will use the optimised implementation in this file when possible. Instances
// of this type only exist when hasGCMAsm returns true.
type aesCipherGCM struct {
	aesCipherAsm
}

// Assert that aesCipherGCM implements the gcmAble interface.
var _ gcmAble = (*aesCipherGCM)(nil)

var supportsAVX512VAES = cpu.X86.HasAVX512 && cpu.X86.HasAVX512VAES && cpu.X86.HasAVX512VPCLMULQDQ

// NewGCM returns the AES cipher wrapped in Galois Counter Mode. This is only
// called by crypto/cipher.NewGCM via the gcmAble interface.
func (c *aesCipherGCM) NewGCM(nonceSize, tagSize int) (cipher.AEAD, error) {
	if supportsAVX512VAES {
		g := &gcmAsmVaes{nonceSize: nonceSize, tagSize: tagSize}
		g.keyData, g.tag, g.ctx, g.encrypt, g.decrypt = gcmInit(c.key, nonceSize, tagSize)
		return g, nil
	} else {
		g := &gcmAsm{ks: c.enc, nonceSize: nonceSize, tagSize: tagSize}
		gcmAesInit(&g.productTable, g.ks)
		return g, nil
	}
}

// vaes implementation
func (g *gcmAsmVaes) NonceSize() int {
	return g.nonceSize
}

func (g *gcmAsmVaes) Overhead() int {
	return g.tagSize
}

// sliceForAppend takes a slice and a requested number of bytes. It returns a
// slice with the contents of the given slice followed by that many bytes and a
// second slice that aliases into it and contains only the extra bytes. If the
// original slice has sufficient capacity then no allocation is performed.
func sliceForAppend(in []byte, n int) (head, tail []byte) {
	if total := len(in) + n; cap(in) >= total {
		head = in[:total]
	} else {
		head = make([]byte, total)
		copy(head, in)
	}
	tail = head[len(in):]
	return
}

// Seal encrypts and authenticates plaintext. See the cipher.AEAD interface for details.
func (g *gcmAsmVaes) Seal(dst, nonce, plaintext, data []byte) []byte {
	if len(nonce) != g.nonceSize {
		panic("crypto/cipher: incorrect nonce length given to GCM")
	}
	if uint64(len(plaintext)) > ((1<<32)-2)*BlockSize {
		panic("crypto/cipher: message too large for GCM")
	}
	ret, out := sliceForAppend(dst, len(plaintext)+g.tagSize)
	g.encrypt(g.keyData, &(g.ctx), out, plaintext, nonce, data, g.tag)
	if subtleoverlap.InexactOverlap(out[:len(plaintext)], plaintext) {
		panic("crypto/cipher: invalid buffer overlap")
	}
	copy(out[len(plaintext):], g.tag[:])
	return ret
}

// Open authenticates and decrypts ciphertext. See the cipher.AEAD interface for details.
func (g *gcmAsmVaes) Open(dst, nonce, ciphertext, data []byte) ([]byte, error) {
	if len(nonce) != g.nonceSize {
		panic("crypto/cipher: incorrect nonce length given to GCM")
	}
	// Sanity check to prevent the authentication from always succeeding if an implementation
	// leaves tagSize uninitialized, for example.
	if g.tagSize < gcmMinimumTagSize {
		panic("crypto/cipher: incorrect GCM tag size")
	}
	if len(ciphertext) < g.tagSize {
		return nil, errOpen
	}
	if uint64(len(ciphertext)) > ((1<<32)-2)*uint64(BlockSize)+uint64(g.tagSize) {
		return nil, errOpen
	}
	tag := ciphertext[len(ciphertext)-g.tagSize:]
	ciphertext = ciphertext[:len(ciphertext)-g.tagSize]
	ret, out := sliceForAppend(dst, len(ciphertext))
	if subtleoverlap.InexactOverlap(out, ciphertext) {
		panic("crypto/cipher: invalid buffer overlap")
	}
	g.decrypt(g.keyData, &(g.ctx), out, ciphertext, nonce, data, g.tag)
	if subtle.ConstantTimeCompare(g.tag[:g.tagSize], tag) != 1 {
		for i := range out {
			out[i] = 0
		}
		return nil, errOpen
	}
	return ret, nil
}

// aes-ni implementation
type gcmAsm struct {
	// ks is the key schedule, the length of which depends on the size of
	// the AES key.
	ks []uint32
	// productTable contains pre-computed multiples of the binary-field
	// element used in GHASH.
	productTable [256]byte
	// nonceSize contains the expected size of the nonce, in bytes.
	nonceSize int
	// tagSize contains the size of the tag, in bytes.
	tagSize int
}

func (g *gcmAsm) NonceSize() int {
	return g.nonceSize
}

func (g *gcmAsm) Overhead() int {
	return g.tagSize
}

// Seal encrypts and authenticates plaintext. See the cipher.AEAD interface for
// details.
func (g *gcmAsm) Seal(dst, nonce, plaintext, data []byte) []byte {
	if len(nonce) != g.nonceSize {
		panic("crypto/cipher: incorrect nonce length given to GCM")
	}
	if uint64(len(plaintext)) > ((1<<32)-2)*BlockSize {
		panic("crypto/cipher: message too large for GCM")
	}

	var counter, tagMask [gcmBlockSize]byte

	if len(nonce) == gcmStandardNonceSize {
		// Init counter to nonce||1
		copy(counter[:], nonce)
		counter[gcmBlockSize-1] = 1
	} else {
		// Otherwise counter = GHASH(nonce)
		gcmAesData(&g.productTable, nonce, &counter)
		gcmAesFinish(&g.productTable, &tagMask, &counter, uint64(len(nonce)), uint64(0))
	}

	encryptBlockAsm(len(g.ks)/4-1, &g.ks[0], &tagMask[0], &counter[0])

	var tagOut [gcmTagSize]byte
	gcmAesData(&g.productTable, data, &tagOut)

	ret, out := sliceForAppend(dst, len(plaintext)+g.tagSize)
	if subtleoverlap.InexactOverlap(out[:len(plaintext)], plaintext) {
		panic("crypto/cipher: invalid buffer overlap")
	}
	if len(plaintext) > 0 {
		gcmAesEnc(&g.productTable, out, plaintext, &counter, &tagOut, g.ks)
	}
	gcmAesFinish(&g.productTable, &tagMask, &tagOut, uint64(len(plaintext)), uint64(len(data)))
	copy(out[len(plaintext):], tagOut[:])

	return ret
}

// Open authenticates and decrypts ciphertext. See the cipher.AEAD interface
// for details.
func (g *gcmAsm) Open(dst, nonce, ciphertext, data []byte) ([]byte, error) {
	if len(nonce) != g.nonceSize {
		panic("crypto/cipher: incorrect nonce length given to GCM")
	}
	// Sanity check to prevent the authentication from always succeeding if an implementation
	// leaves tagSize uninitialized, for example.
	if g.tagSize < gcmMinimumTagSize {
		panic("crypto/cipher: incorrect GCM tag size")
	}

	if len(ciphertext) < g.tagSize {
		return nil, errOpen
	}
	if uint64(len(ciphertext)) > ((1<<32)-2)*uint64(BlockSize)+uint64(g.tagSize) {
		return nil, errOpen
	}

	tag := ciphertext[len(ciphertext)-g.tagSize:]
	ciphertext = ciphertext[:len(ciphertext)-g.tagSize]

	// See GCM spec, section 7.1.
	var counter, tagMask [gcmBlockSize]byte

	if len(nonce) == gcmStandardNonceSize {
		// Init counter to nonce||1
		copy(counter[:], nonce)
		counter[gcmBlockSize-1] = 1
	} else {
		// Otherwise counter = GHASH(nonce)
		gcmAesData(&g.productTable, nonce, &counter)
		gcmAesFinish(&g.productTable, &tagMask, &counter, uint64(len(nonce)), uint64(0))
	}

	encryptBlockAsm(len(g.ks)/4-1, &g.ks[0], &tagMask[0], &counter[0])

	var expectedTag [gcmTagSize]byte
	gcmAesData(&g.productTable, data, &expectedTag)

	ret, out := sliceForAppend(dst, len(ciphertext))
	if subtleoverlap.InexactOverlap(out, ciphertext) {
		panic("crypto/cipher: invalid buffer overlap")
	}
	if len(ciphertext) > 0 {
		gcmAesDec(&g.productTable, out, ciphertext, &counter, &expectedTag, g.ks)
	}
	gcmAesFinish(&g.productTable, &tagMask, &expectedTag, uint64(len(ciphertext)), uint64(len(data)))

	if subtle.ConstantTimeCompare(expectedTag[:g.tagSize], tag) != 1 {
		for i := range out {
			out[i] = 0
		}
		return nil, errOpen
	}

	return ret, nil
}

// gcmInit helper functions
// required to avoid dependencies on hex string package
// these convert string constants to byte arrays
// future - remove helper funcs, replace ascii strings with byte arrays
// future +1 - add support for 64-byte aligned constants to go asm
// making mask and constant declarations unnecessary at the go level
// replace with aligned defined byte tables in gcmv_amd64.s

// fromHexChar converts a hex character into its value and a success flag.
func fromHexChar(c byte) byte {
	switch {
	case '0' <= c && c <= '9':
		return c - '0'
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10
	}
	return 0
}

// Decode decodes src into DecodedLen(len(src)) bytes,
// returning the actual number of bytes written to dst.
func decode(dst, src []byte) int {
	i, j := 0, 1
	for ; j < len(src); j += 2 {
		a := fromHexChar(src[j-1])
		b := fromHexChar(src[j])
		dst[i] = (a << 4) | b
		i++
	}
	return i
}

// DecodeString returns the bytes represented by the hexadecimal string s.
func decodeString(s string) []byte {
	src := []byte(s)
	n := decode(src, src)
	return src[:n]
}

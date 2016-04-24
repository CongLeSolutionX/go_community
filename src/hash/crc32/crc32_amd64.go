// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package crc32

// This file contains the code to call the SSE 4.2 version of the Castagnoli
// and IEEE CRC.

// haveSSE41/haveSSE42/haveCLMUL are defined in crc_amd64.s and use
// CPUID to test for SSE 4.1, 4.2 and CLMUL support.
func haveSSE41() bool
func haveSSE42() bool
func haveCLMUL() bool

// castagnoliSSE42 is defined in crc_amd64.s and uses the SSE4.2 CRC32
// instruction.
//go:noescape
func castagnoliSSE42(crc uint32, p []byte) uint32

//go:noescape
func castagnoliSSE42String(crc uint32, s string) uint32

// ieeeCLMUL is defined in crc_amd64.s and uses the PCLMULQDQ
// instruction as well as SSE 4.1.
//go:noescape
func ieeeCLMUL(crc uint32, p []byte) uint32

//go:noescape
func ieeeCLMULString(crc uint32, s string) uint32

var sse42 = haveSSE42()
var useFastIEEE = haveCLMUL() && haveSSE41()

func updateCastagnoli(crc uint32, p []byte) uint32 {
	if sse42 {
		return castagnoliSSE42(crc, p)
	}
	return updateCastagnoliGeneric(crc, p)
}

func updateCastagnoliString(crc uint32, s string) uint32 {
	if sse42 {
		return castagnoliSSE42String(crc, s)
	}
	return updateCastagnoliGenericString(crc, s)
}

func updateIEEE(crc uint32, p []byte) uint32 {
	if useFastIEEE && len(p) >= 64 {
		left := len(p) & 15
		do := len(p) - left
		crc = ^ieeeCLMUL(^crc, p[:do])
		if left > 0 {
			crc = update(crc, IEEETable, p[do:])
		}
		return crc
	}
	return updateIEEEGeneric(crc, p)
}

func updateIEEEString(crc uint32, s string) uint32 {
	if useFastIEEE && len(s) >= 64 {
		left := len(s) & 15
		do := len(s) - left
		crc = ^ieeeCLMULString(^crc, s[:do])
		if left > 0 {
			crc = updateString(crc, IEEETable, s[do:])
		}
		return crc
	}
	return updateIEEEGenericString(crc, s)
}

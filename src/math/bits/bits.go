// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package bits implements bit counting and manipulation
// functions for the predeclared unsigned integer types.
package bits

// LeadingZerosN returns the number of leading zero bits in x.
// N is one of 8, 16, 32, 64.
// The result is N for x == 0.
func LeadingZeros8(x uint8) int   { return nlz(uint64(x)<<(64-8) | 1<<(64-8-1)) }
func LeadingZeros16(x uint16) int { return nlz(uint64(x)<<(64-16) | 1<<(64-16-1)) }
func LeadingZeros32(x uint32) int { return nlz(uint64(x)<<(64-32) | 1<<(64-32-1)) }
func LeadingZeros64(x uint64) int { return nlz(x) }

// TrailingZerosN returns the number of trailing zero bits in x.
// N is one of 8, 16, 32, 64.
// The result is N for x == 0.
func TrailingZeros8(x uint8) int   { return ntz(uint64(x) | 1<<8) }
func TrailingZeros16(x uint16) int { return ntz(uint64(x) | 1<<16) }
func TrailingZeros32(x uint32) int { return ntz(uint64(x) | 1<<32) }
func TrailingZeros64(x uint64) int { return ntz(x) }

// OnesN returns the number of set bits in x.
// N is one of 8, 16, 32, 64.
func Ones8(x uint8) int   { return pop(uint64(x)) }
func Ones16(x uint16) int { return pop(uint64(x)) }
func Ones32(x uint32) int { return pop(uint64(x)) }
func Ones64(x uint64) int { return pop(x) }

// RotateLeftN returns the value of x rotated left by k bits.
// N is one of 8, 16, 32, 64.
func RotateLeft8(x uint8, k uint) uint8    { return uint8(rot(uint64(x), 8, k%8)) }
func RotateLeft16(x uint16, k uint) uint16 { return uint16(rot(uint64(x), 16, k%16)) }
func RotateLeft32(x uint32, k uint) uint32 { return uint32(rot(uint64(x), 32, k%32)) }
func RotateLeft64(x uint64, k uint) uint64 { return rot(x, 64, k%64) }

// RotateRightN returns the value of x rotated right by k bits.
// N is one of 8, 16, 32, 64.
func RotateRight8(x uint8, k uint) uint8    { return uint8(rot(uint64(x), 8, 8-k%8)) }
func RotateRight16(x uint16, k uint) uint16 { return uint16(rot(uint64(x), 16, 16-k%16)) }
func RotateRight32(x uint32, k uint) uint32 { return uint32(rot(uint64(x), 32, 32-k%32)) }
func RotateRight64(x uint64, k uint) uint64 { return rot(x, 64, 64-k%64) }

// ReverseN returns the value of x with its bits in reversed order.
// N is one of 8, 16, 32, 64.
func Reverse8(x uint8) uint8    { return uint8(rev(uint64(x)) >> (64 - 8)) }
func Reverse16(x uint16) uint16 { return uint16(rev(uint64(x)) >> (64 - 16)) }
func Reverse32(x uint32) uint32 { return uint32(rev(uint64(x)) >> (64 - 32)) }
func Reverse64(x uint64) uint64 { return rev(x) }

// ReverseBytesN returns the value of x with its bytes in reversed order.
func ReverseBytes16(x uint16) uint16 { return uint16(swap(uint64(x)) >> (64 - 16)) }
func ReverseBytes32(x uint32) uint32 { return uint32(swap(uint64(x)) >> (64 - 32)) }
func ReverseBytes64(x uint64) uint64 { return swap(x) }

// LogN returns the integer binary logarithm of x.
// N is one of 8, 16, 32, 64.
// The result is the integer n with 2**n <= x < 2**(n+1) for x != 0.
// The result is -1 for x == 0.
func Log8(x uint8) int   { return log(uint64(x)) }
func Log16(x uint16) int { return log(uint64(x)) }
func Log32(x uint32) int { return log(uint64(x)) }
func Log64(x uint64) int { return log(x) }

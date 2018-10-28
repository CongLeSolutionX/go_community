// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math

import "unsafe"

// Float32bits returns the IEEE 754 binary representation of f.
// the most significant bit of the return value will contain the sign bit,
// the next 8 most significant bits the exponent,
// and the 23 least significant bits the significand or mantissa.
func Float32bits(f float32) uint32 { return *(*uint32)(unsafe.Pointer(&f)) }

// Float32frombits returns the floating point number corresponding
// to the IEEE 754 binary representation b.
// The most significant bit of b should contain the sign bit,
// the next 8 most significant bits the exponent,
// and the 23 least significant bits the significand or mantissa.
func Float32frombits(b uint32) float32 { return *(*float32)(unsafe.Pointer(&b)) }

// Float64bits returns the IEEE 754 binary representation of f.
// the most significant bit of the return value will contain the sign bit,
// the next 12 most significant bits the exponent,
// and the 51 least significant bits the significand or mantissa.
func Float64bits(f float64) uint64 { return *(*uint64)(unsafe.Pointer(&f)) }

// Float64frombits returns the floating point number corresponding
// the IEEE 754 binary representation b.
// The most significant bit of b should contain the sign bit,
// the next 12 most significant bits the exponent,
// and the 51 least significant bits the significand or mantissa.
func Float64frombits(b uint64) float64 { return *(*float64)(unsafe.Pointer(&b)) }

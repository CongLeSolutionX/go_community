// -lang=go1.12

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Check Go language version-specific errors.

package p

// numeric literals
const (
	_ = 1_000 // ERR underscores in numeric literals requires go1.13 or later
	_ = 0b111 // ERR binary literals requires go1.13 or later
	_ = 0o567 // ERROR 0o/0O-style octal literals requires go1.13 or later
	_ = 0xabc // ok
	_ = 0x0p1 // ERR hexadecimal floating-point literals requires go1.13 or later

	_ = 0B111 // ERR binary
	_ = 0O567 // ERR octal
	_ = 0Xabc // ok
	_ = 0X0P1 // ERR hexadecimal floating-point

	_ = 1_000i // ERR underscores
	_ = 0b111i // ERR binary
	_ = 0o567i // ERR octal
	_ = 0xabci // ERR hexadecimal floating-point
	_ = 0x0p1i // ERR hexadecimal floating-point
)

// signed shift counts
var (
	s int
	_ = 1 << s // ERR invalid operation: signed shift count s (variable of type int) requires go1.13 or later
	_ = 1 >> s // ERR signed shift count
)

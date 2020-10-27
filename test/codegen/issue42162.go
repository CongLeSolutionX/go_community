// asmcheck

// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test that generating ORshiftLL/ORshiftRL instead of
// bitfield operations. Issue 42162.

package codegen

func orshiftLL(hi, lo uint32) uint64 {
	return uint64(hi) << 18  | uint64(lo)  // arm64:"ORR\tR[0-9]+<<18", -"UBFIZ"
}

func orshiftRL(hi, lo uint32) uint64 {
	return uint64(hi) >> 30 | uint64(lo)   // arm64:"ORR\tR[0-9]+>>30", -"UBFX"
}

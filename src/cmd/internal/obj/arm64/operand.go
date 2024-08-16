// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package arm64

const (
	REG_R = iota
	REG_F
	REG_V
	REG_Z
	REG_P
	REG_V_INDEXED
	REG_Z_INDEXED
	REG_P_INDEXED
	MEM_ADDR
)
const typeMask = (1 << 7) - 1
const subtypeOffset = 8
const subtypeMask = ((1 << 8) - 1) << subtypeOffset

// Bit extractors for Type and Extension from Register.Format()
func getType(fmt int) int {
	return int(uint(fmt) & typeMask)
}

func getSubtype(fmt int) int {
	return int(uint(fmt) & subtypeMask)
}

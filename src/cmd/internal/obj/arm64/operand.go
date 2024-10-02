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
	IMM
	PRFOP
	PATTERN
	REGLIST_R
	REGLIST_F
	REGLIST_V
	REGLIST_Z
	REGLIST_P
)

/* An operand format specifies the type of an operand, used for pattern matching
 * the Addr structures passed in by the assembly writer/front-end of the compiler,
 * and mapping them to instructions.
 *
 * 63                 16             8          0
 *  +------------------+-------------+----------+
 *  | unspecified (48) | subtype (8) | type (8) |
 *  +------------------+-------------+----------+
 * type: The primary type of the operand listed in the enumeration above, e.g.
 *       register/memory address.
 * subtype: Range used for sub-types of an operand.
 *
 * The top 48 bits of the format may be used for any means.
 */
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

// Supported immediate types
const (
	IMM_INT = iota << subtypeOffset
	IMM_FLOAT
)

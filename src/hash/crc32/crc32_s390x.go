// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package crc32

const (
	vxMinLen    = 64
	vxAlignMask = 15 // align to 16 bytes
)

// hasVectorFacility reports whether the machine has the z/Architecture
// vector facility installed and enabled.
func hasVectorFacility() bool

var hasVX = hasVectorFacility()

// vectorizedCastagnoli implements CRC32 using vector instructions.
// It is defined in crc32_s390x.s.
//go:noescape
func vectorizedCastagnoli(crc uint32, p []byte) uint32

// vectorizedIEEE implements CRC32 using vector instructions.
// It is defined in crc32_s390x.s.
//go:noescape
func vectorizedIEEE(crc uint32, p []byte) uint32

func archAvailableCastagnoli() bool {
	return hasVX
}

func archInitCastagnoli() {
	if !hasVX {
		panic("not available")
	}
	// No initialization necessary.
}

// archUpdateCastagnoli calculates the checksum of p using
// vectorizedCastagnoli.
func archUpdateCastagnoli(crc uint32, p []byte) uint32 {
	if !hasVX {
		panic("not available")
	}
	// Use vectorized function if data length is above threshold.
	if len(p) >= vxMinLen {
		aligned := len(p) & ^vxAlignMask
		crc = vectorizedCastagnoli(crc, p[:aligned])
		p = p[aligned:]
	}
	if len(p) == 0 {
		return crc
	}
	return simpleUpdate(crc, castagnoliTable, p)
}

func archAvailableIEEE() bool {
	return hasVX
}

func archInitIEEE() {
	if !hasVX {
		panic("not available")
	}
	// No initialization necessary.
}

// updateIEEE calculates the checksum of p using vectorizedIEEE if
// possible and falling back onto genericIEEE as needed.
func updateIEEE(crc uint32, p []byte) uint32 {
	if !hasVX {
		panic("not available")
	}
	// Use vectorized function if data length is above threshold.
	if len(p) >= vxMinLen {
		aligned := len(p) & ^vxAlignMask
		crc = vectorizedIEEE(crc, p[:aligned])
		p = p[aligned:]
	}
	if len(p) == 0 {
		return crc
	}
	return simpleUpdate(crc, IEEETable, p)
}

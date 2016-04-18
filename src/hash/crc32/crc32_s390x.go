// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package crc32

import (
	"unsafe"
)

const (
	VX_MIN_LEN    = 64
	VX_ALIGNMENT  = 16
	VX_ALIGN_MASK = (VX_ALIGNMENT - 1)
)

// check if the machine has z/Architecture vector facility installed and enabled
func hasVectorFacility() bool

var hasVX = hasVectorFacility()

// crc32c_le_vgfm and crc32_le_vgfm are defined in crc32_s390x.s
func crc32c_le_vgfm(crc uint32, p []byte) uint32

func crc32_le_vgfm(crc uint32, p []byte) uint32

type checksumFunc func(crc uint32, p []byte) uint32

// software implementations of CRC32 must comply with the checksumFunc interface

func softwareCastagnoli(crc uint32, p []byte) uint32 {
	// Use slicing-by-8 on larger inputs.
	if len(p) >= sliceBy8Cutoff {
		return updateSlicingBy8(crc, castagnoliTable8, p)
	}
	return update(crc, castagnoliTable, p)
}

func softwareIEEE(crc uint32, p []byte) uint32 {
	// Use slicing-by-8 on larger inputs.
	if len(p) >= sliceBy8Cutoff {
		ieeeTable8Once.Do(func() {
			ieeeTable8 = makeTable8(IEEE)
		})
		return updateSlicingBy8(crc, ieeeTable8, p)
	}
	return update(crc, IEEETable, p)
}

// generic vectorized CRC32 algorithm
func vectorizedCRC32(crc uint32, p []byte, swCksum checksumFunc, vxCksum checksumFunc) uint32 {

	p_addr := uintptr(unsafe.Pointer(&p[0]))
	prealign := 0

	// prealigned data to suit the vector register if needed
	if (p_addr & VX_ALIGN_MASK) != 0 {
		prealign = VX_ALIGNMENT - (int(p_addr & uintptr(VX_ALIGN_MASK)))
		crc = ^crc
		crc = swCksum(crc, p[0:prealign])
		crc = ^crc
	}

	p_len := len(p) - prealign

	aligned := p_len & ^VX_ALIGN_MASK
	remaining := p_len & VX_ALIGN_MASK

	// Call accelerated vector go routine
	crc = vxCksum(crc, p[prealign:aligned])

	// process remainig data
	if remaining != 0 {
		crc = ^crc
		crc = swCksum(crc, p[prealign+aligned:])
		crc = ^crc
	}
	return crc
}

func updateCastagnoli(crc uint32, p []byte) uint32 {
	// use vectorized function if vector facility is available and data length > 64 bytes
	if hasVX && (len(p) > VX_MIN_LEN) {
		crc = ^crc
		crc = vectorizedCRC32(crc, p, softwareCastagnoli, crc32c_le_vgfm)
		return ^crc
	} else {
		return softwareCastagnoli(crc, p)

	}
}

func updateIEEE(crc uint32, p []byte) uint32 {
	// use vectorized function if vector facility is available and data length > 64 bytes
	if hasVX && (len(p) > VX_MIN_LEN) {
		crc = ^crc
		crc = vectorizedCRC32(crc, p, softwareIEEE, crc32_le_vgfm)
		return ^crc
	} else {
		return softwareIEEE(crc, p)
	}
}

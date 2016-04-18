// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package crc32

import (
	"unsafe"
)

const (
	vxMinLen    = 64
	vxAlignment = 16
	vxAlignMask = vxAlignment - 1
)

// hasVectorFacility reports whether the machine has the z/Architecture
// vector facility installed and enabled.
func hasVectorFacility() bool

var hasVX = hasVectorFacility()

// checksumFunc is implemented by CRC32 implementations.
type checksumFunc func(crc uint32, p []byte) uint32

// vectorizedCastagnoli implements CRC32 using vector instructions.
// It is defined in crc32_s390x.s.
//go:noescape
func vectorizedCastagnoli(crc uint32, p []byte) uint32

// vectorizedIEEE implements CRC32 using vector instructions.
// It is defined in crc32_s390x.s.
//go:noescape
func vectorizedIEEE(crc uint32, p []byte) uint32

func genericCastagnoli(crc uint32, p []byte) uint32 {
	// Use slicing-by-8 on larger inputs.
	if len(p) >= sliceBy8Cutoff {
		return updateSlicingBy8(crc, castagnoliTable8, p)
	}
	return update(crc, castagnoliTable, p)
}

func genericIEEE(crc uint32, p []byte) uint32 {
	// Use slicing-by-8 on larger inputs.
	if len(p) >= sliceBy8Cutoff {
		ieeeTable8Once.Do(func() {
			ieeeTable8 = makeTable8(IEEE)
		})
		return updateSlicingBy8(crc, ieeeTable8, p)
	}
	return update(crc, IEEETable, p)
}

// vectorizedCRC32 calculates the checksum using the callbacks given.
// generic is used to calculate the checksum of the head and tail of
// the data if the data is unaligned. vector is used to calculate the
// checksum of the aligned body of the data.
func vectorizedCRC32(crc uint32, p []byte, generic, vector checksumFunc) uint32 {
	pAddr := uintptr(unsafe.Pointer(&p[0]))
	prealign := 0

	// prealigned data to suit the vector register if needed
	if pAddr&vxAlignMask != 0 {
		prealign = vxAlignment - int(pAddr&vxAlignMask)
		crc = ^crc
		crc = generic(crc, p[:prealign])
		crc = ^crc
	}

	pLen := len(p) - prealign
	aligned := pLen & ^vxAlignMask
	remaining := pLen & vxAlignMask

	crc = vector(crc, p[prealign:aligned])

	// process remaining data
	if remaining != 0 {
		crc = ^crc
		crc = generic(crc, p[prealign+aligned:])
		crc = ^crc
	}
	return crc
}

func updateCastagnoli(crc uint32, p []byte) uint32 {
	// Use vectorized function if vector facility is available and
	// data length is above threshold.
	if hasVX && len(p) > vxMinLen {
		crc = ^crc
		crc = vectorizedCRC32(crc, p, genericCastagnoli, vectorizedCastagnoli)
		return ^crc
	}
	return genericCastagnoli(crc, p)
}

func updateIEEE(crc uint32, p []byte) uint32 {
	// Use vectorized function if vector facility is available and
	// data length is above threshold.
	if hasVX && len(p) > vxMinLen {
		crc = ^crc
		crc = vectorizedCRC32(crc, p, genericIEEE, vectorizedIEEE)
		return ^crc
	}
	return genericIEEE(crc, p)
}

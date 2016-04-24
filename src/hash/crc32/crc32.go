// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package crc32 implements the 32-bit cyclic redundancy check, or CRC-32,
// checksum. See http://en.wikipedia.org/wiki/Cyclic_redundancy_check for
// information.
//
// Polynomials are represented in LSB-first form also known as reversed representation.
//
// See http://en.wikipedia.org/wiki/Mathematics_of_cyclic_redundancy_checks#Reversed_representations_and_reciprocal_polynomials
// for information.
package crc32

import (
	"hash"
	"sync"
)

// The size of a CRC-32 checksum in bytes.
const Size = 4

// Use "slice by 8" when payload >= this value.
const sliceBy8Cutoff = 16

// Predefined polynomials.
const (
	// IEEE is by far and away the most common CRC-32 polynomial.
	// Used by ethernet (IEEE 802.3), v.42, fddi, gzip, zip, png, ...
	IEEE = 0xedb88320

	// Castagnoli's polynomial, used in iSCSI.
	// Has better error detection characteristics than IEEE.
	// http://dx.doi.org/10.1109/26.231911
	Castagnoli = 0x82f63b78

	// Koopman's polynomial.
	// Also has better error detection characteristics than IEEE.
	// http://dx.doi.org/10.1109/DSN.2002.1028931
	Koopman = 0xeb31d82e
)

// Table is a 256-word table representing the polynomial for efficient processing.
type Table [256]uint32

// castagnoliTable points to a lazily initialized Table for the Castagnoli
// polynomial. MakeTable will always return this value when asked to make a
// Castagnoli table so we can compare against it to find when the caller is
// using this polynomial.
var castagnoliTable *Table
var castagnoliTable8 *slicing8Table
var castagnoliOnce sync.Once

func castagnoliInit() {
	castagnoliTable = makeTable(Castagnoli)
	castagnoliTable8 = makeTable8(Castagnoli)
}

// IEEETable is the table for the IEEE polynomial.
var IEEETable = makeTable(IEEE)

// slicing8Table is array of 8 Tables
type slicing8Table [8]Table

// ieeeTable8 is the slicing8Table for IEEE
var ieeeTable8 *slicing8Table
var ieeeTable8Once sync.Once

// MakeTable returns a Table constructed from the specified polynomial.
// The contents of this Table must not be modified.
func MakeTable(poly uint32) *Table {
	switch poly {
	case IEEE:
		return IEEETable
	case Castagnoli:
		castagnoliOnce.Do(castagnoliInit)
		return castagnoliTable
	}
	return makeTable(poly)
}

// makeTable returns the Table constructed from the specified polynomial.
func makeTable(poly uint32) *Table {
	t := new(Table)
	for i := 0; i < 256; i++ {
		crc := uint32(i)
		for j := 0; j < 8; j++ {
			if crc&1 == 1 {
				crc = (crc >> 1) ^ poly
			} else {
				crc >>= 1
			}
		}
		t[i] = crc
	}
	return t
}

// makeTable8 returns slicing8Table constructed from the specified polynomial.
func makeTable8(poly uint32) *slicing8Table {
	t := new(slicing8Table)
	t[0] = *makeTable(poly)
	for i := 0; i < 256; i++ {
		crc := t[0][i]
		for j := 1; j < 8; j++ {
			crc = t[0][crc&0xFF] ^ (crc >> 8)
			t[j][i] = crc
		}
	}
	return t
}

// digest represents the partial evaluation of a checksum.
type digest struct {
	crc uint32
	tab *Table
}

// New creates a new hash.Hash32 computing the CRC-32 checksum
// using the polynomial represented by the Table.
// Its Sum method will lay the value out in big-endian byte order.
func New(tab *Table) hash.Hash32 { return &digest{0, tab} }

// NewIEEE creates a new hash.Hash32 computing the CRC-32 checksum
// using the IEEE polynomial.
// Its Sum method will lay the value out in big-endian byte order.
func NewIEEE() hash.Hash32 { return New(IEEETable) }

func (d *digest) Size() int { return Size }

func (d *digest) BlockSize() int { return 1 }

func (d *digest) Reset() { d.crc = 0 }

func update(crc uint32, tab *Table, p []byte) uint32 {
	crc = ^crc
	for _, v := range p {
		crc = tab[byte(crc)^v] ^ (crc >> 8)
	}
	return ^crc
}

// updateSlicingBy8 updates CRC using Slicing-by-8
func updateSlicingBy8(crc uint32, tab *slicing8Table, p []byte) uint32 {
	crc = ^crc
	for len(p) > 8 {
		crc ^= uint32(p[0]) | uint32(p[1])<<8 | uint32(p[2])<<16 | uint32(p[3])<<24
		crc = tab[0][p[7]] ^ tab[1][p[6]] ^ tab[2][p[5]] ^ tab[3][p[4]] ^
			tab[4][crc>>24] ^ tab[5][(crc>>16)&0xFF] ^
			tab[6][(crc>>8)&0xFF] ^ tab[7][crc&0xFF]
		p = p[8:]
	}
	crc = ^crc
	if len(p) == 0 {
		return crc
	}
	return update(crc, &tab[0], p)
}

// updateSlicingBy8String updates CRC using Slicing-by-8
func updateSlicingBy8String(crc uint32, tab *slicing8Table, s string) uint32 {
	crc = ^crc
	for len(s) > 8 {
		crc ^= uint32(s[0]) | uint32(s[1])<<8 | uint32(s[2])<<16 | uint32(s[3])<<24
		crc = tab[0][s[7]] ^ tab[1][s[6]] ^ tab[2][s[5]] ^ tab[3][s[4]] ^
			tab[4][crc>>24] ^ tab[5][(crc>>16)&0xFF] ^
			tab[6][(crc>>8)&0xFF] ^ tab[7][crc&0xFF]
		s = s[8:]
	}
	crc = ^crc
	if len(s) == 0 {
		return crc
	}
	return updateString(crc, &tab[0], s)
}

func updateString(crc uint32, tab *Table, s string) uint32 {
	crc = ^crc
	for i := 0; i < len(s); i++ {
		crc = tab[byte(crc)^s[i]] ^ (crc >> 8)
	}
	return ^crc
}

func updateCastagnoliGeneric(crc uint32, p []byte) uint32 {
	// Use slicing-by-8 on larger inputs.
	if len(p) >= sliceBy8Cutoff {
		return updateSlicingBy8(crc, castagnoliTable8, p)
	}
	return update(crc, castagnoliTable, p)
}

func updateCastagnoliGenericString(crc uint32, s string) uint32 {
	// Use slicing-by-8 on larger inputs.
	if len(s) >= sliceBy8Cutoff {
		return updateSlicingBy8String(crc, castagnoliTable8, s)
	}
	return updateString(crc, castagnoliTable, s)
}

func updateIEEEGeneric(crc uint32, p []byte) uint32 {
	// Use slicing-by-8 on larger inputs.
	if len(p) >= sliceBy8Cutoff {
		ieeeTable8Once.Do(func() {
			ieeeTable8 = makeTable8(IEEE)
		})
		return updateSlicingBy8(crc, ieeeTable8, p)
	}
	return update(crc, IEEETable, p)
}

func updateIEEEGenericString(crc uint32, s string) uint32 {
	// Use slicing-by-8 on larger inputs.
	if len(s) >= sliceBy8Cutoff {
		ieeeTable8Once.Do(func() {
			ieeeTable8 = makeTable8(IEEE)
		})
		return updateSlicingBy8String(crc, ieeeTable8, s)
	}
	return updateString(crc, IEEETable, s)
}

// Update returns the result of adding the bytes in p to the crc.
func Update(crc uint32, tab *Table, p []byte) uint32 {
	switch tab {
	case castagnoliTable:
		return updateCastagnoli(crc, p)
	case IEEETable:
		return updateIEEE(crc, p)
	}
	return update(crc, tab, p)
}

// UpdateString returns the result of adding the bytes in s to the crc.
func UpdateString(crc uint32, tab *Table, s string) uint32 {
	switch tab {
	case castagnoliTable:
		return updateCastagnoliGenericString(crc, s)
	case IEEETable:
		return updateIEEEGenericString(crc, s)
	}
	return updateString(crc, tab, s)
}

func (d *digest) Write(p []byte) (n int, err error) {
	d.crc = Update(d.crc, d.tab, p)
	return len(p), nil
}

func (d *digest) WriteString(s string) (n int, err error) {
	d.crc = UpdateString(d.crc, d.tab, s)
	return len(s), nil
}

func (d *digest) Sum32() uint32 { return d.crc }

func (d *digest) Sum(in []byte) []byte {
	s := d.Sum32()
	return append(in, byte(s>>24), byte(s>>16), byte(s>>8), byte(s))
}

// Checksum returns the CRC-32 checksum of data
// using the polynomial represented by the Table.
func Checksum(data []byte, tab *Table) uint32 { return Update(0, tab, data) }

// ChecksumString returns the CRC-32 checksum of s
// using the polynomial represented by the Table.
func ChecksumString(s string, tab *Table) uint32 {
	return UpdateString(0, tab, s)
}

// ChecksumIEEE returns the CRC-32 checksum of data
// using the IEEE polynomial.
func ChecksumIEEE(data []byte) uint32 { return updateIEEE(0, data) }

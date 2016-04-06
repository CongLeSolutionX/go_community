// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sys

import "encoding/binary"

// ArchChar represents an architecture family.
type ArchChar byte

const (
	Char386    ArchChar = '8'
	CharAMD64  ArchChar = '6'
	CharARM    ArchChar = '5'
	CharARM64  ArchChar = '7'
	CharMIPS64 ArchChar = '0'
	CharPPC64  ArchChar = '9'
	CharS390X  ArchChar = 'z'
)

// Arch represents an individual architecture.
type Arch struct {
	Name string
	Char ArchChar

	ByteOrder binary.ByteOrder

	IntSize int
	PtrSize int
	RegSize int

	MinLC int
}

// HasChar reports whether a is a member of any of the specified
// architecture families.
func (a *Arch) HasChar(xs ...ArchChar) bool {
	for _, x := range xs {
		if a.Char == x {
			return true
		}
	}
	return false
}

var Arch386 = Arch{
	Name:      "386",
	Char:      Char386,
	ByteOrder: binary.LittleEndian,
	IntSize:   4,
	PtrSize:   4,
	RegSize:   4,
	MinLC:     1,
}

var ArchAMD64 = Arch{
	Name:      "amd64",
	Char:      CharAMD64,
	ByteOrder: binary.LittleEndian,
	IntSize:   8,
	PtrSize:   8,
	RegSize:   8,
	MinLC:     1,
}

var ArchAMD64P32 = Arch{
	Name:      "amd64p32",
	Char:      CharAMD64,
	ByteOrder: binary.LittleEndian,
	IntSize:   4,
	PtrSize:   4,
	RegSize:   8,
	MinLC:     1,
}

var ArchARM = Arch{
	Name:      "arm",
	Char:      CharARM,
	ByteOrder: binary.LittleEndian,
	IntSize:   4,
	PtrSize:   4,
	RegSize:   4,
	MinLC:     4,
}

var ArchARM64 = Arch{
	Name:      "arm64",
	Char:      CharARM64,
	ByteOrder: binary.LittleEndian,
	IntSize:   8,
	PtrSize:   8,
	RegSize:   8,
	MinLC:     4,
}

var ArchPPC64 = Arch{
	Name:      "ppc64",
	Char:      CharPPC64,
	ByteOrder: binary.BigEndian,
	IntSize:   8,
	PtrSize:   8,
	RegSize:   8,
	MinLC:     4,
}

var ArchPPC64LE = Arch{
	Name:      "ppc64le",
	Char:      CharPPC64,
	ByteOrder: binary.LittleEndian,
	IntSize:   8,
	PtrSize:   8,
	RegSize:   8,
	MinLC:     4,
}

var ArchMIPS64 = Arch{
	Name:      "mips64",
	Char:      CharMIPS64,
	ByteOrder: binary.BigEndian,
	IntSize:   8,
	PtrSize:   8,
	RegSize:   8,
	MinLC:     4,
}

var ArchMIPS64LE = Arch{
	Name:      "mips64le",
	Char:      CharMIPS64,
	ByteOrder: binary.LittleEndian,
	IntSize:   8,
	PtrSize:   8,
	RegSize:   8,
	MinLC:     4,
}

var ArchS390X = Arch{
	Name:      "s390x",
	Char:      CharS390X,
	ByteOrder: binary.BigEndian,
	IntSize:   8,
	PtrSize:   8,
	RegSize:   8,
	MinLC:     2,
}

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package obj

// X86suffixes is a complete list of possible opcode suffix combinations.
// Basically, it "maps" uint8 suffix bits to their string representation.
// This the exception of first and last elements, order is not important.
var X86suffixes = [...]string{
	"", // Map empty suffix to empty string.

	"Z",

	"SAE",
	"SAE.Z",

	"RN_SAE",
	"RZ_SAE",
	"RD_SAE",
	"RU_SAE",
	"RN_SAE.Z",
	"RZ_SAE.Z",
	"RD_SAE.Z",
	"RU_SAE.Z",

	"BCST",
	"BCST.Z",

	"<bad suffix>",
}

// X86suffix represents x86-specific opcode suffix.
// Compound (multi-part) suffixes expressed with single X86suffix value.
//
// uint8 type is used to fit obj.Prog.Scond (see link.go).
type X86suffix uint8

// x86badSuffix is used to represent all invalid suffix combinations.
const x86badSuffix = X86suffix(len(X86suffixes) - 1)

// NewX86suffix returns X86suffix object that matches suffixes string.
//
// If no matching suffix is found, special "invalid" suffix is returned.
// Use IsValid method to check against this case.
func NewX86suffix(suffixes string) X86suffix {
	for i := range X86suffixes {
		if X86suffixes[i] == suffixes {
			return X86suffix(i)
		}
	}
	return x86badSuffix
}

// IsValid returns true for valid suffix.
// Empty suffix is considered as valid.
func (suffix X86suffix) IsValid() bool {
	return suffix != x86badSuffix
}

// String returns suffix printed representation.
// It matches the string that was used to create suffix with NewX86Suffix()
// for valid suffixes.
// For all invalid suffixes, special marker is returned.
func (suffix X86suffix) String() string {
	return X86suffixes[suffix]
}

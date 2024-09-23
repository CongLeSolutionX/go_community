// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !purego

package sha256

import "internal/cpu"

//go:noescape
func blockAMD64(dig *digest, p []byte)

var useAVX2 = cpu.X86.HasAVX2 && cpu.X86.HasBMI2

//go:noescape
func blockAVX2(dig *digest, p []byte)

var useSHANI = useAVX2 && cpu.X86.HasSHA

//go:noescape
func blockSHANI(dig *digest, p []byte)

func block(dig *digest, p []byte) {
	if useSHANI {
		blockSHANI(dig, p)
	} else if useAVX2 {
		blockAVX2(dig, p)
	} else {
		blockAMD64(dig, p)
	}
}

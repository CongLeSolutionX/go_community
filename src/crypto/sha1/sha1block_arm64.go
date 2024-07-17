// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !purego

package sha1

import "internal/cpu"
import "crypto/internal/sharc"

var k = sharc.K1

//go:noescape
func sha1block(h []uint32, p []byte, k []uint32)

func block(dig *digest, p []byte) {
	if !cpu.ARM64.HasSHA1 {
		blockGeneric(dig, p)
	} else {
		h := dig.h[:]
		sha1block(h, p, k)
	}
}

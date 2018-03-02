// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !amd64,!s390x,!arm64

package bytealg

import _ "unsafe" // for go:linkname

func IndexBytesByte(b []byte, c byte) int {
	for i, x := range b {
		if x == c {
			return i
		}
	}
	return -1
}

func IndexStringByte(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

//go:linkname IndexBytesByte bytes.IndexByte
//go:linkname IndexStringByte strings.IndexByte

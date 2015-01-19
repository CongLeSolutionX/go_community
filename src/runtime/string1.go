// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_core "runtime/internal/core"
	"unsafe"
)

func findnullw(s *uint16) int {
	if s == nil {
		return 0
	}
	p := (*[_core.MaxMem/2/2 - 1]uint16)(unsafe.Pointer(s))
	l := 0
	for p[l] != 0 {
		l++
	}
	return l
}

func strcmp(s1, s2 *byte) int32 {
	p1 := (*[_core.MaxMem/2 - 1]byte)(unsafe.Pointer(s1))
	p2 := (*[_core.MaxMem/2 - 1]byte)(unsafe.Pointer(s2))

	for i := uintptr(0); ; i++ {
		c1 := p1[i]
		c2 := p2[i]
		if c1 < c2 {
			return -1
		}
		if c1 > c2 {
			return +1
		}
		if c1 == 0 {
			return 0
		}
	}
}

func strncmp(s1, s2 *byte, n uintptr) int32 {
	p1 := (*[_core.MaxMem/2 - 1]byte)(unsafe.Pointer(s1))
	p2 := (*[_core.MaxMem/2 - 1]byte)(unsafe.Pointer(s2))

	for i := uintptr(0); i < n; i++ {
		c1 := p1[i]
		c2 := p2[i]
		if c1 < c2 {
			return -1
		}
		if c1 > c2 {
			return +1
		}
		if c1 == 0 {
			break
		}
	}
	return 0
}

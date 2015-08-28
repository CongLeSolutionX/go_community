// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package base

import (
	"unsafe"
)

type StringStruct struct {
	Str unsafe.Pointer
	Len int
}

func Index(s, t string) int {
	if len(t) == 0 {
		return 0
	}
	for i := 0; i < len(s); i++ {
		if s[i] == t[0] && hasprefix(s[i:], t) {
			return i
		}
	}
	return -1
}

func contains(s, t string) bool {
	return Index(s, t) >= 0
}

func hasprefix(s, t string) bool {
	return len(s) >= len(t) && s[:len(t)] == t
}

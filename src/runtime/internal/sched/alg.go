// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sched

import (
	"unsafe"
)

var UseAeshash bool

// in asm_*.s
func aeshash(p unsafe.Pointer, h, s uintptr) uintptr

// used in hash{32,64}.go to seed the hash function
var Hashkey [4]uintptr

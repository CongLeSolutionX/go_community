// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schedinit

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	"unsafe"
)

// Small malloc size classes >= 16 are the multiples of 16: 16, 32, 48, 64, 80, 96, 112, 128, 144, ...
// Each P holds a pool for defers with small arg sizes.
// Assign defer allocations to pools by rounding to 16, to match malloc size classes.

const (
	deferHeaderSize = unsafe.Sizeof(_core.Defer{})
	minDeferAlloc   = (deferHeaderSize + 15) &^ 15
	minDeferArgs    = minDeferAlloc - deferHeaderSize
)

// defer size class for arg size sz
//go:nosplit
func Deferclass(siz uintptr) uintptr {
	if siz <= minDeferArgs {
		return 0
	}
	return (siz - minDeferArgs + 15) / 16
}

// total size of memory block for defer with arg size sz
func Totaldefersize(siz uintptr) uintptr {
	if siz <= minDeferArgs {
		return minDeferAlloc
	}
	return deferHeaderSize + siz
}

// Ensure that defer arg sizes that map to the same defer size class
// also map to the same malloc size class.
func testdefersizes() {
	var m [len(_core.P{}.Deferpool)]int32

	for i := range m {
		m[i] = -1
	}
	for i := uintptr(0); ; i++ {
		defersc := Deferclass(i)
		if defersc >= uintptr(len(m)) {
			break
		}
		siz := Goroundupsize(Totaldefersize(i))
		if m[defersc] < 0 {
			m[defersc] = int32(siz)
			continue
		}
		if m[defersc] != int32(siz) {
			print("bad defer size class: i=", i, " siz=", siz, " defersc=", defersc, "\n")
			_lock.Throw("bad defer size class")
		}
	}
}

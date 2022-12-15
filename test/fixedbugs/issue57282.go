// run

// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build loong64 || mips64 || mips64le || riscv64

package main

import (
	"fmt"
	. "sync/atomic"
)

const (
	_magic = 0xfbcdefaf
)

func main() {
	var x struct {
		before uint32
		i      uint32
		after  uint32

		o uint32
		n uint32
	}

	x.before = _magic
	x.after = _magic

	for t := uint32(0x7FFFFFF0); t < 0x80000003; t += 1 {
		x.i = t + 0
		x.o = t + 0
		x.n = t + 1

		if !CompareAndSwapUint32(&x.i, x.o, x.n) {
			panic(fmt.Sprintf("should have swapped %#x %#x", x.o, x.n))
		}

		if x.i != x.n {
			panic(fmt.Sprintf("wrong x.i after swap: x.i=%#x x.n=%#x", x.i, x.n))
		}

		if x.before != _magic || x.after != _magic {
			panic(fmt.Sprintf("wrong magic: %#x _ %#x != %#x _ %#x", x.before, x.after, _magic, _magic))
		}
	}
}

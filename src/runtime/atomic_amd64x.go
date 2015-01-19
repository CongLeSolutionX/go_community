// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build amd64 amd64p32

package runtime

import (
	_core "runtime/internal/core"
	"unsafe"
)

//go:noescape
func xchg(ptr *uint32, new uint32) uint32

// xchgp cannot have a go:noescape annotation, because
// while ptr does not escape, new does. If new is marked as
// not escaping, the compiler will make incorrect escape analysis
// decisions about the value being xchg'ed.
// Instead, make xchgp a wrapper around the actual atomic.
// When calling the wrapper we mark ptr as noescape explicitly.

//go:nosplit
func xchgp(ptr unsafe.Pointer, new unsafe.Pointer) unsafe.Pointer {
	return xchgp1(_core.Noescape(ptr), new)
}

func xchgp1(ptr unsafe.Pointer, new unsafe.Pointer) unsafe.Pointer

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_base "runtime/internal/base"
	"unsafe"
)

// describes how to handle callback
type wincallbackcontext struct {
	gobody       unsafe.Pointer // go function to call
	argsize      uintptr        // callback arguments size (in bytes)
	restorestack uintptr        // adjust stack on return by (in bytes) (386 only)
	cleanstack   bool
}

type sigtabtt struct {
	flags int32
	name  *int8
}

type forcegcstate struct {
	lock _base.Mutex
	g    *_base.G
	idle uint32
}

/*
 * known to compiler
 */
const (
	_Structrnd = _base.RegSize
)

// startup_random_data holds random bytes initialized at startup.  These come from
// the ELF AT_RANDOM auxiliary vector (vdso_linux_amd64.go or os_linux_386.go).
var startupRandomData []byte

// extendRandom extends the random numbers in r[:n] to the whole slice r.
// Treats n<0 as n==0.
func extendRandom(r []byte, n int) {
	if n < 0 {
		n = 0
	}
	for n < len(r) {
		// Extend random bits using hash function & time seed
		w := n
		if w > 16 {
			w = 16
		}
		h := _base.Memhash(unsafe.Pointer(&r[n-w]), uintptr(_base.Nanotime()), uintptr(w))
		for i := 0; i < _base.PtrSize && n < len(r); i++ {
			r[n] = byte(h)
			n++
			h >>= 8
		}
	}
}

var (
	emptystring string
	lastg       *_base.G
	goos        *int8
	signote     _base.Note
	forcegc     forcegcstate

	// Information about what cpu features are available.
	// Set on startup in asm_{x86,amd64}.s.
	cpuid_ecx         uint32
	cpuid_edx         uint32
	lfenceBeforeRdtsc bool
)

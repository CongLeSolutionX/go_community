// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	"unsafe"
)

//go:linkname secret_count runtime/secret.count
func secret_count() int64 {
	return getg().secret
}

//go:linkname secret_inc runtime/secret.inc
func secret_inc() {
	getg().secret++
}

//go:linkname secret_dec runtime/secret.dec
func secret_dec() {
	gp := getg()
	gp.secret--

	// Note: we're doing this zeroing on every secret_dec, not just
	// the one that returns us to 0. Normally, gp.secret will only
	// ever be 0 or 1, so that isn't a big deal.

	// TODO: clear all the registers (except known safe ones like SP).
	// Q: what does "all" mean in this context? Do we need to
	// worry about just Go-allocatable registers, or do we need
	// to consider any register potentially used by assembly?

	// Figure out where the current bottom of stack is.
	sp := getsp()
	lo := gp.stack.lo

	// TODO: keep some sort of low water mark so that we don't have
	// to zero a potentially large stack if we used just a little
	// bit of it. That will allow us to use a higher value for
	// lo than gp.stack.lo.

	// Zero all the unused stack (below sp). Note: this relies on
	// the fact that memclrNoHeapPointers doesn't use any
	// stack. If it ever did need stack, we would have to wrap
	// this call in systemstack (and then also take into account
	// the stack space needed for the systemstack switch).
	switch GOARCH {
	default:
		memclrNoHeapPointers(unsafe.Pointer(lo), sp-lo)
	case "amd64":
		// Minor adjustment to avoid clobbering the return
		// address of the memclr call.  This is ok secret-wise
		// because the write of the return address by the call
		// instruction will "clear" that word for us.
		memclrNoHeapPointers(unsafe.Pointer(lo), (sp-8)-lo)
	case "386":
		// Ditto.
		memclrNoHeapPointers(unsafe.Pointer(lo), (sp-4)-lo)
	}

	// TODO: what to do when panicking? This code is not sufficient.

	// Clear the frame of secret_dec itself.  At this point none
	// of this frame is being used, so it is safe to do so.  Note
	// that secret_dec's frame actually has uninitialized entries!
	// The two arg slots reserved for the memclr call never get
	// set to anything. So this step is necessary to ensure no
	// secrets remain on the stack.
	csp := getcallersp()
	switch GOARCH {
	default:
		memclrNoHeapPointers(unsafe.Pointer(sp), csp-sp)
	case "amd64":
		// Minor adjustment to not clobber the return address
		// of secret_dec.  That stack slot was "cleared" by
		// the write of the return address by the call to
		// secret_dec in runtime/secret.Do.
		memclrNoHeapPointers(unsafe.Pointer(sp), (csp-8)-sp)
	case "386":
		// Ditto.
		memclrNoHeapPointers(unsafe.Pointer(sp), (csp-4)-sp)
	}

	// Don't put any code here: the stack frame's contents are gone!
}

//go:noinline
//go:nosplit
func getsp() uintptr {
	return getcallersp()
}

// specialSecret tracks whether we need to zero an object immediately
// upon freeing.
type specialSecret struct {
	special special
}

// addSecret records the fact that we need to zero p immediately
// when it is freed.
func addSecret(p unsafe.Pointer) {
	lock(&mheap_.speciallock)
	s := (*specialSecret)(mheap_.specialSecretAlloc.alloc())
	s.special.kind = _KindSpecialSecret
	unlock(&mheap_.speciallock)
	addspecial(p, &s.special)
}

// TODO: figure out how to test all of this.

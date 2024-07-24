// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	"unsafe"
)

//go:linkname secret_count runtime/secret.count
func secret_count() uint32 {
	return getg().secret
}

//go:linkname secret_inc runtime/secret.inc
func secret_inc() {
	getg().secret++
}

//go:linkname secret_dec runtime/secret.dec
func secret_dec() {
	getg().secret--
}

//go:linkname secret_eraseSecrets runtime/secret.eraseSecrets
func secret_eraseSecrets() {
	// Clear all the registers (except known safe ones like SP).
	// Q: what does "all" mean in this context? Do we need to
	// worry about just Go-allocatable registers, or do we need
	// to consider any register potentially used by assembly?
	secretEraseRegisters()

	// Figure out the regions of stack we need to zero.
	// Our frame layout is as diagrammed below. (With memclr*
	// the sole function we're going to call.)
	// +----------------------+
	// | secret.Do            |
	// +----------------------+  <- csp (caller stack pointer)
	// | secret_eraseSecrets  |
	// +----------------------+  <- sp (stack pointer)
	// | memclrNoHeapPointers |
	// +----------------------+
	// |                      |
	// | ... stack used by    |
	// | secret function ...  |
	// |                      |
	// +----------------------+  <- lo (bottom of stack)
	sp := getsp()
	csp := getcallersp()
	lo := getg().stack.lo

	// TODO: keep some sort of low water mark so that we don't have
	// to zero a potentially large stack if we used just a little
	// bit of it. That will allow us to use a higher value for
	// lo than gp.stack.lo.

	// We now need to zero all the stack memory below csp, which
	// is the stack pointer of secret.Do.  Unfortunately, we have
	// to avoid a few words that we're not allowed to zero, like
	// wherever a return address is stored.

	// Note: this relies on the fact that memclrNoHeapPointers
	// doesn't use any stack. If it ever did need stack, we would
	// have to wrap this call in systemstack (and then also take
	// into account the stack space needed for the systemstack
	// switch).
	switch GOARCH {
	default:
		// TODO: figure out other archs and where they
		// save their return address.
		memclrNoHeapPointers(unsafe.Pointer(lo), sp-lo)
		memclrNoHeapPointers(unsafe.Pointer(sp), csp-sp)
	case "amd64":
		// There are 2 words we really can't clobber. Those are the
		// return address from this function, and the return
		// address from (the about to be called) memclr. Both
		// of those words will be "cleared" by the call
		// instructions themselves.
		// In addition, we avoid clobbering the saved frame pointer
		// fields. memclr keeps its frame pointer in a register,
		// so there's only one of those to avoid.
		//
		// | secret.Do                    |
		// +------------------------------+  <- csp (caller stack pointer)
		// | secret_eraseSecrets  retaddr |
		// |                      savedFP | (fp of secret.Do)
		// |                              |
		// +------------------------------+  <- sp (stack pointer)
		// | memclrNoHeapPointers retaddr |
		// +------------------------------+
		memclrNoHeapPointers(unsafe.Pointer(lo), (sp-8)-lo)
		memclrNoHeapPointers(unsafe.Pointer(sp), (csp-16)-sp)
	case "arm64":
		// Similarly for arm64, with the complication that
		// the return address is at [sp:sp+8] and the saved frame
		// pointer is at [sp-8:sp] (ugh!).
		// memclrNoHeapPointers keeps both its return address
		// and frame pointer in registers.
		//
		// | secret.Do                    |
		// +------------------------------+  <- csp (caller stack pointer)
		// |                      savedFP | (fp of secret.Do's parent)
		// | secret_eraseSecrets          |
		// |                      retaddr |
		// +------------------------------+  <- sp (stack pointer)
		// |                      savedFP | (fp of secret.Do)
		// | memclrNoHeapPointers         |
		// +------------------------------+
		memclrNoHeapPointers(unsafe.Pointer(lo), (sp-8)-lo)
		memclrNoHeapPointers(unsafe.Pointer(sp+8), (csp-8)-(sp+8))
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

// secret_getStack returns the memory range of the
// current goroutine's stack.
// For testing only.
// Note that this is kind of tricky, as the goroutine can
// be copied and/or exit before the result is used, at which
// point it may no longer be valid.
//go:linkname secret_getStack runtime/secret.getStack
func secret_getStack() (uintptr, uintptr) {
	gp := getg()
	return gp.stack.lo, gp.stack.hi
}

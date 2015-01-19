// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build amd64 amd64p32 386

package runtime

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	"unsafe"
)

// adjust Gobuf as it if executed a call to fn with context ctxt
// and then did an immediate gosave.
func gostartcall(buf *_core.Gobuf, fn, ctxt unsafe.Pointer) {
	sp := buf.Sp
	if _lock.RegSize > _core.PtrSize {
		sp -= _core.PtrSize
		*(*uintptr)(unsafe.Pointer(sp)) = 0
	}
	sp -= _core.PtrSize
	*(*uintptr)(unsafe.Pointer(sp)) = buf.Pc
	buf.Sp = sp
	buf.Pc = uintptr(fn)
	buf.Ctxt = ctxt
}

// Called to rewind context saved during morestack back to beginning of function.
// To help us, the linker emits a jmp back to the beginning right after the
// call to morestack. We just have to decode and apply that jump.
func rewindmorestack(buf *_core.Gobuf) {
	pc := (*[8]byte)(unsafe.Pointer(buf.Pc))
	if pc[0] == 0xe9 { // jmp 4-byte offset
		buf.Pc = buf.Pc + 5 + uintptr(int64(*(*int32)(unsafe.Pointer(&pc[1]))))
		return
	}
	if pc[0] == 0xeb { // jmp 1-byte offset
		buf.Pc = buf.Pc + 2 + uintptr(int64(*(*int8)(unsafe.Pointer(&pc[1]))))
		return
	}
	if pc[0] == 0xcc {
		// This is a breakpoint inserted by gdb.  We could use
		// runtime·findfunc to find the function.  But if we
		// do that, then we will continue execution at the
		// function entry point, and we will not hit the gdb
		// breakpoint.  So for this case we don't change
		// buf.pc, so that when we return we will execute
		// the jump instruction and carry on.  This means that
		// stack unwinding may not work entirely correctly
		// (http://golang.org/issue/5723) but the user is
		// running under gdb anyhow.
		return
	}
	print("runtime: pc=", pc, " ", _core.Hex(pc[0]), " ", _core.Hex(pc[1]), " ", _core.Hex(pc[2]), " ", _core.Hex(pc[3]), " ", _core.Hex(pc[4]), "\n")
	_lock.Gothrow("runtime: misuse of rewindmorestack")
}

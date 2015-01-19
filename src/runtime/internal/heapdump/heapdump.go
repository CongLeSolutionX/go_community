// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Implementation of runtime/debug.WriteHeapDump.  Writes all
// objects in the heap plus additional info (roots, threads,
// finalizers, etc.) to a file.

// The format of the dumped file is described at
// http://golang.org/s/go14heapdump.

package heapdump

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	"unsafe"
)

const (
	FieldKindEol       = 0
	FieldKindPtr       = 1
	FieldKindIface     = 2
	FieldKindEface     = 3
	TagEOF             = 0
	TagObject          = 1
	TagOtherRoot       = 2
	TagType            = 3
	TagGoroutine       = 4
	TagStackFrame      = 5
	TagParams          = 6
	TagFinalizer       = 7
	TagItab            = 8
	TagOSThread        = 9
	TagMemStats        = 10
	TagQueuedFinalizer = 11
	TagData            = 12
	TagBSS             = 13
	TagDefer           = 14
	TagPanic           = 15
	TagMemProf         = 16
	TagAllocSample     = 17
)

var Dumpfd uintptr // fd to write the dump to.

// buffer of pending write data
const (
	bufSize = 4096
)

var Buf [bufSize]byte
var Nbuf uintptr

func Dwrite(data unsafe.Pointer, len uintptr) {
	if len == 0 {
		return
	}
	if Nbuf+len <= bufSize {
		copy(Buf[Nbuf:], (*[bufSize]byte)(data)[:len])
		Nbuf += len
		return
	}

	_core.Write(Dumpfd, (unsafe.Pointer)(&Buf), int32(Nbuf))
	if len >= bufSize {
		_core.Write(Dumpfd, data, int32(len))
		Nbuf = 0
	} else {
		copy(Buf[:], (*[bufSize]byte)(data)[:len])
		Nbuf = len
	}
}

// dump a uint64 in a varint format parseable by encoding/binary
func Dumpint(v uint64) {
	var buf [10]byte
	var n int
	for v >= 0x80 {
		buf[n] = byte(v | 0x80)
		n++
		v >>= 7
	}
	buf[n] = byte(v)
	n++
	Dwrite(unsafe.Pointer(&buf), uintptr(n))
}

// dump varint uint64 length followed by memory contents
func Dumpmemrange(data unsafe.Pointer, len uintptr) {
	Dumpint(uint64(len))
	Dwrite(data, len)
}

func Dumpstr(s string) {
	sp := (*_lock.StringStruct)(unsafe.Pointer(&s))
	Dumpmemrange(sp.Str, uintptr(sp.Len))
}

func dumpotherroot(description string, to unsafe.Pointer) {
	Dumpint(TagOtherRoot)
	Dumpstr(description)
	Dumpint(uint64(uintptr(to)))
}

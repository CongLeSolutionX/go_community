// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin,amd64

// Package macOS provides cgo-less wrappers for Core Foundation and
// Security.framework, similarly to how package syscall provides access to
// libSystem.dylib.
package macOS

import (
	"errors"
	"runtime"
	"unsafe"
)

// CFRef is an opaque reference to a Core Foundation object. It is a pointer,
// but to memory not owned by Go, so not an unsafe.Pointer.
type CFRef uintptr

func CFDataCopyGoBytes(data CFRef) []byte {
	length := CFDataGetLength(data)
	ptr := CFDataGetBytePtr(data)
	src := (*[1 << 20]byte)(unsafe.Pointer(ptr))[:length:length]
	out := make([]byte, length)
	copy(out, src)
	return out
}

const kCFStringEncodingUTF8 = 0x08000100

//go:linkname x509_CFStringCreateWithBytes x509_CFStringCreateWithBytes
//go:cgo_import_dynamic x509_CFStringCreateWithBytes CFStringCreateWithBytes "/System/Library/Frameworks/CoreFoundation.framework/Versions/A/CoreFoundation"
func CFStringCreateWithBytes(s string) CFRef {
	b := []byte(s)
	ret, _ := syscall(funcPC(x509_CFStringCreateWithBytes_trampoline), 0 /* alloc */, uintptr(unsafe.Pointer(&b[0])),
		uintptr(len(s)), uintptr(kCFStringEncodingUTF8), 0 /* isExternalRepresentation */, 0)
	runtime.KeepAlive(b)
	return CFRef(ret)
}
func x509_CFStringCreateWithBytes_trampoline()

//go:linkname x509_CFDictionaryGetValueIfPresent x509_CFDictionaryGetValueIfPresent
//go:cgo_import_dynamic x509_CFDictionaryGetValueIfPresent CFDictionaryGetValueIfPresent "/System/Library/Frameworks/CoreFoundation.framework/Versions/A/CoreFoundation"
func CFDictionaryGetValueIfPresent(dict, key CFRef) (value CFRef, ok bool) {
	ret, _ := syscall(funcPC(x509_CFDictionaryGetValueIfPresent_trampoline), uintptr(dict), uintptr(key),
		uintptr(unsafe.Pointer(&value)), 0, 0, 0)
	if ret == 0 {
		return 0, false
	}
	return value, true
}
func x509_CFDictionaryGetValueIfPresent_trampoline()

const kCFNumberSInt32Type = 3

//go:linkname x509_CFNumberGetValue x509_CFNumberGetValue
//go:cgo_import_dynamic x509_CFNumberGetValue CFNumberGetValue "/System/Library/Frameworks/CoreFoundation.framework/Versions/A/CoreFoundation"
func CFNumberGetValue(num CFRef) (int32, error) {
	var value int32
	ret, _ := syscall(funcPC(x509_CFNumberGetValue_trampoline), uintptr(num), uintptr(kCFNumberSInt32Type),
		uintptr(unsafe.Pointer(&value)), 0, 0, 0)
	if ret == 0 {
		return 0, errors.New("CFNumberGetValue error")
	}
	return value, nil
}
func x509_CFNumberGetValue_trampoline()

//go:linkname x509_CFDataGetLength x509_CFDataGetLength
//go:cgo_import_dynamic x509_CFDataGetLength CFDataGetLength "/System/Library/Frameworks/CoreFoundation.framework/Versions/A/CoreFoundation"
func CFDataGetLength(data CFRef) int {
	ret, _ := syscall(funcPC(x509_CFDataGetLength_trampoline), uintptr(data), 0, 0, 0, 0, 0)
	return int(ret)
}
func x509_CFDataGetLength_trampoline()

//go:linkname x509_CFDataGetBytePtr x509_CFDataGetBytePtr
//go:cgo_import_dynamic x509_CFDataGetBytePtr CFDataGetBytePtr "/System/Library/Frameworks/CoreFoundation.framework/Versions/A/CoreFoundation"
func CFDataGetBytePtr(data CFRef) uintptr {
	ret, _ := syscall(funcPC(x509_CFDataGetBytePtr_trampoline), uintptr(data), 0, 0, 0, 0, 0)
	return ret
}
func x509_CFDataGetBytePtr_trampoline()

//go:linkname x509_CFArrayGetCount x509_CFArrayGetCount
//go:cgo_import_dynamic x509_CFArrayGetCount CFArrayGetCount "/System/Library/Frameworks/CoreFoundation.framework/Versions/A/CoreFoundation"
func CFArrayGetCount(array CFRef) int {
	ret, _ := syscall(funcPC(x509_CFArrayGetCount_trampoline), uintptr(array), 0, 0, 0, 0, 0)
	return int(ret)
}
func x509_CFArrayGetCount_trampoline()

//go:linkname x509_CFArrayGetValueAtIndex x509_CFArrayGetValueAtIndex
//go:cgo_import_dynamic x509_CFArrayGetValueAtIndex CFArrayGetValueAtIndex "/System/Library/Frameworks/CoreFoundation.framework/Versions/A/CoreFoundation"
func CFArrayGetValueAtIndex(array CFRef, index int) CFRef {
	ret, _ := syscall(funcPC(x509_CFArrayGetValueAtIndex_trampoline), uintptr(array), uintptr(index), 0, 0, 0, 0)
	return CFRef(ret)
}
func x509_CFArrayGetValueAtIndex_trampoline()

//go:linkname x509_CFEqual x509_CFEqual
//go:cgo_import_dynamic x509_CFEqual CFEqual "/System/Library/Frameworks/CoreFoundation.framework/Versions/A/CoreFoundation"
func CFEqual(a, b CFRef) bool {
	ret, _ := syscall(funcPC(x509_CFEqual_trampoline), uintptr(a), uintptr(b), 0, 0, 0, 0)
	return ret == 1
}
func x509_CFEqual_trampoline()

//go:linkname x509_CFRelease x509_CFRelease
//go:cgo_import_dynamic x509_CFRelease CFRelease "/System/Library/Frameworks/CoreFoundation.framework/Versions/A/CoreFoundation"
func CFRelease(ref CFRef) {
	syscall(funcPC(x509_CFRelease_trampoline), uintptr(ref), 0, 0, 0, 0, 0)
}
func x509_CFRelease_trampoline()

// syscall is implemented in the runtime package (runtime/sys_darwin.go)
func syscall(fn, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr)

// funcPC returns the entry point for f. See comments in runtime/proc.go
// for the function of the same name.
//go:nosplit
func funcPC(f func()) uintptr {
	return **(**uintptr)(unsafe.Pointer(&f))
}

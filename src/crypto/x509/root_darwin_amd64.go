// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package x509

import (
	"runtime"
	"unsafe"
)

func loadSystemRoots() (*CertPool, error) {
	d := CFDataCreate("6 x 9 = 42")
	println(CFDataGetLength(d))
	CFRelease(d)

	return NewCertPool(), nil
}

// CFRef is an opaque reference to a Core Foundation object. It is a pointer,
// but from memory not owned by Go, so not an unsafe.Pointer.
type CFRef uintptr

func CFDataCreate(s string) CFRef {
	b := []byte(s)
	ret, _ := syscall(funcPC(x509_CFDataCreate_trampoline), 0, uintptr(unsafe.Pointer(&b[0])), uintptr(len(s)), 0, 0, 0)
	runtime.KeepAlive(b)
	return CFRef(ret)
}
func x509_CFDataCreate_trampoline()

//go:linkname x509_CFDataCreate x509_CFDataCreate
//go:cgo_import_dynamic x509_CFDataCreate CFDataCreate "/System/Library/Frameworks/CoreFoundation.framework/Versions/A/CoreFoundation"

func CFDataGetLength(data CFRef) int {
	ret, _ := syscall(funcPC(x509_CFDataGetLength_trampoline), uintptr(data), 0, 0, 0, 0, 0)
	return int(ret)
}
func x509_CFDataGetLength_trampoline()

//go:linkname x509_CFDataGetLength x509_CFDataGetLength
//go:cgo_import_dynamic x509_CFDataGetLength CFDataGetLength "/System/Library/Frameworks/CoreFoundation.framework/Versions/A/CoreFoundation"

func CFRelease(ref CFRef) {
	syscall(funcPC(x509_CFRelease_trampoline), uintptr(ref), 0, 0, 0, 0, 0)
}
func x509_CFRelease_trampoline()

//go:linkname x509_CFRelease x509_CFRelease
//go:cgo_import_dynamic x509_CFRelease CFRelease "/System/Library/Frameworks/CoreFoundation.framework/Versions/A/CoreFoundation"

// syscall is implemented in the runtime package (runtime/sys_darwin.go)
func syscall(fn, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr)

// funcPC returns the entry point for f. See comments in runtime/proc.go
// for the function of the same name.
//go:nosplit
func funcPC(f func()) uintptr {
	return **(**uintptr)(unsafe.Pointer(&f))
}

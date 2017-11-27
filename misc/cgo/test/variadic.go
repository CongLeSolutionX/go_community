// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cgotest

// #include <stdio.h>
// #include <stdlib.h>
import "C"
import (
	"testing"
	"unsafe"
)

func testVariadic(t *testing.T) {
	buf1 := make([]byte, 10)
	format1 := C.CString("1 + 2 = %d")
	defer C.free(unsafe.Pointer(format1))
	C.sprintf((*C.char)(unsafe.Pointer(&buf1[0])), format1, 3)
	if got, want := string(buf1[:9]), "1 + 2 = 3"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	buf2 := make([]byte, 10)
	format2 := C.CString("4 + 5 = %d")
	defer C.free(unsafe.Pointer(format2))
	C.sprintf((*C.char)(unsafe.Pointer(&buf2[0])), format2, 9)
	if got, want := string(buf2[:9]), "4 + 5 = 9"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	buf3 := make([]byte, 10)
	format3 := C.CString("%ld")
	defer C.free(unsafe.Pointer(format3))
	C.sprintf((*C.char)(unsafe.Pointer(&buf3[0])), format3, C.long(2))
	if got, want := string(buf3[:1]), "2"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

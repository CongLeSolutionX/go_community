// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build cgo,!netgo
// +build darwin linux

package socket

/*
#include <sys/types.h>
#include <sys/socket.h>

#include <netdb.h>
*/
import "C"

import "unsafe"

func getnameinfoPTR(b []byte, sa *C.struct_sockaddr, salen C.socklen_t) error {
	errno, _ := C.getnameinfo(sa, salen, (*C.char)(unsafe.Pointer(&b[0])), C.socklen_t(len(b)), nil, 0, C.NI_NAMEREQD)
	if errno != 0 {
		return AddrinfoErrno(errno)
	}
	return nil
}

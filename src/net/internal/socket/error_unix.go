// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build cgo,!netgo
// +build !nacl,!plan9,!solaris,!windows

package socket

/*
#include <sys/types.h>
#include <sys/socket.h>

#include <netdb.h>
*/
import "C"

// An AddrinfoErrno represents a getaddrinfo, getnameinfo-specific
// error number. It's a signed number and a zero value is a non-error
// by convention.
type AddrinfoErrno int

func (eai AddrinfoErrno) Error() string {
	return C.GoString(C.gai_strerror(C.int(eai)))
}

func (eai AddrinfoErrno) Temporary() bool {
	return eai == C.EAI_AGAIN
}

func (eai AddrinfoErrno) Timeout() bool {
	return false
}

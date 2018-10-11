// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !aix,!darwin,!dragonfly,!freebsd,!linux,!netbsd,!openbsd,!solaris

package net

import (
	"errors"
	"os"
	"runtime"
)

func newSyscallConn(network string, f *os.File) (*SyscallConn, error) {
	return nil, errors.New("not implemented on %s " + runtime.GOOS)
}

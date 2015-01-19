// Copyright 2012 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package lock

import (
	_core "runtime/internal/core"
)

const (
	SIG_DFL uintptr = 0
	SIG_IGN uintptr = 1
)

func Crash() {
	if GOOS == "darwin" {
		// OS X core dumps are linear dumps of the mapped memory,
		// from the first virtual byte to the last, with zeros in the gaps.
		// Because of the way we arrange the address space on 64-bit systems,
		// this means the OS X core file will be >128 GB and even on a zippy
		// workstation can take OS X well over an hour to write (uninterruptible).
		// Save users from making that mistake.
		if _core.PtrSize == 8 {
			return
		}
	}

	unblocksignals()
	Setsig(SIGABRT, SIG_DFL, false)
	Raise(SIGABRT)
}

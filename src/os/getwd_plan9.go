// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build plan9

package os

import "syscall"

// Use syscall.Getwd directly on plan9 for reasons laid out in CL 89575.
func getwd() (dir string, err error) {
	dir, err = syscall.Getwd()
	return dir, NewSyscallError("getwd", err)
}

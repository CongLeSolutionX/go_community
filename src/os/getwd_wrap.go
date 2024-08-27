// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build windows || plan9

package os

import "syscall"

func getwd() (dir string, err error) {
	dir, err = syscall.Getwd()
	return dir, NewSyscallError("getwd", err)
}

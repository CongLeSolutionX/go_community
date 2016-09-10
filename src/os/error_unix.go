// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

package os

import "syscall"

func isExist(err error) bool {
	if err == nil {
		return false
	}
	err = convertError(err)
	return err == syscall.EEXIST || err == syscall.ENOTEMPTY || err == ErrExist
}

func isNotExist(err error) bool {
	if err == nil {
		return false
	}
	err = convertError(err)
	return err == syscall.ENOENT || err == ErrNotExist
}

func isPermission(err error) bool {
	if err == nil {
		return false
	}
	err = convertError(err)
	return err == syscall.EACCES || err == syscall.EPERM || err == ErrPermission
}

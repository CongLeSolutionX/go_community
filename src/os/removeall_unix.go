// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build unix

package os

import (
	"internal/syscall/unix"
	"syscall"
)

func unlinkat(fd int, name string) error {
	return ignoringEINTR(func() error {
		return unix.Unlinkat(fd, name, 0)
	})
}

func rmdirat(fd int, name string) error {
	return ignoringEINTR(func() error {
		return unix.Unlinkat(fd, name, unix.AT_REMOVEDIR)
	})
}

// openDirAt opens a directory name relative to the directory referred to by
// the file descriptor dirfd. If name is anything but a directory (this
// includes a symlink to one), it should return an error. Other than that this
// should act like openFileNolog.
//
// This acts like openFileNolog rather than OpenFile because
// we are going to (try to) remove the file.
// The contents of this file are not relevant for test caching.
func openDirAt(dirfd int, name string) (*File, error) {
	r, err := ignoringEINTR2(func() (int, error) {
		return unix.Openat(dirfd, name, O_RDONLY|syscall.O_CLOEXEC|syscall.O_DIRECTORY|syscall.O_NOFOLLOW, 0)
	})
	if err != nil {
		return nil, err
	}

	if !supportsCloseOnExec {
		syscall.CloseOnExec(r)
	}

	// We use kindNoPoll because we know that this is a directory.
	return newFile(r, name, kindNoPoll, false), nil
}

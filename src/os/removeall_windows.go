// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build windows

package os

import (
	"internal/syscall/windows"
	"syscall"
)

func unlinkat(fd sysfdType, name string) error {
	return removeat(fd, name)
}

func rmdirat(fd sysfdType, name string) error {
	return removeat(fd, name)
}

// openDirAt opens a directory name relative to the directory referred to by
// the file descriptor dirfd. If name is anything but a directory (this
// includes a symlink to one), it should return an error. Other than that this
// should act like openFileNolog.
//
// This acts like openFileNolog rather than OpenFile because
// we are going to (try to) remove the file.
// The contents of this file are not relevant for test caching.
func openDirAt(fd sysfdType, name string) (*File, error) {
	h, err := openat(fd, name, syscall.O_RDONLY|syscall.O_CLOEXEC|windows.O_DIRECTORY, 0)
	if err != nil {
		return nil, err
	}
	return newFile(h, name, "file"), nil
}

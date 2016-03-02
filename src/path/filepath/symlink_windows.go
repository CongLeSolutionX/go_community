// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package filepath

import (
	"os"
	"syscall"

	"internal/syscall/windows"
)

func open(path string) (fd syscall.Handle, err error) {
	if len(path) == 0 {
		return syscall.InvalidHandle, syscall.ERROR_FILE_NOT_FOUND
	}
	pathp, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return syscall.InvalidHandle, err
	}

	return syscall.CreateFile(pathp, syscall.GENERIC_READ, syscall.FILE_SHARE_READ, nil, syscall.OPEN_EXISTING, syscall.FILE_FLAG_BACKUP_SEMANTICS, 0)
}

func evalSymlinks(path string) (string, error) {
	fd, err := open(path)
	if err != nil {
		return "", err
	}

	abs, err := windows.FinalPathByHandle(fd)
	if err != nil {
		return "", err
	}

	if IsAbs(path) {
		return abs, nil
	}

	if isUNC(path) {
		return abs, nil
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	if volume := VolumeName(path); volume != "" {
		if volume == VolumeName(wd) {
			rel, err := Rel(wd, abs)
			return volume + rel, err
		}

		volume = VolumeName(abs)

		return volume + abs[len(volume)+1:], nil // trim beginning \
	}

	return Rel(wd, abs)
}

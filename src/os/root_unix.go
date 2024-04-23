// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build aix || darwin || dragonfly || freebsd || (js && wasm) || wasip1 || linux || netbsd || openbsd || solaris

package os

import (
	"internal/syscall/unix"
	"syscall"
)

type sysfdType = int

// rootOpenRootNolog is Root.OpenRoot.
func rootOpenRootNolog(root *Root, name string) (*Root, error) {
	fd, err := doInRoot(root, name, rootOpenDir)
	if err != nil {
		return nil, err
	}
	f := NewFile(uintptr(fd), root.Name()+string(PathSeparator)+name)
	return &Root{f: f}, nil
}

// rootOpenFileNolog is Root.OpenFile.
func rootOpenFileNolog(root *Root, name string, flag int, perm FileMode) (*File, error) {
	fd, err := doInRoot(root, name, func(parent int, name string) (fd int, err error) {
		ignoringEINTR(func() error {
			fd, err = unix.Openat(parent, name, syscall.O_NOFOLLOW|syscall.O_CLOEXEC|flag, uint32(perm))
			if err == syscall.ELOOP {
				err = eloopErr(parent, name)
			}
			return err
		})
		return fd, err
	})
	if err != nil {
		return nil, err
	}
	f := newFile(fd, root.Name()+string(PathSeparator)+name, kindOpenFile, unix.HasNonblockFlag(flag))
	return f, nil
}

func rootOpenDir(parent int, name string) (int, error) {
	var (
		fd  int
		err error
	)
	ignoringEINTR(func() error {
		fd, err = unix.Openat(parent, name, syscall.O_NOFOLLOW|syscall.O_CLOEXEC|syscall.O_RDONLY, 0)
		if err == syscall.ELOOP {
			err = eloopErr(parent, name)
		}
		return err
	})
	return fd, err
}

func mkdirat(fd int, name string, perm FileMode) error {
	return ignoringEINTR(func() error {
		return unix.Mkdirat(fd, name, syscallMode(perm))
	})
}

// eloopErr resolves the symlink name in parent,
// and returns errSymlink with the link contents.
func eloopErr(parent int, name string) error {
	link, err := readlinkat(parent, name)
	if err != nil {
		return syscall.ELOOP
	}
	return errSymlink(link)
}

func readlinkat(fd int, name string) (string, error) {
	for len := 128; ; len *= 2 {
		b := make([]byte, len)
		var (
			n int
			e error
		)
		ignoringEINTR(func() error {
			n, e = unix.Readlinkat(fd, name, b)
			return e
		})
		if e == syscall.ERANGE {
			continue
		}
		if e != nil {
			return "", &PathError{Op: "readlinkat", Path: name, Err: e}
		}
		if n < 0 {
			n = 0
		}
		if n < len {
			return string(b[0:n]), nil
		}
	}
}

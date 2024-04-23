// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || wasip1

package os

import (
	"errors"
	"internal/syscall/unix"
	"runtime"
	"syscall"
)

type sysfdType = int

func openRoot(name string) (*Root, error) {
	var fd int
	err := ignoringEINTR(func() error {
		var err error
		fd, _, err = open(name, syscall.O_CLOEXEC, 0)
		return err
	})
	if err != nil {
		return nil, &PathError{Op: "open", Path: name, Err: err}
	}
	return newRoot(fd, name)
}

func newRoot(fd int, name string) (*Root, error) {
	var st syscall.Stat_t
	err := ignoringEINTR(func() error {
		return syscall.Fstat(fd, &st)
	})
	if err == nil && st.Mode&syscall.S_IFDIR == 0 {
		syscall.Close(fd)
		return nil, &PathError{Op: "open", Path: name, Err: errors.New("not a directory")}
	}

	// There's a race here with fork/exec, which we are
	// content to live with. See ../syscall/exec_unix.go.
	if !supportsCloseOnExec {
		syscall.CloseOnExec(fd)
	}

	r := &Root{root{
		fd:   fd,
		name: name,
	}}
	runtime.SetFinalizer(&r.root, (*root).Close)
	return r, nil
}

// openRootInRoot is Root.OpenRoot.
func openRootInRoot(r *Root, name string) (*Root, error) {
	fd, err := doInRoot(r, name, func(parent int, name string) (fd int, err error) {
		ignoringEINTR(func() error {
			fd, err = unix.Openat(parent, name, syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0)
			if err == syscall.ELOOP {
				err = eloopErr(parent, name)
			}
			return err
		})
		return fd, err
	})
	if err != nil {
		return nil, &PathError{Op: "openat", Path: name, Err: err}
	}
	return newRoot(fd, name)
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
		return nil, &PathError{Op: "openat", Path: name, Err: err}
	}
	f := newFile(fd, joinPath(root.Name(), name), kindOpenFile, unix.HasNonblockFlag(flag))
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
	err := ignoringEINTR(func() error {
		return unix.Mkdirat(fd, name, syscallMode(perm))
	})
	if err != nil {
		return &PathError{Op: "mkdirat", Path: name, Err: err}
	}
	return nil
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
			return "", e
		}
		if n < 0 {
			n = 0
		}
		if n < len {
			return string(b[0:n]), nil
		}
	}
}

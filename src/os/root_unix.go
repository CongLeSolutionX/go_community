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

// openRootNolog is OpenRoot.
func openRootNolog(name string) (*Root, error) {
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

// newRoot returns a new Root.
// If fd is not a directory, it closes it and returns an error.
func newRoot(fd int, name string) (*Root, error) {
	var fs fileStat
	err := ignoringEINTR(func() error {
		return syscall.Fstat(fd, &fs.sys)
	})
	fillFileStatFromSys(&fs, name)
	if err == nil && !fs.IsDir() {
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
				err = eloopErr(parent, name, err)
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
			if err == syscall.ELOOP || err == syscall.ENOTDIR {
				err = eloopErr(parent, name, err)
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
		fd, err = unix.Openat(parent, name, syscall.O_NOFOLLOW|syscall.O_CLOEXEC|syscall.O_DIRECTORY, 0)
		if err == syscall.ELOOP || err == syscall.ENOTDIR {
			err = eloopErr(parent, name, err)
		} else if err == syscall.ENOTSUP || err == syscall.EOPNOTSUPP {
			// ENOTSUP and EOPNOTSUPP are often, but not always, the same errno.
			// Translate both to ENOTDIR, since this indicates a non-terminal
			// path component was not a directory.
			err = syscall.ENOTDIR
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
//
// If name is not a symlink, return origError.
func eloopErr(parent int, name string, origError error) error {
	link, err := readlinkat(parent, name)
	if err != nil {
		return origError
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

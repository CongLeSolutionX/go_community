// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package lockedfile creates and manipulates files whose contents should only
// change atomically.
package lockedfile

import (
	"os"
	"sort"
	"sync"

	"cmd/go/internal/base"
)

// A File is a locked *os.File.
//
// Closing the file releases the lock.
//
// If the program exits while a file is locked, the operating system releases
// the lock but may not do so promptly: callers must ensure that all locked
// files are closed before exiting.
type File struct {
	osFile
	closed bool
}

// osFile embeds a *os.File while keeping the pointer itself unexported.
// (When we close a File, it must be the same file descriptor that we opened!)
type osFile struct {
	*os.File
}

var openFiles struct {
	sync.Mutex
	m map[*File]struct{}
}

func init() {
	openFiles.m = make(map[*File]struct{})

	// Per https://docs.microsoft.com/en-us/windows/desktop/api/fileapi/nf-fileapi-lockfileex:
	//
	// “If a process terminates with a portion of a file locked or closes a file
	// that has outstanding locks, the locks are unlocked by the operating system.
	// However, the time it takes for the operating system to unlock these locks
	// depends upon available system resources.”
	//
	// That makes it important that we actually Close locked files, and don't just
	// allow them to leak. Use a base.AtExit hook to report violations.
	base.AtExit(func() {
		openFiles.Lock()
		defer openFiles.Unlock()

		if len(openFiles.m) > 0 {
			names := make([]string, 0, len(openFiles.m))
			for f := range openFiles.m {
				names = append(names, f.Name())
			}
			sort.Strings(names)
			base.Errorf("exiting with files still locked: %s", names)
		}
	})
}

// OpenFile is like os.OpenFile, but returns a locked file.
// If flag includes os.O_WRONLY or os.O_RDWR, the file is write-locked;
// otherwise, it is read-locked.
func OpenFile(name string, flag int, perm os.FileMode) (_ *File, err error) {
	f := new(File)
	f.osFile.File, err = openFile(name, flag, perm)
	if err != nil {
		return nil, err
	}

	openFiles.Lock()
	openFiles.m[f] = struct{}{}
	openFiles.Unlock()

	return f, nil
}

// Open is like os.Open, but returns a read-locked file.
func Open(name string) (*File, error) {
	return OpenFile(name, os.O_RDONLY, 0)
}

// Create is like os.Create, but returns a write-locked file.
func Create(name string) (*File, error) {
	return OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
}

// Edit creates the named file with mode 0666 (before umask),
// but does not truncate existing contents.
//
// If successful, methods on the returned File can be used for I/O; the
// associated file descriptor has mode O_RDWR and the file is write-locked.
func Edit(name string) (*File, error) {
	return OpenFile(name, os.O_RDWR|os.O_CREATE, 0666)
}

// Close unlocks and closes the underlying file.
//
// Close may be called multiple times; all calls after the first will return a
// non-nil error.
func (f *File) Close() error {
	if f.closed {
		return &os.PathError{
			Op:   "close",
			Path: f.Name(),
			Err:  os.ErrClosed,
		}
	}
	f.closed = true

	err := closeFile(f.osFile.File)

	openFiles.Lock()
	delete(openFiles.m, f)
	openFiles.Unlock()

	return err
}

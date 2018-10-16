// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package filelock provides a platform-independent API for advisory file
// locking. Calls to functions in this package on platforms that do not support
// advisory locks will return errors for which IsNotSupported returns true.
//
// ⚠ This package provides a very low-level API. Prefer to use lockedfile
// instead when possible.
package filelock

import (
	"errors"
	"os"
)

// File is the subset of os.File required for file locking.
// File implementations must be usable as map keys.
type File interface {
	Name() string
	Fd() uintptr
	Stat() (os.FileInfo, error)
}

// Lock places an advisory write lock on the file, blocking until it can be
// locked.
//
// If Lock returns nil, no other process will be able to place a read or write
// lock on the file until this process exits, closes f, or calls Unlock on it.
//
// If f is already read- or write-locked, the behavior of Lock is unspecified.
//
// Closing the file may or may not release the lock promptly. Callers should
// ensure that Unlock is always called if Lock succeeds.
func Lock(f File) error {
	return lock(f, writeLock)
}

// RLock places an advisory read lock on the file, blocking until it can be locked.
//
// If RLock returns nil, no other process will be able to place a write lock on
// the file until this process exits, closes f, or calls Unlock on it.
//
// If f is already read- or write-locked, the behavior of RLock is unspecified.
//
// Closing the file may or may not release the lock promptly. Callers should
// ensure that Unlock is always called if RLock succeeds.
func RLock(f File) error {
	return lock(f, readLock)
}

// Unlock removes an advisory lock placed on f by this process.
//
// It is safe to unlock a file that has not been locked, but the error returned
// in that case (if any) is unspecified.
func Unlock(f File) error {
	return unlock(f)
}

// String implements fmt.Stringer for a platform-specific lockType type.
func (lt lockType) String() string {
	switch lt {
	case readLock:
		return "RLock"
	case writeLock:
		return "Lock"
	default:
		return "Unlock"
	}
}

// IsNotSupported returns a boolean indicating whether the error is known to
// report that a function is not supported (possibly for a specific input).
// It is satisfied by ErrNotSupported as well as some syscall errors.
func IsNotSupported(err error) bool {
	return isNotSupported(underlyingError(err))
}

// IsLocked returns a boolean indicating whether the error is known to report
// that a file has a conflicting lock. It is satisfied by ErrLocked as well as
// some syscall errors.
func IsLocked(err error) bool {
	return isLocked(underlyingError(err))
}

var (
	ErrNotSupported = errors.New("operation not supported")
	ErrLocked       = errors.New("file is locked")
)

// underlyingError returns the underlying error for known os error types.
func underlyingError(err error) error {
	switch err := err.(type) {
	case *os.PathError:
		return err.Err
	case *os.LinkError:
		return err.Err
	case *os.SyscallError:
		return err.Err
	}
	return err
}

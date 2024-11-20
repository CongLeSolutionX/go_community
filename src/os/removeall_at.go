// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build unix || windows

package os

import (
	"io"
	"runtime"
	"syscall"
)

func removeAll(path string) error {
	if path == "" {
		// fail silently to retain compatibility with previous behavior
		// of RemoveAll. See issue 28830.
		return nil
	}

	// The rmdir system call does not permit removing ".",
	// so we don't permit it either.
	if endsWithDot(path) {
		return &PathError{Op: "RemoveAll", Path: path, Err: syscall.EINVAL}
	}

	// Simple case: if Remove works, we're done.
	err := Remove(path)
	if err == nil || IsNotExist(err) {
		return nil
	}

	dir, err := Open(path)
	if IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	err = removeAllFrom(dir)
	dir.Close()
	if err != nil {
		return prependErrorPathPrefix(err, path)
	}
	if err := Remove(path); err != nil && !IsNotExist(err) {
		return err
	}
	return nil
}

func prependErrorPathPrefix(err error, prefix string) error {
	pe, ok := err.(*PathError)
	if !ok {
		return err
	}
	if pe.Path == "." {
		pe.Path = prefix
	} else if len(prefix) > 0 && !IsPathSeparator(prefix[len(prefix)-1]) {
		pe.Path = prefix + string(PathSeparator) + pe.Path
	} else {
		pe.Path = prefix + pe.Path
	}
	return pe
}

// removeAllFromChild recursively removes the file name in dirfd.
func removeAllFromChild(dirfd sysfdType, name string) error {
	// If we can unlink the file, then we're done.
	unlinkErr := unlinkat(dirfd, name)
	if unlinkErr == nil || IsNotExist(unlinkErr) {
		return nil
	}

	// Possibly this is a directory that we can recurse into.
	// Try to open it.
	child, err := openDirAt(dirfd, name)
	if err != nil {
		if IsNotExist(err) {
			return nil
		}
		if runtime.GOOS != "windows" {
			// On Unix, unlink and rmdir are different operations.
			// We've failed to open this file as a directory,
			// but possibly it's an empty directory that we don't
			// have permission to open. Try rmdir.
			if err := rmdirat(dirfd, name); err == nil {
				return nil
			}
		}
		return &PathError{Op: "RemoveAll", Path: name, Err: unlinkErr}
	}

	err = removeAllFrom(child)
	child.Close()
	if err != nil {
		return prependErrorPathPrefix(err, name)
	}

	err = rmdirat(dirfd, name)
	if err != nil && !IsNotExist(err) {
		return &PathError{Op: "RemoveAll", Path: name, Err: err}
	}
	return nil
}

// removeAllFrom recursively removes the contents of dir.
func removeAllFrom(dir *File) error {
	dirfd := (sysfdType)(dir.Fd())

	var firstError error
	for {
		const reqSize = 1024
		names, err := dir.Readdirnames(reqSize)
		if err != nil && err != io.EOF {
			// Don't return this error.
			// The caller will try to delete the directory we just failed to list,
			// and either succeed or fail; if it fails, we use the error from
			// deleting the directory rather than this one.
			break
		}

		success := false
		errs := 0
		for _, name := range names {
			err := removeAllFromChild(dirfd, name)
			if err != nil {
				errs++
				if firstError == nil {
					firstError = err
				}
			} else {
				success = true
			}
		}

		if len(names) < reqSize {
			break
		}

		// Removing files from the directory may have caused
		// the OS to reshuffle it. Simply calling Readdirnames
		// again may skip some entries. The only reliable way
		// to avoid this is to close and re-open the
		// directory. See issue 20841.
		//
		// If we didn't delete anything, keep iterating from the current position.
		if success {
			if info := dir.dirinfo.Swap(nil); info != nil {
				info.close()
			}
			dir.Seek(0, io.SeekStart)
		}
	}

	return firstError
}

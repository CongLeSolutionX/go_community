// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os

import (
	"internal/filepathlite"
	"io"
	"syscall"
)

// RemoveAll removes path and any children it contains.
// It removes everything it can but returns the first error
// it encounters. If the path does not exist, RemoveAll
// returns nil (no error).
// If there is an error, it will be of type [*PathError].
func RemoveAll(path string) error {
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

	// RemoveAll recurses by deleting the path base from its parent directory
	parentDir, base := filepathlite.Split(path)
	for len(parentDir) > 0 && IsPathSeparator(parentDir[len(parentDir)-1]) {
		parentDir = parentDir[:len(parentDir)-1]
	}

	parent, err := OpenRoot(parentDir)
	if IsNotExist(err) {
		// If parent does not exist, base cannot exist. Fail silently.
		return nil
	}
	if err != nil {
		return err
	}
	defer parent.Close()

	if err := removeAllFrom(parent, base); err != nil {
		return prependErrorPathPrefix(err, parentDir)
	}
	return nil
}

// endsWithDot reports whether the final component of path is ".".
func endsWithDot(path string) bool {
	if path == "." {
		return true
	}
	if len(path) >= 2 && path[len(path)-1] == '.' && IsPathSeparator(path[len(path)-2]) {
		return true
	}
	return false
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

func removeAllFrom(parent *Root, base string) error {
	// Simple case: if remove works, we're done.
	removeErr := parent.Remove(base)
	if removeErr == nil || IsNotExist(removeErr) {
		return nil
	}

	// This might be a directory we need to recurse into.
	// Try opening it to see.
	child, err := parent.OpenRoot(base)
	if err != nil {
		if IsNotExist(err) {
			return nil
		}
		return err
	}
	defer child.Close()

	var firstError error
	var file *File
	defer func() {
		if file != nil {
			file.Close()
		}
	}()
	for {
		if file == nil {
			file, err = child.Open(".")
			if err != nil {
				return &PathError{Op: "readdirnames", Path: base, Err: underlyingError(err)}
			}
		}
		const reqSize = 1024
		names, err := file.Readdirnames(reqSize)
		if err != nil && err != io.EOF {
			return &PathError{Op: "readdirnames", Path: base, Err: underlyingError(err)}
		}
		numErr := 0
		for _, name := range names {
			err := removeAllFrom(child, name)
			if err != nil {
				numErr++
				if firstError == nil {
					firstError = prependErrorPathPrefix(err, base)
				}
			}
		}

		// If we can delete any entry, then close the file
		// and reopen it on the next iteration.
		// Removing files from the directory may have caused
		// the OS to reshuffle it. Simply calling Readdirnames
		// again may skip some entries. The only reliable way
		// to avoid this is to close and re-open the
		// directory. See issue 20841.
		if numErr != len(names) {
			file.Close()
			file = nil
		}

		// Finish when the end of the directory is reached
		if len(names) < reqSize {
			break
		}
	}
	file.Close()
	file = nil

	err = parent.Remove(base)
	if firstError != nil {
		return firstError
	}
	if err != nil && !IsNotExist(err) {
		return err
	}
	return nil
}

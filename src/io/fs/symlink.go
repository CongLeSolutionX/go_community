// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fs

import "errors"

// SymlinkFS is the interface implemented by a file system
// that supports symbolic links.
type SymlinkFS interface {
	FS

	// ReadLink returns the destination of the named symbolic link.
	// Link destinations will always be slash-separated paths relative to
	// the link's directory. The link destination is guaranteed to be
	// a path inside FS.
	// If there is an error, it should be of type *PathError.
	ReadLink(name string) (string, error)

	// Lstat returns a FileInfo describing the file without following any symbolic links.
	// If there is an error, it should be of type *PathError.
	Lstat(name string) (FileInfo, error)
}

// ReadLink returns the destination of the named symbolic link.
//
// If fsys does not implement [SymlinkFS], then ReadLink returns an error.
func ReadLink(fsys FS, name string) (string, error) {
	sym, ok := fsys.(SymlinkFS)
	if !ok {
		return "", &PathError{Op: "readlink", Path: name, Err: errors.New("not implemented")}
	}
	return sym.ReadLink(name)
}

// Lstat returns a FileInfo describing the file without following any symbolic links.
//
// If fsys does not implement [SymlinkFS], then Lstat is identical to [Stat].
func Lstat(fsys FS, name string) (FileInfo, error) {
	sym, ok := fsys.(SymlinkFS)
	if !ok {
		return Stat(fsys, name)
	}
	return sym.Lstat(name)
}

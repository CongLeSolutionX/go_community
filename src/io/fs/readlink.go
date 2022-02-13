// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fs

import "errors"

// ReadLinkFS is the interface implemented by a file system
// that supports symbolic links.
type ReadLinkFS interface {
	FS

	// ReadLink returns the destination of the named symbolic link.
	// Link destinations will always be slash-separated paths relative to
	// the link's directory. The link destination is guaranteed to be
	// a path inside FS.
	ReadLink(name string) (string, error)
}

// ReadLink returns the destination of the named symbolic link.
//
// If fsys does not implement ReadLinkFS, then ReadLink returns an error.
func ReadLink(fsys FS, name string) (string, error) {
	rl, ok := fsys.(ReadLinkFS)
	if !ok {
		return "", &PathError{Op: "readlink", Path: name, Err: errors.New("not implemented")}
	}
	return rl.ReadLink(name)
}

// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os

import (
	"errors"
	"internal/testlog"
	"runtime"
)

// Root represents a directory.
//
// Methods on Root can only access files and directories within that directory.
// If any component of a file name passed to a method of Root references a location
// outside the root, the method returns an error.
// File names may reference the directory itself (.).
// On Windows, file names may not reference reserved device names such as COM1.
//
// File names may contain symbolic links, but symbolic links may not
// reference a location outside the root.
// Symbolic links must not be absolute.
//
// When GOOS=js, Root is vulnerable to TOCTOU (time-of-check-time-of-use) attacks
// in symlink validation, and cannot ensure that operations will not escape the root.
//
// Methods on Root do not prohibit traversal of filesystem boundaries,
// Linux bind mounts, /proc special files, or access to Unix device files.
//
// Methods on Root are safe to be used from multiple goroutines simultaneously.
//
// On most platforms, creating a Root opens a file descriptor or handle referencing
// the directory. If the directory is moved, methods on Root reference the original
// directory. When GOOS is plan9 or js, a Root instead references the path name of a
// directory and does not track renames.
type Root struct {
	root root
}

const (
	// Maximum number of symbolic links we will follow when resolving a file in a root.
	// 8 is __POSIX_SYMLOOP_MAX (the minimum allowed value for SYMLOOP_MAX),
	// and a commonly limit.
	rootMaxSymlinks = 8
)

// OpenRoot opens the named directory.
// If there is an error, it will be of type *PathError.
func OpenRoot(name string) (*Root, error) {
	return openRoot(name)
}

// Name returns the name of the directory presented to OpenRoot.
//
// It is safe to call Name after [Close].
func (r *Root) Name() string {
	return r.root.Name()
}

// Close closes the Root.
// After Close is called, methods on Root return errors.
func (r *Root) Close() error {
	return r.root.Close()
}

// Open opens the named file in the root for reading.
// See [Open] for more details.
func (r *Root) Open(name string) (*File, error) {
	return r.OpenFile(name, O_RDONLY, 0)
}

// Create creates or truncates the named file in the root.
// See [Create] for more details.
func (r *Root) Create(name string) (*File, error) {
	return r.OpenFile(name, O_RDWR|O_CREATE|O_TRUNC, 0666)
}

// OpenFile opens the named file in the root.
// See [OpenFile] for more details.
//
// If perm contains bits other than the nine least-significant bits (0o777),
// OpenFile returns an error.
func (r *Root) OpenFile(name string, flag int, perm FileMode) (*File, error) {
	if perm&0o777 != perm {
		return nil, &PathError{Op: "openat", Path: name, Err: errors.New("unsupported file mode")}
	}
	r.logOpen(name)
	rf, err := rootOpenFileNolog(r, name, flag, perm)
	if err != nil {
		return nil, err
	}
	rf.appendMode = flag&O_APPEND != 0
	return rf, nil
}

// OpenDir opens the named directory in the root.
// If there is an error, it will be of type *PathError.
func (r *Root) OpenRoot(name string) (*Root, error) {
	r.logOpen(name)
	return openRootInRoot(r, name)
}

// Mkdir creates a new directory in the root
// with the specified name and permission bits (before umask).
// See [Mkdir] for more details.
//
// If perm contains bits other than the nine least-significant bits (0o777),
// OpenFile returns an error.
func (r *Root) Mkdir(name string, perm FileMode) error {
	if perm&0o777 != perm {
		return &PathError{Op: "mkdirat", Path: name, Err: errors.New("unsupported file mode")}
	}
	return rootMkdir(r, name, perm)
}

func (r *Root) logOpen(name string) {
	if log := testlog.Logger(); log != nil {
		// This won't be right if r's name has changed since it was opened,
		// but it's the best we can do.
		log.Open(joinPath(r.Name(), name))
	}
}

// splitPathInRoot splits a path into components
// and joins it with the given prefix and suffix.
//
// The path is relative to a Root, and must not be
// absolute, volume-relative, or "".
//
// "." components are removed, except in the last component.
//
// Path separators following the last component are preserved.
func splitPathInRoot(s string, prefix, suffix []string) (_ []string, err error) {
	if len(s) == 0 {
		return nil, errors.New("empty path")
	}
	if isSeparator(s[0]) {
		return nil, errPathEscapes
	}

	if runtime.GOOS == "windows" {
		// Windows cleans paths before opening them.
		s, err = rootCleanPath(s, prefix, suffix)
		if err != nil {
			return nil, err
		}
		prefix = nil
		suffix = nil
	}

	parts := append([]string{}, prefix...)
	i, j := 0, 1
	for {
		if j < len(s) && !isSeparator(s[j]) {
			// Keep looking for the end of this component.
			j++
			continue
		}
		parts = append(parts, s[i:j])
		// Advance to the next component, or end of the path.
		for j < len(s) && isSeparator(s[j]) {
			j++
		}
		if j == len(s) {
			// If this is the last path component,
			// preserve any trailing path separators.
			parts[len(parts)-1] = s[i:]
			break
		}
		if parts[len(parts)-1] == "." {
			// Remove "." components, except at the end.
			parts = parts[:len(parts)-1]
		}
		i = j
	}
	parts = append(parts, suffix...)
	return parts, nil
}

func isSeparator(c byte) bool {
	if runtime.GOOS == "windows" {
		return c == '/' || c == '\\'
	} else {
		return c == '/'
	}
}

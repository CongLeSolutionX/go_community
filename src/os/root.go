// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os

import (
	"errors"
	"internal/testlog"
	"runtime"
	"slices"
	"syscall"
)

// Root represents a directory.
//
// Methods on Root can only access files and directories within that directory.
// If any component of a file name passed to a method of Root references a location
// outside the root or (on Windows) a reserved device name such as NUL,
// the method returns an error. File names may reference the directory itself (.).
//
// Symbolic links in file names may not reference a location outside o
// File names may contain symbolic links, but symbolic links may not
// reference a location outside the root.
// Symbolic links must not be absolute.
//
// On platforms which do not provide an openat-style function,
// Root is vulnerable to TOCTOU (time-of-check-time-of-use) attacks
// in symlink validation, and cannot ensure that operations will not
// escape the root. This limitation is currently present when GOOS is
// plan9 or js.
//
// Methods on Root do not prohibit traversal of filesystem boundaries,
// Linux bind mounts, /proc special files, or access to Unix device files.
//
// Methods on Root are safe to be used from multiple goroutines simultaneously.
//
// On most platforms, creating a Root opens a file descriptor or handle referencing
// the directory. If the directory is moved, methods on Root reference the original
// handle. When GOOS is plan9 or js, a Root instead references the path name of a
// directory and does not track renames.
type Root struct {
	f *File
}

// OpenRoot opens the named directory.
// If there is an error, it will be of type *PathError.
func OpenRoot(name string) (*Root, error) {
	f, err := OpenFile(name, O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	fi, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}
	if !fi.IsDir() {
		f.Close()
		return nil, errors.New("not a directory")
	}
	return &Root{f: f}, nil
}

// Name returns the name of the directory presented to OpenRoot.
//
// It is safe to call Name after [Close].
func (r *Root) Name() string { return r.f.Name() }

// Close closes the Root.
// After Close is called, methods on Root return errors.
func (r *Root) Close() error {
	r.f.Close()
	return nil
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
func (r *Root) OpenFile(name string, flag int, perm FileMode) (*File, error) {
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
	return rootOpenRootNolog(r, name)
}

// Mkdir creates a new directory in the root
// with the specified name and permission bits (before umask).
// See [Mkdir] for more details.
func (r *Root) Mkdir(name string, perm FileMode) error {
	_, err := doInRoot(r, name, func(parent sysfdType, name string) (struct{}, error) {
		return struct{}{}, mkdirat(parent, name, perm)
	})
	return err
}

func (r *Root) logOpen(name string) {
	if log := testlog.Logger(); log != nil {
		// This won't be right if f's name has changed since it was opened,
		// but it's the best we can do.
		log.Open(r.Name() + string(PathSeparator) + name)
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
// Path separators in the last component are preserved.
func splitPathInRoot(s string, prefix, suffix []string) (_ []string, err error) {
	if len(s) == 0 {
		return nil, errors.New("empty path")
	}
	if isSeparator(s[0]) {
		return nil, errPathEscapes
	}

	if runtime.GOOS == "windows" {
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

type errSymlink string

func (errSymlink) Error() string { return "os: BUG: errSymlink is not user-visible" }

// doInRoot performs an operation on a path in a Root.
//
// It opens the directory containing the final step of the path,
// and calls f with the directory FD and name of the final step.
//
// If the path refers to a symlink, then f must return errSymlink.
// doInRoot will follow the symlink and call f again.
func doInRoot[T any](root *Root, name string, f func(parent sysfdType, name string) (T, error)) (ret T, err error) {
	parts, err := splitPathInRoot(name, nil, nil)
	if err != nil {
		return ret, err
	}

	rootfd := root.f.pfd.Sysfd
	dirfd := rootfd
	defer func() {
		if dirfd != rootfd {
			syscall.Close(dirfd)
		}
	}()

	// Maximum number of opens we will make across the entire operation.
	// TODO: Figure out if there's precedent for an appropriate value here.
	const maxSteps = 255
	if len(parts) > maxSteps {
		// Fail early if we know the path is going to have too many steps to it.
		return ret, errors.New("too many path steps")
	}

	// Maximum number of symbolic links we will follow.
	// TODO: Set this to a more considered value.
	const maxSymlinks = 16

	i := 0
	steps := 0
	symlinks := 0
	for {
		steps++
		if steps > maxSteps {
			return ret, errors.New("too many path steps")
		}

		if parts[i] == ".." {
			// Resolve one or more parent ("..") path components.
			//
			// We can't openat(dirfd, ".."), because the directory
			// may have moved since we opened it and we have no portable
			// way of verifying its parent doesn't escaped the root.
			//
			// Instead, we rewrite the original path, removing the
			// steps eliminated by ".." components, and start over
			// from the beginning.
			end := i + 1
			for end < len(parts) && parts[end] == ".." {
				end++
			}
			count := end - i
			if count > i {
				return ret, errPathEscapes
			}
			parts = slices.Delete(parts, i-count, end)
			i = 0
			if dirfd != rootfd {
				syscall.Close(dirfd)
			}
			dirfd = rootfd
			continue
		}

		if i == len(parts)-1 {
			// This is the last path step.
			// Call f to decide what to do with it.
			// If f returns ELOOP, this step is a symlink
			// which should be followed.
			ret, err = f(dirfd, parts[i])
			if _, ok := err.(errSymlink); !ok && err != syscall.ELOOP {
				return ret, err
			}
		} else {
			var fd sysfdType
			fd, err = rootOpenDir(dirfd, parts[i])
			if err == nil {
				if dirfd != rootfd {
					syscall.Close(dirfd)
				}
				dirfd = fd
			} else if _, ok := err.(errSymlink); !ok && err != syscall.ELOOP {
				return ret, err
			}
		}

		if e, ok := err.(errSymlink); ok {
			symlinks++
			if symlinks > maxSymlinks {
				return ret, errors.New("too many symlinks")
			}
			newparts, err := splitPathInRoot(string(e), parts[:i], parts[i+1:])
			if err != nil {
				return ret, err
			}
			if len(newparts) < i || !slices.Equal(parts[:i], newparts[:i]) {
				// Some component in the path which we have already traversed
				// has changed. We need to restart parsing from the root.
				i = 0
				if dirfd != rootfd {
					syscall.Close(dirfd)
				}
				dirfd = rootfd
			}
			parts = newparts
			continue
		}

		i++
	}
}

// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || windows || wasip1

package os

import (
	"runtime"
	"slices"
	"sync/atomic"
	"syscall"
)

// root implementation for platforms with a function to open a file
// relative to a directory.
type root struct {
	name string

	// refs is the refcount of open operations using fd.
	// Close sets the rootClosed bit in refs.
	// When refs == rootClosed (closed, and not in use), the fd is closed.
	refs atomic.Int64
	fd   sysfdType
}

const rootClosed = 1 << 61

func (r *root) Close() error {
	old := r.refs.Or(rootClosed)
	if old == 0 {
		syscall.Close(r.fd)
	}
	runtime.SetFinalizer(r, nil) // no need for a finalizer any more
	return nil
}

func (r *root) incref() error {
	old := r.refs.Add(1)
	if old&rootClosed != 0 {
		r.refs.Add(-1)
		return ErrClosed
	}
	return nil
}

func (r *root) decref() {
	old := r.refs.Add(-1)
	if old == rootClosed {
		syscall.Close(r.fd)
	}
}

func (r *root) Name() string {
	return r.name
}

func rootMkdir(r *Root, name string, perm FileMode) error {
	_, err := doInRoot(r, name, func(parent sysfdType, name string) (struct{}, error) {
		return struct{}{}, mkdirat(parent, name, perm)
	})
	if err != nil {
		return &PathError{Op: "mkdirat", Path: name, Err: err}
	}
	return err
}

// doInRoot performs an operation on a path in a Root.
//
// It opens the directory containing the final step of the path,
// and calls f with the directory FD and name of the final step.
//
// If the path refers to a symlink, then f must return errSymlink.
// doInRoot will follow the symlink and call f again.
func doInRoot[T any](r *Root, name string, f func(parent sysfdType, name string) (T, error)) (ret T, err error) {
	if err := r.root.incref(); err != nil {
		return ret, err
	}
	defer r.root.decref()

	parts, err := splitPathInRoot(name, nil, nil)
	if err != nil {
		return ret, err
	}

	rootfd := r.root.fd
	dirfd := rootfd
	defer func() {
		if dirfd != rootfd {
			syscall.Close(dirfd)
		}
	}()

	// When resolving .. path components, we restart path resolution from the root.
	// (We can't openat(dir, "..") to move up to the parent directory,
	// because dir may have moved since we opened it.)
	// To limit how many opens a malicious path can cause us to perform, we set
	// a limit on the total number of path steps and the total number of restarts
	// caused by .. components. If *both* limits are exceeded, we halt the operation.
	const maxSteps = 255
	const maxRestarts = 8

	i := 0
	steps := 0
	restarts := 0
	symlinks := 0
	for {
		steps++
		if steps > maxSteps && restarts > maxRestarts {
			return ret, syscall.ENAMETOOLONG
		}

		if parts[i] == ".." {
			// Resolve one or more parent ("..") path components.
			//
			// Rewrite the original path,
			// removing the steps eliminated by ".." components,
			// and start over from the beginning.
			restarts++
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
			if _, ok := err.(errSymlink); !ok {
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
			} else if _, ok := err.(errSymlink); !ok {
				return ret, err
			}
		}

		if e, ok := err.(errSymlink); ok {
			symlinks++
			if symlinks > rootMaxSymlinks {
				return ret, syscall.ELOOP
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

// errSymlink reports that a file being operated on is actually a symlink,
// and the target of that symlink.
type errSymlink string

func (errSymlink) Error() string { return "os: BUG: errSymlink is not user-visible" }

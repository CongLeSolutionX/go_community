// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package filepath

import (
	"internal/filepath"
	"strings"
	"syscall"
)

// HasPrefix exists for historical compatibility and should not be used.
//
// Deprecated: HasPrefix does not respect path boundaries and
// does not ignore case when required.
func HasPrefix(p, prefix string) bool {
	if strings.HasPrefix(p, prefix) {
		return true
	}
	return strings.HasPrefix(strings.ToLower(p), strings.ToLower(prefix))
}

func splitList(path string) []string {
	// The same implementation is used in LookPath in os/exec;
	// consider changing os/exec when changing this.

	if path == "" {
		return []string{}
	}

	// Split path, respecting but preserving quotes.
	list := []string{}
	start := 0
	quo := false
	for i := 0; i < len(path); i++ {
		switch c := path[i]; {
		case c == '"':
			quo = !quo
		case c == ListSeparator && !quo:
			list = append(list, path[start:i])
			start = i + 1
		}
	}
	list = append(list, path[start:])

	// Remove quotes.
	for i, s := range list {
		list[i] = strings.ReplaceAll(s, `"`, ``)
	}

	return list
}

func abs(path string) (string, error) {
	if path == "" {
		// syscall.FullPath returns an error on empty path, because it's not a valid path.
		// To implement Abs behavior of returning working directory on empty string input,
		// special-case empty path by changing it to "." path. See golang.org/issue/24441.
		path = "."
	}
	fullPath, err := syscall.FullPath(path)
	if err != nil {
		return "", err
	}
	return Clean(fullPath), nil
}

func join(elem []string) string {
	for i, e := range elem {
		if e != "" {
			return joinNonEmpty(elem[i:])
		}
	}
	return ""
}

// joinNonEmpty is like join, but it assumes that the first element is non-empty.
func joinNonEmpty(elem []string) string {
	if len(elem[0]) == 2 && elem[0][1] == ':' {
		// First element is drive letter without terminating slash.
		// Keep path relative to current directory on that drive.
		// Skip empty elements.
		i := 1
		for ; i < len(elem); i++ {
			if elem[i] != "" {
				break
			}
		}
		return Clean(elem[0] + strings.Join(elem[i:], string(Separator)))
	}
	// The following logic prevents Join from inadvertently creating a
	// UNC path on Windows. Unless the first element is a UNC path, Join
	// shouldn't create a UNC path. See golang.org/issue/9167.
	p := Clean(strings.Join(elem, string(Separator)))
	if !isUNC(p) {
		return p
	}
	// p == UNC only allowed when the first element is a UNC path.
	head := Clean(elem[0])
	if isUNC(head) {
		return p
	}
	// head + tail == UNC, but joining two non-UNC paths should not result
	// in a UNC path. Undo creation of UNC path.
	tail := Clean(strings.Join(elem[1:], string(Separator)))
	if head[len(head)-1] == Separator {
		return head + tail
	}
	return head + string(Separator) + tail
}

// isUNC reports whether path is a UNC path.
func isUNC(path string) bool {
	return filepath.VolumeNameLen(path) > 2
}

func sameWord(a, b string) bool {
	return strings.EqualFold(a, b)
}

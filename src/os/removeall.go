// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os

// endsWithDot returns whether the final component of path is ".".
func endsWithDot(path string) bool {
	if path == "." {
		return true
	}
	if len(path) >= 2 && path[len(path)-1] == '.' && IsPathSeparator(path[len(path)-2]) {
		return true
	}
	return false
}

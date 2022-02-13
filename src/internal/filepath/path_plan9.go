// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package filepath

import "strings"

// IsAbs reports whether the path is absolute.
func IsAbs(path string) bool {
	return strings.HasPrefix(path, "/") || strings.HasPrefix(path, "#")
}

// VolumeNameLen returns length of the leading volume name on Windows.
// It returns 0 elsewhere.
func VolumeNameLen(path string) int {
	return 0
}

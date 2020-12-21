// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package renameio writes files atomically by renaming temporary files.
package renameio

import (
	"path/filepath"
)

const patternSuffix = ".tmp"

// Pattern returns a glob pattern that matches the unrenamed temporary files
// created when writing to filename.
func Pattern(filename string) string {
	return filepath.Join(filepath.Dir(filename), filepath.Base(filename)+patternSuffix)
}

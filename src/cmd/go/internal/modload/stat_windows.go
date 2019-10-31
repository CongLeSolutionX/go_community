// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package modload

import (
	"os"
)

func hasWritePermSys(fi os.FileInfo) bool {
	// fi.Sys() returns a value of type *syscall.Win32FileAttributeData,
	// but that doesn't tell us anything about file permissions.
	return false
}

// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package renameio

import (
	"internal/syscall/windows"
	"os"
	_ "unsafe"
)

//go:linkname os_fixLongPath os.fixLongPath
func os_fixLongPath(path string) string

// rename is like os.Rename, but uses ReplaceFile instead of MoveFileEx.
//
// TODO(bcmills): For Go 1.14, try changing os.Rename itself to do this.
func rename(oldpath, newpath string) error {
	e := windows.Rename(os_fixLongPath(oldpath), os_fixLongPath(newpath))
	if e != nil {
		return &os.LinkError{Op: "rename", Old: oldpath, New: newpath, Err: e}
	}
	return nil
}

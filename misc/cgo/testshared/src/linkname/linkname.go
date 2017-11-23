// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package linkname

import (
	_ "runtime"
	_ "unsafe"
)

type errStr string

//go:linkname errStr.Error runtime.errorString.Error
func (e errStr) Error() string

//go:linkname stopTheWorld runtime.stopTheWorld
func stopTheWorld(reason string)

//go:linkname startTheWorld runtime.startTheWorld
func startTheWorld()

//go:linkname allgs runtime.allgs
var allgs []uintptr

// Exercise the use of linknamed variables, functions and methods.
func Test(s string) (string, int) {
	stopTheWorld("exe4")
	l := len(allgs)
	startTheWorld()
	return errStr(s).Error(), l
}

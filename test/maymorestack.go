// run -gcflags=all=-d=maymorestack=main.mayMoreStack

// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test compiler's -d maymorestack option.

package main

import (
	"sync/atomic"
	_ "unsafe" // For go:linkname
)

var ok uint32

func main() {
	if atomic.LoadUint32(&ok) == 0 {
		panic("mayMoreStack not called")
	}
}

// This must be ABI0 so it can be called from obj.
//go:linkname mayMoreStack
//go:nosplit
func mayMoreStack() {
	atomic.StoreUint32(&ok, 1)
}

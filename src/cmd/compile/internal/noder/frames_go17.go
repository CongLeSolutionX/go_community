// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build go1.7

package noder

import "runtime"

func walkFrames(pcs []uintptr, visit frameVisitor) {
	frames := runtime.CallersFrames(pcs)
	for {
		frame, more := frames.Next()
		if !more {
			return
		}
		visit(frame.File, frame.Line, frame.Function, frame.PC-frame.Entry)
	}
}

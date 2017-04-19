// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime_test

import (
	"fmt"
	"runtime"
	"strings"
)

func ExampleFrames() {
	c := func() {
		// Ask runtime.Callers for up to 10 pcs, including runtime.Callers itself.
		pc := make([]uintptr, 10)
		n := runtime.Callers(0, pc)
		if n == 0 {
			// No pcs available. Stop now.
			// This can happen if the first argument to runtime.Callers is large.
			return
		}

		// Pass only valid pcs to runtime.CallersFrames.
		pc = pc[:n]
		frames := runtime.CallersFrames(pc)
		var frame runtime.Frame
		more := true
		for more {
			frame, more = frames.Next()
			// Those (up to) 10 pcs could correspond to
			// an indefinite number of Frames, due to inlining.
			// To keep this example's output stable,
			// even if the compiler's inlining decisions change,
			// we stop the moment that we leave package runtime.
			if !strings.Contains(frame.File, "runtime/") {
				break
			}
			fmt.Printf("- more:%v | %s\n", more, frame.Function)
		}
	}

	b := func() { c() }
	a := func() { b() }

	a()
	// Output:
	// - more:true | runtime.Callers
	// - more:true | runtime_test.ExampleFrames.func1
	// - more:true | runtime_test.ExampleFrames.func2
	// - more:true | runtime_test.ExampleFrames.func3
	// - more:true | runtime_test.ExampleFrames
}

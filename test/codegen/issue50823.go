// asmcheck

// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package codegen

func f() int {
	ch1 := make(chan int)
	ch2 := make(chan int)

	select {
	// amd64:-"JEQ",-"JNE"
	case <-ch1:
		return 1
	case <-ch2:
		return 2
	}

}

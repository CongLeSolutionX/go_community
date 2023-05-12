// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !amd64.v3

package runtime

import "unsafe"

// This exists for architectures where [memclrNoHeapPointers] is able to atomically clear pointers.

//go:noescape
//go:linkname doMemclrWithPointers reflect.memclrNoHeapPointers
func doMemclrWithPointers(ptr unsafe.Pointer, n uintptr)

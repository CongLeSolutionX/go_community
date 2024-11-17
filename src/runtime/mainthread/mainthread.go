// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package mainthread mediates access to the program's main thread.
//
// Most Go programs do not need to run on specific threads
// and can ignore this package, but some C libraries, often GUI-related libraries,
// only work when invoked from the program's main thread.
//
// [Do] runs a function on the main thread. No other code can run on the main thread
// until that function returns.
//
// Each package's initialization functions always run on the main thread,
// as if by successive calls to Do(init).
//
// For compatibility with earlier versions of Go, if an init function calls [runtime.LockOSThread],
// then package main's func main also runs on the main thread, as if by Do(main).
package mainthread

import _ "unsafe"

// Do calls f on the main thread.
// Nothing else runs on the main thread until f returns.
// If f calls Do, the nested call panics.
//
// Package initialization functions run as if by Do(init).
// If an init function calls [runtime.LockOSThread], then package main's func main
// runs as if by Do(main), until the thread is unlocked using [runtime.UnlockOSThread].
//
// Do panics if the Go runtime is not in control of the main thread, such as in build modes
// c-shared and c-archive.
//
//go:linkname Do runtime.mainThreadDo
func Do(f func())

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bytealg

import "unsafe"

// Note: there's no equal_generic.go because every platform must implement at least memequal_varlen in assembly.

//go:noescape
func Equal(a, b []byte) bool

// The following are defined in assembly in this package, but exported
// to other packages. Provide Go declarations to go with their
// assembly definitions.

//go:linkname bytes_Equal bytes.Equal
func bytes_Equal(a, b []byte) bool

// The compiler generates calls to runtime.memequal and runtime.memequal_varlen.
// In addition, the runtime calls runtime.memequal explicitly.
// Those functions are implemented in this package.

//go:linkname runtime_memequal runtime.memequal
func runtime_memequal(a, b unsafe.Pointer, size uintptr) bool

//go:linkname runtime_memequal_varlen runtime.memequal_varlen
func runtime_memequal_varlen(a, b unsafe.Pointer) bool

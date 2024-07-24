// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build arm64 || amd64

// Export for testing.

package secret

import "unsafe"

func GetStack() (uintptr, uintptr) {
	return getStack()
}

func LoadRegisters(p unsafe.Pointer) {
	loadRegisters(p)
}

func SpillRegisters(p unsafe.Pointer) uintptr {
	return spillRegisters(p)
}

func Read(addr uintptr) int64 {
	return read(addr)
}

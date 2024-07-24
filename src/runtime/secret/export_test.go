// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Export for testing.

package secret

import (
	"internal/cpu"
	"unsafe"
)

func GetStack() (uintptr, uintptr) {
	return getStack()
}

func LoadRegisters(p unsafe.Pointer) {
	loadRegisters(p)
}

func SpillRegisters(p unsafe.Pointer) {
	spillRegisters(p)
}

// in assembly
//
//go:noescape
func loadRegisters(p unsafe.Pointer) // load data from p into test registers
//go:noescape
func spillRegisters(p unsafe.Pointer) // spill data from test registers into p

const (
	offsetX86HasAVX     = unsafe.Offsetof(cpu.X86.HasAVX)
	offsetX86HasAVX512F = unsafe.Offsetof(cpu.X86.HasAVX512F)
)

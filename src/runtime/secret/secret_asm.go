// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build arm64 || amd64

package secret

import (
	"internal/cpu"
	"unsafe"
)

// load data from p into test registers.
//
//go:noescape
func loadRegisters(p unsafe.Pointer)

// spill data from test registers into p.
// Returns the amount of space filled in.
//
//go:noescape
func spillRegisters(p unsafe.Pointer) uintptr

const (
	offsetX86HasAVX     = unsafe.Offsetof(cpu.X86.HasAVX)
	offsetX86HasAVX512F = unsafe.Offsetof(cpu.X86.HasAVX512F)
)

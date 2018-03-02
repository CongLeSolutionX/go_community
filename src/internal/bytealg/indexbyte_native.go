// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build amd64 s390x arm64

package bytealg

import (
	"internal/cpu"
	"unsafe"
)

// Offsets into internal/cpu records for use in assembly
// TODO: find a better way to do this?
const x86_HasAVX2 = unsafe.Offsetof(cpu.X86.HasAVX2)
const s390x_HasVX = unsafe.Offsetof(cpu.S390X.HasVX)

//go:noescape
func IndexStringByte(s string, c byte) int

//go:noescape
func IndexBytesByte(b []byte, c byte) int

// dummy declaration to induce empty args stackmap for indexbytebody.
// TODO: why is this needed?
func indexbytebody()

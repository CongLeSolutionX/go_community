// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ppc64 ppc64le

package runtime

// For go:linkname
import _ "unsafe"

// ppc64x doesn't have a 'cpuid' instruction equivalent and relies on
// HWCAP/HWCAP2 bits for hardware capabilities.

//go:linkname cpu_hwcap internal/cpu.hwcap
//go:linkname cpu_hwcap2 internal/cpu.hwcap2
var cpu_hwcap uintptr
var cpu_hwcap2 uintptr

func archauxv(tag, val uintptr) {
	switch tag {
	case _AT_HWCAP:
		cpu_hwcap = val
	case _AT_HWCAP2:
		cpu_hwcap2 = val
	}
}

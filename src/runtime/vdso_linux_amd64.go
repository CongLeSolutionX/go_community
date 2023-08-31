// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

const (
	// vdsoArrayMax is the byte-size of a maximally sized array on this architecture.
	// See cmd/compile/internal/amd64/galign.go arch.MAXWIDTH initialization.
	vdsoArrayMax = 1<<50 - 1
)

const vdsoLinuxVersion = "LINUX_2.6"

var vdsoSymbolKeys = []vdsoSymbolKey{
	{"__vdso_gettimeofday", &vdsoGettimeofdaySym},
	{"__vdso_clock_gettime", &vdsoClockgettimeSym},
}

var (
	vdsoGettimeofdaySym uintptr
	vdsoClockgettimeSym uintptr
)

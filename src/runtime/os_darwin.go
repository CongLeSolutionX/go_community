// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

func bsdthread_register() int32
func mach_thread_self() uint32

//go:noescape
func sysctl(mib *uint32, miblen uint32, out *byte, size *uintptr, dst *byte, ndst uintptr) int32

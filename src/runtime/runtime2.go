// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	"unsafe"
)

// describes how to handle callback
type wincallbackcontext struct {
	gobody       unsafe.Pointer // go function to call
	argsize      uintptr        // callback arguments size (in bytes)
	restorestack uintptr        // adjust stack on return by (in bytes) (386 only)
	cleanstack   bool
}

type sigtabtt struct {
	flags int32
	name  *int8
}

type forcegcstate struct {
	lock _core.Mutex
	g    *_core.G
	idle uint32
}

/*
 * known to compiler
 */
const (
	_Structrnd = _lock.RegSize
)

// startup_random_data holds random bytes initialized at startup.  These come from
// the ELF AT_RANDOM auxiliary vector (vdso_linux_amd64.go or os_linux_386.go).
var startupRandomData []byte

var (
	emptystring string
	goos        *int8
	cpuid_edx   uint32
	forcegc     forcegcstate
)

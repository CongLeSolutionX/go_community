// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package seq

import (
	_maps "runtime/internal/maps"
	_sched "runtime/internal/sched"
	"unsafe"
)

// rawmem returns a chunk of pointerless memory.  It is
// not zeroed.
func rawmem(size uintptr) unsafe.Pointer {
	return _maps.Mallocgc(size, nil, _sched.FlagNoScan|_sched.FlagNoZero)
}

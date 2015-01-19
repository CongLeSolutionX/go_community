// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Parallel for algorithm.

package gc

import (
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
	"unsafe"
)

func parforsetup(desc *_sched.Parfor, nthr, n uint32, ctx unsafe.Pointer, wait bool, body func(*_sched.Parfor, uint32)) {
	if desc == nil || nthr == 0 || nthr > desc.Nthrmax || body == nil {
		print("desc=", desc, " nthr=", nthr, " count=", n, " body=", body, "\n")
		_lock.Throw("parfor: invalid args")
	}

	desc.Body = *(*unsafe.Pointer)(unsafe.Pointer(&body))
	desc.Done = 0
	desc.Nthr = nthr
	desc.Thrseq = 0
	desc.Cnt = n
	desc.Ctx = ctx
	desc.Wait = wait
	desc.Nsteal = 0
	desc.Nstealcnt = 0
	desc.Nprocyield = 0
	desc.Nosyield = 0
	desc.Nsleep = 0

	for i := uint32(0); i < nthr; i++ {
		begin := uint32(uint64(n) * uint64(i) / uint64(nthr))
		end := uint32(uint64(n) * uint64(i+1) / uint64(nthr))
		pos := &_sched.Desc_thr_index(desc, i).Pos
		if uintptr(unsafe.Pointer(pos))&7 != 0 {
			_lock.Throw("parforsetup: pos is not aligned")
		}
		*pos = uint64(begin) | uint64(end)<<32
	}
}

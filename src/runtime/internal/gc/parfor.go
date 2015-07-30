// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Parallel for algorithm.

package gc

import (
	_base "runtime/internal/base"
)

func parforalloc(nthrmax uint32) *_base.Parfor {
	return &_base.Parfor{
		Thr: make([]_base.Parforthread, nthrmax),
	}
}

// Parforsetup initializes desc for a parallel for operation with nthr
// threads executing n jobs.
//
// On return the nthr threads are each expected to call parfordo(desc)
// to run the operation. During those calls, for each i in [0, n), one
// thread will be used invoke body(desc, i).
// If wait is true, no parfordo will return until all work has been completed.
// If wait is false, parfordo may return when there is a small amount
// of work left, under the assumption that another thread has that
// work well in hand.
func parforsetup(desc *_base.Parfor, nthr, n uint32, wait bool, body func(*_base.Parfor, uint32)) {
	if desc == nil || nthr == 0 || nthr > uint32(len(desc.Thr)) || body == nil {
		print("desc=", desc, " nthr=", nthr, " count=", n, " body=", body, "\n")
		_base.Throw("parfor: invalid args")
	}

	desc.Body = body
	desc.Done = 0
	desc.Nthr = nthr
	desc.Thrseq = 0
	desc.Cnt = n
	desc.Wait = wait
	desc.Nsteal = 0
	desc.Nstealcnt = 0
	desc.Nprocyield = 0
	desc.Nosyield = 0
	desc.Nsleep = 0

	for i := range desc.Thr {
		begin := uint32(uint64(n) * uint64(i) / uint64(nthr))
		end := uint32(uint64(n) * uint64(i+1) / uint64(nthr))
		desc.Thr[i].Pos = uint64(begin) | uint64(end)<<32
	}
}

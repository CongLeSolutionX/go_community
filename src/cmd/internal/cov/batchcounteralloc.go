// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cov

type BatchCounterAlloc struct {
	pool []uint32
}

func (ca *BatchCounterAlloc) AllocateCounters(n int) []uint32 {
	const chunk = 8192
	if n > cap(ca.pool) {
		siz := chunk
		if n > chunk {
			siz = n
		}
		ca.pool = make([]uint32, siz)
	}
	rv := ca.pool[:n]
	ca.pool = ca.pool[n:]
	return rv
}

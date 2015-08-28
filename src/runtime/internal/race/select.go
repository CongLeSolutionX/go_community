// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package race

import (
	_base "runtime/internal/base"
	"unsafe"
)

func (c *Hchan) Sortkey() uintptr {
	// TODO(khr): if we have a moving garbage collector, we'll need to
	// change this function.
	return uintptr(unsafe.Pointer(c))
}

func (q *Waitq) DequeueSudoG(sgp *_base.Sudog) {
	x := sgp.Prev
	y := sgp.Next
	if x != nil {
		if y != nil {
			// middle of queue
			x.Next = y
			y.Prev = x
			sgp.Next = nil
			sgp.Prev = nil
			return
		}
		// end of queue
		x.Next = nil
		q.last = x
		sgp.Prev = nil
		return
	}
	if y != nil {
		// start of queue
		y.Prev = nil
		q.First = y
		sgp.Next = nil
		return
	}

	// x==y==nil.  Either sgp is the only element in the queue,
	// or it has already been removed.  Use q.first to disambiguate.
	if q.First == sgp {
		q.First = nil
		q.last = nil
	}
}

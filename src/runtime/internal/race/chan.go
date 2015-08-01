// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package race

import (
	_base "runtime/internal/base"
	"unsafe"
)

type Hchan struct {
	Qcount   uint           // total data in the queue
	Dataqsiz uint           // size of the circular queue
	Buf      unsafe.Pointer // points to an array of dataqsiz elements
	Elemsize uint16
	Closed   uint32
	Elemtype *_base.Type // element type
	Sendx    uint        // send index
	Recvx    uint        // receive index
	Recvq    Waitq       // list of recv waiters
	Sendq    Waitq       // list of send waiters
	Lock     _base.Mutex
}

type Waitq struct {
	First *_base.Sudog
	last  *_base.Sudog
}

// chanbuf(c, i) is pointer to the i'th slot in the buffer.
func Chanbuf(c *Hchan, i uint) unsafe.Pointer {
	return _base.Add(c.Buf, uintptr(i)*uintptr(c.Elemsize))
}

func (q *Waitq) Enqueue(sgp *_base.Sudog) {
	sgp.Next = nil
	x := q.last
	if x == nil {
		sgp.Prev = nil
		q.First = sgp
		q.last = sgp
		return
	}
	sgp.Prev = x
	x.Next = sgp
	q.last = sgp
}

func (q *Waitq) Dequeue() *_base.Sudog {
	for {
		sgp := q.First
		if sgp == nil {
			return nil
		}
		y := sgp.Next
		if y == nil {
			q.First = nil
			q.last = nil
		} else {
			y.Prev = nil
			q.First = y
			sgp.Next = nil // mark as removed (see dequeueSudog)
		}

		// if sgp participates in a select and is already signaled, ignore it
		if sgp.Selectdone != nil {
			// claim the right to signal
			if *sgp.Selectdone != 0 || !_base.Cas(sgp.Selectdone, 0, 1) {
				continue
			}
		}

		return sgp
	}
}

func Racesync(c *Hchan, sg *_base.Sudog) {
	Racerelease(Chanbuf(c, 0))
	raceacquireg(sg.G, Chanbuf(c, 0))
	racereleaseg(sg.G, Chanbuf(c, 0))
	_base.Raceacquire(Chanbuf(c, 0))
}

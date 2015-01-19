// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package channels

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
	_sem "runtime/internal/sem"
	"unsafe"
)

const (
	MaxAlign  = 8
	HchanSize = unsafe.Sizeof(Hchan{}) + uintptr(-int(unsafe.Sizeof(Hchan{}))&(MaxAlign-1))
	DebugChan = false
)

// chanbuf(c, i) is pointer to the i'th slot in the buffer.
func chanbuf(c *Hchan, i uint) unsafe.Pointer {
	return _core.Add(unsafe.Pointer(c.Buf), uintptr(i)*uintptr(c.Elemsize))
}

/*
 * generic single channel send/recv
 * If block is not nil,
 * then the protocol will not
 * sleep but return if it could
 * not complete.
 *
 * sleep can wake up with g.param == nil
 * when a channel involved in the sleep has
 * been closed.  it is easiest to loop and re-run
 * the operation; we'll see that it's now closed.
 */
func Chansend(t *Chantype, c *Hchan, ep unsafe.Pointer, block bool, callerpc uintptr) bool {
	if _sched.Raceenabled {
		RaceReadObjectPC(t.Elem, ep, callerpc, _lock.FuncPC(Chansend))
	}

	if c == nil {
		if !block {
			return false
		}
		_sched.Gopark(nil, nil, "chan send (nil chan)")
		_lock.Throw("unreachable")
	}

	if DebugChan {
		print("chansend: chan=", c, "\n")
	}

	if _sched.Raceenabled {
		Racereadpc(unsafe.Pointer(c), callerpc, _lock.FuncPC(Chansend))
	}

	// Fast path: check for failed non-blocking operation without acquiring the lock.
	//
	// After observing that the channel is not closed, we observe that the channel is
	// not ready for sending. Each of these observations is a single word-sized read
	// (first c.closed and second c.recvq.first or c.qcount depending on kind of channel).
	// Because a closed channel cannot transition from 'ready for sending' to
	// 'not ready for sending', even if the channel is closed between the two observations,
	// they imply a moment between the two when the channel was both not yet closed
	// and not ready for sending. We behave as if we observed the channel at that moment,
	// and report that the send cannot proceed.
	//
	// It is okay if the reads are reordered here: if we observe that the channel is not
	// ready for sending and then observe that it is not closed, that implies that the
	// channel wasn't closed during the first observation.
	if !block && c.Closed == 0 && ((c.Dataqsiz == 0 && c.Recvq.first == nil) ||
		(c.Dataqsiz > 0 && c.Qcount == c.Dataqsiz)) {
		return false
	}

	var t0 int64
	if _sem.Blockprofilerate > 0 {
		t0 = _sched.Cputicks()
	}

	_lock.Lock(&c.Lock)
	if c.Closed != 0 {
		_lock.Unlock(&c.Lock)
		panic("send on closed channel")
	}

	if c.Dataqsiz == 0 { // synchronous channel
		sg := c.Recvq.Dequeue()
		if sg != nil { // found a waiting receiver
			if _sched.Raceenabled {
				racesync(c, sg)
			}
			_lock.Unlock(&c.Lock)

			recvg := sg.G
			if sg.Elem != nil {
				_sched.Memmove(unsafe.Pointer(sg.Elem), ep, uintptr(c.Elemsize))
				sg.Elem = nil
			}
			recvg.Param = unsafe.Pointer(sg)
			if sg.Releasetime != 0 {
				sg.Releasetime = _sched.Cputicks()
			}
			_sched.Goready(recvg)
			return true
		}

		if !block {
			_lock.Unlock(&c.Lock)
			return false
		}

		// no receiver available: block on this channel.
		gp := _core.Getg()
		mysg := _sem.AcquireSudog()
		mysg.Releasetime = 0
		if t0 != 0 {
			mysg.Releasetime = -1
		}
		mysg.Elem = ep
		mysg.Waitlink = nil
		gp.Waiting = mysg
		mysg.G = gp
		mysg.Selectdone = nil
		gp.Param = nil
		c.Sendq.enqueue(mysg)
		_sched.Goparkunlock(&c.Lock, "chan send")

		// someone woke us up.
		if mysg != gp.Waiting {
			_lock.Throw("G waiting list is corrupted!")
		}
		gp.Waiting = nil
		if gp.Param == nil {
			if c.Closed == 0 {
				_lock.Throw("chansend: spurious wakeup")
			}
			panic("send on closed channel")
		}
		gp.Param = nil
		if mysg.Releasetime > 0 {
			_sem.Blockevent(int64(mysg.Releasetime)-t0, 2)
		}
		_sem.ReleaseSudog(mysg)
		return true
	}

	// asynchronous channel
	// wait for some space to write our data
	var t1 int64
	for c.Qcount >= c.Dataqsiz {
		if !block {
			_lock.Unlock(&c.Lock)
			return false
		}
		gp := _core.Getg()
		mysg := _sem.AcquireSudog()
		mysg.Releasetime = 0
		if t0 != 0 {
			mysg.Releasetime = -1
		}
		mysg.G = gp
		mysg.Elem = nil
		mysg.Selectdone = nil
		c.Sendq.enqueue(mysg)
		_sched.Goparkunlock(&c.Lock, "chan send")

		// someone woke us up - try again
		if mysg.Releasetime > 0 {
			t1 = mysg.Releasetime
		}
		_sem.ReleaseSudog(mysg)
		_lock.Lock(&c.Lock)
		if c.Closed != 0 {
			_lock.Unlock(&c.Lock)
			panic("send on closed channel")
		}
	}

	// write our data into the channel buffer
	if _sched.Raceenabled {
		_sched.Raceacquire(chanbuf(c, c.sendx))
		Racerelease(chanbuf(c, c.sendx))
	}
	_sched.Memmove(chanbuf(c, c.sendx), ep, uintptr(c.Elemsize))
	c.sendx++
	if c.sendx == c.Dataqsiz {
		c.sendx = 0
	}
	c.Qcount++

	// wake up a waiting receiver
	sg := c.Recvq.Dequeue()
	if sg != nil {
		recvg := sg.G
		_lock.Unlock(&c.Lock)
		if sg.Releasetime != 0 {
			sg.Releasetime = _sched.Cputicks()
		}
		_sched.Goready(recvg)
	} else {
		_lock.Unlock(&c.Lock)
	}
	if t1 > 0 {
		_sem.Blockevent(t1-t0, 2)
	}
	return true
}

// entry points for <- c from compiled code
//go:nosplit
func chanrecv1(t *Chantype, c *Hchan, elem unsafe.Pointer) {
	chanrecv(t, c, elem, true)
}

//go:nosplit
func chanrecv2(t *Chantype, c *Hchan, elem unsafe.Pointer) (received bool) {
	_, received = chanrecv(t, c, elem, true)
	return
}

// chanrecv receives on channel c and writes the received data to ep.
// ep may be nil, in which case received data is ignored.
// If block == false and no elements are available, returns (false, false).
// Otherwise, if c is closed, zeros *ep and returns (true, false).
// Otherwise, fills in *ep with an element and returns (true, true).
func chanrecv(t *Chantype, c *Hchan, ep unsafe.Pointer, block bool) (selected, received bool) {
	// raceenabled: don't need to check ep, as it is always on the stack.

	if DebugChan {
		print("chanrecv: chan=", c, "\n")
	}

	if c == nil {
		if !block {
			return
		}
		_sched.Gopark(nil, nil, "chan receive (nil chan)")
		_lock.Throw("unreachable")
	}

	// Fast path: check for failed non-blocking operation without acquiring the lock.
	//
	// After observing that the channel is not ready for receiving, we observe that the
	// channel is not closed. Each of these observations is a single word-sized read
	// (first c.sendq.first or c.qcount, and second c.closed).
	// Because a channel cannot be reopened, the later observation of the channel
	// being not closed implies that it was also not closed at the moment of the
	// first observation. We behave as if we observed the channel at that moment
	// and report that the receive cannot proceed.
	//
	// The order of operations is important here: reversing the operations can lead to
	// incorrect behavior when racing with a close.
	if !block && (c.Dataqsiz == 0 && c.Sendq.first == nil ||
		c.Dataqsiz > 0 && atomicloaduint(&c.Qcount) == 0) &&
		_lock.Atomicload(&c.Closed) == 0 {
		return
	}

	var t0 int64
	if _sem.Blockprofilerate > 0 {
		t0 = _sched.Cputicks()
	}

	_lock.Lock(&c.Lock)
	if c.Dataqsiz == 0 { // synchronous channel
		if c.Closed != 0 {
			return recvclosed(c, ep)
		}

		sg := c.Sendq.Dequeue()
		if sg != nil {
			if _sched.Raceenabled {
				racesync(c, sg)
			}
			_lock.Unlock(&c.Lock)

			if ep != nil {
				_sched.Memmove(ep, sg.Elem, uintptr(c.Elemsize))
			}
			sg.Elem = nil
			gp := sg.G
			gp.Param = unsafe.Pointer(sg)
			if sg.Releasetime != 0 {
				sg.Releasetime = _sched.Cputicks()
			}
			_sched.Goready(gp)
			selected = true
			received = true
			return
		}

		if !block {
			_lock.Unlock(&c.Lock)
			return
		}

		// no sender available: block on this channel.
		gp := _core.Getg()
		mysg := _sem.AcquireSudog()
		mysg.Releasetime = 0
		if t0 != 0 {
			mysg.Releasetime = -1
		}
		mysg.Elem = ep
		mysg.Waitlink = nil
		gp.Waiting = mysg
		mysg.G = gp
		mysg.Selectdone = nil
		gp.Param = nil
		c.Recvq.enqueue(mysg)
		_sched.Goparkunlock(&c.Lock, "chan receive")

		// someone woke us up
		if mysg != gp.Waiting {
			_lock.Throw("G waiting list is corrupted!")
		}
		gp.Waiting = nil
		if mysg.Releasetime > 0 {
			_sem.Blockevent(mysg.Releasetime-t0, 2)
		}
		haveData := gp.Param != nil
		gp.Param = nil
		_sem.ReleaseSudog(mysg)

		if haveData {
			// a sender sent us some data. It already wrote to ep.
			selected = true
			received = true
			return
		}

		_lock.Lock(&c.Lock)
		if c.Closed == 0 {
			_lock.Throw("chanrecv: spurious wakeup")
		}
		return recvclosed(c, ep)
	}

	// asynchronous channel
	// wait for some data to appear
	var t1 int64
	for c.Qcount <= 0 {
		if c.Closed != 0 {
			selected, received = recvclosed(c, ep)
			if t1 > 0 {
				_sem.Blockevent(t1-t0, 2)
			}
			return
		}

		if !block {
			_lock.Unlock(&c.Lock)
			return
		}

		// wait for someone to send an element
		gp := _core.Getg()
		mysg := _sem.AcquireSudog()
		mysg.Releasetime = 0
		if t0 != 0 {
			mysg.Releasetime = -1
		}
		mysg.Elem = nil
		mysg.G = gp
		mysg.Selectdone = nil

		c.Recvq.enqueue(mysg)
		_sched.Goparkunlock(&c.Lock, "chan receive")

		// someone woke us up - try again
		if mysg.Releasetime > 0 {
			t1 = mysg.Releasetime
		}
		_sem.ReleaseSudog(mysg)
		_lock.Lock(&c.Lock)
	}

	if _sched.Raceenabled {
		_sched.Raceacquire(chanbuf(c, c.recvx))
		Racerelease(chanbuf(c, c.recvx))
	}
	if ep != nil {
		_sched.Memmove(ep, chanbuf(c, c.recvx), uintptr(c.Elemsize))
	}
	_core.Memclr(chanbuf(c, c.recvx), uintptr(c.Elemsize))

	c.recvx++
	if c.recvx == c.Dataqsiz {
		c.recvx = 0
	}
	c.Qcount--

	// ping a sender now that there is space
	sg := c.Sendq.Dequeue()
	if sg != nil {
		gp := sg.G
		_lock.Unlock(&c.Lock)
		if sg.Releasetime != 0 {
			sg.Releasetime = _sched.Cputicks()
		}
		_sched.Goready(gp)
	} else {
		_lock.Unlock(&c.Lock)
	}

	if t1 > 0 {
		_sem.Blockevent(t1-t0, 2)
	}
	selected = true
	received = true
	return
}

// recvclosed is a helper function for chanrecv.  Handles cleanup
// when the receiver encounters a closed channel.
// Caller must hold c.lock, recvclosed will release the lock.
func recvclosed(c *Hchan, ep unsafe.Pointer) (selected, recevied bool) {
	if _sched.Raceenabled {
		_sched.Raceacquire(unsafe.Pointer(c))
	}
	_lock.Unlock(&c.Lock)
	if ep != nil {
		_core.Memclr(ep, uintptr(c.Elemsize))
	}
	return true, false
}

// compiler implements
//
//	select {
//	case c <- v:
//		... foo
//	default:
//		... bar
//	}
//
// as
//
//	if selectnbsend(c, v) {
//		... foo
//	} else {
//		... bar
//	}
//
func selectnbsend(t *Chantype, c *Hchan, elem unsafe.Pointer) (selected bool) {
	return Chansend(t, c, elem, false, _lock.Getcallerpc(unsafe.Pointer(&t)))
}

// compiler implements
//
//	select {
//	case v = <-c:
//		... foo
//	default:
//		... bar
//	}
//
// as
//
//	if selectnbrecv(&v, c) {
//		... foo
//	} else {
//		... bar
//	}
//
func selectnbrecv(t *Chantype, elem unsafe.Pointer, c *Hchan) (selected bool) {
	selected, _ = chanrecv(t, c, elem, false)
	return
}

// compiler implements
//
//	select {
//	case v, ok = <-c:
//		... foo
//	default:
//		... bar
//	}
//
// as
//
//	if c != nil && selectnbrecv2(&v, &ok, c) {
//		... foo
//	} else {
//		... bar
//	}
//
func selectnbrecv2(t *Chantype, elem unsafe.Pointer, received *bool, c *Hchan) (selected bool) {
	// TODO(khr): just return 2 values from this function, now that it is in Go.
	selected, *received = chanrecv(t, c, elem, false)
	return
}

//go:linkname reflect_chanrecv reflect.chanrecv
func reflect_chanrecv(t *Chantype, c *Hchan, nb bool, elem unsafe.Pointer) (selected bool, received bool) {
	return chanrecv(t, c, elem, !nb)
}

func (q *waitq) enqueue(sgp *_core.Sudog) {
	sgp.Next = nil
	x := q.last
	if x == nil {
		sgp.Prev = nil
		q.first = sgp
		q.last = sgp
		return
	}
	sgp.Prev = x
	x.Next = sgp
	q.last = sgp
}

func (q *waitq) Dequeue() *_core.Sudog {
	for {
		sgp := q.first
		if sgp == nil {
			return nil
		}
		y := sgp.Next
		if y == nil {
			q.first = nil
			q.last = nil
		} else {
			y.Prev = nil
			q.first = y
			sgp.Next = nil // mark as removed (see dequeueSudog)
		}

		// if sgp participates in a select and is already signaled, ignore it
		if sgp.Selectdone != nil {
			// claim the right to signal
			if *sgp.Selectdone != 0 || !_sched.Cas(sgp.Selectdone, 0, 1) {
				continue
			}
		}

		return sgp
	}
}

func racesync(c *Hchan, sg *_core.Sudog) {
	Racerelease(chanbuf(c, 0))
	raceacquireg(sg.G, chanbuf(c, 0))
	racereleaseg(sg.G, chanbuf(c, 0))
	_sched.Raceacquire(chanbuf(c, 0))
}

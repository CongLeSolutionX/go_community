// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_base "runtime/internal/base"
	_gc "runtime/internal/gc"
	_iface "runtime/internal/iface"
	_race "runtime/internal/race"
	"unsafe"
)

const (
	maxAlign  = 8
	hchanSize = unsafe.Sizeof(_race.Hchan{}) + uintptr(-int(unsafe.Sizeof(_race.Hchan{}))&(maxAlign-1))
	debugChan = false
)

//go:linkname reflect_makechan reflect.makechan
func reflect_makechan(t *chantype, size int64) *_race.Hchan {
	return makechan(t, size)
}

func makechan(t *chantype, size int64) *_race.Hchan {
	elem := t.elem

	// compiler checks this but be safe.
	if elem.Size >= 1<<16 {
		_base.Throw("makechan: invalid channel element type")
	}
	if hchanSize%maxAlign != 0 || elem.Align > maxAlign {
		_base.Throw("makechan: bad alignment")
	}
	if size < 0 || int64(uintptr(size)) != size || (elem.Size > 0 && uintptr(size) > (_base.MaxMem-hchanSize)/uintptr(elem.Size)) {
		panic("makechan: size out of range")
	}

	var c *_race.Hchan
	if elem.Kind&_iface.KindNoPointers != 0 || size == 0 {
		// Allocate memory in one call.
		// Hchan does not contain pointers interesting for GC in this case:
		// buf points into the same allocation, elemtype is persistent.
		// SudoG's are referenced from their owning thread so they can't be collected.
		// TODO(dvyukov,rlh): Rethink when collector can move allocated objects.
		c = (*_race.Hchan)(_iface.Mallocgc(hchanSize+uintptr(size)*uintptr(elem.Size), nil, _base.FlagNoScan))
		if size > 0 && elem.Size != 0 {
			c.Buf = _base.Add(unsafe.Pointer(c), hchanSize)
		} else {
			// race detector uses this location for synchronization
			// Also prevents us from pointing beyond the allocation (see issue 9401).
			c.Buf = unsafe.Pointer(c)
		}
	} else {
		c = new(_race.Hchan)
		c.Buf = newarray(elem, uintptr(size))
	}
	c.Elemsize = uint16(elem.Size)
	c.Elemtype = elem
	c.Dataqsiz = uint(size)

	if debugChan {
		print("makechan: chan=", c, "; elemsize=", elem.Size, "; elemalg=", elem.Alg, "; dataqsiz=", size, "\n")
	}
	return c
}

// entry point for c <- x from compiled code
//go:nosplit
func chansend1(t *chantype, c *_race.Hchan, elem unsafe.Pointer) {
	chansend(t, c, elem, true, _base.Getcallerpc(unsafe.Pointer(&t)))
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
func chansend(t *chantype, c *_race.Hchan, ep unsafe.Pointer, block bool, callerpc uintptr) bool {
	if _base.Raceenabled {
		_race.RaceReadObjectPC(t.elem, ep, callerpc, _base.FuncPC(chansend))
	}

	if c == nil {
		if !block {
			return false
		}
		_base.Gopark(nil, nil, "chan send (nil chan)", _base.TraceEvGoStop, 2)
		_base.Throw("unreachable")
	}

	if debugChan {
		print("chansend: chan=", c, "\n")
	}

	if _base.Raceenabled {
		_race.Racereadpc(unsafe.Pointer(c), callerpc, _base.FuncPC(chansend))
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
	if !block && c.Closed == 0 && ((c.Dataqsiz == 0 && c.Recvq.First == nil) ||
		(c.Dataqsiz > 0 && c.Qcount == c.Dataqsiz)) {
		return false
	}

	var t0 int64
	if _gc.Blockprofilerate > 0 {
		t0 = _base.Cputicks()
	}

	_base.Lock(&c.Lock)
	if c.Closed != 0 {
		_base.Unlock(&c.Lock)
		panic("send on closed channel")
	}

	if c.Dataqsiz == 0 { // synchronous channel
		sg := c.Recvq.Dequeue()
		if sg != nil { // found a waiting receiver
			if _base.Raceenabled {
				_race.Racesync(c, sg)
			}
			_base.Unlock(&c.Lock)

			recvg := sg.G
			if sg.Elem != nil {
				syncsend(c, sg, ep)
			}
			recvg.Param = unsafe.Pointer(sg)
			if sg.Releasetime != 0 {
				sg.Releasetime = _base.Cputicks()
			}
			_gc.Goready(recvg, 3)
			return true
		}

		if !block {
			_base.Unlock(&c.Lock)
			return false
		}

		// no receiver available: block on this channel.
		gp := _base.Getg()
		mysg := _gc.AcquireSudog()
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
		c.Sendq.Enqueue(mysg)
		_base.Goparkunlock(&c.Lock, "chan send", _base.TraceEvGoBlockSend, 3)

		// someone woke us up.
		if mysg != gp.Waiting {
			_base.Throw("G waiting list is corrupted!")
		}
		gp.Waiting = nil
		if gp.Param == nil {
			if c.Closed == 0 {
				_base.Throw("chansend: spurious wakeup")
			}
			panic("send on closed channel")
		}
		gp.Param = nil
		if mysg.Releasetime > 0 {
			_gc.Blockevent(int64(mysg.Releasetime)-t0, 2)
		}
		_gc.ReleaseSudog(mysg)
		return true
	}

	// asynchronous channel
	// wait for some space to write our data
	var t1 int64
	for futile := byte(0); c.Qcount >= c.Dataqsiz; futile = _base.TraceFutileWakeup {
		if !block {
			_base.Unlock(&c.Lock)
			return false
		}
		gp := _base.Getg()
		mysg := _gc.AcquireSudog()
		mysg.Releasetime = 0
		if t0 != 0 {
			mysg.Releasetime = -1
		}
		mysg.G = gp
		mysg.Elem = nil
		mysg.Selectdone = nil
		c.Sendq.Enqueue(mysg)
		_base.Goparkunlock(&c.Lock, "chan send", _base.TraceEvGoBlockSend|futile, 3)

		// someone woke us up - try again
		if mysg.Releasetime > 0 {
			t1 = mysg.Releasetime
		}
		_gc.ReleaseSudog(mysg)
		_base.Lock(&c.Lock)
		if c.Closed != 0 {
			_base.Unlock(&c.Lock)
			panic("send on closed channel")
		}
	}

	// write our data into the channel buffer
	if _base.Raceenabled {
		_iface.Raceacquire(_race.Chanbuf(c, c.Sendx))
		_race.Racerelease(_race.Chanbuf(c, c.Sendx))
	}
	_iface.Typedmemmove(c.Elemtype, _race.Chanbuf(c, c.Sendx), ep)
	c.Sendx++
	if c.Sendx == c.Dataqsiz {
		c.Sendx = 0
	}
	c.Qcount++

	// wake up a waiting receiver
	sg := c.Recvq.Dequeue()
	if sg != nil {
		recvg := sg.G
		_base.Unlock(&c.Lock)
		if sg.Releasetime != 0 {
			sg.Releasetime = _base.Cputicks()
		}
		_gc.Goready(recvg, 3)
	} else {
		_base.Unlock(&c.Lock)
	}
	if t1 > 0 {
		_gc.Blockevent(t1-t0, 2)
	}
	return true
}

func syncsend(c *_race.Hchan, sg *_base.Sudog, elem unsafe.Pointer) {
	// Send on unbuffered channel is the only operation
	// in the entire runtime where one goroutine
	// writes to the stack of another goroutine. The GC assumes that
	// stack writes only happen when the goroutine is running and are
	// only done by that goroutine. Using a write barrier is sufficient to
	// make up for violating that assumption, but the write barrier has to work.
	// typedmemmove will call heapBitsBulkBarrier, but the target bytes
	// are not in the heap, so that will not help. We arrange to call
	// memmove and typeBitsBulkBarrier instead.
	_base.Memmove(sg.Elem, elem, c.Elemtype.Size)
	typeBitsBulkBarrier(c.Elemtype, uintptr(sg.Elem), c.Elemtype.Size)
	sg.Elem = nil
}

func closechan(c *_race.Hchan) {
	if c == nil {
		panic("close of nil channel")
	}

	_base.Lock(&c.Lock)
	if c.Closed != 0 {
		_base.Unlock(&c.Lock)
		panic("close of closed channel")
	}

	if _base.Raceenabled {
		callerpc := _base.Getcallerpc(unsafe.Pointer(&c))
		_race.Racewritepc(unsafe.Pointer(c), callerpc, _base.FuncPC(closechan))
		_race.Racerelease(unsafe.Pointer(c))
	}

	c.Closed = 1

	// release all readers
	for {
		sg := c.Recvq.Dequeue()
		if sg == nil {
			break
		}
		gp := sg.G
		sg.Elem = nil
		gp.Param = nil
		if sg.Releasetime != 0 {
			sg.Releasetime = _base.Cputicks()
		}
		_gc.Goready(gp, 3)
	}

	// release all writers
	for {
		sg := c.Sendq.Dequeue()
		if sg == nil {
			break
		}
		gp := sg.G
		sg.Elem = nil
		gp.Param = nil
		if sg.Releasetime != 0 {
			sg.Releasetime = _base.Cputicks()
		}
		_gc.Goready(gp, 3)
	}
	_base.Unlock(&c.Lock)
}

// entry points for <- c from compiled code
//go:nosplit
func chanrecv1(t *chantype, c *_race.Hchan, elem unsafe.Pointer) {
	chanrecv(t, c, elem, true)
}

//go:nosplit
func chanrecv2(t *chantype, c *_race.Hchan, elem unsafe.Pointer) (received bool) {
	_, received = chanrecv(t, c, elem, true)
	return
}

// chanrecv receives on channel c and writes the received data to ep.
// ep may be nil, in which case received data is ignored.
// If block == false and no elements are available, returns (false, false).
// Otherwise, if c is closed, zeros *ep and returns (true, false).
// Otherwise, fills in *ep with an element and returns (true, true).
func chanrecv(t *chantype, c *_race.Hchan, ep unsafe.Pointer, block bool) (selected, received bool) {
	// raceenabled: don't need to check ep, as it is always on the stack.

	if debugChan {
		print("chanrecv: chan=", c, "\n")
	}

	if c == nil {
		if !block {
			return
		}
		_base.Gopark(nil, nil, "chan receive (nil chan)", _base.TraceEvGoStop, 2)
		_base.Throw("unreachable")
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
	if !block && (c.Dataqsiz == 0 && c.Sendq.First == nil ||
		c.Dataqsiz > 0 && _iface.Atomicloaduint(&c.Qcount) == 0) &&
		_base.Atomicload(&c.Closed) == 0 {
		return
	}

	var t0 int64
	if _gc.Blockprofilerate > 0 {
		t0 = _base.Cputicks()
	}

	_base.Lock(&c.Lock)
	if c.Dataqsiz == 0 { // synchronous channel
		if c.Closed != 0 {
			return recvclosed(c, ep)
		}

		sg := c.Sendq.Dequeue()
		if sg != nil {
			if _base.Raceenabled {
				_race.Racesync(c, sg)
			}
			_base.Unlock(&c.Lock)

			if ep != nil {
				_iface.Typedmemmove(c.Elemtype, ep, sg.Elem)
			}
			sg.Elem = nil
			gp := sg.G
			gp.Param = unsafe.Pointer(sg)
			if sg.Releasetime != 0 {
				sg.Releasetime = _base.Cputicks()
			}
			_gc.Goready(gp, 3)
			selected = true
			received = true
			return
		}

		if !block {
			_base.Unlock(&c.Lock)
			return
		}

		// no sender available: block on this channel.
		gp := _base.Getg()
		mysg := _gc.AcquireSudog()
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
		c.Recvq.Enqueue(mysg)
		_base.Goparkunlock(&c.Lock, "chan receive", _base.TraceEvGoBlockRecv, 3)

		// someone woke us up
		if mysg != gp.Waiting {
			_base.Throw("G waiting list is corrupted!")
		}
		gp.Waiting = nil
		if mysg.Releasetime > 0 {
			_gc.Blockevent(mysg.Releasetime-t0, 2)
		}
		haveData := gp.Param != nil
		gp.Param = nil
		_gc.ReleaseSudog(mysg)

		if haveData {
			// a sender sent us some data. It already wrote to ep.
			selected = true
			received = true
			return
		}

		_base.Lock(&c.Lock)
		if c.Closed == 0 {
			_base.Throw("chanrecv: spurious wakeup")
		}
		return recvclosed(c, ep)
	}

	// asynchronous channel
	// wait for some data to appear
	var t1 int64
	for futile := byte(0); c.Qcount <= 0; futile = _base.TraceFutileWakeup {
		if c.Closed != 0 {
			selected, received = recvclosed(c, ep)
			if t1 > 0 {
				_gc.Blockevent(t1-t0, 2)
			}
			return
		}

		if !block {
			_base.Unlock(&c.Lock)
			return
		}

		// wait for someone to send an element
		gp := _base.Getg()
		mysg := _gc.AcquireSudog()
		mysg.Releasetime = 0
		if t0 != 0 {
			mysg.Releasetime = -1
		}
		mysg.Elem = nil
		mysg.G = gp
		mysg.Selectdone = nil

		c.Recvq.Enqueue(mysg)
		_base.Goparkunlock(&c.Lock, "chan receive", _base.TraceEvGoBlockRecv|futile, 3)

		// someone woke us up - try again
		if mysg.Releasetime > 0 {
			t1 = mysg.Releasetime
		}
		_gc.ReleaseSudog(mysg)
		_base.Lock(&c.Lock)
	}

	if _base.Raceenabled {
		_iface.Raceacquire(_race.Chanbuf(c, c.Recvx))
		_race.Racerelease(_race.Chanbuf(c, c.Recvx))
	}
	if ep != nil {
		_iface.Typedmemmove(c.Elemtype, ep, _race.Chanbuf(c, c.Recvx))
	}
	_base.Memclr(_race.Chanbuf(c, c.Recvx), uintptr(c.Elemsize))

	c.Recvx++
	if c.Recvx == c.Dataqsiz {
		c.Recvx = 0
	}
	c.Qcount--

	// ping a sender now that there is space
	sg := c.Sendq.Dequeue()
	if sg != nil {
		gp := sg.G
		_base.Unlock(&c.Lock)
		if sg.Releasetime != 0 {
			sg.Releasetime = _base.Cputicks()
		}
		_gc.Goready(gp, 3)
	} else {
		_base.Unlock(&c.Lock)
	}

	if t1 > 0 {
		_gc.Blockevent(t1-t0, 2)
	}
	selected = true
	received = true
	return
}

// recvclosed is a helper function for chanrecv.  Handles cleanup
// when the receiver encounters a closed channel.
// Caller must hold c.lock, recvclosed will release the lock.
func recvclosed(c *_race.Hchan, ep unsafe.Pointer) (selected, recevied bool) {
	if _base.Raceenabled {
		_iface.Raceacquire(unsafe.Pointer(c))
	}
	_base.Unlock(&c.Lock)
	if ep != nil {
		_base.Memclr(ep, uintptr(c.Elemsize))
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
func selectnbsend(t *chantype, c *_race.Hchan, elem unsafe.Pointer) (selected bool) {
	return chansend(t, c, elem, false, _base.Getcallerpc(unsafe.Pointer(&t)))
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
func selectnbrecv(t *chantype, elem unsafe.Pointer, c *_race.Hchan) (selected bool) {
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
func selectnbrecv2(t *chantype, elem unsafe.Pointer, received *bool, c *_race.Hchan) (selected bool) {
	// TODO(khr): just return 2 values from this function, now that it is in Go.
	selected, *received = chanrecv(t, c, elem, false)
	return
}

//go:linkname reflect_chansend reflect.chansend
func reflect_chansend(t *chantype, c *_race.Hchan, elem unsafe.Pointer, nb bool) (selected bool) {
	return chansend(t, c, elem, !nb, _base.Getcallerpc(unsafe.Pointer(&t)))
}

//go:linkname reflect_chanrecv reflect.chanrecv
func reflect_chanrecv(t *chantype, c *_race.Hchan, nb bool, elem unsafe.Pointer) (selected bool, received bool) {
	return chanrecv(t, c, elem, !nb)
}

//go:linkname reflect_chanlen reflect.chanlen
func reflect_chanlen(c *_race.Hchan) int {
	if c == nil {
		return 0
	}
	return int(c.Qcount)
}

//go:linkname reflect_chancap reflect.chancap
func reflect_chancap(c *_race.Hchan) int {
	if c == nil {
		return 0
	}
	return int(c.Dataqsiz)
}

//go:linkname reflect_chanclose reflect.chanclose
func reflect_chanclose(c *_race.Hchan) {
	closechan(c)
}

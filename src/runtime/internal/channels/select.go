// Copyright 2009 The Go Authors. All rights reserved.
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
	DebugSelect = false
)

var (
	chansendpc = _lock.FuncPC(Chansend)
	chanrecvpc = _lock.FuncPC(chanrecv)
)

func Selectsize(size uintptr) uintptr {
	selsize := unsafe.Sizeof(Select{}) +
		(size-1)*unsafe.Sizeof(Select{}.Scase[0]) +
		size*unsafe.Sizeof(*Select{}.lockorder) +
		size*unsafe.Sizeof(*Select{}.pollorder)
	return _lock.Round(selsize, _lock.Int64Align)
}

func Newselect(sel *Select, selsize int64, size int32) {
	if selsize != int64(Selectsize(uintptr(size))) {
		print("runtime: bad select size ", selsize, ", want ", Selectsize(uintptr(size)), "\n")
		_lock.Throw("bad select size")
	}
	sel.Tcase = uint16(size)
	sel.Ncase = 0
	sel.lockorder = (**Hchan)(_core.Add(unsafe.Pointer(&sel.Scase), uintptr(size)*unsafe.Sizeof(Select{}.Scase[0])))
	sel.pollorder = (*uint16)(_core.Add(unsafe.Pointer(sel.lockorder), uintptr(size)*unsafe.Sizeof(*Select{}.lockorder)))

	if DebugSelect {
		print("newselect s=", sel, " size=", size, "\n")
	}
}

//go:nosplit
func selectrecv(sel *Select, c *Hchan, elem unsafe.Pointer) (selected bool) {
	// nil cases do not compete
	if c != nil {
		SelectrecvImpl(sel, c, _lock.Getcallerpc(unsafe.Pointer(&sel)), elem, nil, uintptr(unsafe.Pointer(&selected))-uintptr(unsafe.Pointer(&sel)))
	}
	return
}

//go:nosplit
func selectrecv2(sel *Select, c *Hchan, elem unsafe.Pointer, received *bool) (selected bool) {
	// nil cases do not compete
	if c != nil {
		SelectrecvImpl(sel, c, _lock.Getcallerpc(unsafe.Pointer(&sel)), elem, received, uintptr(unsafe.Pointer(&selected))-uintptr(unsafe.Pointer(&sel)))
	}
	return
}

func SelectrecvImpl(sel *Select, c *Hchan, pc uintptr, elem unsafe.Pointer, received *bool, so uintptr) {
	i := sel.Ncase
	if i >= sel.Tcase {
		_lock.Throw("selectrecv: too many cases")
	}
	sel.Ncase = i + 1
	cas := (*Scase)(_core.Add(unsafe.Pointer(&sel.Scase), uintptr(i)*unsafe.Sizeof(sel.Scase[0])))
	cas.Pc = pc
	cas.Chan = c
	cas.So = uint16(so)
	cas.Kind = CaseRecv
	cas.Elem = elem
	cas.receivedp = received

	if DebugSelect {
		print("selectrecv s=", sel, " pc=", _core.Hex(cas.Pc), " chan=", cas.Chan, " so=", cas.So, "\n")
	}
}

func sellock(sel *Select) {
	lockslice := _sched.SliceStruct{unsafe.Pointer(sel.lockorder), int(sel.Ncase), int(sel.Ncase)}
	lockorder := *(*[]*Hchan)(unsafe.Pointer(&lockslice))
	var c *Hchan
	for _, c0 := range lockorder {
		if c0 != nil && c0 != c {
			c = c0
			_lock.Lock(&c.Lock)
		}
	}
}

func selunlock(sel *Select) {
	// We must be very careful here to not touch sel after we have unlocked
	// the last lock, because sel can be freed right after the last unlock.
	// Consider the following situation.
	// First M calls runtime·park() in runtime·selectgo() passing the sel.
	// Once runtime·park() has unlocked the last lock, another M makes
	// the G that calls select runnable again and schedules it for execution.
	// When the G runs on another M, it locks all the locks and frees sel.
	// Now if the first M touches sel, it will access freed memory.
	n := int(sel.Ncase)
	r := 0
	lockslice := _sched.SliceStruct{unsafe.Pointer(sel.lockorder), n, n}
	lockorder := *(*[]*Hchan)(unsafe.Pointer(&lockslice))
	// skip the default case
	if n > 0 && lockorder[0] == nil {
		r = 1
	}
	for i := n - 1; i >= r; i-- {
		c := lockorder[i]
		if i > 0 && c == lockorder[i-1] {
			continue // will unlock it on the next iteration
		}
		_lock.Unlock(&c.Lock)
	}
}

func selparkcommit(gp *_core.G, sel unsafe.Pointer) bool {
	selunlock((*Select)(sel))
	return true
}

// overwrites return pc on stack to signal which case of the select
// to run, so cannot appear at the top of a split stack.
//go:nosplit
func selectgo(sel *Select) {
	pc, offset := SelectgoImpl(sel)
	*(*bool)(_core.Add(unsafe.Pointer(&sel), uintptr(offset))) = true
	setcallerpc(unsafe.Pointer(&sel), pc)
}

// selectgoImpl returns scase.pc and scase.so for the select
// case which fired.
func SelectgoImpl(sel *Select) (uintptr, uint16) {
	if DebugSelect {
		print("select: sel=", sel, "\n")
	}

	scaseslice := _sched.SliceStruct{unsafe.Pointer(&sel.Scase), int(sel.Ncase), int(sel.Ncase)}
	scases := *(*[]Scase)(unsafe.Pointer(&scaseslice))

	var t0 int64
	if _sem.Blockprofilerate > 0 {
		t0 = _sched.Cputicks()
		for i := 0; i < int(sel.Ncase); i++ {
			scases[i].releasetime = -1
		}
	}

	// The compiler rewrites selects that statically have
	// only 0 or 1 cases plus default into simpler constructs.
	// The only way we can end up with such small sel.ncase
	// values here is for a larger select in which most channels
	// have been nilled out.  The general code handles those
	// cases correctly, and they are rare enough not to bother
	// optimizing (and needing to test).

	// generate permuted order
	pollslice := _sched.SliceStruct{unsafe.Pointer(sel.pollorder), int(sel.Ncase), int(sel.Ncase)}
	pollorder := *(*[]uint16)(unsafe.Pointer(&pollslice))
	for i := 0; i < int(sel.Ncase); i++ {
		pollorder[i] = uint16(i)
	}
	for i := 1; i < int(sel.Ncase); i++ {
		o := pollorder[i]
		j := int(_lock.Fastrand1()) % (i + 1)
		pollorder[i] = pollorder[j]
		pollorder[j] = o
	}

	// sort the cases by Hchan address to get the locking order.
	// simple heap sort, to guarantee n log n time and constant stack footprint.
	lockslice := _sched.SliceStruct{unsafe.Pointer(sel.lockorder), int(sel.Ncase), int(sel.Ncase)}
	lockorder := *(*[]*Hchan)(unsafe.Pointer(&lockslice))
	for i := 0; i < int(sel.Ncase); i++ {
		j := i
		c := scases[j].Chan
		for j > 0 && lockorder[(j-1)/2].sortkey() < c.sortkey() {
			k := (j - 1) / 2
			lockorder[j] = lockorder[k]
			j = k
		}
		lockorder[j] = c
	}
	for i := int(sel.Ncase) - 1; i >= 0; i-- {
		c := lockorder[i]
		lockorder[i] = lockorder[0]
		j := 0
		for {
			k := j*2 + 1
			if k >= i {
				break
			}
			if k+1 < i && lockorder[k].sortkey() < lockorder[k+1].sortkey() {
				k++
			}
			if c.sortkey() < lockorder[k].sortkey() {
				lockorder[j] = lockorder[k]
				j = k
				continue
			}
			break
		}
		lockorder[j] = c
	}
	/*
		for i := 0; i+1 < int(sel.ncase); i++ {
			if lockorder[i].sortkey() > lockorder[i+1].sortkey() {
				print("i=", i, " x=", lockorder[i], " y=", lockorder[i+1], "\n")
				throw("select: broken sort")
			}
		}
	*/

	// lock all the channels involved in the select
	sellock(sel)

	var (
		gp     *_core.G
		done   uint32
		sg     *_core.Sudog
		c      *Hchan
		k      *Scase
		sglist *_core.Sudog
		sgnext *_core.Sudog
	)

loop:
	// pass 1 - look for something already waiting
	var dfl *Scase
	var cas *Scase
	for i := 0; i < int(sel.Ncase); i++ {
		cas = &scases[pollorder[i]]
		c = cas.Chan

		switch cas.Kind {
		case CaseRecv:
			if c.Dataqsiz > 0 {
				if c.Qcount > 0 {
					goto asyncrecv
				}
			} else {
				sg = c.Sendq.Dequeue()
				if sg != nil {
					goto syncrecv
				}
			}
			if c.Closed != 0 {
				goto rclose
			}

		case CaseSend:
			if _sched.Raceenabled {
				Racereadpc(unsafe.Pointer(c), cas.Pc, chansendpc)
			}
			if c.Closed != 0 {
				goto sclose
			}
			if c.Dataqsiz > 0 {
				if c.Qcount < c.Dataqsiz {
					goto asyncsend
				}
			} else {
				sg = c.Recvq.Dequeue()
				if sg != nil {
					goto syncsend
				}
			}

		case CaseDefault:
			dfl = cas
		}
	}

	if dfl != nil {
		selunlock(sel)
		cas = dfl
		goto retc
	}

	// pass 2 - enqueue on all chans
	gp = _core.Getg()
	done = 0
	for i := 0; i < int(sel.Ncase); i++ {
		cas = &scases[pollorder[i]]
		c = cas.Chan
		sg := _sem.AcquireSudog()
		sg.G = gp
		// Note: selectdone is adjusted for stack copies in stack.c:adjustsudogs
		sg.Selectdone = (*uint32)(_core.Noescape(unsafe.Pointer(&done)))
		sg.Elem = cas.Elem
		sg.Releasetime = 0
		if t0 != 0 {
			sg.Releasetime = -1
		}
		sg.Waitlink = gp.Waiting
		gp.Waiting = sg

		switch cas.Kind {
		case CaseRecv:
			c.Recvq.enqueue(sg)

		case CaseSend:
			c.Sendq.enqueue(sg)
		}
	}

	// wait for someone to wake us up
	gp.Param = nil
	_sched.Gopark(selparkcommit, unsafe.Pointer(sel), "select", _sched.TraceEvGoBlockSelect)

	// someone woke us up
	sellock(sel)
	sg = (*_core.Sudog)(gp.Param)
	gp.Param = nil

	// pass 3 - dequeue from unsuccessful chans
	// otherwise they stack up on quiet channels
	// record the successful case, if any.
	// We singly-linked up the SudoGs in case order, so when
	// iterating through the linked list they are in reverse order.
	cas = nil
	sglist = gp.Waiting
	// Clear all elem before unlinking from gp.waiting.
	for sg1 := gp.Waiting; sg1 != nil; sg1 = sg1.Waitlink {
		sg1.Selectdone = nil
		sg1.Elem = nil
	}
	gp.Waiting = nil
	for i := int(sel.Ncase) - 1; i >= 0; i-- {
		k = &scases[pollorder[i]]
		if sglist.Releasetime > 0 {
			k.releasetime = sglist.Releasetime
		}
		if sg == sglist {
			// sg has already been dequeued by the G that woke us up.
			cas = k
		} else {
			c = k.Chan
			if k.Kind == CaseSend {
				c.Sendq.dequeueSudoG(sglist)
			} else {
				c.Recvq.dequeueSudoG(sglist)
			}
		}
		sgnext = sglist.Waitlink
		sglist.Waitlink = nil
		_sem.ReleaseSudog(sglist)
		sglist = sgnext
	}

	if cas == nil {
		goto loop
	}

	c = cas.Chan

	if c.Dataqsiz > 0 {
		_lock.Throw("selectgo: shouldn't happen")
	}

	if DebugSelect {
		print("wait-return: sel=", sel, " c=", c, " cas=", cas, " kind=", cas.Kind, "\n")
	}

	if cas.Kind == CaseRecv {
		if cas.receivedp != nil {
			*cas.receivedp = true
		}
	}

	if _sched.Raceenabled {
		if cas.Kind == CaseRecv && cas.Elem != nil {
			raceWriteObjectPC(c.Elemtype, cas.Elem, cas.Pc, chanrecvpc)
		} else if cas.Kind == CaseSend {
			RaceReadObjectPC(c.Elemtype, cas.Elem, cas.Pc, chansendpc)
		}
	}

	selunlock(sel)
	goto retc

asyncrecv:
	// can receive from buffer
	if _sched.Raceenabled {
		if cas.Elem != nil {
			raceWriteObjectPC(c.Elemtype, cas.Elem, cas.Pc, chanrecvpc)
		}
		_sched.Raceacquire(chanbuf(c, c.recvx))
		Racerelease(chanbuf(c, c.recvx))
	}
	if cas.receivedp != nil {
		*cas.receivedp = true
	}
	if cas.Elem != nil {
		Typedmemmove(c.Elemtype, cas.Elem, chanbuf(c, c.recvx))
	}
	_core.Memclr(chanbuf(c, c.recvx), uintptr(c.Elemsize))
	c.recvx++
	if c.recvx == c.Dataqsiz {
		c.recvx = 0
	}
	c.Qcount--
	sg = c.Sendq.Dequeue()
	if sg != nil {
		gp = sg.G
		selunlock(sel)
		if sg.Releasetime != 0 {
			sg.Releasetime = _sched.Cputicks()
		}
		_sched.Goready(gp)
	} else {
		selunlock(sel)
	}
	goto retc

asyncsend:
	// can send to buffer
	if _sched.Raceenabled {
		_sched.Raceacquire(chanbuf(c, c.sendx))
		Racerelease(chanbuf(c, c.sendx))
		RaceReadObjectPC(c.Elemtype, cas.Elem, cas.Pc, chansendpc)
	}
	Typedmemmove(c.Elemtype, chanbuf(c, c.sendx), cas.Elem)
	c.sendx++
	if c.sendx == c.Dataqsiz {
		c.sendx = 0
	}
	c.Qcount++
	sg = c.Recvq.Dequeue()
	if sg != nil {
		gp = sg.G
		selunlock(sel)
		if sg.Releasetime != 0 {
			sg.Releasetime = _sched.Cputicks()
		}
		_sched.Goready(gp)
	} else {
		selunlock(sel)
	}
	goto retc

syncrecv:
	// can receive from sleeping sender (sg)
	if _sched.Raceenabled {
		if cas.Elem != nil {
			raceWriteObjectPC(c.Elemtype, cas.Elem, cas.Pc, chanrecvpc)
		}
		racesync(c, sg)
	}
	selunlock(sel)
	if DebugSelect {
		print("syncrecv: sel=", sel, " c=", c, "\n")
	}
	if cas.receivedp != nil {
		*cas.receivedp = true
	}
	if cas.Elem != nil {
		Typedmemmove(c.Elemtype, cas.Elem, sg.Elem)
	}
	sg.Elem = nil
	gp = sg.G
	gp.Param = unsafe.Pointer(sg)
	if sg.Releasetime != 0 {
		sg.Releasetime = _sched.Cputicks()
	}
	_sched.Goready(gp)
	goto retc

rclose:
	// read at end of closed channel
	selunlock(sel)
	if cas.receivedp != nil {
		*cas.receivedp = false
	}
	if cas.Elem != nil {
		_core.Memclr(cas.Elem, uintptr(c.Elemsize))
	}
	if _sched.Raceenabled {
		_sched.Raceacquire(unsafe.Pointer(c))
	}
	goto retc

syncsend:
	// can send to sleeping receiver (sg)
	if _sched.Raceenabled {
		RaceReadObjectPC(c.Elemtype, cas.Elem, cas.Pc, chansendpc)
		racesync(c, sg)
	}
	selunlock(sel)
	if DebugSelect {
		print("syncsend: sel=", sel, " c=", c, "\n")
	}
	if sg.Elem != nil {
		Typedmemmove(c.Elemtype, sg.Elem, cas.Elem)
	}
	sg.Elem = nil
	gp = sg.G
	gp.Param = unsafe.Pointer(sg)
	if sg.Releasetime != 0 {
		sg.Releasetime = _sched.Cputicks()
	}
	_sched.Goready(gp)

retc:
	if cas.releasetime > 0 {
		_sem.Blockevent(cas.releasetime-t0, 2)
	}
	return cas.Pc, cas.So

sclose:
	// send on closed channel
	selunlock(sel)
	panic("send on closed channel")
}

func (c *Hchan) sortkey() uintptr {
	// TODO(khr): if we have a moving garbage collector, we'll need to
	// change this function.
	return uintptr(unsafe.Pointer(c))
}

// These values must match ../reflect/value.go:/SelectDir.
type SelectDir int

func (q *waitq) dequeueSudoG(sgp *_core.Sudog) {
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
		q.first = y
		sgp.Next = nil
		return
	}

	// x==y==nil.  Either sgp is the only element in the queue,
	// or it has already been removed.  Use q.first to disambiguate.
	if q.first == sgp {
		q.first = nil
		q.last = nil
	}
}

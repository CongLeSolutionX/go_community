// Copyright 2009 The Go Authors. All rights reserved.
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
	debugSelect = false

	// scase.kind
	caseRecv = iota
	caseSend
	caseDefault
)

// Select statement header.
// Known to compiler.
// Changes here must also be made in src/cmd/internal/gc/select.go's selecttype.
type hselect struct {
	tcase     uint16        // total count of scase[]
	ncase     uint16        // currently filled scase[]
	pollorder *uint16       // case poll order
	lockorder **_race.Hchan // channel lock order
	scase     [1]scase      // one per case (in order of appearance)
}

// Select case descriptor.
// Known to compiler.
// Changes here must also be made in src/cmd/internal/gc/select.go's selecttype.
type scase struct {
	elem        unsafe.Pointer // data element
	c           *_race.Hchan   // chan
	pc          uintptr        // return pc
	kind        uint16
	so          uint16 // vararg of selected bool
	receivedp   *bool  // pointer to received bool (recv2)
	releasetime int64
}

var (
	chansendpc = _base.FuncPC(chansend)
	chanrecvpc = _base.FuncPC(chanrecv)
)

func selectsize(size uintptr) uintptr {
	selsize := unsafe.Sizeof(hselect{}) +
		(size-1)*unsafe.Sizeof(hselect{}.scase[0]) +
		size*unsafe.Sizeof(*hselect{}.lockorder) +
		size*unsafe.Sizeof(*hselect{}.pollorder)
	return _base.Round(selsize, _base.Int64Align)
}

func newselect(sel *hselect, selsize int64, size int32) {
	if selsize != int64(selectsize(uintptr(size))) {
		print("runtime: bad select size ", selsize, ", want ", selectsize(uintptr(size)), "\n")
		_base.Throw("bad select size")
	}
	sel.tcase = uint16(size)
	sel.ncase = 0
	sel.lockorder = (**_race.Hchan)(_base.Add(unsafe.Pointer(&sel.scase), uintptr(size)*unsafe.Sizeof(hselect{}.scase[0])))
	sel.pollorder = (*uint16)(_base.Add(unsafe.Pointer(sel.lockorder), uintptr(size)*unsafe.Sizeof(*hselect{}.lockorder)))

	if debugSelect {
		print("newselect s=", sel, " size=", size, "\n")
	}
}

//go:nosplit
func selectsend(sel *hselect, c *_race.Hchan, elem unsafe.Pointer) (selected bool) {
	// nil cases do not compete
	if c != nil {
		selectsendImpl(sel, c, _base.Getcallerpc(unsafe.Pointer(&sel)), elem, uintptr(unsafe.Pointer(&selected))-uintptr(unsafe.Pointer(&sel)))
	}
	return
}

// cut in half to give stack a chance to split
func selectsendImpl(sel *hselect, c *_race.Hchan, pc uintptr, elem unsafe.Pointer, so uintptr) {
	i := sel.ncase
	if i >= sel.tcase {
		_base.Throw("selectsend: too many cases")
	}
	sel.ncase = i + 1
	cas := (*scase)(_base.Add(unsafe.Pointer(&sel.scase), uintptr(i)*unsafe.Sizeof(sel.scase[0])))

	cas.pc = pc
	cas.c = c
	cas.so = uint16(so)
	cas.kind = caseSend
	cas.elem = elem

	if debugSelect {
		print("selectsend s=", sel, " pc=", _base.Hex(cas.pc), " chan=", cas.c, " so=", cas.so, "\n")
	}
}

//go:nosplit
func selectrecv(sel *hselect, c *_race.Hchan, elem unsafe.Pointer) (selected bool) {
	// nil cases do not compete
	if c != nil {
		selectrecvImpl(sel, c, _base.Getcallerpc(unsafe.Pointer(&sel)), elem, nil, uintptr(unsafe.Pointer(&selected))-uintptr(unsafe.Pointer(&sel)))
	}
	return
}

//go:nosplit
func selectrecv2(sel *hselect, c *_race.Hchan, elem unsafe.Pointer, received *bool) (selected bool) {
	// nil cases do not compete
	if c != nil {
		selectrecvImpl(sel, c, _base.Getcallerpc(unsafe.Pointer(&sel)), elem, received, uintptr(unsafe.Pointer(&selected))-uintptr(unsafe.Pointer(&sel)))
	}
	return
}

func selectrecvImpl(sel *hselect, c *_race.Hchan, pc uintptr, elem unsafe.Pointer, received *bool, so uintptr) {
	i := sel.ncase
	if i >= sel.tcase {
		_base.Throw("selectrecv: too many cases")
	}
	sel.ncase = i + 1
	cas := (*scase)(_base.Add(unsafe.Pointer(&sel.scase), uintptr(i)*unsafe.Sizeof(sel.scase[0])))
	cas.pc = pc
	cas.c = c
	cas.so = uint16(so)
	cas.kind = caseRecv
	cas.elem = elem
	cas.receivedp = received

	if debugSelect {
		print("selectrecv s=", sel, " pc=", _base.Hex(cas.pc), " chan=", cas.c, " so=", cas.so, "\n")
	}
}

//go:nosplit
func selectdefault(sel *hselect) (selected bool) {
	selectdefaultImpl(sel, _base.Getcallerpc(unsafe.Pointer(&sel)), uintptr(unsafe.Pointer(&selected))-uintptr(unsafe.Pointer(&sel)))
	return
}

func selectdefaultImpl(sel *hselect, callerpc uintptr, so uintptr) {
	i := sel.ncase
	if i >= sel.tcase {
		_base.Throw("selectdefault: too many cases")
	}
	sel.ncase = i + 1
	cas := (*scase)(_base.Add(unsafe.Pointer(&sel.scase), uintptr(i)*unsafe.Sizeof(sel.scase[0])))
	cas.pc = callerpc
	cas.c = nil
	cas.so = uint16(so)
	cas.kind = caseDefault

	if debugSelect {
		print("selectdefault s=", sel, " pc=", _base.Hex(cas.pc), " so=", cas.so, "\n")
	}
}

func sellock(sel *hselect) {
	lockslice := _base.Slice{unsafe.Pointer(sel.lockorder), int(sel.ncase), int(sel.ncase)}
	lockorder := *(*[]*_race.Hchan)(unsafe.Pointer(&lockslice))
	var c *_race.Hchan
	for _, c0 := range lockorder {
		if c0 != nil && c0 != c {
			c = c0
			_base.Lock(&c.Lock)
		}
	}
}

func selunlock(sel *hselect) {
	// We must be very careful here to not touch sel after we have unlocked
	// the last lock, because sel can be freed right after the last unlock.
	// Consider the following situation.
	// First M calls runtime·park() in runtime·selectgo() passing the sel.
	// Once runtime·park() has unlocked the last lock, another M makes
	// the G that calls select runnable again and schedules it for execution.
	// When the G runs on another M, it locks all the locks and frees sel.
	// Now if the first M touches sel, it will access freed memory.
	n := int(sel.ncase)
	r := 0
	lockslice := _base.Slice{unsafe.Pointer(sel.lockorder), n, n}
	lockorder := *(*[]*_race.Hchan)(unsafe.Pointer(&lockslice))
	// skip the default case
	if n > 0 && lockorder[0] == nil {
		r = 1
	}
	for i := n - 1; i >= r; i-- {
		c := lockorder[i]
		if i > 0 && c == lockorder[i-1] {
			continue // will unlock it on the next iteration
		}
		_base.Unlock(&c.Lock)
	}
}

func selparkcommit(gp *_base.G, sel unsafe.Pointer) bool {
	selunlock((*hselect)(sel))
	return true
}

func block() {
	_base.Gopark(nil, nil, "select (no cases)", _base.TraceEvGoStop, 1) // forever
}

// overwrites return pc on stack to signal which case of the select
// to run, so cannot appear at the top of a split stack.
//go:nosplit
func selectgo(sel *hselect) {
	pc, offset := selectgoImpl(sel)
	*(*bool)(_base.Add(unsafe.Pointer(&sel), uintptr(offset))) = true
	setcallerpc(unsafe.Pointer(&sel), pc)
}

// selectgoImpl returns scase.pc and scase.so for the select
// case which fired.
func selectgoImpl(sel *hselect) (uintptr, uint16) {
	if debugSelect {
		print("select: sel=", sel, "\n")
	}

	scaseslice := _base.Slice{unsafe.Pointer(&sel.scase), int(sel.ncase), int(sel.ncase)}
	scases := *(*[]scase)(unsafe.Pointer(&scaseslice))

	var t0 int64
	if _gc.Blockprofilerate > 0 {
		t0 = _base.Cputicks()
		for i := 0; i < int(sel.ncase); i++ {
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
	pollslice := _base.Slice{unsafe.Pointer(sel.pollorder), int(sel.ncase), int(sel.ncase)}
	pollorder := *(*[]uint16)(unsafe.Pointer(&pollslice))
	for i := 1; i < int(sel.ncase); i++ {
		j := int(_base.Fastrand1()) % (i + 1)
		pollorder[i] = pollorder[j]
		pollorder[j] = uint16(i)
	}

	// sort the cases by Hchan address to get the locking order.
	// simple heap sort, to guarantee n log n time and constant stack footprint.
	lockslice := _base.Slice{unsafe.Pointer(sel.lockorder), int(sel.ncase), int(sel.ncase)}
	lockorder := *(*[]*_race.Hchan)(unsafe.Pointer(&lockslice))
	for i := 0; i < int(sel.ncase); i++ {
		j := i
		c := scases[j].c
		for j > 0 && lockorder[(j-1)/2].Sortkey() < c.Sortkey() {
			k := (j - 1) / 2
			lockorder[j] = lockorder[k]
			j = k
		}
		lockorder[j] = c
	}
	for i := int(sel.ncase) - 1; i >= 0; i-- {
		c := lockorder[i]
		lockorder[i] = lockorder[0]
		j := 0
		for {
			k := j*2 + 1
			if k >= i {
				break
			}
			if k+1 < i && lockorder[k].Sortkey() < lockorder[k+1].Sortkey() {
				k++
			}
			if c.Sortkey() < lockorder[k].Sortkey() {
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
		gp     *_base.G
		done   uint32
		sg     *_base.Sudog
		c      *_race.Hchan
		k      *scase
		sglist *_base.Sudog
		sgnext *_base.Sudog
		futile byte
	)

loop:
	// pass 1 - look for something already waiting
	var dfl *scase
	var cas *scase
	for i := 0; i < int(sel.ncase); i++ {
		cas = &scases[pollorder[i]]
		c = cas.c

		switch cas.kind {
		case caseRecv:
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

		case caseSend:
			if _base.Raceenabled {
				_race.Racereadpc(unsafe.Pointer(c), cas.pc, chansendpc)
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

		case caseDefault:
			dfl = cas
		}
	}

	if dfl != nil {
		selunlock(sel)
		cas = dfl
		goto retc
	}

	// pass 2 - enqueue on all chans
	gp = _base.Getg()
	done = 0
	for i := 0; i < int(sel.ncase); i++ {
		cas = &scases[pollorder[i]]
		c = cas.c
		sg := _gc.AcquireSudog()
		sg.G = gp
		// Note: selectdone is adjusted for stack copies in stack1.go:adjustsudogs
		sg.Selectdone = (*uint32)(_base.Noescape(unsafe.Pointer(&done)))
		sg.Elem = cas.elem
		sg.Releasetime = 0
		if t0 != 0 {
			sg.Releasetime = -1
		}
		sg.Waitlink = gp.Waiting
		gp.Waiting = sg

		switch cas.kind {
		case caseRecv:
			c.Recvq.Enqueue(sg)

		case caseSend:
			c.Sendq.Enqueue(sg)
		}
	}

	// wait for someone to wake us up
	gp.Param = nil
	_base.Gopark(selparkcommit, unsafe.Pointer(sel), "select", _base.TraceEvGoBlockSelect|futile, 2)

	// someone woke us up
	sellock(sel)
	sg = (*_base.Sudog)(gp.Param)
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
	for i := int(sel.ncase) - 1; i >= 0; i-- {
		k = &scases[pollorder[i]]
		if sglist.Releasetime > 0 {
			k.releasetime = sglist.Releasetime
		}
		if sg == sglist {
			// sg has already been dequeued by the G that woke us up.
			cas = k
		} else {
			c = k.c
			if k.kind == caseSend {
				c.Sendq.DequeueSudoG(sglist)
			} else {
				c.Recvq.DequeueSudoG(sglist)
			}
		}
		sgnext = sglist.Waitlink
		sglist.Waitlink = nil
		_gc.ReleaseSudog(sglist)
		sglist = sgnext
	}

	if cas == nil {
		futile = _base.TraceFutileWakeup
		goto loop
	}

	c = cas.c

	if c.Dataqsiz > 0 {
		_base.Throw("selectgo: shouldn't happen")
	}

	if debugSelect {
		print("wait-return: sel=", sel, " c=", c, " cas=", cas, " kind=", cas.kind, "\n")
	}

	if cas.kind == caseRecv {
		if cas.receivedp != nil {
			*cas.receivedp = true
		}
	}

	if _base.Raceenabled {
		if cas.kind == caseRecv && cas.elem != nil {
			_race.RaceWriteObjectPC(c.Elemtype, cas.elem, cas.pc, chanrecvpc)
		} else if cas.kind == caseSend {
			_race.RaceReadObjectPC(c.Elemtype, cas.elem, cas.pc, chansendpc)
		}
	}

	selunlock(sel)
	goto retc

asyncrecv:
	// can receive from buffer
	if _base.Raceenabled {
		if cas.elem != nil {
			_race.RaceWriteObjectPC(c.Elemtype, cas.elem, cas.pc, chanrecvpc)
		}
		_base.Raceacquire(_race.Chanbuf(c, c.Recvx))
		_race.Racerelease(_race.Chanbuf(c, c.Recvx))
	}
	if cas.receivedp != nil {
		*cas.receivedp = true
	}
	if cas.elem != nil {
		_iface.Typedmemmove(c.Elemtype, cas.elem, _race.Chanbuf(c, c.Recvx))
	}
	_base.Memclr(_race.Chanbuf(c, c.Recvx), uintptr(c.Elemsize))
	c.Recvx++
	if c.Recvx == c.Dataqsiz {
		c.Recvx = 0
	}
	c.Qcount--
	sg = c.Sendq.Dequeue()
	if sg != nil {
		gp = sg.G
		selunlock(sel)
		if sg.Releasetime != 0 {
			sg.Releasetime = _base.Cputicks()
		}
		_base.Goready(gp, 3)
	} else {
		selunlock(sel)
	}
	goto retc

asyncsend:
	// can send to buffer
	if _base.Raceenabled {
		_base.Raceacquire(_race.Chanbuf(c, c.Sendx))
		_race.Racerelease(_race.Chanbuf(c, c.Sendx))
		_race.RaceReadObjectPC(c.Elemtype, cas.elem, cas.pc, chansendpc)
	}
	_iface.Typedmemmove(c.Elemtype, _race.Chanbuf(c, c.Sendx), cas.elem)
	c.Sendx++
	if c.Sendx == c.Dataqsiz {
		c.Sendx = 0
	}
	c.Qcount++
	sg = c.Recvq.Dequeue()
	if sg != nil {
		gp = sg.G
		selunlock(sel)
		if sg.Releasetime != 0 {
			sg.Releasetime = _base.Cputicks()
		}
		_base.Goready(gp, 3)
	} else {
		selunlock(sel)
	}
	goto retc

syncrecv:
	// can receive from sleeping sender (sg)
	if _base.Raceenabled {
		if cas.elem != nil {
			_race.RaceWriteObjectPC(c.Elemtype, cas.elem, cas.pc, chanrecvpc)
		}
		_race.Racesync(c, sg)
	}
	selunlock(sel)
	if debugSelect {
		print("syncrecv: sel=", sel, " c=", c, "\n")
	}
	if cas.receivedp != nil {
		*cas.receivedp = true
	}
	if cas.elem != nil {
		_iface.Typedmemmove(c.Elemtype, cas.elem, sg.Elem)
	}
	sg.Elem = nil
	gp = sg.G
	gp.Param = unsafe.Pointer(sg)
	if sg.Releasetime != 0 {
		sg.Releasetime = _base.Cputicks()
	}
	_base.Goready(gp, 3)
	goto retc

rclose:
	// read at end of closed channel
	selunlock(sel)
	if cas.receivedp != nil {
		*cas.receivedp = false
	}
	if cas.elem != nil {
		_base.Memclr(cas.elem, uintptr(c.Elemsize))
	}
	if _base.Raceenabled {
		_base.Raceacquire(unsafe.Pointer(c))
	}
	goto retc

syncsend:
	// can send to sleeping receiver (sg)
	if _base.Raceenabled {
		_race.RaceReadObjectPC(c.Elemtype, cas.elem, cas.pc, chansendpc)
		_race.Racesync(c, sg)
	}
	selunlock(sel)
	if debugSelect {
		print("syncsend: sel=", sel, " c=", c, "\n")
	}
	if sg.Elem != nil {
		syncsend(c, sg, cas.elem)
	}
	sg.Elem = nil
	gp = sg.G
	gp.Param = unsafe.Pointer(sg)
	if sg.Releasetime != 0 {
		sg.Releasetime = _base.Cputicks()
	}
	_base.Goready(gp, 3)

retc:
	if cas.releasetime > 0 {
		_gc.Blockevent(cas.releasetime-t0, 2)
	}
	return cas.pc, cas.so

sclose:
	// send on closed channel
	selunlock(sel)
	panic("send on closed channel")
}

// A runtimeSelect is a single case passed to rselect.
// This must match ../reflect/value.go:/runtimeSelect
type runtimeSelect struct {
	dir selectDir
	typ unsafe.Pointer // channel type (not used here)
	ch  *_race.Hchan   // channel
	val unsafe.Pointer // ptr to data (SendDir) or ptr to receive buffer (RecvDir)
}

// These values must match ../reflect/value.go:/SelectDir.
type selectDir int

const (
	_             selectDir = iota
	selectSend              // case Chan <- Send
	selectRecv              // case <-Chan:
	selectDefault           // default
)

//go:linkname reflect_rselect reflect.rselect
func reflect_rselect(cases []runtimeSelect) (chosen int, recvOK bool) {
	// flagNoScan is safe here, because all objects are also referenced from cases.
	size := selectsize(uintptr(len(cases)))
	sel := (*hselect)(_iface.Mallocgc(size, nil, _base.XFlagNoScan))
	newselect(sel, int64(size), int32(len(cases)))
	r := new(bool)
	for i := range cases {
		rc := &cases[i]
		switch rc.dir {
		case selectDefault:
			selectdefaultImpl(sel, uintptr(i), 0)
		case selectSend:
			if rc.ch == nil {
				break
			}
			selectsendImpl(sel, rc.ch, uintptr(i), rc.val, 0)
		case selectRecv:
			if rc.ch == nil {
				break
			}
			selectrecvImpl(sel, rc.ch, uintptr(i), rc.val, r, 0)
		}
	}

	pc, _ := selectgoImpl(sel)
	chosen = int(pc)
	recvOK = *r
	return
}

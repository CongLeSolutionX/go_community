// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_channels "runtime/internal/channels"
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	_maps "runtime/internal/maps"
	_sched "runtime/internal/sched"
	"unsafe"
)

// TODO(khr): make hchan.buf an unsafe.Pointer, not a *uint8

//go:linkname reflect_makechan reflect.makechan
func reflect_makechan(t *_channels.Chantype, size int64) *_channels.Hchan {
	return makechan(t, size)
}

func makechan(t *_channels.Chantype, size int64) *_channels.Hchan {
	elem := t.Elem

	// compiler checks this but be safe.
	if elem.Size >= 1<<16 {
		_lock.Throw("makechan: invalid channel element type")
	}
	if _channels.HchanSize%_channels.MaxAlign != 0 || elem.Align > _channels.MaxAlign {
		_lock.Throw("makechan: bad alignment")
	}
	if size < 0 || int64(uintptr(size)) != size || (elem.Size > 0 && uintptr(size) > (_core.MaxMem-_channels.HchanSize)/uintptr(elem.Size)) {
		panic("makechan: size out of range")
	}

	var c *_channels.Hchan
	if elem.Kind&_channels.KindNoPointers != 0 || size == 0 {
		// Allocate memory in one call.
		// Hchan does not contain pointers interesting for GC in this case:
		// buf points into the same allocation, elemtype is persistent.
		// SudoG's are referenced from their owning thread so they can't be collected.
		// TODO(dvyukov,rlh): Rethink when collector can move allocated objects.
		c = (*_channels.Hchan)(_maps.Mallocgc(_channels.HchanSize+uintptr(size)*uintptr(elem.Size), nil, _sched.XFlagNoScan))
		if size > 0 && elem.Size != 0 {
			c.Buf = (*uint8)(_core.Add(unsafe.Pointer(c), _channels.HchanSize))
		} else {
			// race detector uses this location for synchronization
			// Also prevents us from pointing beyond the allocation (see issue 9401).
			c.Buf = (*uint8)(unsafe.Pointer(c))
		}
	} else {
		c = new(_channels.Hchan)
		c.Buf = (*uint8)(_maps.Newarray(elem, uintptr(size)))
	}
	c.Elemsize = uint16(elem.Size)
	c.Elemtype = elem
	c.Dataqsiz = uint(size)

	if _channels.DebugChan {
		print("makechan: chan=", c, "; elemsize=", elem.Size, "; elemalg=", elem.Alg, "; dataqsiz=", size, "\n")
	}
	return c
}

func closechan(c *_channels.Hchan) {
	if c == nil {
		panic("close of nil channel")
	}

	_lock.Lock(&c.Lock)
	if c.Closed != 0 {
		_lock.Unlock(&c.Lock)
		panic("close of closed channel")
	}

	if _sched.Raceenabled {
		callerpc := _lock.Getcallerpc(unsafe.Pointer(&c))
		_maps.Racewritepc(unsafe.Pointer(c), callerpc, _lock.FuncPC(closechan))
		_channels.Racerelease(unsafe.Pointer(c))
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
			sg.Releasetime = _sched.Cputicks()
		}
		_sched.Goready(gp)
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
			sg.Releasetime = _sched.Cputicks()
		}
		_sched.Goready(gp)
	}
	_lock.Unlock(&c.Lock)
}

//go:linkname reflect_chansend reflect.chansend
func reflect_chansend(t *_channels.Chantype, c *_channels.Hchan, elem unsafe.Pointer, nb bool) (selected bool) {
	return _channels.Chansend(t, c, elem, !nb, _lock.Getcallerpc(unsafe.Pointer(&t)))
}

//go:linkname reflect_chanlen reflect.chanlen
func reflect_chanlen(c *_channels.Hchan) int {
	if c == nil {
		return 0
	}
	return int(c.Qcount)
}

//go:linkname reflect_chancap reflect.chancap
func reflect_chancap(c *_channels.Hchan) int {
	if c == nil {
		return 0
	}
	return int(c.Dataqsiz)
}

//go:linkname reflect_chanclose reflect.chanclose
func reflect_chanclose(c *_channels.Hchan) {
	closechan(c)
}

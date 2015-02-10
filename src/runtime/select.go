// Copyright 2009 The Go Authors. All rights reserved.
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

//go:nosplit
func selectsend(sel *_channels.Select, c *_channels.Hchan, elem unsafe.Pointer) (selected bool) {
	// nil cases do not compete
	if c != nil {
		selectsendImpl(sel, c, _lock.Getcallerpc(unsafe.Pointer(&sel)), elem, uintptr(unsafe.Pointer(&selected))-uintptr(unsafe.Pointer(&sel)))
	}
	return
}

// cut in half to give stack a chance to split
func selectsendImpl(sel *_channels.Select, c *_channels.Hchan, pc uintptr, elem unsafe.Pointer, so uintptr) {
	i := sel.Ncase
	if i >= sel.Tcase {
		_lock.Throw("selectsend: too many cases")
	}
	sel.Ncase = i + 1
	cas := (*_channels.Scase)(_core.Add(unsafe.Pointer(&sel.Scase), uintptr(i)*unsafe.Sizeof(sel.Scase[0])))

	cas.Pc = pc
	cas.Chan = c
	cas.So = uint16(so)
	cas.Kind = _channels.CaseSend
	cas.Elem = elem

	if _channels.DebugSelect {
		print("selectsend s=", sel, " pc=", _core.Hex(cas.Pc), " chan=", cas.Chan, " so=", cas.So, "\n")
	}
}

//go:nosplit
func selectdefault(sel *_channels.Select) (selected bool) {
	selectdefaultImpl(sel, _lock.Getcallerpc(unsafe.Pointer(&sel)), uintptr(unsafe.Pointer(&selected))-uintptr(unsafe.Pointer(&sel)))
	return
}

func selectdefaultImpl(sel *_channels.Select, callerpc uintptr, so uintptr) {
	i := sel.Ncase
	if i >= sel.Tcase {
		_lock.Throw("selectdefault: too many cases")
	}
	sel.Ncase = i + 1
	cas := (*_channels.Scase)(_core.Add(unsafe.Pointer(&sel.Scase), uintptr(i)*unsafe.Sizeof(sel.Scase[0])))
	cas.Pc = callerpc
	cas.Chan = nil
	cas.So = uint16(so)
	cas.Kind = _channels.CaseDefault

	if _channels.DebugSelect {
		print("selectdefault s=", sel, " pc=", _core.Hex(cas.Pc), " so=", cas.So, "\n")
	}
}

// A runtimeSelect is a single case passed to rselect.
// This must match ../reflect/value.go:/runtimeSelect
type runtimeSelect struct {
	dir _channels.SelectDir
	typ unsafe.Pointer   // channel type (not used here)
	ch  *_channels.Hchan // channel
	val unsafe.Pointer   // ptr to data (SendDir) or ptr to receive buffer (RecvDir)
}

const (
	_             _channels.SelectDir = iota
	selectSend                        // case Chan <- Send
	selectRecv                        // case <-Chan:
	selectDefault                     // default
)

//go:linkname reflect_rselect reflect.rselect
func reflect_rselect(cases []runtimeSelect) (chosen int, recvOK bool) {
	// flagNoScan is safe here, because all objects are also referenced from cases.
	size := _channels.Selectsize(uintptr(len(cases)))
	sel := (*_channels.Select)(_maps.Mallocgc(size, nil, _sched.XFlagNoScan))
	_channels.Newselect(sel, int64(size), int32(len(cases)))
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
			_channels.SelectrecvImpl(sel, rc.ch, uintptr(i), rc.val, r, 0)
		}
	}

	pc, _ := _channels.SelectgoImpl(sel)
	chosen = int(pc)
	recvOK = *r
	return
}

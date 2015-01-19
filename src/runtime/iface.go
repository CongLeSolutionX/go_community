// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_core "runtime/internal/core"
	_hash "runtime/internal/hash"
	_ifacestuff "runtime/internal/ifacestuff"
	_sched "runtime/internal/sched"
	"unsafe"
)

func typ2Itab(t *_core.Type, inter *_core.Interfacetype, cache **_core.Itab) *_core.Itab {
	tab := _ifacestuff.Getitab(inter, t, false)
	_sched.Atomicstorep(unsafe.Pointer(cache), unsafe.Pointer(tab))
	return tab
}

//go:linkname reflect_ifaceE2I reflect.ifaceE2I
func reflect_ifaceE2I(inter *_core.Interfacetype, e interface{}, dst *_ifacestuff.FInterface) {
	*dst = _ifacestuff.AssertE2I(inter, e)
}

func ifacethash(i _ifacestuff.FInterface) uint32 {
	ip := (*_hash.Iface)(unsafe.Pointer(&i))
	tab := ip.Tab
	if tab == nil {
		return 0
	}
	return tab.Type.Hash
}

func efacethash(e interface{}) uint32 {
	ep := (*_core.Eface)(unsafe.Pointer(&e))
	t := ep.Type
	if t == nil {
		return 0
	}
	return t.Hash
}

func iterate_itabs(fn func(*_core.Itab)) {
	for _, h := range &_hash.Hash {
		for ; h != nil; h = h.Link {
			fn(h)
		}
	}
}

func ifaceE2I2(inter *_core.Interfacetype, e interface{}, r *_ifacestuff.FInterface) (ok bool) {
	*r, ok = _ifacestuff.AssertE2I2(inter, e)
	return
}

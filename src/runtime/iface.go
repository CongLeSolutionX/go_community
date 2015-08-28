// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_base "runtime/internal/base"
	_iface "runtime/internal/iface"
	"unsafe"
)

func typ2Itab(t *_base.Type, inter *_iface.Interfacetype, cache **_iface.Itab) *_iface.Itab {
	tab := _iface.Getitab(inter, t, false)
	_base.Atomicstorep(unsafe.Pointer(cache), unsafe.Pointer(tab))
	return tab
}

func panicdottype(have, want, iface *_base.Type) {
	haveString := ""
	if have != nil {
		haveString = *have.String
	}
	panic(&_iface.TypeAssertionError{*iface.String, haveString, *want.String, ""})
}

//go:linkname reflect_ifaceE2I reflect.ifaceE2I
func reflect_ifaceE2I(inter *_iface.Interfacetype, e interface{}, dst *_iface.FInterface) {
	_iface.AssertE2I(inter, e, dst)
}

func ifacethash(i _iface.FInterface) uint32 {
	ip := (*_iface.Iface)(unsafe.Pointer(&i))
	tab := ip.Tab
	if tab == nil {
		return 0
	}
	return tab.Type.Hash
}

func efacethash(e interface{}) uint32 {
	ep := (*_iface.Eface)(unsafe.Pointer(&e))
	t := ep.Type
	if t == nil {
		return 0
	}
	return t.Hash
}

func iterate_itabs(fn func(*_iface.Itab)) {
	for _, h := range &_iface.Hash {
		for ; h != nil; h = h.Link {
			fn(h)
		}
	}
}

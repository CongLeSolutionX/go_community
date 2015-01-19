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

// Reflect stubs.  Called from ../reflect/asm_*.s

//go:linkname reflect_makemap reflect.makemap
func reflect_makemap(t *_maps.Maptype) *_maps.Hmap {
	return _maps.Makemap(t, 0)
}

//go:linkname reflect_mapaccess reflect.mapaccess
func reflect_mapaccess(t *_maps.Maptype, h *_maps.Hmap, key unsafe.Pointer) unsafe.Pointer {
	val, ok := _maps.Mapaccess2(t, h, key)
	if !ok {
		// reflect wants nil for a missing element
		val = nil
	}
	return val
}

//go:linkname reflect_mapassign reflect.mapassign
func reflect_mapassign(t *_maps.Maptype, h *_maps.Hmap, key unsafe.Pointer, val unsafe.Pointer) {
	_maps.Mapassign1(t, h, key, val)
}

//go:linkname reflect_mapdelete reflect.mapdelete
func reflect_mapdelete(t *_maps.Maptype, h *_maps.Hmap, key unsafe.Pointer) {
	_maps.Mapdelete(t, h, key)
}

//go:linkname reflect_mapiterinit reflect.mapiterinit
func reflect_mapiterinit(t *_maps.Maptype, h *_maps.Hmap) *_maps.Hiter {
	it := new(_maps.Hiter)
	_maps.Mapiterinit(t, h, it)
	return it
}

//go:linkname reflect_mapiterkey reflect.mapiterkey
func reflect_mapiterkey(it *_maps.Hiter) unsafe.Pointer {
	return it.Key
}

//go:linkname reflect_maplen reflect.maplen
func reflect_maplen(h *_maps.Hmap) int {
	if h == nil {
		return 0
	}
	if _sched.Raceenabled {
		callerpc := _lock.Getcallerpc(unsafe.Pointer(&h))
		_channels.Racereadpc(unsafe.Pointer(h), callerpc, _lock.FuncPC(reflect_maplen))
	}
	return h.Count
}

//go:linkname reflect_ismapkey reflect.ismapkey
func reflect_ismapkey(t *_core.Type) bool {
	return _maps.Ismapkey(t)
}

// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ifacestuff

import (
	_channels "runtime/internal/channels"
	_core "runtime/internal/core"
	_hash "runtime/internal/hash"
	_lock "runtime/internal/lock"
	_maps "runtime/internal/maps"
	_prof "runtime/internal/prof"
	_sched "runtime/internal/sched"
	"unsafe"
)

var (
	ifaceLock _core.Mutex // lock for accessing hash
)

// fInterface is our standard non-empty interface.  We use it instead
// of interface{f()} in function prototypes because gofmt insists on
// putting lots of newlines in the otherwise concise interface{f()}.
type FInterface interface {
	f()
}

func Getitab(inter *_core.Interfacetype, typ *_core.Type, canfail bool) *_core.Itab {
	if len(inter.Mhdr) == 0 {
		_lock.Throw("internal error - misuse of itab")
	}

	// easy case
	x := typ.X
	if x == nil {
		if canfail {
			return nil
		}
		panic(&TypeAssertionError{"", *typ.String, *inter.Typ.String, *inter.Mhdr[0].Name})
	}

	// compiler has provided some good hash codes for us.
	h := inter.Typ.Hash
	h += 17 * typ.Hash
	// TODO(rsc): h += 23 * x.mhash ?
	h %= _hash.HashSize

	// look twice - once without lock, once with.
	// common case will be no lock contention.
	var m *_core.Itab
	var locked int
	for locked = 0; locked < 2; locked++ {
		if locked != 0 {
			_lock.Lock(&ifaceLock)
		}
		for m = (*_core.Itab)(_prof.Atomicloadp(unsafe.Pointer(&_hash.Hash[h]))); m != nil; m = m.Link {
			if m.Inter == inter && m.Type == typ {
				if m.Bad != 0 {
					m = nil
					if !canfail {
						// this can only happen if the conversion
						// was already done once using the , ok form
						// and we have a cached negative result.
						// the cached result doesn't record which
						// interface function was missing, so jump
						// down to the interface check, which will
						// do more work but give a better error.
						goto search
					}
				}
				if locked != 0 {
					_lock.Unlock(&ifaceLock)
				}
				return m
			}
		}
	}

	m = (*_core.Itab)(_lock.Persistentalloc(unsafe.Sizeof(_core.Itab{})+uintptr(len(inter.Mhdr)-1)*_core.PtrSize, 0, &_lock.Memstats.Other_sys))
	m.Inter = inter
	m.Type = typ

search:
	// both inter and typ have method sorted by name,
	// and interface names are unique,
	// so can iterate over both in lock step;
	// the loop is O(ni+nt) not O(ni*nt).
	ni := len(inter.Mhdr)
	nt := len(x.Mhdr)
	j := 0
	for k := 0; k < ni; k++ {
		i := &inter.Mhdr[k]
		iname := i.Name
		ipkgpath := i.Pkgpath
		itype := i.Type
		for ; j < nt; j++ {
			t := &x.Mhdr[j]
			if t.Mtyp == itype && t.Name == iname && t.Pkgpath == ipkgpath {
				if m != nil {
					*(*unsafe.Pointer)(_core.Add(unsafe.Pointer(&m.Fun[0]), uintptr(k)*_core.PtrSize)) = t.Ifn
				}
				goto nextimethod
			}
		}
		// didn't find method
		if !canfail {
			if locked != 0 {
				_lock.Unlock(&ifaceLock)
			}
			panic(&TypeAssertionError{"", *typ.String, *inter.Typ.String, *iname})
		}
		m.Bad = 1
		break
	nextimethod:
	}
	if locked == 0 {
		_lock.Throw("invalid itab locking")
	}
	m.Link = _hash.Hash[h]
	_sched.Atomicstorep(unsafe.Pointer(&_hash.Hash[h]), unsafe.Pointer(m))
	_lock.Unlock(&ifaceLock)
	if m.Bad != 0 {
		return nil
	}
	return m
}

func convT2E(t *_core.Type, elem unsafe.Pointer) (e interface{}) {
	ep := (*_core.Eface)(unsafe.Pointer(&e))
	if _hash.IsDirectIface(t) {
		ep.Type = t
		_channels.Typedmemmove(t, unsafe.Pointer(&ep.Data), elem)
	} else {
		x := _maps.Newobject(t)
		// TODO: We allocate a zeroed object only to overwrite it with
		// actual data.  Figure out how to avoid zeroing.  Also below in convT2I.
		_channels.Typedmemmove(t, x, elem)
		ep.Type = t
		ep.Data = x
	}
	return
}

func convT2I(t *_core.Type, inter *_core.Interfacetype, cache **_core.Itab, elem unsafe.Pointer) (i FInterface) {
	tab := (*_core.Itab)(_prof.Atomicloadp(unsafe.Pointer(cache)))
	if tab == nil {
		tab = Getitab(inter, t, false)
		_sched.Atomicstorep(unsafe.Pointer(cache), unsafe.Pointer(tab))
	}
	pi := (*_hash.Iface)(unsafe.Pointer(&i))
	if _hash.IsDirectIface(t) {
		pi.Tab = tab
		_channels.Typedmemmove(t, unsafe.Pointer(&pi.Data), elem)
	} else {
		x := _maps.Newobject(t)
		_channels.Typedmemmove(t, x, elem)
		pi.Tab = tab
		pi.Data = x
	}
	return
}

func assertI2T(t *_core.Type, i FInterface, r unsafe.Pointer) {
	ip := (*_hash.Iface)(unsafe.Pointer(&i))
	tab := ip.Tab
	if tab == nil {
		panic(&TypeAssertionError{"", "", *t.String, ""})
	}
	if tab.Type != t {
		panic(&TypeAssertionError{*tab.Inter.Typ.String, *tab.Type.String, *t.String, ""})
	}
	if r != nil {
		if _hash.IsDirectIface(t) {
			_channels.Writebarrierptr((*uintptr)(r), uintptr(ip.Data))
		} else {
			_channels.Typedmemmove(t, r, ip.Data)
		}
	}
}

func assertI2T2(t *_core.Type, i FInterface, r unsafe.Pointer) bool {
	ip := (*_hash.Iface)(unsafe.Pointer(&i))
	tab := ip.Tab
	if tab == nil || tab.Type != t {
		if r != nil {
			_core.Memclr(r, uintptr(t.Size))
		}
		return false
	}
	if r != nil {
		if _hash.IsDirectIface(t) {
			_channels.Writebarrierptr((*uintptr)(r), uintptr(ip.Data))
		} else {
			_channels.Typedmemmove(t, r, ip.Data)
		}
	}
	return true
}

func assertE2T(t *_core.Type, e interface{}, r unsafe.Pointer) {
	ep := (*_core.Eface)(unsafe.Pointer(&e))
	if ep.Type == nil {
		panic(&TypeAssertionError{"", "", *t.String, ""})
	}
	if ep.Type != t {
		panic(&TypeAssertionError{"", *ep.Type.String, *t.String, ""})
	}
	if r != nil {
		if _hash.IsDirectIface(t) {
			_channels.Writebarrierptr((*uintptr)(r), uintptr(ep.Data))
		} else {
			_channels.Typedmemmove(t, r, ep.Data)
		}
	}
}

func assertE2T2(t *_core.Type, e interface{}, r unsafe.Pointer) bool {
	ep := (*_core.Eface)(unsafe.Pointer(&e))
	if ep.Type != t {
		if r != nil {
			_core.Memclr(r, uintptr(t.Size))
		}
		return false
	}
	if r != nil {
		if _hash.IsDirectIface(t) {
			_channels.Writebarrierptr((*uintptr)(r), uintptr(ep.Data))
		} else {
			_channels.Typedmemmove(t, r, ep.Data)
		}
	}
	return true
}

func convI2E(i FInterface) (r interface{}) {
	ip := (*_hash.Iface)(unsafe.Pointer(&i))
	tab := ip.Tab
	if tab == nil {
		return
	}
	rp := (*_core.Eface)(unsafe.Pointer(&r))
	rp.Type = tab.Type
	rp.Data = ip.Data
	return
}

func assertI2E(inter *_core.Interfacetype, i FInterface, r *interface{}) {
	ip := (*_hash.Iface)(unsafe.Pointer(&i))
	tab := ip.Tab
	if tab == nil {
		// explicit conversions require non-nil interface value.
		panic(&TypeAssertionError{"", "", *inter.Typ.String, ""})
	}
	rp := (*_core.Eface)(unsafe.Pointer(r))
	rp.Type = tab.Type
	rp.Data = ip.Data
	return
}

func assertI2E2(inter *_core.Interfacetype, i FInterface, r *interface{}) bool {
	ip := (*_hash.Iface)(unsafe.Pointer(&i))
	tab := ip.Tab
	if tab == nil {
		return false
	}
	if r != nil {
		rp := (*_core.Eface)(unsafe.Pointer(r))
		rp.Type = tab.Type
		rp.Data = ip.Data
	}
	return true
}

func convI2I(inter *_core.Interfacetype, i FInterface) (r FInterface) {
	ip := (*_hash.Iface)(unsafe.Pointer(&i))
	tab := ip.Tab
	if tab == nil {
		return
	}
	rp := (*_hash.Iface)(unsafe.Pointer(&r))
	if tab.Inter == inter {
		rp.Tab = tab
		rp.Data = ip.Data
		return
	}
	rp.Tab = Getitab(inter, tab.Type, false)
	rp.Data = ip.Data
	return
}

func assertI2I(inter *_core.Interfacetype, i FInterface, r *FInterface) {
	ip := (*_hash.Iface)(unsafe.Pointer(&i))
	tab := ip.Tab
	if tab == nil {
		// explicit conversions require non-nil interface value.
		panic(&TypeAssertionError{"", "", *inter.Typ.String, ""})
	}
	rp := (*_hash.Iface)(unsafe.Pointer(r))
	if tab.Inter == inter {
		rp.Tab = tab
		rp.Data = ip.Data
		return
	}
	rp.Tab = Getitab(inter, tab.Type, false)
	rp.Data = ip.Data
}

func assertI2I2(inter *_core.Interfacetype, i FInterface, r *FInterface) bool {
	ip := (*_hash.Iface)(unsafe.Pointer(&i))
	tab := ip.Tab
	if tab == nil {
		if r != nil {
			*r = nil
		}
		return false
	}
	if tab.Inter != inter {
		tab = Getitab(inter, tab.Type, true)
		if tab == nil {
			if r != nil {
				*r = nil
			}
			return false
		}
	}
	if r != nil {
		rp := (*_hash.Iface)(unsafe.Pointer(r))
		rp.Tab = tab
		rp.Data = ip.Data
	}
	return true
}

func AssertE2I(inter *_core.Interfacetype, e interface{}, r *FInterface) {
	ep := (*_core.Eface)(unsafe.Pointer(&e))
	t := ep.Type
	if t == nil {
		// explicit conversions require non-nil interface value.
		panic(&TypeAssertionError{"", "", *inter.Typ.String, ""})
	}
	rp := (*_hash.Iface)(unsafe.Pointer(r))
	rp.Tab = Getitab(inter, t, false)
	rp.Data = ep.Data
}

func AssertE2I2(inter *_core.Interfacetype, e interface{}, r *FInterface) bool {
	ep := (*_core.Eface)(unsafe.Pointer(&e))
	t := ep.Type
	if t == nil {
		if r != nil {
			*r = nil
		}
		return false
	}
	tab := Getitab(inter, t, true)
	if tab == nil {
		if r != nil {
			*r = nil
		}
		return false
	}
	if r != nil {
		rp := (*_hash.Iface)(unsafe.Pointer(r))
		rp.Tab = tab
		rp.Data = ep.Data
	}
	return true
}

func assertE2E(inter *_core.Interfacetype, e interface{}, r *interface{}) {
	ep := (*_core.Eface)(unsafe.Pointer(&e))
	if ep.Type == nil {
		// explicit conversions require non-nil interface value.
		panic(&TypeAssertionError{"", "", *inter.Typ.String, ""})
	}
	*r = e
}

func assertE2E2(inter *_core.Interfacetype, e interface{}, r *interface{}) bool {
	ep := (*_core.Eface)(unsafe.Pointer(&e))
	if ep.Type == nil {
		if r != nil {
			*r = nil
		}
		return false
	}
	if r != nil {
		*r = e
	}
	return true
}

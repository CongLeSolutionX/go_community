// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ifacestuff

import (
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
		i := (*_core.Imethod)(_core.Add(unsafe.Pointer(inter), unsafe.Sizeof(_core.Interfacetype{})))
		panic(&TypeAssertionError{"", *typ.String, *inter.Typ.String, *i.Name})
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

	m = (*_core.Itab)(_lock.Persistentalloc(unsafe.Sizeof(_core.Itab{})+uintptr(len(inter.Mhdr))*_core.PtrSize, 0, &_lock.Memstats.Other_sys))
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
		i := (*_core.Imethod)(_core.Add(unsafe.Pointer(inter), unsafe.Sizeof(_core.Interfacetype{})+uintptr(k)*unsafe.Sizeof(_core.Imethod{})))
		iname := i.Name
		ipkgpath := i.Pkgpath
		itype := i.Type
		for ; j < nt; j++ {
			t := (*_core.Method)(_core.Add(unsafe.Pointer(x), unsafe.Sizeof(_core.Uncommontype{})+uintptr(j)*unsafe.Sizeof(_core.Method{})))
			if t.Mtyp == itype && t.Name == iname && t.Pkgpath == ipkgpath {
				if m != nil {
					*(*unsafe.Pointer)(_core.Add(unsafe.Pointer(m), unsafe.Sizeof(_core.Itab{})+uintptr(k)*_core.PtrSize)) = t.Ifn
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
	size := uintptr(t.Size)
	ep := (*_core.Eface)(unsafe.Pointer(&e))
	if _hash.IsDirectIface(t) {
		ep.Type = t
		_sched.Memmove(unsafe.Pointer(&ep.Data), elem, size)
	} else {
		x := _maps.Newobject(t)
		// TODO: We allocate a zeroed object only to overwrite it with
		// actual data.  Figure out how to avoid zeroing.  Also below in convT2I.
		_sched.Memmove(x, elem, size)
		ep.Type = t
		ep.Data = x
	}
	return
}

func ConvT2I(t *_core.Type, inter *_core.Interfacetype, cache **_core.Itab, elem unsafe.Pointer) (i FInterface) {
	return convT2I(t, inter, cache, elem)
}

func convT2I(t *_core.Type, inter *_core.Interfacetype, cache **_core.Itab, elem unsafe.Pointer) (i FInterface) {
	tab := (*_core.Itab)(_prof.Atomicloadp(unsafe.Pointer(cache)))
	if tab == nil {
		tab = Getitab(inter, t, false)
		_sched.Atomicstorep(unsafe.Pointer(cache), unsafe.Pointer(tab))
	}
	size := uintptr(t.Size)
	pi := (*_hash.Iface)(unsafe.Pointer(&i))
	if _hash.IsDirectIface(t) {
		pi.Tab = tab
		_sched.Memmove(unsafe.Pointer(&pi.Data), elem, size)
	} else {
		x := _maps.Newobject(t)
		_sched.Memmove(x, elem, size)
		pi.Tab = tab
		pi.Data = x
	}
	return
}

// TODO: give these routines a pointer to the result area instead of writing
// extra data in the outargs section.  Then we can get rid of go:nosplit.
//go:nosplit
func assertI2T(t *_core.Type, i FInterface) (r struct{}) {
	ip := (*_hash.Iface)(unsafe.Pointer(&i))
	tab := ip.Tab
	if tab == nil {
		panic(&TypeAssertionError{"", "", *t.String, ""})
	}
	if tab.Type != t {
		panic(&TypeAssertionError{*tab.Inter.Typ.String, *tab.Type.String, *t.String, ""})
	}
	size := uintptr(t.Size)
	if _hash.IsDirectIface(t) {
		_sched.Memmove(unsafe.Pointer(&r), unsafe.Pointer(&ip.Data), size)
	} else {
		_sched.Memmove(unsafe.Pointer(&r), ip.Data, size)
	}
	return
}

//go:nosplit
func assertI2T2(t *_core.Type, i FInterface) (r byte) {
	ip := (*_hash.Iface)(unsafe.Pointer(&i))
	size := uintptr(t.Size)
	ok := (*bool)(_core.Add(unsafe.Pointer(&r), size))
	tab := ip.Tab
	if tab == nil || tab.Type != t {
		*ok = false
		_core.Memclr(unsafe.Pointer(&r), size)
		return
	}
	*ok = true
	if _hash.IsDirectIface(t) {
		_sched.Memmove(unsafe.Pointer(&r), unsafe.Pointer(&ip.Data), size)
	} else {
		_sched.Memmove(unsafe.Pointer(&r), ip.Data, size)
	}
	return
}

func assertI2TOK(t *_core.Type, i FInterface) bool {
	ip := (*_hash.Iface)(unsafe.Pointer(&i))
	tab := ip.Tab
	return tab != nil && tab.Type == t
}

//go:nosplit
func assertE2T(t *_core.Type, e interface{}) (r struct{}) {
	ep := (*_core.Eface)(unsafe.Pointer(&e))
	if ep.Type == nil {
		panic(&TypeAssertionError{"", "", *t.String, ""})
	}
	if ep.Type != t {
		panic(&TypeAssertionError{"", *ep.Type.String, *t.String, ""})
	}
	size := uintptr(t.Size)
	if _hash.IsDirectIface(t) {
		_sched.Memmove(unsafe.Pointer(&r), unsafe.Pointer(&ep.Data), size)
	} else {
		_sched.Memmove(unsafe.Pointer(&r), ep.Data, size)
	}
	return
}

//go:nosplit
func assertE2T2(t *_core.Type, e interface{}) (r byte) {
	ep := (*_core.Eface)(unsafe.Pointer(&e))
	size := uintptr(t.Size)
	ok := (*bool)(_core.Add(unsafe.Pointer(&r), size))
	if ep.Type != t {
		*ok = false
		_core.Memclr(unsafe.Pointer(&r), size)
		return
	}
	*ok = true
	if _hash.IsDirectIface(t) {
		_sched.Memmove(unsafe.Pointer(&r), unsafe.Pointer(&ep.Data), size)
	} else {
		_sched.Memmove(unsafe.Pointer(&r), ep.Data, size)
	}
	return
}

func assertE2TOK(t *_core.Type, e interface{}) bool {
	ep := (*_core.Eface)(unsafe.Pointer(&e))
	return t == ep.Type
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

func assertI2E(inter *_core.Interfacetype, i FInterface) (r interface{}) {
	ip := (*_hash.Iface)(unsafe.Pointer(&i))
	tab := ip.Tab
	if tab == nil {
		// explicit conversions require non-nil interface value.
		panic(&TypeAssertionError{"", "", *inter.Typ.String, ""})
	}
	rp := (*_core.Eface)(unsafe.Pointer(&r))
	rp.Type = tab.Type
	rp.Data = ip.Data
	return
}

func assertI2E2(inter *_core.Interfacetype, i FInterface) (r interface{}, ok bool) {
	ip := (*_hash.Iface)(unsafe.Pointer(&i))
	tab := ip.Tab
	if tab == nil {
		return
	}
	rp := (*_core.Eface)(unsafe.Pointer(&r))
	rp.Type = tab.Type
	rp.Data = ip.Data
	ok = true
	return
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

func assertI2I(inter *_core.Interfacetype, i FInterface) (r FInterface) {
	ip := (*_hash.Iface)(unsafe.Pointer(&i))
	tab := ip.Tab
	if tab == nil {
		// explicit conversions require non-nil interface value.
		panic(&TypeAssertionError{"", "", *inter.Typ.String, ""})
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

func assertI2I2(inter *_core.Interfacetype, i FInterface) (r FInterface, ok bool) {
	ip := (*_hash.Iface)(unsafe.Pointer(&i))
	tab := ip.Tab
	if tab == nil {
		return
	}
	rp := (*_hash.Iface)(unsafe.Pointer(&r))
	if tab.Inter == inter {
		rp.Tab = tab
		rp.Data = ip.Data
		ok = true
		return
	}
	tab = Getitab(inter, tab.Type, true)
	if tab == nil {
		rp.Data = nil
		rp.Tab = nil
		ok = false
		return
	}
	rp.Tab = tab
	rp.Data = ip.Data
	ok = true
	return
}

func assertE2I(inter *_core.Interfacetype, e interface{}) (r FInterface) {
	return AssertE2I(inter, e)
}

func AssertE2I(inter *_core.Interfacetype, e interface{}) (r FInterface) {
	ep := (*_core.Eface)(unsafe.Pointer(&e))
	t := ep.Type
	if t == nil {
		// explicit conversions require non-nil interface value.
		panic(&TypeAssertionError{"", "", *inter.Typ.String, ""})
	}
	rp := (*_hash.Iface)(unsafe.Pointer(&r))
	rp.Tab = Getitab(inter, t, false)
	rp.Data = ep.Data
	return
}

func assertE2I2(inter *_core.Interfacetype, e interface{}) (r FInterface, ok bool) {
	return AssertE2I2(inter, e)
}

func AssertE2I2(inter *_core.Interfacetype, e interface{}) (r FInterface, ok bool) {
	ep := (*_core.Eface)(unsafe.Pointer(&e))
	t := ep.Type
	if t == nil {
		return
	}
	tab := Getitab(inter, t, true)
	if tab == nil {
		return
	}
	rp := (*_hash.Iface)(unsafe.Pointer(&r))
	rp.Tab = tab
	rp.Data = ep.Data
	ok = true
	return
}

func assertE2E(inter *_core.Interfacetype, e interface{}) interface{} {
	ep := (*_core.Eface)(unsafe.Pointer(&e))
	if ep.Type == nil {
		// explicit conversions require non-nil interface value.
		panic(&TypeAssertionError{"", "", *inter.Typ.String, ""})
	}
	return e
}

func assertE2E2(inter *_core.Interfacetype, e interface{}) (interface{}, bool) {
	ep := (*_core.Eface)(unsafe.Pointer(&e))
	if ep.Type == nil {
		return nil, false
	}
	return e, true
}

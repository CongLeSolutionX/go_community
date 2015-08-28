// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package iface

import (
	_base "runtime/internal/base"
	"unsafe"
)

const (
	hashSize = 1009
)

var (
	ifaceLock _base.Mutex // lock for accessing hash
	Hash      [hashSize]*Itab
)

// fInterface is our standard non-empty interface.  We use it instead
// of interface{f()} in function prototypes because gofmt insists on
// putting lots of newlines in the otherwise concise interface{f()}.
type FInterface interface {
	f()
}

func Getitab(inter *Interfacetype, typ *_base.Type, canfail bool) *Itab {
	if len(inter.Mhdr) == 0 {
		_base.Throw("internal error - misuse of itab")
	}

	// easy case
	x := typ.X
	if x == nil {
		if canfail {
			return nil
		}
		panic(&TypeAssertionError{"", *typ.String, *inter.typ.String, *inter.Mhdr[0].name})
	}

	// compiler has provided some good hash codes for us.
	h := inter.typ.Hash
	h += 17 * typ.Hash
	// TODO(rsc): h += 23 * x.mhash ?
	h %= hashSize

	// look twice - once without lock, once with.
	// common case will be no lock contention.
	var m *Itab
	var locked int
	for locked = 0; locked < 2; locked++ {
		if locked != 0 {
			_base.Lock(&ifaceLock)
		}
		for m = (*Itab)(Atomicloadp(unsafe.Pointer(&Hash[h]))); m != nil; m = m.Link {
			if m.inter == inter && m.Type == typ {
				if m.bad != 0 {
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
					_base.Unlock(&ifaceLock)
				}
				return m
			}
		}
	}

	m = (*Itab)(_base.Persistentalloc(unsafe.Sizeof(Itab{})+uintptr(len(inter.Mhdr)-1)*_base.PtrSize, 0, &_base.Memstats.Other_sys))
	m.inter = inter
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
		iname := i.name
		ipkgpath := i.pkgpath
		itype := i._type
		for ; j < nt; j++ {
			t := &x.Mhdr[j]
			if t.Mtyp == itype && (t.Name == iname || *t.Name == *iname) && t.Pkgpath == ipkgpath {
				if m != nil {
					*(*unsafe.Pointer)(_base.Add(unsafe.Pointer(&m.fun[0]), uintptr(k)*_base.PtrSize)) = t.Ifn
				}
				goto nextimethod
			}
		}
		// didn't find method
		if !canfail {
			if locked != 0 {
				_base.Unlock(&ifaceLock)
			}
			panic(&TypeAssertionError{"", *typ.String, *inter.typ.String, *iname})
		}
		m.bad = 1
		break
	nextimethod:
	}
	if locked == 0 {
		_base.Throw("invalid itab locking")
	}
	m.Link = Hash[h]
	_base.Atomicstorep(unsafe.Pointer(&Hash[h]), unsafe.Pointer(m))
	_base.Unlock(&ifaceLock)
	if m.bad != 0 {
		return nil
	}
	return m
}

func convT2E(t *_base.Type, elem unsafe.Pointer, x unsafe.Pointer) (e interface{}) {
	ep := (*Eface)(unsafe.Pointer(&e))
	if IsDirectIface(t) {
		ep.Type = t
		Typedmemmove(t, unsafe.Pointer(&ep.Data), elem)
	} else {
		if x == nil {
			x = Newobject(t)
		}
		// TODO: We allocate a zeroed object only to overwrite it with
		// actual data.  Figure out how to avoid zeroing.  Also below in convT2I.
		Typedmemmove(t, x, elem)
		ep.Type = t
		ep.Data = x
	}
	return
}

func convT2I(t *_base.Type, inter *Interfacetype, cache **Itab, elem unsafe.Pointer, x unsafe.Pointer) (i FInterface) {
	tab := (*Itab)(Atomicloadp(unsafe.Pointer(cache)))
	if tab == nil {
		tab = Getitab(inter, t, false)
		_base.Atomicstorep(unsafe.Pointer(cache), unsafe.Pointer(tab))
	}
	pi := (*Iface)(unsafe.Pointer(&i))
	if IsDirectIface(t) {
		pi.Tab = tab
		Typedmemmove(t, unsafe.Pointer(&pi.Data), elem)
	} else {
		if x == nil {
			x = Newobject(t)
		}
		Typedmemmove(t, x, elem)
		pi.Tab = tab
		pi.Data = x
	}
	return
}

func assertI2T(t *_base.Type, i FInterface, r unsafe.Pointer) {
	ip := (*Iface)(unsafe.Pointer(&i))
	tab := ip.Tab
	if tab == nil {
		panic(&TypeAssertionError{"", "", *t.String, ""})
	}
	if tab.Type != t {
		panic(&TypeAssertionError{*tab.inter.typ.String, *tab.Type.String, *t.String, ""})
	}
	if r != nil {
		if IsDirectIface(t) {
			Writebarrierptr((*uintptr)(r), uintptr(ip.Data))
		} else {
			Typedmemmove(t, r, ip.Data)
		}
	}
}

func assertI2T2(t *_base.Type, i FInterface, r unsafe.Pointer) bool {
	ip := (*Iface)(unsafe.Pointer(&i))
	tab := ip.Tab
	if tab == nil || tab.Type != t {
		if r != nil {
			_base.Memclr(r, uintptr(t.Size))
		}
		return false
	}
	if r != nil {
		if IsDirectIface(t) {
			Writebarrierptr((*uintptr)(r), uintptr(ip.Data))
		} else {
			Typedmemmove(t, r, ip.Data)
		}
	}
	return true
}

func assertE2T(t *_base.Type, e interface{}, r unsafe.Pointer) {
	ep := (*Eface)(unsafe.Pointer(&e))
	if ep.Type == nil {
		panic(&TypeAssertionError{"", "", *t.String, ""})
	}
	if ep.Type != t {
		panic(&TypeAssertionError{"", *ep.Type.String, *t.String, ""})
	}
	if r != nil {
		if IsDirectIface(t) {
			Writebarrierptr((*uintptr)(r), uintptr(ep.Data))
		} else {
			Typedmemmove(t, r, ep.Data)
		}
	}
}

// The compiler ensures that r is non-nil.
func assertE2T2(t *_base.Type, e interface{}, r unsafe.Pointer) bool {
	ep := (*Eface)(unsafe.Pointer(&e))
	if ep.Type != t {
		_base.Memclr(r, uintptr(t.Size))
		return false
	}
	if IsDirectIface(t) {
		Writebarrierptr((*uintptr)(r), uintptr(ep.Data))
	} else {
		Typedmemmove(t, r, ep.Data)
	}
	return true
}

func convI2E(i FInterface) (r interface{}) {
	ip := (*Iface)(unsafe.Pointer(&i))
	tab := ip.Tab
	if tab == nil {
		return
	}
	rp := (*Eface)(unsafe.Pointer(&r))
	rp.Type = tab.Type
	rp.Data = ip.Data
	return
}

func assertI2E(inter *Interfacetype, i FInterface, r *interface{}) {
	ip := (*Iface)(unsafe.Pointer(&i))
	tab := ip.Tab
	if tab == nil {
		// explicit conversions require non-nil interface value.
		panic(&TypeAssertionError{"", "", *inter.typ.String, ""})
	}
	rp := (*Eface)(unsafe.Pointer(r))
	rp.Type = tab.Type
	rp.Data = ip.Data
	return
}

// The compiler ensures that r is non-nil.
func assertI2E2(inter *Interfacetype, i FInterface, r *interface{}) bool {
	ip := (*Iface)(unsafe.Pointer(&i))
	tab := ip.Tab
	if tab == nil {
		return false
	}
	rp := (*Eface)(unsafe.Pointer(r))
	rp.Type = tab.Type
	rp.Data = ip.Data
	return true
}

func convI2I(inter *Interfacetype, i FInterface) (r FInterface) {
	ip := (*Iface)(unsafe.Pointer(&i))
	tab := ip.Tab
	if tab == nil {
		return
	}
	rp := (*Iface)(unsafe.Pointer(&r))
	if tab.inter == inter {
		rp.Tab = tab
		rp.Data = ip.Data
		return
	}
	rp.Tab = Getitab(inter, tab.Type, false)
	rp.Data = ip.Data
	return
}

func assertI2I(inter *Interfacetype, i FInterface, r *FInterface) {
	ip := (*Iface)(unsafe.Pointer(&i))
	tab := ip.Tab
	if tab == nil {
		// explicit conversions require non-nil interface value.
		panic(&TypeAssertionError{"", "", *inter.typ.String, ""})
	}
	rp := (*Iface)(unsafe.Pointer(r))
	if tab.inter == inter {
		rp.Tab = tab
		rp.Data = ip.Data
		return
	}
	rp.Tab = Getitab(inter, tab.Type, false)
	rp.Data = ip.Data
}

func assertI2I2(inter *Interfacetype, i FInterface, r *FInterface) bool {
	ip := (*Iface)(unsafe.Pointer(&i))
	tab := ip.Tab
	if tab == nil {
		if r != nil {
			*r = nil
		}
		return false
	}
	if tab.inter != inter {
		tab = Getitab(inter, tab.Type, true)
		if tab == nil {
			if r != nil {
				*r = nil
			}
			return false
		}
	}
	if r != nil {
		rp := (*Iface)(unsafe.Pointer(r))
		rp.Tab = tab
		rp.Data = ip.Data
	}
	return true
}

func AssertE2I(inter *Interfacetype, e interface{}, r *FInterface) {
	ep := (*Eface)(unsafe.Pointer(&e))
	t := ep.Type
	if t == nil {
		// explicit conversions require non-nil interface value.
		panic(&TypeAssertionError{"", "", *inter.typ.String, ""})
	}
	rp := (*Iface)(unsafe.Pointer(r))
	rp.Tab = Getitab(inter, t, false)
	rp.Data = ep.Data
}

var testingAssertE2I2GC bool

func AssertE2I2(inter *Interfacetype, e interface{}, r *FInterface) bool {
	if testingAssertE2I2GC {
		GC()
	}
	ep := (*Eface)(unsafe.Pointer(&e))
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
		rp := (*Iface)(unsafe.Pointer(r))
		rp.Tab = tab
		rp.Data = ep.Data
	}
	return true
}

func assertE2E(inter *Interfacetype, e interface{}, r *interface{}) {
	ep := (*Eface)(unsafe.Pointer(&e))
	if ep.Type == nil {
		// explicit conversions require non-nil interface value.
		panic(&TypeAssertionError{"", "", *inter.typ.String, ""})
	}
	*r = e
}

// The compiler ensures that r is non-nil.
func assertE2E2(inter *Interfacetype, e interface{}, r *interface{}) bool {
	ep := (*Eface)(unsafe.Pointer(&e))
	if ep.Type == nil {
		*r = nil
		return false
	}
	*r = e
	return true
}

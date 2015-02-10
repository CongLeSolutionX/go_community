// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package printf

import (
	_core "runtime/internal/core"
	_hash "runtime/internal/hash"
	_ifacestuff "runtime/internal/ifacestuff"
	_lock "runtime/internal/lock"
	"unsafe"
)

func Bytes(s string) (ret []byte) {
	rp := (*_core.Slice)(unsafe.Pointer(&ret))
	sp := (*String)(_core.Noescape(unsafe.Pointer(&s)))
	rp.Array = sp.Str
	rp.Len = uint(sp.len)
	rp.Cap = uint(sp.len)
	return
}

// write to goroutine-local buffer if diverting output,
// or else standard error.
func gwrite(b []byte) {
	if len(b) == 0 {
		return
	}
	gp := _core.Getg()
	if gp == nil || gp.Writebuf == nil {
		writeErr(b)
		return
	}

	n := copy(gp.Writebuf[len(gp.Writebuf):cap(gp.Writebuf)], b)
	gp.Writebuf = gp.Writebuf[:len(gp.Writebuf)+n]
}

func printbool(v bool) {
	if v {
		print("true")
	} else {
		print("false")
	}
}

func printbyte(c byte) {
	gwrite((*[1]byte)(unsafe.Pointer(&c))[:])
}

func printfloat(v float64) {
	switch {
	case v != v:
		print("NaN")
		return
	case v+v == v && v > 0:
		print("+Inf")
		return
	case v+v == v && v < 0:
		print("-Inf")
		return
	}

	const n = 7 // digits printed
	var buf [n + 7]byte
	buf[0] = '+'
	e := 0 // exp
	if v == 0 {
		if 1/v < 0 {
			buf[0] = '-'
		}
	} else {
		if v < 0 {
			v = -v
			buf[0] = '-'
		}

		// normalize
		for v >= 10 {
			e++
			v /= 10
		}
		for v < 1 {
			e--
			v *= 10
		}

		// round
		h := 5.0
		for i := 0; i < n; i++ {
			h /= 10
		}
		v += h
		if v >= 10 {
			e++
			v /= 10
		}
	}

	// format +d.dddd+edd
	for i := 0; i < n; i++ {
		s := int(v)
		buf[i+2] = byte(s + '0')
		v -= float64(s)
		v *= 10
	}
	buf[1] = buf[2]
	buf[2] = '.'

	buf[n+2] = 'e'
	buf[n+3] = '+'
	if e < 0 {
		e = -e
		buf[n+3] = '-'
	}

	buf[n+4] = byte(e/100) + '0'
	buf[n+5] = byte(e/10)%10 + '0'
	buf[n+6] = byte(e%10) + '0'
	gwrite(buf[:])
}

func printcomplex(c complex128) {
	print("(", real(c), imag(c), "i)")
}

func printuint(v uint64) {
	var buf [100]byte
	i := len(buf)
	for i--; i > 0; i-- {
		buf[i] = byte(v%10 + '0')
		if v < 10 {
			break
		}
		v /= 10
	}
	gwrite(buf[i:])
}

func printint(v int64) {
	if v < 0 {
		print("-")
		v = -v
	}
	printuint(uint64(v))
}

func printhex(v uint64) {
	const dig = "0123456789abcdef"
	var buf [100]byte
	i := len(buf)
	for i--; i > 0; i-- {
		buf[i] = dig[v%16]
		if v < 16 {
			break
		}
		v /= 16
	}
	i--
	buf[i] = 'x'
	i--
	buf[i] = '0'
	gwrite(buf[i:])
}

func printpointer(p unsafe.Pointer) {
	printhex(uint64(uintptr(p)))
}

func printstring(s string) {
	if uintptr(len(s)) > _lock.Maxstring {
		gwrite(Bytes("[string too long]"))
		return
	}
	gwrite(Bytes(s))
}

func printslice(s []byte) {
	sp := (*_core.Slice)(unsafe.Pointer(&s))
	print("[", len(s), "/", cap(s), "]")
	printpointer(unsafe.Pointer(sp.Array))
}

func printeface(e interface{}) {
	ep := (*_core.Eface)(unsafe.Pointer(&e))
	print("(", ep.Type, ",", ep.Data, ")")
}

func printiface(i _ifacestuff.FInterface) {
	ip := (*_hash.Iface)(unsafe.Pointer(&i))
	print("(", ip.Tab, ",", ip.Data, ")")
}

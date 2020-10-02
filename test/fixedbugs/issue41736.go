// compile

// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

type I struct {
	x int64
}

type F struct {
	x float64
}

type C struct {
	x *complex128
}

type D struct {
	x complex64
}

//go:noinline
func (i I) X() C {
	cx := complex(0, float64(i.x))
	return C{&cx}
}

//go:noinline
func (f F) X() C {
	cx := complex(f.x, 0)
	return C{&cx}
}

//go:noinline
func (c C) X() C {
	cx := complex(imag(*c.x), real(*c.x))
	return C{&cx}
}

//go:noinline
func (d D) X() C {
	cx := complex(float64(imag(d.x)), -float64(real(d.x)))
	return C{&cx}
}

//go:noinline
func (i I) id() I {
	return i
}

//go:noinline
func (f F) id() F {
	return f
}

//go:noinline
func (c C) id() C {
	return c
}


//go:noinline
func (d D) id() D {
	return d
}


type T interface {
	X() C
}

func G(x []T) []T {
	var y []T
	for _, a := range x {
		var v T
		switch u := a.(type) {
		case I:
			v = u.id()
		case F:
			v = u.id()
		case C:
			v = u.id()
		case D:
			v = u.id()
		}
		y = append(y, v)
	}
	return y
}

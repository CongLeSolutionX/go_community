// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package b

type S struct{}

func (S) M0() (_ struct{ f string }) {
	return
}

func (S) M1() (_ struct{ string }) {
	return
}

func (S) M2() (_ struct{ f struct{ f string } }) {
	return
}

func (S) M3() (_ struct{ F struct{ f string } }) {
	return
}

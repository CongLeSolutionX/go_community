// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package b

type S struct{}

func (S) M0(_ struct{ f string }) {
}

func (S) M1(_ struct{ string }) {
}

func (S) M2(_ struct{ f struct{ f string } }) {
}

func (S) M3(_ struct{ F struct{ f string } }) {
}

type t struct{ A int }

func (S) M4(_ struct {
	*string
}) {
}

func (S) M5(_ struct {
	S
	t
}) {
}

// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

var _ interface{ m() } = struct /* ERR m is a field, not a method */ {
	m func()
}{}

var _ interface{ m() } = & /* ERR m is a field, not a method */ struct {
	m func()
}{}

var _ interface{ M() } = struct /* ERR missing method M */ {
	m func()
}{}

var _ interface{ M() } = & /* ERR missing method M */ struct {
	m func()
}{}

// test case from issue
type I interface{ m() }
type T struct{ m func() }
type M struct{}

func (M) m() {}

func _() {
	var t T
	var m M
	var i I

	i = m
	i = t // ERR m is a field, not a method
	_ = i
}

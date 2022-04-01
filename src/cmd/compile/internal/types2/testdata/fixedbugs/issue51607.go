// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

// test cases from issue

type _ interface {
	interface {bool | int} | interface {bool | string}
}

type _ interface {
	interface {bool | int} ; interface {bool | string}
}

type _ interface {
	interface {bool; int} ; interface {bool; string}
}

type _ interface {
	interface {bool; int} | interface {bool; string}
}
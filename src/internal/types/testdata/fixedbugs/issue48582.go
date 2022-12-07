// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

type N /* ERR invalid recursive type */ interface {
	int | N
}

type A /* ERR invalid recursive type */ interface {
	int | B
}

type B interface {
	int | A
}

type S /* ERR invalid recursive type */ struct {
	I // ERR interface contains type constraints
}

type I interface {
	int | S
}

type P interface {
	*P // ERR interface contains type constraints
}

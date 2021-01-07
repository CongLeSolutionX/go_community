// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package x

type S struct {
	a Key
}

func (s S) A() Key {
	return s.a
}

type Key struct {
	key int64
}

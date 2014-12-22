// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package foo

type Foo struct {
	Bar string
	baz string
}

type Foo2 *Foo

func (Foo) bar() {}

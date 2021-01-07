// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "issue43551.dir/x"

type S   x.S
type Key x.Key

func (s S) A() Key {
	return Key(x.S(s).A())
}

func main() {
	// Do this to get proper "go build" behavior.  It doesn't really need to run.
}

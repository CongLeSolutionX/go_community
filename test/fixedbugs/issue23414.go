// compile

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func main() {
	var s = struct {
		t1 struct{}
		_  int
		t2 struct{}
	}{}
	print(&s.t1 == &s.t2)
}

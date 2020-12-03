// errorcheck

// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 1606.

package main

func main() {
	var x interface{}
	switch t := x.(type) {
	case 0:		// ERROR "type"
<<<<<<< HEAD   (00e572 [dev.regabi] cmd/compile: remove okAs)
		t.x = 1
		x.x = 1 // ERROR "type interface \{\}|reference to undefined field or method"
=======
		t.x = 1 // ERROR "type interface \{\}|reference to undefined field or method|interface with no methods"
>>>>>>> BRANCH (d0c0dc doc/go1.16: document os package changes)
	}
}

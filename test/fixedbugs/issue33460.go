// errorcheck

// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

const (
	zero = iota
	one
	two
	three
)

const iii int = 0x3

func f(v int) {
	switch v {
	case zero, one:
	case two, one: // ERROR "previous case at LINE-1"

	case three:
	case 3: // ERROR "previous case at LINE-1"
	case iii: // ERROR "previous case at LINE-2"
	}
}

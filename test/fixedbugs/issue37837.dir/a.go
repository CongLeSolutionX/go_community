// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package a

func F(i interface{}) int {
	switch i.(type) {
	case nil:
		return 0
	case int:
		return 1
	case float64:
		return 2
	default:
		return 3
	}
}

func G(i interface{}) interface{} {
	switch i := i.(type) {
	case nil:
		return &i
	case int:
		return &i
	case float64:
		return &i
	case string, []byte:
		return &i
	default:
		return &i
	}
}

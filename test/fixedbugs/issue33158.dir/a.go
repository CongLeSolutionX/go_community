// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package a

var GS string

func M() string {
	if s := getname("Fred"); s != "" {
		return s
	}
	if s := getname("Joe"); s != "" {
		return s
	}

	return string("Alex")
}

// could be any function returning a string, just has to be non-inlinable.
func getname(s string) string {
	defer func() { GS = s }()
	if GS == s {
		return GS + "foo"
	}
	return s
}

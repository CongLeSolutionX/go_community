// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gc

func Atoi(s string) int {
	n := 0
	for len(s) > 0 && '0' <= s[0] && s[0] <= '9' {
		n = n*10 + int(s[0]) - '0'
		s = s[1:]
	}
	return n
}

// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gover

import (
	"strings"
)

// Max returns the max of x and y as toolchain names
// like go1.19.4, comparing the versions.
func Max(x, y string) string {
	if Compare(x, y) >= 0 {
		return x
	}
	return y
}

// Compare returns -1, 0, or +1 depending on whether
// x < y, x == y, or x > y, interpreted as toolchain versions.
// The versions x and y may begin with a "go" prefix, as in "go1.21",
// or not, as in "1.21".
// All versions not beginning with "go1" or "1" (for example, "devel ...")
// compare equal and greater than versions that do have those prefixes.
func Compare(x, y string) int {
	if x == y {
		return 0
	}
	if y == "" {
		return +1
	}
	if x == "" {
		return -1
	}
	x = strings.TrimPrefix(x, "go")
	y = strings.TrimPrefix(y, "go")
	if !strings.HasPrefix(x, "1") && !strings.HasPrefix(y, "1") {
		return 0
	}
	if !strings.HasPrefix(x, "1") {
		return +1
	}
	if !strings.HasPrefix(y, "1") {
		return -1
	}
	for x != "" || y != "" {
		if x == y {
			return 0
		}
		xN, xRest := versionCut(x)
		yN, yRest := versionCut(y)
		if xN > yN {
			return +1
		}
		if xN < yN {
			return -1
		}
		x = xRest
		y = yRest
	}
	return 0
}

// versionCut cuts the version x after the next dot or before the next non-digit,
// returning the leading decimal found and the remainder of the string.
func versionCut(x string) (int, string) {
	// Treat empty string as infinite source of .0.0.0...
	if x == "" {
		return 0, ""
	}
	i := 0
	v := 0
	for i < len(x) && '0' <= x[i] && x[i] <= '9' {
		v = v*10 + int(x[i]-'0')
		i++
	}
	// Treat non-empty non-number as -1 (for release candidates, etc),
	// but stop at next number.
	if i == 0 {
		for i < len(x) && (x[i] < '0' || '9' < x[i]) {
			i++
		}
		if i < len(x) && x[i] == '.' {
			i++
		}
		if strings.Contains(x[:i], "alpha") {
			return -3, x[i:]
		}
		if strings.Contains(x[:i], "beta") {
			return -2, x[i:]
		}
		return -1, x[i:]
	}
	if i < len(x) && x[i] == '.' {
		i++
	}
	return v, x[i:]
}

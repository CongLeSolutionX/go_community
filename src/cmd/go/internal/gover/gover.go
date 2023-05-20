// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package gover implements support for Go toolchain versions like 1.21.0 and 1.21rc1.
// (For historical reasons, Go does not use semver for its toolchains.)
// This package provides the same basic analysis that golang.org/x/mod/semver does for semver.
// It also provides some helpers for extracting versions from go.mod files
// and for dealing with module.Versions that may use Go versions or semver
// depending on the module path.
package gover

import "cmp"

// A version is a parsed Go version: major[.minor[.patch]][kind[pre]]
// The numbers are the original decimal strings to avoid integer overflows
// and since there is very little actual math. (Probably overflow doesn't matter in practice,
// but at the time this code was written, there was an existing test that used
// go1.99999999999, which does not fit in an int on 32-bit platforms.
// The "big decimal" representation avoids the problem entirely.)
type version struct {
	major string // decimal
	minor string // decimal or ""
	patch string // decimal or ""
	kind  string // "", "alpha", "beta", "rc"
	pre   string // decimal or ""
}

// Compare returns -1, 0, or +1 depending on whether
// x < y, x == y, or x > y, interpreted as toolchain versions.
// The versions x and y must not begin with a "go" prefix: just "1.21" not "go1.21".
// Malformed versions compare less than well-formed versions and equal to each other.
// The version family "1.21" compares less than
func Compare(x, y string) int {
	vx := parse(x)
	vy := parse(y)

	// patch missing is same as "0" for older versions,
	// empty string sorting before "0" for Go 1.21 and later.
	if vx.patch == "" && vx.pre == "" && lessInt(vx.minor, "21") {
		vx.patch = "0"
	}
	if vy.patch == "" && vy.pre == "" && lessInt(vy.minor, "21") {
		vy.patch = "0"
	}

	if c := cmpInt(vx.major, vy.major); c != 0 {
		return c
	}
	if c := cmpInt(vx.minor, vy.minor); c != 0 {
		return c
	}
	if c := cmpInt(vx.patch, vy.patch); c != 0 {
		return c
	}
	if c := cmp.Compare(vx.kind, vy.kind); c != 0 { // "" < alpha < beta < rc
		return c
	}
	if c := cmpInt(vx.pre, vy.pre); c != 0 {
		return c
	}
	return 0
}

// IsDev reports whether v denotes the overall Go toolchain development branch
// and not a specific release. Starting with the Go 1.21 release, "1.x" denotes
// the overall development branch; the first release is "1.x.0".
// The distinction is important because the relative ordering is
//
//	1.21 < 1.21rc1 < 1.21.0
//
// meaning that Go 1.21rc1 and Go 1.21.0 will both handle go.mod files that
// say "go 1.21", but Go 1.21rc1 will not handle files that say "go 1.21.0".
func IsDev(x string) bool {
	return parse(x).kind == "#dev"
}

// Dev returns the Go dev version. For example, Dev("1.2.3") == "1.2".
// The Dev version is also the language version.
func Dev(x string) string {
	v := parse(x)
	if v.minor == "" {
		return v.major
	}
	return v.major + "." + v.minor
}

// Prev returns the Go major release immediately preceding v,
// or v itself if v is the first Go major release (1.0) or not a supported
// Go version.
//
// Examples:
//
//	Prev("1.2") = "1.1"
//	Prev("1.3rc4") = "1.2"
//
func Prev(x string) string {
	v := parse(x)
	if cmpInt(v.minor, "1") <= 0 {
		return v.major
	}
	return v.major + "." + decInt(v.minor)
}

// IsValid reports whether the version x is valid.
func IsValid(x string) bool {
	return parse(x) != version{}
}

// parse parses the Go version string x into a version.
// It returns the zero version if x is malformed.
func parse(x string) version {
	var v version

	// Parse major version.
	var ok bool
	v.major, x, ok = cutInt(x)
	if !ok {
		return version{}
	}
	if x == "" {
		return v
	}

	// Parse . before minor version.
	if x[0] != '.' {
		return version{}
	}

	// Parse minor version.
	v.minor, x, ok = cutInt(x[1:])
	if !ok {
		return version{}
	}
	if x == "" {
		return v
	}

	// Parse patch if present
	if x[0] == '.' {
		v.patch, x, ok = cutInt(x[1:])
		if !ok || x != "" {
			return version{}
		}
		return v
	}

	// Parse prerelease.
	i := 0
	for i < len(x) && (x[i] < '0' || '9' < x[i]) {
		i++
	}
	if i == 0 {
		return version{}
	}
	v.kind, x = x[:i], x[i:]
	if x == "" {
		return v
	}
	v.pre, x, ok = cutInt(x)
	if !ok || x != "" {
		return version{}
	}

	return v
}

// cutInt scans the leading decimal number at the start of x to an integer
// and returns that value and the rest of the string.
func cutInt(x string) (n, rest string, ok bool) {
	i := 0
	for i < len(x) && '0' <= x[i] && x[i] <= '9' {
		i++
	}
	if i == 0 || x[0] == '0' && i != 1 {
		return "", "", false
	}
	return x[:i], x[i:], true
}

// cmpInt returns cmp.Compare(x, y) interpreting x and y as decimal numbers.
func cmpInt(x, y string) int {
	if lessInt(x, y) {
		return -1
	}
	if lessInt(y, x) {
		return +1
	}
	return 0
}

// lessInt returns x < y interpreting x and y as decimal numbers.
func lessInt(x, y string) bool {
	return len(x) < len(y) || len(x) == len(y) && x < y
}

// decInt returns the string corresponding to x - 1, truncated at 0.
func decInt(x string) string {
	if x == "" || x == "0" {
		return x
	}
	b := []byte(x)
	i := len(b) - 1
	for i >= 0 && b[i] == '0' {
		i--
	}
	if i < 0 {
		return "0"
	}
	b[i]--
	for i++; i < len(b); i++ {
		b[i] = '9'
	}
	if b[0] == '0' {
		b = b[1:]
	}
	return string(b)
}

// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gover

import (
	"strconv"
	"strings"

	"golang.org/x/mod/semver"
)

// Compare returns -1, 0, or +1 depending on whether
// x < y, x == y, or x > y, interpreted as toolchain versions.
// The versions x and y must not begin with a "go" prefix: just "1.21" not "go1.21".
// Malformed versions compare less than well-formed versions and equal to each other.
func Compare(x, y string) int {
	return semver.Compare(ToSemver(x), ToSemver(y))
}

// ToSemver converts the Go version x to an equivalent semver form.
func ToSemver(x string) string {
	if strings.HasPrefix(x, "v") {
		return x
	}
	i := 0
	for i < len(x) && ('0' <= x[i] && x[i] <= '9' || x[i] == '.') {
		i++
	}
	v := "v" + x[:i]
	for strings.Count(v, ".") < 2 {
		v += ".0"
	}
	p := ""
	if i < len(x) {
		if x[i] == '-' {
			p = x[i:]
		} else {
			p = "-" + x[i:]
			j := 0
			for j < len(p) && (p[j] < '0' || '9' < p[j]) {
				j++
			}
			if j < len(p) {
				p = p[:j] + "." + p[j:]
			}
		}
	}
	return semver.Canonical(v + p)
}

// FromSemver converts a semver form returned by ToSemver back to a Go version.
// The round trip may canonicalize the Go version.
//
// For example:
//
//	FromSemver(ToSemver("go1.2.0")) == "go1.2"
//	FromSemver(ToSemver("go1.23")) == "go1.23.0" // trailing .0 started with Go 1.21
func FromSemver(v string) string {
	v = semver.Canonical(v)
	if v == "" {
		return ""
	}
	p := semver.Prerelease(v)
	v = v[len("v") : len(v)-len(p)]

	// Prior to Go 1.21, trailing zeros were dropped.
	// Still dropped for release candidates.
	if min, err := strconv.Atoi(strings.TrimPrefix(semver.MajorMinor("v"+v), "v1.")); (err == nil && min < 21) || p != "" {
		for strings.HasSuffix(v, ".0") {
			v = v[:len(v)-len(".0")]
		}
	}
	if p != "" {
		// Convert -rc.1 to rc1
		p = strings.TrimPrefix(p, "-")
		kind, num, _ := strings.Cut(p, ".")
		p = kind + num
	}
	return v + p

}

// Major returns the Go major version. For example, Major("1.2.3") == "1.2".
// Note that Go terminology differs from semver; in terms of semver this is semver.MajorMinor.
func Major(v string) string {
	i := strings.Index(v, ".")
	if i < 0 {
		return v
	}
	j := strings.Index(v[i+1:], ".")
	if j < 0 {
		return v
	}
	return v[:i+1+j]
}

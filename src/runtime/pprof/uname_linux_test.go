// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux

package pprof

import (
	"fmt"
	"strconv"
	"strings"
	"syscall"
)

func linuxKernelVersion() (major, minor, patch int, err error) {
	var uname syscall.Utsname
	if err := syscall.Uname(&uname); err != nil {
		return 0, 0, 0, err
	}

	var buf [65]byte
	for i, b := range uname.Release {
		buf[i] = byte(b)
	}
	rl := string(buf[:])

	// Ignore everything after a dash (distro sub-version).
	rl, _, _ = strings.Cut(rl, "-")

	// Just the base version number remains (e.g., "5.4.3").
	s := strings.Split(rl, ".")

	v, err := strconv.ParseInt(s[0], 10, 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("error parsing major version %q in %s: %w", s[0], rl, err)
	}
	major = int(v)

	if len(s) >= 2 {
		v, err := strconv.ParseInt(s[1], 10, 64)
		if err != nil {
			return 0, 0, 0, fmt.Errorf("error parsing minor version %q in %s: %w", s[1], rl, err)
		}
		minor = int(v)
	}

	if len(s) >= 3 {
		v, err := strconv.ParseInt(s[2], 10, 64)
		if err != nil {
			return 0, 0, 0, fmt.Errorf("error parsing patch version %q in %s: %w", s[2], rl, err)
		}
		patch = int(v)
	}

	return
}

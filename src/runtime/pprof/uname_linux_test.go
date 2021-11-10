// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux

package pprof

import (
	"fmt"
	"regexp"
	"strconv"
	"syscall"
)

var versionRe = regexp.MustCompile(`^(\d+)(?:\.(\d+)(?:\.(\d+))).*$`)

func linuxKernelVersion() (major, minor, patch int, err error) {
	var uname syscall.Utsname
	if err := syscall.Uname(&uname); err != nil {
		return 0, 0, 0, err
	}

	var buf [65]byte
	last := 0
	for i, b := range uname.Release {
		if b == 0 {
			break
		}
		buf[i] = byte(b)
		last = i
	}
	rl := string(buf[:last])

	m := versionRe.FindStringSubmatch(rl)
	if m == nil {
		return 0, 0, 0, fmt.Errorf("error matching version number in %q", rl)
	}

	v, err := strconv.ParseInt(m[1], 10, 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("error parsing major version %q in %s: %w", m[1], rl, err)
	}
	major = int(v)

	if len(m) >= 3 {
		v, err := strconv.ParseInt(m[2], 10, 64)
		if err != nil {
			return 0, 0, 0, fmt.Errorf("error parsing minor version %q in %s: %w", m[2], rl, err)
		}
		minor = int(v)
	}

	if len(m) >= 4 {
		v, err := strconv.ParseInt(m[3], 10, 64)
		if err != nil {
			return 0, 0, 0, fmt.Errorf("error parsing patch version %q in %s: %w", m[3], rl, err)
		}
		patch = int(v)
	}

	return
}

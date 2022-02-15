// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux || darwin || freebsd || openbsd || netbsd || solaris || dragonfly || aix

package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

const tp = `package t
var t%d = 1`

func TestOpenFileLimit(t *testing.T) {
	if testing.Short() {
		t.Skip("skip in short mode")
	}

	const (
		limit = 20
		count = 50
		fntmp = "d%d.go"
	)

	rl := &syscall.Rlimit{}
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, rl)
	if err != nil {
		t.Error(err)
	}
	if rl.Cur < limit || rl.Max < limit {
		t.Skipf("can't test within open file limit less than %d", limit)
	}
	// recover original limit
	defer func() {
		err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, rl)
		if err != nil {
			t.Error(err)
		}
	}()

	tmpdir := t.TempDir()
	for i := 0; i < count; i++ {
		err = os.WriteFile(filepath.Join(tmpdir, fmt.Sprintf(fntmp, i)),
			[]byte(fmt.Sprintf(tp, i)), 0600)
		if err != nil {
			t.Error(err)
		}
	}

	err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &syscall.Rlimit{Cur: limit, Max: rl.Max})
	if err != nil {
		t.Error(err)
	}

	var buf, errBuf bytes.Buffer
	s := newSequencer(4<<20, &buf, &errBuf)
	for i := 0; i < count; i++ {
		fn := filepath.Join(tmpdir, fmt.Sprintf(fntmp, i))
		info, err := os.Stat(fn)
		if err != nil {
			t.Error(err)
		}
		s.Add(fileWeight(fn, info), func(r *reporter) error {
			return s.processFile(fn, info, nil, r)
		})
	}

	if errBuf.Len() > 0 || s.GetExitCode() != 0 {
		t.Error("format failed:", errBuf.String())
	}

	if buf.Len() == 0 {
		t.Error("nothing formated")
	}
}

// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os_test

import (
	"os"
	"testing"
)

func TestFindProcess_exists(t *testing.T) {
	pid := os.Getpid()
	proc, err := os.FindProcess(pid)
	if err != nil {
		t.Fatalf("Unexpectedly got a non-nil error: %v", err)
	}
	notExistLike := os.IsNotExist(err)
	if notExistLike {
		t.Fatal("os.IsNotExist should have returned false")
	}
	if proc == nil {
		t.Fatal("Expected back a process")
	}
	if g, w := proc.Pid, pid; g != w {
		t.Fatalf("Pid mismatch: got %d want %d", g, w)
	}
}

func TestFindProcess_nonExistent(t *testing.T) {
	proc, err := os.FindProcess(-1)
	if g, w := err, os.ErrProcessNotExist; g != w {
		t.Fatalf("Error mismatch, got %v\nwant: %v", g, w)
	}
	notExistLike := os.IsNotExist(err)
	if !notExistLike {
		t.Fatal("os.IsNotExist should have returned true")
	}
	if proc != nil {
		t.Fatalf("Unexpectedly returned a process: %#v", proc)
	}
}

func BenchmarkFindProcess(b *testing.B) {
	pid := os.Getpid()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		proc, err := os.FindProcess(pid)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
		if proc == nil {
			b.Fatal("Unexpectedly returned nil for process")
		}
	}
}

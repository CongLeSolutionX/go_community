// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Only works on systems with syscall.Close.
// We need a fast system call to provoke the race,
// and Close(-1) is nearly universally fast.

// +build aix darwin dragonfly freebsd linux netbsd openbsd plan9

package runtime_test

import (
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"unsafe"
)

func TestBadOpen(t *testing.T) {
	// make sure we get the correct error code if open fails. Same for
	// read/write/close on the resulting -1 fd. See issue 10052.
	nonfile := []byte("/notreallyafile")
	fd := runtime.Open(&nonfile[0], 0, 0)
	if fd != -1 {
		t.Errorf("open(\"%s\")=%d, want -1", string(nonfile), fd)
	}
	var buf [32]byte
	r := runtime.Read(-1, unsafe.Pointer(&buf[0]), int32(len(buf)))
	if r != -int32(syscall.EBADF) {
		t.Errorf("read()=%d, want %d", r, -int32(syscall.EBADF))
	}
	w := runtime.Write(^uintptr(0), unsafe.Pointer(&buf[0]), int32(len(buf)))
	if w != -int32(syscall.EBADF) {
		t.Errorf("write()=%d, want %d", w, -int32(syscall.EBADF))
	}
	c := runtime.Close(-1)
	if c != -1 {
		t.Errorf("close()=%d, want -1", c)
	}
}

func TestGoroutineProfile(t *testing.T) {
	// GoroutineProfile used to use the wrong starting sp for
	// goroutines coming out of system calls, causing possible
	// crashes.
	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(100))

	var stop uint32
	defer atomic.StoreUint32(&stop, 1) // in case of panic

	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			for atomic.LoadUint32(&stop) == 0 {
				syscall.Close(-1)
			}
			wg.Done()
		}()
	}

	max := 10000
	if testing.Short() {
		max = 100
	}
	stk := make([]runtime.StackRecord, 128)
	for n := 0; n < max; n++ {
		_, ok := runtime.GoroutineProfile(stk)
		if !ok {
			t.Fatalf("GoroutineProfile failed")
		}
	}

	// If the program didn't crash, we passed.
	atomic.StoreUint32(&stop, 1)
	wg.Wait()
}

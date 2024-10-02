// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime_test

import (
	"runtime"
	"testing"
	"unsafe"
)

func TestCleanup(t *testing.T) {
	ch := make(chan bool, 1)
	want := 97531
	done := make(chan bool, 1)
	go func() {
		// allocate struct with pointer to avoid hitting tinyalloc.
		// Otherwise we can't be sure when the allocation will
		// be freed.
		type T struct {
			v int
			p unsafe.Pointer
		}
		v := &new(T).v
		*v = 97531
		cleanup := func(x int) {
			if x != want {
				t.Errorf("cleanup %d, want %d", x, want)
			}
			ch <- true
		}
		runtime.AddCleanup(v, cleanup, 97531)
		v = nil
		done <- true
	}()
	<-done
	runtime.GC()
	<-ch
}

func TestCleanupType(t *testing.T) {}

func TestCleanupInvalidInput(t *testing.T) {}

func TestCleanupZeroSizedStruct(t *testing.T) {
	type Z struct{}
	z := new(Z)
	runtime.AddCleanup(z, func(s string) {}, "foo")
}

func TestCleanupAfterFinalizer(t *testing.T) {}

func BenchmarkCleanup(b *testing.B) {}

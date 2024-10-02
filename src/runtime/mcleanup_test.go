// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime_test

import (
	"runtime"
	"testing"
)

func TestCleanup(t *testing.T) {
	ch := make(chan bool, 2)
	done := make(chan bool, 1)
	go func() {
		obj := new(int)
		*obj = 123
		runtime.AddCleanup(obj, func(b bool) { ch <- b }, true)
		obj = nil
		done <- true
	}()
	<-done
	runtime.GC()
	<-ch
}

func TestCleanupInvalidInput(t *testing.T) {}

func TestCleanupZeroSizedStruct(t *testing.T) {
	type Z struct{}
	z := new(Z)
	runtime.AddCleanup(z, func(s string) {}, "foo")
}

func TestCleanupAfterFinalizer(t *testing.T) {}

func BenchmarkCleanup(b *testing.B) {}

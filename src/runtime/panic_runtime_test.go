// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime_test

import (
	"encoding/binary"
	"runtime"
	"testing"
	"unsafe"
)

func BenchmarkReadvarintUnsafe(b *testing.B) {
	var buf [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(buf[:], 1<<35-1)
	s := unsafe.Slice(&buf[0], n)

	b.Run("old", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			runtime.ReadvarintUnsafeOld(unsafe.Pointer(&s[0]))
		}
	})
	b.Run("new", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			runtime.ReadvarintUnsafe(unsafe.Pointer(&s[0]))
		}
	})
}

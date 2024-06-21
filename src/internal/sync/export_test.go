// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sync

import "unsafe"

// NewBadHashTrieMap creates a new HashTrieMap for the provided key and value
// but with an intentionally bad hash function.
func NewBadHashTrieMap[K, V comparable]() *HashTrieMap[K, V] {
	// Stub out the good hash function with a terrible one.
	// Everything should still work as expected.
	m := NewHashTrieMap[K, V]()
	m.keyHash = func(_ unsafe.Pointer, _ uintptr) uintptr {
		return 0
	}
	return m
}

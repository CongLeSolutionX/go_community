// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cnet

import "internal/trace/testdata/cmd/gclab/heap"

type LAddr64 uint64

func (a LAddr64) ArenaID() heap.ArenaID {
	return heap.ArenaID(heap.Words(a) / heap.ArenaWords)
}

func (a LAddr64) ArenaWord() heap.Words {
	return heap.Words(a) % heap.ArenaWords
}

func (a LAddr64) ToLAddr() heap.LAddr {
	return heap.LAddr(uint64(a) * uint64(heap.WordBytes))
}

type LAddr32 uint32

func (a LAddr32) ArenaWord() heap.Words {
	return heap.Words(a) % heap.ArenaWords
}

func (a LAddr32) ToLAddr(base heap.LAddr) heap.LAddr {
	return heap.LAddr(uint64(a)*uint64(heap.WordBytes)) | base
}

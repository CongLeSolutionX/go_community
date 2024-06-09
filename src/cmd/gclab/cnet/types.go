// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cnet

import (
	"cmd/gclab/heap"
	"fmt"
)

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

func (a LAddr64) String() string {
	aID, aOff := a.ArenaID(), a.ArenaWord().Bytes()
	return fmt.Sprintf("LAddr64(%d/%#07x)", aID, uint64(aOff))
}

type LAddr32 uint32

func (a LAddr32) ArenaWord() heap.Words {
	return heap.Words(a) % heap.ArenaWords
}

func (a LAddr32) ToLAddr(base heap.LAddr) heap.LAddr {
	return heap.LAddr(uint64(a)*uint64(heap.WordBytes)) | base
}

func (a LAddr32) String() string {
	aID := heap.ArenaID(heap.Words(a) / heap.ArenaWords) // Possibly truncated
	aOff := a.ArenaWord().Bytes()
	return fmt.Sprintf("LAddr32(?%d/%#07x)", aID, uint64(aOff))
}

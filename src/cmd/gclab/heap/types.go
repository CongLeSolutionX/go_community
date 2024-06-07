// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package heap

import (
	"fmt"
	"unsafe"
)

// Bytes is a count of bytes or a byte offset.
type Bytes uint64

func (a Bytes) Div(b Bytes) int {
	return int(a / b)
}

func (a Bytes) CeilDiv(b Bytes) int {
	return int((a + b - 1) / b)
}

func (a Bytes) Mul(b int) Bytes {
	return a * Bytes(b)
}

func (a Bytes) Words() Words {
	return Words(a / WordBytes)
}

func (a Bytes) String() string {
	if a == 0 {
		return "0 bytes"
	} else if a%TiB == 0 {
		return fmt.Sprintf("%d TiB", a/TiB)
	} else if a%GiB == 0 {
		return fmt.Sprintf("%d GiB", a/GiB)
	} else if a%MiB == 0 {
		return fmt.Sprintf("%d MiB", a/MiB)
	} else if a%KiB == 0 {
		return fmt.Sprintf("%d KiB", a/KiB)
	}
	return fmt.Sprintf("%d bytes", a)
}

const (
	KiB Bytes = 1 << 10
	MiB Bytes = 1 << 20
	GiB Bytes = 1 << 30
	TiB Bytes = 1 << 40
)

// Words is a count of words or a word offset.
type Words uint64

func (a Words) Bytes() Bytes {
	return Bytes(a) * WordBytes
}

func (a Words) Mul(b int) Words {
	return a * Words(b)
}

func (a Words) Div(b Words) int {
	return int(a / b)
}

// VAddr is a virtual address: a byte offset into the address space.
type VAddr uint64

func (a VAddr) ArenaIndex() int {
	return Bytes(a).Div(ArenaBytes)
}

func (a VAddr) ArenaOffset() Bytes {
	return Bytes(a) % ArenaBytes
}

func (a VAddr) Arena() (aIdx int, aOff Bytes) {
	return Bytes(a).Div(ArenaBytes), Bytes(a) % ArenaBytes
}

func (a VAddr) Page() VPage {
	return VPage(Bytes(a) / PageBytes)
}

func (a VAddr) Plus(b Bytes) VAddr {
	c, ok := a.PlusOK(b)
	if !ok {
		panic(fmt.Sprintf("%s+%s overflowed", a, b))
	}
	return c
}

func (a VAddr) PlusOK(b Bytes) (VAddr, bool) {
	c := a + VAddr(b)
	if c < a {
		return 0, false
	}
	return c, true
}

func (a VAddr) Minus(b VAddr) Bytes {
	c := a - b
	if c > a {
		panic(fmt.Sprintf("%s-%s overflowed", a, b))
	}
	return Bytes(c)
}

func (a VAddr) String() string {
	return fmt.Sprintf("0x%016x", uint64(a))
}

// VPage is a virtual page index.
type VPage uint64

func (p VPage) Start() VAddr {
	return VAddr(Bytes(p) * PageBytes)
}

func (p VPage) Plus(b int) VPage {
	c := p + VPage(b)
	if c < p {
		panic(fmt.Sprintf("%s+%v overflowed", p, b))
	}
	return c
}

func (p VPage) ArenaOffset() int {
	return int(p) % ArenaBytes.Div(PageBytes)
}

type LAddr uint64

const NoLAddr = ^LAddr(0)

func (a LAddr) ArenaID() ArenaID {
	return ArenaID(Bytes(a).Div(ArenaBytes))
}

func (a LAddr) ArenaOffset() Bytes {
	return Bytes(a) % ArenaBytes
}

func (a LAddr) Arena() (aID ArenaID, aOff Bytes) {
	return ArenaID(Bytes(a).Div(ArenaBytes)), Bytes(a) % ArenaBytes
}

func (a LAddr) Plus(b Bytes) LAddr {
	// TODO: Check if we overflowed the arena bounds. For faster batch
	// operations, I could add a check bit between the arena ID and the offset
	// and then just OR together a bunch of check bits and do a single "did
	// something go wrong" check. Though that halves the space I can store in an
	// LAddr32.
	c := a + LAddr(b)
	if c < a {
		panic(fmt.Sprintf("%s+%s overflowed", a, b))
	}
	return c
}

func (a LAddr) FloorDiv(b Bytes) int {
	return int(Bytes(a) / b)
}

func (a LAddr) String() string {
	if a == NoLAddr {
		return "NoLAddr"
	}
	aID, aOff := a.Arena()
	return fmt.Sprintf("LAddr(%d/%#07x)", aID, uint64(aOff))
}

type Range struct {
	Start VAddr
	Len   Bytes
}

var FullRange = Range{0, ^Bytes(0)}

func (r Range) End() VAddr {
	end, ok := r.Start.PlusOK(r.Len)
	if !ok {
		panic(fmt.Sprintf("range end overflowed: %s", r))
	}
	return end
}

func (r Range) Contains(x VAddr) bool {
	return r.Start < x && x.Minus(r.Start) < r.Len
}

func (r Range) Overlaps(r2 Range) bool {
	return r.Start < r2.End() && r2.Start < r.End()
}

func (r Range) String() string {
	return fmt.Sprintf("[%s,%s)", r.Start, r.End())
}

func (r Range) ShortString() string {
	return fmt.Sprintf("[%#x,%#x)", uint64(r.Start), uint64(r.End()))
}

func CastSlice[To any](src []byte) []To {
	// TODO: It would be nice if we could limit this to pointer-free types. That
	// would make this "safe".
	d := (*To)(unsafe.Pointer(unsafe.SliceData(src)))
	return unsafe.Slice(d, len(src)/int(unsafe.Sizeof(*d)))
}

// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rand

import _ "unsafe" // for linkname

type ChaCha8 struct {
	seed    [32]byte
	buf     [32]uint64
	counter uint64
	n       uint
}

func NewChaCha8(seed [32]byte) *ChaCha8 {
	c := new(ChaCha8)
	c.Seed(seed)
	return c
}

func (c *ChaCha8) Seed(seed [32]byte) {
	c.seed = seed
}

func (c *ChaCha8) Uint64() uint64 {
	if c.n == 0 {
		c.refill()
	}
	// c.n is how many bytes are left in buffer.
	c.n--
	return c.buf[uint(len(c.buf)-1)-c.n]
}

//go:noinline
func (c *ChaCha8) refill() {
	chacha8block(c.counter, &c.seed, &c.buf)
	c.counter += 4
	c.n = uint(len(c.buf))
}

//go:linkname chacha8block runtime.chacha8block
func chacha8block(counter uint64, seed *[32]byte, blocks *[32]uint64)

// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cnet

import "cmd/gclab/heap"

type condenser struct {
	width       uint
	mask        uint64
	startDstBit int

	steps []condenseStep
}

type condenseStep struct {
	// Each step processes one word of the source and emits to one or two words
	// of dst.

	// If botMask != 0, then the low order bits of this source word are a
	// partial field.
	botMask  uint64
	botShift uint64

	// Then there are nFields complete fields in the source word that are
	// emitted to the same output word.
	nFields int

	// If nextFields != 0, then move on to the next output word and process
	// another nextFields complete fields.
	nextFields int
}

// newCondenser returns a condenser that takes an input src that's limit bits
// long and an output dst and sets bit i of dst to the logical OR of src bits
// [i*width, (i+1)*width).
func newCondenser(width, limit heap.Words) *condenser {
	var steps []condenseStep
	dstFree := 64
	for sbit := heap.Words(0); sbit < limit; sbit += 64 {
		sWidth := min(64, limit-sbit)

		// Process any low-order bits of this word that are part of the previous
		// field.
		var step condenseStep
		if botBits := (sbit+width-1)/width*width - sbit; botBits != 0 {
			if botBits >= 64 {
				// OR the whole thing in.
				steps = append(steps, condenseStep{botMask: ^uint64(0)})
				// XXX Make sure we don't advance obit.
				continue
			}

			// Get the bottom botBits bits
			step.botMask = uint64(1)<<botBits - 1
			step.botShift = uint64(botBits)
			sWidth -= botBits
		}

		// How many complete fields can we get?
		step.nFields = int(sWidth+width-1) / int(width)
		// How many can we fit in dst?
		if step.nFields < dstFree {
			dstFree -= step.nFields
		} else if step.nFields == dstFree {
			// Flush dst, but there aren't any more fields in src.
			step.nextFields = -1
			dstFree = 64
		} else {
			step.nextFields = step.nFields - dstFree
			step.nFields = dstFree
			dstFree = 64 - step.nextFields
		}

		steps = append(steps, step)
	}
	// The final flush is implied.
	if steps[len(steps)-1].nextFields == -1 {
		steps[len(steps)-1].nextFields = 0
	}

	mask := ^uint64(0)
	if width < 64 {
		mask = 1<<width - 1
	}
	return &condenser{uint(width), mask, 0, steps}
}

// slice returns a new condenser that only processes bits [start,end) of src.
// start and end must be multiples of 64.
// If end is ^0, it is ignored.
func (c *condenser) slice(start, end heap.Words) *condenser {
	if start%64 != 0 {
		panic("start must be a multiple of 64")
	}
	if end == ^heap.Words(0) {
		end = heap.Words(len(c.steps) * 64)
	}
	if end%64 != 0 {
		panic("end must be a multiple of 64")
	}
	c2 := *c
	c2.startDstBit = (int(start) + int(c.width) - 1) / int(c.width)
	c2.steps = c.steps[start/64 : end/64]
	return &c2
}

func (c *condenser) do(src, dst []uint64) {
	// We process src because it's easy to OR more stuff in dst, but it's hard
	// to pick out all the pieces of src we need for a particular bit of dst.
	var out uint64
	oidx := c.startDstBit / 64
	obit := uint(c.startDstBit % 64)
	mask := c.mask
	_ = src[len(c.steps)-1]
	for i, step := range c.steps {
		sw := src[i]

		// Process any low-order bits of sw that are part of the previous field
		// and align sw to the start of a field.
		if step.botMask != 0 {
			val := sw & step.botMask
			out |= uint64(bool2int(val != 0)) << (obit - 1)
			sw >>= step.botShift
		}

		// Remaining fields
		for range step.nFields {
			val := sw & mask
			out |= uint64(bool2int(val != 0)) << (obit % 64)
			obit++
			sw >>= c.width % 64
		}

		if step.nextFields != 0 {
			// Flush to dst
			dst[oidx] = out
			oidx++
			out = 0
			obit = 0
			// Process remaining fields
			for range step.nextFields {
				val := sw & mask
				out |= uint64(bool2int(val != 0)) << (obit % 64)
				obit++
				sw >>= c.width % 64
			}
		}
	}
	// Final flush to dst
	dst[oidx] = out
}

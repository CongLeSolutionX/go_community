// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

import (
	"cmd/internal/obj"
	"sort"
	"sync"
)

// A Cache holds reusable compiler state.
// It is intended to be re-used for multiple Func compilations.
type Cache struct {
	// batch allocator for various types
	nValues      int
	valueBatches []*[valueBatchSize]Value
	nBlocks      int
	blockBatches []*[blockBatchSize]Block

	// Storage for regalloc result
	locs [2000]Location

	// Reusable stackAllocState.
	// See stackalloc.go's {new,put}StackAllocState.
	stackAllocState *stackAllocState

	domblockstore   []ID            // scratch space for computing dominators
	scrSparseSet    []*sparseSet    // scratch sparse sets to be re-used.
	scrSparseMap    []*sparseMap    // scratch sparse maps to be re-used.
	scrSparseMapPos []*sparseMapPos // scratch sparse maps to be re-used.
	scrPoset        []*poset        // scratch poset to be reused
	// deadcode contains reusable slices specifically for the deadcode pass.
	// It gets special treatment because of the frequency with which it is run.
	deadcode struct {
		liveOrderStmts []*Value
		live           []bool
		q              []*Value
	}
	// Reusable regalloc state.
	regallocValues []valState

	ValueToProgAfter []*obj.Prog
	debugState       debugState

	Liveness interface{} // *gc.livenessFuncCache
}

const valueBatchSize = 256
const blockBatchSize = 64

var valueBatchPool = sync.Pool{New: func() interface{} { return new([valueBatchSize]Value) }}
var blockBatchPool = sync.Pool{New: func() interface{} { return new([blockBatchSize]Block) }}

func (c *Cache) allocValue() *Value {
	// Grab a value from the cache allocator.
	n := c.nValues
	batchId, batchOff := uint(n)/valueBatchSize, uint(n)%valueBatchSize
	if batchId == uint(len(c.valueBatches)) {
		c.valueBatches = append(c.valueBatches, valueBatchPool.Get().(*[valueBatchSize]Value))
	}
	c.nValues = n + 1
	return &c.valueBatches[batchId][batchOff]
}

func (c *Cache) allocBlock() *Block {
	// Grab a block from the cache allocator.
	n := c.nBlocks
	batchId, batchOff := uint(n)/blockBatchSize, uint(n)%blockBatchSize
	if batchId == uint(len(c.blockBatches)) {
		c.blockBatches = append(c.blockBatches, blockBatchPool.Get().(*[blockBatchSize]Block))
	}
	c.nBlocks = n + 1
	return &c.blockBatches[batchId][batchOff]
}

func (c *Cache) Reset() {
	// Zero the values we've used.
	for i, b := range c.valueBatches {
		n := c.nValues - i*valueBatchSize
		if n < valueBatchSize {
			// zero partial batch
			s := b[:n]
			for j := range s {
				s[j] = Value{}
			}
		} else {
			// zero full batch
			*b = [valueBatchSize]Value{}
		}
		valueBatchPool.Put(b)
	}
	c.valueBatches = nil
	c.nValues = 0

	// Zero the blocks we've used.
	for i, b := range c.blockBatches {
		n := c.nBlocks - i*blockBatchSize
		if n < blockBatchSize {
			// zero partial batch
			s := b[:n]
			for j := range s {
				s[j] = Block{}
			}
		} else {
			// zero full batch
			*b = [blockBatchSize]Block{}
		}
		blockBatchPool.Put(b)
	}
	c.blockBatches = nil
	c.nBlocks = 0

	nl := sort.Search(len(c.locs), func(i int) bool { return c.locs[i] == nil })
	xl := c.locs[:nl]
	for i := range xl {
		xl[i] = nil
	}

	// regalloc sets the length of c.regallocValues to whatever it may use,
	// so clear according to length.
	for i := range c.regallocValues {
		c.regallocValues[i] = valState{}
	}

	// liveOrderStmts gets used multiple times during compilation of a function.
	// We don't know where the high water mark was, so reslice to cap and search.
	c.deadcode.liveOrderStmts = c.deadcode.liveOrderStmts[:cap(c.deadcode.liveOrderStmts)]
	no := sort.Search(len(c.deadcode.liveOrderStmts), func(i int) bool { return c.deadcode.liveOrderStmts[i] == nil })
	xo := c.deadcode.liveOrderStmts[:no]
	for i := range xo {
		xo[i] = nil
	}
	c.deadcode.q = c.deadcode.q[:cap(c.deadcode.q)]
	nq := sort.Search(len(c.deadcode.q), func(i int) bool { return c.deadcode.q[i] == nil })
	xq := c.deadcode.q[:nq]
	for i := range xq {
		xq[i] = nil
	}
}

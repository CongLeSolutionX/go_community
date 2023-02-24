// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package json

import (
	"bytes"
	"io"
	"math/bits"
	"slices"
	"sync"
)

// pooledBuffer is similar to bytes.Buffer,
// but uses a series of segmented buffers instead of a single contiguous buffer.
// Each segment never has a smaller capacity than the previous buffer and
// are retrieved from a tiered list of sync.Pool as needed.
//
// Invariant: len(pooledBuffer.fullList()) <= maxNumSegments
type pooledBuffer struct {
	list [][]byte // list of segments except the last one
	last []byte
}

func (b *pooledBuffer) fullList() [][]byte {
	if cap(b.last) > 0 {
		b.list = slices.Grow(b.list, 1)
		return append(b.list, b.last) // avoid changing length of b.list
	}
	return b.list
}

func (b *pooledBuffer) Len() (n int) {
	if len(b.list) == 0 {
		return len(b.last)
	}
	for _, p := range b.fullList() {
		n += len(p)
	}
	return n
}

func (b *pooledBuffer) Cap() int                { return b.Len() + b.Available() }
func (b *pooledBuffer) Available() int          { return cap(b.AvailableBuffer()) }
func (b *pooledBuffer) AvailableBuffer() []byte { return b.last[len(b.last):] }

func (b *pooledBuffer) Grow(n int) {
	if cap(b.last)-len(b.last) < n {
		b.growSlow(n)
	}
}
func (b *pooledBuffer) growSlow(n int) {
	if cap(b.last) > 0 {
		if n <= cap(b.last) {
			n = cap(b.last) // ensure segments never decrease in capacity
			if n < maxRetainSegmentSize {
				n *= 2 // only double capacity until maxRetainSegmentSize
			}
		}
		if len(b.last) == 0 {
			putBuffer(b.last)
		} else {
			b.list = append(b.list, b.last)
		}
	}
	b.last = getBuffer(n)
}

func (b *pooledBuffer) Write(p []byte) (int, error) {
	b.Grow(len(p))
	b.last = b.last[:len(b.last)+len(p)]
	copy(b.last[len(b.last)-len(p):], p)
	return len(p), nil
}

func (b *pooledBuffer) WriteString(s string) (int, error) {
	b.Grow(len(s))
	b.last = b.last[:len(b.last)+len(s)]
	copy(b.last[len(b.last)-len(s):], s)
	return len(s), nil
}

func (b *pooledBuffer) WriteByte(c byte) error {
	b.Grow(1)
	b.last = append(b.last, c)
	return nil
}

func (b *pooledBuffer) WriteTo(w io.Writer) (n int64, err error) {
	for _, p := range b.fullList() {
		m, err := w.Write(p)
		if err != nil {
			return n, err
		}
		n += int64(m)
	}
	return n, err
}

// Bytes returns the buffer content as a single contiguous buffer.
// It may need to merge segments to produce a contiguous buffer.
func (b *pooledBuffer) Bytes() []byte {
	if len(b.list) == 0 {
		return b.last
	}
	p := getBuffer(b.Len())
	for i, list := 0, b.fullList(); i < len(list); i++ {
		p = append(p, list[i]...)
		putBuffer(list[i])
		list[i] = nil // allow GC to reclaim the buffer
	}
	b.list = b.list[:0]
	b.last = p
	return b.last
}

// BytesClone returns a copy of the buffer content.
func (b *pooledBuffer) BytesClone() []byte {
	if len(b.list) == 0 {
		return bytes.Clone(b.last)
	}
	return bytes.Join(b.fullList(), nil)
}

func (b *pooledBuffer) Reset() {
	// Return all segments to the pools except one segment,
	// which may be retained locally if sufficiently small.
	if len(b.list) == 0 && cap(b.last) <= maxRetainSegmentSize {
		b.last = b.last[:0]
		return
	}
	var retain []byte
	list := b.fullList()
	for i := len(list) - 1; i >= 0; i-- {
		if retain == nil && cap(list[i]) <= maxRetainSegmentSize {
			retain = list[i][:0] // retain locally, but clear the length
		} else {
			putBuffer(list[i])
		}
		list[i] = nil // allow GC to reclaim the buffer
	}
	b.list = b.list[:0]
	b.last = retain
}

const (
	minPooledSegmentShift = 12 // minumum size of buffer to pool
	maxRetainSegmentShift = 16 // maximum size of buffer to retain locally in buffer
	maxRetainSegmentSize  = 1 << maxRetainSegmentShift
	maxNumSegments        = bits.UintSize - minPooledSegmentShift
)

// TODO(https://go.dev/issue/47657): Use sync.PoolOf.
// The putBuffer must allocate a slice header whenever calling sync.Pool.Put.
// Using *[]byte instead of []byte avoids the allocation,
// but complicates the logic of pooledBuffer and is slower as a result.

// bufferPools is a list of buffer pools.
// Each pool manages buffers of capacity within [1<<shift : 2<<shift),
// where shift is (index+minPooledSegmentShift).
var bufferPools [maxNumSegments]sync.Pool

// getBuffer acquires an empty buffer with enough capacity to hold n bytes.
func getBuffer(n int) []byte {
	if n < 1<<minPooledSegmentShift {
		n = 1 << minPooledSegmentShift
	}
	shift := bits.Len(uint(n - 1))
	if p, _ := bufferPools[shift-minPooledSegmentShift].Get().([]byte); p != nil {
		return p[:0]
	}
	return make([]byte, 0, 1<<shift)
}

// putBuffer releases a buffer back to the pools.
func putBuffer(p []byte) {
	if cap(p) < 1<<minPooledSegmentShift {
		return
	}
	shift := bits.Len(uint(cap(p)) - 1)
	// TODO: In race detector mode, asynchronously write to the buffer to detect
	// buffers that may have accidentally leaked to users.
	// See https://go.dev/issue/58452 for inspiration.
	bufferPools[shift-minPooledSegmentShift].Put(p)
}

// Copyright 2017 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	"runtime/internal/atomic"
	"unsafe"
)

// A profBuf is a lock-free buffer for profiling events,
// safe for concurrent use by one reader and one writer.
// The writer may be a signal handler running without a user g.
// The reader is assumed to be a user g.
//
// Each logged event corresponds to a fixed size header, a list of
// uintptrs (typically a stack), and exactly one unsafe.Pointer tag.
// The header and uintptrs are stored in the circular buffer data and the
// tag is stored in a circular buffer tags, running in parallel.
// In the circular buffer data, each event takes 2+hdrsize+len(stk)
// words: the value 2+hdrsize+len(stk), then the time of the event, then
// hdrsize words giving the fixed-size header, and then len(stk) words
// for the stack.
//
// The current effective offsets into the tags and data circular buffers
// for reading and writing are stored in the high 30 and low 32 bits of r and w.
// The bottom bits of the high 32 are additional flag bits in w, unused in r.
// "Effective" offsets means the total number of reads or writes, mod 2^length.
// The offset in the buffer is the effective offset mod the length of the buffer.
// To make wraparound mod 2^length match wraparound mod length of the buffer,
// the length of the buffer must be a power of two.
//
// If the reader catches up to the writer, a flag passed to read controls
// whether the read blocks until more data is available. A read returns a
// pointer to the buffer data itself; the caller is assumed to be done with
// that data at the next read. The read offset rNext tracks the next offset to
// be returned by read. By definition, r ≤ rNext ≤ w (before wraparound),
// and rNext is only used by the reader, so it can be accessed without atomics.
//
// If the writer gets ahead of the reader, so that the buffer fills,
// future writes are discarded and replaced in the output stream by an
// overflow entry, which has size 2+hdrsize+1, time set to the time of
// the first discarded write, a header of all zeroed words, and a "stack"
// containing one word, the number of discarded writes.
//
// Between the time the buffer fills and the buffer becomes empty enough
// to hold more data, the overflow entry is stored as a pending overflow
// entry in the fields overflow and overflowTime. The pending overflow
// entry can be turned into a real record by either the writer or the
// reader. If the writer is called to write a new record and finds that
// the output buffer has room for both the pending overflow entry and the
// new record, the writer emits the pending overflow entry and the new
// record into the buffer. If the reader is called to read data and finds
// that the output buffer is empty but that there is a pending overflow
// entry, the reader will return a synthesized record for the pending
// overflow entry.
//
// Only the writer can create or add to a pending overflow entry, but
// either the reader or the writer can clear the pending overflow entry.
// A pending overflow entry is indicated by the low 32 bits of 'overflow'
// holding the number of discarded writes, and overflowTime holding the
// time of the first discarded write. The high 32 bits of 'overflow'
// increment each time the low 32 bits transition from zero to non-zero
// or vice versa. This sequence number avoids ABA problems in the use of
// compare-and-swap to coordinate between reader and writer.
// The overflowTime is only written when the low 32 bits of overflow are
// zero, that is, only when there is no pending overflow entry, in
// preparation for creating a new one. The reader can therefore fetch and
// clear the entry atomically using
//
//	for {
//		overflow = load(&b.overflow)
//		if uint32(overflow) == 0 {
//			// no pending entry
//			break
//		}
//		time = load(&b.overflowTime)
//		if cas(&b.overflow, overflow, ((overflow>>32)+1)<<32) {
//			// pending entry cleared
//			break
//		}
//	}
//	if uint32(overflow) > 0 {
//		emit entry for uint32(overflow), time
//	}
//
type profBuf struct {
	// accessed atomically
	r, w         uint64
	overflow     uint64
	overflowTime uint64
	eof          uint32

	// immutable (excluding slice content)
	hdrsize uintptr
	data    []uint64
	tags    []unsafe.Pointer

	// owned by reader
	rNext       uint64
	overflowBuf []uint64 // for use by reader to return overflow record
	wait        note
}

const (
	profReaderSleeping uint64 = 1 << 32
	profWriteExtra     uint64 = 1 << 33 // overflow or eof waiting
)

func profDataCount(x uint64) uint32 {
	return uint32(x)
}

func profTagCount(x uint64) uint32 {
	return uint32(x >> 34)
}

func profAddCounts(x uint64, data, tag uint32) uint64 {
	return (x>>34+uint64(tag))<<34 | uint64(uint32(x)+data)
}

// newProfBuf returns a new profiling buffer with room for
// a header of hdrsize words and a buffer of at least bufwords words.
func newProfBuf(hdrsize, bufwords, tags int) *profBuf {
	if min := 2 + hdrsize + 1; bufwords < min {
		bufwords = min
	}

	// Buffer sizes must be power of two, so that we don't have to
	// worry about uint32 wraparound changing the effective position
	// within the buffers.
	if bufwords >= 1<<30 || tags >= 1<<30 {
		panic("newProfBuf: buffer too large")
	}
	var i int
	for i = 1; i < bufwords; i <<= 1 {
	}
	bufwords = i
	for i = 1; i < tags; i <<= 1 {
	}
	tags = i

	b := new(profBuf)
	b.hdrsize = uintptr(hdrsize)
	b.data = make([]uint64, bufwords)
	b.tags = make([]unsafe.Pointer, tags)
	b.overflowBuf = make([]uint64, 2+b.hdrsize+1)
	return b
}

// canWriteRecord reports whether the buffer has room
// for a single record with a stack of length nstk.
func (b *profBuf) canWriteRecord(nstk int) bool {
	br := uint64(atomic.Load64(&b.r))
	bw := uint64(atomic.Load64(&b.w))

	// room for tag?
	if uintptr(profTagCount(br))+uintptr(len(b.tags))-uintptr(profTagCount(bw)) < 1 {
		return false
	}

	// room for data?
	nd := int(uintptr(profDataCount(br)) + uintptr(len(b.data)) - uintptr(profDataCount(bw)))
	want := 2 + int(b.hdrsize) + nstk
	i := int(uintptr(profDataCount(bw)) % uintptr(len(b.data)))
	if i+want > len(b.data) {
		// account for loss of trailing fragment of slice
		nd -= len(b.data) - i
	}
	return nd >= want
}

// canWriteTwoRecords reports whether the buffer has room
// for two records with stack lengths nstk1, nstk2, in that order.
func (b *profBuf) canWriteTwoRecords(nstk1, nstk2 int) bool {
	br := uint64(atomic.Load64(&b.r))
	bw := uint64(atomic.Load64(&b.w))

	// room for tag?
	if uintptr(profTagCount(br))+uintptr(len(b.tags))-uintptr(profTagCount(bw)) < 2 {
		return false
	}

	// room for data?
	nd := int(uintptr(profDataCount(br)) + uintptr(len(b.data)) - uintptr(profDataCount(bw)))

	// first record
	want := 2 + int(b.hdrsize) + nstk1
	i := int(uintptr(profDataCount(bw)) % uintptr(len(b.data)))
	if i+want > len(b.data) {
		// account for loss of trailing fragment of slice
		nd -= len(b.data) - i
		i = 0
	}
	i += want
	nd -= want

	// second record
	want = 2 + int(b.hdrsize) + nstk2
	if i+want > len(b.data) {
		nd -= len(b.data) - i
		i = 0
	}
	return nd >= want
}

// write writes an entry to the profiling buffer b.
// The entry begins with a fixed hdr, which must have
// length b.hdrsize, followed by a variable-sized stack
// and a single tag pointer *tagptr (or nil if tagptr is nil).
// The tagptr is read while holding tagLock as a cas-based spin lock,
// for garbage collection / write barrier synchronization. (See comment below.)
// No write barriers allowed because this might be called from a signal handler.
//go:nowritebarrierrec
func (b *profBuf) write(tagptr *unsafe.Pointer, tagLock *uint32, now int64, hdr []uint64, stk []uintptr) {
	if b == nil {
		return
	}
	if len(hdr) > int(b.hdrsize) {
		throw("misuse of profBuf.write")
	}

	if overflow := atomic.Load64(&b.overflow); uint32(overflow) > 0 && b.canWriteTwoRecords(1, len(stk)) {
		// Room for both an overflow record and the one being written.
		// Write the overflow record if the reader hasn't gotten to it yet.
		// Only racing against reader, not other writers.
		for {
			// Increment generation, clear overflow count in low bits.
			if atomic.Cas64(&b.overflow, overflow, ((overflow>>32)+1)<<32) {
				break
			}
			overflow = atomic.Load64(&b.overflow)
		}
		if uint32(overflow) > 0 {
			var count [1]uintptr
			count[0] = uintptr(uint32(overflow))
			b.write(nil, nil, int64(atomic.Load64(&b.overflowTime)), nil, count[:])
		}
	} else if uint32(overflow) > 0 || !b.canWriteRecord(len(stk)) {
		// Pending overflow without room to write overflow and new records
		// or no overflow but also no room for new record.
		for {
			// We need to set overflowTime if we're incrementing b.overflow from 0.
			// We're racing with reader who wants to set b.overflow to 0.
			// Once we see b.overflow reach 0, it's stable: no one else is changing it underfoot.
			if uint32(overflow) != 0 && atomic.Cas64(&b.overflow, overflow, overflow+1) {
				break
			}
			atomic.Store64(&b.overflowTime, uint64(now))
			atomic.Store64(&b.overflow, (((overflow>>32)+1)<<32)+1)
			break
		}
		b.wakeupExtra()
		return
	}

	// There's room: write the record.
	br := uint64(atomic.Load64(&b.r))
	bw := uint64(atomic.Load64(&b.w))

	// Profiling tag
	//
	// The tag is a pointer, but we can't run a write barrier here.
	// Instead we lock the tag field in the g so that it cannot be
	// overwritten while we copy it. Doing so ensures that at the
	// time our copy finishes, we know the tag we copied is still
	// visible to the GC as gp.labels.
	// From that point on, any new GC will see the tag in b.tags,
	// and any GC in progress could only correctly collect the tag
	// if gp.labels is overwritten, in which case the deletion barrier
	// will mark it.
	// We also arrange that the entry we're overwriting in b.tags
	// is nil, so there is no need for a deletion barrier on the write.
	wt := profTagCount(bw) % uint32(len(b.tags))
	if tagptr != nil {
		for !atomic.Cas(tagLock, 0, 1) {
			// loop, we're possibly in a signal handler
		}
		*(*uintptr)(unsafe.Pointer(&b.tags[wt])) = uintptr(unsafe.Pointer(*tagptr))
		atomic.Store(tagLock, 0)
	}

	// Main record.
	// It has to fit in a contiguous section of the slice, so if it doesn't fit at the end,
	// leave a rewind marker (0) and start over at the beginning of the slice.
	wd := uintptr(profDataCount(bw) % uint32(len(b.data)))
	nd := uintptr(profDataCount(br)) + uintptr(len(b.data)) - uintptr(profDataCount(bw))
	skip := uintptr(0)
	if wd+2+b.hdrsize+uintptr(len(stk)) > uintptr(len(b.data)) {
		b.data[wd] = 0
		skip = uintptr(len(b.data)) - wd
		nd -= skip
		wd = 0
	}
	data := b.data[wd:]
	data[0] = uint64(2 + b.hdrsize + uintptr(len(stk))) // length
	data[1] = uint64(now)                               // time stamp
	// header, zero-padded
	i := uintptr(copy(data[2:2+b.hdrsize], hdr))
	for ; i < b.hdrsize; i++ {
		data[2+i] = 0
	}
	for i, pc := range stk {
		data[2+b.hdrsize+uintptr(i)] = uint64(pc)
	}

	for {
		// Commit write.
		old := atomic.Load64(&b.w)
		new := profAddCounts(old, uint32(skip+2+b.hdrsize+uintptr(len(stk))), 1)
		if !atomic.Cas64(&b.w, old, new) {
			continue
		}
		// If there was a reader, wake it up.
		if old&profReaderSleeping != 0 {
			notewakeup(&b.wait)
		}
		break
	}
}

// close signals that there will be no more writes on the buffer.
// Once all the data has been read from the buffer, reads will return eof=true.
func (b *profBuf) close() {
	if atomic.Load(&b.eof) > 0 {
		panic("runtime: profBuf already closed")
	}
	atomic.Store(&b.eof, 1)
	b.wakeupExtra()
}

// wakeupExtra must be called after setting one of the "extra"
// atomic fields b.overflow or b.eof.
// It records the change in b.w and wakes up the reader if needed.
func (b *profBuf) wakeupExtra() {
	for {
		old := uint64(atomic.Load64(&b.w))
		new := old | profWriteExtra
		if !atomic.Cas64(&b.w, old, new) {
			continue
		}
		if old&profReaderSleeping != 0 {
			notewakeup(&b.wait)
		}
		break
	}
}

// profBufReadMode specifies whether to block when no data is available to read.
type profBufReadMode int

const (
	profBufBlocking profBufReadMode = iota
	profBufNonBlocking
)

var overflowTag [1]unsafe.Pointer // always nil

func (b *profBuf) read(mode profBufReadMode) (data []uint64, tags []unsafe.Pointer, eof bool) {
	if b == nil {
		return nil, nil, true
	}

	br := b.rNext

	// Commit previous read.
	// First clear tags that have now been read, both to avoid holding
	// up the memory they point at for longer than necessary
	// and so that b.write can assume it is always overwriting
	// nil tag entries (see comment in b.write).
	rPrev := atomic.Load64(&b.r)
	if rPrev != br {
		ntag := uintptr(profTagCount(br) - profTagCount(rPrev))
		ti := int(profTagCount(rPrev) % uint32(len(b.tags)))
		for i := uintptr(0); i < ntag; i++ {
			b.tags[ti] = nil
			if ti++; ti == len(b.tags) {
				ti = 0
			}
		}
		atomic.Store64(&b.r, br)
	}

Read:
	bw := uint64(atomic.Load64(&b.w))
	numData := uintptr(profDataCount(bw) - profDataCount(br))
	if numData == 0 {
		if atomic.Load64(&b.overflow) > 0 {
			// No data to read, but there is overflow to report.
			// Racing with writer flushing b.overflow into a real record.
			overflow := atomic.Load64(&b.overflow)
			time := atomic.Load64(&b.overflowTime)
			for {
				if uint32(overflow) == 0 {
					// Lost the race, go around again.
					goto Read
				}
				if atomic.Cas64(&b.overflow, overflow, ((overflow>>32)+1)<<32) {
					break
				}
				overflow = atomic.Load64(&b.overflow)
				time = atomic.Load64(&b.overflowTime)
			}

			// Won the race, report overflow.
			dst := b.overflowBuf
			dst[0] = uint64(2 + b.hdrsize + 1)
			dst[1] = uint64(time)
			for i := uintptr(0); i < b.hdrsize; i++ {
				dst[2+i] = 0
			}
			dst[2+b.hdrsize] = uint64(uint32(overflow))
			return dst[:2+b.hdrsize+1], overflowTag[:1], false
		}
		if atomic.Load(&b.eof) > 0 {
			// No data, no overflow, EOF set: done.
			return nil, nil, true
		}
		if bw&profWriteExtra != 0 {
			// Writer claims to have published extra information (overflow or eof).
			// Attempt to clear notification and then check again.
			// If we fail to clear the notification it means w64 changed,
			// so we still need to check again.
			atomic.Cas64(&b.w, bw, bw&^profWriteExtra)
			goto Read
		}

		// Nothing to read right now.
		// Return or sleep according to mode.
		if mode == profBufNonBlocking {
			return nil, nil, false
		}
		if !atomic.Cas64(&b.w, bw, bw|profReaderSleeping) {
			goto Read
		}
		// Committed to sleeping.
		notetsleepg(&b.wait, -1)
		noteclear(&b.wait)
		goto Read
	}
	data = b.data[profDataCount(br)%uint32(len(b.data)):]
	if uintptr(len(data)) > numData {
		data = data[:numData]
	} else {
		numData -= uintptr(len(data)) // available in case of rewind
	}
	skip := 0
	if data[0] == 0 {
		skip = len(data)
		data = b.data
		if uintptr(len(data)) > numData {
			data = data[:numData]
		}
	}

	ntag := uintptr(profTagCount(bw) - profTagCount(br))
	if ntag == 0 {
		panic("runtime: malformed profBuf buffer - tag and data out of sync")
	}
	tags = b.tags[profTagCount(br)%uint32(len(b.tags)):]
	if uintptr(len(tags)) > ntag {
		tags = tags[:ntag]
	}

	// Count out whole data records until either data or tags is done.
	// They are always in sync in the buffer, but due to an end-of-slice
	// wraparound we might need to stop early and return the rest
	// in the next call.
	di := 0
	ti := 0
	for di < len(data) && data[di] != 0 && ti < len(tags) {
		if uintptr(di)+uintptr(data[di]) > uintptr(len(data)) {
			panic("runtime: malformed profBuf buffer - invalid size")
		}
		di += int(data[di])
		ti++
	}

	// Remember how much we returned, to commit read on next call.
	b.rNext = profAddCounts(br, uint32(skip+di), uint32(ti))

	return data[:di], tags[:ti], false
}

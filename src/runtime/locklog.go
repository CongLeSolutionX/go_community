// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build locklog

// This file implements tracing of lock operations for constructing
// the runtime lock order and checking for cycles. To use this
// facility, build with
//
//  -tags=locklog -gcflags=all=-d=maymorestack=runtime.lockLogMoreStack
//
// The "maymorestack" flag may be omitted if you don't care to track
// potential lock acquisitions caused by stack growths.

package runtime

import (
	"runtime/internal/atomic"
	"unsafe"
)

// lockLogBytes is the number of bytes to buffer on each M.
//
// 64KiB is enough to buffer all lock operations before the os package
// can connect to the log daemon.
const lockLogBytes = 64 << 10

// lockLog stores the globals used by lock logging.
var lockLog struct {
	flushLock    mutex
	fd           int
	flushedBytes uint32
	// labels is a global log of label operations. This is flushed
	// before any per-M log is flushed to ensure that label
	// operations appear before acquire/release operations of the
	// same lock. It is protected by flushLock.
	labels lockLogPerM

	// clsID is the next lock class ID - 1.
	clsID uint32
}

// lockLogInit initializes the lock logging facility to log to file
// descriptor fd and report exePath as the path to this binary.
//
// During initialization, the os package connects to the lock logging
// server socket at GOLOCKLOG, finds this binary's executable path,
// and calls this function. Many locking operations will already have
// happened at this point, so this will panic if the operation buffer
// has already been filled.
//
//go:linkname lockLogInit os.runtime_lockLogInit
func lockLogInit(fd int, exePath string) {
	lock(&lockLog.flushLock)
	lockLog.fd = fd
	if lockLog.flushedBytes != 0 {
		println("runtime: lockLogBytes is too small; already flushed", lockLog.flushedBytes)
		throw("lost lock log operations")
	}

	// Send executable path and wait for acknowledgment.
	var hdr [8*2 + 4]byte
	copy(hdr[:], "locklog\x00")
	lockLogPutUint64(hdr[8:], uint64(funcPC(main)))
	lockLogPutUint32(hdr[8*2:], uint32(len(exePath)))
	lockLogWriteAll(hdr[:])
	lockLogWriteAll(bytes(exePath))
	if n := read(int32(fd), noescape(unsafe.Pointer(&hdr[0])), 1); n < 0 {
		println("runtime: lock log read returned error", -n)
		throw("lock log handshake failed")
	}
	unlock(&lockLog.flushLock)
}

// A lockClass represents the lock class of a non-static lock such as
// a heap-allocated lock.
//
// This is normally used in a statically-initialized global. E.g.,
//
//    var hchanLockClass = &lockClass{name: "runtime.hchan.log"}
type lockClass struct {
	id   uint32
	name string
}

// lockLabeled is like lock(l), but records the lock class and rank
// for a non-static lock acquisition.
func lockLabeled(l *mutex, cls *lockClass, rank uint64) {
	id := atomic.Load(&cls.id)
	if id == 0 {
		// Assign the lock class an ID and record it in the
		// label log.
		lock(&lockLog.flushLock)
		id = atomic.Load(&cls.id)
		if id == 0 {
			id = lockLog.clsID + 1
			lockLog.clsID++
			atomic.Store(&cls.id, id)

			// Log the new class.
			systemstack(func() {
				log := &lockLog.labels
				buf := log.acquireBuf(1 + 4 + len(cls.name))
				buf[0] = byte(lockLogOpNewClass)
				lockLogPutUint32(buf[1:], uint32(len(cls.name)))
				copy(buf[1+4:], cls.name)
				log.releaseBuf()
			})
		}
		unlock(&lockLog.flushLock)
	}
	// Disallow migrating Ms or stack growths between the label
	// message and the lock acquired message.
	systemstack(func() {
		log := &getg().m.lockLog
		buf := log.acquireBuf(1 + 4 + 8)
		buf[0] = byte(lockLogOpLabel)
		lockLogPutUint32(buf[1:], id)
		lockLogPutUint64(buf[1+4:], rank)
		log.releaseBuf()

		// Actually lock the lock.
		log.labeling = true
		lock(l)
		log.labeling = false
	})
}

// The following functions are the entry-points to record lock
// operations.
//
// All of these are nosplit and switch to the system stack immediately
// to avoid stack growths. Since a stack growth could itself have lock
// operations, this prevents re-entrant calls.

//go:nosplit
func lockLogAcquire(l *mutex) {
	if l == &lockLog.flushLock {
		return
	}
	pc := getcallerpc()
	sp := getcallersp()
	gp := getg()
	systemstack(func() {
		// Since acquiring a mutex locks the M as well, this won't
		// race accessing the per-M buffer.
		log := &getg().m.lockLog
		var pcs [8]uintptr
		var skip = 1
		if log.labeling {
			// Skip lockLabeled.
			skip = 4
		}
		n := gentraceback(pc, sp, 0, gp, skip, &pcs[0], len(pcs), nil, nil, _TraceJumpStack)
		addr := uint64(uintptr(unsafe.Pointer(l)))
		log.last = addr
		buf := log.acquireBuf(1 + 8 + 8*n)
		buf[0] = byte(lockLogOpAcquire) | byte(n<<lockLogOpBits)
		lockLogPutUint64(buf[1:], addr)
		for i := 0; i < n; i++ {
			lockLogPutUint64(buf[1+8+8*i:], uint64(pcs[i]))
		}
		log.releaseBuf()
	})
}

//go:nosplit
func lockLogRelease(l *mutex) {
	if l == &lockLog.flushLock {
		return
	}
	systemstack(func() {
		addr := uint64(uintptr(unsafe.Pointer(l)))
		log := &getg().m.lockLog
		if addr == log.last {
			buf := log.acquireBuf(1)
			buf[0] = byte(lockLogOpReleaseLast)
		} else {
			buf := log.acquireBuf(1 + 8)
			buf[0] = byte(lockLogOpRelease)
			lockLogPutUint64(buf[1:], uint64(addr))
		}
		log.releaseBuf()
	})
}

//go:nosplit
func lockLogMayAcquire(l *mutex) {
	pc := getcallerpc()
	sp := getcallersp()
	gp := getg()
	systemstack(func() {
		var pcs [8]uintptr
		n := gentraceback(pc, sp, 0, gp, 0, &pcs[0], len(pcs), nil, nil, _TraceJumpStack)
		addr := uint64(uintptr(unsafe.Pointer(l)))
		log := &getg().m.lockLog
		buf := log.acquireBuf(1 + 8 + 8*n)
		// For now, there's always 1 frame in the traceback.
		buf[0] = byte(lockLogOpMayAcquire) | byte(n<<lockLogOpBits)
		lockLogPutUint64(buf[1:], uint64(addr))
		for i := 0; i < n; i++ {
			lockLogPutUint64(buf[1+8+8*i:], uint64(pcs[i]))
		}
		log.releaseBuf()
	})
}

//go:nosplit
func lockLogFlushAll() {
	systemstack(func() {
		me := acquirem()
		lock(&sched.lock)
		for mp := allm; mp != nil; mp = mp.alllink {
			if mp != me {
				// This is technically racy, since mp could be
				// writing to its log, but this only happens
				// at exit.
				mp.lockLog.flush()
			}
		}
		unlock(&sched.lock)
		// Finally, flush our own so we get the sched.lock unlock.
		me.lockLog.flush()
		releasem(me)
	})
}

// Lock log encoding utilities.

func lockLogPutUint32(buf []byte, x uint32) {
	_ = buf[3]
	buf[0] = byte(x)
	buf[1] = byte(x >> (1 * 8))
	buf[2] = byte(x >> (2 * 8))
	buf[3] = byte(x >> (3 * 8))
}

func lockLogPutUint64(buf []byte, x uint64) {
	_ = buf[7]
	buf[0] = byte(x)
	buf[1] = byte(x >> (1 * 8))
	buf[2] = byte(x >> (2 * 8))
	buf[3] = byte(x >> (3 * 8))
	buf[4] = byte(x >> (4 * 8))
	buf[5] = byte(x >> (5 * 8))
	buf[6] = byte(x >> (6 * 8))
	buf[7] = byte(x >> (7 * 8))
}

// Lock log buffer management functions.
//
// These must be called on the system stack since otherwise logging
// stack growths could cause re-entrant use of the log buffer.

// lockLogPerM is a per-M buffer of recent lock operations.
//
// Only the current M may write into the buffer, but other Ms may
// flush it (specifically, on exit). Hence, writing a record first
// makes a reservation and then only commits that reservation once the
// record is filled in. If that commit fails, we discard the record.
type lockLogPerM struct {
	buf          [lockLogBytes]byte
	last         uint64 // Lock most recently acquired
	pos          uint32 // buf[:pos] is valid. Accessed atomically.
	rStart, rEnd uint32 // Reserved range
	flushing     int

	labeling bool // In lockLabeled
}

type logOp byte

const (
	lockLogOpAcquire logOp = iota
	lockLogOpRelease
	lockLogOpReleaseLast
	lockLogOpMayAcquire

	// lockLogOpNewClass records a new lock class's name. These
	// operations are tracked in a separate log buffer that will
	// be flushed before any acquire/release operations that may
	// reference this class.
	lockLogOpNewClass
	// lockLogLabel records the class ID and rank of the next
	// acquire operation. This is used when acquiring
	// heap-allocated locks.
	lockLogOpLabel

	lockLogOpBits = 3 // Number of bits to store lockLogOp*
)

//go:systemstack
func (lo *lockLogPerM) acquireBuf(bytes int) []byte {
	pos := atomic.Load(&lo.pos)
	if pos+uint32(bytes) >= uint32(len(lo.buf)) {
		lo.flush()
		pos = 0
	}
	lo.rStart = pos
	if pos == 0 {
		// Fresh buffer. Make room for position header and add M ID.
		pos = 4 + 4
		lockLogPutUint32(lo.buf[4:], uint32(getg().m.id))
	}
	buf := lo.buf[pos : pos+uint32(bytes)]
	lo.rEnd = pos + uint32(bytes)
	return buf
}

//go:systemstack
func (lo *lockLogPerM) releaseBuf() {
	// If this cas fails, we raced with an on-exit flush, so we
	// should discard our record anyway.
	atomic.Cas(&lo.pos, lo.rStart, lo.rEnd)
}

//go:systemstack
func (lo *lockLogPerM) flush() {
	lo.flushing++
	if lo.flushing > 1 {
		throw("recursive flush")
	}
	lock(&lockLog.flushLock)
	// Flush the labels log first.
	if lockLog.labels.pos != 0 {
		lockLog.labels.flush1()
	}
	lo.flush1()
	unlock(&lockLog.flushLock)
	lo.flushing--
}

//go:systemstack
func (lo *lockLogPerM) flush1() {
	pos := atomic.Load(&lo.pos)
	lockLog.flushedBytes += pos
	if pos != 0 && lockLog.fd != 0 {
		// Fill the position header.
		lockLogPutUint32(lo.buf[0:], pos)
		// Write out the buffer.
		lockLogWriteAll(lo.buf[:lo.pos])
	}
	atomic.Store(&lo.pos, 0)
}

func lockLogWriteAll(data []byte) {
	for len(data) > 0 {
		// XXX If the pipe is broken, this gets a SIGPIPE,
		// which self-deadlocks on the flush lock.
		n := write(uintptr(lockLog.fd), unsafe.Pointer(&data[0]), int32(len(data)))
		if n < 0 {
			// XXX We can't even print from here because
			// it will cause a recursive flush.
			println("runtime: lock log write returned error", -n)
			throw("lock log flush failed")
		}
		data = data[n:]
	}
}

// lockLogMoreStack records that a conditional morestack call may
// acquire the heap lock. This should be used with the compiler's
// -d=maymorestack=runtime.lockLogMoreStack flag.
//
//go:linkname lockLogMoreStack
//go:nosplit
func lockLogMoreStack() {
	// Only log if we're on a user stack (so it could possibly
	// grow) and if we already hold a lock (otherwise this can't
	// contribute to the lock graph anyway).
	gp := getg()
	if gp == nil || gp.m == nil || gp.m.locks == 0 || gp != gp.m.curg {
		return
	}

	// It's safe to call lockLogMayAcquire directly because it
	// doesn't grow the stack.
	lockLogMayAcquire(&mheap_.lock)
}

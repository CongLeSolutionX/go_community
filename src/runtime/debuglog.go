// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !nodebuglog

// This file provides an internal debug logging facility. The debug
// log is a lightweight, in-memory, per-P ring buffer. By default, the
// runtime prints the debug log on panic.
//
// The debug log is designed so it can be used even from very
// constrained corners of the runtime such as signal handlers or
// inside the write barrier. One known limitation right now is that it
// cannot be used if the caller may not own its current P (e.g.,
// during syscall entry and exit). However, it can be used if there is
// no current P.
//
// The dlog() function is the main entry point to the debug logger.
// For example,
//
//    dlog().string("new G ").int64(gp.goid).end()
//
// dlog() starts a new message and returns a *debugLog object that has
// methods for appending various types to the log message. debugLog
// also has methods for higher-level formatting such as "callers" to
// record the current traceback.
//
// Every dlog() call must be paired with a call to the end() method to
// end the log message.
//
// This facility can be disabled entirely and compiled away by setting
// -tags nodebuglog when building.

package runtime

import (
	"runtime/internal/atomic"
	"unsafe"
)

// debugLogBytes is the size of each per-P ring buffer. This is
// allocated off-heap to avoid blowing up the P and hence the GC'd
// heap size.
const debugLogBytes = 16 << 10

// A debugLogger records debug messages in a low-overhead binary format.
//
// To write to the debug log, start with a call to dlog().
//
// This consists of two debugLog rings in order to allow a signal
// handler to log in the middle of a non-signal log operation.
type debugLogger struct {
	active uint32
	rings  [2]debugLog
}

// A debugLog is a ring buffer of binary debug log records.
//
// A log record consists of a header emitted by dlog(), a sequence of
// fields, and a footer emitted by debugLog.end(). Each field starts
// with a byte indicating its type, followed by type-specific data.
//
// Because this is a ring buffer, new records will eventually
// overwrite old records. Hence, the record format is designed such
// that it can read backwards to find the earliest still-complete
// record, and then read forwards to print the records.
type debugLog struct {
	pid int32

	// committed gives information about the last committed
	// record. committed&1 indicates the index into lastTick and
	// lastNano of the tick and nano at the start of the last
	// committed record. committed>>1 indicates the cursor
	// position just after the last committed record.
	committed int
	write     uint64
	data      []byte

	lastTick, lastNano [2]uint64

	printPos             int
	printTick, printNano uint64
	printFirst           bool
}

// globalDebugLog is a global debugLogger for use when a per-P
// debugLogger isn't available.
//
// TODO: If debugLogger were per-M instead of per-P, I wouldn't need
// this and debugLog wouldn't have issues in contexts where the P
// might be weird (e.g., syscall exit).
var globalDebugLog struct {
	lock mutex
	l    debugLogger
}

//go:nosplit
func (l *debugLog) head() int {
	return l.committed >> 1
}

const (
	debugLogInt = 1 + iota
	debugLogUint
	debugLogHex
	debugLogBoolTrue
	debugLogBoolFalse
	debugLogConstString
	debugLogVarString
	debugLogSpace
	debugLogPC
	debugLogTraceback

	// Type codes 0x80..0xFF indicate a two byte record trailer,
	// where the record payload length is the value - 0x8000. If
	// the length is 0xFFFF, the record is bad.
)

//go:nosplit
func (l *debugLog) reserve(bytes int) int {
	// TODO: We should try to do this fewer times per record.
	// Ideally we could do it just once if we had all of the
	// fields. Alternatively, we could acquire write space in
	// chunks.
	for {
		pos := atomic.Load64(&l.write)
		for pos&(1<<63) != 0 {
			systemstack(func() { usleep(10) })
			pos = atomic.Load64(&l.write)
		}
		if atomic.Cas64(&l.write, pos, pos+uint64(bytes)) {
			return int(pos)
		}
	}
}

//go:nosplit
func (l *debugLog) rawByte(x byte) {
	pos := l.reserve(1)
	l.data[pos%len(l.data)] = x
}

//go:nosplit
func (l *debugLog) rawUint16BE(x uint16) {
	pos := l.reserve(2)
	l.data[pos%len(l.data)] = uint8(x >> 8)
	l.data[(pos+1)%len(l.data)] = uint8(x)
}

//go:nosplit
func (l *debugLog) rawVarint(u uint64) {
	buf := make([]byte, 10)
	i := 0
	for u >= 0x80 {
		buf[i] = byte(u) | 0x80
		u >>= 7
		i++
	}
	buf[i] = byte(u)
	i++
	l.rawBytes(buf[:i])
}

//go:nosplit
func (l *debugLog) rawBytes(x []byte) {
	pos := l.reserve(len(x))
	for len(x) > 0 {
		n := copy(l.data[pos%len(l.data):], x)
		pos += n
		x = x[n:]
	}
}

//go:nosplit
func (l *debugLog) rawString(x string) {
	pos := l.reserve(len(x))
	for len(x) > 0 {
		// The noescape works around issue #15730.
		n := copy(l.data[pos%len(l.data):], *(*string)(noescape(unsafe.Pointer(&x))))
		pos += n
		x = x[n:]
	}
}

//go:nosplit
func (l *debugLog) bool(x bool) *debugLog {
	if l == nil {
		return nil
	}
	if x {
		l.rawByte(debugLogBoolTrue)
	} else {
		l.rawByte(debugLogBoolFalse)
	}
	return l
}

//go:nosplit
func (l *debugLog) int64(x int64) *debugLog {
	if l == nil {
		return nil
	}
	l.rawByte(debugLogInt)
	var u uint64
	if x < 0 {
		u = (^uint64(x) << 1) | 1 // complement i, bit 0 is 1
	} else {
		u = (uint64(x) << 1) // do not complement i, bit 0 is 0
	}
	l.rawVarint(u)
	return l
}

//go:nosplit
func (l *debugLog) int(x int) *debugLog {
	if l == nil {
		return nil
	}
	return l.int64(int64(x))
}

//go:nosplit
func (l *debugLog) uint64(x uint64) *debugLog {
	if l == nil {
		return nil
	}
	l.rawByte(debugLogUint)
	l.rawVarint(x)
	return l
}

//go:nosplit
func (l *debugLog) hex(x uintptr) *debugLog {
	if l == nil {
		return nil
	}
	// TODO: Hex is often used for pointers. Can we save space by
	// saving the lowest heap address we've seen and recording
	// these relative to that?
	l.rawByte(debugLogHex)
	l.rawVarint(uint64(x))
	return l
}

//go:nosplit
func (l *debugLog) pc(pc uintptr) *debugLog {
	if l == nil {
		return nil
	}
	l.rawByte(debugLogPC)
	l.rawVarint(uint64(pc))
	return l
}

//go:nosplit
func (l *debugLog) traceback(pcs []uintptr) *debugLog {
	if l == nil {
		return nil
	}
	l.rawByte(debugLogTraceback)
	l.rawVarint(uint64(len(pcs)))
	for _, pc := range pcs {
		l.rawVarint(uint64(pc))
	}
	return l
}

//go:nosplit
func (l *debugLog) callers(skip, limit int) *debugLog {
	if l == nil {
		return nil
	}
	var stk [16]uintptr
	if limit < 0 || limit > len(stk) {
		limit = len(stk)
	}

	// This is equivalent to callers(), but with _TraceJumpStack.
	sp := getcallersp()
	pc := getcallerpc()
	gp := getg()
	var nstk int
	systemstack(func() {
		nstk = gentraceback(pc, sp, 0, gp, skip, &stk[0], limit, nil, nil, _TraceJumpStack)
	})

	return l.traceback(stk[:nstk])
}

//go:nosplit
func (l *debugLog) string(x string) *debugLog {
	if l == nil {
		return nil
	}
	if x == " " {
		return l.sp()
	}
	str := stringStructOf(&x)
	datap := &firstmoduledata
	// TODO: If it's a really short constant string, just emit the string.
	if datap.etext <= uintptr(str.str) && uintptr(str.str) < datap.end {
		// String constants are in the rodata section, which
		// isn't recorded in moduledata. But it has to be
		// somewhere between etext and end.
		l.rawByte(debugLogConstString)
		l.rawVarint(uint64(str.len))
		l.rawVarint(uint64(uintptr(str.str) - datap.etext))
	} else {
		l.rawByte(debugLogVarString)
		l.rawVarint(uint64(len(x)))
		l.rawString(x)
	}
	return l
}

//go:nosplit
func (l *debugLog) sp() *debugLog {
	if l == nil {
		return nil
	}
	l.rawByte(debugLogSpace)
	return l
}

// dlog starts a new debug log entry and returns the debug log to
// write the entry to. dlog() must always be paired with a call to
// debugLog.end().
//
// For example, to log the start of a new goroutine:
//     dlog().string("new G ").int64(gp.goid).end()
//
//go:nosplit
func dlog() *debugLog {
	var l *debugLogger
	var pid int32 = -1
	gp := getg()
	if gp == nil || gp.m == nil {
		// Use global ring.
		systemstack(func() { lock(&globalDebugLog.lock) })
		l = &globalDebugLog.l
	} else {
		mp := acquirem()
		if mp.p == 0 {
			// Use global ring.
			releasem(mp)
			systemstack(func() { lock(&globalDebugLog.lock) })
			l = &globalDebugLog.l
		} else {
			// Use per-P ring.
			l = &mp.p.ptr().debugLogger
			pid = mp.p.ptr().id
		}
	}
	active := int(atomic.Xadd(&l.active, 1)) - 1
	if active >= len(l.rings) {
		throw("recursive dlog or missing debugLog.end()")
	}
	lr := &l.rings[active]

	// Initialize debugLog if necessary.
	if lr.data == nil {
		lr.pid = pid
		// Allocate log memory.
		systemstack(func() {
			buf := persistentalloc(debugLogBytes, 0, &memstats.other_sys)
			// TODO: I should be able to use a notinheap
			// [debugLogBytes]byte type here, but the slice
			// assignment still has a write barrier.
			sh := (*slice)(unsafe.Pointer(&lr.data))
			*(*uintptr)(unsafe.Pointer(&sh.array)) = uintptr(buf)
			sh.len, sh.cap = debugLogBytes, debugLogBytes
		})
	}

	// Emit header.
	index := lr.committed & 1
	tick, nano := uint64(cputicks()), uint64(nanotime())
	lr.rawVarint(tick - lr.lastTick[index])
	lr.rawVarint(nano - lr.lastNano[index])
	lr.lastTick[1-index], lr.lastNano[1-index] = tick, nano
	return lr
}

//go:nosplit
func (l *debugLog) end() {
	if l == nil {
		return
	}
	// Emit footer.
	len := l.write - uint64(l.head())
	if 0x8000+len >= 0xFFFF {
		// Unrepresentable length (plus we wrapped around the
		// whole ring, so it doesn't matter).
		l.rawUint16BE(0xFFFF)
	} else {
		l.rawUint16BE(uint16(0x8000 + len))
	}
	index := l.committed & 1
	// Commit this record and flip the active lastTick/Nano
	l.committed = int(l.write)<<1 | (1 - index)

	var logger *debugLogger
	if l.pid == -1 {
		logger = &globalDebugLog.l
	} else {
		logger = &allp[l.pid].debugLogger
	}

	active := atomic.Xadd(&logger.active, -1)
	if &logger.rings[active] != l {
		throw("mismatched debugLog.end()")
	}

	if l.pid == -1 {
		// No M or no P. We're using the global ring.
		systemstack(func() { unlock(&globalDebugLog.lock) })
	} else {
		releasem(getg().m)
	}
}

// printDebugLog prints the debug log.
func printDebugLog() {
	getVarint := func(l *debugLog) uint64 {
		var u uint64
		for i := uint(0); ; i += 7 {
			if l.printPos >= l.head() {
				throw("truncated record")
			}
			b := l.data[l.printPos%len(l.data)]
			l.printPos++
			u |= uint64(b&^0x80) << i
			if b&0x80 == 0 {
				break
			}
		}
		return u
	}

	printlock()

	// Find the beginning position, tick, and nano of each debug log.
	prepareLog := func(logger *debugLogger) {
		for ringi := range logger.rings {
			l := &logger.rings[ringi]

			// Lock this log. printDebugLog is the only
			// thing that locks logs, so we're only racing
			// with increments of the write position, not
			// with other lockers.
			writePos := l.write
			for {
				if atomic.Cas64(&l.write, writePos, writePos|(1<<63)) {
					break
				}
				l.write = atomic.Load64(&l.write)
			}

			// We may be in the middle of a log write. Figure out
			// where we've overwritten to and don't rewind further
			// than that.
			tail := int(writePos) - len(l.data)
			if tail < 0 {
				tail = 0
			}

			// Walk records backwards until we get to the
			// beginning or to a partial record. Work out
			// the timestamp of the first complete record
			// by walking back from the timestamp of the
			// most recent complete record.
			pos := l.head()
			index := l.committed & 1
			tick, nano := l.lastTick[index], l.lastNano[index]
			var tickDelta, nanoDelta uint64
			for {
				npos := pos - 2
				if npos < tail {
					break
				}
				rlen := uint16(l.data[npos%len(l.data)]) << 8
				rlen |= uint16(l.data[(npos+1)%len(l.data)])
				if rlen == 0xFFFF {
					break
				}
				rlen -= 0x8000
				npos -= int(rlen)
				if npos < tail {
					break
				}
				// Successfully consumed a record.
				tick -= tickDelta
				nano -= nanoDelta
				pos = npos
				// Get the tick and nano delta between
				// this record and the previous
				// record.
				l.printPos = pos
				tickDelta, nanoDelta = getVarint(l), getVarint(l)
			}
			l.printPos = pos
			l.printFirst = true
			l.printTick = tick
			l.printNano = nano
		}
	}
	for _, p := range allp {
		prepareLog(&p.debugLogger)
	}
	prepareLog(&globalDebugLog.l)

	// Find and print next record until we run out.
	for {
		var best struct {
			time  uint64
			log   *debugLog
			ringi int
		}
		best.time = ^uint64(0)
		for _, p := range allp {
			for ringi := range p.debugLogger.rings {
				l := &p.debugLogger.rings[ringi]
				if l.printPos == l.head() {
					continue
				}

				if l.printTick < best.time {
					best.time, best.log, best.ringi = l.printTick, l, ringi
				}
			}
		}
		for ringi := range globalDebugLog.l.rings {
			l := &globalDebugLog.l.rings[ringi]
			if l.printPos == l.head() {
				continue
			}

			if l.printTick < best.time {
				best.time, best.log, best.ringi = l.printTick, l, ringi
			}
		}
		if best.log == nil {
			break
		}

		// Print record from bestLog.
		l := best.log
		if l.printFirst {
			print(">> begin P ", l.pid, " ring ", best.ringi, " log")
			if l.printPos != 0 {
				print("; lost first ", l.printPos>>10, " KB")
			}
			print(" <<\n")
			l.printFirst = false
			// The rest of the code expects the cursor to
			// always be just past the deltas in the
			// header since it needs to know the time of
			// the next record. Skip the first headers
			// here.
			getVarint(l)
			getVarint(l)
		}

		print("[")
		var tmpbuf [21]byte
		pnano := int64(l.printNano) - runtimeInitTime
		if pnano < 0 {
			// Logged before runtimeInitTime was set.
			pnano = 0
		}
		// TODO: I disabled recordForPanic in gwrite because
		// it was causing tmpbuf to escape because of a
		// "recursive call".
		gwrite(itoaDiv(tmpbuf[:], uint64(pnano), 9))
		print(" P ", l.pid, "] ")
	fields:
		for {
			typ := l.data[l.printPos%len(l.data)]
			l.printPos++
			if typ >= 0x80 {
				// End of record trailer.
				l.printPos++
				if l.printPos > l.head() {
					throw("bad trailer")
				}
				break fields
			}

			switch typ {
			default:
				print("<unknown field type ", hex(typ), " pos ", l.printPos, " head ", l.head(), " write ", l.write&^(1<<63), "> aborting P log\n")
				l.printPos = l.head()
				//throw("unknown field type")
				break fields

			case debugLogInt:
				u := getVarint(l)
				var v int64
				if u&1 == 0 {
					v = int64(u >> 1)
				} else {
					v = ^int64(u >> 1)
				}
				print(v)

			case debugLogUint:
				u := getVarint(l)
				print(u)

			case debugLogHex:
				u := getVarint(l)
				print(hex(u))

			case debugLogBoolTrue:
				print(true)

			case debugLogBoolFalse:
				print(false)

			case debugLogConstString:
				len, ptr := int(getVarint(l)), uintptr(getVarint(l))
				ptr += firstmoduledata.etext
				str := stringStruct{
					str: unsafe.Pointer(ptr),
					len: len,
				}
				s := *(*string)(unsafe.Pointer(&str))
				print(s)

			case debugLogVarString:
				slen := int(getVarint(l))
				s := l.data[l.printPos%len(l.data):]
				slen1 := len(s)
				if slen1 > slen {
					slen1 = slen
				}
				gwrite(s[:slen1])
				l.printPos += slen1
				slen -= slen1
				if slen != 0 {
					// String wrapped around the ring.
					gwrite(l.data[:slen])
					l.printPos += slen
				}

			case debugLogSpace:
				printsp()

			case debugLogPC:
				pc := uintptr(getVarint(l))
				printDebugLogPC(pc)

			case debugLogTraceback:
				nstk := int(getVarint(l))
				for i := 0; i < nstk; i++ {
					pc := uintptr(getVarint(l))
					print("\n    ")
					printDebugLogPC(pc)
				}
			}
		}
		println()

		// Consume deltas to the next record, update the next
		// record time so we can sort it in, and place the
		// cursor just after the header.
		if l.printPos != l.head() {
			l.printTick += getVarint(l)
			l.printNano += getVarint(l)
		}
	}

	// Unlock logs.
	for _, p := range allp {
		for ringi := range p.debugLogger.rings {
			l := &p.debugLogger.rings[ringi]
			atomic.Store64(&l.write, l.write&^(1<<63))
		}
	}

	printunlock()
}

func printDebugLogPC(pc uintptr) {
	print(hex(pc))
	fn := findfunc(pc)
	if !fn.valid() {
		print(" [unknown PC]")
	} else {
		name := funcname(fn)
		file, line := funcline(fn, pc)
		print(" [", name, "+", hex(pc-fn.entry),
			" ", file, ":", line, "]")
	}
}

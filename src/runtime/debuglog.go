// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	"runtime/internal/atomic"
	"runtime/internal/sys"
	"unsafe"
)

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
	logger *debugLogger
	head   int
	write  uint64
	data   [16 << 10]byte

	tickBase uint64

	printPos   int
	printTick  uint64
	printFirst bool
}

const (
	debugLogInt = 1 + iota
	debugLogHex
	debugLogConstString
	debugLogVarString
	debugLogSpace

	// Type codes 0x80..0xFF indicate a two byte record trailer,
	// where the record payload length is the value - 0x8000. If
	// the length is 0xFFFF, the record is bad.
)

func (l *debugLog) reserve(bytes int) int {
	// TODO: We should try to do this fewer times per record.
	// Ideally we could do it just once if we had all of the
	// fields. Alternatively, we could acquire write space in
	// chunks.
	for {
		pos := l.write
		for pos&(1<<63) != 0 {
			usleep(10)
			pos = atomic.Load64(&l.write)
		}
		if atomic.Cas64(&l.write, pos, pos+uint64(bytes)) {
			return int(pos)
		}
	}
}

func (l *debugLog) rawByte(x byte) {
	pos := l.reserve(1)
	l.data[pos%len(l.data)] = x
}

func (l *debugLog) rawUint16BE(x uint16) {
	pos := l.reserve(2)
	l.data[pos%len(l.data)] = uint8(x >> 8)
	l.data[(pos+1)%len(l.data)] = uint8(x)
}

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

func (l *debugLog) rawBytes(x []byte) {
	pos := l.reserve(len(x))
	for len(x) > 0 {
		n := copy(l.data[pos%len(l.data):], x)
		pos += n
		x = x[n:]
	}
}

func (l *debugLog) rawString(x string) {
	pos := l.reserve(len(x))
	for len(x) > 0 {
		// The noescape works around issue #15730.
		n := copy(l.data[pos%len(l.data):], *(*string)(noescape(unsafe.Pointer(&x))))
		pos += n
		x = x[n:]
	}
}

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

func (l *debugLog) int(x int) *debugLog {
	if l == nil {
		return nil
	}
	return l.int64(int64(x))
}

func (l *debugLog) hex(x uintptr) *debugLog {
	if l == nil {
		return nil
	}
	l.rawByte(debugLogHex)
	l.rawVarint(uint64(x))
	return l
}

func (l *debugLog) string(x string) *debugLog {
	if l == nil {
		return nil
	}
	str := stringStructOf(&x)
	datap := &firstmoduledata
	// TODO: If it's a really short constant string, just emit the string.
	if datap.noptrdata <= uintptr(str.str) && uintptr(str.str) < datap.enoptrdata {
		l.rawByte(debugLogConstString)
		str := stringStructOf(&x)
		l.rawVarint(uint64(str.len))
		l.rawVarint(uint64(uintptr(str.str) - datap.noptrdata))
	} else {
		l.rawByte(debugLogVarString)
		l.rawVarint(uint64(len(x)))
		l.rawString(x)
	}
	return l
}

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
//     dlog().string("new G ").int64(gp.goid0.end()
func dlog() *debugLog {
	gp := getg()
	if gp == nil || gp.m == nil {
		// TODO: Have a global ring for this.
		return nil
	}
	mp := acquirem()
	if mp.p == 0 {
		// TODO: Have a global ring for this.
		releasem(mp)
		return nil
	}
	l := &mp.p.ptr().debugLogger
	active := int(atomic.Xadd(&l.active, 1)) - 1
	if active >= len(l.rings) {
		throw("recursive dlog or missing debugLog.end()")
	}
	lr := &l.rings[active]
	if lr.tickBase == 0 {
		lr.tickBase = uint64(cputicks())
	}
	// Emit header.
	lr.rawVarint(uint64(cputicks()) - lr.tickBase)
	lr.rawVarint(uint64(nanotime() - runtimeInitTime))
	return lr
}

func (l *debugLog) end() {
	if l == nil {
		return
	}
	// Emit footer.
	len := l.write - uint64(l.head)
	if 0x8000+len >= 0xFFFF {
		// Unrepresentable length (plus we wrapped around the
		// whole ring, so it doesn't matter).
		l.rawUint16BE(0xFFFF)
	} else {
		l.rawUint16BE(uint16(0x8000 + len))
	}
	l.head = int(l.write)

	mp := getg().m
	logger := &mp.p.ptr().debugLogger
	active := atomic.Xadd(&logger.active, -1)
	if &logger.rings[active] != l {
		throw("mismatched debugLog.end()")
	}
	releasem(mp)
}

// printDebugLog prints the debug log.
func printDebugLog() {
	getVarint := func(l *debugLog) uint64 {
		var u uint64
		for i := uint(0); ; i += 7 {
			if l.printPos >= l.head {
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

	// Find the beginning of each debug log.
	for _, p := range allp {
		if p == nil {
			continue
		}
		for ringi := range p.debugLogger.rings {
			l := &p.debugLogger.rings[ringi]

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
			// beginning or to a partial record.
			pos := l.head
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
				pos = npos
			}
			l.printPos = pos
			l.printFirst = true

			// Decode initial record time.
			if l.printPos != l.head {
				l.printTick = getVarint(l) + l.tickBase
			}
		}
	}

	// Find and print next record until we run out.
	for {
		var best struct {
			time  uint64
			p     *p
			ringi int
		}
		best.time = ^uint64(0)
		for _, p := range allp {
			if p == nil {
				continue
			}
			for ringi := range p.debugLogger.rings {
				l := &p.debugLogger.rings[ringi]
				if l.printPos == l.head {
					continue
				}

				if l.printTick < best.time {
					best.time, best.p, best.ringi = l.printTick, p, ringi
				}
			}
		}
		if best.p == nil {
			break
		}

		// Print record from bestLog.
		l := &best.p.debugLogger.rings[best.ringi]
		if l.printFirst {
			print(">> begin P ", best.p.id, " ring ", best.ringi, " log <<\n")
			l.printFirst = false
		}

		print("[")
		var tmpbuf [20]byte
		t0 := getVarint(l)
		gwrite(itoaDiv(tmpbuf[:], t0, 9))
		print(" P ", best.p.id, "] ")
	fields:
		for {
			typ := l.data[l.printPos%len(l.data)]
			l.printPos++
			if typ >= 0x80 {
				// End of record trailer.
				l.printPos++
				if l.printPos > l.head {
					throw("bad trailer")
				}
				break fields
			}

			switch typ {
			default:
				print("<unknown field type", hex(typ), "pos", l.printPos, "head", l.head, "write", l.write&^(1<<63), "> aborting P log")
				l.printPos = l.head
				//throw("unknown field type")
				break fields

			case debugLogInt:
				u := getVarint(l)
				var v int64
				if u&1 == 0 {
					v = int64(u >> 1)
				} else {
					v = -int64(u >> 1)
				}
				print(v)

			case debugLogHex:
				u := getVarint(l)
				print(hex(u))

			case debugLogConstString:
				len, ptr := int(getVarint(l)), uintptr(getVarint(l))
				ptr += firstmoduledata.noptrdata
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
			}
		}
		println()

		// Get timestamp of next record.
		if l.printPos != l.head {
			l.printTick = getVarint(l) + l.tickBase
		}
	}

	// Unlock logs.
	for _, p := range allp {
		if p == nil {
			continue
		}
		for ringi := range p.debugLogger.rings {
			l := &p.debugLogger.rings[ringi]
			atomic.Store64(&l.write, l.write&^(1<<63))
		}
	}

	printunlock()
}

func hexdumpWords(p, end uintptr, mark uintptr) {
	p1 := func(x uintptr) {
		var buf [2 * sys.PtrSize]byte
		for i := len(buf) - 1; i >= 0; i-- {
			if x&0xF < 10 {
				buf[i] = byte(x&0xF) + '0'
			} else {
				buf[i] = byte(x&0xF) - 10 + 'a'
			}
			x >>= 4
		}
		gwrite(buf[:])
	}

	printlock()
	for i := uintptr(0); p+i < end; i += sys.PtrSize {
		if i%16 == 0 {
			if i != 0 {
				println()
			}
			p1(p + i)
			print(":")
		}

		print(" ")
		val := *(*uintptr)(unsafe.Pointer(p + i))
		p1(val)

		if p+i <= mark && mark < p+i+sys.PtrSize {
			print("*")
		} else {
			print(" ")
		}
	}
	println()
	printunlock()
}

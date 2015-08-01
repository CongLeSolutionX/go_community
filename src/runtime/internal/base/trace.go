// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Go execution tracer.
// The tracer captures a wide range of execution events like goroutine
// creation/blocking/unblocking, syscall enter/exit/block, GC-related events,
// changes of heap size, processor start/stop, etc and writes them to a buffer
// in a compact form. A precise nanosecond-precision timestamp and a stack
// trace is captured for most events.
// See https://golang.org/s/go15trace for more info.

package base

import (
	"unsafe"
)

// Event types in the trace, args are given in square brackets.
const (
	TraceEvNone           = 0  // unused
	TraceEvBatch          = 1  // start of per-P batch of events [pid, timestamp]
	TraceEvFrequency      = 2  // contains tracer timer frequency [frequency (ticks per second)]
	TraceEvStack          = 3  // stack [stack id, number of PCs, array of PCs]
	TraceEvGomaxprocs     = 4  // current value of GOMAXPROCS [timestamp, GOMAXPROCS, stack id]
	TraceEvProcStart      = 5  // start of P [timestamp, thread id]
	TraceEvProcStop       = 6  // stop of P [timestamp]
	TraceEvGCStart        = 7  // GC start [timestamp, stack id]
	TraceEvGCDone         = 8  // GC done [timestamp]
	TraceEvGCScanStart    = 9  // GC scan start [timestamp]
	TraceEvGCScanDone     = 10 // GC scan done [timestamp]
	TraceEvGCSweepStart   = 11 // GC sweep start [timestamp, stack id]
	TraceEvGCSweepDone    = 12 // GC sweep done [timestamp]
	TraceEvGoCreate       = 13 // goroutine creation [timestamp, new goroutine id, start PC, stack id]
	TraceEvGoStart        = 14 // goroutine starts running [timestamp, goroutine id]
	TraceEvGoEnd          = 15 // goroutine ends [timestamp]
	TraceEvGoStop         = 16 // goroutine stops (like in select{}) [timestamp, stack]
	TraceEvGoSched        = 17 // goroutine calls Gosched [timestamp, stack]
	TraceEvGoPreempt      = 18 // goroutine is preempted [timestamp, stack]
	TraceEvGoSleep        = 19 // goroutine calls Sleep [timestamp, stack]
	TraceEvGoBlock        = 20 // goroutine blocks [timestamp, stack]
	TraceEvGoUnblock      = 21 // goroutine is unblocked [timestamp, goroutine id, stack]
	TraceEvGoBlockSend    = 22 // goroutine blocks on chan send [timestamp, stack]
	TraceEvGoBlockRecv    = 23 // goroutine blocks on chan recv [timestamp, stack]
	TraceEvGoBlockSelect  = 24 // goroutine blocks on select [timestamp, stack]
	TraceEvGoBlockSync    = 25 // goroutine blocks on Mutex/RWMutex [timestamp, stack]
	TraceEvGoBlockCond    = 26 // goroutine blocks on Cond [timestamp, stack]
	TraceEvGoBlockNet     = 27 // goroutine blocks on network [timestamp, stack]
	TraceEvGoSysCall      = 28 // syscall enter [timestamp, stack]
	TraceEvGoSysExit      = 29 // syscall exit [timestamp, goroutine id, real timestamp]
	TraceEvGoSysBlock     = 30 // syscall blocks [timestamp]
	TraceEvGoWaiting      = 31 // denotes that goroutine is blocked when tracing starts [goroutine id]
	TraceEvGoInSyscall    = 32 // denotes that goroutine is in syscall when tracing starts [goroutine id]
	TraceEvHeapAlloc      = 33 // memstats.heap_live change [timestamp, heap_alloc]
	TraceEvNextGC         = 34 // memstats.next_gc change [timestamp, next_gc]
	TraceEvTimerGoroutine = 35 // denotes timer goroutine [timer goroutine id]
	TraceEvFutileWakeup   = 36 // denotes that the previous wakeup of this goroutine was futile [timestamp]
	TraceEvCount          = 37
)

const (
	// Timestamps in trace are cputicks/traceTickDiv.
	// This makes absolute values of timestamp diffs smaller,
	// and so they are encoded in less number of bytes.
	// 64 on x86 is somewhat arbitrary (one tick is ~20ns on a 3GHz machine).
	// The suggested increment frequency for PowerPC's time base register is
	// 512 MHz according to Power ISA v2.07 section 6.2, so we use 16 on ppc64
	// and ppc64le.
	// Tracing won't work reliably for architectures where cputicks is emulated
	// by nanotime, so the value doesn't matter for those architectures.
	TraceTickDiv = 16 + 48*(goarch_386|goarch_amd64|goarch_amd64p32)
	// Maximum number of PCs in a single stack trace.
	// Since events contain only stack id rather than whole stack trace,
	// we can allow quite large values here.
	TraceStackSize = 128
	// Identifier of a fake P that is used when we trace without a real P.
	TraceGlobProc = -1
	// Maximum number of bytes to encode uint64 in base-128.
	TraceBytesPerNumber = 10
	// Shift of the number of arguments in the first event byte.
	TraceArgCountShift = 6
	// Flag passed to traceGoPark to denote that the previous wakeup of this
	// goroutine was futile. For example, a goroutine was unblocked on a mutex,
	// but another goroutine got ahead and acquired the mutex before the first
	// goroutine is scheduled, so the first goroutine has to block again.
	// Such wakeups happen on buffered channels and sync.Mutex,
	// but are generally not interesting for end user.
	TraceFutileWakeup byte = 128
)

// trace is global tracing context.
var Trace struct {
	Lock          Mutex     // protects the following members
	LockOwner     *G        // to avoid deadlocks during recursive lock locks
	Enabled       bool      // when set runtime traces events
	Shutdown      bool      // set when we are waiting for trace reader to finish after setting enabled to false
	HeaderWritten bool      // whether ReadTrace has emitted trace header
	FooterWritten bool      // whether ReadTrace has emitted trace footer
	ShutdownSema  uint32    // used to wait for ReadTrace completion
	TicksStart    int64     // cputicks when tracing was started
	TicksEnd      int64     // cputicks when tracing was stopped
	TimeStart     int64     // nanotime when tracing was started
	TimeEnd       int64     // nanotime when tracing was stopped
	Reading       *TraceBuf // buffer currently handed off to user
	Empty         *TraceBuf // stack of empty buffers
	FullHead      *TraceBuf // queue of full buffers
	FullTail      *TraceBuf
	Reader        *G              // goroutine that called ReadTrace, or nil
	StackTab      traceStackTable // maps stack traces to unique ids

	BufLock Mutex     // protects buf
	Buf     *TraceBuf // global trace buffer, used when running without a p
}

// traceBufHeader is per-P tracing buffer.
type TraceBufHeader struct {
	Link      *TraceBuf               // in trace.empty/full
	lastTicks uint64                  // when we wrote the last event
	Buf       []byte                  // trace data, always points to traceBuf.arr
	stk       [TraceStackSize]uintptr // scratch buffer for traceback
}

// traceBuf is per-P tracing buffer.
type TraceBuf struct {
	TraceBufHeader
	arr [64<<10 - unsafe.Sizeof(TraceBufHeader{})]byte // underlying buffer for traceBufHeader.buf
}

// traceReader returns the trace reader that should be woken up, if any.
func traceReader() *G {
	if Trace.Reader == nil || (Trace.FullHead == nil && !Trace.Shutdown) {
		return nil
	}
	Lock(&Trace.Lock)
	if Trace.Reader == nil || (Trace.FullHead == nil && !Trace.Shutdown) {
		Unlock(&Trace.Lock)
		return nil
	}
	gp := Trace.Reader
	Trace.Reader = nil
	Unlock(&Trace.Lock)
	return gp
}

// traceFullQueue queues buf into queue of full buffers.
func TraceFullQueue(buf *TraceBuf) {
	buf.Link = nil
	if Trace.FullHead == nil {
		Trace.FullHead = buf
	} else {
		Trace.FullTail.Link = buf
	}
	Trace.FullTail = buf
}

// traceEvent writes a single event to trace buffer, flushing the buffer if necessary.
// ev is event type.
// If skip > 0, write current stack id as the last argument (skipping skip top frames).
// If skip = 0, this event type should contain a stack, but we don't want
// to collect and remember it for this particular call.
func TraceEvent(ev byte, skip int, args ...uint64) {
	mp, pid, bufp := traceAcquireBuffer()
	// Double-check trace.enabled now that we've done m.locks++ and acquired bufLock.
	// This protects from races between traceEvent and StartTrace/StopTrace.

	// The caller checked that trace.enabled == true, but trace.enabled might have been
	// turned off between the check and now. Check again. traceLockBuffer did mp.locks++,
	// StopTrace does stopTheWorld, and stopTheWorld waits for mp.locks to go back to zero,
	// so if we see trace.enabled == true now, we know it's true for the rest of the function.
	// Exitsyscall can run even during stopTheWorld. The race with StartTrace/StopTrace
	// during tracing in exitsyscall is resolved by locking trace.bufLock in traceLockBuffer.
	if !Trace.Enabled && !mp.Startingtrace {
		traceReleaseBuffer(pid)
		return
	}
	buf := *bufp
	const maxSize = 2 + 4*TraceBytesPerNumber // event type, length, timestamp, stack id and two add params
	if buf == nil || cap(buf.Buf)-len(buf.Buf) < maxSize {
		buf = traceFlush(buf)
		*bufp = buf
	}

	ticks := uint64(Cputicks()) / TraceTickDiv
	tickDiff := ticks - buf.lastTicks
	if len(buf.Buf) == 0 {
		data := buf.Buf
		data = append(data, TraceEvBatch|1<<TraceArgCountShift)
		data = TraceAppend(data, uint64(pid))
		data = TraceAppend(data, ticks)
		buf.Buf = data
		tickDiff = 0
	}
	buf.lastTicks = ticks
	narg := byte(len(args))
	if skip >= 0 {
		narg++
	}
	// We have only 2 bits for number of arguments.
	// If number is >= 3, then the event type is followed by event length in bytes.
	if narg > 3 {
		narg = 3
	}
	data := buf.Buf
	data = append(data, ev|narg<<TraceArgCountShift)
	var lenp *byte
	if narg == 3 {
		// Reserve the byte for length assuming that length < 128.
		data = append(data, 0)
		lenp = &data[len(data)-1]
	}
	data = TraceAppend(data, tickDiff)
	for _, a := range args {
		data = TraceAppend(data, a)
	}
	if skip == 0 {
		data = append(data, 0)
	} else if skip > 0 {
		_g_ := Getg()
		gp := mp.Curg
		var nstk int
		if gp == _g_ {
			nstk = Callers(skip, buf.stk[:])
		} else if gp != nil {
			gp = mp.Curg
			nstk = Gcallers(gp, skip, buf.stk[:])
		}
		if nstk > 0 {
			nstk-- // skip runtime.goexit
		}
		if nstk > 0 && gp.Goid == 1 {
			nstk-- // skip runtime.main
		}
		id := Trace.StackTab.put(buf.stk[:nstk])
		data = TraceAppend(data, uint64(id))
	}
	evSize := len(data) - len(buf.Buf)
	if evSize > maxSize {
		Throw("invalid length of trace event")
	}
	if lenp != nil {
		// Fill in actual length.
		*lenp = byte(evSize - 2)
	}
	buf.Buf = data
	traceReleaseBuffer(pid)
}

// traceAcquireBuffer returns trace buffer to use and, if necessary, locks it.
func traceAcquireBuffer() (mp *M, pid int32, bufp **TraceBuf) {
	mp = Acquirem()
	if p := mp.P.Ptr(); p != nil {
		return mp, p.Id, &p.Tracebuf
	}
	Lock(&Trace.BufLock)
	return mp, TraceGlobProc, &Trace.Buf
}

// traceReleaseBuffer releases a buffer previously acquired with traceAcquireBuffer.
func traceReleaseBuffer(pid int32) {
	if pid == TraceGlobProc {
		Unlock(&Trace.BufLock)
	}
	Releasem(Getg().M)
}

// traceFlush puts buf onto stack of full buffers and returns an empty buffer.
func traceFlush(buf *TraceBuf) *TraceBuf {
	owner := Trace.LockOwner
	dolock := owner == nil || owner != Getg().M.Curg
	if dolock {
		Lock(&Trace.Lock)
	}
	if buf != nil {
		if &buf.Buf[0] != &buf.arr[0] {
			Throw("trace buffer overflow")
		}
		TraceFullQueue(buf)
	}
	if Trace.Empty != nil {
		buf = Trace.Empty
		Trace.Empty = buf.Link
	} else {
		buf = (*TraceBuf)(SysAlloc(unsafe.Sizeof(TraceBuf{}), &Memstats.Other_sys))
		if buf == nil {
			Throw("trace: out of memory")
		}
	}
	buf.Link = nil
	buf.Buf = buf.arr[:0]
	buf.lastTicks = 0
	if dolock {
		Unlock(&Trace.Lock)
	}
	return buf
}

// traceAppend appends v to buf in little-endian-base-128 encoding.
func TraceAppend(buf []byte, v uint64) []byte {
	for ; v >= 0x80; v >>= 7 {
		buf = append(buf, 0x80|byte(v))
	}
	buf = append(buf, byte(v))
	return buf
}

// traceStackTable maps stack traces (arrays of PC's) to unique uint32 ids.
// It is lock-free for reading.
type traceStackTable struct {
	lock Mutex
	seq  uint32
	mem  traceAlloc
	tab  [1 << 13]*traceStack
}

// traceStack is a single stack in traceStackTable.
type traceStack struct {
	link *traceStack
	hash uintptr
	id   uint32
	n    int
	stk  [0]uintptr // real type [n]uintptr
}

// stack returns slice of PCs.
func (ts *traceStack) stack() []uintptr {
	return (*[TraceStackSize]uintptr)(unsafe.Pointer(&ts.stk))[:ts.n]
}

// put returns a unique id for the stack trace pcs and caches it in the table,
// if it sees the trace for the first time.
func (tab *traceStackTable) put(pcs []uintptr) uint32 {
	if len(pcs) == 0 {
		return 0
	}
	hash := Memhash(unsafe.Pointer(&pcs[0]), uintptr(len(pcs))*unsafe.Sizeof(pcs[0]), 0)
	// First, search the hashtable w/o the mutex.
	if id := tab.find(pcs, hash); id != 0 {
		return id
	}
	// Now, double check under the mutex.
	Lock(&tab.lock)
	if id := tab.find(pcs, hash); id != 0 {
		Unlock(&tab.lock)
		return id
	}
	// Create new record.
	tab.seq++
	stk := tab.newStack(len(pcs))
	stk.hash = hash
	stk.id = tab.seq
	stk.n = len(pcs)
	stkpc := stk.stack()
	for i, pc := range pcs {
		stkpc[i] = pc
	}
	part := int(hash % uintptr(len(tab.tab)))
	stk.link = tab.tab[part]
	Atomicstorep(unsafe.Pointer(&tab.tab[part]), unsafe.Pointer(stk))
	Unlock(&tab.lock)
	return stk.id
}

// find checks if the stack trace pcs is already present in the table.
func (tab *traceStackTable) find(pcs []uintptr, hash uintptr) uint32 {
	part := int(hash % uintptr(len(tab.tab)))
Search:
	for stk := tab.tab[part]; stk != nil; stk = stk.link {
		if stk.hash == hash && stk.n == len(pcs) {
			for i, stkpc := range stk.stack() {
				if stkpc != pcs[i] {
					continue Search
				}
			}
			return stk.id
		}
	}
	return 0
}

// newStack allocates a new stack of size n.
func (tab *traceStackTable) newStack(n int) *traceStack {
	return (*traceStack)(tab.mem.alloc(unsafe.Sizeof(traceStack{}) + uintptr(n)*PtrSize))
}

// dump writes all previously cached stacks to trace buffers,
// releases all memory and resets state.
func (tab *traceStackTable) Dump() {
	var tmp [(2 + TraceStackSize) * TraceBytesPerNumber]byte
	buf := traceFlush(nil)
	for _, stk := range tab.tab {
		for ; stk != nil; stk = stk.link {
			maxSize := 1 + (3+stk.n)*TraceBytesPerNumber
			if cap(buf.Buf)-len(buf.Buf) < maxSize {
				buf = traceFlush(buf)
			}
			// Form the event in the temp buffer, we need to know the actual length.
			tmpbuf := tmp[:0]
			tmpbuf = TraceAppend(tmpbuf, uint64(stk.id))
			tmpbuf = TraceAppend(tmpbuf, uint64(stk.n))
			for _, pc := range stk.stack() {
				tmpbuf = TraceAppend(tmpbuf, uint64(pc))
			}
			// Now copy to the buffer.
			data := buf.Buf
			data = append(data, TraceEvStack|3<<TraceArgCountShift)
			data = TraceAppend(data, uint64(len(tmpbuf)))
			data = append(data, tmpbuf...)
			buf.Buf = data
		}
	}

	Lock(&Trace.Lock)
	TraceFullQueue(buf)
	Unlock(&Trace.Lock)

	tab.mem.drop()
	*tab = traceStackTable{}
}

// traceAlloc is a non-thread-safe region allocator.
// It holds a linked list of traceAllocBlock.
type traceAlloc struct {
	head *traceAllocBlock
	off  uintptr
}

// traceAllocBlock is a block in traceAlloc.
type traceAllocBlock struct {
	next *traceAllocBlock
	data [64<<10 - PtrSize]byte
}

// alloc allocates n-byte block.
func (a *traceAlloc) alloc(n uintptr) unsafe.Pointer {
	n = Round(n, PtrSize)
	if a.head == nil || a.off+n > uintptr(len(a.head.data)) {
		if n > uintptr(len(a.head.data)) {
			Throw("trace: alloc too large")
		}
		block := (*traceAllocBlock)(SysAlloc(unsafe.Sizeof(traceAllocBlock{}), &Memstats.Other_sys))
		if block == nil {
			Throw("trace: out of memory")
		}
		block.next = a.head
		a.head = block
		a.off = 0
	}
	p := &a.head.data[a.off]
	a.off += n
	return unsafe.Pointer(p)
}

// drop frees all previously allocated memory and resets the allocator.
func (a *traceAlloc) drop() {
	for a.head != nil {
		block := a.head
		a.head = block.next
		SysFree(unsafe.Pointer(block), unsafe.Sizeof(traceAllocBlock{}), &Memstats.Other_sys)
	}
}

func TraceProcStart() {
	TraceEvent(TraceEvProcStart, -1, uint64(Getg().M.Id))
}

func TraceProcStop(pp *P) {
	// Sysmon and stopTheWorld can stop Ps blocked in syscalls,
	// to handle this we temporary employ the P.
	mp := Acquirem()
	oldp := mp.P
	mp.P.Set(pp)
	TraceEvent(TraceEvProcStop, -1)
	mp.P = oldp
	Releasem(mp)
}

func TraceGCScanStart() {
	TraceEvent(TraceEvGCScanStart, -1)
}

func TraceGCScanDone() {
	TraceEvent(TraceEvGCScanDone, -1)
}

func TraceGoStart() {
	TraceEvent(TraceEvGoStart, -1, uint64(Getg().M.Curg.Goid))
}

func traceGoPark(traceEv byte, skip int, gp *G) {
	if traceEv&TraceFutileWakeup != 0 {
		TraceEvent(TraceEvFutileWakeup, -1)
	}
	TraceEvent(traceEv & ^TraceFutileWakeup, skip)
}

func TraceGoUnpark(gp *G, skip int) {
	TraceEvent(TraceEvGoUnblock, skip, uint64(gp.Goid))
}

func TraceGoSysCall() {
	TraceEvent(TraceEvGoSysCall, 4)
}

func traceGoSysExit(ts int64) {
	if ts != 0 && ts < Trace.TicksStart {
		// The timestamp was obtained during a previous tracing session, ignore.
		return
	}
	TraceEvent(TraceEvGoSysExit, -1, uint64(Getg().M.Curg.Goid), uint64(ts)/TraceTickDiv)
}

func TraceGoSysBlock(pp *P) {
	// Sysmon and stopTheWorld can declare syscalls running on remote Ps as blocked,
	// to handle this we temporary employ the P.
	mp := Acquirem()
	oldp := mp.P
	mp.P.Set(pp)
	TraceEvent(TraceEvGoSysBlock, -1)
	mp.P = oldp
	Releasem(mp)
}

// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	"internal/abi"
	"internal/goarch"
	"internal/goexperiment"
	"runtime/internal/atomic"
	"runtime/internal/sys"
	"unsafe"
)

var publishStats struct {
	mallocBytes atomic.Uint64
	globalBytes atomic.Uint64
	// localBytes if mallocBytes - globalBytes
}

func publishStatsReport() {
	mallocBytes := publishStats.mallocBytes.Load()
	globalBytes := publishStats.globalBytes.Load()
	localBytes := mallocBytes - globalBytes
	print("local: ", localBytes, "/", mallocBytes, "\n")
}

func init() {
	// Printing of course breaks lots of tests. Maybe append to a log somewhere?
	if false && goexperiment.CgoCheck2 {
		addExitHook(publishStatsReport, false)
	}
}

const debugLeak = true

func dlogLeak() *dlogger {
	if !debugLeak {
		return nil
	}
	return dlog().s("leak").u(uint(myOwnerID()))
}

type ownerID uint16

func myOwnerID() ownerID {
	gp := getg().m.curg
	if gp == nil {
		return 0 // Global owner
	}
	return ownerIDOf(gp)
}

func ownerIDOf(gp *g) ownerID {
	myId := ownerID(gp.goid)
	if myId == 0 {
		myId = 0xFFFF
	}
	return myId
}

func ownerMalloc(obj unsafe.Pointer, span *mspan, isTiny bool, typ *abi.Type, callPC uintptr) {
	if !goexperiment.CgoCheck2 {
		return
	}
	objIndex := span.objIndex(uintptr(obj))
	publishStats.mallocBytes.Add(int64(span.elemsize))
	if isTiny {
		// This is kind of unfortunate, but at least tiny allocs will never
		// cause anything else to be published.
		span.owners[objIndex] = 0
		publishStats.globalBytes.Add(int64(span.elemsize))
	} else {
		span.owners[objIndex] = myOwnerID()
	}
	if debugLeak {
		var pcs [10]uintptr
		n := callers(3, pcs[:])
		//dlogLeak().s("malloc").p(obj).pc(callPC).end()
		dlogLeak().s("malloc").p(obj).s(toRType(typ).string()).s("tiny").b(isTiny).traceback(pcs[:n]).end()
	}
}

type publishedType uint8

const (
	publishedNone publishedType = iota // Zero value, so you can use >0 checks.
	publishedGlobal
	publishedHeap
)

//go:nosplit
func isPublished(ptr unsafe.Pointer) publishedType {
	// The stack may move, so we need to keep ptr as a pointer.
	span := spanOf(uintptr(ptr))
	if span == nil {
		// Assume this is a global. Globals are implicitly published. (Do we
		// need to check this?)
		return publishedGlobal
	}
	if state := span.state.get(); state != mSpanInUse {
		if state != mSpanManual {
			throw("write into dead span")
		}
		// This is a stack.
		gp := getg().m.curg
		if gp != nil && gp.stack.lo < uintptr(ptr) && uintptr(ptr) <= gp.stack.hi {
			// This is our stack.
			//
			// TODO: Track this for up-stack escapes.
			return publishedNone
		}
		g0 := getg().m.g0
		if g0.stack.lo < uintptr(ptr) && uintptr(ptr) <= g0.stack.hi {
			// This can happen if we're running on g0 and do an up-stack write,
			// or sendTime, which sends on a channel while on g0.
			return publishedNone
		}
		// This is someone else's stack.
		g0, gsig := getg().m.g0, getg().m.gsignal
		dlogLeak().s("write to another stack").p(ptr).
			s("g").hex(uint64(gp.stack.lo)).hex(uint64(gp.stack.hi)).
			s("g0").hex(uint64(g0.stack.lo)).hex(uint64(g0.stack.hi)).
			s("gsignal").hex(uint64(gsig.stack.lo)).hex(uint64(gsig.stack.hi)).end()
		throw("write to another stack")
	}

	objIndex := span.objIndex(uintptr(ptr))
	owner := span.owners[objIndex]
	if owner == 0 {
		return publishedHeap
	} else if owner != myOwnerID() {
		// We somehow reached an object that's local to another goroutine.
		dlogLeak().s("observed remote object").p(ptr).end()
		throw("object leaked without us noticing")
	}
	// This G owns the destination object, so we're not leaking src.
	return publishedNone
}

//go:nosplit
func isPublishable(ptr unsafe.Pointer) bool {
	// This is very similar to isPublished, but is meant for the target of a
	// write into a published object.
	span := spanOfHeap(uintptr(ptr))
	if span == nil {
		// Globals are already published and stack objects can't be published,
		// so there's nothing to do.
		return false
	}

	objIndex := span.objIndex(uintptr(ptr))
	owner := span.owners[objIndex]
	if owner == 0 {
		// It's already published.
		return false
	} else if owner != myOwnerID() {
		// We somehow reached an object that's local to another goroutine.
		dlogLeak().s("observed remote object").p(ptr).end()
		throw("object leaked without us noticing")
	}
	// It's local and in the heap. We can publish it.
	return true
}

func findGlobalBitmap(ptr uintptr) (bitmap *uint8, bitOffset uintptr) {
	for _, datap := range activeModules() {
		if datap.data <= ptr && ptr <= datap.edata {
			doff := (ptr - datap.data) / goarch.PtrSize
			return datap.gcdatamask.bytedata, doff
		}
		if datap.bss <= ptr && ptr <= datap.ebss {
			boff := (ptr - datap.bss) / goarch.PtrSize
			return datap.gcbssmask.bytedata, boff
		}
	}
	return nil, 0
}

func classifyPtr(ptr uintptr) string {
	if ptr == 0 {
		return "nil"
	}
	span := spanOf(ptr)
	if span == nil {
		gp := getg().m.curg
		if gp != nil && gp.stack.lo < uintptr(ptr) && uintptr(ptr) <= gp.stack.hi {
			return "g stack"
		}
		gp = getg().m.g0
		if gp.stack.lo < uintptr(ptr) && uintptr(ptr) <= gp.stack.hi {
			return "g0 stack"
		}
		gp = getg().m.gsignal
		if gp != nil && gp.stack.lo < uintptr(ptr) && uintptr(ptr) <= gp.stack.hi {
			return "gsignal stack"
		}
		for _, datap := range activeModules() {
			if datap.text <= ptr && ptr < datap.etext {
				return "text"
			} else if datap.noptrdata <= ptr && ptr < datap.enoptrdata {
				return "noptrdata"
			} else if datap.data <= ptr && ptr < datap.edata {
				return "data"
			} else if datap.bss <= ptr && ptr < datap.ebss {
				return "bss"
			} else if datap.noptrbss <= ptr && ptr < datap.enoptrbss {
				return "noptrbss"
			} else if datap.types <= ptr && ptr < datap.etypes {
				return "types"
			}
		}
		return "nil span"
	}
	switch span.state.get() {
	case mSpanInUse:
		return "heap"
	case mSpanManual:
		return "stack"
	}
	return "unknown span"
}

// publishPtrWrite implements the barrier before *dst = src. It publishes src if dst is published.
//
// This happens between a write barrier check and a write, so it must not allow preemption.
//
//go:nosplit
func publishPtrWrite(dst *unsafe.Pointer, src unsafe.Pointer) {
	if !goexperiment.CgoCheck2 {
		return
	}
	if isPublished(unsafe.Pointer(dst)) > 0 && isPublishable(src) {
		mp := acquirem()
		// dst is a slot in a global object. Are we leaking src? We can only
		// leak src if it points to a heap object, so there's no complexity
		// around where to get the pointer bitmap from in this case.
		srcBase, srcSpan, srcIndex := findObject(uintptr(src), 0, 0)
		if srcBase == 0 {
			// src isn't in the heap (this is probably bad because it indicates
			// a heap->stack pointer).
			releasem(mp)
			return
		}
		var pr publishRoot
		pr.init().doObject(srcBase, srcSpan, srcIndex, unsafe.Pointer(dst))
		pr.flush()
		releasem(mp)
	}
}

// publishMemmove implements the barrier before memmove(dst, src, size). If dst
// is published, it publishes everything reachable from src.
//
// This happens between a write barrier check and a write, so it must not allow preemption.
//
//go:nosplit
func publishMemmove(dst, src unsafe.Pointer, size uintptr) {
	// This gets called *before* the memmove, which means we have to get the
	// data from src. However, src may be on the stack; hence, we use dst's
	// pointer bitmap, which we can always get.

	if !goexperiment.CgoCheck2 {
		return
	}
	dpub := isPublished(dst)
	// TODO: Check if src is already published and, if so, do nothing because
	// everything src points to must already be published as well. However,
	// there are weird edge cases to this check: src may be in another stack
	// (for some channel ops), in which case we need to conservatively publish.
	if dpub > 0 {
		// src may be on the stack. This means we have to use dst's bitmap and
		// can't let the stack move. (TODO: There can only be one level of
		// src-on-stackness; would it suffice to use unsafe.Pointer below and
		// allow stack moves? We'd still have to disable preemption.)
		systemstack(func() {
			var pr publishRoot
			switch dpub {
			case publishedHeap:
				//dlogLeak().s("memmove").p(dst).uptr(size).end() // Noisy and not very useful
				hbits := heapBitsForAddr(uintptr(dst), size)
				pr.init().doRange(hbits, uintptr(src))
				pr.flush()
			case publishedGlobal:
				//dlogLeak().s("memmove").p(dst).uptr(size).end()
				bitmap, bitOffset := findGlobalBitmap(uintptr(dst))
				pr.init().doRangeBitmap(src, size, bitmap, bitOffset)
				pr.flush()
			}
		})
	}
}

// publishObject unconditionally publishes the object at ptr. It's called in a
// handful of one-off places where the runtime is doing odd things.
func publishObject(ptr unsafe.Pointer, reason string) {
	if !goexperiment.CgoCheck2 {
		return
	}
	span := spanOfHeap(uintptr(ptr))
	if span == nil {
		if debugLeak {
			var pcs [10]uintptr
			n := callers(2, pcs[:])
			dlogLeak().s("publishObject on non-heap ptr").p(ptr).traceback(pcs[:n]).end()
		}
		return
	}
	var pr publishRoot
	pr.init().doObject(uintptr(ptr), span, span.objIndex(uintptr(ptr)), nil)
	pr.flush()
}

// publishChan is called in various places where channels are doing copies off
// other goroutine's stacks. It requires that base be either published or owned
// by owner (which may not be the calling goroutine), and uses the pointer
// bitmap from typ, since base may be on a stack.
func publishChan(owner *g, base unsafe.Pointer, typ *abi.Type) {
	if !goexperiment.CgoCheck2 {
		return
	}
	// base may point to a stack. Channels are the one case where we copy stack
	// to stack. If we're doing a direct receive, the current owner is actually
	// the goroutine we're receiving from.
	if typ.PtrBytes == 0 {
		return
	}
	if typ.Kind_&kindGCProg != 0 {
		throw("type " + toRType(typ).string() + " has GC prog")
	}
	// base may point to the stack, but we keep it as an unsafe.Pointer so it
	// can be adjusted.
	var pr publishRoot
	pr.initAt(owner).doRangeBitmap(base, typ.PtrBytes, typ.GCData, 0)
	pr.flush()
}

type publishRoot struct {
	id    ownerID
	stack *publishStack
}

func (pr *publishRoot) init() publisher {
	pr.id = myOwnerID()
	return publisher{pr, publisherLimit, 0}
}

func (pr *publishRoot) initAt(gp *g) publisher {
	pr.id = ownerIDOf(gp)
	return publisher{pr, publisherLimit, 0}
}

func (pr *publishRoot) push(ptr uintptr) {
	if pr.stack == nil || pr.stack.len == len(pr.stack.ptrs) {
		// Allocate a new stack segment.
		lock(&publishStackLock)
		stack := (*publishStack)(publishStackAlloc.alloc())
		unlock(&publishStackLock)
		pr.stack, stack.prev = stack, pr.stack
		stack.len = 0
	}
	pr.stack.ptrs[pr.stack.len] = ptr
	pr.stack.len++
}

func (pr *publishRoot) flush() {
	for pr.stack != nil {
		stack := pr.stack
		stack.len--
		ptr := stack.ptrs[stack.len]
		if stack.len == 0 {
			pr.stack = stack.prev
			lock(&publishStackLock)
			publishStackAlloc.free(unsafe.Pointer(stack))
			unlock(&publishStackLock)
		}

		span := spanOf(ptr)
		index := span.objIndex(ptr)
		publisher{pr, publisherLimit, 0}.doObject(ptr, span, index, nil)
	}
}

// publisherLimit is the maximum recursion depth while publishing. This needs to
// be small enough to fit the call stack on the g0 stack. If we exceed this,
// we'll push the next level onto the publisherStack, but for small publish
// operations, this avoids any allocation of pointer buffers.
const publisherLimit = 16

type publisher struct {
	*publishRoot
	limit  int
	indent int
}

func (p publisher) down() publisher {
	p.limit--
	p.indent += 4
	return p
}

func (p publisher) dlog() *dlogger {
	return dlogLeak().indent(p.indent)
}

var publishStackLock mutex
var publishStackAlloc fixalloc

type publishStack struct {
	_    sys.NotInHeap
	prev *publishStack
	len  int
	ptrs [4096/goarch.PtrSize - 2]uintptr
}

func (p publisher) doObject(base uintptr, span *mspan, objIndex uintptr, cause unsafe.Pointer) {
	owner := span.owners[objIndex]
	if owner == 0 {
		// Object is already global.
		return
	}
	if p.limit <= 0 {
		// We're out of code stack. Push this object to revisit from the top loop.
		p.dlog().s("push").hex(uint64(base)).end()
		p.push(base)
		return
	}
	p.dlog().s("publish").hex(uint64(base)).s("<-").p(cause).end()
	if owner != p.id {
		// We reached an object that's neither global nor owned by us.
		var pcs [10]uintptr
		n := callers(2, pcs[:])
		p.dlog().s("publishing object").hex(uint64(base)).s("owned by").u(uint(owner)).traceback(pcs[:n]).end()
		throw("object leaked without us noticing")
	}
	// Publish the object.
	publishStats.globalBytes.Add(int64(span.elemsize))
	span.owners[objIndex] = 0

	// Publish any objects reachable from the object.
	hbits := heapBitsForAddr(base, span.elemsize)
	p.down().doRange(hbits, base)
}

func (p publisher) doRange(hbits heapBits, base uintptr) {
	offset := base - hbits.addr
	for {
		var addr uintptr
		if hbits, addr = hbits.nextFast(); addr == 0 {
			if hbits, addr = hbits.next(); addr == 0 {
				break
			}
		}
		// For memmove, hbits may point at the destination (where we can get
		// hbits), while base points to the source. Adjust addr to point at
		// base.
		addr += offset

		obj := *(*uintptr)(unsafe.Pointer(addr))
		if obj == 0 {
			continue
		}

		// Find the heap object, if any, obj points to. This may also point to a
		// global, but globals are always implicitly published, so we don't have
		// to traverse into globals. In some runtime-internal cases, it can
		// point to the stack; we assume those cases never need to cause
		// publication.
		if tBase, tSpan, tObjIndex := findObject(obj, base, addr-base); tBase != 0 {
			p.doObject(tBase, tSpan, tObjIndex, unsafe.Pointer(addr))
		} else if debugLeak {
			c := classifyPtr(obj)
			p.dlog().hex(uint64(obj)).s("<-").hex(uint64(addr)).s("type:").s(c).end()
		}
	}
}

func (p publisher) doRangeBitmap(base unsafe.Pointer, size uintptr, bitmap *uint8, bitOffset uintptr) {
	bitmap = addb(bitmap, bitOffset/8)
	bitOffset = bitOffset % 8
	for i := uintptr(0); i < size; i += goarch.PtrSize {
		if (*bitmap>>uint8(bitOffset))&1 != 0 {
			slot := add(base, i)
			obj := *(*uintptr)(slot)
			if obj != 0 {
				if tBase, tSpan, tObjIndex := findObject(obj, uintptr(base), i); tBase != 0 {
					p.doObject(tBase, tSpan, tObjIndex, slot)
				}
			}
		}

		bitOffset++
		if bitOffset == 8 {
			bitmap = addb(bitmap, 1)
			bitOffset = 0
		}
	}
}

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Garbage collector: stack objects and stack tracing

package runtime

import (
	"runtime/internal/sys"
	"unsafe"
)

const stackTraceDebug = false

// Buffer for pointers found during stack tracing.
// Must be smaller than or equal to workbuf.
//
//go:notinheap
type stackWorkBuf struct {
	stackWorkBufHdr
	obj [(_WorkbufSize - unsafe.Sizeof(stackWorkBufHdr{})) / sys.PtrSize]uintptr
}

// Header declaration must come after the buf declaration above, because of issue #14620.
//
//go:notinheap
type stackWorkBufHdr struct {
	workbufhdr
	next *stackWorkBuf // linked list of workbufs
	// Note: we could theoretically repurpose lfnode.next as this next pointer.
	// It would save 1 word, but that probably isn't worth busting open
	// the lfnode API.
}

// Buffer for stack objects found on a goroutine stack.
// Must be smaller than or equal to workbuf.
//
//go:notinheap
type stackObjectBuf struct {
	stackObjectBufHdr
	obj [(_WorkbufSize - unsafe.Sizeof(stackObjectBufHdr{})) / unsafe.Sizeof(stackObject{})]stackObject
}

//go:notinheap
type stackObjectBufHdr struct {
	workbufhdr
	next *stackObjectBuf
}

func init() {
	if unsafe.Sizeof(stackWorkBuf{}) > unsafe.Sizeof(workbuf{}) {
		panic("stackWorkBuf too big")
	}
	if unsafe.Sizeof(stackObjectBuf{}) > unsafe.Sizeof(workbuf{}) {
		panic("stackObjectBuf too big")
	}
}

// A stackObject represents a variable on the stack that has had
// its address taken.
//
//go:notinheap
type stackObject struct {
	off   uint32       // offset above stack.lo
	size  uint32       // size of object
	typ   *_type       // type info (for ptr/nonptr bits). nil if object has been scanned.
	left  *stackObject // objects with lower addresses
	right *stackObject // objects with higher addresses
}

// obj.typ = typ, but with no write barrier.
//go:nowritebarrier
func (obj *stackObject) setType(typ *_type) {
	// Types of stack objects are always in read-only memory, not the heap.
	// So not using a write barrier is ok.
	*(*uintptr)(unsafe.Pointer(&obj.typ)) = uintptr(unsafe.Pointer(typ))
}

// A stackScanState keeps track of the state used during the GC walk
// of a goroutine.
//
//go:notinheap
type stackScanState struct {
	cache pcvalueCache

	// stack limits
	stack stack

	// buf contains the set of possible pointers to stack objects.
	// Organized as a LIFO linked list of buffers.
	// All buffers except possibly the head buffer are full.
	buf     *stackWorkBuf
	freeBuf *stackWorkBuf // keep around one free buffer for allocation hysteresis

	// list of stack objects
	// Objects are in increasing address order.
	head  *stackObjectBuf
	tail  *stackObjectBuf
	nobjs int

	// root of binary tree for fast object lookup by address
	root *stackObject
}

// Add p as a potential pointer to a stack object.
// p must be a stack address.
func (s *stackScanState) putPtr(p uintptr) {
	if p < s.stack.lo || p >= s.stack.hi {
		throw("address not a stack address")
	}
	buf := s.buf
	if buf == nil {
		// Initial setup.
		buf = (*stackWorkBuf)(unsafe.Pointer(getempty()))
		buf.nobj = 0
		buf.next = nil
		s.buf = buf
	} else if buf.nobj == len(buf.obj) {
		if s.freeBuf != nil {
			buf = s.freeBuf
			s.freeBuf = nil
		} else {
			buf = (*stackWorkBuf)(unsafe.Pointer(getempty()))
		}
		buf.nobj = 0
		buf.next = s.buf
		s.buf = buf
	}
	buf.obj[buf.nobj] = p
	buf.nobj++
}

// Remove and return a potential pointer to a stack object.
// Returns 0 if there are no more pointers available.
func (s *stackScanState) getPtr() uintptr {
	buf := s.buf
	if buf == nil {
		// Never had any data.
		return 0
	}
	if buf.nobj == 0 {
		if s.freeBuf != nil {
			// Free old freeBuf.
			putempty((*workbuf)(unsafe.Pointer(s.freeBuf)))
		}
		// Move buf to the freeBuf.
		s.freeBuf = buf
		buf = buf.next
		s.buf = buf
		if buf == nil {
			// No more data.
			putempty((*workbuf)(unsafe.Pointer(s.freeBuf)))
			s.freeBuf = nil
			return 0
		}
	}
	buf.nobj--
	return buf.obj[buf.nobj]
}

// addObject adds a stack object at addr of type typ to the set of stack objects.
func (s *stackScanState) addObject(addr uintptr, typ *_type) {
	x := s.tail
	if x == nil {
		// initial setup
		x = (*stackObjectBuf)(unsafe.Pointer(getempty()))
		x.next = nil
		s.head = x
		s.tail = x
	}
	if x.nobj > 0 && uint32(addr-s.stack.lo) < x.obj[x.nobj-1].off+x.obj[x.nobj-1].size {
		throw("objects added out of order or overlapping")
	}
	if x.nobj == len(x.obj) {
		// full buffer - allocate a new buffer, add to end of linked list
		y := (*stackObjectBuf)(unsafe.Pointer(getempty()))
		y.next = nil
		x.next = y
		s.tail = y
		x = y
	}
	obj := &x.obj[x.nobj]
	x.nobj++
	obj.off = uint32(addr - s.stack.lo)
	obj.size = uint32(typ.size)
	obj.setType(typ)
	// obj.left and obj.right will be initalized by buildIndex before use.
	s.nobjs++
}

// buildIndex initializes s.root to a binary search tree.
// It should be called after all addObject calls but before
// any call of findObject.
func (s *stackScanState) buildIndex() {
	s.root, _, _ = binarySearchTree(s.head, 0, s.nobjs)
}

// Build a binary search tree with the n objects in the list
// x.obj[idx], x.obj[idx+1], ..., x.next.obj[0], ...
// Returns the root of that tree, and the buf+idx of the nth object after x.obj[idx].
// (The first object that was not included in the binary search tree.)
// If n == 0, returns nil, x.
func binarySearchTree(x *stackObjectBuf, idx int, n int) (root *stackObject, restBuf *stackObjectBuf, restIdx int) {
	if n == 0 {
		return nil, x, idx
	}
	var left, right *stackObject
	left, x, idx = binarySearchTree(x, idx, n/2)
	root = &x.obj[idx]
	idx++
	if idx == len(x.obj) {
		x = x.next
		idx = 0
	}
	right, x, idx = binarySearchTree(x, idx, n-n/2-1)
	root.left = left
	root.right = right
	return root, x, idx
}

// findObject returns the stack object containing address a, if any.
// Must have called buildIndex previously.
func (s *stackScanState) findObject(a uintptr) *stackObject {
	off := uint32(a - s.stack.lo)
	obj := s.root
	for {
		if obj == nil {
			return nil
		}
		if off < obj.off {
			obj = obj.left
			continue
		}
		if off >= obj.off+obj.size {
			obj = obj.right
			continue
		}
		return obj
	}
}

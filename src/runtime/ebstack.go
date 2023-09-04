package runtime

import (
	"internal/cpu"
	"internal/goarch"
	"runtime/internal/atomic"
	"unsafe"
)

// elimination-backoff stack algorithm, to reduce cas contention
// see https://people.csail.mit.edu/shanir/publications/Lock_Free.pdf
type ebstack struct {
	head      uint64
	_         cpu.CacheLinePad
	exchanger [ebStackArrayLen]struct {
		e uint64
		_ [cpu.CacheLinePadSize - goarch.PtrSize]byte
	}
}

const (
	eliminationBackoffEmpty   uint64 = 0
	eliminationBackoffWaiting uint64 = 1
	eliminationBackoffBusy    uint64 = 2

	ebStackPinTimeout int64 = 10000

	ebStackArrayLen = 32
)

func (ebs *ebstack) push(node *lfnode) {
	node.pushcnt++
	new := lfstackPack(node, node.pushcnt)
	if node1 := lfstackUnpack(new); node1 != node {
		print("runtime: lfstack.push invalid packing: node=", node, " cnt=", hex(node.pushcnt), " packed=", hex(new), " -> node=", node1, "\n")
		throw("lfstack.push")
	}
	for {
		old := atomic.Load64((*uint64)(&ebs.head))
		node.next = old
		if atomic.Cas64((*uint64)(&ebs.head), old, new) {
			return
		}
		// try backoff
		y, success := ebs.exchange(node)
		if !success {
			continue
		}
		if y == nil {
			return
		}
	}
}

func (ebs *ebstack) pop() unsafe.Pointer {
	for {
		old := atomic.Load64((*uint64)(&ebs.head))
		if old == 0 {
			return nil
		}
		node := lfstackUnpack(old)
		next := atomic.Load64(&node.next)
		if atomic.Cas64((*uint64)(&ebs.head), old, next) {
			return unsafe.Pointer(node)
		}
		y, success := ebs.exchange(nil)
		if !success {
			continue
		}
		if y != nil {
			return unsafe.Pointer(y)
		}
	}
}

func (ebs *ebstack) empty() bool {
	return atomic.Load64(&ebs.head) == 0
}

func (ebs *ebstack) exchange(node *lfnode) (*lfnode, bool) {
	i := fastrandn(ebStackArrayLen)

	deadline := nanotime() + ebStackPinTimeout

	var old, status uint64
	var oldNode *lfnode

	for nanotime() < deadline {
		old = atomic.Load64(&ebs.exchanger[i].e)
		oldNode, status = ebstackUnpack(old)

		if status == eliminationBackoffEmpty {
			new := ebstackPack(node, uintptr(eliminationBackoffWaiting))
			if atomic.Cas64(&ebs.exchanger[i].e, old, new) {
				// wait for the other thread to exchange
				for nanotime() < deadline {
					old = atomic.Load64(&ebs.exchanger[i].e)
					oldNode, status = ebstackUnpack(old)
					if status != eliminationBackoffBusy {
						procyield(10)
						continue
					}
					// exchange successfully, reset to EMPTY
					atomic.Store64(&ebs.exchanger[i].e, 0)
					return oldNode, true
				}
				// timeout, try reset state to EMPTY
				if atomic.Cas64(&ebs.exchanger[i].e, new, 0) {
					return nil, false
				}
				// fails, some exchang-ing thread must have shown up,
				// exchange completes
				old = atomic.Load64(&ebs.exchanger[i].e)
				oldNode, _ = ebstackUnpack(old)
				atomic.Store64(&ebs.exchanger[i].e, 0)
				return oldNode, true
			}
		} else if status == eliminationBackoffWaiting {
			new := ebstackPack(node, 2)
			if atomic.Cas64(&ebs.exchanger[i].e, old, new) {
				return oldNode, true
			}
		}
		procyield(10)
	}

	return nil, false
}

func ebstackPack(node *lfnode, cnt uintptr) uint64 {
	return uint64(taggedPointerPack(unsafe.Pointer(node), cnt))
}

func ebstackUnpack(val uint64) (*lfnode, uint64) {
	tp := taggedPointer(val)
	return (*lfnode)(tp.pointer()), uint64(tp.tag())
}

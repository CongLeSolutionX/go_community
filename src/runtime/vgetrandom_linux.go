// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux && amd64

package runtime

import "unsafe"

var vgrnd struct {
	states     []uintptr
	statesLock mutex
	stateSize  uintptr
	mmapProt   uint32
	mmapFlags  uint32
}

func vgrndInit() {
	if vdsoGetrandomSym == 0 {
		return
	}

	var params struct {
		SizeOfOpaqueParams uint32
		MmapProt           uint32
		MmapFlags          uint32
		reserved           [13]uint32
	}
	if vgetrandom(nil, 0, 0, uintptr(unsafe.Pointer(&params)), ^uint(0)) < 0 {
		return
	}
	vgrnd.stateSize = uintptr(params.SizeOfOpaqueParams)
	vgrnd.mmapProt = params.MmapProt
	vgrnd.mmapFlags = params.MmapFlags

	lockInit(&vgrnd.statesLock, lockRankLeafRank)
}

func vgrndGetState() uintptr {
	lock(&vgrnd.statesLock)
	if len(vgrnd.states) == 0 {
		num := uintptr(ncpu)
		allocSize := (num*vgrnd.stateSize + physPageSize - 1) & (^(physPageSize - 1))
		num = (physPageSize / vgrnd.stateSize) * (allocSize / physPageSize)
		p, err := mmap(nil, allocSize, int32(vgrnd.mmapProt), int32(vgrnd.mmapFlags), -1, 0)
		if err != 0 {
			unlock(&vgrnd.statesLock)
			return 0
		}
		newBlock := uintptr(p)
		if vgrnd.states == nil {
			vgrnd.states = make([]uintptr, 0, num)
		}
		for i := uintptr(0); i < num; i++ {
			if (newBlock&(physPageSize-1))+vgrnd.stateSize > physPageSize {
				newBlock = (newBlock + physPageSize - 1) & (^(physPageSize - 1))
			}
			vgrnd.states = append(vgrnd.states, newBlock)
			newBlock += vgrnd.stateSize
		}
	}
	state := vgrnd.states[len(vgrnd.states)-1]
	vgrnd.states = vgrnd.states[:len(vgrnd.states)-1]
	unlock(&vgrnd.statesLock)
	return state
}

func vgrndPutState(state uintptr) {
	lock(&vgrnd.statesLock)
	vgrnd.states = append(vgrnd.states, state)
	unlock(&vgrnd.statesLock)
}

func vgetrandom(buf *byte, length uint, flags uint32, opaquestate uintptr, opaquestateSize uint) (ret int)

//go:linkname getrandomVDSO
//go:nosplit
func getrandomVDSO(p []byte, flags uint32) (ret int, supported bool) {
	if vgrnd.stateSize == 0 {
		return -1, false
	}

	mp := getg().m
	//TODO: take some kind of lock so that another g doesn't take our m. i
	//added go:nosplit, hoping that'd do it, but i think that's not right
	//and this needs to be revisited.

	if mp.vdsoGetRandomState == 0 {
		state := vgrndGetState()
		if state == 0 {
			return -1, false
		}
		mp.vdsoGetRandomState = state
	}

	supported = true

	var buf *byte
	if len(p) > 0 {
		buf = &p[0]
	}
	return vgetrandom(buf, uint(len(p)), flags, mp.vdsoGetRandomState, uint(vgrnd.stateSize)), true
}

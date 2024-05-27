// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	"internal/abi"
	"internal/goarch"
	"unsafe"
)

type shadowStack struct {
	pcs []uintptr
}

func ShadowFPCallers(pcs []uintptr) (i int) {
	const shadowDepth = 2
	var (
		gp           = getg()
		fp           = unsafe.Pointer(getfp())
		oldShadowPtr unsafe.Pointer
		newShadowPtr unsafe.Pointer
	)
	for i = 0; i < len(pcs) && fp != nil; i++ {
		pcPtr := unsafe.Pointer(uintptr(fp) + goarch.PtrSize)
		pc := *(*uintptr)(pcPtr)
		if i == shadowDepth {
			newShadowPtr = pcPtr
		}
		if pc == abi.FuncPCABIInternal(shadowTrampolineASM) {
			i += copy(pcs[i:], gp.shadowStack.pcs)
			oldShadowPtr = pcPtr
			break
		}
		pcs[i] = pc
		fp = unsafe.Pointer(*(*uintptr)(fp))
	}
	if oldShadowPtr == newShadowPtr {
		return
	}
	if oldShadowPtr != nil {
		*(*uintptr)(oldShadowPtr) = gp.shadowStack.pcs[0]
	}
	gp.shadowStack.pcs = append(gp.shadowStack.pcs[0:0], pcs[shadowDepth:i]...)
	*(*uintptr)(newShadowPtr) = abi.FuncPCABIInternal(shadowTrampolineASM)
	return
}

func shadowTrampolineASM()

func shadowTrampolineGo() (retpc uintptr) {
	gp := getg()
	retpc = gp.shadowStack.pcs[0]
	gp.shadowStack.pcs = gp.shadowStack.pcs[1:]
	*(*uintptr)(unsafe.Pointer(getcallersp())) = abi.FuncPCABIInternal(shadowTrampolineASM)
	return
}

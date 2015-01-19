// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package channels

import (
	_core "runtime/internal/core"
	"unsafe"
)

//#define	MAXALIGN	8

type waitq struct {
	first *_core.Sudog
	last  *_core.Sudog
}

type Hchan struct {
	Qcount   uint // total data in the q
	Dataqsiz uint // size of the circular q
	Buf      *byte
	Elemsize uint16
	Closed   uint32
	Elemtype *_core.Type // element type
	sendx    uint        // send index
	recvx    uint        // receive index
	Recvq    waitq       // list of recv waiters
	Sendq    waitq       // list of send waiters
	Lock     _core.Mutex
}

// Buffer follows Hchan immediately in memory.
// chanbuf(c, i) is pointer to the i'th slot in the buffer.
// #define chanbuf(c, i) ((byte*)((c)->buf)+(uintptr)(c)->elemsize*(i))

const (
	// scase.kind
	CaseRecv = iota
	CaseSend
	CaseDefault
)

// Known to compiler.
// Changes here must also be made in src/cmd/gc/select.c's selecttype.
type Scase struct {
	Elem        unsafe.Pointer // data element
	Chan        *Hchan         // chan
	Pc          uintptr        // return pc
	Kind        uint16
	So          uint16 // vararg of selected bool
	receivedp   *bool  // pointer to received bool (recv2)
	releasetime int64
}

// Known to compiler.
// Changes here must also be made in src/cmd/gc/select.c's selecttype.
type Select struct {
	Tcase     uint16   // total count of scase[]
	Ncase     uint16   // currently filled scase[]
	pollorder *uint16  // case poll order
	lockorder **Hchan  // channel lock order
	Scase     [1]Scase // one per case (in order of appearance)
}

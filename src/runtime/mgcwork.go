// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_base "runtime/internal/base"
)

// putpartial puts empty buffers on the work.empty queue,
// full buffers on the work.full queue and
// others on the work.partial queue.
// entry is used to provide a brief history of ownership
// using entry + xxx00000 to
// indicating that two call chain line numbers.
//go:nowritebarrier
func putpartial(b *_base.Workbuf, entry int) {
	if b.Nobj == 0 {
		_base.Putempty(b, entry+81500000)
	} else if b.Nobj < len(b.Obj) {
		b.Logput(entry)
		_base.Lfstackpush(&_base.Work.Partial, &b.Node)
	} else if b.Nobj == len(b.Obj) {
		b.Logput(entry)
		_base.Lfstackpush(&_base.Work.Full, &b.Node)
	} else {
		_base.Throw("putpartial: bad Workbuf b.nobj")
	}
}

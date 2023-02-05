// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

// inlinedCall is the encoding of entries in the FUNCDATA_InlTree table.
type inlinedCall struct {
	funcID    funcID // type of the called function
	_         [3]byte
	nameOff   int32 // offset into pclntab for name of called function
	parentPc  int32 // position of an instruction whose source position is the call site (offset from entry)
	startLine int32 // line number of start of function (func keyword/TEXT directive)
}

// An inlineUnwinder iterates over the stack of inlined calls at a PC by
// decoding the inline table. The last step of iteration is always the frame of
// the physical function, so there's always at least one frame.
//
// This is typically used as:
//
//	for u := newInlineUnwinder(...); u.valid(); u.next() { ... }
//
// Implementation note: This is used in contexts that disallow write barriers.
// Hence, the constructor returns this by value and pointer receiver methods
// must not mutate pointer fields. Luckily, expanding inline frames requires
// very little state.
type inlineUnwinder struct {
	f       funcInfo
	cache   *pcvalueCache
	inlTree *[1 << 20]inlinedCall

	// pc is the PC giving the file/line metadata of the current frame. This is
	// 0 when the iterator is exhausted.
	pc uintptr

	// index is the index of the current record in inlTree, or -1 if we are in
	// the outermost function.
	index int32
}

// newInlineUnwinder creates an inlineUnwinder initially set to the inner-most
// inlined frame at PC.
//
// This unwinder uses non-strict handling of PC because it's assumed this is
// only ever used for symbolic debugging. If things go really wrong, it'll just
// fall back to the outermost frame.
func newInlineUnwinder(f funcInfo, pc uintptr, cache *pcvalueCache) inlineUnwinder {
	inldata := funcdata(f, _FUNCDATA_InlTree)
	if inldata == nil {
		return inlineUnwinder{f, nil, nil, pc, -1}
	}
	inlTree := (*[1 << 20]inlinedCall)(inldata)
	u := inlineUnwinder{f, cache, inlTree, 0, 0}
	u.resolveInternal(pc)
	return u
}

func (u *inlineUnwinder) resolveInternal(pc uintptr) {
	u.pc = pc
	// Conveniently, this returns -1 if there's an error, which is the same
	// value we use for the outermost frame.
	u.index = pcdatavalue1(u.f, _PCDATA_InlTreeIndex, pc, u.cache, false)
}

func (u *inlineUnwinder) valid() bool {
	return u.pc != 0
}

// next steps u to the next caller.
func (u *inlineUnwinder) next() {
	if u.index < 0 {
		u.pc = 0
		return
	}
	parentPc := u.inlTree[u.index].parentPc
	u.resolveInternal(u.f.entry() + uintptr(parentPc))
	return
}

// isInlined returns whether the current frame of u is an inlined frame.
func (u *inlineUnwinder) isInlined() bool {
	return u.index >= 0
}

// srcFunc returns the srcFunc representing the current frame.
func (u *inlineUnwinder) srcFunc() srcFunc {
	if u.index < 0 {
		return u.f.srcFunc()
	}
	t := &u.inlTree[u.index]
	return srcFunc{
		u.f.datap,
		t.nameOff,
		t.startLine,
		t.funcID,
	}
}

// fileLine returns the file name and line number of the call within the current
// frame. As a convenience, for the innermost frame, it returns the file and
// line of the PC this unwinder was started at (often this is a call to another
// physical function).
//
// It returns "?", 0 if something goes wrong.
func (u *inlineUnwinder) fileLine() (file string, line int) {
	file, line32 := funcline1(u.f, u.pc, false)
	return file, int(line32)
}

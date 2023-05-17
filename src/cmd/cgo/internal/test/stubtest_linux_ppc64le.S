// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Verify the call stubs inserted run on the target machine. The go linker
// should generate working stubs for any GOPPC64 and buildmode combination.

// A function which does not guarantee R2 is preserved.
	.globl	notoc_nor2_func
	.type	notoc_nor2_func, @function
notoc_nor2_func:
	.localentry notoc_nor2_func,1
	li	2,0
	blr

// A function which expects R2 to hold TOC, and has a distinct local entry.
	.globl	toc_func
	.type	toc_func, @function
toc_func:
	addis	2,12,.TOC.-toc_func@ha
	addi	2,2,.TOC.-toc_func@l
	.localentry toc_func, .-toc_func
	mflr	0
	std	0,16(1)
	stdu	1,-32(1)

	// Call a NOTOC function which clobbers R2.
	bl	notoc_nor2_func
	nop

	// Call libc random. This should generate a TOC relative plt stub.
	bl	random
	nop

	addi	1,1,32
	ld 	0,16(1)
	mtlr	0
	blr

// An ELFv2 st_other==0 function. It preserves R2 (TOC), but does not use it.
	.globl	notoc_func
	.type	notoc_func, @function
notoc_func:
	// Save R2 and LR and stack a frame.
	mflr	0
	std	0,16(1)
	stdu	1,-32(1)

	// Save R2 in TOC save slot.
	std	2,24(1)

	// clobber R2
	li	2,0
	
	// Call type2_func. A call stub from notoc to toc should be inserted.
	bl	toc_func@notoc

	// Call libc random. A notoc plt stub should be inserted.
	bl	random@notoc

	// Restore R2
	ld	2,24(1)

	// Restore LR and pop stack
	addi	1,1,32
	ld 	0,16(1)
	mtlr	0
	blr

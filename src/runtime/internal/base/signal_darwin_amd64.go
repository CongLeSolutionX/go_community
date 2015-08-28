// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package base

import (
	"unsafe"
)

type Sigctxt struct {
	Info *Siginfo
	Ctxt unsafe.Pointer
}

func (c *Sigctxt) regs() *regs64   { return &(*ucontext)(c.Ctxt).uc_mcontext.ss }
func (c *Sigctxt) rax() uint64     { return c.regs().rax }
func (c *Sigctxt) rbx() uint64     { return c.regs().rbx }
func (c *Sigctxt) rcx() uint64     { return c.regs().rcx }
func (c *Sigctxt) rdx() uint64     { return c.regs().rdx }
func (c *Sigctxt) rdi() uint64     { return c.regs().rdi }
func (c *Sigctxt) rsi() uint64     { return c.regs().rsi }
func (c *Sigctxt) rbp() uint64     { return c.regs().rbp }
func (c *Sigctxt) rsp() uint64     { return c.regs().rsp }
func (c *Sigctxt) r8() uint64      { return c.regs().r8 }
func (c *Sigctxt) r9() uint64      { return c.regs().r9 }
func (c *Sigctxt) r10() uint64     { return c.regs().r10 }
func (c *Sigctxt) r11() uint64     { return c.regs().r11 }
func (c *Sigctxt) r12() uint64     { return c.regs().r12 }
func (c *Sigctxt) r13() uint64     { return c.regs().r13 }
func (c *Sigctxt) r14() uint64     { return c.regs().r14 }
func (c *Sigctxt) r15() uint64     { return c.regs().r15 }
func (c *Sigctxt) rip() uint64     { return c.regs().rip }
func (c *Sigctxt) rflags() uint64  { return c.regs().rflags }
func (c *Sigctxt) cs() uint64      { return c.regs().cs }
func (c *Sigctxt) fs() uint64      { return c.regs().fs }
func (c *Sigctxt) gs() uint64      { return c.regs().gs }
func (c *Sigctxt) Sigcode() uint64 { return uint64(c.Info.si_code) }
func (c *Sigctxt) sigaddr() uint64 { return c.Info.si_addr }

func (c *Sigctxt) set_rip(x uint64)     { c.regs().rip = x }
func (c *Sigctxt) set_rsp(x uint64)     { c.regs().rsp = x }
func (c *Sigctxt) set_sigcode(x uint64) { c.Info.si_code = int32(x) }
func (c *Sigctxt) set_sigaddr(x uint64) { c.Info.si_addr = x }

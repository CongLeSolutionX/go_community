// created by cgo -cdefs and then converted to Go
// cgo -cdefs defs_darwin.go

package sched

import (
	_core "runtime/internal/core"
)

type siginfo struct {
	si_signo  int32
	si_errno  int32
	si_code   int32
	si_pid    int32
	si_uid    uint32
	si_status int32
	si_addr   *byte
	si_value  [8]byte
	si_band   int64
	__pad     [7]uint64
}

type timeval struct {
	tv_sec    int64
	tv_usec   int32
	pad_cgo_0 [4]byte
}

func (tv *timeval) set_usec(x int32) {
	tv.tv_usec = x
}

type itimerval struct {
	it_interval timeval
	it_value    timeval
}

type timespec struct {
	tv_sec  int64
	tv_nsec int64
}

type Fpcontrol struct {
	pad_cgo_0 [2]byte
}

type Fpstatus struct {
	pad_cgo_0 [2]byte
}

type Regmmst struct {
	mmst_reg  [10]int8
	mmst_rsrv [6]int8
}

type Regxmm struct {
	xmm_reg [16]int8
}

type regs64 struct {
	rax    uint64
	rbx    uint64
	rcx    uint64
	rdx    uint64
	rdi    uint64
	rsi    uint64
	rbp    uint64
	rsp    uint64
	r8     uint64
	r9     uint64
	r10    uint64
	r11    uint64
	r12    uint64
	r13    uint64
	r14    uint64
	r15    uint64
	rip    uint64
	rflags uint64
	cs     uint64
	fs     uint64
	gs     uint64
}

type floatstate64 struct {
	fpu_reserved  [2]int32
	fpu_fcw       Fpcontrol
	fpu_fsw       Fpstatus
	fpu_ftw       uint8
	fpu_rsrv1     uint8
	fpu_fop       uint16
	fpu_ip        uint32
	fpu_cs        uint16
	fpu_rsrv2     uint16
	fpu_dp        uint32
	fpu_ds        uint16
	fpu_rsrv3     uint16
	fpu_mxcsr     uint32
	fpu_mxcsrmask uint32
	fpu_stmm0     Regmmst
	fpu_stmm1     Regmmst
	fpu_stmm2     Regmmst
	fpu_stmm3     Regmmst
	fpu_stmm4     Regmmst
	fpu_stmm5     Regmmst
	fpu_stmm6     Regmmst
	fpu_stmm7     Regmmst
	fpu_xmm0      Regxmm
	fpu_xmm1      Regxmm
	fpu_xmm2      Regxmm
	fpu_xmm3      Regxmm
	fpu_xmm4      Regxmm
	fpu_xmm5      Regxmm
	fpu_xmm6      Regxmm
	fpu_xmm7      Regxmm
	fpu_xmm8      Regxmm
	fpu_xmm9      Regxmm
	fpu_xmm10     Regxmm
	fpu_xmm11     Regxmm
	fpu_xmm12     Regxmm
	fpu_xmm13     Regxmm
	fpu_xmm14     Regxmm
	fpu_xmm15     Regxmm
	fpu_rsrv4     [96]int8
	fpu_reserved1 int32
}

type exceptionstate64 struct {
	trapno     uint16
	cpu        uint16
	err        uint32
	faultvaddr uint64
}

type mcontext64 struct {
	es        exceptionstate64
	ss        regs64
	fs        floatstate64
	pad_cgo_0 [4]byte
}

type ucontext struct {
	uc_onstack  int32
	uc_sigmask  uint32
	uc_stack    _core.Stackt
	uc_link     *ucontext
	uc_mcsize   uint64
	uc_mcontext *mcontext64
}

type Keventt struct {
	Ident  uint64
	Filter int16
	Flags  uint16
	Fflags uint32
	Data   int64
	Udata  *byte
}

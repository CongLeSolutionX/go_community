// created by cgo -cdefs and then converted to Go
// cgo -cdefs defs_darwin.go

package base

import (
	"unsafe"
)

const (
	EINTR  = 0x4
	EFAULT = 0xe

	PROT_NONE  = 0x0
	PROT_READ  = 0x1
	PROT_WRITE = 0x2
	PROT_EXEC  = 0x4

	MAP_ANON    = 0x1000
	MAP_PRIVATE = 0x2
	MAP_FIXED   = 0x10

	MADV_DONTNEED = 0x4
	MADV_FREE     = 0x5

	MACH_MSG_TYPE_MOVE_RECEIVE   = 0x10
	MACH_MSG_TYPE_MOVE_SEND      = 0x11
	MACH_MSG_TYPE_MOVE_SEND_ONCE = 0x12
	MACH_MSG_TYPE_COPY_SEND      = 0x13
	MACH_MSG_TYPE_MAKE_SEND      = 0x14
	MACH_MSG_TYPE_MAKE_SEND_ONCE = 0x15
	MACH_MSG_TYPE_COPY_RECEIVE   = 0x16

	MACH_MSG_PORT_DESCRIPTOR         = 0x0
	MACH_MSG_OOL_DESCRIPTOR          = 0x1
	MACH_MSG_OOL_PORTS_DESCRIPTOR    = 0x2
	MACH_MSG_OOL_VOLATILE_DESCRIPTOR = 0x3

	MACH_MSGH_BITS_COMPLEX = 0x80000000

	MACH_SEND_MSG  = 0x1
	MACH_RCV_MSG   = 0x2
	MACH_RCV_LARGE = 0x4

	MACH_SEND_TIMEOUT   = 0x10
	MACH_SEND_INTERRUPT = 0x40
	MACH_SEND_ALWAYS    = 0x10000
	MACH_SEND_TRAILER   = 0x20000
	MACH_RCV_TIMEOUT    = 0x100
	MACH_RCV_NOTIFY     = 0x200
	MACH_RCV_INTERRUPT  = 0x400
	MACH_RCV_OVERWRITE  = 0x1000

	NDR_PROTOCOL_2_0      = 0x0
	NDR_INT_BIG_ENDIAN    = 0x0
	NDR_INT_LITTLE_ENDIAN = 0x1
	NDR_FLOAT_IEEE        = 0x0
	NDR_CHAR_ASCII        = 0x0

	SA_SIGINFO   = 0x40
	SA_RESTART   = 0x2
	SA_ONSTACK   = 0x1
	SA_USERTRAMP = 0x100
	SA_64REGSET  = 0x200

	SIGHUP    = 0x1
	SIGINT    = 0x2
	SIGQUIT   = 0x3
	SIGILL    = 0x4
	SIGTRAP   = 0x5
	SIGABRT   = 0x6
	SIGEMT    = 0x7
	SIGFPE    = 0x8
	SIGKILL   = 0x9
	SIGBUS    = 0xa
	SIGSEGV   = 0xb
	SIGSYS    = 0xc
	SIGPIPE   = 0xd
	SIGALRM   = 0xe
	SIGTERM   = 0xf
	SIGURG    = 0x10
	SIGSTOP   = 0x11
	SIGTSTP   = 0x12
	SIGCONT   = 0x13
	SIGCHLD   = 0x14
	SIGTTIN   = 0x15
	SIGTTOU   = 0x16
	SIGIO     = 0x17
	SIGXCPU   = 0x18
	SIGXFSZ   = 0x19
	SIGVTALRM = 0x1a
	SIGPROF   = 0x1b
	SIGWINCH  = 0x1c
	SIGINFO   = 0x1d
	SIGUSR1   = 0x1e
	SIGUSR2   = 0x1f

	FPE_INTDIV = 0x7
	FPE_INTOVF = 0x8
	FPE_FLTDIV = 0x1
	FPE_FLTOVF = 0x2
	FPE_FLTUND = 0x3
	FPE_FLTRES = 0x4
	FPE_FLTINV = 0x5
	FPE_FLTSUB = 0x6

	BUS_ADRALN = 0x1
	BUS_ADRERR = 0x2
	BUS_OBJERR = 0x3

	SEGV_MAPERR = 0x1
	SEGV_ACCERR = 0x2

	ITIMER_REAL    = 0x0
	ITIMER_VIRTUAL = 0x1
	ITIMER_PROF    = 0x2

	EV_ADD       = 0x1
	EV_DELETE    = 0x2
	EV_CLEAR     = 0x20
	EV_RECEIPT   = 0x40
	EV_ERROR     = 0x4000
	EVFILT_READ  = -0x1
	EVFILT_WRITE = -0x2
)

type Machbody struct {
	Msgh_descriptor_count uint32
}

type Machheader struct {
	Msgh_bits        uint32
	Msgh_size        uint32
	Msgh_remote_port uint32
	msgh_local_port  uint32
	msgh_reserved    uint32
	Msgh_id          int32
}

type machndr struct {
	mig_vers     uint8
	if_vers      uint8
	reserved1    uint8
	mig_encoding uint8
	int_rep      uint8
	char_rep     uint8
	float_rep    uint8
	reserved2    uint8
}

type Machport struct {
	Name        uint32
	pad1        uint32
	pad2        uint16
	Disposition uint8
	Type        uint8
}

type stackt struct {
	ss_sp     *byte
	ss_size   uintptr
	ss_flags  int32
	pad_cgo_0 [4]byte
}

type sigactiont struct {
	__sigaction_u [8]byte
	sa_tramp      unsafe.Pointer
	sa_mask       uint32
	sa_flags      int32
}

type Siginfo struct {
	si_signo  int32
	si_errno  int32
	si_code   int32
	si_pid    int32
	si_uid    uint32
	si_status int32
	si_addr   uint64
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
	uc_stack    stackt
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

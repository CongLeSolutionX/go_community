// created by cgo -cdefs and then converted to Go
// cgo -cdefs defs_darwin.go

package runtime

import (
	_base "runtime/internal/base"
)

type regs32 struct {
	eax    uint32
	ebx    uint32
	ecx    uint32
	edx    uint32
	edi    uint32
	esi    uint32
	ebp    uint32
	esp    uint32
	ss     uint32
	eflags uint32
	eip    uint32
	cs     uint32
	ds     uint32
	es     uint32
	fs     uint32
	gs     uint32
}

type floatstate32 struct {
	fpu_reserved  [2]int32
	fpu_fcw       _base.Fpcontrol
	fpu_fsw       _base.Fpstatus
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
	fpu_stmm0     _base.Regmmst
	fpu_stmm1     _base.Regmmst
	fpu_stmm2     _base.Regmmst
	fpu_stmm3     _base.Regmmst
	fpu_stmm4     _base.Regmmst
	fpu_stmm5     _base.Regmmst
	fpu_stmm6     _base.Regmmst
	fpu_stmm7     _base.Regmmst
	fpu_xmm0      _base.Regxmm
	fpu_xmm1      _base.Regxmm
	fpu_xmm2      _base.Regxmm
	fpu_xmm3      _base.Regxmm
	fpu_xmm4      _base.Regxmm
	fpu_xmm5      _base.Regxmm
	fpu_xmm6      _base.Regxmm
	fpu_xmm7      _base.Regxmm
	fpu_rsrv4     [224]int8
	fpu_reserved1 int32
}

type exceptionstate32 struct {
	trapno     uint16
	cpu        uint16
	err        uint32
	faultvaddr uint32
}

type mcontext32 struct {
	es exceptionstate32
	ss regs32
	fs floatstate32
}

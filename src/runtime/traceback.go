// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_base "runtime/internal/base"
	_gc "runtime/internal/gc"
	_iface "runtime/internal/iface"
)

func tracebackinit() {
	// Go variable initialization happens late during runtime startup.
	// Instead of initializing the variables above in the declarations,
	// schedinit calls this function so that the variables are
	// initialized and available earlier in the startup sequence.
	_base.GoexitPC = _base.FuncPC(_base.Goexit)
	_base.JmpdeferPC = _base.FuncPC(jmpdefer)
	_base.McallPC = _base.FuncPC(_base.Mcall)
	_base.MorestackPC = _base.FuncPC(morestack)
	_base.MstartPC = _base.FuncPC(_base.Mstart)
	_base.Rt0_goPC = _base.FuncPC(rt0_go)
	_base.SigpanicPC = _base.FuncPC(_base.Sigpanic)
	_base.RunfinqPC = _base.FuncPC(runfinq)
	_base.BackgroundgcPC = _base.FuncPC(_iface.Backgroundgc)
	_base.BgsweepPC = _base.FuncPC(bgsweep)
	_base.ForcegchelperPC = _base.FuncPC(forcegchelper)
	_base.TimerprocPC = _base.FuncPC(_iface.Timerproc)
	_base.GcBgMarkWorkerPC = _base.FuncPC(_gc.GcBgMarkWorker)
	_gc.Systemstack_switchPC = _base.FuncPC(systemstack_switch)
	_base.SystemstackPC = _base.FuncPC(_base.Systemstack)
	_base.StackBarrierPC = _base.FuncPC(stackBarrier)

	// used by sigprof handler
	_base.GogoPC = _base.FuncPC(_base.Gogo)
}

var gScanStatusStrings = [...]string{
	0:               "scan",
	_base.Grunnable: "scanrunnable",
	_base.Grunning:  "scanrunning",
	_base.Gsyscall:  "scansyscall",
	_base.Gwaiting:  "scanwaiting",
	_base.Gdead:     "scandead",
	_base.Genqueue:  "scanenqueue",
}

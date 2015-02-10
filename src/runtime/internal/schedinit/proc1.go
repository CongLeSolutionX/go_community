// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schedinit

import (
	_core "runtime/internal/core"
	_gc "runtime/internal/gc"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
)

// The bootstrap sequence is:
//
//	call osinit
//	call schedinit
//	make & queue new G
//	call runtime·mstart
//
// The new G calls runtime·main.
func schedinit() {
	// raceinit must be the first call to race detector.
	// In particular, it must be done before mallocinit below calls racemapshadow.
	_g_ := _core.Getg()
	if _sched.Raceenabled {
		_g_.Racectx = raceinit()
	}

	_core.Sched.Maxmcount = 10000

	// Cache the framepointer experiment.  This affects stack unwinding.
	_lock.Framepointer_enabled = haveexperiment("framepointer")

	tracebackinit()
	symtabinit()
	stackinit()
	mallocinit()
	_sched.Mcommoninit(_g_.M)

	goargs()
	goenvs()
	parsedebugvars()
	wbshadowinit()
	gcinit()

	_core.Sched.Lastpoll = uint64(_lock.Nanotime())
	procs := 1
	if n := atoi(Gogetenv("GOMAXPROCS")); n > 0 {
		if n > _lock.MaxGomaxprocs {
			n = _lock.MaxGomaxprocs
		}
		procs = n
	}
	if _gc.Procresize(int32(procs)) != nil {
		_lock.Throw("unknown runnable goroutine during bootstrap")
	}

	if buildVersion == "" {
		// Condition should never trigger.  This code just serves
		// to ensure runtime·buildVersion is kept in the resulting binary.
		buildVersion = "unknown"
	}
}

func haveexperiment(name string) bool {
	x := Goexperiment
	for x != "" {
		xname := ""
		i := _lock.Index(x, ",")
		if i < 0 {
			xname, x = x, ""
		} else {
			xname, x = x[:i], x[i+1:]
		}
		if xname == name {
			return true
		}
	}
	return false
}

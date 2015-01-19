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

/*
SchedT	sched;
int32	gomaxprocs;
uint32	needextram;
bool	iscgo;
M	m0;
G	g0;	// idle goroutine for m0
G*	lastg;
M*	allm;
M*	extram;
P*	allp[MaxGomaxprocs+1];
int8*	goos;
int32	ncpu;
int32	newprocs;

Mutex allglock;	// the following vars are protected by this lock or by stoptheworld
G**	allg;
Slice	allgs;
uintptr allglen;
ForceGCState	forcegc;

void mstart(void);
static void runqput(P*, G*);
static G* runqget(P*);
static bool runqputslow(P*, G*, uint32, uint32);
static G* runqsteal(P*, P*);
static void mput(M*);
static M* mget(void);
static void mcommoninit(M*);
static void schedule(void);
static void procresize(int32);
static void acquirep(P*);
static P* releasep(void);
static void newm(void(*)(void), P*);
static void stopm(void);
static void startm(P*, bool);
static void handoffp(P*);
static void wakep(void);
static void stoplockedm(void);
static void startlockedm(G*);
static void sysmon(void);
static uint32 retake(int64);
static void incidlelocked(int32);
static void checkdead(void);
static void exitsyscall0(G*);
void park_m(G*);
static void goexit0(G*);
static void gfput(P*, G*);
static G* gfget(P*);
static void gfpurge(P*);
static void globrunqput(G*);
static void globrunqputbatch(G*, G*, int32);
static G* globrunqget(P*, int32);
static P* pidleget(void);
static void pidleput(P*);
static void injectglist(G*);
static bool preemptall(void);
static bool preemptone(P*);
static bool exitsyscallfast(void);
static bool haveexperiment(int8*);
void allgadd(G*);
static void dropg(void);

extern String buildVersion;
*/

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

	tracebackinit()
	symtabinit()
	stackinit()
	mallocinit()
	_sched.Mcommoninit(_g_.M)

	goargs()
	goenvs()
	parsedebugvars()
	gcinit()

	_core.Sched.Lastpoll = uint64(_lock.Nanotime())
	procs := 1
	if n := goatoi(Gogetenv("GOMAXPROCS")); n > 0 {
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

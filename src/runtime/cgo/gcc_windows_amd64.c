// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#define WIN32_LEAN_AND_MEAN
#include <windows.h>
#include <process.h>
#include <psapi.h>
#include <stdlib.h>
#include <stdio.h>
#include <errno.h>
#include <inttypes.h>
#include "libcgo.h"
#include "libcgo_windows.h"

static void threadentry(void*);
static void (*setg_gcc)(void*);

void
x_cgo_init(G *g, void (*setg)(void*), void **tlsg, void **tlsbase)
{
	setg_gcc = setg;
}

void
dump_memory(void)
{
	fprintf(stderr, "runtime: process memory stats:\n");

	PROCESS_MEMORY_COUNTERS pmc;
	BOOL ok = GetProcessMemoryInfo((HANDLE)-1, &pmc, sizeof(pmc));
	if (!ok) {
		fprintf(stderr, "runtime: failed to query process memory stats (%ld)\n", GetLastError());
		return;
	}

	fprintf(stderr, "\tPageFaultCount: 0x%08" PRIx64 "\n", (uint64_t)pmc.PageFaultCount);
	fprintf(stderr, "\tPeakWorkingSetSize: 0x%08" PRIx64 "\n", (uint64_t)pmc.PeakWorkingSetSize);
	fprintf(stderr, "\tWorkingSetSize: 0x%08" PRIx64 "\n", (uint64_t)pmc.WorkingSetSize);
	fprintf(stderr, "\tQuotaPeakPagedPoolUsage: 0x%08" PRIx64 "\n", (uint64_t)pmc.QuotaPeakPagedPoolUsage);
	fprintf(stderr, "\tQuotaPagedPoolUsage: 0x%08" PRIx64 "\n", (uint64_t)pmc.QuotaPagedPoolUsage);
	fprintf(stderr, "\tQuotaPeakNonPagedPoolUsage: 0x%08" PRIx64 "\n", (uint64_t)pmc.QuotaPeakNonPagedPoolUsage);
	fprintf(stderr, "\tQuotaNonPagedPoolUsage: 0x%08" PRIx64 "\n", (uint64_t)pmc.QuotaNonPagedPoolUsage);
	fprintf(stderr, "\tPagefileUsage: 0x%08" PRIx64 "\n", (uint64_t)pmc.PagefileUsage);
	fprintf(stderr, "\tPeakPagefileUsage: 0x%08" PRIx64 "\n", (uint64_t)pmc.PeakPagefileUsage);
	//fprintf(stderr, "\tPrivateUsage: 0x%08" PRIx64 "\n", (uint64_t)pmc.PrivateUsage);

	fprintf(stderr, "runtime: global memory stats:\n");

	MEMORYSTATUSEX statex;
	ok = GlobalMemoryStatusEx(&statex);
	if (!ok) {
		fprintf(stderr, "runtime: failed to query global memory stats (%ld)\n", GetLastError());
		return;
	}

	fprintf(stderr, "\tMemoryLoad: %" PRIu64 "\n", (uint64_t)statex.dwMemoryLoad);
	fprintf(stderr, "\tTotalPhys: 0x%08" PRIx64 "\n", (uint64_t)statex.ullTotalPhys);
	fprintf(stderr, "\tAvailPhys: 0x%08" PRIx64 "\n", (uint64_t)statex.ullAvailPhys);
	fprintf(stderr, "\tTotalPageFile: 0x%08" PRIx64 "\n", (uint64_t)statex.ullTotalPageFile);
	fprintf(stderr, "\tAvailPageFile: 0x%08" PRIx64 "\n", (uint64_t)statex.ullAvailPageFile);
	fprintf(stderr, "\tTotalVirtual: 0x%08" PRIx64 "\n", (uint64_t)statex.ullTotalVirtual);
	fprintf(stderr, "\tAvailVirtual: 0x%08" PRIx64 "\n", (uint64_t)statex.ullAvailVirtual);
}

void
_cgo_sys_thread_start(ThreadStart *ts)
{
	uintptr_t thandle;

	thandle = _beginthread(threadentry, 0, ts);
	if(thandle == -1) {
		fprintf(stderr, "runtime: failed to create new OS thread (%d)\n", errno);
		dump_memory();
		abort();
	}
}

static void
threadentry(void *v)
{
	ThreadStart ts;

	ts = *(ThreadStart*)v;
	free(v);

	// minit queries stack bounds from the OS.

	/*
	 * Set specific keys in thread local storage.
	 */
	asm volatile (
	  "movq %0, %%gs:0x28\n"	// MOVL tls0, 0x28(GS)
	  :: "r"(ts.tls)
	);

	crosscall_amd64(ts.fn, setg_gcc, (void*)ts.g);
}

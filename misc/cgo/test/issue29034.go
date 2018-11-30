// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !windows

package cgotest

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"
	"testing"
	"unsafe"
)

// The fake traceback must include at least 2 frames since the bottom
// runtime.goexit frame is deliberately dropped by runtime/pprof on output.
// Also the leaf PC (0x100) is incremented by runtime/pprof to correct the
// address from the signal, hence the top function in this test is
// "dummypc_101".

/*
#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>
#include <sys/time.h>

struct cgoTracebackArg {
    uintptr_t  context;
    uintptr_t  sigContext;
    uintptr_t* buf;
    uintptr_t  max;
};

struct cgoSymbolizerArg {
    uintptr_t   pc;
    const char* file;
    uintptr_t   lineno;
    const char* func;
    uintptr_t   entry;
    uintptr_t   more;
    uintptr_t   data;
};

void cgoTraceback(void* parg) {
    struct cgoTracebackArg* arg = (struct cgoTracebackArg*)(parg);
    arg->buf[0] = 0x100;
    arg->buf[1] = 0x200;
    arg->buf[2] = 0x300;
    arg->buf[3] = 0;
}
void cgoSymbolizer(void* parg) {
    struct cgoSymbolizerArg* arg = (struct cgoSymbolizerArg*)(parg);
    if (!arg->pc) {
        return;
    }
    char *fn = malloc(32);
    snprintf(fn, 32, "dummypc_%x", arg->pc);
    arg->file = "dummyfile.c";
    arg->lineno = arg->pc;
    arg->func = fn;
    arg->entry = arg->pc - 16;
    arg->more = 0;
    arg->data = 0;
}

void spin(struct timeval timeout) {
    struct timeval deadline, now;
    gettimeofday(&deadline, NULL);
    timeradd(&deadline, &timeout, &deadline);
    do {
        gettimeofday(&now, NULL);
    } while (timercmp(&now, &deadline, <));
}
*/
import "C"

func nextTrace(rd *bufio.Scanner) []string {
	var trace []string
	for rd.Scan() {
		line := rd.Text()
		if strings.HasPrefix(line, "---") {
			return trace
		}

		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) == 0 {
			continue
		}
		// Last field contains the function name.
		trace = append(trace, fields[len(fields)-1])
	}
	return nil
}

func parseTraces(buf []byte) [][]string {
	s := bufio.NewScanner(bytes.NewReader(buf))
	nextTrace(s) // Skip the header.
	var traces [][]string
	for t := nextTrace(s); len(t) > 0; t = nextTrace(s) {
		traces = append(traces, t)
	}
	return traces
}

func findTrace(traces [][]string, top string) []string {
	for _, t := range traces {
		if t[0] == top {
			return t
		}
	}
	return nil
}

func init() {
	if os.Getenv("test29034HelperProcess") == "1" {
		runtime.SetCgoTraceback(0, unsafe.Pointer(C.cgoTraceback), nil, unsafe.Pointer(C.cgoSymbolizer))
	}
}

func goCmd() string {
	var exeSuffix string
	if runtime.GOOS == "windows" {
		exeSuffix = ".exe"
	}
	path := filepath.Join(runtime.GOROOT(), "bin", "go"+exeSuffix)
	if _, err := os.Stat(path); err == nil {
		return path
	}
	return "go"
}

func test29034Helper() {
	f, err := ioutil.TempFile("", "issue29034-")
	if err != nil {
		fmt.Fprintf(os.Stderr, "create profile failed: %v", err)
		os.Exit(1)
	}
	pprof.StartCPUProfile(f)
	C.spin(C.struct_timeval{0, 100000})
	pprof.StopCPUProfile()
	f.Close()
	defer os.Remove(f.Name())

	cmd := exec.Command(goCmd(), "tool", "pprof", "-traces", f.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "pprof failed: %v\n", err)
		os.Exit(1)
	}
}

func test29034(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	if os.Getenv("test29034HelperProcess") == "1" {
		test29034Helper()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=Test29034")
	cmd.Env = append(os.Environ(), "test29034HelperProcess=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("helper failed: %v", err)
	}

	traces := parseTraces(out)
	if len(traces) == 0 {
		t.Fatalf("pprof traces not found:\n%s", out)
	}

	const top = "dummypc_101"
	trace := findTrace(traces, top)
	if len(trace) == 0 {
		t.Fatalf("%s traceback missing. Found:\n%v", top, traces)
	}
	if trace[len(trace)-1] != "runtime.main" {
		t.Fatalf("invalid %s traceback origin: got=%v; want=[... runtime.main]", top, trace)
	}
}

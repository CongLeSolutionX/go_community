// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build windows

package exec_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"testing"
)

var (
	quitSignal os.Signal = nil
	pipeSignal           = syscall.SIGPIPE
)

func init() {
	registerHelperCommand("pipehandle", cmdPipeHandle)
}

func cmdPipeHandle(args ...string) {
	handle, _ := strconv.ParseUint(args[0], 16, 64)
	pipe := os.NewFile(uintptr(handle), "")
	_, err := fmt.Fprint(pipe, args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "writing to pipe failed: %v\n", err)
		os.Exit(1)
	}
	pipe.Close()
}

func TestPipePassing(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Error(err)
	}
	const marker = "arrakis, dune, desert planet"
	childProc := helperCommand(t, "pipehandle", strconv.FormatUint(uint64(w.Fd()), 16), marker)
	childProc.SysProcAttr = &syscall.SysProcAttr{AdditionalInheritedHandles: []syscall.Handle{syscall.Handle(w.Fd())}}
	err = childProc.Start()
	if err != nil {
		t.Error(err)
	}
	w.Close()
	response, err := io.ReadAll(r)
	if err != nil {
		t.Error(err)
	}
	r.Close()
	if string(response) != marker {
		t.Errorf("got %q; want %q", string(response), marker)
	}
	err = childProc.Wait()
	if err != nil {
		t.Error(err)
	}
}

func TestNoInheritHandles(t *testing.T) {
	cmd := exec.Command("cmd", "/c exit 88")
	cmd.SysProcAttr = &syscall.SysProcAttr{NoInheritHandles: true}
	err := cmd.Run()
	exitError, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("got error %v; want ExitError", err)
	}
	if exitError.ExitCode() != 88 {
		t.Fatalf("got exit code %d; want 88", exitError.ExitCode())
	}
}

// start a child process without the user code explicitly starting
// with a copy of the parent's. (The Windows SYSTEMROOT issue: Issue
// 25210)
func TestChildCriticalEnv(t *testing.T) {
	cmd := helperCommand(t, "echoenv", "SYSTEMROOT")
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(out)) == "" {
		t.Error("no SYSTEMROOT found")
	}
}

func TestStartRejectsUnsupportedInterrupt(t *testing.T) {
	for _, sig := range []os.Signal{
		os.Interrupt, // explicitly not implemented

		// “invented values” as described by the syscall package.
		syscall.SIGHUP,
		syscall.SIGQUIT,

		// Note that os.Kill actually is supported, and is tested separately.
	} {
		t.Run(sig.String(), func(t *testing.T) {
			cmd := exec.CommandContext(context.Background(), exePath(), "-sleep=1ms")
			cmd.Interrupt = sig
			err := cmd.Start()

			if err == nil {
				t.Errorf("Start succeeded unexpectedly")
				cmd.Wait()
			} else if !errors.Is(err, syscall.EWINDOWS) {
				t.Errorf("Start: %v\nwant %v", err, syscall.EWINDOWS)
			}
		})
	}
}

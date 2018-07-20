// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime_test

import (
	"bytes"
	"internal/testenv"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"testing"
	"time"
)

func TestPanicSystemstack(t *testing.T) {
	// Test that GOTRACEBACK=crash prints both the system and user
	// stack of other threads.

	// The GOTRACEBACK=crash handler takes 0.1 seconds even if
	// it's not writing a core file and potentially much longer if
	// it is. Skip in short mode.
	if testing.Short() {
		t.Skip("Skipping in short mode (GOTRACEBACK=crash is slow)")
	}

	t.Parallel()
	cmd := exec.Command(os.Args[0], "testPanicSystemstackInternal")
	cmd = testenv.CleanCmdEnv(cmd)
	cmd.Env = append(cmd.Env, "GOTRACEBACK=crash")
	pr, pw, err := os.Pipe()
	if err != nil {
		t.Fatal("creating pipe: ", err)
	}
	cmd.Stderr = pw
	if err := cmd.Start(); err != nil {
		t.Fatal("starting command: ", err)
	}
	defer cmd.Process.Wait()
	defer cmd.Process.Kill()
	if err := pw.Close(); err != nil {
		t.Log("closing write pipe: ", err)
	}
	defer pr.Close()

	// Wait for "x\nx\n" to indicate readiness.
	buf := make([]byte, 4)
	_, err = io.ReadFull(pr, buf)
	if err != nil || string(buf) != "x\nx\n" {
		t.Fatal("subprocess failed; output:\n", string(buf))
	}

	// Get traceback.
	tb, err := ioutil.ReadAll(pr)
	if err != nil {
		t.Fatal("reading traceback from pipe: ", err)
	}

	// Traceback should have two testPanicSystemstackInternal's
	// and two blockOnSystemStackInternal's.
	if bytes.Count(tb, []byte("testPanicSystemstackInternal")) != 2 {
		t.Fatal("traceback missing user stack:\n", string(tb))
	} else if bytes.Count(tb, []byte("blockOnSystemStackInternal")) != 2 {
		t.Fatal("traceback missing system stack:\n", string(tb))
	}
}

func init() {
	if len(os.Args) >= 2 && os.Args[1] == "testPanicSystemstackInternal" {
		// Get three threads running on the system stack with
		// something recognizable in the stack trace.
		runtime.GOMAXPROCS(3)
		go testPanicSystemstackInternal()
		go testPanicSystemstackRaiseException()
		testPanicSystemstackInternal()
	}
}

func testPanicSystemstackRaiseException() {
	// TODO: !!!! wait for less then one second or do not wait at all !!!!
	time.Sleep(1 * time.Second)

	const EXCEPTION_NONCONTINUABLE = 1
	mod := syscall.MustLoadDLL("kernel32.dll")
	proc := mod.MustFindProc("RaiseException")
	proc.Call(0xbad, EXCEPTION_NONCONTINUABLE, 0, 0)
	println("RaiseException should not return")

	os.Exit(1) // Should be unreachable.
}

func testPanicSystemstackInternal() {
	runtime.BlockOnSystemStack()
	os.Exit(1) // Should be unreachable.
}

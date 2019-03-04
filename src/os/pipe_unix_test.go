// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !windows,!plan9,!nacl,!js

package os_test

import (
	"fmt"
	"internal/testenv"
	"os"
	osexec "os/exec"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"
)

func TestEPIPE(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}

	// Every time we write to the pipe we should get an EPIPE.
	for i := 0; i < 20; i++ {
		_, err = w.Write([]byte("hi"))
		if err == nil {
			t.Fatal("unexpected success of Write to broken pipe")
		}
		if pe, ok := err.(*os.PathError); ok {
			err = pe.Err
		}
		if se, ok := err.(*os.SyscallError); ok {
			err = se.Err
		}
		if err != syscall.EPIPE {
			t.Errorf("iteration %d: got %v, expected EPIPE", i, err)
		}
	}
}

// Test broken pipes on Unix systems.
func TestStdPipe(t *testing.T) {
	testenv.MustHaveExec(t)
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
	// Invoke the test program to run the test and write to a closed pipe.
	// If sig is false:
	// writing to stdout or stderr should cause an immediate SIGPIPE;
	// writing to descriptor 3 should fail with EPIPE and then exit 0.
	// If sig is true:
	// all writes should fail with EPIPE and then exit 0.
	for _, sig := range []bool{false, true} {
		for dest := 1; dest < 4; dest++ {
			cmd := osexec.Command(os.Args[0], "-test.run", "TestStdPipeHelper")
			cmd.Stdout = w
			cmd.Stderr = w
			cmd.ExtraFiles = []*os.File{w}
			cmd.Env = append(os.Environ(), fmt.Sprintf("GO_TEST_STD_PIPE_HELPER=%d", dest))
			if sig {
				cmd.Env = append(cmd.Env, "GO_TEST_STD_PIPE_HELPER_SIGNAL=1")
			}
			if err := cmd.Run(); err == nil {
				if !sig && dest < 3 {
					t.Errorf("unexpected success of write to closed pipe %d sig %t in child", dest, sig)
				}
			} else if ee, ok := err.(*osexec.ExitError); !ok {
				t.Errorf("unexpected exec error type %T: %v", err, err)
			} else if ws, ok := ee.Sys().(syscall.WaitStatus); !ok {
				t.Errorf("unexpected wait status type %T: %v", ee.Sys(), ee.Sys())
			} else if ws.Signaled() && ws.Signal() == syscall.SIGPIPE {
				if sig || dest > 2 {
					t.Errorf("unexpected SIGPIPE signal for descriptor %d sig %t", dest, sig)
				}
			} else {
				t.Errorf("unexpected exit status %v for descriptor %d sig %t", err, dest, sig)
			}
		}
	}
}

// This is a helper for TestStdPipe. It's not a test in itself.
func TestStdPipeHelper(t *testing.T) {
	if os.Getenv("GO_TEST_STD_PIPE_HELPER_SIGNAL") != "" {
		signal.Notify(make(chan os.Signal, 1), syscall.SIGPIPE)
	}
	switch os.Getenv("GO_TEST_STD_PIPE_HELPER") {
	case "1":
		os.Stdout.Write([]byte("stdout"))
	case "2":
		os.Stderr.Write([]byte("stderr"))
	case "3":
		if _, err := os.NewFile(3, "3").Write([]byte("3")); err == nil {
			os.Exit(3)
		}
	default:
		t.Skip("skipping test helper")
	}
	// For stdout/stderr, we should have crashed with a broken pipe error.
	// The caller will be looking for that exit status,
	// so just exit normally here to cause a failure in the caller.
	// For descriptor 3, a normal exit is expected.
	os.Exit(0)
}

// Issue 20915: Reading on nonblocking fd should not return "waiting
// for unsupported file type." Currently it returns EAGAIN; it is
// possible that in the future it will simply wait for data.
func TestReadNonblockingFd(t *testing.T) {
	if os.Getenv("GO_WANT_READ_NONBLOCKING_FD") == "1" {
		fd := int(os.Stdin.Fd())
		syscall.SetNonblock(fd, true)
		defer syscall.SetNonblock(fd, false)
		_, err := os.Stdin.Read(make([]byte, 1))
		if err != nil {
			if perr, ok := err.(*os.PathError); !ok || perr.Err != syscall.EAGAIN {
				t.Fatalf("read on nonblocking stdin got %q, should have gotten EAGAIN", err)
			}
		}
		os.Exit(0)
	}

	testenv.MustHaveExec(t)
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer w.Close()
	cmd := osexec.Command(os.Args[0], "-test.run="+t.Name())
	cmd.Env = append(os.Environ(), "GO_WANT_READ_NONBLOCKING_FD=1")
	cmd.Stdin = r
	output, err := cmd.CombinedOutput()
	t.Logf("%s", output)
	if err != nil {
		t.Errorf("child process failed: %v", err)
	}
}

func TestCloseWithBlockingReadByNewFile(t *testing.T) {
	var p [2]int
	err := syscall.Pipe(p[:])
	if err != nil {
		t.Fatal(err)
	}
	// os.NewFile returns a blocking mode file.
	testCloseWithBlockingRead(t, os.NewFile(uintptr(p[0]), "reader"), os.NewFile(uintptr(p[1]), "writer"))
}

func TestCloseWithBlockingReadByFd(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	// Calling Fd will put the file into blocking mode.
	_ = r.Fd()
	testCloseWithBlockingRead(t, r, w)
}

// Issue 24481.
func TestFdRace(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer w.Close()

	var wg sync.WaitGroup
	call := func() {
		defer wg.Done()
		w.Fd()
	}

	const tries = 100
	for i := 0; i < tries; i++ {
		wg.Add(1)
		go call()
	}
	wg.Wait()
}

func TestFdReadRace(t *testing.T) {
	t.Parallel()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer w.Close()

	c := make(chan bool)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		var buf [10]byte
		r.SetReadDeadline(time.Now().Add(time.Second))
		c <- true
		if _, err := r.Read(buf[:]); os.IsTimeout(err) {
			t.Error("read timed out")
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-c
		// Give the other goroutine a chance to enter the Read.
		// It doesn't matter if this occasionally fails, the test
		// will still pass, it just won't test anything.
		time.Sleep(10 * time.Millisecond)
		r.Fd()

		// The bug was that Fd would hang until Read timed out.
		// If the bug is fixed, then closing r here will cause
		// the Read to exit before the timeout expires.
		r.Close()
	}()

	wg.Wait()
}

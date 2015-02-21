// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pprof_test

import (
	"bytes"
	"internal/trace"
	"net"
	"os"
	"runtime"
	. "runtime/pprof"
	"sync"
	"testing"
	"time"
)

func skipTraceTestsIfNeeded(t *testing.T) {
	switch runtime.GOOS {
	case "solaris":
		t.Skip("skipping: solaris timer can go backwards (http://golang.org/issue/8976)")
	}

	switch runtime.GOARCH {
	case "arm":
		t.Skip("skipping: arm tests fail with 'failed to parse trace' (http://golang.org/issue/9725)")
	}
}

func TestTraceStartStop(t *testing.T) {
	skipTraceTestsIfNeeded(t)
	buf := new(bytes.Buffer)
	if err := StartTrace(buf); err != nil {
		t.Fatalf("failed to start tracing: %v", err)
	}
	StopTrace()
	size := buf.Len()
	if size == 0 {
		t.Fatalf("trace is empty")
	}
	time.Sleep(100 * time.Millisecond)
	if size != buf.Len() {
		t.Fatalf("trace writes after stop: %v -> %v", size, buf.Len())
	}
}

func TestTraceDoubleStart(t *testing.T) {
	skipTraceTestsIfNeeded(t)
	StopTrace()
	buf := new(bytes.Buffer)
	if err := StartTrace(buf); err != nil {
		t.Fatalf("failed to start tracing: %v", err)
	}
	if err := StartTrace(buf); err == nil {
		t.Fatalf("succeed to start tracing second time")
	}
	StopTrace()
	StopTrace()
}

func TestTrace(t *testing.T) {
	skipTraceTestsIfNeeded(t)
	buf := new(bytes.Buffer)
	if err := StartTrace(buf); err != nil {
		t.Fatalf("failed to start tracing: %v", err)
	}
	StopTrace()
	_, err := trace.Parse(buf)
	if err != nil {
		t.Fatalf("failed to parse trace: %v", err)
	}
}

func TestTraceStress(t *testing.T) {
	skipTraceTestsIfNeeded(t)

	var wg sync.WaitGroup
	done := make(chan bool)

	// Create a goroutine blocked before tracing.
	wg.Add(1)
	go func() {
		<-done
		wg.Done()
	}()

	// Create a goroutine blocked in syscall before tracing.
	rp, wp, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	defer func() {
		rp.Close()
		wp.Close()
	}()
	wg.Add(1)
	go func() {
		var tmp [1]byte
		rp.Read(tmp[:])
		<-done
		wg.Done()
	}()
	time.Sleep(time.Millisecond)

	buf := new(bytes.Buffer)
	if err := StartTrace(buf); err != nil {
		t.Fatalf("failed to start tracing: %v", err)
	}

	procs := runtime.GOMAXPROCS(10)

	go func() {
		runtime.LockOSThread()
		for {
			select {
			case <-done:
				return
			default:
				runtime.Gosched()
			}
		}
	}()

	runtime.GC()
	// Trigger GC from malloc.
	for i := 0; i < 1e3; i++ {
		_ = make([]byte, 1<<20)
	}

	// Create a bunch of busy goroutines to load all Ps.
	for p := 0; p < 10; p++ {
		wg.Add(1)
		go func() {
			// Do something useful.
			tmp := make([]byte, 1<<16)
			for i := range tmp {
				tmp[i]++
			}
			_ = tmp
			<-done
			wg.Done()
		}()
	}

	// Block in syscall.
	wg.Add(1)
	go func() {
		var tmp [1]byte
		rp.Read(tmp[:])
		<-done
		wg.Done()
	}()

	// Test timers.
	timerDone := make(chan bool)
	go func() {
		time.Sleep(time.Millisecond)
		timerDone <- true
	}()
	<-timerDone

	// A bit of network.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	defer ln.Close()
	go func() {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		time.Sleep(time.Millisecond)
		var buf [1]byte
		c.Write(buf[:])
		c.Close()
	}()
	c, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	var tmp [1]byte
	c.Read(tmp[:])
	c.Close()

	go func() {
		runtime.Gosched()
		select {}
	}()

	// Unblock helper goroutines and wait them to finish.
	wp.Write(tmp[:])
	wp.Write(tmp[:])
	close(done)
	wg.Wait()

	runtime.GOMAXPROCS(procs)

	StopTrace()
	_, err = trace.Parse(buf)
	if err != nil {
		t.Fatalf("failed to parse trace: %v", err)
	}
}

// Do a bunch of various stuff (timers, GC, network, etc) in a separate goroutine.
// And concurrently with all that start/stop trace 3 times.
func TestTraceStressStartStop(t *testing.T) {
	skipTraceTestsIfNeeded(t)

	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(8))
	outerDone := make(chan bool)

	go func() {
		defer func() {
			outerDone <- true
		}()

		var wg sync.WaitGroup
		done := make(chan bool)

		wg.Add(1)
		go func() {
			<-done
			wg.Done()
		}()

		rp, wp, err := os.Pipe()
		if err != nil {
			t.Fatalf("failed to create pipe: %v", err)
		}
		defer func() {
			rp.Close()
			wp.Close()
		}()
		wg.Add(1)
		go func() {
			var tmp [1]byte
			rp.Read(tmp[:])
			<-done
			wg.Done()
		}()
		time.Sleep(time.Millisecond)

		go func() {
			runtime.LockOSThread()
			for {
				select {
				case <-done:
					return
				default:
					runtime.Gosched()
				}
			}
		}()

		runtime.GC()
		// Trigger GC from malloc.
		for i := 0; i < 1e3; i++ {
			_ = make([]byte, 1<<20)
		}

		// Create a bunch of busy goroutines to load all Ps.
		for p := 0; p < 10; p++ {
			wg.Add(1)
			go func() {
				// Do something useful.
				tmp := make([]byte, 1<<16)
				for i := range tmp {
					tmp[i]++
				}
				_ = tmp
				<-done
				wg.Done()
			}()
		}

		// Block in syscall.
		wg.Add(1)
		go func() {
			var tmp [1]byte
			rp.Read(tmp[:])
			<-done
			wg.Done()
		}()

		runtime.GOMAXPROCS(runtime.GOMAXPROCS(1))

		// Test timers.
		timerDone := make(chan bool)
		go func() {
			time.Sleep(time.Millisecond)
			timerDone <- true
		}()
		<-timerDone

		// A bit of network.
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("listen failed: %v", err)
		}
		defer ln.Close()
		go func() {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			time.Sleep(time.Millisecond)
			var buf [1]byte
			c.Write(buf[:])
			c.Close()
		}()
		c, err := net.Dial("tcp", ln.Addr().String())
		if err != nil {
			t.Fatalf("dial failed: %v", err)
		}
		var tmp [1]byte
		c.Read(tmp[:])
		c.Close()

		go func() {
			runtime.Gosched()
			select {}
		}()

		// Unblock helper goroutines and wait them to finish.
		wp.Write(tmp[:])
		wp.Write(tmp[:])
		close(done)
		wg.Wait()
	}()

	for i := 0; i < 3; i++ {
		buf := new(bytes.Buffer)
		if err := StartTrace(buf); err != nil {
			t.Fatalf("failed to start tracing: %v", err)
		}
		time.Sleep(time.Millisecond)
		StopTrace()
		if _, err := trace.Parse(buf); err != nil {
			t.Fatalf("failed to parse trace: %v", err)
		}
	}
	<-outerDone
}

// TestTraceSymbolize tests symbolization and that events has proper stacks.
// In particular that we strip bottom uninteresting frames like goexit,
// top uninteresting frames (runtime guts).
func TestTraceSymbolize(t *testing.T) {
	skipTraceTestsIfNeeded(t)
	if runtime.GOOS == "nacl" {
		t.Skip("skipping: nacl tests fail with 'failed to symbolize trace: failed to start addr2line'")
	}
	buf := new(bytes.Buffer)
	if err := StartTrace(buf); err != nil {
		t.Fatalf("failed to start tracing: %v", err)
	}

	// Now we will do a bunch of things for which we verify stacks later.
	// It is impossible to ensure that a goroutine has actually blocked
	// on a channel, in a select or otherwise. So we kick off goroutines
	// that need to block first in the hope that while we are executing
	// the rest of the test, they will block.
	go func() {
		select {}
	}()
	go func() {
		var c chan int
		c <- 0
	}()
	go func() {
		var c chan int
		<-c
	}()
	done1 := make(chan bool)
	go func() {
		<-done1
	}()
	done2 := make(chan bool)
	go func() {
		done2 <- true
	}()
	c1 := make(chan int)
	c2 := make(chan int)
	go func() {
		select {
		case <-c1:
		case <-c2:
		}
	}()
	var mu sync.Mutex
	mu.Lock()
	go func() {
		mu.Lock()
		mu.Unlock()
	}()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		wg.Wait()
	}()
	cv := sync.NewCond(&sync.Mutex{})
	go func() {
		cv.L.Lock()
		cv.Wait()
		cv.L.Unlock()
	}()
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	go func() {
		c, err := ln.Accept()
		if err != nil {
			t.Fatalf("failed to accept: %v", err)
		}
		c.Close()
	}()

	time.Sleep(time.Millisecond)
	runtime.GC()
	runtime.Gosched()
	time.Sleep(time.Millisecond) // the last chance for the goroutines above to block
	done1 <- true
	<-done2
	c1 <- 0
	mu.Unlock()
	wg.Done()
	cv.Signal()
	c, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("failed to dial: %v", err)
	}
	c.Close()

	StopTrace()
	events, err := trace.Parse(buf)
	if err != nil {
		t.Fatalf("failed to parse trace: %v", err)
	}
	err = trace.Symbolize(events, os.Args[0])
	if err != nil {
		t.Fatalf("failed to symbolize trace: %v", err)
	}

	// Now check that the stacks are correct.
	type frame struct {
		Fn   string
		Line int
	}
	type eventDesc struct {
		Type byte
		Stk  []frame
	}
	want := []eventDesc{
		eventDesc{trace.EvGCStart, []frame{
			frame{"runtime.GC", 0},
			frame{"runtime/pprof_test.TestTraceSymbolize", 0},
			frame{"testing.tRunner", 0},
		}},
		eventDesc{trace.EvGoSched, []frame{
			frame{"runtime/pprof_test.TestTraceSymbolize", 0},
			frame{"testing.tRunner", 0},
		}},
		eventDesc{trace.EvGoCreate, []frame{
			frame{"runtime/pprof_test.TestTraceSymbolize", 0},
			frame{"testing.tRunner", 0},
		}},
		eventDesc{trace.EvGoStop, []frame{
			frame{"runtime.block", 0},
			frame{"runtime/pprof_test.TestTraceSymbolize.func1", 0},
		}},
		eventDesc{trace.EvGoStop, []frame{
			frame{"runtime.chansend1", 0},
			frame{"runtime/pprof_test.TestTraceSymbolize.func2", 0},
		}},
		eventDesc{trace.EvGoStop, []frame{
			frame{"runtime.chanrecv1", 0},
			frame{"runtime/pprof_test.TestTraceSymbolize.func3", 0},
		}},
		eventDesc{trace.EvGoBlockRecv, []frame{
			frame{"runtime.chanrecv1", 0},
			frame{"runtime/pprof_test.TestTraceSymbolize.func4", 0},
		}},
		eventDesc{trace.EvGoBlockSend, []frame{
			frame{"runtime.chansend1", 0},
			frame{"runtime/pprof_test.TestTraceSymbolize.func5", 0},
		}},
		eventDesc{trace.EvGoBlockSelect, []frame{
			frame{"runtime.selectgo", 0},
			frame{"runtime/pprof_test.TestTraceSymbolize.func6", 0},
		}},
		eventDesc{trace.EvGoBlockSync, []frame{
			frame{"sync.(*Mutex).Lock", 0},
			frame{"runtime/pprof_test.TestTraceSymbolize.func7", 0},
		}},
		eventDesc{trace.EvGoBlockSync, []frame{
			frame{"sync.(*WaitGroup).Wait", 0},
			frame{"runtime/pprof_test.TestTraceSymbolize.func8", 0},
		}},
		eventDesc{trace.EvGoBlockCond, []frame{
			frame{"sync.(*Cond).Wait", 0},
			frame{"runtime/pprof_test.TestTraceSymbolize.func9", 0},
		}},
		eventDesc{trace.EvGoBlockNet, []frame{
			frame{"net.(*netFD).accept", 0},
			frame{"net.(*TCPListener).AcceptTCP", 0},
			frame{"net.(*TCPListener).Accept", 0},
			frame{"runtime/pprof_test.TestTraceSymbolize.func10", 0},
		}},
		eventDesc{trace.EvGoSleep, []frame{
			frame{"time.Sleep", 0},
			frame{"runtime/pprof_test.TestTraceSymbolize", 0},
			frame{"testing.tRunner", 0},
		}},
	}
	matched := make([]bool, len(want))
	for _, ev := range events {
	wantLoop:
		for i, w := range want {
			if matched[i] || w.Type != ev.Type || len(w.Stk) != len(ev.Stk) {
				continue
			}

			for fi, f := range ev.Stk {
				wf := w.Stk[fi]
				if wf.Fn != f.Fn || wf.Line != 0 && wf.Line != f.Line {
					continue wantLoop
				}
			}
			matched[i] = true
		}
	}
	for i, m := range matched {
		if m {
			continue
		}
		w := want[i]
		t.Errorf("did not match event %v at %v:%v", trace.EventDescriptions[w.Type].Name, w.Stk[0].Fn, w.Stk[0].Line)
		t.Errorf("seen the following events of this type:")
		for _, ev := range events {
			if ev.Type != w.Type {
				continue
			}
			for _, f := range ev.Stk {
				t.Logf("  %v:%v", f.Fn, f.Line)
			}
			t.Logf("---")
		}
	}
}

package main

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"runtime/pprof"
)

func init() {
	runtime.MemProfileRate = 1
}

var globalNonzero = []byte{1, 2, 3, 4}
var global []byte // = []byte{1, 2, 3, 4}

var globalWithInteriorPointers [3]*int

func main() {
	deferred := make([]byte, 1e6)
	defer func() {
		fmt.Fprintln(io.Discard, deferred)
	}()

	// TODO: tiny allocation?

	m := make([]byte, 1e6) // zero (bss) segment
	n := make([]byte, 1e6) // data segment
	o := make([]byte, 1e6) // kept alive by stack frame
	global = m
	fmt.Printf("global: %p\n", &global)
	globalNonzero = n
	fmt.Fprintln(io.Discard, m)
	runtime.GC()

	fmt.Printf("global with pointer: %p\n", &globalWithInteriorPointers)
	z := new(int)
	globalWithInteriorPointers[1] = z

	ch := make(chan struct{})
	a := new([100000]int)
	runtime.SetFinalizer(a, func(*[100000]int) {
		ch <- struct{}{}
	})
	runtime.GC()
	// TODO: if we only did this first GC, we wouldn't
	// yet see a root for a, which is odd since it ought to be
	// kept alive by the finalizer which hasn't yet returned
	runtime.GC()
	runtime.GC()

	f, _ := os.Create("heap.pprof")
	defer f.Close()
	pprof.Lookup("heap").WriteTo(f, 0)
	runtime.KeepAlive(o)
	fmt.Fprintln(io.Discard, global)
	fmt.Fprintln(io.Discard, globalNonzero)
	fmt.Fprintln(io.Discard, globalWithInteriorPointers)
	<-ch
}

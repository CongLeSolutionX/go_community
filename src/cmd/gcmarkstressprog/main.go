package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"time"
)

func init() {
	runtime.SetMutexProfileFraction(1000)
}

func recursiveMemoryRetainer(data []byte, depth int, c chan struct{}) {
	var payload any
	if err := json.NewDecoder(bytes.NewReader(data)).Decode(&payload); err != nil {
		panic(err)
	}
	if depth == 0 {
		c <- struct{}{}
		<-c
		fmt.Fprintln(io.Discard, payload)
		return
	}
	recursiveMemoryRetainer(data, depth-1, c)
	fmt.Fprintln(io.Discard, payload)
}

func main() {
	prefix := os.Getenv("PREFIX")
	if prefix == "" {
		panic("no prefix")
	}

	cpuprof := prefix + "-cpu.pprof"
	f, err := os.Create(cpuprof)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	data, err := os.ReadFile("code.json")
	if err != nil {
		panic(err)
	}

	var payloads []any
	for i := 0; i < 10; i++ {
		var payload any
		if err := json.NewDecoder(bytes.NewReader(data)).Decode(&payload); err != nil {
			panic(err)
		}
		payloads = append(payloads, payload)
	}

	allocstart := time.Now()
	const numGoroutines = 10
	var chans []chan struct{}
	for i := 0; i < numGoroutines; i++ {
		ch := make(chan struct{})
		go recursiveMemoryRetainer(data, 20, ch)
		chans = append(chans, ch)
	}
	for _, ch := range chans {
		<-ch
	}
	fmt.Printf("time to alloc: %v\n", time.Since(allocstart))

	numGCs := 0
	x := time.Now()
	for time.Since(x) < 4*time.Second {
		runtime.GC()
		numGCs++
	}
	fmt.Printf("gcs/second: %f\n", float64(numGCs)/time.Since(x).Seconds())
	fmt.Printf("gcs: %d\n", numGCs)

	tracepath := prefix + "-trace.out"
	tracefile, err := os.Create(tracepath)
	if err != nil {
		panic(err)
	}
	defer tracefile.Close()
	trace.Start(tracefile)
	runtime.GC()
	trace.Stop()

	// TODO: just keepalive?
	fmt.Fprintln(io.Discard, payloads)
	heapprofile, _ := os.Create(prefix + "-heap.pprof")
	defer heapprofile.Close()
	pprof.Lookup("heap").WriteTo(heapprofile, 0)
	mutexprofile, _ := os.Create(prefix + "-mutex.pprof")
	defer mutexprofile.Close()
	pprof.Lookup("mutex").WriteTo(mutexprofile, 0)
	for _, ch := range chans {
		close(ch)
	}
}

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package test

import (
	"fmt"
	"internal/coverage"
	"internal/coverage/decodecounter"
	"internal/coverage/encodecounter"
	"os"
	"path/filepath"
	"testing"
)

func TestCounterDataFileWriterReader(t *testing.T) {
	flavors := []coverage.CounterFlavor{
		coverage.CtrRaw,
		coverage.CtrULeb128,
	}
	mkfunc := func(p uint32, f uint32, c []uint32) decodecounter.FuncPayload {
		return decodecounter.FuncPayload{
			PkgIdx:   p,
			FuncIdx:  f,
			Counters: c,
		}
	}

	isDead := func(fp decodecounter.FuncPayload) bool {
		for _, v := range fp.Counters {
			if v != 0 {
				return false
			}
		}
		return true
	}

	funcs := []decodecounter.FuncPayload{
		mkfunc(0, 0, []uint32{1, 2, 3}),
		mkfunc(0, 1, []uint32{4, 5}),
		mkfunc(1, 0, []uint32{0, 7, 8, 9, 10, 11, 12, 13, 14}),
	}

	writeVisitor := func(f encodecounter.CounterVisitorFcn) bool {
		for _, fn := range funcs {
			f(fn.PkgIdx, fn.FuncIdx, fn.Counters)
		}
		return true
	}

	for kf, flav := range flavors {

		// Open a counter data file in preparation for emitting data.
		d := t.TempDir()
		cfpath := filepath.Join(d, fmt.Sprintf("covcounters.hash.0.%d", kf))
		of, err := os.OpenFile(cfpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			t.Fatalf("opening covcounters: %v", err)
		}

		// Perform the encode and write.
		args := []string{"arg0", "arg1", "arg_________2"}
		cdfw := encodecounter.NewCoverageDataFileWriter(cfpath, of, args, flav)
		finalHash := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 0}
		if !cdfw.Write(finalHash, writeVisitor) {
			t.Fatalf("counter file Write failed")
		}
		if err := of.Close(); err != nil {
			t.Fatalf("closing covcounters: %v", err)
		}
		cdfw = nil

		// Decode the same file.
		var cdr *decodecounter.CounterDataReader
		inf, err := os.Open(cfpath)
		defer inf.Close()
		if err != nil {
			t.Fatalf("reopening covcounters file: %v", err)
		}
		if cdr, err = decodecounter.NewCounterDataReader(cfpath, inf, inf); err != nil {
			t.Fatalf("opening covcounters for read: %v", err)
		}
		decodedArgs := cdr.Args()
		aWant := fmt.Sprintf("%+v", args)
		aGot := fmt.Sprintf("%+v", decodedArgs)
		if aWant != aGot {
			t.Errorf("reading decoded args, got %s want %s", aGot, aWant)
		}
		for i := range funcs {
			if isDead(funcs[i]) {
				continue
			}
			var fp decodecounter.FuncPayload
			if err := cdr.NextFunc(&fp); err != nil {
				t.Fatalf("reading func %d: %v", i, err)
			}
			got := fmt.Sprintf("%+v", fp)
			want := fmt.Sprintf("%+v", funcs[i])
			if got != want {
				t.Errorf("cdr.NextFunc iter %d got %+v expected %+v", i, got, want)
			}
		}
	}
}

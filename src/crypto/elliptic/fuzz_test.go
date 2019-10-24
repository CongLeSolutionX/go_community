// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package elliptic

import (
	"context"
	"crypto/rand"
	"flag"
	"runtime"
	"sync"
	"testing"
	"time"
)

var stressFlag = flag.Bool("stress", false, "run slow stress tests")

func TestP256Fuzz(t *testing.T) {
	if runtime.GOARCH == "wasm" {
		t.Skip("too slow on wasm")
	}

	// Feed random inputs into the generic and assembly implementations
	// of P256 and compare the results. On platforms without an assembly
	// implementation of P256 the generic implementation will just be
	// run twice.
	fuzz := func(ctx context.Context, wg *sync.WaitGroup) {
		defer wg.Done()

		asm := P256()              // assembly implementation (if available)
		generic := P256().Params() // generic implementation
		for i := 1; ; i++ {
			var scalar1 [32]byte
			if _, err := rand.Read(scalar1[:]); err != nil {
				t.Errorf("error while reading from random source: %v", err)
				return
			}
			x, y := asm.ScalarBaseMult(scalar1[:])
			x2, y2 := generic.ScalarBaseMult(scalar1[:])

			var scalar2 [32]byte
			if _, err := rand.Read(scalar2[:]); err != nil {
				t.Errorf("error while reading from random source: %v", err)
				return
			}
			xx, yy := asm.ScalarMult(x, y, scalar2[:])
			xx2, yy2 := generic.ScalarMult(x2, y2, scalar2[:])

			if x.Cmp(x2) != 0 || y.Cmp(y2) != 0 {
				t.Errorf("ScalarBaseMult does not match reference result with scalar: %x", scalar1)
				t.Log("please report this error to security@golang.org")
				return
			}

			if xx.Cmp(xx2) != 0 || yy.Cmp(yy2) != 0 {
				t.Errorf("ScalarMult does not match reference result with scalars: %x and %x", scalar1, scalar2)
				t.Log("please report this error to security@golang.org")
				return
			}

			// Check at the end so we run at least one iteration
			// before exiting the goroutine.
			select {
			case <-ctx.Done():
				t.Logf("iterations: %v", i)
				return
			default:
			}
		}
	}

	// Run the test for a fixed period of time.
	timeout := 2 * time.Second
	if testing.Short() {
		timeout = 10 * time.Millisecond
	} else if *stressFlag {
		timeout = 1 * time.Minute
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel() // clean up the context

	// Start 1 goroutine per hardware thread to maximize the total iterations.
	wg := new(sync.WaitGroup)
	instances := runtime.GOMAXPROCS(-1)
	for i := 0; i < instances; i++ {
		wg.Add(1)
		go fuzz(ctx, wg)
	}
	wg.Wait()
}

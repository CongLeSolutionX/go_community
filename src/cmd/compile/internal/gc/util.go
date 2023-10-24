// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gc

import (
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	tracepkg "runtime/trace"

	"cmd/compile/internal/base"
)

func startProfile() {
	if base.Flag.CPUProfile != "" {
		f, err := os.Create(base.Flag.CPUProfile)
		if err != nil {
			base.Fatalf("%v", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			base.Fatalf("%v", err)
		}
		base.AtExit(pprof.StopCPUProfile)
	}
	if base.Flag.MemProfile != "" {
		if base.Flag.MemProfileRate != 0 {
			runtime.MemProfileRate = base.Flag.MemProfileRate
		}
		var writeLegacyFormat = 1
		fn := base.Flag.MemProfile
		if fi, statErr := os.Stat(fn); statErr == nil && fi.IsDir() {
			fn = filepath.Join(fn, base.Ctxt.Pkgpath+".mprof")
			os.MkdirAll(filepath.Dir(fn), 0775)
			writeLegacyFormat = 0
		}

		f, err := os.Create(fn)

		if err != nil {
			base.Fatalf("%v", err)
		}
		base.AtExit(func() {
			// Profile all outstanding allocations.
			runtime.GC()
			// compilebench parses the memory profile to extract memstats,
			// which are only written in the legacy pprof format.
			// See golang.org/issue/18641 and runtime/pprof/pprof.go:writeHeap.
			if err := pprof.Lookup("heap").WriteTo(f, writeLegacyFormat); err != nil {
				base.Fatalf("%v", err)
			}
		})
	} else {
		// Not doing memory profiling; disable it entirely.
		runtime.MemProfileRate = 0
	}
	if base.Flag.BlockProfile != "" {
		f, err := os.Create(base.Flag.BlockProfile)
		if err != nil {
			base.Fatalf("%v", err)
		}
		runtime.SetBlockProfileRate(1)
		base.AtExit(func() {
			pprof.Lookup("block").WriteTo(f, 0)
			f.Close()
		})
	}
	if base.Flag.MutexProfile != "" {
		f, err := os.Create(base.Flag.MutexProfile)
		if err != nil {
			base.Fatalf("%v", err)
		}
		runtime.SetMutexProfileFraction(1)
		base.AtExit(func() {
			pprof.Lookup("mutex").WriteTo(f, 0)
			f.Close()
		})
	}
	if base.Flag.TraceProfile != "" {
		f, err := os.Create(base.Flag.TraceProfile)
		if err != nil {
			base.Fatalf("%v", err)
		}
		if err := tracepkg.Start(f); err != nil {
			base.Fatalf("%v", err)
		}
		base.AtExit(tracepkg.Stop)
	}
}

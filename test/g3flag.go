// run

// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test compiling the below listed files (cfiles) in the test directory with the
// -G=3 option and make sure there are no errors or panics. This is to test the
// translation of types2 types to regular types, and the user of these types later
// in the compiler. A few files (rfiles) are fully compiled and run with -G=3.
package main

import (
	"flag"
	"log"
	"os/exec"
)

var cfiles []string = []string{
	"index0.go",
	"gc.go",
	"args.go",
	"escape_hash_maphash.go",
	"func6.go",
	"import.go",
	"inline_math_bits_rotate.go",
	"gc1.go",
	"compos.go",
	"closure4.go",
	"int_lit.go",
	"varinit.go",
	"parentype.go",
	"bom.go",
	"struct0.go",
	"inline_variadic.go",
	"escape_selfassign.go",
	"reflectmethod5.go",
	"defernil.go",
	"reflectmethod6.go",
	"inline_math_bits_rotate.go",
	"reflectmethod4.go",
	"reflectmethod1.go",
	"deferprint.go",
	"func8.go",
	"align.go",
	"initempty.go",
	"char_lit.go",
	"strcopy.go",
	"escape_goto.go",
	"goprint.go",
	"sliceopt.go",
	"crlf.go",
	"const3.go",
	"method3.go",
	"devirt.go",
	"reflectmethod3.go",
	"reflectmethod2.go",
	"linkx.go",
	"env.go",
	"intrinsic_atomic.go",
	"alias1.go",
	"nilptr5_aix.go",
	"nilptr5.go",
	"alg.go",
	"convert.go",
	"for.go",
	"sinit_run.go",
	"gcstring.go",
	"escape_runtime_atomic.go",
	"defer.go",
	"nilptr4.go",
	"escape_sync_atomic.go",
	"sieve.go",
	"inline_literal.go",
	"strength.go",
	"escape_struct_return.go",
	"simassign.go",
	"live_syscall.go",
	"simassign.go",
	"divide.go",
	"recover2.go",
	"nul1.go",
	"initialize.go",
	"notinheap.go",
	"initcomma.go",
	"stackobj.go",
	"uintptrescapes3.go",
	"const4.go",
	"turing.go",
	"typeswitch1.go",
	"decl.go",
	"deferfin.go",
	"escape4.go",
	"if.go",
	"defererrcheck.go",
	"stackobj3.go",
	"func5.go",
	"live1.go",
	"func.go",
	"escape4.go",
	"if.go",
	"indirect.go",
	"defererrcheck.go",
	"stackobj3.go",
	"stringrange.go",
	"mapclear.go",
	"inline_sync.go",
	"rename.go",
	"235.go",
	"print.go",
	"chancap.go",
	"typeswitch1.go",
	"linkx_run.go",
	"phiopt.go",
	"closure.go",
	"stack.go",
	"linkmain_run.go",
	"escape_level.go",
	"slicecap.go",
	"winbatch.go",
	"opt_branchlikely.go",
	"bigalg.go",
	"escape_unsafe.go",
	"recover4.go",
	"uintptrescapes2.go",
	"switch.go",

	// These use untyped constants (so we need code to create extra literal nodes with
	// the real type at each use)
	"rune.go",
	"atomicload.go",
	"gc2.go",
	"init1.go",
	"clearfat.go",
	"tinyfin.go",
	"typeswitch.go",
	"finprofiled.go",
	"chanlinear.go",
	"iota.go",

	// These all have circular or forward references on types
	"escape3.go",
	"live2.go",
	"notinheap3.go",
	"escape_calls.go",
	"func2.go",
	"mallocfin.go",
	"stackobj2.go",
	"peano.go",
}

var rfiles []string = []string{
	"closure1.go",
	"sigchld.go",
	"func7.go",
	"printbig.go",
	"gc1.go",
	"helloworld.go",
}

var verbose = flag.Bool("v", false, "show output for successful compiles")

func main() {
	flag.Parse()
	for _, s := range cfiles {
		out, err := exec.Command("go", "tool", "compile", "-G=3", s).CombinedOutput()
		if err != nil {
			log.Fatalf("go tool compile -G=3 %s: %s\n%s", s, err, out)
		}
		if *verbose {
			log.Printf("go tool compile -G=3 %s:\n%s", s, out)
		}
	}
	for _, s := range rfiles {
		out, err := exec.Command("go", "run", "-gcflags=-G=3", s).CombinedOutput()
		if err != nil {
			log.Fatalf("go run -gcflags=-G=3 %s: %s\n%s", s, err, out)
		}
		if *verbose {
			log.Printf("go run -gcflag=-G=3 %s:\n%s", s, out)
		}
	}
}

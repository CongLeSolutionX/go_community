// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"flag"
	"fmt"
	"internal/buildcfg"
	"log"
	"os"
	"regexp"
	"sort"

	"cmd/asm/internal/arch"
	"cmd/asm/internal/asm"
	"cmd/asm/internal/flags"
	"cmd/asm/internal/lex"

	"cmd/internal/bio"
	"cmd/internal/obj"
	"cmd/internal/objabi"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("asm: ")

	buildcfg.Check()
	GOARCH := buildcfg.GOARCH

	flags.Parse()

	architecture := arch.Set(GOARCH, *flags.Shared || *flags.Dynlink)
	if architecture == nil {
		log.Fatalf("unrecognized architecture %s", GOARCH)
	}
	ctxt := obj.Linknew(architecture.LinkArch)
	ctxt.Debugasm = flags.PrintOut
	ctxt.Debugvlog = flags.DebugV
	ctxt.Flag_dynlink = *flags.Dynlink
	ctxt.Flag_linkshared = *flags.Linkshared
	ctxt.Flag_shared = *flags.Shared || *flags.Dynlink
	ctxt.Flag_maymorestack = flags.DebugFlags.MayMoreStack
	ctxt.Debugpcln = flags.DebugFlags.PCTab
	ctxt.IsAsm = true
	ctxt.Pkgpath = *flags.Importpath
	switch *flags.Spectre {
	default:
		log.Printf("unknown setting -spectre=%s", *flags.Spectre)
		os.Exit(2)
	case "":
		// nothing
	case "index":
		// known to compiler; ignore here so people can use
		// the same list with -gcflags=-spectre=LIST and -asmflags=-spectrre=LIST
	case "all", "ret":
		ctxt.Retpoline = true
	}

	ctxt.Bso = bufio.NewWriter(os.Stdout)
	defer ctxt.Bso.Flush()

	architecture.Init(ctxt)

	// Create object file, write header.
	buf, err := bio.Create(*flags.OutputFile)
	if err != nil {
		log.Fatal(err)
	}
	defer buf.Close()

	if !*flags.SymABIs {
		buf.WriteString(objabi.HeaderString())
		fmt.Fprintf(buf, "!\n")
	}

	// Set macros for GOEXPERIMENTs so we can easily switch
	// runtime assembly code based on them.
	if objabi.LookupPkgSpecial(ctxt.Pkgpath).AllowAsmABI {
		for _, exp := range buildcfg.Experiment.Enabled() {
			flags.D = append(flags.D, "GOEXPERIMENT_"+exp)
		}
	}

	var ok bool
	var failedFile string
	errorMsgs := []string{}
	for _, f := range flag.Args() {
		lexer := lex.NewLexer(f)
		parser := asm.NewParser(ctxt, architecture, lexer)
		ctxt.DiagFunc = func(format string, args ...interface{}) {
			msg := fmt.Sprintf(format, args...)
			errorMsgs = append(errorMsgs, msg)
		}
		ctxt.DiagFlush = func() {
			if len(errorMsgs) == 0 {
				return
			}
			// Try to sort the error messages by file line, then we can remove
			// the identical error message.
			sort.Slice(errorMsgs, func(i, j int) bool {
				fileLine := regexp.MustCompile(`\(.*\.s:[0-9]+\)`)
				fileLine1, fileLine2 := "", ""
				if m := fileLine.FindStringSubmatch(errorMsgs[i]); m != nil {
					fileLine1 = m[0]
				}
				if m := fileLine.FindStringSubmatch(errorMsgs[j]); m != nil {
					fileLine2 = m[0]
				}
				if fileLine1 != fileLine2 {
					return fileLine1 < fileLine2
				}
				return errorMsgs[i] < errorMsgs[j]
			})
			for i, err := range errorMsgs {
				if i == 0 || err != errorMsgs[i-1] {
					log.Print(err)
				}
			}
			errorMsgs = errorMsgs[:0]
			ctxt.Errors = 0
		}
		ctxt.DiagShrink = func(n int) {
			if ctxt.Errors <= n {
				return
			}
			ctxt.Errors = n
			errorMsgs = errorMsgs[:n]
		}
		if *flags.SymABIs {
			ok = parser.ParseSymABIs(buf)
		} else {
			pList := new(obj.Plist)
			pList.Firstpc, ok = parser.Parse()
			// reports errors to parser.Errorf
			if ok {
				obj.Flushplist(ctxt, pList, nil)
			}
		}
		if !ok {
			failedFile = f
			break
		}
	}
	if ok && !*flags.SymABIs {
		ctxt.NumberSyms()
		obj.WriteObjFile(ctxt, buf)
	}
	if !ok || ctxt.Errors > 0 {
		ctxt.DiagFlush()
		if failedFile != "" {
			log.Printf("assembly of %s failed", failedFile)
		} else {
			log.Print("assembly failed")
		}
		buf.Close()
		os.Remove(*flags.OutputFile)
		os.Exit(1)
	}
}

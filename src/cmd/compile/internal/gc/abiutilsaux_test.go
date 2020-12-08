// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gc

// This file contains utility routines and harness infrastructure used
// by the ABI tests in "abiutils_test.go".

import (
	"cmd/compile/internal/ir"
	"cmd/compile/internal/types"
	"cmd/internal/src"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"
	"text/scanner"
)

func mkParamResultField(t *types.Type, s *types.Sym, which ir.Class) *types.Field {
	field := types.NewField(src.NoXPos, s, t)
	n := NewName(s)
	n.SetClass(which)
	field.Nname = n
	n.SetType(t)
	return field
}

func mkFuncType(rcvr *types.Type, ins []*types.Type, outs []*types.Type) *types.Type {
	q := lookup("?")
	inf := []*types.Field{}
	for _, it := range ins {
		inf = append(inf, mkParamResultField(it, q, ir.PPARAM))
	}
	outf := []*types.Field{}
	for _, ot := range outs {
		outf = append(outf, mkParamResultField(ot, q, ir.PPARAMOUT))
	}
	var rf *types.Field
	if rcvr != nil {
		rf = mkParamResultField(rcvr, q, ir.PPARAM)
	}
	return types.NewSignature(types.LocalPkg, rf, inf, outf)
}

type expectedDump struct {
	dump string
	file string
	line int
}

func tokenize(src string) []string {
	var s scanner.Scanner
	s.Init(strings.NewReader(src))
	res := []string{}
	for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
		res = append(res, s.TokenText())
	}
	return res
}

func verifyParamResultOffset(t *testing.T, f *types.Field, r ABIParamAssignment, which string, idx int) int {
	n := ir.AsNode(f.Nname)
	if n == nil {
		panic("not expected")
	}
	if n.Offset() != int64(r.Offset) {
		t.Errorf("%s %d: got offset %d wanted %d t=%v",
			which, idx, r.Offset, n.Offset(), f.Type)
		return 1
	}
	return 0
}

func makeExpectedDump(e string) expectedDump {
	var pcs [24]uintptr
	runtime.Callers(0, pcs[:])
	frames := runtime.CallersFrames(pcs[:])
	for {
		frame, more := frames.Next()
		if strings.Contains(frame.Function, "TestABIUtils") {
			return expectedDump{
				dump: e,
				file: frame.File,
				line: frame.Line,
			}
		}
		if !more {
			break
		}
	}
	return expectedDump{dump: e}
}

func difftokens(atoks []string, etoks []string) string {
	if len(atoks) != len(etoks) {
		return fmt.Sprintf("expected %d tokens got %d",
			len(etoks), len(atoks))
	}
	for i := 0; i < len(etoks); i++ {
		if etoks[i] == atoks[i] {
			continue
		}

		return fmt.Sprintf("diff at token %d: expected %q got %q",
			i, etoks[i], atoks[i])
	}
	return ""
}

var dumpFileCount int
var remOnce sync.Once
var remasterScriptFile *os.File

func dumpFileName(tag string, count int) string {
	return fmt.Sprintf("/tmp/%s.%d.txt", tag, count)
}

func emitStringToDumpFile(tag string, count int, dump string) {
	fn := dumpFileName(tag, count)
	fl := os.O_RDWR | os.O_CREATE | os.O_TRUNC
	of, err := os.OpenFile(fn, fl, 0644)
	if err != nil {
		panic("unable to emit " + fn)
	}
	fmt.Fprintf(of, "%s\n", dump)
	fmt.Fprintf(os.Stderr, "... emitted dump file %s\n", fn)
	if cerr := of.Close(); cerr != nil {
		panic("error closing dump file" + fn)
	}
}

func complain(t *testing.T, reason string, exp expectedDump, actual string) {
	fmt.Fprintf(os.Stderr, "expected:\n%s\n", strings.TrimSpace(exp.dump))
	fmt.Fprintf(os.Stderr, "got:\n%s\n", actual)
	t.Errorf("failure reason: %s\n", reason)

	emitRemasterScript := os.Getenv("ABITEST_EMIT_REMASTER_SCRIPT") != ""
	emitDumpFiles := emitRemasterScript || os.Getenv("ABITEST_EMIT_DUMP_FILES") != ""

	if emitDumpFiles {
		emitStringToDumpFile("expected", dumpFileCount, exp.dump)
		emitStringToDumpFile("actual", dumpFileCount, actual)
		if emitRemasterScript {
			remOnce.Do(func() {
				fn := "/tmp/remaster-inputs.txt"
				fl := os.O_RDWR | os.O_CREATE | os.O_TRUNC
				if f, err := os.OpenFile(fn, fl, 0644); err != nil {
					panic("unable to open " + fn)
				} else {
					remasterScriptFile = f
					fmt.Fprintf(os.Stderr, "emitting remaster script to file %q\n", fn)
					fmt.Fprintf(f, "# emitted by tests in abiutils_test.go\n")
					fmt.Fprintf(f, "# for use with remaster-tests.go\n")
					remasterScriptFile.Sync()
				}
			})
			fmt.Fprintf(remasterScriptFile, "%s %d %s %s\n", exp.file, exp.line,
				dumpFileName("expected", dumpFileCount),
				dumpFileName("actual", dumpFileCount))
			remasterScriptFile.Sync()
		}
		dumpFileCount++
	}
}

func abitest(t *testing.T, ft *types.Type, exp expectedDump) {

	dowidth(ft)

	// Analyze with full set of registers.
	regRes := ABIAnalyze(ft, configAMD64)
	regResString := strings.TrimSpace(regRes.String())

	// Check results.
	reason := difftokens(tokenize(regResString), tokenize(exp.dump))
	if reason != "" {
		complain(t, reason, exp, regResString)
	}

	// Analyze again with empty register set.
	empty := ABIConfig{}
	emptyRes := ABIAnalyze(ft, empty)
	emptyResString := emptyRes.String()

	// Walk the results and make sure the offsets assigned match
	// up with those assiged by dowidth. This checks to make sure that
	// when we have no available registers the ABI assignment degenerates
	// back to the original ABI0.

	// receiver
	failed := 0
	rfsl := ft.Recvs().Fields().Slice()
	poff := 0
	if len(rfsl) != 0 {
		failed |= verifyParamResultOffset(t, rfsl[0], emptyRes.inparams[0], "receiver", 0)
		poff = 1
	}
	// params
	pfsl := ft.Params().Fields().Slice()
	for k, f := range pfsl {
		verifyParamResultOffset(t, f, emptyRes.inparams[k+poff], "param", k)
	}
	// results
	ofsl := ft.Results().Fields().Slice()
	for k, f := range ofsl {
		failed |= verifyParamResultOffset(t, f, emptyRes.outparams[k], "result", k)
	}

	if failed != 0 {
		fmt.Fprintf(os.Stderr, "emptyres:\n%s\n", emptyResString)
	}
}

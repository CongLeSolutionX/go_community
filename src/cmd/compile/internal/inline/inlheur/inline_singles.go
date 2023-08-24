// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package inlheur

import (
	"cmd/compile/internal/ir"
	"cmd/compile/internal/types"
	"fmt"
	"os"
	"sort"
)

type siteAndCount struct {
	cs    *CallSite
	count int
}

func InlineSingleCallFuncs(isBigFunc func(*ir.Func) bool) {
	// NB: the call site tab is a pre-inlining view of the world
	csCount := make(map[*ir.Func]siteAndCount)
	cstab := CallSiteTable()
	for _, cs := range cstab {
		if cs.Callee.Sym().Pkg != types.LocalPkg {
			continue
		}
		// no methods for now
		if ir.IsMethod(cs.Callee) {
			continue
		}
		// skip callee marked "noinline"
		if cs.Callee.Pragma&ir.Noinline != 0 {
			continue
		}
		// remove anything that looks like it will be inlined
		if cs.Score <= 80 && cs.Callee.Inl != nil {
			continue
		}
		if sac, ok := csCount[cs.Callee]; ok {
			sac.count++
			csCount[cs.Callee] = sac
			continue
		}
		csCount[cs.Callee] = siteAndCount{
			cs:    cs,
			count: 1,
		}
	}
	cands := []*CallSite{}
	for _, sac := range csCount {
		if sac.count != 1 {
			continue
		}
		if sac.cs.Callee.Nname.Addrtaken() {
			continue
		}
		if types.IsExported(sac.cs.Callee.Sym().Name) {
			continue
		}
		if sac.cs.Caller.Inl != nil {
			continue
		}
		//fmt.Fprintf(os.Stderr, "=-= %+v\n", sac.cs.Caller)
		if isBigFunc(sac.cs.Caller) {
			continue
		}
		cands = append(cands, sac.cs)
	}
	encands := []string{}
	for _, cs := range cands {
		encands = append(encands, EncodeCallSiteKey(cs)+" "+fmt.Sprintf("%v", cs.Callee))
	}
	sort.Strings(encands)

	if len(encands) != 0 {
		fmt.Fprintf(os.Stderr, ">> single call func candidates for this package:\n")
		for _, ecs := range encands {
			fmt.Fprintf(os.Stderr, "-- %s\n", ecs)
		}
	}
}

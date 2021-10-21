// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cov

import (
	"fmt"
	"internal/coverage"
	"os"
	"path/filepath"
	"regexp"
	"sort"
)

// A "pod" is a set of files emitted during the executions of a
// coverage-instrumented binary. Each pod contains a single meta-data
// file, and then 0 or more counter data files that refer to that
// meta-data file. This file

// Pod holds encapsulates info on the names of the file(s) that hold
// the results of a set of runs of a given coverage-instrumented
// binary.
type Pod struct {
	MetaFile         string
	CounterDataFiles []string
}

// CollectPods visits the files contained within the directory
// 'dirpath' and returns the set of pods it contains, or an error if
// something went wrong during directory/file reading. CollectPods
// skips over any file in the directory that is not related to
// coverage (e.g. skips things that are not meta-data files or
// counter-data files). CollectPods also skips over 'orphaned' counter
// data files (e.g. counter data files for which we can't find the
// corresponding meta-data file). If "warn" is true, CollectPods will
// issue warnings to stderr when it encounters non-fatal problems (for
// orphans or a directory with no meta-data files).
func CollectPods(dirpath string, warn bool) ([]Pod, error) {
	files := []string{}
	dents, err := os.ReadDir(dirpath)
	if err != nil {
		return nil, err
	}
	for _, e := range dents {
		if e.IsDir() {
			continue
		}
		files = append(files, filepath.Join(dirpath, e.Name()))
	}
	return CollectPodsFromFiles(files, warn), nil
}

// CollectPodsFromFiles functions the same as "CollectPods" but
// operates on an explicit list of files instead of a directory.
func CollectPodsFromFiles(files []string, warn bool) []Pod {
	metaRE := regexp.MustCompile(fmt.Sprintf(`^%s\.(\S+)$`, coverage.MetaFilePref))
	mm := make(map[string]Pod)
	for _, f := range files {
		base := filepath.Base(f)
		if m := metaRE.FindStringSubmatch(base); m != nil {
			tag := m[1]
			mm[tag] = Pod{MetaFile: f}
		}
	}
	counterRE := regexp.MustCompile(fmt.Sprintf(`^%s\.(\S+)\.\d+$`, coverage.CounterFilePref))
	for _, f := range files {
		base := filepath.Base(f)
		if m := counterRE.FindStringSubmatch(base); m != nil {
			tag := m[1]
			fmt.Fprintf(os.Stderr, "=-=   counterdata matched %s tag %s\n", base, tag)
			if v, ok := mm[tag]; ok {
				v.CounterDataFiles = append(v.CounterDataFiles, f)
				mm[tag] = v
			} else {
				if warn {
					warning("skipping orphaned counter file: %s", f)
				}
			}
		}
	}
	if len(mm) == 0 {
		if warn {
			warning("no coverage data files found")
		}
		return nil
	}
	pods := make([]Pod, 0, len(mm))
	for _, p := range mm {
		sort.Slice(p.CounterDataFiles, func(i, j int) bool {
			return p.CounterDataFiles[i] < p.CounterDataFiles[j]
		})
		pods = append(pods, p)
	}
	sort.Slice(pods, func(i, j int) bool {
		return pods[i].MetaFile < pods[j].MetaFile
	})
	return pods
}

func warning(s string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "warning: ")
	fmt.Fprintf(os.Stderr, s, a...)
	fmt.Fprintf(os.Stderr, "\n")
}

// Copyright 2022 The Go Authors. All rights reserved.
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
	"strconv"
)

// Pod encapsulates a set of files emitted during the executions of a
// coverage-instrumented binary. Each pod contains a single meta-data
// file, and then 0 or more counter data files that refer to that
// meta-data file. Pods are intended to simplify processing of
// coverage output files in the case where we have several coverage
// output directories containing output files derived from more
// than one instrumented executable. In the case where the files that
// make up a pod are spread out across multiple directories, each
// element of the "Origins" field below will be populated with the
// index of the originating directory for the corresponding counter
// data file (within the slice of input dirs handed to CollectPods).
// The ProcessIDs field will be populated with the process ID of each
// data file in the CounterDataFiles slice.
type Pod struct {
	MetaFile         string
	CounterDataFiles []string
	Origins          []int
	ProcessIDs       []int
}

// CollectPods visits the files contained within the directories in
// the list 'dirs' and returns the set of pods they contains, or an
// error if something went wrong during directory/file reading.
// CollectPods skips over any file that is not
// related to coverage (e.g. skips things that are not meta-data files
// or counter-data files). CollectPods also skips over 'orphaned'
// counter data files (e.g. counter data files for which we can't find
// the corresponding meta-data file). If "warn" is true, CollectPods
// will issue warnings to stderr when it encounters non-fatal problems
// (for orphans or a directory with no meta-data files).
func CollectPods(dirs []string, warn bool) ([]Pod, error) {
	files := []string{}
	dirIndices := []int{}
	for k, dir := range dirs {
		dents, err := os.ReadDir(dir)
		if err != nil {
			return nil, err
		}
		for _, e := range dents {
			if e.IsDir() {
				continue
			}
			files = append(files, filepath.Join(dir, e.Name()))
			dirIndices = append(dirIndices, k)
		}
	}
	return collectPodsImpl(files, dirIndices, warn), nil
}

// CollectPodsFromFiles functions the same as "CollectPods" but
// operates on an explicit list of files instead of a directory.
func CollectPodsFromFiles(files []string, warn bool) []Pod {
	return collectPodsImpl(files, nil, warn)
}

type fileWithAnnotations struct {
	file   string
	origin int
	pid    int
}

type protoPod struct {
	mf       string
	elements []fileWithAnnotations
}

func collectPodsImpl(files []string, dirIndices []int, warn bool) []Pod {
	metaRE := regexp.MustCompile(fmt.Sprintf(`^%s\.(\S+)$`, coverage.MetaFilePref))
	mm := make(map[string]protoPod)
	for _, f := range files {
		base := filepath.Base(f)
		if m := metaRE.FindStringSubmatch(base); m != nil {
			tag := m[1]
			// We need to allow for the possibility of duplicates
			// meta-data files. If we hit this case, use the
			// first encountered as the canonical version.
			if _, ok := mm[tag]; !ok {
				mm[tag] = protoPod{mf: f}
			}
			// FIXME: should probably check file length and hash here for
			// the duplicate.
		}
	}
	counterRE := regexp.MustCompile(fmt.Sprintf(coverage.CounterFileRegexp, coverage.CounterFilePref))
	for k, f := range files {
		base := filepath.Base(f)
		if m := counterRE.FindStringSubmatch(base); m != nil {
			tag := m[1] // meta hash
			pid, err := strconv.Atoi(m[2])
			if err != nil {
				continue
			}
			if v, ok := mm[tag]; ok {
				idx := -1
				if dirIndices != nil {
					idx = dirIndices[k]
				}
				fo := fileWithAnnotations{file: f, origin: idx, pid: pid}
				v.elements = append(v.elements, fo)
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
		sort.Slice(p.elements, func(i, j int) bool {
			return p.elements[i].file < p.elements[j].file
		})
		pod := Pod{
			MetaFile:         p.mf,
			CounterDataFiles: make([]string, 0, len(p.elements)),
			Origins:          make([]int, 0, len(p.elements)),
			ProcessIDs:       make([]int, 0, len(p.elements)),
		}
		for _, e := range p.elements {
			pod.CounterDataFiles = append(pod.CounterDataFiles, e.file)
			pod.Origins = append(pod.Origins, e.origin)
			pod.ProcessIDs = append(pod.ProcessIDs, e.pid)
		}
		pods = append(pods, pod)
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

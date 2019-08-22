// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package symbolize

import (
	"debug/dwarf"
	"io"
	"sort"

	"lockcheck/cache"
)

// LineTab stores the function and line table for a DWARF object.
//
// It is safe to access concurrently.
type LineTab struct {
	dd         *dwarf.Data
	funcRanges []funcRange
	names      map[string]*dwarf.Entry

	cuCache cache.Cache // map[*dwarf.Entry]{*cuCache|error}
}

type LineInfo struct {
	Func *dwarf.Entry
	File *dwarf.LineFile
	Line int
}

type funcRange struct {
	pcs     [2]uint64
	fun, cu *dwarf.Entry
}

// cuCache records the decoded line table for a CU.
type cuCache struct {
	// seqs is the set of lines sequences in this CU, sorted by
	// starting PC.
	seqs [][]dwarf.LineEntry

	// files is the file table from this CU's line table.
	files []*dwarf.LineFile
}

// NewLineTab loads the source line table from DWARF data dd.
func NewLineTab(dd *dwarf.Data) (*LineTab, error) {
	// Index the PC ranges of functions.
	r := dd.Reader()
	var cu *dwarf.Entry
	var funcRanges []funcRange
	names := make(map[string]*dwarf.Entry)
	for {
		ent, err := r.Next()
		if err != nil {
			return nil, err
		}
		if ent == nil {
			break
		}

		switch ent.Tag {
		default:
			r.SkipChildren()

		case dwarf.TagCompileUnit:
			cu = ent

		case dwarf.TagModule, dwarf.TagNamespace:
			// Descend into these.

		case dwarf.TagSubprogram:
			name, ok := ent.Val(dwarf.AttrName).(string)
			if ok {
				names[name] = ent
			}
			ranges, err := dd.Ranges(ent)
			if err != nil {
				return nil, err
			}
			for _, r := range ranges {
				funcRanges = append(funcRanges, funcRange{r, ent, cu})
			}
			r.SkipChildren()
		}
	}
	sort.Slice(funcRanges, func(i, j int) bool {
		return funcRanges[i].pcs[0] < funcRanges[j].pcs[0]
	})

	l := &LineTab{dd: dd, funcRanges: funcRanges, names: names}
	l.cuCache.New = l.loadCU
	return l, nil
}

// FuncByName returns the DIE for the named function, or nil if no
// such function exists.
func (l *LineTab) FuncByName(name string) *dwarf.Entry {
	return l.names[name]
}

// decodeCU decodes the line table information for cu.
func (l *LineTab) decodeCU(cu *dwarf.Entry) (cuCache, error) {
	val := l.cuCache.Get(cu)
	switch val := val.(type) {
	case *cuCache:
		return *val, nil
	case error:
		return cuCache{}, val
	default:
		panic("bad type")
	}
}

// loadCU is the cache populator for decodeCU. key is a *dwarf.Entry
// of a TagCompileUnit. It returns either a *cuCache or an error.
func (l *LineTab) loadCU(key interface{}) interface{} {
	cu := key.(*dwarf.Entry)

	// Decode line table.
	lr, err := l.dd.LineReader(cu)
	if err != nil {
		return err
	}
	var seq []dwarf.LineEntry
	var entry dwarf.LineEntry
	for {
		err := lr.Next(&entry)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		seq = append(seq, entry)
	}

	// Split the table into sequences.
	var seqs [][]dwarf.LineEntry
	start := 0
	for i := range seq {
		if seq[i].EndSequence {
			seqs = append(seqs, seq[start:i+1])
			start = i + 1
		}
	}
	if start != len(seq) {
		seqs = append(seqs, seq[start:])
	}

	// Sort sequences by start address.
	sort.Slice(seqs, func(i, j int) bool {
		return seqs[i][0].Address < seqs[j][0].Address
	})

	return &cuCache{seqs, lr.Files()}
}

// Lookup returns the symbolic line information for pc. If no
// functions contain pc, it returns nil, nil. If pc is part of an
// inlined stack of functions, it returns multiple LineInfos, starting
// from the inner-most frame.
func (l *LineTab) Lookup(pc uint64) ([]LineInfo, error) {
	// Find the function containing PC.
	i := sort.Search(len(l.funcRanges), func(i int) bool {
		return pc < l.funcRanges[i].pcs[0]
	}) - 1
	if i < 0 || pc > l.funcRanges[i].pcs[1] {
		return nil, nil
	}
	fr := l.funcRanges[i]

	// Get decoded line table sequences.
	cu, err := l.decodeCU(fr.cu)
	if err != nil {
		return nil, err
	}

	// Find the right line table sequence.
	i = sort.Search(len(cu.seqs), func(i int) bool {
		return pc < cu.seqs[i][0].Address
	}) - 1
	if i < 0 {
		return nil, nil
	}
	seq := cu.seqs[i]
	// Find the entry within the sequence.
	i = sort.Search(len(seq), func(i int) bool {
		return pc < seq[i].Address
	}) - 1
	lineEntry := seq[i]
	if lineEntry.EndSequence {
		return nil, nil
	}

	// Walk the function DIE to find inlines at this PC.
	// Build this stack starting from the outer-most frame.
	//
	// TODO: We could cache the decoded inline tree for the
	// function, but this doesn't seem to cost much.
	var stack []LineInfo
	innerFunc := fr.fun
	r := l.dd.Reader()
	r.Seek(fr.fun.Offset)
	depth := 0
	for {
		ent, err := r.Next()
		if err != nil {
			return nil, err
		}
		if ent == nil {
			break
		}

		if ent.Tag == dwarf.TagInlinedSubroutine {
			ranges, err := l.dd.Ranges(ent)
			if err != nil {
				return nil, err
			}
			if rangesContains(ranges, pc) {
				var callFile *dwarf.LineFile
				callFileN, ok := ent.Val(dwarf.AttrCallFile).(int64)
				if ok {
					callFile = cu.files[callFileN]
				}
				callLine, _ := ent.Val(dwarf.AttrCallLine).(int64)
				stack = append(stack, LineInfo{innerFunc, callFile, int(callLine)})
				innerFunc = ent
			}
		}

		if ent.Tag == 0 {
			depth--
		} else if ent.Children {
			depth++
		}
		if depth <= 0 {
			// We're back at the level of the original DIE.
			break
		}
	}

	// Finish up the inner-most frame in the stack.
	stack = append(stack, LineInfo{innerFunc, lineEntry.File, lineEntry.Line})

	// Reverse the stack.
	for i, j := 0, len(stack)-1; i < j; i, j = i+1, j-1 {
		stack[i], stack[j] = stack[j], stack[i]
	}

	return stack, nil
}

func rangesContains(ranges [][2]uint64, x uint64) bool {
	for _, r := range ranges {
		if r[0] <= x && x < r[1] {
			return true
		}
	}
	return false
}

// FuncName returns the name of a function given its TagSubprogram or
// TagInlinedSubroutine DIE.
func (l *LineTab) FuncName(ent *dwarf.Entry) string {
	name, ok := ent.Val(dwarf.AttrName).(string)
	if ok {
		return name
	}

	r := l.dd.Reader()
	for {
		origin, ok := ent.Val(dwarf.AttrAbstractOrigin).(dwarf.Offset)
		if !ok {
			return ""
		}
		r.Seek(origin)
		var err error
		ent, err = r.Next()
		if err != nil {
			return ""
		}

		name, ok := ent.Val(dwarf.AttrName).(string)
		if ok {
			return name
		}
	}
}

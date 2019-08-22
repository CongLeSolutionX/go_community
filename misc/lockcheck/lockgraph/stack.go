// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lockgraph

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"sync"
)

// A Stack is a symbolized call stack. The representation is compact,
// but must be referenced against a StackTable.
//
// Stack can be serialized to and from JSON.
type Stack []SrcPos

func (s Stack) Hash() uint64 {
	var h uint64
	for _, pos := range s {
		h += uint64(pos.File)
		h += h << 10
		h ^= h >> 6
		h += uint64(pos.Func)
		h += h << 10
		h ^= h >> 6
		h += uint64(pos.Line)
		h += h << 10
		h ^= h >> 6
	}
	h += h << 3
	h ^= h >> 11
	return h
}

func (s Stack) Equals(s2 Stack) bool {
	if len(s) != len(s2) {
		return false
	}
	for i := range s {
		if s[i] != s2[i] {
			return false
		}
	}
	return true
}

// SrcPos is a source position.
//
// SrcPos can be serialized to and from JSON.
type SrcPos struct {
	File  int // Index in StackTable.Files
	Func  int // Index in StackTable.Funcs
	Line  int
	Error string // If non-zero, an error string
}

func (p *SrcPos) MarshalJSON() ([]byte, error) {
	if p.Error != "" {
		return json.Marshal(p.Error)
	}
	return json.Marshal([3]int{p.File, p.Func, p.Line})
}

func (p *SrcPos) UnmarshalJSON(data []byte) error {
	*p = SrcPos{}
	if data[0] == '"' {
		return json.Unmarshal(data, &p.Error)
	}
	var parts [3]int
	if err := json.Unmarshal(data, &parts); err != nil {
		log.Print(string(data))
		panic(err)
		return err
	}
	p.File, p.Func, p.Line = parts[0], parts[1], parts[2]
	return nil
}

// A StackTable compactly stores the data to symbolize a set of
// Stacks.
//
// StackTable is safe to access concurrently.
//
// StackTable can be serialized to and from JSON.
type StackTable struct {
	Files, Funcs StringTable
}

func (st *StackTable) StringStack(stack Stack) string {
	var buf bytes.Buffer
	for i, pos := range stack {
		if i > 0 {
			buf.WriteByte('\n')
		}
		if pos.Error != "" {
			buf.WriteString(pos.Error)
			continue
		}
		fmt.Fprintf(&buf, "%s %s:%d", st.Funcs.Get(pos.Func), st.Funcs.Get(pos.File), pos.Line)
	}
	return buf.String()
}

// A StringTable is a table of interned strings.
//
// StringTable is safe to access concurrently.
//
// StringTable can be serialized to and from JSON.
type StringTable struct {
	strings []string
	index   map[string]int
	lock    sync.RWMutex
}

// Add adds str to the string table and returns its compact index.
func (st *StringTable) Add(str string) int {
	st.lock.RLock()
	idx, ok := st.index[str]
	st.lock.RUnlock()
	if !ok {
		st.lock.Lock()
		if st.index == nil {
			// Lazily populate the index.
			st.index = make(map[string]int)
			for i, str := range st.strings {
				st.index[str] = i
			}
		}
		// Check again
		idx, ok = st.index[str]
		if !ok {
			idx = len(st.strings)
			st.index[str] = idx
			st.strings = append(st.strings, str)
		}
		st.lock.Unlock()
	}
	return idx
}

// Get returns the string by index i.
func (st *StringTable) Get(i int) string {
	st.lock.RLock()
	defer st.lock.RUnlock()
	return st.strings[i]
}

func (st *StringTable) MarshalJSON() ([]byte, error) {
	st.lock.RLock()
	defer st.lock.RUnlock()
	return json.Marshal(st.strings)
}

func (st *StringTable) UnmarshalJSON(data []byte) error {
	st.lock.Lock()
	defer st.lock.Unlock()
	if err := json.Unmarshal(data, &st.strings); err != nil {
		return err
	}
	st.index = nil
	return nil
}

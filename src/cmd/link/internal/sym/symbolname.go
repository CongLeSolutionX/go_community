// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sym

import (
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
)

// SymName is a token representing the name string for a Go linker
// symbol. Many Go symbol names are of the form X.Y where "X" is a
// token of some sort (ex: "type", or packagepath) and Y is a name
// (function name or type name). Since many symbols share the same
// token, we store the X and Y separately.
//
// Each component of the name (pref and suf) is an index into a
// SymNameTable's "allnames" array. The string "suf" is guaranteed
// to not contain a ".", whereas "pref" is either empty or ends
// with a "." character.

type SymName struct {
	pref uint64
	suf  uint64
}

func (s *SymName) Equal(o SymName) bool {
	return s.pref == o.pref && s.suf == o.suf
}

type nameHasher interface {
	Sum64(input []byte) uint64
}

type defaultHasher struct {
}

const (
	offset64 = 14695981039346656037
	prime64  = 1099511628211
)

func (dh defaultHasher) Sum64(input []byte) uint64 {
	var hash uint64 = offset64
	for idx := 0; idx < len(input); idx++ {
		hash *= prime64
		hash ^= uint64(input[idx])
	}
	return hash
}

// SymNameTable is a container/manager for symbol names. The chief
// entry point is the Lookup() method, which looks up a string and
// returns a SymName token for that name (taking advantage of string
// interning).
//
// Given a string S that is a component of a SymbolName (see above),
// we hash the string and look up the result in a map, which leads us
// to a slot in the 'allnames' slice. In the absence of any collisions,
// the value for each map key is a simple slot in the 'allnames' array.
// If two strings SA and SB hash to the same hashcode, then the
// value for each is a slot in the collisions array (indicated by
// setting the high bit on the value).
//
type SymNameTable struct {
	hash       map[uint64]uint64
	allnames   []string
	collisions [][]uint64
	hasher     nameHasher
	locked     bool
}

// NewSymNameTable creates and initializes a new name table.
func NewSymNameTable(sizeHint int) SymNameTable {
	h := make(map[uint64]uint64, sizeHint)
	a := make([]string, 0, sizeHint)
	f := defaultHasher{}
	c := make([][]uint64, 0, 16)
	ret := SymNameTable{hash: h, allnames: a, hasher: f, collisions: c}
	ret.install("")
	return ret
}

func (snt *SymNameTable) Lock() {
	if snt.locked {
		panic("SymNameTable double lock")
	}
	snt.locked = true
}

func (snt *SymNameTable) Unlock() {
	if !snt.locked {
		panic("SymNameTable not locked")
	}
	snt.locked = false
}

func (snt *SymNameTable) String(sn SymName) string {
	return snt.allnames[sn.pref] + snt.allnames[sn.suf]
}

func (snt *SymNameTable) addString(s string) uint64 {
	idx := uint64(len(snt.allnames))
	if false {
		if idx == math.MaxUint32 {
			panic("SymNameTable overflow (more than 2^32 entries)")
		}
	}
	snt.allnames = append(snt.allnames, s)
	return idx
}

func isCollision(v uint64) bool {
	return (v & uint64(1<<63)) != 0
}

func getCollisionSlot(v uint64) uint64 {
	return (v << 1) >> 1
}

func makeCollisionToken(v uint64) uint64 {
	return v | uint64(1<<63)
}

func (snt *SymNameTable) hashcode(s string) uint64 {
	return snt.hasher.Sum64([]byte(s))
}

func (snt *SymNameTable) install(s string) uint64 {
	hc := snt.hasher.Sum64([]byte(s))
	if val, ok := snt.hash[hc]; ok {
		if isCollision(val) || s != snt.allnames[val] {
			// handle collisions
			return snt.handleCollision(s, hc, val)
		} else {
			return val
		}
	}
	idx := snt.addString(s)
	if snt.locked {
		log.Printf("write during lock for string '%s' id %d hc %d\n", s, idx, hc)
		panic("write during lock")
	}
	snt.hash[hc] = uint64(idx)
	return idx
}

func (snt *SymNameTable) handleCollision(s string, hc uint64, val uint64) uint64 {
	var colslot uint64

	// Is this a collision that was encountered previously?
	if isCollision(val) {
		colslot = getCollisionSlot(val)
		for _, slot := range snt.collisions[colslot] {
			if s == snt.allnames[slot] {
				return slot
			}
		}
	} else {
		// new collision -- add entry to collisions slice and val for hash code
		colslot = uint64(len(snt.collisions))
		snt.collisions = append(snt.collisions, []uint64{val})
		if snt.locked {
			panic("write during lock")
		}
		snt.hash[hc] = makeCollisionToken(colslot)
	}

	// string has not been seen before -- insert into allnames
	idx := snt.addString(s)

	// update existing collision entry
	snt.collisions[colslot] = append(snt.collisions[colslot], idx)

	return idx
}

func (snt *SymNameTable) Lookup(s string) SymName {
	pref := ""
	suf := s
	doti := strings.LastIndex(s, ".")
	if doti != -1 {
		pref = s[:doti+1]
		suf = s[doti+1:]
	}
	prefidx := snt.install(pref)
	sufidx := snt.install(suf)
	return SymName{pref: prefidx, suf: sufidx}
}

func (snt *SymNameTable) HasPrefix(sn SymName, pref string) bool {
	snp := snt.allnames[sn.pref]

	// 'pref' smaller than 'sn.pref'? check just against 'sn.pref'
	pl := len(pref)
	snpl := len(snp)
	if pl < snpl {
		return snp[0:pl] == pref
	}

	// check to make sure first part of pref matches sn.pref
	if snp != pref[0:snpl] {
		return false
	}
	if pl == snpl {
		return true
	}

	// check remainder
	sns := snt.allnames[sn.suf]
	return strings.HasPrefix(sns, pref[snpl:])
}

func (snt *SymNameTable) HasSuffix(sn SymName, suf string) bool {

	sns := snt.allnames[sn.suf]

	// 'suf' smaller than 'sn.suf'? check just against 'sn.suf'
	sl := len(suf)
	snsl := len(sns)
	if sl < snsl {
		return sns[snsl-sl:] == suf
	}

	// check to make sure second part of suf matches all of sn.suf
	if sns != suf[sl-snsl:] {
		return false
	}
	if sl == snsl {
		return true
	}

	// check remainder
	snp := snt.allnames[sn.pref]
	return strings.HasSuffix(snp, suf[:sl-snsl])
}

func (snt *SymNameTable) NameEqString(sn SymName, s string) bool {
	snp := snt.allnames[sn.pref]
	sns := snt.allnames[sn.suf]
	if len(s) != len(snp)+len(sns) {
		return false
	}
	return s[0:len(snp)] == snp && s[len(snp):len(snp)+len(sns)] == sns
}

type blob struct {
	hc  uint64
	val uint64
}
type blobs []blob

func (bl blobs) Len() int           { return len(bl) }
func (bl blobs) Swap(i, j int)      { bl[i], bl[j] = bl[j], bl[i] }
func (bl blobs) Less(i, j int) bool { return bl[i].val < bl[j].val }

// for unit testing (preferrably on small tables)
func (snt *SymNameTable) dump() string {
	// string builder (no strings.Builder to preserve 1.4 bootstrap)
	sb := ""
	var bl blobs
	for k, v := range snt.hash {
		bl = append(bl, blob{hc: k, val: v})
	}
	sort.Sort(blobs(bl))
	sb += "hash {\n"
	for idx, b := range bl {
		s := fmt.Sprintf("  %d: hc=0x%x slot=%d\n", idx, b.hc, b.val)
		if isCollision(b.val) {
			s = fmt.Sprintf("  %d: hc=0x%x *** coll=%d\n",
				idx, b.hc, getCollisionSlot(b.val))
		}
		sb += s
	}
	sb += "}\n"
	sb += "allnames {\n"
	for idx, s := range snt.allnames {
		sb += fmt.Sprintf("  %d: '%s'\n", idx, s)
	}
	sb += "}\n"
	sb += "collisions {\n"
	for idx, c := range snt.collisions {
		sb += fmt.Sprintf("  %d: %v\n", idx, c)
	}
	sb += "}\n"
	return sb
}

// for unit testing
func (snt *SymNameTable) explode(sn SymName) string {
	return "{" + snt.allnames[sn.pref] + ":" + snt.allnames[sn.suf] + "}"
}

// for unit testing only. has side effect of clearing table.
func (snt *SymNameTable) replaceHash(newhasher nameHasher) {
	snt.hasher = newhasher
	for k := range snt.hash {
		delete(snt.hash, k)
	}
	snt.allnames = snt.allnames[:0]
	snt.collisions = snt.collisions[:0]
	snt.install("")
}

type SNTStats struct {
	Entries        uint64 // Total number of strings stored
	TotalStringLen uint64 // Total length of all strings in the table
	Collisions     uint64 // Number of collisions
}

// for unit testing
func (snt *SymNameTable) stats() SNTStats {
	var st SNTStats
	for _, s := range snt.allnames {
		st.TotalStringLen += uint64(len(s))
	}
	st.Entries = uint64(len(snt.allnames))
	for _, sl := range snt.collisions {
		st.Collisions += uint64(len(sl))
	}
	return st
}

// for unit testing.
func (snt *SymNameTable) fragmentString(fragment uint64) string {
	return snt.allnames[fragment]
}

// for unit testing.
func (st SNTStats) String() string {
	col := ""
	if st.Collisions != 0 {
		col = fmt.Sprintf("\nCollisions: %d", st.Collisions)
	}
	return fmt.Sprintf("Entries: %d\nTotalStringLen: %d%s",
		st.Entries, st.TotalStringLen, col)
}

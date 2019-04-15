// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sym

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"sort"
	"strings"
	"syscall"
	"testing"
	"unsafe"
)

type spair struct {
	orig     string
	exploded string
}

func TestSymNames(t *testing.T) {
	snt := NewSymNameTable(1)

	items := [...]spair{
		{"", "{:}"},
		{".", "{.:}"},
		{"..", "{..:}"},
		{"foo.bar", "{foo.:bar}"},
		{"explosive.hamster.iguana", "{explosive.hamster.:iguana}"},
		{"foo.foo.bar", "{foo.foo.:bar}"},
		{"foo.bar.bar", "{foo.bar.:bar}"},
	}
	for idx := 0; idx < len(items); idx++ {
		sn := snt.Lookup(items[idx].orig)
		ex := snt.explode(sn)
		if ex != items[idx].exploded {
			t.Errorf("expected exploded %s got %s", items[idx].exploded, ex)
		}
	}
	var esn SymName
	zsn := snt.Lookup("")
	if esn != zsn {
		t.Errorf("expected zero sn, got %v\n", zsn)
	}
	foobarsn := snt.Lookup("foo.bar.bar")
	foobarsn2 := snt.Lookup("foo.bar.bar")
	if foobarsn != foobarsn2 {
		t.Errorf("expected equality; orig=%v next=%v", foobarsn, foobarsn2)
	}

	// Very basic test of HasPrefix method
	if !snt.HasPrefix(foobarsn, "foo") {
		t.Errorf("HasPrefix(foo) false for sn %s", snt.String(foobarsn))
	}
	if !snt.HasPrefix(foobarsn, "foo.bar") {
		t.Errorf("HasPrefix(foo.bar) false for sn %s", snt.String(foobarsn))
	}
	if !snt.HasPrefix(foobarsn, "foo.bar.b") {
		t.Errorf("HasPrefix(foo.bar.b) false for sn %s", snt.String(foobarsn))
	}

	// Very basic test of HasSuffix method
	if !snt.HasSuffix(foobarsn, "ar") {
		t.Errorf("HasSuffix(ar) false for sn %s", snt.String(foobarsn))
	}
	if !snt.HasSuffix(foobarsn, "bar") {
		t.Errorf("HasSuffix(bar) false for sn %s", snt.String(foobarsn))
	}
	if !snt.HasSuffix(foobarsn, "o.bar.bar") {
		t.Errorf("HasSuffix(o.bar.bar) false for sn %s", snt.String(foobarsn))
	}
}

func doPerfTests() string {
	testfile := os.Getenv("SYMBOLNAME_PERFTESTS")
	if testfile == "" {
		return ""
	}
	f, err := os.Open(testfile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening %s: %v\n", testfile, err)
		return ""
	}
	f.Close()
	return testfile
}

type snst struct {
	count    uint64
	fragment uint64
}

func topTen(sl []snst, snt *SymNameTable) string {
	var sb strings.Builder
	sort.Slice(sl, func(j, i int) bool {
		if sl[i].count != sl[j].count {
			return sl[i].count < sl[j].count
		}
		return sl[i].fragment < sl[j].fragment
	})
	for idx := 0; idx < 10 && idx < len(sl); idx++ {
		sb.WriteString(fmt.Sprintf("%d: '%s'\n", sl[idx].count,
			snt.fragmentString(sl[idx].fragment)))
	}
	return sb.String()
}

const smalltestfile = "testdata/symbolnames.txt"

func TestSymCommoning(t *testing.T) {
	snt := NewSymNameTable(3)

	testfile := os.Getenv("SYMBOLNAME_PERFTESTS")
	if testfile == "" {
		testfile = smalltestfile
	}

	// Construct from file.
	f, err := os.Open(testfile)
	if err != nil {
		t.Errorf("can't open '" + testfile + "'")
	}
	scanner := bufio.NewScanner(f)
	snames := []SymName{}
	for scanner.Scan() {
		line := scanner.Text()
		snames = append(snames, snt.Lookup(line))
	}
	f.Close()

	// Verify stats
	stats := snt.stats()
	fmt.Fprintf(os.Stderr, "stats: %v\n", stats)
	if testfile == smalltestfile {
		estats := SNTStats{
			Entries:        1425,
			TotalStringLen: 32469,
			Collisions:     0,
		}
		if stats.String() != estats.String() {
			t.Errorf("stats mismatch: expected: '%s' got '%s'", estats, stats)
		}
	}

	// Look for most populate prefixes and suffixes
	pmap := make(map[uint64]uint64)
	smap := make(map[uint64]uint64)
	for _, sn := range snames {
		pmap[sn.pref] += 1
		smap[sn.suf] += 1
	}

	var prefixSavings uint64
	prefst := []snst{}
	for k, v := range pmap {
		prefst = append(prefst, snst{count: v, fragment: k})
		if v > 1 {
			prefixSavings += uint64(len(snt.fragmentString(k))) * (v - 1)
		}
	}
	fmt.Fprintf(os.Stderr, "Prefix savings: %d\n", prefixSavings)
	fmt.Fprintf(os.Stderr, "Top 10 prefixes:\n%s", topTen(prefst, &snt))

	var suffixSavings uint64
	sufst := []snst{}
	for k, v := range smap {
		sufst = append(sufst, snst{count: v, fragment: k})
		if v > 1 {
			suffixSavings += uint64(len(snt.fragmentString(k))) * (v - 1)
		}
	}
	fmt.Fprintf(os.Stderr, "Suffix savings: %d\n", suffixSavings)
	fmt.Fprintf(os.Stderr, "Top 10 suffixes:\n%s", topTen(sufst, &snt))
}

func constructTableFromFile(fn string) *SymNameTable {
	ret := NewSymNameTable(5)

	// Construct from file.
	f, err := os.Open(fn)
	if err != nil {
		panic("can't read file")
	}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		tbytes := scanner.Bytes()
		ret.Lookup(string(tbytes))
	}
	f.Close()
	//fmt.Fprintf(os.Stderr, "constructTableFromFile: %s\n", ret.dump())
	return &ret
}

type stringStruct struct {
	str unsafe.Pointer
	len int
}

func mkROString(rodata []byte, soff int64, slen int) string {
	ss := stringStruct{str: unsafe.Pointer(&rodata[soff]), len: slen}
	s := *(*string)(unsafe.Pointer(&ss))
	return s
}

func constructTableFromFileWithMmap(fn string) *SymNameTable {
	ret := NewSymNameTable(101)

	// Construct from file.
	f, err := os.Open(fn)
	if err != nil {
		panic("can't read file")
	}
	fi, err := f.Stat()
	if err != nil {
		panic("can't stat file")
	}

	// Cread a read-only mapping of the file in question. Offset zero, length
	// equal to the length of the file. Q: do we need to round up to page size?
	rodata, err := syscall.Mmap(int(f.Fd()), int64(0), int(fi.Size()),
		syscall.PROT_READ, syscall.MAP_PRIVATE|syscall.MAP_FILE)
	if err != nil {
		log.Fatalf("%v", err)
	}

	// Walk through the file collecting strings.
	scanner := bufio.NewScanner(f)
	var off int64
	for scanner.Scan() {
		tbytes := scanner.Bytes()
		ret.Lookup(mkROString(rodata, off, len(tbytes)))
		off += int64(len(tbytes)) + 1
	}
	f.Close()
	//fmt.Fprintf(os.Stderr, "constructTableFromFileWithMmap: %s\n", ret.dump())
	return &ret
}

type strsl struct {
	sl []string
}

func sliceFromFile(fn string) *strsl {
	var ret strsl
	ret.sl = make([]string, 0, 100000)

	f, err := os.Open(fn)
	if err != nil {
		panic("can't read file")
	}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		ret.sl = append(ret.sl, line)
	}
	f.Close()
	return &ret
}

type emptys struct {
}

type strmap struct {
	m map[string]emptys
}

func mapFromFile(fn string) *strmap {
	var ret strmap
	ret.m = make(map[string]emptys, 100000)

	f, err := os.Open(fn)
	if err != nil {
		panic("can't read file")
	}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if _, ok := ret.m[line]; !ok {
			ret.m[line] = emptys{}
		}
	}
	f.Close()
	return &ret
}

func alloc() uint64 {
	var stats runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&stats)
	return stats.Alloc
}

func TestSymTableStorageConsumptionMemstats1(t *testing.T) {
	testfile := doPerfTests()
	if testfile == "" {
		t.Skip("skipping performance test")
	}

	baseAlloc := alloc()
	snt := constructTableFromFile(testfile)
	afterAlloc := alloc()
	fmt.Fprintf(os.Stderr, "SNT allocation delta: %v (%v - %v)\n",
		afterAlloc-baseAlloc, afterAlloc, baseAlloc)
	fmt.Fprintf(os.Stderr, "SNT stats: %v\n", snt.stats())
}

func TestSymTableStorageConsumptionMemprof1(t *testing.T) {
	testfile := doPerfTests()
	if testfile == "" {
		t.Skip("skipping performance test")
	}

	runtime.MemProfileRate = 1
	f, err := os.Create("/tmp/snt.mem.p")
	if err != nil {
		log.Fatalf("%v", err)
	}
	snt := constructTableFromFile(testfile)
	runtime.GC()
	if err := pprof.WriteHeapProfile(f); err != nil {
		log.Fatalf("%v", err)
	}
	f.Close()
	snt.stats()
}

func TestSymTableStorageConsumptionMemstats2(t *testing.T) {
	testfile := doPerfTests()
	if testfile == "" {
		t.Skip("skipping performance test")
	}

	baseAlloc := alloc()
	strslp := sliceFromFile(testfile)
	afterAlloc := alloc()

	fmt.Fprintf(os.Stderr, "string slice allocation delta: %v (%v - %v)\n",
		afterAlloc-baseAlloc, afterAlloc, baseAlloc)
	fmt.Fprintf(os.Stderr, "sl[0] is %s\n", strslp.sl[0])
}

func TestSymTableStorageConsumptionMemprof2(t *testing.T) {
	testfile := doPerfTests()
	if testfile == "" {
		t.Skip("skipping performance test")
	}

	runtime.MemProfileRate = 1
	f, err := os.Create("/tmp/strsl.mem.p")
	if err != nil {
		log.Fatalf("%v", err)
	}
	strslp := sliceFromFile(testfile)
	runtime.GC()
	if err := pprof.WriteHeapProfile(f); err != nil {
		log.Fatalf("%v", err)
	}
	f.Close()
	fmt.Fprintf(os.Stderr, "sl[0] is %s\n", strslp.sl[0])
}

func TestSymTableStorageConsumptionMemstats3(t *testing.T) {
	testfile := doPerfTests()
	if testfile == "" {
		t.Skip("skipping performance test")
	}

	baseAlloc := alloc()
	snt := constructTableFromFileWithMmap(testfile)
	afterAlloc := alloc()
	fmt.Fprintf(os.Stderr, "SNT allocation delta: %v (%v - %v)\n",
		afterAlloc-baseAlloc, afterAlloc, baseAlloc)
	snt.stats()
}

func TestSymTableStorageConsumptionMemprof3(t *testing.T) {
	testfile := doPerfTests()
	if testfile == "" {
		t.Skip("skipping performance test")
	}

	runtime.MemProfileRate = 1
	f, err := os.Create("/tmp/sntm.mem.p")
	if err != nil {
		log.Fatalf("%v", err)
	}
	snt := constructTableFromFileWithMmap(testfile)
	runtime.GC()
	if err := pprof.WriteHeapProfile(f); err != nil {
		log.Fatalf("%v", err)
	}
	f.Close()
	snt.stats()
}

func TestSymTableStorageConsumptionMemstats4(t *testing.T) {
	testfile := doPerfTests()
	if testfile == "" {
		t.Skip("skipping performance test")
	}

	baseAlloc := alloc()
	smp := mapFromFile(testfile)
	afterAlloc := alloc()
	fmt.Fprintf(os.Stderr, "SMP allocation delta: %v (%v - %v)\n",
		afterAlloc-baseAlloc, afterAlloc, baseAlloc)
	fmt.Fprintf(os.Stderr, "m[\"\"] is %s\n", smp.m[""])
}

func TestSymTableStorageConsumptionMemprof4(t *testing.T) {
	testfile := doPerfTests()
	if testfile == "" {
		t.Skip("skipping performance test")
	}

	runtime.MemProfileRate = 1
	f, err := os.Create("/tmp/smp.mem.p")
	if err != nil {
		log.Fatalf("%v", err)
	}
	smp := mapFromFile(testfile)
	runtime.GC()
	if err := pprof.WriteHeapProfile(f); err != nil {
		log.Fatalf("%v", err)
	}
	f.Close()
	fmt.Fprintf(os.Stderr, "m[\"\"] is %s\n", smp.m[""])
}

// Dummy hasher with pathological performance.
type badHash struct {
}

func (b badHash) Sum64(p []byte) uint64 {
	if len(p) != 0 {
		return uint64(p[0])
	}
	return 0
}

func populateSNT(snt *SymNameTable) ([]SymName, []string) {
	var bh badHash
	snt.replaceHash(bh)
	items := []string{"", "22", "a", "ab", "abc", "b", "bc", "bcd", "q",
		"a.", ".b", "abc.def", "quix.frob", "xyzzy.xyzzy", "fruit",
		"wrap"}
	sns := []SymName{}
	for idx := 0; idx < len(items); idx++ {
		sn := snt.Lookup(items[idx])
		sns = append(sns, sn)
	}
	return sns, items
}

func TestSymTableCollisions(t *testing.T) {
	snt := NewSymNameTable(10101)
	sns, items := populateSNT(&snt)
	for idx := 0; idx < len(items); idx++ {
		again := snt.String(sns[idx])
		if items[idx] != again {
			t.Errorf("String(Lookup(%s)) = %s", items[idx], again)
		}
	}
	stats := snt.stats()
	if stats.Collisions == 0 {
		t.Errorf("expected to see nonzero collisions in this test")
	}
}

func TestSymTableConcurrentAccess(t *testing.T) {
	snt := NewSymNameTable(1031)
	sns, items := populateSNT(&snt)
	snt.Lock()
	for i := 0; i < 10; i++ {
		go func() {
			for idx := 0; idx < len(items); idx++ {
				again := snt.String(sns[idx])
				if items[idx] != again {
					t.Errorf("String(Lookup(%s)) = %s", items[idx], again)
				}
			}
		}()
	}
	snt.Unlock()
}

func TestSymTableHasPrefix(t *testing.T) {
	snt := NewSymNameTable(3)
	sns, items := populateSNT(&snt)
	for idx := 0; idx < len(items); idx++ {
		s := snt.String(sns[idx])
		for j := 0; j < len(s); j++ {
			act := snt.HasPrefix(sns[idx], s[0:j])
			exp := strings.HasPrefix(s, s[0:j])
			if act != exp {
				t.Errorf("snt.HasPrefix(%s,%s) returned %v expected %v",
					s, s[0:j], act, exp)
			}
			nope := snt.HasPrefix(sns[idx], s[0:j]+"~")
			if nope {
				t.Errorf("snt.HasPrefix(%s,%s~) returned true", s, s[0:j])
			}
		}
	}
}

func TestSymTableHasSuffix(t *testing.T) {
	snt := NewSymNameTable(3)
	sns, items := populateSNT(&snt)
	for idx := 0; idx < len(items); idx++ {
		s := snt.String(sns[idx])
		for j := 0; j < len(s); j++ {
			act := snt.HasSuffix(sns[idx], s[j:])
			exp := strings.HasSuffix(s, s[j:])
			if act != exp {
				t.Errorf("snt.HasSuffix(%s,%s) returned %v expected %v",
					s, s[0:j], act, exp)
			}
			nope := snt.HasSuffix(sns[idx], "~"+s[j:])
			if nope {
				t.Errorf("snt.HasSuffix(%s,~%s) returned true", s, s[j:])
			}
		}
	}
}
func TestSymTableNameEqString(t *testing.T) {
	snt := NewSymNameTable(3)
	sns, items := populateSNT(&snt)
	for idx := 0; idx < len(items); idx++ {
		item := items[idx]
		sn := sns[idx]
		if !snt.NameEqString(sn, item) {
			t.Errorf("snt.NameEqual(%v,%s) returned false expected true",
				sn, item)
		}
		if snt.NameEqString(sn, item+"~") {
			t.Errorf("snt.NameEqual(%v,%s) returned true expected false",
				sn, item+"~")
		}
		if snt.NameEqString(sn, "~"+item) {
			t.Errorf("snt.NameEqual(%v,%s) returned true expected false",
				sn, "~"+item)
		}
	}
}

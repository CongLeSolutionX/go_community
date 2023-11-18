// test -v -gcflags=-l escape_reflect.go

// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package escape

import (
	"flag"
	"reflect"
	"regexp"
	"runtime"
	"testing"
	"time"
	"unsafe"
)

var (
	recoverAllFlag  = flag.Bool("recoverall", false, "recover all panics; can be helpful with 'go test -v' to see which functions panic")
	funcTimeoutFlag = flag.Duration("functimeout", time.Minute, "fail if any function takes longer than `duration`")
	runNamesFlag    = flag.String("runnames", "", "test only the names matching `regexp`, ignoring any marked skip")
	listNamesFlag   = flag.String("listnames", "", "list names matching `regexp`, ignoring any marked skip")
)

func TestCallInputFunctions(t *testing.T) {
	tests := []struct {
		name       string
		f          func()
		skip       bool
		allowPanic bool
	}{
		{name: "typ", f: func() { i := gi; typ(i) }},
		{name: "kind", f: func() { i := gi; kind(i) }},
		{name: "int1", f: func() { i := gi; int1(i) }},
		{name: "ptr A", f: func() { i := gi; pi := &i; ptr(pi) }},
		{name: "ptr B", f: func() { i := gi; pi := &i; esc(pi); ptr(pi) }},
		{name: "bytes1 A", f: func() { x := make([]byte, 10); bytes1(x) }},
		{name: "bytes1 B", f: func() { x := make([]byte, 10); esc(x); bytes1(x) }},
		{name: "bytes2 A", f: func() { x := make([]byte, 10); bytes2(x) }},
		{name: "bytes2 B", f: func() { x := make([]byte, 10); esc(x); bytes2(x) }},
		{name: "string1 A", f: func() { s := gs; string1(s) }},
		{name: "string1 B", f: func() { s := gs; esc(s); string1(s) }},
		{name: "string2", f: func() { i := gi; string2(i) }},
		{name: "interface1 A", f: func() { var x any = gi; interface1(x) }},
		{name: "interface1 B", f: func() { var x any = gi; esc(x); interface1(x) }},
		{name: "interface2", f: func() { i := gi; interface2(i) }},
		{name: "interface3", f: func() { i := gi; interface3(i) }},
		{name: "interface4 A", f: func() { i := gi; pi := &i; interface4(pi) }},
		{name: "interface4 B", f: func() { i := gi; pi := &i; esc(pi); interface4(pi) }},
		{name: "addr A", f: func() { i := gi; pi := &i; addr(pi) }},
		{name: "addr B", f: func() { i := gi; pi := &i; esc(pi); addr(pi) }},
		{name: "uintptr1 A", f: func() { i := gi; pi := &i; uintptr1(pi) }},
		{name: "uintptr1 B", f: func() { i := gi; pi := &i; esc(pi); uintptr1(pi) }},
		{name: "unsafeaddr A", f: func() { i := gi; pi := &i; unsafeaddr(pi) }},
		{name: "unsafeaddr B", f: func() { i := gi; pi := &i; esc(pi); unsafeaddr(pi) }},
		{name: "ifacedata A", f: func() { var x any = gi; ifacedata(x) }},
		{name: "ifacedata B", f: func() { var x any = gi; esc(x); ifacedata(x) }},
		{name: "can", f: func() { i := gi; can(i) }},
		{name: "is", f: func() { i := gi; is(i) }},
		{name: "is2 A", f: func() { var x [2]int; is2(x) }},
		{name: "is2 B", f: func() { var x [2]int; esc(x); is2(x) }},
		{name: "is3 A", f: func() { var x struct{a int; b int}; is3(x) }},
		{name: "is3 B", f: func() { var x struct{a int; b int}; esc(x); is3(x) }},
		{name: "overflow", f: func() { i := gi; overflow(i) }},
		{name: "len1 A", f: func() { x := make([]int, 10); len1(x) }},
		{name: "len1 B", f: func() { x := make([]int, 10); esc(x); len1(x) }},
		{name: "len2 A", f: func() { var x [3]int; len2(x) }},
		{name: "len2 B", f: func() { var x [3]int; esc(x); len2(x) }},
		{name: "len3 A", f: func() { s := gs; len3(s) }},
		{name: "len3 B", f: func() { s := gs; esc(s); len3(s) }},
		{name: "len4 A", f: func() { x := make(map[int]int, 10); len4(x) }},
		{name: "len4 B", f: func() { x := make(map[int]int, 10); esc(x); len4(x) }},
		{name: "len5 A", f: func() { x := make(chan int, 10); len5(x) }},
		{name: "len5 B", f: func() { x := make(chan int, 10); esc(x); len5(x) }},
		{name: "cap1 A", f: func() { x := make([]int, 10); cap1(x) }},
		{name: "cap1 B", f: func() { x := make([]int, 10); esc(x); cap1(x) }},
		{name: "cap2 A", f: func() { var x [3]int; cap2(x) }},
		{name: "cap2 B", f: func() { var x [3]int; esc(x); cap2(x) }},
		{name: "cap3 A", f: func() { x := make(chan int, 10); cap3(x) }},
		{name: "cap3 B", f: func() { x := make(chan int, 10); esc(x); cap3(x) }},
		{name: "setlen A", allowPanic: true, f: func() { x := make([]int, 10); px := &x; i := gi; setlen(px, i) }},
		{name: "setlen B", allowPanic: true, f: func() { x := make([]int, 10); px := &x; i := gi; esc(px); setlen(px, i) }},
		{name: "setcap A", allowPanic: true, f: func() { x := make([]int, 10); px := &x; i := gi; setcap(px, i) }},
		{name: "setcap B", allowPanic: true, f: func() { x := make([]int, 10); px := &x; i := gi; esc(px); setcap(px, i) }},
		{name: "slice1 A", f: func() { x := make([]byte, 10); slice1(x) }},
		{name: "slice1 B", f: func() { x := make([]byte, 10); esc(x); slice1(x) }},
		{name: "slice2 A", f: func() { s := gs; slice2(s) }},
		{name: "slice2 B", f: func() { s := gs; esc(s); slice2(s) }},
		{name: "slice3 A", allowPanic: true, f: func() { var x [10]byte; slice3(x) }},
		{name: "slice3 B", allowPanic: true, f: func() { var x [10]byte; esc(x); slice3(x) }},
		{name: "elem1 A", f: func() { i := gi; pi := &i; elem1(pi) }},
		{name: "elem1 B", f: func() { i := gi; pi := &i; esc(pi); elem1(pi) }},
		{name: "elem2 A", f: func() { s := gs; ps := &s; elem2(ps) }},
		{name: "elem2 B", f: func() { s := gs; ps := &s; esc(ps); elem2(ps) }},
		{name: "field1 A", f: func() { var x S; field1(x) }},
		{name: "field1 B", f: func() { var x S; esc(x); field1(x) }},
		{name: "field2 A", f: func() { var x S; field2(x) }},
		{name: "field2 B", f: func() { var x S; esc(x); field2(x) }},
		{name: "numfield A", f: func() { var x S; numfield(x) }},
		{name: "numfield B", f: func() { var x S; esc(x); numfield(x) }},
		{name: "index1 A", f: func() { x := make([]int, 10); index1(x) }},
		{name: "index1 B", f: func() { x := make([]int, 10); esc(x); index1(x) }},
		{name: "index2 A", f: func() { x := make([]string, 10); index2(x) }},
		{name: "index2 B", f: func() { x := make([]string, 10); esc(x); index2(x) }},
		{name: "index3 A", f: func() { var x [3]int; index3(x) }},
		{name: "index3 B", f: func() { var x [3]int; esc(x); index3(x) }},
		{name: "index4 A", f: func() { var x [3]string; index4(x) }},
		{name: "index4 B", f: func() { var x [3]string; esc(x); index4(x) }},
		{name: "index5 A", f: func() { s := gs; index5(s) }},
		{name: "index5 B", f: func() { s := gs; esc(s); index5(s) }},
		{name: "call1 A", allowPanic: true, f: func() { var f func(int); i := gi; call1(f, i) }},
		{name: "call1 B", allowPanic: true, f: func() { var f func(int); i := gi; esc(f); call1(f, i) }},
		{name: "call2 A", allowPanic: true, f: func() { var f func(*int); i := gi; pi := &i; call2(f, pi) }},
		{name: "call2 B", allowPanic: true, f: func() { var f func(*int); i := gi; pi := &i; esc(f); call2(f, pi) }},
		{name: "call2 C", allowPanic: true, f: func() { var f func(*int); i := gi; pi := &i; esc(pi); call2(f, pi) }},
		{name: "call2 D", allowPanic: true, f: func() { var f func(*int); i := gi; pi := &i; esc(f); esc(pi); call2(f, pi) }},
		{name: "method A", f: func() { var x S; method(x) }},
		{name: "method B", f: func() { var x S; esc(x); method(x) }},
		{name: "nummethod A", f: func() { var x S; nummethod(x) }},
		{name: "nummethod B", f: func() { var x S; esc(x); nummethod(x) }},
		{name: "mapindex A", f: func() { m := make(map[string]string, 10); s := gs; mapindex(m, s) }},
		{name: "mapindex B", f: func() { m := make(map[string]string, 10); s := gs; esc(m); mapindex(m, s) }},
		{name: "mapindex C", f: func() { m := make(map[string]string, 10); s := gs; esc(s); mapindex(m, s) }},
		{name: "mapindex D", f: func() { m := make(map[string]string, 10); s := gs; esc(m); esc(s); mapindex(m, s) }},
		{name: "mapkeys A", f: func() { m := make(map[string]string, 10); mapkeys(m) }},
		{name: "mapkeys B", f: func() { m := make(map[string]string, 10); esc(m); mapkeys(m) }},
		{name: "mapiter1 A", f: func() { m := make(map[string]string, 10); mapiter1(m) }},
		{name: "mapiter1 B", f: func() { m := make(map[string]string, 10); esc(m); mapiter1(m) }},
		{name: "mapiter2 A", f: func() { m := make(map[string]string, 10); mapiter2(m) }},
		{name: "mapiter2 B", f: func() { m := make(map[string]string, 10); esc(m); mapiter2(m) }},
		{name: "mapiter3 A", f: func() { m := make(map[string]string, 10); var i reflect.MapIter; pi := &i; mapiter3(m, pi) }},
		{name: "mapiter3 B", f: func() { m := make(map[string]string, 10); var i reflect.MapIter; pi := &i; esc(m); mapiter3(m, pi) }},
		{name: "mapiter3 C", f: func() { m := make(map[string]string, 10); var i reflect.MapIter; pi := &i; esc(pi); mapiter3(m, pi) }},
		{name: "mapiter3 D", f: func() { m := make(map[string]string, 10); var i reflect.MapIter; pi := &i; esc(m); esc(pi); mapiter3(m, pi) }},
		{name: "recv1 A", skip: true, f: func() { c := make(chan string, 10); recv1(c) }},
		{name: "recv1 B", skip: true, f: func() { c := make(chan string, 10); esc(c); recv1(c) }},
		{name: "recv2 A", f: func() { c := make(chan string, 10); recv2(c) }},
		{name: "recv2 B", f: func() { c := make(chan string, 10); esc(c); recv2(c) }},
		{name: "send1 A", f: func() { c := make(chan string, 10); s := gs; send1(c, s) }},
		{name: "send1 B", f: func() { c := make(chan string, 10); s := gs; esc(c); send1(c, s) }},
		{name: "send1 C", f: func() { c := make(chan string, 10); s := gs; esc(s); send1(c, s) }},
		{name: "send1 D", f: func() { c := make(chan string, 10); s := gs; esc(c); esc(s); send1(c, s) }},
		{name: "send2 A", f: func() { c := make(chan string, 10); s := gs; send2(c, s) }},
		{name: "send2 B", f: func() { c := make(chan string, 10); s := gs; esc(c); send2(c, s) }},
		{name: "send2 C", f: func() { c := make(chan string, 10); s := gs; esc(s); send2(c, s) }},
		{name: "send2 D", f: func() { c := make(chan string, 10); s := gs; esc(c); esc(s); send2(c, s) }},
		{name: "close1 A", f: func() { c := make(chan string, 10); close1(c) }},
		{name: "close1 B", f: func() { c := make(chan string, 10); esc(c); close1(c) }},
		{name: "select1 A", skip: true, f: func() { c := make(chan string, 10); select1(c) }},
		{name: "select1 B", skip: true, f: func() { c := make(chan string, 10); esc(c); select1(c) }},
		{name: "select2 A", f: func() { c := make(chan string, 10); s := gs; select2(c, s) }},
		{name: "select2 B", f: func() { c := make(chan string, 10); s := gs; esc(c); select2(c, s) }},
		{name: "select2 C", f: func() { c := make(chan string, 10); s := gs; esc(s); select2(c, s) }},
		{name: "select2 D", f: func() { c := make(chan string, 10); s := gs; esc(c); esc(s); select2(c, s) }},
		{name: "convert1", f: func() { i := gi; convert1(i) }},
		{name: "convert2 A", f: func() { x := make([]byte, 10); convert2(x) }},
		{name: "convert2 B", f: func() { x := make([]byte, 10); esc(x); convert2(x) }},
		{name: "set1 A", f: func() { i := gi; v := reflect.ValueOf(&i).Elem(); i2 := gi; set1(v, i2) }},
		{name: "set1 B", f: func() { i := gi; v := reflect.ValueOf(&i).Elem(); i2 := gi; esc(v); set1(v, i2) }},
		{name: "set2", f: func() { i := gi; set2(i) }},
		{name: "set3 A", f: func() { i := gi; v := reflect.ValueOf(&i).Elem(); i2 := gi; set3(v, i2) }},
		{name: "set3 B", f: func() { i := gi; v := reflect.ValueOf(&i).Elem(); i2 := gi; esc(v); set3(v, i2) }},
		{name: "set4", f: func() { i := gi; set4(i) }},
		{name: "set5 A", f: func() { s := gs; v := reflect.ValueOf(&s).Elem(); s2 := gs; set5(v, s2) }},
		{name: "set5 B", f: func() { s := gs; v := reflect.ValueOf(&s).Elem(); s2 := gs; esc(v); set5(v, s2) }},
		{name: "set5 C", f: func() { s := gs; v := reflect.ValueOf(&s).Elem(); s2 := gs; esc(s2); set5(v, s2) }},
		{name: "set5 D", f: func() { s := gs; v := reflect.ValueOf(&s).Elem(); s2 := gs; esc(v); esc(s2); set5(v, s2) }},
		{name: "set6 A", f: func() { x := make([]byte, 10); v := reflect.ValueOf(&x).Elem(); x2 := make([]byte, 10); set6(v, x2) }},
		{name: "set6 B", f: func() { x := make([]byte, 10); v := reflect.ValueOf(&x).Elem(); x2 := make([]byte, 10); esc(v); set6(v, x2) }},
		{name: "set6 C", f: func() { x := make([]byte, 10); v := reflect.ValueOf(&x).Elem(); x2 := make([]byte, 10); esc(x2); set6(v, x2) }},
		{name: "set6 D", f: func() { x := make([]byte, 10); v := reflect.ValueOf(&x).Elem(); x2 := make([]byte, 10); esc(v); esc(x2); set6(v, x2) }},
		{name: "set7 A", f: func() { var u unsafe.Pointer; v := reflect.ValueOf(&u).Elem(); var u2 unsafe.Pointer; set7(v, u2) }},
		{name: "set7 B", f: func() { var u unsafe.Pointer; v := reflect.ValueOf(&u).Elem(); var u2 unsafe.Pointer; esc(v); set7(v, u2) }},
		{name: "set7 C", f: func() { var u unsafe.Pointer; v := reflect.ValueOf(&u).Elem(); var u2 unsafe.Pointer; esc(u2); set7(v, u2) }},
		{name: "set7 D", f: func() { var u unsafe.Pointer; v := reflect.ValueOf(&u).Elem(); var u2 unsafe.Pointer; esc(v); esc(u2); set7(v, u2) }},
		{name: "setmapindex A", f: func() { m := make(map[string]string, 10); s := gs; s2 := gs; setmapindex(m, s, s2) }},
		{name: "setmapindex B", f: func() { m := make(map[string]string, 10); s := gs; s2 := gs; esc(m); setmapindex(m, s, s2) }},
		{name: "setmapindex C", f: func() { m := make(map[string]string, 10); s := gs; s2 := gs; esc(s); setmapindex(m, s, s2) }},
		{name: "setmapindex D", f: func() { m := make(map[string]string, 10); s := gs; s2 := gs; esc(m); esc(s); setmapindex(m, s, s2) }},
		{name: "setmapindex E", f: func() { m := make(map[string]string, 10); s := gs; s2 := gs; esc(s2); setmapindex(m, s, s2) }},
		{name: "setmapindex F", f: func() { m := make(map[string]string, 10); s := gs; s2 := gs; esc(m); esc(s2); setmapindex(m, s, s2) }},
		{name: "setmapindex G", f: func() { m := make(map[string]string, 10); s := gs; s2 := gs; esc(s); esc(s2); setmapindex(m, s, s2) }},
		{name: "setmapindex H", f: func() { m := make(map[string]string, 10); s := gs; s2 := gs; esc(m); esc(s); esc(s2); setmapindex(m, s, s2) }},
		{name: "mapdelete A", f: func() { m := make(map[string]string, 10); s := gs; mapdelete(m, s) }},
		{name: "mapdelete B", f: func() { m := make(map[string]string, 10); s := gs; esc(m); mapdelete(m, s) }},
		{name: "mapdelete C", f: func() { m := make(map[string]string, 10); s := gs; esc(s); mapdelete(m, s) }},
		{name: "mapdelete D", f: func() { m := make(map[string]string, 10); s := gs; esc(m); esc(s); mapdelete(m, s) }},
		{name: "setiterkey1 A", allowPanic: true, f: func() { var v reflect.Value; var i reflect.MapIter; pi := &i; setiterkey1(v, pi) }},
		{name: "setiterkey1 B", allowPanic: true, f: func() { var v reflect.Value; var i reflect.MapIter; pi := &i; esc(v); setiterkey1(v, pi) }},
		{name: "setiterkey1 C", allowPanic: true, f: func() { var v reflect.Value; var i reflect.MapIter; pi := &i; esc(pi); setiterkey1(v, pi) }},
		{name: "setiterkey1 D", allowPanic: true, f: func() { var v reflect.Value; var i reflect.MapIter; pi := &i; esc(v); esc(pi); setiterkey1(v, pi) }},
		{name: "setiterkey2 A", allowPanic: true, f: func() { m := make(map[string]string, 10); v := reflect.ValueOf(&m).Elem(); m2 := make(map[string]string, 10); setiterkey2(v, m2) }},
		{name: "setiterkey2 B", allowPanic: true, f: func() { m := make(map[string]string, 10); v := reflect.ValueOf(&m).Elem(); m2 := make(map[string]string, 10); esc(v); setiterkey2(v, m2) }},
		{name: "setiterkey2 C", allowPanic: true, f: func() { m := make(map[string]string, 10); v := reflect.ValueOf(&m).Elem(); m2 := make(map[string]string, 10); esc(m2); setiterkey2(v, m2) }},
		{name: "setiterkey2 D", allowPanic: true, f: func() { m := make(map[string]string, 10); v := reflect.ValueOf(&m).Elem(); m2 := make(map[string]string, 10); esc(v); esc(m2); setiterkey2(v, m2) }},
		{name: "setitervalue1 A", allowPanic: true, f: func() { var v reflect.Value; var i reflect.MapIter; pi := &i; setitervalue1(v, pi) }},
		{name: "setitervalue1 B", allowPanic: true, f: func() { var v reflect.Value; var i reflect.MapIter; pi := &i; esc(v); setitervalue1(v, pi) }},
		{name: "setitervalue1 C", allowPanic: true, f: func() { var v reflect.Value; var i reflect.MapIter; pi := &i; esc(pi); setitervalue1(v, pi) }},
		{name: "setitervalue1 D", allowPanic: true, f: func() { var v reflect.Value; var i reflect.MapIter; pi := &i; esc(v); esc(pi); setitervalue1(v, pi) }},
		{name: "setitervalue2 A", allowPanic: true, f: func() { m := make(map[string]string, 10); v := reflect.ValueOf(&m).Elem(); m2 := make(map[string]string, 10); setitervalue2(v, m2) }},
		{name: "setitervalue2 B", allowPanic: true, f: func() { m := make(map[string]string, 10); v := reflect.ValueOf(&m).Elem(); m2 := make(map[string]string, 10); esc(v); setitervalue2(v, m2) }},
		{name: "setitervalue2 C", allowPanic: true, f: func() { m := make(map[string]string, 10); v := reflect.ValueOf(&m).Elem(); m2 := make(map[string]string, 10); esc(m2); setitervalue2(v, m2) }},
		{name: "setitervalue2 D", allowPanic: true, f: func() { m := make(map[string]string, 10); v := reflect.ValueOf(&m).Elem(); m2 := make(map[string]string, 10); esc(v); esc(m2); setitervalue2(v, m2) }},
		{name: "append1 A", f: func() { s := make([]int, 10); i := gi; append1(s, i) }},
		{name: "append1 B", f: func() { s := make([]int, 10); i := gi; esc(s); append1(s, i) }},
		{name: "append2 A", f: func() { s := make([]int, 10); x := make([]int, 10); append2(s, x) }},
		{name: "append2 B", f: func() { s := make([]int, 10); x := make([]int, 10); esc(s); append2(s, x) }},
		{name: "append2 C", f: func() { s := make([]int, 10); x := make([]int, 10); esc(x); append2(s, x) }},
		{name: "append2 D", f: func() { s := make([]int, 10); x := make([]int, 10); esc(s); esc(x); append2(s, x) }},
	}

	listNamesRe := regexp.MustCompile(*listNamesFlag)
	runNamesRe := regexp.MustCompile(*runNamesFlag)
	// Run our test functions one after another in a single goroutine
	// that calls runtime.GC at the end. In some cases, this might help
	// the runtime recognize an illegal heap pointer to a stack variable.
	starting := make(chan string)
	done := make(chan struct{})
	go func() {
		for _, tt := range tests {
			if tt.skip {
				continue
			}
			if *listNamesFlag != "" && listNamesRe.MatchString(tt.name) {
				t.Log("func name:", tt.name)
				continue
			}

			if runNamesRe.MatchString(tt.name) {
				starting <- tt.name
				func() {
					defer func() {
						if r := recover(); r != nil {
							if !tt.allowPanic && !*recoverAllFlag {
								t.Log("panic in:", tt.name)
								panic(r)
							}
							t.Log("recovered panic:", tt.name)
						}
					}()

					// Call the function under test.
					tt.f()
				}()
			}
		}

		for i := 0; i < 5; i++ {
			runtime.GC()
		}
		close(done)
	}()

	var current string
	for {
		select {
		case current = <-starting:
			t.Log("starting func:", current)
		case <-time.After(*funcTimeoutFlag):
			t.Fatal("timed out func:", current)
		case <-done:
			return
		}
	}
}

// Global int and string values to help avoid some heap allocations being optimized away
// via the compiler or runtime recognizing constants or small integer values.
var (
	gi = 1000
	gs = "abcd"
)

// esc forces x to escape.
func esc(x any) {
	if escSink.b {
		escSink.x = x
	}
}

var escSink struct {
	b bool
	x any
}

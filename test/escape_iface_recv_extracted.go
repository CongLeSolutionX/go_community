// errorcheck -0 -m -l

// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test escape analysis for interface receivers using
// simplified examples extracted from std or cmd.
//
// These were usually selected because their behavior
// changed under different implementations of escape
// analysis, including some WIP implementations.

package example

// TODO: we could eliminate all or most of these to reduce
// sensitivity to changes in stdlib, though might be OK to
// keep some (e.g., if using interfaces declared in the io package).
import (
	"bytes"
	"compress/lzw"
	"encoding/binary"
	"io"
	"io/fs"
	"sort"
)

// Example loosely modeled on the fmt print family.
// It does not emulate the various reflect operations used in the real fmt.

func usePrint() {
	val := 1000
	Print(val) // ERROR "... argument does not escape$" "val does not escape$"
}

func Print(args ...any) { // ERROR "might leak param content: args$"
	doPrint(args)
}

func doPrint(args []any) { // ERROR "might leak param content: args$"
	for _, a := range args {
		if handleMethods(inputArg{eface: a}) {
			continue
		}
		switch v := a.(type) {
		case int:
			println(v)
			continue
		}
		println(a)
	}
}

func handleMethods(arg inputArg) bool { // ERROR "might leak param: arg$"
	if v, ok := arg.eface.(GoStringer); ok {
		println(v.GoString())
		return true
	}
	switch v := arg.eface.(type) {
	case Stringer:
		println(v.String())
		return true
	}
	return false
}

type inputArg struct {
	eface any
	x     int
}
type Stringer interface{ String() string }
type GoStringer interface{ GoString() string }

// Example similar to the prior one, but manually inlined.

func usePrintInlined() {
	val := 1000
	args := []any{val} // ERROR "\[\]any{...} does not escape$" "val does not escape$"
	for _, a := range args {
		arg := inputArg{eface: a}
		if v, ok := arg.eface.(GoStringer); ok {
			println(v.GoString())
			continue
		}
		switch v := arg.eface.(type) {
		case Stringer:
			println(v.String())
			continue
		}
		switch v := a.(type) {
		case int:
			println(v)
			continue
		}
		println(a)
	}
}

// Example loosely modeled on math/big.byteReader.

type byteReader struct {
	ScanState
}

func (r byteReader) ReadByte() (byte, error) { // ERROR "leaking param: r$"
	ch, _, err := r.ReadRune()
	return byte(ch), err
}

func (r byteReader) UnreadByte() error { // ERROR "leaking param: r$"
	return r.UnreadRune()
}

type ScanState interface {
	// Similar to fmt.ScanState.
	ReadRune() (r rune, size int, err error)
	UnreadRune() error
}

// Example loosely modeled on context.stringify.

type stringer interface{ String() string }

func stringify(v any) string { // ERROR "leaking param: v to result ~r0 level=1$" "might leak param: v$"
	switch s := v.(type) {
	case stringer:
		return s.String()
	case string:
		return s
	}
	return "<not Stringer>"
}

// Example loosely modeled on context.withoutCancelCtx.

type withoutCancelCtx struct {
	c Context
}

func (withoutCancelCtx) Done() <-chan struct{} {
	return nil
}

func (withoutCancelCtx) Err() error {
	return nil
}

func (c withoutCancelCtx) Value(key any) any { // ERROR "leaking param: c$" "leaking param: key$"
	return value(c, key) // ERROR "c escapes to heap$"
}

// TODO: is "leaking param content: c" correct here? (Go 1.21 has "leaking param: c").
// -m=2 output includes:
//
//	.\t.go:20:12: parameter c leaks to <storage for ctx> in value with derefs=1:
//	.\t.go:20:12:   flow: <temp> ← c:
//	.\t.go:20:12:   flow: ctx ← *<temp>:
//	.\t.go:20:12:     from case <node DYNAMICTYPE>: return ctx (switch case) at .\t.go:23:3
//	.\t.go:20:12:   flow: <storage for ctx> ← ctx:
//	.\t.go:20:12:     from ctx (interface-converted) at .\t.go:24:11
//	.\t.go:20:12: leaking param content: c
func value(c Context, key any) any { // ERROR "leaking param content: c$" "leaking param: key$" "might leak param: c$"
	for {
		switch ctx := c.(type) {
		case withoutCancelCtx:
			return ctx // ERROR "ctx escapes to heap$"
		default:
			return c.Value(key)
		}
	}
}

type Context interface {
	Done() <-chan struct{}
	Err() error
	Value(key any) any
}

// Example loosely modeled on debug/gosym.funcTab.

type LineTable struct {
	binary  binary.ByteOrder
	functab []byte
}

type funcTab struct {
	*LineTable
	sz int
}

func (f1 funcTab) funcOff(i int) uint64 { // ERROR "leaking param content: f1$"
	return f1.uint(f1.functab[(2*i+1)*f1.sz:])
}

func (f2 funcTab) uint(b []byte) uint64 { // ERROR "leaking param content: f2$" "leaking param: b$"
	if f2.sz == 4 {
		return uint64(f2.binary.Uint32(b))
	}
	return f2.binary.Uint64(b)
}

// Example loosely modeled on image/gif.encoder.writeImageBlock.
//
// TODO: consider removing this example, or otherwise removing lzw import.

func writeImageBlock(pm Paletted) { // ERROR "leaking param: pm$"
	lzww := lzw.NewWriter(io.Discard, lzw.LSB, 999)
	lzww.Write(pm.Pix[:999])
}

type Paletted struct {
	Pix []uint8
}

// Example loosely modeled on io.nopCloserWriterTo.

func NopCloser(r io.Reader) io.ReadCloser { // ERROR "leaking param: r$"
	if _, ok := r.(io.WriterTo); ok {
		return nopCloserWriterTo{r} // ERROR "nopCloserWriterTo{...} escapes to heap$"
	}
	return nopCloser{r} // ERROR "nopCloser{...} escapes to heap$"
}

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

type nopCloserWriterTo struct {
	io.Reader
}

func (nopCloserWriterTo) Close() error { return nil }

func (c nopCloserWriterTo) WriteTo(w io.Writer) (n int64, err error) { // ERROR "leaking param: c$" "leaking param: w$"
	return c.Reader.(io.WriterTo).WriteTo(w)
}

// Example loosely modeled on src/testing/fstest.fsOnly.
type fsOnly struct{ fs.FS }

type MapFS map[string]*MapFile

// TODO: is "leaking param: fsys" correct? (Same as reported by Go 1.21).
func (fsys MapFS) ReadFile(name string) ([]byte, error) { // ERROR "leaking param: fsys$" "leaking param: name$"
	return fs.ReadFile(fsOnly{fsys}, name) // ERROR "fsOnly{...} does not escape$"
}

func (fsys MapFS) Open(name string) (fs.File, error) { return nil, nil } // ERROR "fsys does not escape$" "name does not escape$"

type MapFile struct{ Sys any }

// Example loosely modeled on src/io/fs.dirInfo.

type dirInfo struct {
	fileInfo FileInfo
}

func (di1 dirInfo) Info() (FileInfo, error) { // ERROR "leaking param: di1 to result ~r0 level=0$"
	return di1.fileInfo, nil
}

func (di2 dirInfo) Name() string { // ERROR "leaking param: di2$"
	return di2.fileInfo.Name()
}

func (di3 dirInfo) String() string { // ERROR "leaking param: di3$"
	return FormatDirEntry(di3) // ERROR "di3 escapes to heap$"
}

func FormatDirEntry(dir DirEntry) string { // ERROR "might leak param: dir$"
	return dir.Name()
}

type FileInfo interface {
	Name() string
}

type DirEntry interface {
	Name() string
}

// Example loosely modeled on go/parser.ParseFile.

func useParseFile() {
	var b []byte
	_ = ParseFile(b) // ERROR "b does not escape$"
}

func ParseFile(src any) []byte { // ERROR "leaking param: src to result ~r0 level=1$" "might leak param: src$"
	return readSource(src)
}

func readSource(src any) []byte { // ERROR "leaking param: src to result ~r0 level=1$" "might leak param: src$"
	switch s := src.(type) {
	case string:
		return []byte(s) // ERROR "\(\[\]byte\)\(s\) escapes to heap$"
	case []byte:
		return s
	case *bytes.Buffer:
		if s != nil {
			return s.Bytes()
		}
	case io.Reader:
		b, _ := io.ReadAll(s)
		return b
	}
	return nil
}

// Example loosely modeled on runtime/pprof.elfBuildID.

func useByteOrder(order int) uint32 {
	var byteOrder binary.ByteOrder
	switch order {
	case 1:
		byteOrder = binary.LittleEndian // ERROR "binary.LittleEndian does not escape$"
	case 2:
		byteOrder = binary.BigEndian // ERROR "binary.BigEndian does not escape$"
	}

	b := []byte{0x1, 0x2, 0x3, 0x4} // ERROR "\[\]byte{...} escapes to heap$"
	ret := byteOrder.Uint32(b)
	return ret
}

// Example using sort.Reverse.

func sortReverse(s []int) []int { // ERROR "leaking param: s$"
	i := sort.IntSlice(s)
	r := sort.Reverse(i) // ERROR "i escapes to heap$"
	sort.Sort(r)
	return s
}

// Example loosely modeled on strings.(*genericReplacer).Replace.

type stringWriter struct {
	w io.Writer
}

func (wRecv stringWriter) WriteString(s string) (int, error) { // ERROR "leaking param: wRecv$" "s does not escape$"
	b := []byte(s) // ERROR "\(\[\]byte\)\(s\) escapes to heap$"
	return wRecv.w.Write(b)
}

func getStringWriter(w1 io.Writer) io.StringWriter { // ERROR "leaking param: w1$"
	sw, ok := w1.(io.StringWriter)
	if !ok {
		sw = stringWriter{w1} // ERROR "stringWriter{...} escapes to heap$"
	}
	return sw
}

func genericReplacerWriteString(w2 io.Writer, s string) { // ERROR "leaking param: s$" "leaking param: w2$"
	sw := getStringWriter(w2)
	_, _ = sw.WriteString(s[0:999])
	return
}

// Example loosely modeled on regexp/syntax.cleanClass.

type ranges struct {
	p *[]rune
}

func (ra ranges) Less(i, j int) bool { return false } // ERROR "ra does not escape$"
func (ra ranges) Len() int           { return 0 }     // ERROR "ra does not escape$"
func (ra ranges) Swap(i, j int)      {}               // ERROR "ra does not escape$"

func cleanClass(rp *[]rune) []rune { // ERROR "leaking param: rp to result ~r0 level=1$"
	sort.Sort(ranges{rp})
	r := *rp
	return r
}

// Example loosely modeled on archive/tar.readSpecialFile.

func readSpecialFile(r io.Reader) ([]byte, error) { // ERROR "leaking param: r$"
	buf, err := io.ReadAll(io.LimitReader(r, 999))
	return buf, err
}

// Example loosely modeled on testing.fuzzResult.

type fuzzResult struct {
	N     int
	Error error
}

// TODO: is this right? is this similar to handleMethods in fmt, but here
// passing the struct as the receiver parameter?
func (r fuzzResult) String() string { // ERROR "leaking param: r$"
	if r.Error == nil {
		return ""
	}
	return r.Error.Error()
}

func useFuzzResult() {
	f := fuzzResult{}
	_ = f.String()
}

// Example loosely modeled on cmd/go/internal/web.hookCloser.

type hookCloser struct {
	io.ReadCloser
	afterClose func()
}

func (c hookCloser) Close() error { // ERROR "leaking param: c$"
	err := c.ReadCloser.Close()
	return err
}

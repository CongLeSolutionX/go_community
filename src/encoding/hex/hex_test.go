// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hex

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"testing"
)

type encDecTest struct {
	enc string
	dec []byte
}

var encDecTests = []encDecTest{
	{"", []byte{}},
	{"0001020304050607", []byte{0, 1, 2, 3, 4, 5, 6, 7}},
	{"08090a0b0c0d0e0f", []byte{8, 9, 10, 11, 12, 13, 14, 15}},
	{"f0f1f2f3f4f5f6f7", []byte{0xf0, 0xf1, 0xf2, 0xf3, 0xf4, 0xf5, 0xf6, 0xf7}},
	{"f8f9fafbfcfdfeff", []byte{0xf8, 0xf9, 0xfa, 0xfb, 0xfc, 0xfd, 0xfe, 0xff}},
	{"67", []byte{'g'}},
	{"e3a1", []byte{0xe3, 0xa1}},
}

func TestEncode(t *testing.T) {
	for i, test := range encDecTests {
		dst := make([]byte, EncodedLen(len(test.dec)))
		n := Encode(dst, test.dec)
		if n != len(dst) {
			t.Errorf("#%d: bad return value: got: %d want: %d", i, n, len(dst))
		}
		if string(dst) != test.enc {
			t.Errorf("#%d: got: %#v want: %#v", i, dst, test.enc)
		}
	}
}

func TestDecode(t *testing.T) {
	// Case for decoding uppercase hex characters, since
	// Encode always uses lowercase.
	decTests := append(encDecTests, encDecTest{"F8F9FAFBFCFDFEFF", []byte{0xf8, 0xf9, 0xfa, 0xfb, 0xfc, 0xfd, 0xfe, 0xff}})
	for i, test := range decTests {
		dst := make([]byte, DecodedLen(len(test.enc)))
		n, err := Decode(dst, []byte(test.enc))
		if err != nil {
			t.Errorf("#%d: bad return value: got:%d want:%d", i, n, len(dst))
		} else if !bytes.Equal(dst, test.dec) {
			t.Errorf("#%d: got: %#v want: %#v", i, dst, test.dec)
		}
	}
}

func TestEncodeToString(t *testing.T) {
	for i, test := range encDecTests {
		s := EncodeToString(test.dec)
		if s != test.enc {
			t.Errorf("#%d got:%s want:%s", i, s, test.enc)
		}
	}
}

func TestDecodeString(t *testing.T) {
	for i, test := range encDecTests {
		dst, err := DecodeString(test.enc)
		if err != nil {
			t.Errorf("#%d: unexpected err value: %s", i, err)
			continue
		}
		if !bytes.Equal(dst, test.dec) {
			t.Errorf("#%d: got: %#v want: #%v", i, dst, test.dec)
		}
	}
}

type errTest struct {
	in  string
	err error
}

var errTests = []errTest{
	{"0", ErrLength},
	{"zd4aa", ErrLength},
	{"0g", InvalidByteError('g')},
	{"00gg", InvalidByteError('g')},
	{"0\x01", InvalidByteError('\x01')},
}

func TestInvalidErr(t *testing.T) {
	for i, test := range errTests {
		dst := make([]byte, DecodedLen(len(test.in)))
		_, err := Decode(dst, []byte(test.in))
		if err == nil {
			t.Errorf("#%d: expected %v; got none", i, test.err)
		} else if err != test.err {
			t.Errorf("#%d: got: %v want: %v", i, err, test.err)
		}
	}
}

func TestInvalidStringErr(t *testing.T) {
	for i, test := range errTests {
		_, err := DecodeString(test.in)
		if err == nil {
			t.Errorf("#%d: expected %v; got none", i, test.err)
		} else if err != test.err {
			t.Errorf("#%d: got: %v want: %v", i, err, test.err)
		}
	}
}

func TestEncoderDecoder(t *testing.T) {
	var buf bytes.Buffer

	for _, multiplier := range []int{1, 128, 192} {
		for i, test := range encDecTests {
			buf.Reset()

			input := bytes.Repeat(test.dec, multiplier)
			output := strings.Repeat(test.enc, multiplier)

			enc := NewEncoder(&buf)
			if n, _ := io.Copy(enc, bytes.NewReader(input)); n != int64(len(input)) {
				t.Errorf("#%d (multipler %d) bytes written: got: %#v want: %#v", i, multiplier, n, len(input))
				continue
			}

			if encDst := buf.String(); encDst != output {
				t.Errorf("#%d (multipler %d): got: %#v want: %#v", i, multiplier, encDst, output)
				continue
			}

			dec := NewDecoder(&buf)
			decDst, err := ioutil.ReadAll(dec)
			if err != nil {
				t.Errorf("#%d (multipler %d): unexpected error value: %s", i, multiplier, err)
				continue
			}
			if !bytes.Equal(decDst, input) {
				t.Errorf("#%d (multipler %d): got: %#v want: #%v", i, multiplier, decDst, input)
			}
		}
	}
}

var errDecoderTests = []errTest{
	{"0", io.ErrUnexpectedEOF},
	{"0g", InvalidByteError('g')},
	{"00gg", InvalidByteError('g')},
	{"0\x01", InvalidByteError('\x01')},
}

func TestDecoderErr(t *testing.T) {
	for i, test := range errDecoderTests {
		dec := NewDecoder(strings.NewReader(test.in))
		_, err := ioutil.ReadAll(dec)
		if err == nil {
			t.Errorf("#%d: expected %v; got none", i, test.err)
		} else if err != test.err {
			t.Errorf("#%d: got: %v want: %v", i, err, test.err)
		}
	}
}

func TestDecoderUnevenErr(t *testing.T) {
	dec := NewDecoder(strings.NewReader("ffeed"))
	var out [1]byte

	_, err := dec.Read(nil)
	if err != nil {
		t.Fatalf("expected nil error; got %v", err)
	}

	n, err := dec.Read(out[:])
	if n != 1 {
		t.Fatalf("expected n = %v; got %d", 1, n)
	}
	if err != nil {
		t.Fatalf("expected no error; got %s", err)
	}
	if out[0] != 0xff {
		t.Fatalf("expected out[0] = 0xff; got 0x%x", out[0])
	}

	n, err = dec.Read(out[:])
	if n != 1 {
		t.Fatalf("expected n = %v; got %d", 1, n)
	}
	if err != nil {
		t.Fatalf("expected no error; got %s", err)
	}
	if out[0] != 0xee {
		t.Fatalf("expected out[0] = 0xee; got 0x%x", out[0])
	}

	_, err = dec.Read(out[:])
	if err != io.ErrUnexpectedEOF {
		t.Fatalf("expected ErrUnexpectedEOF; got %v", err)
	}
}

func TestDumper(t *testing.T) {
	var in [40]byte
	for i := range in {
		in[i] = byte(i + 30)
	}

	for stride := 1; stride < len(in); stride++ {
		var out bytes.Buffer
		dumper := Dumper(&out)
		done := 0
		for done < len(in) {
			todo := done + stride
			if todo > len(in) {
				todo = len(in)
			}
			dumper.Write(in[done:todo])
			done = todo
		}

		dumper.Close()
		if !bytes.Equal(out.Bytes(), expectedHexDump) {
			t.Errorf("stride: %d failed. got:\n%s\nwant:\n%s", stride, out.Bytes(), expectedHexDump)
		}
	}
}

func TestDump(t *testing.T) {
	var in [40]byte
	for i := range in {
		in[i] = byte(i + 30)
	}

	out := []byte(Dump(in[:]))
	if !bytes.Equal(out, expectedHexDump) {
		t.Errorf("got:\n%s\nwant:\n%s", out, expectedHexDump)
	}
}

var expectedHexDump = []byte(`00000000  1e 1f 20 21 22 23 24 25  26 27 28 29 2a 2b 2c 2d  |.. !"#$%&'()*+,-|
00000010  2e 2f 30 31 32 33 34 35  36 37 38 39 3a 3b 3c 3d  |./0123456789:;<=|
00000020  3e 3f 40 41 42 43 44 45                           |>?@ABCDE|
`)

var sink []byte

func BenchmarkEncode(b *testing.B) {
	for _, size := range []int{256, 1024, 4096, 16384} {
		src := bytes.Repeat([]byte{2, 3, 5, 7, 9, 11, 13, 17}, size/8)
		sink = make([]byte, 2*size)

		b.Run(fmt.Sprintf("%v", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				Encode(sink, src)
			}
		})
	}
}

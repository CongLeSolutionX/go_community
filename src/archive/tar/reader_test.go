// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tar

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

type multiFile struct {
	io.Reader
	fc []io.Closer
	fr []io.Reader
}

func openMultiFile(ps ...string) (io.ReadCloser, error) {
	mr := new(multiFile)
	for _, p := range ps {
		f, err := os.Open(p)
		if err != nil {
			mr.Close()
			return nil, err
		}
		mr.fc = append(mr.fc, f)
		mr.fr = append(mr.fr, f)
	}
	mr.Reader = io.MultiReader(mr.fr...)
	return mr, nil
}

func (mr *multiFile) Close() error {
	for _, f := range mr.fc {
		f.Close()
	}
	mr.fc = nil
	return nil
}

type untarTest struct {
	files   []string  // Test input is the logical concatenation of these
	headers []*Header // Expected output headers
	cksums  []string  // MD5 checksums of files, leave nil if not checked
	err     error     // Expected error to occur
}

var gnuTarTest = &untarTest{
	files: []string{"testdata/gnu.tar"},
	headers: []*Header{
		{
			Name:     "small.txt",
			Mode:     0640,
			Uid:      73025,
			Gid:      5000,
			Size:     5,
			ModTime:  time.Unix(1244428340, 0),
			Typeflag: '0',
			Uname:    "dsymonds",
			Gname:    "eng",
		},
		{
			Name:     "small2.txt",
			Mode:     0640,
			Uid:      73025,
			Gid:      5000,
			Size:     11,
			ModTime:  time.Unix(1244436044, 0),
			Typeflag: '0',
			Uname:    "dsymonds",
			Gname:    "eng",
		},
	},
	cksums: []string{
		"e38b27eaccb4391bdec553a7f3ae6b2f",
		"c65bd2e50a56a2138bf1716f2fd56fe9",
	},
}

var sparseTarTest = &untarTest{
	files: []string{"testdata/sparse-formats.tar"},
	headers: []*Header{
		{
			Name:     "sparse-gnu",
			Mode:     420,
			Uid:      1000,
			Gid:      1000,
			Size:     200,
			ModTime:  time.Unix(1392395740, 0),
			Typeflag: 0x53,
			Linkname: "",
			Uname:    "david",
			Gname:    "david",
			Devmajor: 0,
			Devminor: 0,
		},
		{
			Name:     "sparse-posix-0.0",
			Mode:     420,
			Uid:      1000,
			Gid:      1000,
			Size:     200,
			ModTime:  time.Unix(1392342187, 0),
			Typeflag: 0x30,
			Linkname: "",
			Uname:    "david",
			Gname:    "david",
			Devmajor: 0,
			Devminor: 0,
		},
		{
			Name:     "sparse-posix-0.1",
			Mode:     420,
			Uid:      1000,
			Gid:      1000,
			Size:     200,
			ModTime:  time.Unix(1392340456, 0),
			Typeflag: 0x30,
			Linkname: "",
			Uname:    "david",
			Gname:    "david",
			Devmajor: 0,
			Devminor: 0,
		},
		{
			Name:     "sparse-posix-1.0",
			Mode:     420,
			Uid:      1000,
			Gid:      1000,
			Size:     200,
			ModTime:  time.Unix(1392337404, 0),
			Typeflag: 0x30,
			Linkname: "",
			Uname:    "david",
			Gname:    "david",
			Devmajor: 0,
			Devminor: 0,
		},
		{
			Name:     "end",
			Mode:     420,
			Uid:      1000,
			Gid:      1000,
			Size:     4,
			ModTime:  time.Unix(1392398319, 0),
			Typeflag: 0x30,
			Linkname: "",
			Uname:    "david",
			Gname:    "david",
			Devmajor: 0,
			Devminor: 0,
		},
	},
	cksums: []string{
		"6f53234398c2449fe67c1812d993012f",
		"6f53234398c2449fe67c1812d993012f",
		"6f53234398c2449fe67c1812d993012f",
		"6f53234398c2449fe67c1812d993012f",
		"b0061974914468de549a2af8ced10316",
	},
}

var untarTests = []*untarTest{
	gnuTarTest,
	sparseTarTest,
	{
		files: []string{"testdata/ustar.tar"},
		headers: []*Header{
			{
				Name:     strings.Repeat("longname/", 15) + "file.txt",
				Mode:     0644,
				Uid:      501,
				Gid:      20,
				Size:     6,
				ModTime:  time.Unix(1360135598, 0),
				Typeflag: '0',
				Uname:    "shane",
				Gname:    "staff",
			},
		},
	},
	{
		files: []string{"testdata/star.tar"},
		headers: []*Header{
			{
				Name:       "small.txt",
				Mode:       0640,
				Uid:        73025,
				Gid:        5000,
				Size:       5,
				ModTime:    time.Unix(1244592783, 0),
				Typeflag:   '0',
				Uname:      "dsymonds",
				Gname:      "eng",
				AccessTime: time.Unix(1244592783, 0),
				ChangeTime: time.Unix(1244592783, 0),
			},
			{
				Name:       "small2.txt",
				Mode:       0640,
				Uid:        73025,
				Gid:        5000,
				Size:       11,
				ModTime:    time.Unix(1244592783, 0),
				Typeflag:   '0',
				Uname:      "dsymonds",
				Gname:      "eng",
				AccessTime: time.Unix(1244592783, 0),
				ChangeTime: time.Unix(1244592783, 0),
			},
		},
	},
	{
		files: []string{"testdata/star-prefix.tar"},
		headers: []*Header{
			{
				Name:       "prefix/small.txt",
				Mode:       0640,
				Uid:        73025,
				Gid:        5000,
				Size:       5,
				ModTime:    time.Unix(1244592783, 0),
				Typeflag:   '0',
				Uname:      "dsymonds",
				Gname:      "eng",
				AccessTime: time.Unix(1244592783, 0),
				ChangeTime: time.Unix(1244592783, 0),
			},
		},
	},
	{
		files: []string{"testdata/v7.tar"},
		headers: []*Header{
			{
				Name:     "small.txt",
				Mode:     0444,
				Uid:      73025,
				Gid:      5000,
				Size:     5,
				ModTime:  time.Unix(1244593104, 0),
				Typeflag: '\x00',
			},
			{
				Name:     "small2.txt",
				Mode:     0444,
				Uid:      73025,
				Gid:      5000,
				Size:     11,
				ModTime:  time.Unix(1244593104, 0),
				Typeflag: '\x00',
			},
		},
	},
	{
		files: []string{"testdata/pax.tar"},
		headers: []*Header{
			{
				Name:       "a/123456789101112131415161718192021222324252627282930313233343536373839404142434445464748495051525354555657585960616263646566676869707172737475767778798081828384858687888990919293949596979899100",
				Mode:       0664,
				Uid:        1000,
				Gid:        1000,
				Uname:      "shane",
				Gname:      "shane",
				Size:       7,
				ModTime:    time.Unix(1350244992, 23960108),
				ChangeTime: time.Unix(1350244992, 23960108),
				AccessTime: time.Unix(1350244992, 23960108),
				Typeflag:   TypeReg,
			},
			{
				Name:       "a/b",
				Mode:       0777,
				Uid:        1000,
				Gid:        1000,
				Uname:      "shane",
				Gname:      "shane",
				Size:       0,
				ModTime:    time.Unix(1350266320, 910238425),
				ChangeTime: time.Unix(1350266320, 910238425),
				AccessTime: time.Unix(1350266320, 910238425),
				Typeflag:   TypeSymlink,
				Linkname:   "123456789101112131415161718192021222324252627282930313233343536373839404142434445464748495051525354555657585960616263646566676869707172737475767778798081828384858687888990919293949596979899100",
			},
		},
	},
	{
		files: []string{"testdata/nil-uid.tar"}, // golang.org/issue/5290
		headers: []*Header{
			{
				Name:     "P1050238.JPG.log",
				Mode:     0664,
				Uid:      0,
				Gid:      0,
				Size:     14,
				ModTime:  time.Unix(1365454838, 0),
				Typeflag: TypeReg,
				Linkname: "",
				Uname:    "eyefi",
				Gname:    "eyefi",
				Devmajor: 0,
				Devminor: 0,
			},
		},
	},
	{
		files: []string{"testdata/xattrs.tar"},
		headers: []*Header{
			{
				Name:       "small.txt",
				Mode:       0644,
				Uid:        1000,
				Gid:        10,
				Size:       5,
				ModTime:    time.Unix(1386065770, 448252320),
				Typeflag:   '0',
				Uname:      "alex",
				Gname:      "wheel",
				AccessTime: time.Unix(1389782991, 419875220),
				ChangeTime: time.Unix(1389782956, 794414986),
				Xattrs: map[string]string{
					"user.key":  "value",
					"user.key2": "value2",
					// Interestingly, selinux encodes the terminating null inside the xattr
					"security.selinux": "unconfined_u:object_r:default_t:s0\x00",
				},
			},
			{
				Name:       "small2.txt",
				Mode:       0644,
				Uid:        1000,
				Gid:        10,
				Size:       11,
				ModTime:    time.Unix(1386065770, 449252304),
				Typeflag:   '0',
				Uname:      "alex",
				Gname:      "wheel",
				AccessTime: time.Unix(1389782991, 419875220),
				ChangeTime: time.Unix(1386065770, 449252304),
				Xattrs: map[string]string{
					"security.selinux": "unconfined_u:object_r:default_t:s0\x00",
				},
			},
		},
	},
	{
		files: []string{"testdata/gnu-file-atime.tar"},
		headers: []*Header{
			{
				Name:       "test2/",
				Mode:       040755,
				Uid:        1000,
				Gid:        1000,
				Size:       14,
				ModTime:    time.Unix(1441973427, 0),
				Typeflag:   'D',
				Uname:      "rawr",
				Gname:      "dsnet",
				AccessTime: time.Unix(1441974501, 0),
				ChangeTime: time.Unix(1441973436, 0),
			},
			{
				Name:       "test2/foo",
				Mode:       0100644,
				Uid:        1000,
				Gid:        1000,
				Size:       64,
				ModTime:    time.Unix(1441973363, 0),
				Typeflag:   '0',
				Uname:      "rawr",
				Gname:      "dsnet",
				AccessTime: time.Unix(1441974501, 0),
				ChangeTime: time.Unix(1441973436, 0),
			},
			{
				Name:       "test2/sparse",
				Mode:       0100644,
				Uid:        1000,
				Gid:        1000,
				Size:       536870912,
				ModTime:    time.Unix(1441973427, 0),
				Typeflag:   'S',
				Uname:      "rawr",
				Gname:      "dsnet",
				AccessTime: time.Unix(1441991948, 0),
				ChangeTime: time.Unix(1441973436, 0),
			},
		},
	},
	{
		files: []string{
			"testdata/ustar-long-path.tar",
			"testdata/ustar-file684.tar",
		},
		headers: []*Header{
			{
				Name:     "GNU1/GNU1/long-path-name",
				Mode:     0640,
				Uid:      319973,
				Gid:      5000,
				Size:     684,
				ModTime:  time.Unix(1442282516, 0),
				Typeflag: '0',
				Uname:    "joetsai",
				Gname:    "eng",
			},
		},
	},
	{
		files: []string{
			"testdata/pax6-pos-size.tar",
			"testdata/gnu-file684.tar",
		},
		headers: []*Header{
			{
				Name:     "foo",
				Mode:     0640,
				Uid:      319973,
				Gid:      5000,
				Size:     999,
				ModTime:  time.Unix(1442282516, 0),
				Typeflag: '0',
				Uname:    "joetsai",
				Gname:    "eng",
			},
		},
		cksums: []string{
			"0afb597b283fe61b5d4879669a350556",
		},
	},
	{
		files: []string{"testdata/sparse-gnu.tar"},
		headers: []*Header{
			{
				Name:     "sparse.db",
				Mode:     0640,
				Uid:      319973,
				Gid:      5000,
				Size:     8589934592,
				ModTime:  time.Unix(1442271680, 0),
				Typeflag: 'S',
				Uname:    "joetsai",
				Gname:    "eng",
			},
		},
	},
	{
		files: []string{"testdata/sparse-pax.tar"},
		headers: []*Header{
			{
				Name:       "sparse.db",
				Mode:       0640,
				Uid:        319973,
				Gid:        5000,
				Size:       8589934592,
				ModTime:    time.Unix(1442271680, 531154408),
				Typeflag:   '0',
				Uname:      "joetsai",
				Gname:      "eng",
				AccessTime: time.Unix(1442271699, 478914344),
				ChangeTime: time.Unix(1442271680, 531154408),
			},
		},
	},
	{
		files: []string{"testdata/gnu-bad-size.tar"},
		err:   ErrHeader,
	},
	{
		files: []string{"testdata/gnu-bad-mode.tar"},
		err:   ErrHeader,
	},
	{
		files: []string{"testdata/gnu1a-trunc-path.tar"},
		err:   io.ErrUnexpectedEOF,
	},
	{
		files: []string{"testdata/sparse-gnu-bad-magic.tar"},
		err:   ErrHeader,
	},
	{
		files: []string{"testdata/sparse-gnu-neg-size.tar"},
		err:   ErrHeader,
	},
	{
		files: []string{"testdata/sparse-pax-neg-size.tar"},
		err:   ErrHeader,
	},
	{
		files: []string{"testdata/neg-size.tar"},
		err:   ErrHeader,
	},
	{
		files: []string{
			"testdata/pax7-neg-size.tar",
			"testdata/ustar-file684.tar",
		},
		err: ErrHeader,
	},
	{
		files: []string{
			"testdata/pax8-bad-hdr.tar",
			"testdata/ustar-file684.tar",
		},
		err: ErrHeader,
	},
	{
		files: []string{
			"testdata/pax9-bad-mtime.tar",
			"testdata/ustar-file684.tar",
		},
		err: ErrHeader,
	},
	{
		files: []string{"testdata/issue10968.tar"},
		err:   ErrHeader,
	},
	{
		files: []string{"testdata/issue11169.tar"},
		err:   ErrHeader,
	},
	{
		files: []string{"testdata/issue12435.tar"},
		err:   ErrHeader,
	},
	{
		files: []string{"testdata/issue12436.tar"},
		err:   ErrHeader,
	},
	{
		files: []string{"testdata/issue12557.tar"},
		headers: []*Header{
			{
				Name:     "aaa",
				Mode:     0644,
				Uid:      1000,
				Gid:      1000,
				Size:     46,
				ModTime:  time.Unix(1441816230, 0),
				Typeflag: '0',
				Uname:    "rawr",
				Gname:    "dsnet",
			},
		},
		err: io.ErrUnexpectedEOF,
	},
}

func TestReader(t *testing.T) {
	for i, test := range untarTests {
		var err error
		var hdr *Header

		f, err := openMultiFile(test.files...)
		if err != nil {
			t.Errorf("test %d: Unexpected error: %v", i, err)
			continue
		}
		defer f.Close()

		tr := NewReader(f)
		var hdrs []*Header
		var cksums []string
		var rdbuf = make([]uint8, 8)
		for {
			// Parse the header.
			hdr, err = tr.Next()
			if err != nil {
				if err == io.EOF {
					err = nil // Expected error
				}
				break
			}
			hdrs = append(hdrs, hdr)

			// If cksums is not nil, then compute the cksum as well.
			if test.cksums == nil {
				continue
			}
			h := md5.New()
			_, err = io.CopyBuffer(h, tr, rdbuf)
			if err != nil {
				break
			}
			cksums = append(cksums, fmt.Sprintf("%x", h.Sum(nil)))
		}

		for j, hdr := range hdrs {
			if j == len(test.headers) {
				t.Errorf("test %d, entry %d: Unexpected header:\nhave %+v", i, j, *hdr)
				continue
			}

			if !reflect.DeepEqual(*hdr, *test.headers[j]) {
				t.Log(hdr.ModTime.Unix())
				t.Log(hdr.AccessTime.Unix())
				t.Log(hdr.ChangeTime.Unix())
				t.Log(hdr.ModTime.Nanosecond())
				t.Log(hdr.AccessTime.Nanosecond())
				t.Log(hdr.ChangeTime.Nanosecond())
				t.Errorf("test %d, entry %d: Incorrect header:\nhave %+v\nwant %+v",
					i, j, *hdr, *test.headers[j])
			}
		}

		for j, sum := range cksums {
			if j == len(test.cksums) {
				t.Errorf("test %d, entry %d: Unexpected sum:\nhave %v", i, j, sum)
				continue
			}

			if sum != test.cksums[j] {
				t.Errorf("test %d, entry %d: Incorrect sum:\nhave %v\nwant %v",
					i, j, sum, test.cksums[j])
			}
		}

		if err != test.err {
			t.Errorf("test %d: Wrong error: got %v, want %v", i, err, test.err)
		}

		f.Close()
	}
}

func TestPartialRead(t *testing.T) {
	f, err := os.Open("testdata/gnu.tar")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	defer f.Close()

	var entries = []struct {
		cnt    int    // Number of bytes to read
		output string // Expected value of string read
	}{
		{4, "Kilt"},
		{6, "Google"},
	}

	tr := NewReader(f)
	for i, e := range entries {
		hdr, err := tr.Next()
		if err != nil || hdr == nil {
			t.Fatalf("Didn't get file %d: %v", i, err)
		}
		buf := make([]byte, e.cnt)
		if _, err := io.ReadFull(tr, buf); err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if string(buf) != e.output {
			t.Errorf("Contents = %q, want %q", string(buf), e.output)
		}
	}

	_, err = tr.Next()
	if err != io.EOF {
		t.Errorf("Unexpected error: got %v, want %v", err, io.EOF)
	}
}

func TestParsePAXHeader(t *testing.T) {
	var vectors = []struct {
		input  string            // Input data
		inMap  map[string]string // Initial input map
		output map[string]string // Expected output map
		err    error             // Expected error outcome of parsing
	}{
		{"", nil, nil, nil},
		{"1 k=1\n", nil, nil, ErrHeader},
		{"6 k=1\n", nil, map[string]string{"k": "1"}, nil},
		{"6 k~1\n", nil, nil, ErrHeader},
		{"6_k=1\n", nil, nil, ErrHeader},
		{"6 k=1 ", nil, nil, ErrHeader},
		{"632 k=1\n", nil, nil, ErrHeader},
		{"16 longkeyname=hahaha\n", nil, nil, ErrHeader},
		{"27 ☺☻☹=日a本b語ç\n", nil, map[string]string{"☺☻☹": "日a本b語ç"}, nil},
		{"10 a=name\n", nil, map[string]string{"a": "name"}, nil},
		{"9 a=name\n", nil, map[string]string{"a": "name"}, nil},
		{"30 mtime=1350244992.023960108\n", nil,
			map[string]string{"mtime": "1350244992.023960108"}, nil},
		{"3 somelongkey=\n", nil, nil, ErrHeader},
		{"50 tooshort=\n", nil, nil, ErrHeader},
		{"23 GNU.sparse.offset=0\n25 GNU.sparse.numbytes=1\n" +
			"23 GNU.sparse.offset=2\n25 GNU.sparse.numbytes=3\n", nil,
			map[string]string{"GNU.sparse.map": "0,1,2,3"}, nil},
		{"25 GNU.sparse.numbytes=0\n23 GNU.sparse.offset=1\n" +
			"25 GNU.sparse.numbytes=2\n23 GNU.sparse.offset=3\n", nil, nil, ErrHeader},
		{"7 key=\n", nil, nil, nil},
		{"7 key=\n", map[string]string{"key": "hahaha"}, nil, nil},
		{"13 key2=haha\n", map[string]string{"key": "hahaha"},
			map[string]string{"key": "hahaha", "key2": "haha"}, nil},
		{"13 key1=haha\n13 key2=nana\n13 key3=kaka\n", map[string]string{"key2": ""},
			map[string]string{"key1": "haha", "key2": "nana", "key3": "kaka"}, nil},
		{"13 key1=haha\n8 key1=\n13 key3=kaka\n", map[string]string{"key2": "nana"},
			map[string]string{"key2": "nana", "key3": "kaka"}, nil},
	}

	for i, v := range vectors {
		r := bytes.NewReader([]byte(v.input))
		extHdrs, err := parsePAX(r, v.inMap)
		if !reflect.DeepEqual(extHdrs, v.output) && !(len(extHdrs) == 0 && len(v.output) == 0) {
			t.Errorf("test %d, parsePAX(...): got %v, want %v", i, extHdrs, v.output)
		}
		if err != v.err {
			t.Errorf("test %d, unexpected error: got %v, want %v", i, err, v.err)
		}
	}
}

func TestParsePAXTime(t *testing.T) {
	var vectors = []struct {
		input    string    // Input string
		output   time.Time // Expected output time if parsed
		parsible bool      // Expected error outcome of parsing
	}{
		{"1350244992.023960108", time.Unix(1350244992, 23960108), true},
		{"1350244992.02396010", time.Unix(1350244992, 23960100), true},
		{"1350244992.0239601089", time.Unix(1350244992, 23960108), true},
		{"1350244992", time.Unix(1350244992, 0), true},
		{"-1.3", time.Unix(-2, 7E8), true},
		{"0.0", time.Unix(0, 0), true},
		{"0", time.Unix(0, 0), true},
		{"1.", time.Unix(1, 0), true},
		{"", time.Unix(0, 0), true},
		{".", time.Unix(0, 0), true},
		{".5", time.Unix(0, 5E8), true},
		{"-", time.Time{}, false},
		{"+", time.Time{}, false},
		{"-1.-1", time.Time{}, false},
		{"99999999999999999999999999999999999999999999999", time.Time{}, false},
		{"0.123456789abcdef", time.Time{}, false},
	}

	for _, v := range vectors {
		ts, err := parsePAXTime(v.input)
		parsible := (err == nil)
		if v.parsible != parsible {
			if v.parsible {
				t.Errorf("parsePAXTime(%q): want parsing success, but got error", v.input)
			} else {
				t.Errorf("parsePAXTime(%q): want parsing error, but got success", v.input)
			}
		}
		if parsible && !ts.Equal(v.output) {
			t.Errorf("parsePAXTime(%q) = %d.%9d, want %d.%9d",
				v.input, ts.Unix(), ts.Nanosecond(), v.output.Unix(), v.output.Nanosecond())
		}
	}
}

func TestMergePAX(t *testing.T) {
	var vectors = []struct {
		extHdrs map[string]string // The input headers
		header  *Header           // The expected output
		err     error             // The expected error
	}{{
		extHdrs: map[string]string{
			"path":  "a/b/c",
			"uid":   "1000",
			"mtime": "1350244992.023960108",
		},
		header: &Header{
			Name:    "a/b/c",
			Uid:     1000,
			ModTime: time.Unix(1350244992, 23960108),
		},
	}, {
		extHdrs: map[string]string{"gid": "gtgergergersagersgers"},
		err:     ErrHeader,
	}}

	for i, v := range vectors {
		hdr := new(Header)
		err := mergePAX(hdr, v.extHdrs)
		if err != v.err {
			t.Errorf("test %d, unexpected error: got %v, want %v", i, err, v.err)
		}
		if v.err == nil && !reflect.DeepEqual(*hdr, *v.header) {
			t.Errorf("test %d, mergePAX(...): got %v, want %v", i, *hdr, *v.header)
		}
	}
}

func TestSparseFileReader(t *testing.T) {
	var vectors = []struct {
		sparseMap  []sparseEntry // The input sparse map
		realSize   int64         // The real size of the output file
		sparseData string        // The input compact data
		expected   string        // The expected output data
		err        error         // The expected error outcome
	}{
		{[]sparseEntry{{0, 2}, {5, 3}}, 8, "abcde", "ab\x00\x00\x00cde", nil},
		{[]sparseEntry{{0, 2}, {5, 3}}, 10, "abcde", "ab\x00\x00\x00cde\x00\x00", nil},
		{[]sparseEntry{{1, 3}, {6, 2}}, 8, "abcde", "\x00abc\x00\x00de", nil},
		{[]sparseEntry{{1, 3}, {6, 0}, {6, 0}, {6, 2}}, 8, "abcde", "\x00abc\x00\x00de", nil},
		{[]sparseEntry{{1, 3}, {6, 2}}, 10, "abcde", "\x00abc\x00\x00de\x00\x00", nil},
		{[]sparseEntry{{1, 3}, {6, 2}, {8, 0}, {8, 0}, {8, 0}, {8, 0}}, 10, "abcde", "\x00abc\x00\x00de\x00\x00", nil},
		{nil, 2, "", "\x00\x00", nil},
		{nil, -2, "", "", ErrHeader},
		{[]sparseEntry{{1, 3}, {6, 2}}, -10, "abcde", "", ErrHeader},
		{[]sparseEntry{{1, 3}, {6, 5}}, 10, "abcde", "", ErrHeader},
		{[]sparseEntry{{1, 3}, {6, 5}}, 35, "abcde", "", io.ErrUnexpectedEOF},
		{[]sparseEntry{{1, 3}, {6, -5}}, 35, "abcde", "", ErrHeader},
		{[]sparseEntry{{math.MaxInt64, 3}, {6, -5}}, 35, "abcde", "", ErrHeader},
		{[]sparseEntry{{1, 3}, {2, 2}}, 10, "abcde", "", ErrHeader},
	}

	for i, v := range vectors {
		r := bytes.NewReader([]byte(v.sparseData))
		rfr, _ := newRegFileReader(r, int64(len(v.sparseData)))

		var sfr *sparseFileReader
		var err error
		var buf []byte

		sfr, err = newSparseFileReader(rfr, v.sparseMap, v.realSize)
		if err != nil {
			goto check
		}
		if sfr.numBytes() != int64(len(v.sparseData)) {
			t.Errorf("test %d, sfr.numBytes() before reading: got %d, want %d", i, sfr.numBytes(), len(v.sparseData))
		}
		buf, err = ioutil.ReadAll(sfr)
		if err != nil {
			goto check
		}
		if string(buf) != v.expected {
			t.Errorf("test %d, ioutil.ReadAll(sfr): got %q, want %q", i, string(buf), v.expected)
		}
		if sfr.numBytes() != 0 {
			t.Errorf("test %d, sfr.numBytes() after reading: got %d, want %d", i, sfr.numBytes(), len(v.sparseData))
		}

	check:
		if err != v.err {
			t.Errorf("test %d, unexpected error: got %v, want %v", i, err, v.err)
		}
	}
}

func TestSparsePartialRead(t *testing.T) {
	f, err := os.Open("testdata/sparse-formats.tar")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	defer f.Close()

	var entries = []struct {
		cnt    int    // Number of bytes to read
		output string // Expected value of string read
	}{
		{2, "\x00G"},
		{4, "\x00G\x00o"},
		{6, "\x00G\x00o\x00G"},
		{8, "\x00G\x00o\x00G\x00o"},
		{4, "end\n"},
	}

	tr := NewReader(f)
	for i, e := range entries {
		hdr, err := tr.Next()
		if err != nil || hdr == nil {
			t.Fatalf("Didn't get file %d: %v", i, err)
		}
		buf := make([]byte, e.cnt)
		if _, err := io.ReadFull(tr, buf); err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if string(buf) != e.output {
			t.Errorf("Contents = %q, want %q", string(buf), e.output)
		}
	}

	_, err = tr.Next()
	if err != io.EOF {
		t.Errorf("Unexpected error: got %v, want %v", err, io.EOF)
	}
}

func TestReadOldGNUSparseMap(t *testing.T) {
	var makeRawHeader = func(size, sp0, sp1, sp2, sp3, ext string) []byte {
		blk := make([]byte, blockSize)
		copy(blk[483:], size)
		copy(blk[386+00:], sp0)
		copy(blk[386+24:], sp1)
		copy(blk[386+48:], sp2)
		copy(blk[386+72:], sp3)
		copy(blk[482:], ext)
		return blk
	}
	const (
		t00 = "00000000000\x0000000000000\x00"
		t11 = "00000000001\x0000000000001\x00"
		t12 = "00000000001\x0000000000002\x00"
		t21 = "00000000002\x0000000000001\x00"
	)

	var vectors = []struct {
		data   string        // Input data
		format int           // Input format
		hdr    *Header       // Output header (passed by reference)
		rawHdr []byte        // Input raw header
		sp     []sparseEntry // Expected sparse entries to be outputted
		err    error         // Expected errors that may be raised
	}{{
		"", formatUnknown, nil, nil, nil, ErrHeader,
	}, {
		"", formatGNU, new(Header), makeRawHeader("-1", "", "", "", "", ""), nil, ErrHeader,
	}, {
		"", formatGNU, new(Header),
		makeRawHeader("1234", "fewa", "", "", "", ""), nil, ErrHeader,
	}, {
		"", formatGNU, new(Header),
		makeRawHeader("1234", t00, "", "", "", ""), nil, nil,
	}, {
		"", formatGNU, new(Header),
		makeRawHeader("1234", t11, t12, t21, t11, ""),
		[]sparseEntry{{1, 1}, {1, 2}, {2, 1}, {1, 1}}, nil,
	}, {
		"", formatGNU, new(Header),
		makeRawHeader("1234", t11, t12, t21, t11, "\x80"),
		nil, io.ErrUnexpectedEOF,
	}, {
		t11 + t11 + t00, formatGNU, new(Header),
		makeRawHeader("1234", t11, t12, t21, t11, "\x80"),
		nil, io.ErrUnexpectedEOF,
	}, {
		t11 + t21 + t00 + strings.Repeat("\x00", 512), formatGNU, new(Header),
		makeRawHeader("1234", t11, t12, t21, t11, "\x80"),
		[]sparseEntry{{1, 1}, {1, 2}, {2, 1}, {1, 1}, {1, 1}, {2, 1}}, nil,
	}}

	for i, v := range vectors {
		r := bytes.NewReader([]byte(v.data))
		tr := new(Reader)
		sp, err := tr.readOldGNUSparseMap(r, v.format, v.hdr, v.rawHdr)
		if !reflect.DeepEqual(sp, v.sp) {
			t.Errorf("test %d, readOldGNUSparseMap(...): got %v, want %v", i, sp, v.sp)
		}
		if err != v.err {
			t.Errorf("test %d, unexpected error: got %v, want %v", i, err, v.err)
		}
	}
}

func TestReadGNUSparseMap0x1(t *testing.T) {
	const (
		maxUint = ^uint(0)
		maxInt  = int(maxUint >> 1)
	)
	var big1 = fmt.Sprintf("%d", int64(maxInt))
	var big2 = fmt.Sprintf("%d", (int64(maxInt)/2)+1)
	var big3 = fmt.Sprintf("%d", (int64(maxInt) / 3))

	var vectors = []struct {
		extHdrs map[string]string // Input data
		sp      []sparseEntry     // Expected sparse entries to be outputted
		err     error             // Expected errors that may be raised
	}{{
		map[string]string{paxGNUSparseNumBlocks: "-4"}, nil, ErrHeader,
	}, {
		map[string]string{paxGNUSparseNumBlocks: "fee "}, nil, ErrHeader,
	}, {
		map[string]string{
			paxGNUSparseNumBlocks: big1,
			paxGNUSparseMap:       "0,5,10,5,20,5,30,5",
		},
		nil, ErrHeader,
	}, {
		map[string]string{
			paxGNUSparseNumBlocks: big2,
			paxGNUSparseMap:       "0,5,10,5,20,5,30,5",
		},
		nil, ErrHeader,
	}, {
		map[string]string{
			paxGNUSparseNumBlocks: big3,
			paxGNUSparseMap:       "0,5,10,5,20,5,30,5",
		},
		nil, ErrHeader,
	}, {
		map[string]string{
			paxGNUSparseNumBlocks: "4",
			paxGNUSparseMap:       "0.5,5,10,5,20,5,30,5",
		},
		nil, ErrHeader,
	}, {
		map[string]string{
			paxGNUSparseNumBlocks: "4",
			paxGNUSparseMap:       "0,5.5,10,5,20,5,30,5",
		},
		nil, ErrHeader,
	}, {
		map[string]string{
			paxGNUSparseNumBlocks: "4",
			paxGNUSparseMap:       "0,5,10,5,20,5,30,5",
		},
		[]sparseEntry{{0, 5}, {10, 5}, {20, 5}, {30, 5}}, nil,
	}}

	for i, v := range vectors {
		sp, err := readGNUSparseMap0x1(v.extHdrs)
		if !reflect.DeepEqual(sp, v.sp) {
			t.Errorf("test %d, readGNUSparseMap0x1(...): got %v, want %v", i, sp, v.sp)
		}
		if err != v.err {
			t.Errorf("test %d, unexpected error: got %v, want %v", i, err, v.err)
		}
	}
}

func TestReadGNUSparseMap1x0(t *testing.T) {
	var sr = strings.Repeat
	var r = []sparseEntry{{1, 2}, {3, 4}}
	for i := 0; i < 98; i++ {
		r = append(r, sparseEntry{54321, 12345})
	}

	var vectors = []struct {
		input string        // Input data
		cnt   int           // Expected number of bytes read
		sp    []sparseEntry // Expected sparse entries to be outputted
		err   error         // Expected errors that may be raised
	}{
		{"", 0, nil, io.ErrUnexpectedEOF},
		{"ab", 2, nil, io.ErrUnexpectedEOF},
		{sr("\x00", 512), 512, nil, io.ErrUnexpectedEOF},
		{sr("\x00", 511) + "\n", 512, nil, ErrHeader},
		{sr("\n", 512), 512, nil, ErrHeader},
		{"0\n" + sr("\x00", 510), 512, nil, nil},
		{sr("0", 512) + "0\n" + sr("\x00", 510), 1024, nil, nil},
		{sr("0", 1024) + "1\n2\n3\n" + sr("\x00", 506), 1536, []sparseEntry{{2, 3}}, nil},
		{sr("0", 1024) + "1\n2\n\n" + sr("\x00", 509), 1536, nil, ErrHeader},
		{sr("0", 1024) + "1\n2\n" + sr("\x00", 508), 1536, nil, io.ErrUnexpectedEOF},
		{"-1\n2\n\n" + sr("\x00", 506), 512, nil, ErrHeader},
		{"1\nk\n2\n" + sr("\x00", 506), 512, nil, ErrHeader},
		{"100\n1\n2\n3\n4\n" + sr("54321\n0000000000000012345\n", 98) + sr("\x00", 512), 2560, r, nil},
	}

	for i, v := range vectors {
		r := bytes.NewReader([]byte(v.input))
		sp, err := readGNUSparseMap1x0(r)
		if !reflect.DeepEqual(sp, v.sp) && !(len(sp) == 0 && len(v.sp) == 0) {
			t.Errorf("test %d, readGNUSparseMap1x0(...): got %v, want %v", i, sp, v.sp)
		}
		if numBytes := len(v.input) - r.Len(); numBytes != v.cnt {
			t.Errorf("test %d, numBytes after reading: got %v, want %v", i, numBytes, v.cnt)
		}
		if err != v.err {
			t.Errorf("test %d, unexpected error: got %v, want %v", i, err, v.err)
		}
	}
}

func TestUninitializedRead(t *testing.T) {
	test := gnuTarTest
	f, err := openMultiFile(test.files...)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	defer f.Close()

	tr := NewReader(f)
	_, err = tr.Read([]byte{0x00})
	if err != io.EOF {
		t.Errorf("Unexpected error: %v, wanted %v", err, io.EOF)
	}
}

// Test the ending condition on various truncated files and that truncated files
// are still detected even if the underlying io.Reader satisfies io.Seeker.
func TestStreamTruncation(t *testing.T) {
	var ss []string
	for _, p := range []string{
		"testdata/gnu.tar",
		"testdata/ustar-file684.tar",
		"testdata/pax1-path1.tar",
		"testdata/sparse-formats.tar",
	} {
		buf, err := ioutil.ReadFile(p)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		ss = append(ss, string(buf))
	}

	data1, data2, pax, sparse := ss[0], ss[1], ss[2], ss[3]
	data2 += strings.Repeat("\x00", 10*512)
	trash := strings.Repeat("garbage ", 64) // Exactly 512 bytes

	var vectors = []struct {
		input string // Input stream
		cnt   int    // Expected number of headers read
		err   error  // Expected error outcome
	}{
		{"", 0, io.EOF}, // Empty file is a "valid" tar file
		{data1[:511], 0, io.ErrUnexpectedEOF},
		{data1[:512], 1, io.ErrUnexpectedEOF},
		{data1[:1024], 1, io.EOF},
		{data1[:1536], 2, io.ErrUnexpectedEOF},
		{data1[:2048], 2, io.EOF},
		{data1, 2, io.EOF},
		{data1[:2048] + data2[:1536], 3, io.EOF},
		{data2[:511], 0, io.ErrUnexpectedEOF},
		{data2[:512], 1, io.ErrUnexpectedEOF},
		{data2[:1195], 1, io.ErrUnexpectedEOF},
		{data2[:1196], 1, io.EOF}, // Exact end of data and start of padding
		{data2[:1200], 1, io.EOF},
		{data2[:1535], 1, io.EOF},
		{data2[:1536], 1, io.EOF}, // Exact end of padding
		{data2[:1536] + trash[:1], 1, io.ErrUnexpectedEOF},
		{data2[:1536] + trash[:511], 1, io.ErrUnexpectedEOF},
		{data2[:1536] + trash, 1, ErrHeader},
		{data2[:2048], 1, io.EOF}, // Exactly 1 empty block
		{data2[:2048] + trash[:1], 1, io.ErrUnexpectedEOF},
		{data2[:2048] + trash[:511], 1, io.ErrUnexpectedEOF},
		{data2[:2048] + trash, 1, ErrHeader},
		{data2[:2560], 1, io.EOF}, // Exactly 2 empty blocks (normal end-of-stream)
		{data2[:2560] + trash[:1], 1, io.EOF},
		{data2[:2560] + trash[:511], 1, io.EOF},
		{data2[:2560] + trash, 1, io.EOF},
		{data2[:3072], 1, io.EOF},
		{pax, 0, io.EOF}, // PAX header without data is a "valid" tar file
		{pax + trash[:1], 0, io.ErrUnexpectedEOF},
		{pax + trash[:511], 0, io.ErrUnexpectedEOF},
		{sparse[:511], 0, io.ErrUnexpectedEOF},
		{sparse[:512], 0, io.ErrUnexpectedEOF},
		{sparse[:3584], 1, io.EOF},
		{sparse[:9200], 1, io.EOF}, // Terminate in padding of sparse header
		{sparse[:9216], 1, io.EOF},
		{sparse[:9728], 2, io.ErrUnexpectedEOF},
		{sparse[:10240], 2, io.EOF},
		{sparse[:11264], 2, io.ErrUnexpectedEOF},
		{sparse, 5, io.EOF},
		{sparse + trash, 5, io.EOF},
	}

	type reader struct{ io.Reader }
	type readSeeker struct{ io.ReadSeeker }

	for _, v := range vectors {
		for i := 0; i < 4; i++ {
			var tr *Reader
			var s1, s2 string

			switch i {
			case 0:
				tr = NewReader(&reader{bytes.NewBuffer([]byte(v.input))})
				s1, s2 = "io.Reader", "auto"
			case 1:
				tr = NewReader(&reader{bytes.NewBuffer([]byte(v.input))})
				s1, s2 = "io.Reader", "manual"
			case 2:
				tr = NewReader(&readSeeker{bytes.NewReader([]byte(v.input))})
				s1, s2 = "io.SeekReader", "auto"
			case 3:
				tr = NewReader(&readSeeker{bytes.NewReader([]byte(v.input))})
				s1, s2 = "io.SeekReader", "manual"
			}

			var cnt int
			var err error
			for {
				if _, err = tr.Next(); err != nil {
					break
				}
				cnt++
				if s2 == "manual" {
					if _, err = io.Copy(ioutil.Discard, tr); err != nil {
						break
					}
				}
			}
			if err != v.err {
				t.Errorf("NewReader(%s(%d)) with %s discard: want %v, but got %v",
					s1, len(v.input), s2, v.err, err)
			}
			if cnt != v.cnt {
				t.Errorf("NewReader(%s(%d)) with %s discard: want %d, but got %d headers",
					s1, len(v.input), s2, v.cnt, cnt)
			}
		}
	}
}

// Test archives that use multiple meta headers together. This is acceptable
// when using GNU headers, in which the latest value takes precedence. However,
// when both PAX and GNU headers are mixed, the archive is non-compliant and the
// result in not specified. In Go's case, the latest value seen is used.
func TestMultipleHeaders(t *testing.T) {
	var vectors = []struct {
		files  []string // Input data is logical concatenation of all inputs
		header *Header  // The expected output header
	}{
		{
			[]string{
				"testdata/pax4-linkpath2.tar",
				"testdata/pax2-path2.tar",
				"testdata/pax3-linkpath1.tar",
				"testdata/pax1-path1.tar",
				"testdata/ustar-symlink.tar",
			},
			&Header{
				Name:     "PAX1/PAX1/long-path-name",
				Linkname: "PAX3/PAX3/long-linkpath-name",
			},
		},
		{
			[]string{
				"testdata/gnu4-linkpath2.tar",
				"testdata/gnu2-path2.tar",
				"testdata/gnu1-path1.tar",
				"testdata/gnu3-linkpath1.tar",
				"testdata/gnu-symlink.tar",
			},
			&Header{
				Name:     "GNU1/GNU1/long-path-name",
				Linkname: "GNU3/GNU3/long-linkpath-name",
			},
		},
		{
			[]string{
				"testdata/gnu4-linkpath2.tar",
				"testdata/pax4-linkpath2.tar",
				"testdata/gnu2-path2.tar",
				"testdata/pax1-path1.tar",
				"testdata/pax3-linkpath1.tar",
				"testdata/gnu3-linkpath1.tar",
				"testdata/gnu1-path1.tar",
				"testdata/pax2-path2.tar",
				"testdata/gnu-symlink.tar",
			},
			&Header{
				Name:     "PAX2/PAX2/long-path-name",
				Linkname: "GNU3/GNU3/long-linkpath-name",
			},
		},
		{
			[]string{
				"testdata/pax5-path-add-del.tar",
				"testdata/pax4-linkpath2.tar",
				"testdata/pax3-linkpath1.tar",
				"testdata/pax2-path2.tar",
				"testdata/pax1-path1.tar",
				"testdata/ustar-symlink.tar",
			},
			&Header{
				Name:     "PAX1/PAX1/long-path-name",
				Linkname: "PAX3/PAX3/long-linkpath-name",
			},
		},
		{
			[]string{
				"testdata/pax1-path1.tar",
				"testdata/pax3-linkpath1.tar",
				"testdata/pax4-linkpath2.tar",
				"testdata/pax2-path2.tar",
				"testdata/pax5-path-add-del.tar",
				"testdata/ustar-symlink.tar",
			},
			&Header{
				Name:     "bar",
				Linkname: "PAX4/PAX4/long-linkpath-name",
			},
		},
	}

	for i, v := range vectors {
		r, err := openMultiFile(v.files...)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		defer r.Close()

		tr := NewReader(r)
		hdr, err := tr.Next()
		if err != nil {
			t.Errorf("test %d: didn't get entry: %v", i, err)
		}

		// Set common fields and then check for equality.
		v.header.Typeflag = TypeSymlink
		v.header.ModTime = time.Unix(0, 0)
		if !reflect.DeepEqual(*hdr, *v.header) {
			t.Errorf("test %d: incorrect header:\ngot %+v\nwant %+v",
				i, *hdr, *v.header)
		}
	}
}

// Test that Reader does not attempt to read special header-only files.
func TestReadHeaderOnly(t *testing.T) {
	f, err := os.Open("testdata/special.tar")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	defer f.Close()

	var hdrs []*Header
	tr := NewReader(f)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		hdrs = append(hdrs, hdr)
	}

	// Without ignoring the size on header-only files, it would not be possible
	// to read all the entries.
	if len(hdrs) != 16 {
		t.Fatalf("len(hdrs): got %d, want %d", len(hdrs), 16)
	}

	for i := 0; i < 8; i++ {
		// Only difference between the two is that sizes are different.
		hdr1, hdr2 := hdrs[i], hdrs[i+8]
		hdr1.Size, hdr2.Size = 0, 0
		if !reflect.DeepEqual(*hdr1, *hdr2) {
			t.Errorf("incorrect header:\ngot %+v\nwant %+v", *hdr1, *hdr2)
		}
	}
}

// Test reading of various numeric fields in both octal and binary formats.
func TestReadNumeric(t *testing.T) {
	var vectors = []struct {
		input    string // Input data
		output   int64  // Expected output value if parsed
		parsible bool   // Expected error outcome of parsing
	}{
		// Test base-256 encoded values.
		{"", 0, true},
		{"\x80", 0, true},
		{"\x80\x00", 0, true},
		{"\x80\x00\x00", 0, true},
		{"\xbf", (1 << 6) - 1, true},
		{"\xbf\xff", (1 << 14) - 1, true},
		{"\xbf\xff\xff", (1 << 22) - 1, true},
		{"\xff", -1, true},
		{"\xff\xff", -1, true},
		{"\xff\xff\xff", -1, true},
		{"\xc0", -1 * (1 << 6), true},
		{"\xc0\x00", -1 * (1 << 14), true},
		{"\xc0\x00\x00", -1 * (1 << 22), true},
		{"\x87\x76\xa2\x22\xeb\x8a\x72\x61", 537795476381659745, true},
		{"\x80\x00\x00\x00\x07\x76\xa2\x22\xeb\x8a\x72\x61", 537795476381659745, true},
		{"\xf7\x76\xa2\x22\xeb\x8a\x72\x61", -615126028225187231, true},
		{"\xff\xff\xff\xff\xf7\x76\xa2\x22\xeb\x8a\x72\x61", -615126028225187231, true},
		{"\x80\x7f\xff\xff\xff\xff\xff\xff\xff", math.MaxInt64, true},
		{"\x80\x80\x00\x00\x00\x00\x00\x00\x00", 0, false},
		{"\xff\x80\x00\x00\x00\x00\x00\x00\x00", math.MinInt64, true},
		{"\xff\x7f\xff\xff\xff\xff\xff\xff\xff", 0, false},
		{"\xf5\xec\xd1\xc7\x7e\x5f\x26\x48\x81\x9f\x8f\x9b", 0, false},

		// Test octal encoded values.
		{"0000000\x00", 0, true},
		{"00000000227\x00", 0227, true},
		{"032033\x00 ", 032033, true},
		{"320330\x00 ", 0320330, true},
		{"0000660\x00 ", 0660, true},
		{"\x00\x000000660\x00 ", 0660, true},
		{"0123456789abcdef", 0, false},
	}

	var tr Reader
	for _, v := range vectors {
		tr.err = nil
		x := tr.numeric([]byte(v.input))
		parsible := (tr.err == nil)
		if v.parsible != parsible {
			if v.parsible {
				t.Errorf("Reader.numeric(%q): want parsing success, but got error", v.input)
			} else {
				t.Errorf("Reader.numeric(%q): want parsing error, but got success", v.input)
			}
		}
		if parsible && x != v.output {
			t.Errorf("Reader.numeric(%q) = %d, want %d", v.input, x, v.output)
		}
	}
}

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sync_test

import (
	"bytes"
	"compress/gzip"
	"io"
	"log"
	"os"
	"sync"
)

var readerPool = sync.Pool{
	New: func() interface{} {
		// The New function should generally only return pointer types since
		// a pointer can be stored in an interface value without an allocation.
		return new(gzip.Reader)
	},
}

func getReader(r io.Reader) (*gzip.Reader, error) {
	z := readerPool.Get().(*gzip.Reader)
	if err := z.Reset(r); err != nil {
		putReader(z)
		return nil, err
	}
	return z, nil
}

func putReader(z *gzip.Reader) {
	readerPool.Put(z)
}

func ExamplePool() {
	// Produce a compressed file with some testdata.
	b := new(bytes.Buffer)
	zw := gzip.NewWriter(b)
	if _, err := zw.Write([]byte("Hello, world!")); err != nil {
		log.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		log.Fatal(err)
	}

	// Obtain a new GZIP reader.
	// This may allocate a new reader or pull one from the sync.Pool.
	zr := readerPool.Get().(*gzip.Reader)
	defer readerPool.Put(zr)
	if err := zr.Reset(b); err != nil {
		log.Fatal(err)
	}

	if _, err := io.Copy(os.Stdout, zr); err != nil {
		log.Fatal(err)
	}

	// Output: Hello, world!
}

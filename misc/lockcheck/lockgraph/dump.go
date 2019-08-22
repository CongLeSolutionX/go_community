// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lockgraph

import (
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"io"
)

func Dump(w io.Writer, lg *Graph) error {
	w1 := base64.NewEncoder(base64.StdEncoding, w)
	//w2, err := zlib.NewWriterLevel(w1, zlib.BestCompression)
	w2, err := gzip.NewWriterLevel(w1, gzip.BestCompression)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(w2)
	if err := enc.Encode(lg); err != nil {
		return err
	}
	if err := w2.Close(); err != nil {
		return err
	}
	return w1.Close()
}

func Load(r io.Reader) (*Graph, error) {
	r1 := base64.NewDecoder(base64.StdEncoding, r)
	//r2, err := zlib.NewReader(r1)
	r2, err := gzip.NewReader(r1)
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(r2)
	var g Graph
	if err := dec.Decode(&g); err != nil {
		return nil, err
	}
	return &g, nil
}

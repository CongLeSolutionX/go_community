// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gif

import (
	"bytes"
	"image"
	"os"

	"testing"
)

func fuzz(t *testing.T, b []byte) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(b))
	if err != nil {
		return
	}
	if cfg.Width*cfg.Height > 1e6 {
		return
	}
	img, typ, err := image.Decode(bytes.NewReader(b))
	if err != nil || typ != "gif" {
		return
	}
	for q := 1; q <= 256; q++ {
		var w bytes.Buffer
		err := Encode(&w, img, &Options{NumColors: q})
		if err != nil {
			t.Fatalf("failed to encode valid image: %s", err)
		}
		img1, err := Decode(&w)
		if err != nil {
			t.Fatalf("failed to decode roundtripped image: %s", err)
		}
		got := img1.Bounds()
		want := img.Bounds()
		if !got.Eq(want) {
			t.Fatalf("roundtripped image bounds have changed, got: %s, want: %s", got, want)
		}
	}
}

func FuzzDecodeEmptySeed(f *testing.F) {
	f.Add([]byte{})

	f.Fuzz(fuzz)
}

func FuzzDecodeBasicSeed(f *testing.F) {
	b, err := os.ReadFile("../testdata/video-001.gif")
	if err != nil {
		f.Fatalf("failed to read seed: %s", err)
	}
	f.Add(b)

	f.Fuzz(fuzz)
}

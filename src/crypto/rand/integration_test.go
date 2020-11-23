// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rand_test

import (
	"crypto/rand"
	"fmt"
	"io"
	"testing"
)

type pwnReader struct {
	targetSize int
	hijacked   io.Reader
}

const backdoored = "This is not random content. Hmm."

func (pr *pwnReader) Read(b []byte) (n int, err error) {
	if len(b) == pr.targetSize {
		return copy(b, backdoored), nil
	} else {
		return pr.hijacked.Read(b)
	}
}

var _ io.Reader = (*pwnReader)(nil)

func init() {
	rand.Reader = &pwnReader{32, rand.Reader}
}

func TestSwappingReaderDoestNotChangeRead(t *testing.T) {
	sizes := []int{24, 32, 33}
	for _, size := range sizes {
		name := fmt.Sprintf("Size=%d", size)
		t.Run(name, func(t *testing.T) {
			b := make([]byte, size)
			_, _ = rand.Read(b)
			if string(b) == backdoored {
				t.Fatalf("Not randomized\nGot %q", b)
			}
		})
	}
}

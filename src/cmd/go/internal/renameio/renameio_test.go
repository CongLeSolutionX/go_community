// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package renameio

import (
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestConcurrentReadsAndWrites(t *testing.T) {
	dir, err := ioutil.TempDir("", "renameio")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "blob.bin")

	chunkLen := 64 << 10
	buf := make([]byte, chunkLen*2)
	rand.Read(buf)

	n := 1000
	var wg sync.WaitGroup
	for ; n > 0; n-- {
		wg.Add(1)
		go func() {
			defer wg.Done()

			time.Sleep(time.Duration(rand.Intn(100)) * time.Microsecond)
			offset := rand.Intn(len(buf) - chunkLen)
			if err := WriteFile(path, buf[offset:offset+chunkLen], 0666); err != nil {
				t.Error(err)
			}

			time.Sleep(time.Duration(rand.Intn(100)) * time.Microsecond)
			data, err := ioutil.ReadFile(path)
			if err != nil {
				t.Error(err)
			}
			if len(data) != chunkLen {
				t.Errorf("read %d bytes, but each write is a %d-byte file", len(data), chunkLen)
			}
		}()
	}

	wg.Wait()
}

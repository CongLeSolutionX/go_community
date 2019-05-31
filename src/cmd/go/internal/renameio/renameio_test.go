// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package renameio

import (
	"encoding/binary"
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

	const chunkWords = 8 << 10
	buf := make([]byte, 2*chunkWords*8)
	for i := uint64(0); i < 2*chunkWords; i++ {
		binary.LittleEndian.PutUint64(buf[i*8:], i)
	}

	n := 1000
	var wg sync.WaitGroup
	for ; n > 0; n-- {
		wg.Add(1)
		go func() {
			defer wg.Done()

			time.Sleep(time.Duration(rand.Intn(100)) * time.Microsecond)
			offset := rand.Intn(chunkWords)
			chunk := buf[offset*8 : (offset+chunkWords)*8]
			if err := WriteFile(path, chunk, 0666); err != nil {
				t.Error(err)
			}

			time.Sleep(time.Duration(rand.Intn(100)) * time.Microsecond)
			data, err := ioutil.ReadFile(path)
			if err != nil {
				t.Error(err)
				return
			}
			if len(data) != 8*chunkWords {
				t.Errorf("read %d bytes, but each write is a %d-byte file", len(data), 8*chunkWords)
				return
			}

			u := binary.LittleEndian.Uint64(data)
			for i := 1; i < chunkWords; i++ {
				next := binary.LittleEndian.Uint64(data[i*8:])
				if next != u+1 {
					t.Errorf("wrote sequential integers, but read integer out of sequence at offset %d", i)
					return
				}
				u = next
			}
		}()
	}

	wg.Wait()
}

// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package trace

import (
	"io"
	"runtime"
)

const allocTraceBatch = 32 << 10

var empty [allocTraceBatch]byte

func StreamAllocTrace(w io.Writer) {
	done := make(chan struct{})
	go func() {
		for {
			bytes := runtime.ReadAllocTrace()
			if bytes == nil {
				break
			}
			w.Write(bytes)
			w.Write(empty[:allocTraceBatch-len(bytes)])
		}
		close(done)
	}()
	atState.done = done
}

func StopAllocTrace() {
	runtime.StopAllocTrace()
	<-atState.done
}

var atState struct {
	done <-chan struct{}
}

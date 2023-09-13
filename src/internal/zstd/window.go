// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package zstd

type window struct {
	size int    // maximum required window size
	data []byte // window data
	off  int    // read offset
}

func (w *window) reset() {
	w.data = w.data[:0]
	w.off = 0
}

func (w *window) len() uint32 {
	return uint32(len(w.data))
}

func (w *window) save(buf []byte) {
	if w.size == 0 || len(buf) == 0 {
		return
	}

	if len(buf) >= w.size {
		from := len(buf) - w.size
		w.data = append(w.data[:0], buf[from:]...)
		w.off = 0
		return
	}

	tofill := w.size - len(w.data)
	if tofill == 0 {
		n := copy(w.data[w.off:], buf)
		if n < len(buf) {
			w.off = copy(w.data, buf[n:])
		} else {
			w.off += n
		}
	} else if tofill >= len(buf) {
		w.data = append(w.data, buf...)
	} else {
		w.data = append(w.data, buf[:tofill]...)
		w.off = copy(w.data, buf[tofill:])
	}
}

func (w *window) appendTo(buf []byte, from, to uint32) []byte {
	if from == 0 && to == uint32(len(w.data)) {
		buf = append(buf, w.data[w.off:]...)
		return append(buf, w.data[:w.off]...)
	}

	from += uint32(w.off)
	to += uint32(w.off)
	from %= uint32(len(w.data))
	to %= uint32(len(w.data))

	if from <= to {
		return append(buf, w.data[from:to]...)
	} else {
		buf = append(buf, w.data[from:]...)
		return append(buf, w.data[:to]...)
	}
}

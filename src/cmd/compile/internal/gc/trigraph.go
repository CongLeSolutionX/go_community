// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Trigraph support to support Go on mainframes with limited character sets.

package gc

import (
	"bufio"
	"io"
)

type trigraphReader struct {
	r  io.Reader
	br *bufio.Reader
}

func (t *trigraphReader) Read(p []byte) (n int, err error) {
	if t.br == nil {
		t.br = bufio.NewReader(t.r)
	}
	for len(p) > 0 {
		peek, err := t.br.Peek(1)
		if err != nil {
			return n, err
		}
		c := peek[0]
		if c != '?' {
			p[0], _ = t.br.ReadByte()
			p = p[1:]
			n++
			continue
		}
		peek, err = t.br.Peek(3)
		if err == nil && peek[1] == '?' {
			var equiv byte
			switch peek[2] {
			case '=':
				equiv = '#'
			case '/':
				equiv = '\\'
			case '\'':
				equiv = '^'
			case '(':
				equiv = '['
			case ')':
				equiv = ']'
			case '!':
				equiv = '|'
			case '<':
				equiv = '{'
			case '>':
				equiv = '}'
			case '-':
				equiv = '~'
			}
			if equiv != 0 {
				p[0] = equiv
				p = p[1:]
				n++
				for _ = range peek {
					t.br.ReadByte()
				}
				continue
			}
		}
		for i := range peek {
			p[i], _ = t.br.ReadByte()
		}
		p = p[len(peek):]
		n += len(peek)
		if err != nil {
			return n, err
		}
	}
	return n, nil
}

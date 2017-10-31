package strings

import (
	"errors"
	"io"
	"unicode/utf8"
	"unsafe"
)

// A Builder is a buffer of bytes with Write methods that can be turned into
// a string. The zero value for Builder is an empty buffer ready to use.
//
// Unlike a bytes.Buffer, a Builder does not make a final copy of the buffer
// when calling String.
type Builder struct {
	buf []byte
}

// NewBuilderSize creates and initializes a new Builder with a pre-sized buffer.
//
// In most cases, new(Builder) (or just declaring a Builder variable)
// is sufficient to initialize a Builder.
func NewBuilderSize(n int) *Builder {
	return &Builder{buf: make([]byte, 0, n)}
}

// ErrTooLarge is passed to panic if memory cannot be allocated to store data in a Builder.
var ErrTooLarge = errors.New("strings.Builder: too large")

// String returns the contents of b as a string.
func (b *Builder) String() string {
	return *(*string)(unsafe.Pointer(&b.buf))
}

// Len returns the number of buffered bytes; b.Len() == len(b.String()).
func (b *Builder) Len() int { return len(b.buf) }

const maxInt = int(^uint(0) >> 1)

// tryGrowByReslice is an inlineable version of grow for the fast case where the
// internal buffer only needs to be resliced.
// It returns the index where bytes should be written and whether it succeeded.
func (b *Builder) tryGrowByReslice(n int) (int, bool) {
	if l := len(b.buf); n <= cap(b.buf)-l {
		b.buf = b.buf[:l+n]
		return l, true
	}
	return 0, false
}

// grow grows the buffer to guarantee space for n more bytes.
// It returns the index where bytes should be written.
// If the buffer can't grow it will panic with ErrTooLarge.
func (b *Builder) grow(n int) int {
	if i, ok := b.tryGrowByReslice(n); ok {
		return i
	}
	l := len(b.buf)
	c := cap(b.buf)
	if c > maxInt-c-n {
		panic(ErrTooLarge)
	}
	buf := make([]byte, l+n, 2*c+n)
	copy(buf, b.buf)
	b.buf = buf
	return l
}

// Grow grows b's capacity, if necessary, to guarantee space for
// another n bytes. After Grow(n), at least n bytes can be written to b
// without another allocation. If n is negative, Grow panics.
// If b can't grow it will panic with ErrTooLarge.
func (b *Builder) Grow(n int) {
	if n < 0 {
		panic("strings.Builder.Grow: negative count")
	}
	m := b.grow(n)
	b.buf = b.buf[:m]
}

// Write appends the contents of p to b's buffer, growing the buffer as needed.
// Write always returns len(p), nil. If the buffer becomes too large, Write will
// panic with ErrTooLarge.
func (b *Builder) Write(p []byte) (int, error) {
	m, ok := b.tryGrowByReslice(len(p))
	if !ok {
		m = b.grow(len(p))
	}
	return copy(b.buf[m:], p), nil
}

// WriteByte appends the byte c to b's buffer, growing the buffer as needed.
// The returned error is always nil. If the buffer becomes too large, WriteByte
// will panic with ErrTooLarge.
func (b *Builder) WriteByte(c byte) error {
	m, ok := b.tryGrowByReslice(1)
	if !ok {
		m = b.grow(1)
	}
	b.buf[m] = c
	return nil
}

// WriteRune appends the UTF-8 encoding of Unicode code point r to b's buffer,
// growing the buffer as needed. It returns the length of r and a nil error.
// If the buffer becomes too large, WriteRune will panic with ErrTooLarge.
func (b *Builder) WriteRune(r rune) (int, error) {
	if r < utf8.RuneSelf {
		b.WriteByte(byte(r))
		return 1, nil
	}
	m, ok := b.tryGrowByReslice(utf8.UTFMax)
	if !ok {
		m = b.grow(utf8.UTFMax)
	}
	n := utf8.EncodeRune(b.buf[m:m+utf8.UTFMax], r)
	b.buf = b.buf[:m+n]
	return n, nil
}

// WriteString appends the contents of s to b's buffer, growing the buffer
// as needed. It returns the length of s and a nil error.
func (b *Builder) WriteString(s string) (int, error) {
	m, ok := b.tryGrowByReslice(len(s))
	if !ok {
		m = b.grow(len(s))
	}
	return copy(b.buf[m:], s), nil
}

// minRead is the minimum slice passed to a Read call by Builder.ReadFrom.
// It is the same as bytes.MinRead.
const minRead = 512

// errNegativeRead is the panic value if the reader passed to Builder.ReadFrom
// returns a negative count.
var errNegativeRead = errors.New("strings.Builder: reader returned negative count from Read")

// ReadFrom reads data from r until EOF and appends it to b's buffer,
// growing the buffer as needed. The return value n is the number of bytes read.
// Any error except io.EOF encountered during the read is also returned.
// If the buffer becomes too large, ReadFrom will panic with ErrTooLarge.
func (b *Builder) ReadFrom(r io.Reader) (n int64, err error) {
	for {
		i := b.grow(minRead)
		m, e := r.Read(b.buf[i:cap(b.buf)])
		if m < 0 {
			panic(errNegativeRead)
		}
		b.buf = b.buf[:i+m]
		n += int64(m)
		if e == io.EOF {
			return n, nil
		}
		if e != nil {
			return n, e
		}
	}
}

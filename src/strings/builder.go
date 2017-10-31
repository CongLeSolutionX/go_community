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

// String returns the contents of b as a string.
func (b *Builder) String() string {
	return *(*string)(unsafe.Pointer(&b.buf))
}

// Len returns the number of buffered bytes; b.Len() == len(b.String()).
func (b *Builder) Len() int { return len(b.buf) }

// Write appends the contents of p to the buffer, growing the buffer as needed.
// Write always returns len(p), nil.
func (b *Builder) Write(p []byte) (int, error) {
	b.buf = append(b.buf, p...)
	return len(p), nil
}

// WriteByte appends the byte c to the buffer, growing the buffer as needed.
// The returned error is always nil.
func (b *Builder) WriteByte(c byte) error {
	b.buf = append(b.buf, c)
	return nil
}

// WriteRune appends the UTF-8 encoding of Unicode code point r to the buffer,
// growing the buffer as needed. It returns the length of r and a nil error.
func (b *Builder) WriteRune(r rune) (int, error) {
	if r < utf8.RuneSelf {
		b.buf = append(b.buf, byte(r))
		return 1, nil
	}
	m := len(b.buf)
	b.buf = append(b.buf, make([]byte, utf8.UTFMax)...)
	n := utf8.EncodeRune(b.buf[m:m+utf8.UTFMax], r)
	b.buf = b.buf[:m+n]
	return n, nil
}

// WriteString appends the contents of s to the buffer, growing the buffer
// as needed. It returns the length of s and a nil error.
func (b *Builder) WriteString(s string) (int, error) {
	b.buf = append(b.buf, s...)
	return len(s), nil
}

// minRead is the minimum slice passed to a Read call by Builder.ReadFrom.
// It is the same as bytes.MinRead.
const minRead = 512

// errNegativeRead is the panic value if the reader passed to Builder.ReadFrom
// returns a negative count.
var errNegativeRead = errors.New("strings.Builder: reader returned negative count from Read")

// ReadFrom reads data from r until EOF and appends it to the buffer,
// growing the buffer as needed. The return value n is the number of bytes read.
// Any error except io.EOF encountered during the read is also returned.
func (b *Builder) ReadFrom(r io.Reader) (n int64, err error) {
	for {
		if free := cap(b.buf) - len(b.buf); free < minRead {
			m := len(b.buf)
			b.buf = append(b.buf, make([]byte, minRead-free)...)
			b.buf = b.buf[:m]
		}
		m, e := r.Read(b.buf[len(b.buf):cap(b.buf)])
		if m < 0 {
			panic(errNegativeRead)
		}
		b.buf = b.buf[:len(b.buf)+m]
		n += int64(m)
		if e == io.EOF {
			return n, nil
		}
		if e != nil {
			return n, e
		}
	}
}

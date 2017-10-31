package strings_test

import (
	"errors"
	"io"
	. "strings"
	"testing"
	"testing/iotest"
)

func check(t *testing.T, b *Builder, want string) {
	t.Helper()
	got := b.String()
	if got != want {
		t.Errorf("String: got %#q; want %#q", got, want)
		return
	}
	if n := b.Len(); n != len(got) {
		t.Errorf("Len: got %d; but len(String()) is %d", n, len(got))
	}
}

func TestBuilder(t *testing.T) {
	b := NewBuilderSize(8)
	check(t, b, "")
	n, err := b.WriteString("hello")
	if err != nil || n != 5 {
		t.Errorf("WriteString: got %d,%s; want 5,nil", n, err)
	}
	check(t, b, "hello")
	if err = b.WriteByte(' '); err != nil {
		t.Errorf("WriteByte: %s", err)
	}
	check(t, b, "hello ")
	n, err = b.WriteString("world")
	if err != nil || n != 5 {
		t.Errorf("WriteString: got %d,%s; want 5,nil", n, err)
	}
	check(t, b, "hello world")
}

func TestBuilderString(t *testing.T) {
	b := NewBuilderSize(5)
	b.WriteString("alpha")
	check(t, b, "alpha")
	s1 := b.String()
	b.WriteString("beta")
	check(t, b, "alphabeta")
	s2 := b.String()
	b.WriteString("gamma")
	check(t, b, "alphabetagamma")
	s3 := b.String()

	// Check that subsequent operations didn't change the returned strings.
	if want := "alpha"; s1 != want {
		t.Errorf("first String result is now %q; want %q", s1, want)
	}
	if want := "alphabeta"; s2 != want {
		t.Errorf("second String result is now %q; want %q", s2, want)
	}
	if want := "alphabetagamma"; s3 != want {
		t.Errorf("third String result is now %q; want %q", s3, want)
	}
}

func TestBuilderEmpty(t *testing.T) {
	var b Builder
	check(t, &b, "")
}

func TestBuilderWrite2(t *testing.T) {
	const s0 = "hello 世界"
	for _, tt := range []struct {
		name string
		fn   func(b *Builder) (int, error)
		n    int
		want string
	}{
		{
			"Write",
			func(b *Builder) (int, error) { return b.Write([]byte(s0)) },
			len(s0),
			s0,
		},
		{
			"WriteRune",
			func(b *Builder) (int, error) { return b.WriteRune('a') },
			1,
			"a",
		},
		{
			"WriteRuneWide",
			func(b *Builder) (int, error) { return b.WriteRune('世') },
			3,
			"世",
		},
		{
			"WriteString",
			func(b *Builder) (int, error) { return b.WriteString(s0) },
			len(s0),
			s0,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var b Builder
			n, err := tt.fn(&b)
			if err != nil {
				t.Fatalf("first call: got %s", err)
			}
			if n != tt.n {
				t.Errorf("first call: got n=%d; want %d", n, tt.n)
			}
			check(t, &b, tt.want)

			n, err = tt.fn(&b)
			if err != nil {
				t.Fatalf("second call: got %s", err)
			}
			if n != tt.n {
				t.Errorf("second call: got n=%d; want %d", n, tt.n)
			}
			check(t, &b, tt.want+tt.want)
		})
	}
}

func TestBuilderWriteByte(t *testing.T) {
	var b Builder
	if err := b.WriteByte('a'); err != nil {
		t.Error(err)
	}
	if err := b.WriteByte(0); err != nil {
		t.Error(err)
	}
	check(t, &b, "a\x00")
}

func TestBuilderReadFrom(t *testing.T) {
	for _, tt := range []struct {
		name string
		fn   func(io.Reader) io.Reader
	}{
		{"Reader", func(r io.Reader) io.Reader { return r }},
		{"DataErrReader", iotest.DataErrReader},
		{"OneByteReader", iotest.OneByteReader},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var b Builder

			r := tt.fn(NewReader("hello"))
			n, err := b.ReadFrom(r)
			if err != nil {
				t.Fatalf("first call: got %s", err)
			}
			if n != 5 {
				t.Errorf("first call: got n=%d; want 5", n)
			}
			check(t, &b, "hello")

			r = tt.fn(NewReader(" world"))
			n, err = b.ReadFrom(r)
			if err != nil {
				t.Fatalf("first call: got %s", err)
			}
			if n != 6 {
				t.Errorf("first call: got n=%d; want 6", n)
			}
			check(t, &b, "hello world")
		})
	}
}

var errRead = errors.New("boom")

// errorReader sends reads to the underlying reader but returns errRead instead
// of io.EOF.
type errorReader struct {
	r io.Reader
}

func (r errorReader) Read(b []byte) (int, error) {
	n, err := r.r.Read(b)
	if err == io.EOF {
		err = errRead
	}
	return n, err
}

func TestBuilderReadFromError(t *testing.T) {
	var b Builder
	r := errorReader{NewReader("hello")}
	n, err := b.ReadFrom(r)
	if n != 5 {
		t.Errorf("got n=%d; want 5", n)
	}
	if err != errRead {
		t.Errorf("got err=%q; want %q", err, errRead)
	}
	check(t, &b, "hello")
}

type negativeReader struct{}

func (r negativeReader) Read([]byte) (int, error) { return -1, nil }

func TestBuilderReadFromNegativeReader(t *testing.T) {
	var b Builder
	defer func() {
		switch err := recover().(type) {
		case nil:
			t.Fatal("ReadFrom didn't panic")
		case error:
			wantErr := "strings.Builder: reader returned negative count from Read"
			if err.Error() != wantErr {
				t.Fatalf("recovered panic: got %v; want %v", err.Error(), wantErr)
			}
		default:
			t.Fatalf("unexpected panic value: %#v", err)
		}
	}()

	b.ReadFrom(negativeReader{})
}

func TestBuilderAllocs(t *testing.T) {
	var b *Builder
	var s string
	allocs := testing.AllocsPerRun(1, func() {
		if b == nil {
			// First (warm-up) run.
			b = NewBuilderSize(5)
		} else {
			// Second (alloc-measuring) run.
			b.WriteString("hello")
			s = b.String()
		}
	})
	check(t, b, "hello")
	if allocs > 0 {
		t.Fatalf("got %d alloc(s); want 0", allocs)
	}
}

package wrapper_test

import (
	"testing"
	"wrapper"
)

type wrapperT struct{ s string }

func (w wrapperT) Unwrap() wrapper.Wrapper { return nil }

type wrappedT struct {
	s string
	w wrapper.Wrapper
}

func (w wrappedT) Unwrap() wrapper.Wrapper { return w.w }

func TestUnwrap(t *testing.T) {
	w1 := wrapperT{"wrapper"}
	w2 := wrappedT{"wrapper wrap", w1}

	testCases := []struct {
		w    wrapper.Wrapper
		want wrapper.Wrapper
	}{
		{nil, nil},
		{wrappedT{"wrapped", nil}, nil},
		{w1, nil},
		{w2, w1},
		{wrappedT{"wrap 3", w2}, w2},
	}
	for _, tc := range testCases {
		if got := wrapper.Unwrap(tc.w); got != tc.want {
			t.Errorf("Unwrap(%v) = %v, want %v", tc.w, got, tc.want)
		}
	}
}

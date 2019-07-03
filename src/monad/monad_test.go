package monad_test

import (
	"monad"
	"testing"
)

type monadT struct{ s string }

func (m monadT) Unwrap() monad.Monad { return nil }

type wrappedT struct {
	s string
	m monad.Monad
}

func (w wrappedT) Unwrap() monad.Monad { return w.m }

func TestUnwrap(t *testing.T) {
	m1 := monadT{"monad"}
	m2 := wrappedT{"monad wrap", m1}

	testCases := []struct {
		m    monad.Monad
		want monad.Monad
	}{
		{nil, nil},
		{wrappedT{"wrapped", nil}, nil},
		{m1, nil},
		{m2, m1},
		{wrappedT{"wrap 3", m2}, m2},
	}
	for _, tc := range testCases {
		if got := monad.Unwrap(tc.m); got != tc.want {
			t.Errorf("Unwrap(%v) = %v, want %v", tc.m, got, tc.want)
		}
	}
}

package benchmark

import "testing"

func TestMakeBenchString(t *testing.T) {
	tests := []struct {
		have, want string
	}{
		{"foo", "BenchmarkFoo"},
		{"  foo  ", "BenchmarkFoo"},
		{"foo bar", "BenchmarkFooBar"},
	}
	for i, test := range tests {
		if v := makeBenchString(test.have); test.want != v {
			t.Errorf("test[%d] teleakeBenchString(%q) == %q, want %q", i, test.have, v, test.want)
		}
	}
}

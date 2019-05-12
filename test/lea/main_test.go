package main

import "testing"

var global int

func BenchmarkLEA22_1_noinline(b *testing.B) {
	var sink int
	for i := 0; i < b.N; i++ {
		sink = lea22_1_noinline(sink, sink)
	}
	global = sink
}

func BenchmarkLEA22_4_noinline(b *testing.B) {
	var sink int
	for i := 0; i < b.N; i++ {
		sink = lea22_4_noinline(sink, sink)
	}
	global = sink
}

func BenchmarkLEA22_1_inline(b *testing.B) {
	var sink int
	for i := 0; i < b.N; i++ {
		sink = lea22_1_inline(sink, sink)
	}
	global = sink
}

func BenchmarkLEA22_4_inline(b *testing.B) {
	var sink int
	for i := 0; i < b.N; i++ {
		sink = lea22_4_inline(sink, sink)
	}
	global = sink
}

package runtime_test

import (
	"fmt"
	"reflect"
	"runtime"
	"testing"
)

var callerImpls = []struct {
	name string
	fn   func([]uintptr) int
}{
	{"FPCallers", runtime.FPCallers},
	{"ShadowFPCallers", runtime.ShadowFPCallers},
}

func TestShadowCallers(t *testing.T) {
	t.Run("panic", func(t *testing.T) {
		defer func() {
			if err := recover(); err == nil {
				t.Fatal("expected panic")
			}
		}()
		beforeCallN(5, func(_ int) {
			buf := make([]uintptr, 32)
			runtime.ShadowFPCallers(buf)
			panic("panic")
		})
	})

	t.Run("misc", func(t *testing.T) {
		const probeDepth = 1
		const maxDepth = 32
		var golden [probeDepth + 1][]uintptr
		for _, ci := range callerImpls {
			t.Run(ci.name, func(t *testing.T) {
				for i := 0; i < 3; i++ {
					beforeCallN(probeDepth, func(n int) {
						pcs := make([]uintptr, maxDepth)
						pcs = pcs[0:ci.fn(pcs)]
						if golden[n] == nil {
							golden[n] = pcs
						} else if !reflect.DeepEqual(golden[n], pcs) {
							t.Fatalf("got %v, want %v i=%d", pcs, golden[n], i)
						}
					})
				}
			})
		}
	})
}

func BenchmarkShadowCallers(b *testing.B) {
	const maxDepth = 256
	const checkResults = false

	b.Run("best case", func(b *testing.B) {
		type expectKey struct {
			depth int
		}
		expect := map[expectKey][]uintptr{}
		for _, ci := range callerImpls {
			b.Run(ci.name, func(b *testing.B) {
				for depth := 1; depth <= maxDepth; depth = depth * 2 {
					b.Run(fmt.Sprintf("depth=%d", depth), func(b *testing.B) {
						errCh := make(chan error, 1)
						go func() {
							defer func() {
								select {
								case errCh <- nil:
								default:
								}
							}()
							pcs := make([]uintptr, depth*2+8)
							callAtN(depth, func() {
								for i := 0; i < b.N; i++ {
									n := ci.fn(pcs)
									if checkResults {
										got := pcs[0:n]
										key := expectKey{depth}
										if expect[key] == nil {
											expect[key] = append(expect[key], got...)
										} else if !reflect.DeepEqual(expect[key], got) {
											select {
											case errCh <- fmt.Errorf("got=%v want=%v i=%d", got, expect[key], i):
											default:
											}
										}
									}
								}
							})
						}()
						if err := <-errCh; err != nil {
							b.Fatal(err)
						}
					})
				}
			})
		}
	})

	b.Run("worst case", func(b *testing.B) {
		type expectKey struct {
			depth int
			key   int
		}
		expect := map[expectKey][]uintptr{}
		for _, ci := range callerImpls {
			b.Run(ci.name, func(b *testing.B) {
				for depth := 1; depth <= maxDepth; depth = depth * 2 {
					b.Run(fmt.Sprintf("depth=%d", depth), func(b *testing.B) {
						errCh := make(chan error, 1)
						go func() {
							defer func() {
								select {
								case errCh <- nil:
								default:
								}
							}()
							pcs := make([]uintptr, depth*2+8)
							for i := 0; i < max(b.N/depth/2, 1); i++ {
								beforeAfterCallN(depth, func(key int) {
									n := ci.fn(pcs)
									if checkResults {
										got := pcs[0:n]
										key := expectKey{depth, key}
										if expect[key] == nil {
											expect[key] = append(expect[key], got...)
										} else if !reflect.DeepEqual(expect[key], got) {
											select {
											case errCh <- fmt.Errorf("got=%v want=%v i=%d", got, expect[key], i):
											default:
											}
										}
									}
								})
							}
						}()
						if err := <-errCh; err != nil {
							b.Fatal(err)
						}
					})
				}
			})
		}

	})
}

//go:noinline
func callAtN(depth int, fn func()) {
	if depth > 1 {
		callAtN(depth-1, fn)
	} else {
		fn()
	}
}

//go:noinline
func beforeCallN(depth int, fn func(int)) {
	fn(depth)
	if depth > 0 {
		beforeCallN(depth-1, fn)
	}
}

//go:noinline
func beforeAfterCallN(depth int, fn func(int)) {
	if depth > 0 {
		fn(depth*2 - 1)
		beforeAfterCallN(depth-1, fn)
		fn(depth*2 - 2)
	}
}

//go:noinline
func call5(fn func()) { call4(fn) }

//go:noinline
func call4(fn func()) { call3(fn) }

//go:noinline
func call3(fn func()) { call2(fn) }

//go:noinline
func call2(fn func()) { call1(fn) }

//go:noinline
func call1(fn func()) { call0(fn) }

//go:noinline
func call0(fn func()) { fn() }

package main

import (
	"runtime"
)

func main() {
	runtime.PrintStack("main")
	defer func() {
		runtime.PrintStack("defer1")
	}()
	defer func() {
		runtime.PrintStack("defer2")
		_ = recover()
	}()
	//println(g(500))
	f(8)
}

//go:noinline
func f(i int) {
	runtime.PrintStack("f")
	var a [8]int
	a[i] = 0
}

//go:noinline
func g(a int) int {
	if a > 0 {
		return g(a - 1)
	}
	return 42
}

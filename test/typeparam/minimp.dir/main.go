package main

import (
	"a"
	"fmt"
)

func main() {
	const want = 2
	if got := a.Min[int](2, 3); got != want {
		panic(fmt.Sprintf("got %d, want %d", got, want))
	}

	if got := a.Min(2, 3); got != want {
		panic(fmt.Sprintf("want %d, got %d", want, got))
	}

	if got := a.Min[float64](3.5, 2.0); got != want {
		panic(fmt.Sprintf("got %d, want %d", got, want))
	}

	if got := a.Min(3.5, 2.0); got != want {
		panic(fmt.Sprintf("got %d, want %d", got, want))
	}
}

package const_pure_func_test

import "fmt"

//go:noinline
func AddConst(x, y int) int {
	fmt.Printf("AddConst %d, %d called\n", x, y) // THIS IS NOT REALLY CONST
	return x + y
}

//go:noinline
func AddPure(x, y *int) int {
	fmt.Printf("AddPure %d, %d called\n", *x, *y) // THIS IS NOT REALLY PURE
	return *x + *y
}

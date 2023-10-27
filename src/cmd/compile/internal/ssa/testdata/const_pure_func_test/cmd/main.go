package main

import (
	x "const_pure_func_test"
	"fmt"
)

var mem int

func main() {
	a := x.AddConst(3, 4) + x.AddConst(3, 4) // should call once
	b := 6
	c := x.AddPure(&a, &b) + x.AddPure(&a, &b) + x.AddPure(&a, &b) // should call once
	mem = c
	d := x.AddPure(&a, &b) + x.AddPure(&a, &b) + x.AddPure(&a, &b) // should call once
	_ = x.AddConst(6, 7) + x.AddPure(&c, &a)                       // should not call at all
	fmt.Printf("a=%d, c=%d, d=%d\n", a, c, d)
}

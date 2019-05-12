package main

//go:noinline
func lea22_1_noinline(a, b int) int {
	return 1 + a + b
}

func lea22_1_inline(a, b int) int {
	return 1 + a + b
}

//go:noinline
func lea22_4_noinline(a, b int) int {
	return 1 + (a + 4*b)
}

func lea22_4_inline(a, b int) int {
	return 1 + (a + 4*b)
}

func main() {
}

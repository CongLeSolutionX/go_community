// +build amd64
// errorcheck -0 -d=ssaprove

package main

func f0(a []int) int {
	a[0] = 1
	a[0] = 1 // ERROR "Proved IsInBounds$"
	a[6] = 1
	a[6] = 1 // ERROR "Proved IsInBounds$"
	a[5] = 1
	a[5] = 1 // ERROR "Proved IsInBounds$"
	return 13
}

func f1(a []int) int {
	if len(a) <= 5 {
		return 18
	}
	a[0] = 1
	a[0] = 1 // ERROR "Proved IsInBounds$"
	a[6] = 1
	a[6] = 1 // ERROR "Proved IsInBounds$"
	a[5] = 1 // ERROR "Proved constant IsInBounds$"
	a[5] = 1 // ERROR "Proved IsInBounds$"
	return 25
}

func f2(a []int) int {
	for i := range a {
		a[i] = i
		a[i] = i // ERROR "Proved IsInBounds$"
	}
	return 32
}

func f3(a []uint) int {
	for i := uint(0); i < uint(len(a)); i++ {
		a[i] = i // ERROR "Proved IsInBounds$"
	}
	return 38
}

func f4(a, b, c int) int {
	if a < b {
		if a < b { // ERROR "Proved Less64$"
			return 43
		}
		if a > b { // ERROR "Proved Greater64$"
			return 45
		}
		if a == b { // ERROR "Proved Eq64$"
			return 47
		}
		return 48
	}
	if a <= b {
		if a >= b { // ERROR "Proved Geq64$"
			if a == b { // ERROR "Proved Eq64$"
				return 52
			}
			if a != b { // ERROR "Proved Neq64$"
				return 54
			}
			return 55
		}
		return 56
	}
	if a < b { // ERROR "Disproved Less64$"
		if a < c {
			if a < b { // ERROR "Proved Less64$"
				if a < c { // ERROR "Proved Less64$"
					return 61
				}
				return 62
			}
			return 63
		}
		return 64
	}
	if a < b { // ERROR "Disproved Less64$"
		if b > a { // ERROR "Proved Greater64$"
			return 67
		}
		return 68
	}
	if a <= b { // ERROR "Disproved Leq64$"
		if b > a { // ERROR "Proved Greater64$"
			if b == a { // ERROR "Proved Eq64$"
				return 72
			}
			return 73
		}
		if b >= a { // ERROR "Proved Geq64$"
			if b == a { // ERROR "Proved Eq64$"
				return 76
			}
			return 77
		}
		return 78
	}
	return 79
}

func f5(a, b uint) int {
	if a == b {
		if a <= b { // ERROR "Proved Leq64U$"
			return 84
		}
		return 85
	}
	return 86
}

// These comparisons are compile time constants.
func f6(a uint8) int {
	if a < a { // ERROR "Disproved Less8U$"
		return 94
	}
	if a > a { // ERROR "Disproved Greater8U$"
		return 96
	}
	if a <= a { // ERROR "Proved Leq8U$"
		return 98
	}
	if a >= a { // ERROR "Proved Geq8U$"
		return 100
	}
	return 101
}

func f7(a []int, b int) int {
	if b < len(a) {
		a[b] = 3
		if b < len(a) { // ERROR "Proved Less64$"
			a[b] = 5 // ERROR "Proved IsInBounds$"
		}
	}
	return 0
}

func f8(a, b uint) int {
	if a == b {
		return 1
	}
	if a > b {
		return 2
	}
	if a < b { // ERROR "Proved Less64U$"
		return 3
	}
	return 3
}

func main() {
}

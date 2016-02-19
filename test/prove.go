// +build amd64
// errorcheck -0 -d=ssa/prove/debug=3

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
	return 26
}

func f2(a []int) int {
	for i := range a {
		a[i] = i
		a[i] = i // ERROR "Proved IsInBounds$"
	}
	return 34
}

func f3(a []uint) int {
	for i := uint(0); i < uint(len(a)); i++ {
		a[i] = i // ERROR "Proved IsInBounds$"
	}
	return 41
}

func f4(a, b, c int) int {
	if a < b {
		if a < b { // ERROR "Proved Less64$"
			return 47
		}
		if a > b { // ERROR "Proved Greater64$"
			return 50
		}
		if a == b { // ERROR "Proved Eq64$"
			return 53
		}
		return 55
	}
	if a <= b {
		if a >= b { // ERROR "Proved Geq64$"
			if a == b { // ERROR "Proved Eq64$"
				return 60
			}
			if a != b { // ERROR "Proved Neq64$"
				return 63
			}
			return 65
		}
		return 67
	}
	if a < b { // ERROR "Disproved Less64$"
		if a < c {
			if a < b { // ERROR "Proved Less64$"
				if a < c { // ERROR "Proved Less64$"
					return 73
				}
				return 75
			}
			return 77
		}
		return 79
	}
	if a < b { // ERROR "Disproved Less64$"
		if b > a { // ERROR "Proved Greater64$"
			return 83
		}
		return 85
	}
	if a <= b { // ERROR "Disproved Leq64$"
		if b > a { // ERROR "Proved Greater64$"
			if b == a { // ERROR "Proved Eq64$"
				return 90
			}
			return 92
		}
		if b >= a { // ERROR "Proved Geq64$"
			if b == a { // ERROR "Proved Eq64$"
				return 96
			}
			return 98
		}
		return 100
	}
	return 102
}

func f5(a, b uint) int {
	if a == b {
		if a <= b { // ERROR "Proved Leq64U$"
			return 108
		}
		return 110
	}
	return 112
}

// These comparisons are compile time constants.
func f6(a uint8) int {
	if a < a { // ERROR "Disproved Less8U$"
		return 118
	}
	if a > a { // ERROR "Disproved Greater8U$"
		return 121
	}
	if a <= a { // ERROR "Proved Leq8U$"
		return 124
	}
	if a >= a { // ERROR "Proved Geq8U$"
		return 127
	}
	return 129
}

func f7(a []int, b int) int {
	if b < len(a) {
		a[b] = 3
		if b < len(a) { // ERROR "Proved Less64$"
			a[b] = 5 // ERROR "Proved IsInBounds$"
		}
	}
	return 139
}

func f8(a, b uint) int {
	if a == b {
		return 144
	}
	if a > b {
		return 147
	}
	if a < b { // ERROR "Proved Less64U$"
		return 150
	}
	return 152
}

func main() {
}

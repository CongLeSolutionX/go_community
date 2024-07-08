package strings

// Reverse returnes a copy of s but reversed
func Reverse(s string) string {
	// Convert string to a rune slice
	// to reverse strings containing multi-byte characters accurately
	runes := []rune(s)

	// Reverse the rune slice
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	// Return the string representation of the rune slice
	return string(runes)
}

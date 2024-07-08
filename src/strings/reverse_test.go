package strings_test

import (
	. "strings"
	"testing"
)

func TestReverse(t *testing.T) {
	input1 := "abc"
	expected1 := "cba"
	if output := Reverse(input1); output != expected1 {
		t.Errorf("Reverse(%s) = %s; want %s", input1, output, expected1)
	}

	input2 := ""
	expected2 := ""
	if output := Reverse(input2); output != expected2 {
		t.Errorf("Reverse(%s) = %s; want %s", input2, output, expected2)
	}

	input3 := "a"
	expected3 := "a"
	if output := Reverse(input3); output != expected3 {
		t.Errorf("Reverse(%s) = %s; want %s", input3, output, expected3)
	}
}

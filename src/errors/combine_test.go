package errorcollection

import (
	"fmt"
	"testing"
)

// returns an error that combines the specified number of errors.
func testFunction(count int) error {
	var ec error
	i := 0
	for i < count {
		i = i + 1
		ec = Combine(ec, fmt.Errorf("error #%d", i))
	}
	return ec
}

func TestZeroErrorsIsNil(t *testing.T) {
	var err error
	err = testFunction(0)
	if err != nil {
		t.Errorf("testFunction(0) got %v, wanted nil.", err)
	}
}

func TestOneErrorIsNotNil(t *testing.T) {
	var err error
	err = testFunction(1)
	if err == nil {
		t.Errorf("testFunction(1) got nil, wanted not nil")
	}
	if Count(err) != 1 {
		t.Errorf("Count(err) got %v, wanted 1.", Count(err))
	}
	got := err.Error()
	if got != "error #1" {
		t.Errorf("err.Error() got %q, wanted \"error #1\"", got)
	}
}

func TestCombineWithNil(t *testing.T) {
	// a nil errorCollection combined with nil should be nil
	var ec error
	var nilError error
	ec = Combine(ec, nilError)
	if ec != nil {
		t.Errorf("Combine(ec, nil) got %q, wanted nil", ec)
	}
}

func TestNilErrorCollectionIsEmpty(t *testing.T) {
	var ec errorCollection
	if Count(ec) != 0 {
		t.Errorf("Count(ec) got %v, wanted 0.", Count(ec))
	}
}

func TestErrorCollection(t *testing.T) {
	var ec error

	// Add first error:
	ec = Combine(ec, fmt.Errorf("foobar"))
	if ec == nil {
		t.Errorf("ec with one error got nil, wanted not nil")
	}
	if Count(ec) != 1 {
		t.Errorf("Count(ec) got %v, wanted 1", Count(ec))
	}
	got := ec.Error()
	if got != "foobar" {
		t.Errorf("ec.Error() got %q, wanted \"foobar\"", got)
	}

	// Add second error:
	ec = Combine(ec, fmt.Errorf("another error"))
	if Count(ec) != 2 {
		t.Errorf("Count(ec) got %v, wanted 2", Count(ec))
	}
	got = ec.Error()
	if got != "foobar; another error" {
		t.Errorf("ec.Error() got %q, wanted \"foobar; another error\"", got)
	}

	// Combine two collections:
	ec2 := Combine(fmt.Errorf("err1"), fmt.Errorf("err2"))
	ec = Combine(ec, ec2)
	if Count(ec) != 4 {
		t.Errorf("Count(ec) got %v, wanted 4", Count(ec))
	}
	got = ec.Error()
	if got != "foobar; another error; err1; err2" {
		t.Errorf("ec.Error() got %q, wanted \"foobar; another error; err1; err2\"", got)
	}

}

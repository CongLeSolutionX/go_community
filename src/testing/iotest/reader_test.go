package iotest

import (
	"testing"
	"testing/iotest"
)

func TestFailReader(t *testing.T) {
	read, err := iotest.FailReader().Read([]byte{})
	if err != iotest.ErrIO {
		t.Errorf("FailReader.Read(any) should have returned ErrIO, returned %v", err)
	}
	if read != 0 {
		t.Errorf("FailReader.Read(any) should have read 0 bytes, read %v", read)
	}
}

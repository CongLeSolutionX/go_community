package testdata

import (
	"testing"
)

func nonTest() {} // OK because it doesn't start with "Test".

type someType struct{}

func (someType) TesthasReceiver() {} // OK because it has a receiver.

func TestOKSuffix(*testing.T) {} // OK because first char after "Test" is Uppercase.

func TestÜnicodeWorks(*testing.T) {} // OK because the first char after "Test" is Uppercase.

func TestbadSuffix(*testing.T) {} // ERROR "first letter after 'Test' must not be lowercase"

func Test(*testing.T) {} // OK "Test" on its own is considered a test.

func BenchmarkbadSuffix(*testing.B) {} // ERROR "first letter after 'Benchmark' must not be lowercase"

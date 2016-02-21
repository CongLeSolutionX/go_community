package testdata

import (
	"testing"
)

func nonTest() {} // OK because it doesn't start with "Test".

type someType struct{}

func (someType) TestHasReceiver() {} // OK because it has a receiver.

func TestOKSuffix(*testing.T) {} // OK because first char after "Test" is Uppercase.

func TestÜnicodeWorks(*testing.T) {} // OK because the first char after "Test" is Uppercase.

func TestbadSuffix(*testing.T) {} // ERROR "first letter after 'Test' must not be lowercase: badSuffix"

func Test(*testing.T) {} // OK "Test" on its own is considered a test.

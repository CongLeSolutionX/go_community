package runtime_test

import (
	"runtime"
	"testing"
)

// golang.org/issues/54786
func TestPrintLockDeadlock(t *testing.T) {

	for i := 1; i <= runtime.NumCPU(); i++ {
		concurrentPrint(i)
	}
}

func concurrentPrint(i int) {
	runtime.GOMAXPROCS(i)
	for i := 0; i < 10; i++ {
		go func() {
			for i := 0; i < 100000; i++ {
				print("hello world\n")
			}
		}()
	}
}

package poll_test

import (
	"internal/poll"
	"internal/syscall/unix"
	"runtime"
	"syscall"
	"testing"
	"time"
)

func TestSplicePipePool(t *testing.T) {
	p, _, err := poll.GetPipe()
	if err != nil {
		t.Skip("failed to create pipe skip")
	}
	prfd, pwfd := poll.GetPipePair(p)
	poll.PutPipe(p)
	p = nil
	runtime.GC()
	time.Sleep(time.Duration(100+10) * time.Millisecond)
	runtime.GC()
	time.Sleep(time.Duration(2*100+10) * time.Millisecond)
	_, _, errno1 := syscall.Syscall(unix.FcntlSyscall, uintptr(pwfd), syscall.F_GETFD, 0)
	_, _, errno2 := syscall.Syscall(unix.FcntlSyscall, uintptr(prfd), syscall.F_GETFD, 0)
	if errno1 == 0 || errno2 == 0 {
		t.Fatalf("pipe is still open, errno1: %d, errno2: %d\n", errno1, errno2)
	}
}

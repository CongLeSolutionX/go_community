// +build !goexperiment.staticlockranking

package runtime

// Mutual exclusion locks.  In the uncontended case,
// as fast as spin locks (just a few user-level instructions),
// but on the contention path they sleep in the kernel.
// A zeroed Mutex is unlocked (no need to initialize each lock).
// Initialization is helpful for static lock ranking, but not required.
type mutex struct {
	// Futex-based impl treats it as uint32 key,
	// while sema-based impl as M* waitm.
	// Used to be a union, but unions break precise GC.
	key uintptr
}

func lockInit(l *mutex, rank int) {
}

func lock(l *mutex) {
	lock2(l)
}

func unlock(l *mutex) {
	unlock2(l)
}

func lockRankAcquire(l *mutex, rank int) {
	lock2(l)
}

func lockLogMayAcquire(l *mutex, rank int) {
}

// +build goexperiment.staticlockranking

package sync

import "unsafe"

// Approximation of notifyList in runtime/sema.go. Size and alignment must
// agree.
type notifyList struct {
	wait   uint32
	notify uint32
	lock   uintptr // key field of the mutex
	rank   int     // rank field of the mutex
	pad    int
	head   unsafe.Pointer
	tail   unsafe.Pointer
}

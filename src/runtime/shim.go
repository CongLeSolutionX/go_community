package runtime

import _ "unsafe"

import "runtime/internal/lock"

const GOOS = lock.GOOS

const GOARCH = lock.GOARCH

//go:linkname MemProfileRate runtime/internal/lock.MemProfileRate
var MemProfileRate int

//go:linkname runtime/internal/gc.GCprinttimes GCprinttimes
func GCprinttimes()

//go:linkname runtime/internal/gc.Gosched Gosched
func Gosched()

//go:linkname runtime/internal/gc.GC GC
func GC()

// Add a dummy struct decl
type StackRecord struct {
	Stack0 [32]uintptr
}

// Add a dummy struct decl
type MemStats struct {
	Alloc        uint64
	TotalAlloc   uint64
	Sys          uint64
	Lookups      uint64
	Mallocs      uint64
	Frees        uint64
	HeapAlloc    uint64
	HeapSys      uint64
	HeapIdle     uint64
	HeapInuse    uint64
	HeapReleased uint64
	HeapObjects  uint64
	StackInuse   uint64
	StackSys     uint64
	MSpanInuse   uint64
	MSpanSys     uint64
	MCacheInuse  uint64
	MCacheSys    uint64
	BuckHashSys  uint64
	GCSys        uint64
	OtherSys     uint64
	NextGC       uint64
	LastGC       uint64
	PauseTotalNs uint64
	PauseNs      [256]uint64
	PauseEnd     [256]uint64
	NumGC        uint32
	EnableGC     bool
	DebugGC      bool
	BySize       [61]struct {
		Size    uint32
		Mallocs uint64
		Frees   uint64
	}
}

//go:linkname runtime/internal/prof.ReadMemStats ReadMemStats
func ReadMemStats(m *MemStats)

//go:linkname runtime/internal/prof.SetBlockProfileRate SetBlockProfileRate
func SetBlockProfileRate(rate int)

//go:linkname runtime/internal/prof.Stack Stack
func (x *StackRecord) Stack() []uintptr

//go:linkname runtime/internal/prof.MemProfile MemProfile
func MemProfile(p []MemProfileRecord, inuseZero bool) (n int, ok bool)

// Add a dummy struct decl
type MemProfileRecord struct {
	AllocBytes   int64
	FreeBytes    int64
	AllocObjects int64
	FreeObjects  int64
	Stack0       [32]uintptr
}

//go:linkname runtime/internal/prof.SetCPUProfileRate SetCPUProfileRate
func SetCPUProfileRate(hz int)

//go:linkname runtime/internal/prof.Stack Stack
func (x *MemProfileRecord) Stack() []uintptr

//go:linkname runtime/internal/prof.InUseBytes InUseBytes
func (x *MemProfileRecord) InUseBytes() int64

// Add a dummy struct decl
type BlockProfileRecord struct {
	Count       int64
	Cycles      int64
	StackRecord StackRecord
}

//go:linkname runtime/internal/prof.ThreadCreateProfile ThreadCreateProfile
func ThreadCreateProfile(p []StackRecord) (n int, ok bool)

//go:linkname runtime/internal/prof.InUseObjects InUseObjects
func (x *MemProfileRecord) InUseObjects() int64

//go:linkname runtime/internal/prof.BlockProfile BlockProfile
func BlockProfile(p []BlockProfileRecord) (n int, ok bool)

//go:linkname runtime/internal/ifacestuff.Error Error
func (x *TypeAssertionError) Error() string

//go:linkname runtime/internal/ifacestuff.RuntimeError RuntimeError
func (x *TypeAssertionError) RuntimeError()

// Add a dummy struct decl
type TypeAssertionError struct {
	InterfaceString string
	ConcreteString  string
	AssertedString  string
	MissingMethod   string
}

//go:linkname runtime/internal/finalize.SetFinalizer SetFinalizer
func SetFinalizer(obj interface{}, finalizer interface{})

//go:linkname runtime/internal/defers.Goexit Goexit
func Goexit()

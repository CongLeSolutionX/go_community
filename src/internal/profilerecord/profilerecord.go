package profilerecord

type StackRecord struct {
	Stk []uintptr
}

type MemProfileRecord struct {
	AllocBytes, FreeBytes     int64     // number of bytes allocated, freed
	AllocObjects, FreeObjects int64     // number of objects allocated, freed
	Stk                       []uintptr // stack trace
}

func (r *MemProfileRecord) InUseBytes() int64   { return r.AllocBytes - r.FreeBytes }
func (r *MemProfileRecord) InUseObjects() int64 { return r.AllocObjects - r.FreeObjects }

type BlockProfileRecord struct {
	Count  int64
	Cycles int64
	Stk    []uintptr
}

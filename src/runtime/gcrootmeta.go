package runtime

const (
	metaTypeNone = iota
	metaTypeStackFrame
	metaTypeGlobalVariableData
	metaTypeGlobalVariableBSS
	metaTypeFinalizer
	metaTypeDeferred
	metaTypeSpecials
	metaTypeTiny
	metaTypeWBFlush
	metaTypeScheduler
	metaTypeNewAllocation // call site?
)

var GlobalFrameBuffer stackFrameBuffer

type stackFrameBuffer struct {
	frames []stackFrame
	lock   mutex
}

func (buffer *stackFrameBuffer) addframeSynchronized(frame *stkframe, parent int) int {
	if !MemProfileTrackRoots() {
		return -1
	}

	name := funcname(frame.fn)
	metaframe := stackFrame{name: name, parentIdx: parent}

	lock(&buffer.lock)
	idx := len(buffer.frames)
	buffer.frames = append(buffer.frames, metaframe)
	unlock(&buffer.lock)
	return idx
}

// setparent updates the parent index of the frame at frameIdx.
func (buffer *stackFrameBuffer) setparent(frameIdx, parentIdx int) {
	if !MemProfileTrackRoots() {
		return
	}

	lock(&buffer.lock)
	buffer.frames[frameIdx].parentIdx = parentIdx
	unlock(&buffer.lock)
}

// A reference to the stackframe info.
type stackFrame struct {
	name string
	// The index of our parent: -1 if no parent
	parentIdx int
}

type meta struct {
	// What kind of root is this?
	metaType byte

	// The interpretation of the following fields depends on `metaType`.
	// For stack-traces, p is the index of the frame in the stack-frame list
	// For other types, p is the address of the root keeping the memory alive
	p uintptr
}

func (m meta) updatePointer(p uintptr) meta {
	if m.metaType != metaTypeStackFrame && m.p == 0 {
		m.p = p
	}
	return m
}

func (s stackFrame) String() string {
	return s.name
}

// Note: this assumes that the MemProfileRate does not change during a GC cycle.
// Changing this variable is discouraged by the docs so we're not currently
// worrying about it.
// TODO: cache the value when we start each GC run to avoid TOCTTOU risk
func MemProfileTrackRoots() bool {
	return debug.trackroots != 0 && MemProfileRate > 0
}

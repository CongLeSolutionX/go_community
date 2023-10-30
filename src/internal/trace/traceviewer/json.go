package traceviewer

import (
	"encoding/json"
	"io"
	"strconv"

	"internal/trace/traceviewer/format"
)

type TraceConsumer struct {
	ConsumeTimeUnit    func(unit string)
	ConsumeViewerEvent func(v *format.Event, required bool)
	ConsumeViewerFrame func(key string, f format.Frame)
	Flush              func()
}

func ViewerDataTraceConsumer(w io.Writer, start, end int64) TraceConsumer {
	allFrames := make(map[string]format.Frame)
	requiredFrames := make(map[string]format.Frame)
	enc := json.NewEncoder(w)
	written := 0
	index := int64(-1)

	io.WriteString(w, "{")
	return TraceConsumer{
		ConsumeTimeUnit: func(unit string) {
			io.WriteString(w, `"displayTimeUnit":`)
			enc.Encode(unit)
			io.WriteString(w, ",")
		},
		ConsumeViewerEvent: func(v *format.Event, required bool) {
			index++
			if !required && (index < start || index > end) {
				// not in the range. Skip!
				return
			}
			WalkStackFrames(allFrames, v.Stack, func(id int) {
				s := strconv.Itoa(id)
				requiredFrames[s] = allFrames[s]
			})
			WalkStackFrames(allFrames, v.EndStack, func(id int) {
				s := strconv.Itoa(id)
				requiredFrames[s] = allFrames[s]
			})
			if written == 0 {
				io.WriteString(w, `"traceEvents": [`)
			}
			if written > 0 {
				io.WriteString(w, ",")
			}
			enc.Encode(v)
			// TODO: get rid of the extra \n inserted by enc.Encode.
			// Same should be applied to splittingTraceConsumer.
			written++
		},
		ConsumeViewerFrame: func(k string, v format.Frame) {
			allFrames[k] = v
		},
		Flush: func() {
			io.WriteString(w, `], "stackFrames":`)
			enc.Encode(requiredFrames)
			io.WriteString(w, `}`)
		},
	}
}

// WalkStackFrames calls fn for id and all of its parent frames from allFrames.
func WalkStackFrames(allFrames map[string]format.Frame, id int, fn func(id int)) {
	for id != 0 {
		f, ok := allFrames[strconv.Itoa(id)]
		if !ok {
			break
		}
		fn(id)
		id = f.Parent
	}
}

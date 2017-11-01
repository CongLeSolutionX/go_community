package main

import (
	"bytes"
	"fmt"
	"html/template"
	"internal/trace"
	"log"
	"math"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

func init() {
	// TODO: We could classify other types of spans, too, such as
	// GC events.
	http.HandleFunc("/usertasks", httpUserTasks)
	http.HandleFunc("/usertask", httpUserTask)
}

type taskDesc struct {
	name string
	// TODO: task pc.
	id         uint64
	events     []*trace.Event // sorted based on the first timestamp
	spans      []*spanDesc    // sorted based on the first timestamp, and then the last timestamp
	goroutines map[uint64][]*trace.Event
	create     *trace.Event
	end        *trace.Event
	last       *trace.Event // TODO: delete.
	// index to the last GC start event when the task started
	// and the last GC start events when the task ended.
	gc [2]int

	parent   *taskDesc
	children []*taskDesc
}

func newTaskDesc(id uint64) *taskDesc {
	return &taskDesc{
		id:         id,
		name:       "unknown",
		goroutines: make(map[uint64][]*trace.Event),
	}
}

type spanDesc struct {
	name  string
	task  *taskDesc
	goid  uint64
	start *trace.Event
	end   *trace.Event
}

func (span *spanDesc) firstTimestamp() int64 {
	if span.start != nil {
		return span.start.Ts
	}
	return span.task.firstTimestamp()
}

func (span *spanDesc) lastTimestamp() int64 {
	if span.end != nil {
		return span.end.Ts
	}
	return span.task.lastTimestamp()
}

func (span *spanDesc) duration() time.Duration {
	return time.Duration(span.lastTimestamp() - span.firstTimestamp())
}

// overlappingDuration returns the time duration where the specified event
// overlaps with the task if the event is *Start type events whose Link is
// set if the corresponding end event is in the trace.
func (task *taskDesc) overlappingDuration(ev *trace.Event) (time.Duration, bool) {
	s := ev.Ts
	e := lastTimestamp()
	if ev.Link != nil {
		e = ev.Link.Ts
	}

	if task.firstTimestamp() > e || task.lastTimestamp() < s {
		return 0, false
	}

	switch typ := ev.Type; typ {
	case trace.EvGCStart, trace.EvGCSTWStart:
		// Global events
		if t := task.firstTimestamp(); s < t {
			s = t
		}
		if t := task.lastTimestamp(); e > t {
			e = t
		}
		return time.Duration(e-s) * time.Nanosecond, e-s > 0

	case trace.EvGCSweepStart, trace.EvGoStart, trace.EvGoStartLabel, trace.EvGCMarkAssistStart:
		// Goroutine-local events.
		goid := ev.G
		var overlapping int64
		var lastSpanEnd int64 // the end of previous overlapping span
		for _, span := range task.spans {
			if span.goid != goid {
				continue
			}
			ss, se := span.firstTimestamp(), span.lastTimestamp()
			if ss < lastSpanEnd { // skip nested spans
				continue
			}
			if s > se || e < ss { // not overlapping
				continue
			}
			lastSpanEnd = se

			if ss < s {
				ss = s
			}
			if e < se {
				se = e
			}
			overlapping += se - ss
		}
		return time.Duration(overlapping) * time.Nanosecond, overlapping > 0 // TODO: || has annotation
	}
	return 0, false
}

func (task *taskDesc) complete() bool {
	return task.create != nil && task.end != nil
}

// tree returns all the task nodes in the subtree rooted by this task.
func (task *taskDesc) tree() []*taskDesc {
	if task == nil {
		return nil
	}
	res := []*taskDesc{task}
	for i := 0; len(res[i:]) > 0; i++ {
		t := res[i]
		for _, c := range t.children {
			res = append(res, c)
		}
	}
	return res
}

func lastTimestamp() int64 {
	a, err := annotationAnalysis()
	if err != nil {
		panic(err)
	}
	return a.lastTS
}

func firstTimestamp() int64 {
	a, err := annotationAnalysis()
	if err != nil {
		panic(err)
	}
	return a.firstTS
}

func (task *taskDesc) lastTimestamp() int64 {
	if task.end != nil {
		return task.end.Ts
	}
	if task.last != nil {
		return task.last.Ts
	}
	return lastTimestamp()
}

func (task *taskDesc) firstTimestamp() int64 {
	if task.create != nil {
		return task.create.Ts
	}
	return firstTimestamp()
}

func (task *taskDesc) duration() time.Duration {
	return time.Duration(task.lastTimestamp() - task.firstTimestamp())
}

func (task *taskDesc) Goroutines() map[uint64]bool {
	res := map[uint64]bool{}
	for k := range task.goroutines {
		res[k] = true
	}
	return res
}

// RelatedGoroutines returns IDs of goroutines related to the task. A goroutine
// is related to the task if user annotation activities for the task occurred.
// If non-zero depth is provided, this searches all events with BFS and includes
// goroutines unblocked any of related goroutines to the result.
func (task *taskDesc) RelatedGoroutines(events []*trace.Event, depth int) map[uint64]bool {
	start, end := task.firstTimestamp(), task.lastTimestamp()
	gmap := task.Goroutines()
	for i := 0; i < depth; i++ {
		gmap1 := make(map[uint64]bool)
		for g := range gmap {
			gmap1[g] = true
		}
		for _, ev := range events {
			if ev.Ts < start || ev.Ts > end {
				continue
			}
			if ev.Type == trace.EvGoUnblock && gmap[ev.Args[0]] {
				gmap1[ev.G] = true
			}
			gmap = gmap1
		}
	}
	gmap[0] = true // for GC events
	return gmap
}

type keyValue struct {
	K, V string
	Time time.Duration // time since start of the task
}

func (task *taskDesc) keyValues() []keyValue {
	baseTS := task.firstTimestamp()
	var res []keyValue
	for _, ev := range task.events {
		if ev.Type != trace.EvUserLog {
			continue
		}
		res = append(res, keyValue{
			K:    ev.SArgs[0],
			V:    ev.SArgs[1],
			Time: time.Duration(ev.Ts - baseTS),
		})
	}
	return res
}

func httpUserTasks(w http.ResponseWriter, r *http.Request) {
	res, err := annotationAnalysis()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	summary := make(map[string]taskStats)
	for _, task := range res.Tasks {

		stats, ok := summary[task.name]
		if !ok {
			stats.Type = task.name
		}

		stats.add(task, task.duration())
		summary[task.name] = stats
	}

	// Sort tasks by type.
	userTasks := make([]taskStats, 0, len(summary))
	for _, stats := range summary {
		userTasks = append(userTasks, stats)
	}
	sort.Slice(userTasks, func(i, j int) bool {
		return userTasks[i].Type < userTasks[j].Type
	})

	// Emit table.
	err = templUserTaskTypes.Execute(w, userTasks)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to execute template: %v", err), http.StatusInternalServerError)
		return
	}
}

type userTaskFilter struct {
	conditions []func(*taskDesc) bool
}

func (f *userTaskFilter) filter(t *taskDesc) bool {
	for _, c := range f.conditions {
		if !c(t) {
			return false
		}
	}
	return true
}

func buildUserTaskFilter(r *http.Request) (func(*taskDesc) bool, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}

	var f userTaskFilter
	param := r.Form
	if typ, ok := param["type"]; ok && len(typ) > 0 {
		f.conditions = append(f.conditions, func(t *taskDesc) bool {
			return t.name == typ[0]
		})
	}
	if complete := r.FormValue("complete"); complete == "1" {
		f.conditions = append(f.conditions, func(t *taskDesc) bool {
			return t.complete()
		})
	} else if complete == "0" {
		f.conditions = append(f.conditions, func(t *taskDesc) bool {
			return !t.complete()
		})
	}
	if lat, err := time.ParseDuration(r.FormValue("latmin")); err == nil {
		f.conditions = append(f.conditions, func(t *taskDesc) bool {
			return t.duration() >= lat
		})
	}
	if lat, err := time.ParseDuration(r.FormValue("latmax")); err == nil {
		f.conditions = append(f.conditions, func(t *taskDesc) bool {
			return t.duration() < lat
		})
	}

	return f.filter, nil
}

type userLog struct {
	*trace.Event
}

func (ul userLog) String() string {
	k, v := ul.Event.SArgs[0], ul.Event.SArgs[1]
	if k == "" {
		return v
	}
	return fmt.Sprintf("%v=%v", k, v)
}

func httpUserTask(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("type")
	filter, err := buildUserTaskFilter(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res, err := annotationAnalysis()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type event struct {
		WhenString string
		Elapsed    time.Duration
		Go         uint64
		What       string
		// TODO: include stack trace of creation time
	}
	type entry struct {
		WhenString     string
		ID             uint64
		Duration       time.Duration
		Complete       bool
		Events         []event
		StartMS, EndMS float64
		GCTime         time.Duration
	}

	var data []entry

	base := time.Duration(0) // trace start
	for _, task := range res.Tasks {
		if !filter(task) {
			continue
		}
		var events []event
		var last time.Duration

		for i, ev := range task.events {
			when := time.Duration(ev.Ts)*time.Nanosecond - base // time since trace start
			if i == 0 {
				last = time.Duration(ev.Ts) * time.Nanosecond
			}

			elapsed := time.Duration(ev.Ts)*time.Nanosecond - last
			goid := ev.G
			var what string

			switch ev.Type {
			case trace.EvGoCreate:
				what = fmt.Sprintf("new goroutine %d", ev.Args[0])
			case trace.EvUserLog:
				what = userLog{ev}.String()
			case trace.EvUserSpan:
				if ev.Args[1] == 0 {
					duration := "unknown"
					if ev.Link != nil {
						duration = (time.Duration(ev.Link.Ts-ev.Ts) * time.Nanosecond).String()
					}
					what = fmt.Sprintf("span %s started (duration: %v)", ev.SArgs[0], duration)
				}
			case trace.EvUserTaskCreate:
				what = fmt.Sprintf("task %v (id %d, parent %d) created", ev.SArgs[0], ev.Args[0], ev.Args[1])
				// TODO: add child task creation events into the parent task events
			case trace.EvUserTaskEnd:
				what = "end of task"
			}
			if what == "" {
				continue
			}
			events = append(events, event{
				WhenString: fmt.Sprintf("%2.9f", when.Seconds()),
				Elapsed:    elapsed,
				What:       what,
				Go:         goid,
			})
			last = time.Duration(ev.Ts) * time.Nanosecond
		}

		var gcTime time.Duration
		if len(res.gcs) > 0 {
			lastGC := len(res.gcs) - 1
			if task.complete() {
				lastGC = task.gc[1]
			}
			for i := task.gc[0]; i <= lastGC; i++ {
				if i < 0 {
					continue // task started before the first GC in the trace
				}
				overlapping, _ := task.overlappingDuration(res.gcs[i])
				gcTime += overlapping
			}
		}
		data = append(data, entry{
			WhenString: fmt.Sprintf("%2.9fs", (time.Duration(task.firstTimestamp())*time.Nanosecond - base).Seconds()),
			Duration:   task.duration(),
			ID:         task.id,
			Complete:   task.complete(),
			Events:     events,
			StartMS:    float64(task.firstTimestamp()) / 1e6,
			EndMS:      float64(task.lastTimestamp()) / 1e6,
			GCTime:     gcTime,
		})
	}
	sort.Slice(data, func(i, j int) bool {
		return data[i].Duration < data[j].Duration
	})

	// Emit table.
	err = templUserTaskType.Execute(w, struct {
		Name  string
		Entry []entry
	}{
		Name:  name,
		Entry: data,
	})
	if err != nil {
		log.Printf("failed to execute template: %v", err)
		http.Error(w, fmt.Sprintf("failed to execute template: %v", err), http.StatusInternalServerError)
		return
	}
}

type durationHistogram struct {
	Count                int
	Buckets              []int
	MinBucket, MaxBucket int
}

// Five buckets for every power of 10.
var logDiv = math.Log(math.Pow(10, 1.0/5))

func (h *durationHistogram) add(d time.Duration) {
	var bucket int
	if d > 0 {
		bucket = int(math.Log(float64(d)) / logDiv)
	}
	if len(h.Buckets) <= bucket {
		h.Buckets = append(h.Buckets, make([]int, bucket-len(h.Buckets)+1)...)
		h.Buckets = h.Buckets[:cap(h.Buckets)]
	}
	h.Buckets[bucket]++
	if bucket < h.MinBucket || h.MaxBucket == 0 {
		h.MinBucket = bucket
	}
	if bucket > h.MaxBucket {
		h.MaxBucket = bucket
	}
	h.Count++
}

func (h *durationHistogram) BucketMin(bucket int) time.Duration {
	return time.Duration(math.Exp(float64(bucket) * logDiv))
}

func niceDuration(d time.Duration) string {
	var rnd time.Duration
	var unit string
	switch {
	case d < 10*time.Microsecond:
		rnd, unit = time.Nanosecond, "ns"
	case d < 10*time.Millisecond:
		rnd, unit = time.Microsecond, "µs"
	case d < 10*time.Second:
		rnd, unit = time.Millisecond, "ms"
	default:
		rnd, unit = time.Second, "s "
	}
	return fmt.Sprintf("%d%s", d/rnd, unit)
}

func (h *durationHistogram) ToHTML(urlmaker func(min, max time.Duration) string) template.HTML {
	if h == nil || h.Count == 0 {
		return template.HTML("")
	}

	const barWidth = 400

	maxCount := 0
	for _, count := range h.Buckets {
		if count > maxCount {
			maxCount = count
		}
	}

	w := new(bytes.Buffer)
	fmt.Fprintf(w, `<table>`)
	for i := h.MinBucket; i <= h.MaxBucket; i++ {
		// Tick label.
		fmt.Fprintf(w, `<tr><td class="histoTime" align="right"><a href=%s>%s</a></td>`, urlmaker(h.BucketMin(i), h.BucketMin(i+1)), niceDuration(h.BucketMin(i)))
		// Bucket bar.
		width := h.Buckets[i] * barWidth / maxCount
		fmt.Fprintf(w, `<td><div style="width:%dpx;background:blue;top:.6em;position:relative">&nbsp;</div></td>`, width)
		// Bucket count.
		fmt.Fprintf(w, `<td align="right"><div style="top:.6em;position:relative">%d</div></td>`, h.Buckets[i])
		fmt.Fprintf(w, "</tr>\n")

	}
	// Final tick label.
	fmt.Fprintf(w, `<tr><td align="right">%s</td></tr>`, niceDuration(h.BucketMin(h.MaxBucket+1)))
	fmt.Fprintf(w, `</table>`)
	return template.HTML(w.String())
}

func (h *durationHistogram) String() string {
	const barWidth = 40

	labels := []string{}
	maxLabel := 0
	maxCount := 0
	for i := h.MinBucket; i <= h.MaxBucket; i++ {
		// TODO: This formatting is pretty awful.
		label := fmt.Sprintf("[%-12s%-11s)", h.BucketMin(i).String()+",", h.BucketMin(i+1))
		labels = append(labels, label)
		if len(label) > maxLabel {
			maxLabel = len(label)
		}
		count := h.Buckets[i]
		if count > maxCount {
			maxCount = count
		}
	}

	w := new(bytes.Buffer)
	for i := h.MinBucket; i <= h.MaxBucket; i++ {
		count := h.Buckets[i]
		bar := count * barWidth / maxCount
		fmt.Fprintf(w, "%*s %-*s %d\n", maxLabel, labels[i-h.MinBucket], barWidth, strings.Repeat("█", bar), count)
	}
	return w.String()
}

type taskStats struct {
	Type       string
	Count      int
	Complete   durationHistogram
	Incomplete durationHistogram
}

func (s *taskStats) UserTaskURL(complete bool) func(min, max time.Duration) string {
	return func(min, max time.Duration) string {
		return fmt.Sprintf("/usertask?type=%s&complete=%v&latmin=%v&latmax=%v", template.URLQueryEscaper(s.Type), template.URLQueryEscaper(complete), template.URLQueryEscaper(min), template.URLQueryEscaper(max))
	}
}

func (s *taskStats) add(task *taskDesc, duration time.Duration) {
	if task.complete() {
		s.Complete.add(duration)
	} else {
		s.Incomplete.add(duration)
	}
	s.Count++
}

var templUserTaskTypes = template.Must(template.New("").Parse(`
<html>
<style type="text/css">
.histoTime {
   width: 20%;
   white-space:nowrap;
}

</style>
<body>
<table border="1" sortable="1">
<tr>
<th>Task type</th>
<th>Count</th>
<th>Duration distribution( complete )</th>
<th>Duration distribution( incomplete )</th>
</tr>
{{range $}}
  <tr>
    <td>{{.Type}}</td>
    <td><a href="/usertask?type={{.Type}}">{{.Count}}</a></td>
    <td>{{.Complete.ToHTML (.UserTaskURL true)}}</td>
    <td>{{.Incomplete.ToHTML (.UserTaskURL false)}}</td>
  </tr>
{{end}}
</table>
</body>
</html>
`))

var templUserTaskType = template.Must(template.New("userTask").Funcs(template.FuncMap{
	"elapsed":   elapsed,
	"trimSpace": strings.TrimSpace,
}).Parse(`
<html>
<head> <title>User Task: {{.Name}} </title> </head>
        <style type="text/css">
                body {
                        font-family: sans-serif;
                }
                table#req-status td.family {
                        padding-right: 2em;
                }
                table#req-status td.active {
                        padding-right: 1em;
                }
                table#req-status td.empty {
                        color: #aaa;
                }
                table#reqs {
                        margin-top: 1em;
                }
                table#reqs tr.first {
                        font-weight: bold;
                }
                table#reqs td {
                        font-family: monospace;
                }
                table#reqs td.when {
                        text-align: right;
                        white-space: nowrap;
                }
                table#reqs td.elapsed {
                        padding: 0 0.5em;
                        text-align: right;
                        white-space: pre;
                        width: 10em;
                }
                address {
                        font-size: smaller;
                        margin-top: 5em;
                }
        </style>
<body>

<h1>User Task: {{.Name}}</h1>

<table id="reqs">
<tr><th>When</th><th>Elapsed</th><th>Goroutine ID</th><th>Events</th></tr>
     {{range $el := $.Entry}}
        <tr class="first">
                <td class="when">{{$el.WhenString}}</td>
                <td class="elapsed">{{$el.Duration}}</td>
		<td></td>
                <td><a href="/trace?taskid={{$el.ID}}#{{$el.StartMS}}:{{$el.EndMS}}">Task {{$el.ID}}</a> ({{if .Complete}}complete{{else}}incomplete{{end}})</td>
        </tr>
        {{range $el.Events}}
        <tr>
                <td class="when">{{.WhenString}}</td>
                <td class="elapsed">{{elapsed .Elapsed}}</td>
		<td class="goid">{{.Go}}</td>
                <td>{{.What}}</td>
        </tr>
        {{end}}
	<tr>
		<td></td>
		<td></td>
		<td></td>
		<td>GC:{{$el.GCTime}}</td>
    {{end}}
</body>
</html>
`))

var (
	annotationsOnce sync.Once
	annotations     AnnotationAnalysisResult
)

type AnnotationAnalysisResult struct {
	err     error
	firstTS int64
	lastTS  int64
	Tasks   map[uint64]*taskDesc
	gcs     []*trace.Event // GCStart events, sorted.
}

func annotationAnalysis() (AnnotationAnalysisResult, error) {
	annotationsOnce.Do(func() {
		events, err := parseEvents()
		if err != nil {
			annotations.err = err
			return
		}
		annotations = doAnnotationAnalysis(events)
	})
	return annotations, annotations.err
}

func doAnnotationAnalysis(events []*trace.Event) AnnotationAnalysisResult {
	if len(events) == 0 {
		return AnnotationAnalysisResult{}
	}

	tasks := map[uint64]*taskDesc{}
	activeSpans := map[uint64][]*spanDesc{} // goid->span
	var gcs []*trace.Event                  // gc start events

	for _, ev := range events {
		goid := ev.G

		switch typ := ev.Type; typ {
		case trace.EvGCStart:
			gcs = append(gcs, ev)

		case trace.EvUserTaskCreate, trace.EvUserSpan, trace.EvUserTaskEnd, trace.EvUserLog:
			taskid := ev.Args[0]
			parentID := ev.Args[1]

			task := tasks[taskid]
			if task == nil {
				task = newTaskDesc(taskid)
				task.gc[0] = len(gcs) - 1
				tasks[taskid] = task
			}
			ptask := tasks[parentID]
			if ptask == nil {
				ptask = newTaskDesc(parentID)
				task.gc[0] = len(gcs) - 1
				tasks[parentID] = ptask
			}

			task.parent = ptask
			ptask.children = append(ptask.children, task)

			task.events = append(task.events, ev)

			task.goroutines[goid] = append(task.goroutines[goid], ev)

			switch typ {
			case trace.EvUserTaskCreate:
				task.name = ev.SArgs[0]
				task.create = ev
			case trace.EvUserTaskEnd:
				task.end = ev
				task.gc[1] = len(gcs) - 1
			case trace.EvUserSpan:
				spans := activeSpans[goid]
				switch mode := ev.Args[1]; mode {
				case 0: // start
					s := &spanDesc{
						name:  ev.SArgs[0],
						task:  task,
						goid:  goid,
						start: ev,
					}
					activeSpans[goid] = append(spans, s)
					task.spans = append(task.spans, s)
				case 1: // end
					if len(spans) == 0 {
						// If tracing happens in the middle of a span, unmatching span end can appear.
						break
					}
					s := spans[len(spans)-1]
					if s.task.id != taskid {
						return AnnotationAnalysisResult{
							err: fmt.Errorf("misuse of span is detected: span (task %v, %s) ends while span (task %v, %s) is active", taskid, ev.SArgs[0], s.task.id, s.name)}
					}
					s.end = ev
					s.start.Link = ev // TODO: move this to internal/trace/parser.go?
					activeSpans[goid] = spans[:len(spans)-1]
				}
			}

			if task.last == nil || task.last.Ts <= ev.Ts {
				task.last = ev
			}

		case trace.EvGoCreate:
			// ASSUMPTION: goroutine inherits the current task of the creator

			span := activeSpans[goid]
			if len(span) == 0 {
				continue
			}
			task := span[len(span)-1].task
			task.goroutines[goid] = []*trace.Event{ev}

		case trace.EvGoEnd, trace.EvGoStop:
			spans := activeSpans[goid]
			for _, s := range spans {
				s.end = ev
				task := tasks[s.task.id]
				if task.last == nil || task.last.Ts <= ev.Ts {
					task.last = ev
				}
			}
			delete(activeSpans, goid)

		}
	}

	// sorting spans based on the timestamps
	for _, task := range tasks {
		sort.Slice(task.spans, func(i, j int) bool {
			si, sj := task.spans[i].firstTimestamp(), task.spans[j].firstTimestamp()
			if si != sj {
				return si < sj
			}

			return task.spans[i].lastTimestamp() < task.spans[i].lastTimestamp()
		})
	}
	return AnnotationAnalysisResult{
		Tasks:   tasks,
		firstTS: events[0].Ts,
		lastTS:  events[len(events)-1].Ts,
		gcs:     gcs,
	}
}

func elapsed(d time.Duration) string {
	b := []byte(fmt.Sprintf("%.9f", d.Seconds()))

	// For subsecond durations, blank all zeros before decimal point,
	// and all zeros between the decimal point and the first non-zero digit.
	if d < time.Second {
		dot := bytes.IndexByte(b, '.')
		for i := 0; i < dot; i++ {
			b[i] = ' '
		}
		for i := dot + 1; i < len(b); i++ {
			if b[i] == '0' {
				b[i] = ' '
			} else {
				break
			}
		}
	}

	return string(b)
}

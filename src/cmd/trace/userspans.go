// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// User span-related profiles.

package main

import (
	"bytes"
	"fmt"
	"html/template"
	"internal/trace"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"
)

func init() {
	// TODO: We could classify other types of spans, too, such as
	// GC events.
	http.HandleFunc("/userspantypes", httpUserSpanTypes)
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

func (h *durationHistogram) bucketMin(bucket int) time.Duration {
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

func (h *durationHistogram) ToHTML() template.HTML {
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
		fmt.Fprintf(w, `<tr><td align="right">%s</td>`, niceDuration(h.bucketMin(i)))
		// Bucket bar.
		width := h.Buckets[i] * barWidth / maxCount
		fmt.Fprintf(w, `<td><div style="width:%dpx;background:black;top:.6em;position:relative">&nbsp;</div></td>`, width)
		// Bucket count.
		fmt.Fprintf(w, `<td align="right"><div style="top:.6em;position:relative">%d</div></td>`, h.Buckets[i])
		fmt.Fprintf(w, "</tr>\n")

	}
	// Final tick label.
	fmt.Fprintf(w, `<tr><td align="right">%s</td></tr>`, niceDuration(h.bucketMin(h.MaxBucket+1)))
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
		label := fmt.Sprintf("[%-12s%-11s)", h.bucketMin(i).String()+",", h.bucketMin(i+1))
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

type spanStats struct {
	Type string
	Hist durationHistogram
}

func httpUserSpanTypes(w http.ResponseWriter, r *http.Request) {
	events, err := parseEvents()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Collect user spans by type.
	userSpanMap := make(map[string]spanStats)
	for _, ev := range events {
		if ev.Type != trace.EvUserSpan {
			continue
		}
		stats, ok := userSpanMap[ev.SArgs[0]]
		if !ok {
			stats.Type = ev.SArgs[0]
		}
		stats.Hist.add(time.Duration(ev.Link.Ts - ev.Ts))
		userSpanMap[ev.SArgs[0]] = stats
	}

	// Sort spans by type.
	userSpans := make([]spanStats, 0, len(userSpanMap))
	for _, stats := range userSpanMap {
		userSpans = append(userSpans, stats)
	}
	sort.Slice(userSpans, func(i, j int) bool {
		return userSpans[i].Type < userSpans[j].Type
	})

	// Emit table.
	err = templUserSpanTypes.Execute(w, userSpans)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to execute template: %v", err), http.StatusInternalServerError)
		return
	}
}

var templUserSpanTypes = template.Must(template.New("").Parse(`
<html>
<body>
<table border="1" sortable="1">
<tr>
<th>Span type</th>
<th>Count</th>
<th>Duration distribution</th>
</tr>
{{range $}}
  <tr>
    <td>{{.Type}}</td>
    <td>{{.Hist.Count}}</td>
    <td>{{.Hist.ToHTML}}</td>
  </tr>
{{end}}
</table>
</body>
</html>
`))

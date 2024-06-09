// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package stats

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"time"
)

func (d Dist[T]) Plot(pngPath, x, y string) {
	logBase := 0 // TODO

	buf := os.Stdout
	fmt.Fprintf(buf, "set terminal pngcairo\n")
	fmt.Fprintf(buf, "set output %q\n", pngPath)
	//fmt.Fprintf(buf, "set title %q\n", title)

	fmt.Fprintf(buf, "set xlabel %q\n", x)

	fmt.Fprintf(buf, "set ylabel %q\n", y)
	fmt.Fprintf(buf, "set yrange [0:%d]\n", len(d.vals))
	fmt.Fprintf(buf, "set ytics nomirror\n")
	fmt.Fprintf(buf, "set ytics add (\"n=%d\" %d)\n", len(d.vals), len(d.vals))

	fmt.Fprintf(buf, "set y2label %q\n", "quantile")
	fmt.Fprintf(buf, "set y2range [0:1]\n")
	fmt.Fprintf(buf, "set y2tics nomirror\n")

	slices.Sort(d.vals)

	fmtVal := func(val T) string { return fmt.Sprint(val) }
	if x == "zombie percent" {
		fmtVal = func(val T) string { return fmt.Sprintf("%.1f%%", float64(val)) }
	}

	// Do we need to trim the range?
	mid := d.Quantiles(0, 0.01, 0.99, 1)
	lo, hi := mid[0], mid[3]
	if float64(lo) < 1.5*float64(mid[1])-0.5*float64(mid[2]) {
		lo = mid[1]
	}
	if float64(hi) > 1.5*float64(mid[2])-0.5*float64(mid[1]) {
		hi = mid[2]
	}
	if lo == hi {
		lo -= 1
		hi += 1
	}

	// Set log
	if logBase != 0 {
		fmt.Fprintf(buf, "set log x %d\n", logBase)
		if lo < 0 && hi > 0 {
			for _, val := range d.vals {
				if val > 0 {
					lo = val
					break
				}
			}
		}
	}

	// Format duration tick marks.
	if lo, ok := any(lo).(time.Duration); ok {
		hi := any(hi).(time.Duration)
		var tics []string
		for _, tic := range durationTicks(lo, hi, logBase) {
			tics = append(tics, fmt.Sprintf("%q %d %d", tic.label, tic.pos, tic.level))
		}
		fmt.Fprintf(buf, "set xtics (%s)\n", strings.Join(tics, ","))
	}

	// Set X range.
	fmt.Fprintf(buf, "set xrange [%v:%v]\n", float64(lo), float64(hi))

	// Show min and max
	fmt.Fprintf(buf, "set label %q at graph 0,0 offset 0,char -1.75\n", "min "+fmtVal(d.vals[0]))
	fmt.Fprintf(buf, "set label %q at graph 1,0 right offset 0,char -1.75\n", "max "+fmtVal(d.vals[len(d.vals)-1]))

	// Plot data.
	fmt.Fprintf(buf, "plot '-' notitle with steps, ")
	fmt.Fprintf(buf, "'-' notitle axes x1y2 with labels left offset char 0.2,char -0.25 point ps 2,")
	fmt.Fprintf(buf, "'-' notitle axes x1y2 with labels right offset char -1.2,char -0.25 point ps 2\n")
	for i, val := range d.vals {
		fmt.Fprintf(buf, "%v %d\n", float64(val), i)
	}
	fmt.Fprintf(buf, "e\n")

	// Mark quantiles.
	qs := []float64{0.05, 0.25, 0.5, 0.75, 0.95}
	qvs := d.Quantiles(qs...)
	for i, val := range qvs {
		q := qs[i]
		if val <= (lo+hi)/2 {
			fmt.Fprintf(buf, "%v %g %s\n", float64(val), q, fmtVal(val))
		}
	}
	fmt.Fprintf(buf, "e\n")
	for i, val := range qvs {
		q := qs[i]
		if val > (lo+hi)/2 {
			fmt.Fprintf(buf, "%v %g %s\n", float64(val), q, fmtVal(val))
		}
	}
	fmt.Fprintf(buf, "e\n")

	fmt.Fprintf(buf, "unset output\n")
	fmt.Fprintf(buf, "reset\n")
}

type durationTick struct {
	label string
	pos   time.Duration
	level int
}

func durationTicks(lo, hi time.Duration, logBase int) []durationTick {
	var out []durationTick
	add := func(d time.Duration, level int) { out = append(out, durationTick{d.String(), d, level}) }

	if logBase != 0 {
		for _, d := range makeDurationLevels(logBase) {
			add(d, 0)
		}
	} else {
		const maxTicks = 8
		for level := range durationLevels {
			tlo, step, n := ticksAt(lo, hi, level)
			if n <= maxTicks {
				// We found our major ticks. Minor ticks are one level down.
				_, minStep, _ := ticksAt(lo, hi, level-1)
				// Start one below tlo just to produce the right minors.
				for major := -1; major < n; major++ {
					// Emit major tick.
					pos := tlo + step*time.Duration(major)
					add(pos, 0)
					// Emit minor ticks between pos and pos+step.
					for minor := 1; ; minor++ {
						minPos := pos + minStep*time.Duration(minor)
						if minPos >= pos+step {
							break
						}
						add(minPos, 1)
					}
				}
				break
			}
		}
	}
	return out
}

func makeDurationLevels(factors ...int) []time.Duration {
	var out []time.Duration

	fi := 0
	next := func() time.Duration {
		factor := time.Duration(factors[fi%len(factors)])
		fi++
		return factor
	}

	d := time.Nanosecond
	for d < time.Minute {
		out = append(out, d)
		d *= next()
	}
	d, fi = time.Minute, 0
	for d < time.Hour {
		out = append(out, d)
		d *= next()
	}
	d, fi = time.Hour, 0
	for d <= 100000*time.Hour {
		out = append(out, d)
		d *= next()
	}
	return out
}

var durationLevels = makeDurationLevels(5, 2)

func ticksAt(lo, hi time.Duration, level int) (start, step time.Duration, n int) {
	step = durationLevels[level]
	// Round lo up to a multiple of step.
	start = ((lo + step - 1) / step) * step
	// Round hi down to a multiple of step.
	stop := (hi / step) * step
	n = 1 + int((stop-start)/step)
	return
}

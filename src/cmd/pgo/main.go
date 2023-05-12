package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
)

// From cmd/compile/internal/pgo.
type CallStat struct {
	Pkg string
	Pos string

	Caller string

	// Call type. Interface must not be Direct.
	Direct    bool
	Interface bool

	Weight int64

	Hottest       string
	HottestWeight int64

	// Specialized callee if != "".
	//
	// Note that this may be different than Hottest because we apply
	// type-check restrictions, which helps distinguish multiple calls on
	// the same line. Hottest doesn't do that.
	Specialized       string
	SpecializedWeight int64
}

func readStats() ([]CallStat, error) {
	var stats []CallStat
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var stat CallStat
		if err := json.Unmarshal(scanner.Bytes(), &stat); err != nil {
			//log.Printf("Failed to unmarshal %q: %v", scanner.Text(), err)
			continue
		}
		stats = append(stats, stat)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return stats, nil
}

type sum struct {
	direct         int64
	indirectFunc   int64
	indirectMethod int64
}

func (s *sum) total() int64 {
	return s.direct + s.indirectFunc + s.indirectMethod
}

func pct(n, d int64) float64 {
	return 100*float64(n)/float64(d)
}

func run() error {
	stats, err := readStats()
	if err != nil {
		return err
	}

	var (
		count             sum
		weight            sum
		hottestWeight     sum
		specializedCount  sum
		specializedWeight sum
	)

	for _, s := range stats {
		if s.Direct {
			count.direct++
			weight.direct += s.Weight
			hottestWeight.direct += s.Weight
		} else if s.Interface {
			count.indirectMethod++
			weight.indirectMethod += s.Weight
			hottestWeight.indirectMethod += s.HottestWeight
			if s.Specialized != "" {
				specializedCount.indirectMethod++
				specializedWeight.indirectMethod += s.SpecializedWeight
			}
		} else {
			count.indirectFunc++
			weight.indirectFunc += s.Weight
			hottestWeight.indirectFunc += s.HottestWeight
		}
	}

	fmt.Printf("Call count breakdown:\n")
	fmt.Printf("\tTotal: %d\n", count.total())
	fmt.Printf("\tDirect: %d (%.2f%% of total)\n", count.direct, pct(count.direct, count.total()))
	fmt.Printf("\tIndirect func: %d (%.2f%% of total)\n", count.indirectFunc, pct(count.indirectFunc, count.total()))
	fmt.Printf("\tInterface method: %d (%.2f%% of total)\n", count.indirectMethod, pct(count.indirectMethod, count.total()))

	fmt.Printf("Call weight breakdown:\n")
	fmt.Printf("\tTotal: %d\n", weight.total())
	fmt.Printf("\tDirect: %d (%.2f%% of total)\n", weight.direct, pct(weight.direct, weight.total()))
	fmt.Printf("\tIndirect func: %d (%.2f%% of total)\n", weight.indirectFunc, pct(weight.indirectFunc, weight.total()))
	fmt.Printf("\tInterface method: %d (%.2f%% of total)\n", weight.indirectMethod, pct(weight.indirectMethod, weight.total()))

	fmt.Printf("Call hottest weight breakdown:\n")
	fmt.Printf("\tTotal: %d (%.2f%% of total)\n", hottestWeight.total(), pct(hottestWeight.total(), weight.total()))
	fmt.Printf("\tDirect: %d (%.2f%% of direct)\n", hottestWeight.direct, pct(hottestWeight.direct, weight.direct))
	fmt.Printf("\tIndirect func: %d (%.2f%% of indirect func)\n", hottestWeight.indirectFunc, pct(hottestWeight.indirectFunc, weight.indirectFunc))
	fmt.Printf("\tInterface method: %d (%.2f%% of interface method)\n", hottestWeight.indirectMethod, pct(hottestWeight.indirectMethod, weight.indirectMethod))

	fmt.Printf("Specialized call count: %d (%.2f%% of total, %.2f%% of interface method)\n", specializedCount.indirectMethod, pct(specializedCount.indirectMethod, count.total()), pct(specializedCount.indirectMethod, count.indirectMethod))
	fmt.Printf("Specialized call weight: %d (%.2f%% of total, %.2f%% of interface method)\n", specializedWeight.indirectMethod, pct(specializedWeight.indirectMethod, weight.total()), pct(specializedWeight.indirectMethod, weight.indirectMethod))

	const topCount = 100
	fmt.Printf("\nTop %d hottest indirect calls:\n", topCount)
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].HottestWeight < stats[j].HottestWeight
	})
	printed := 0
	var topWeight, topHottestWeight int64
	for i := len(stats)-1; i >= 0 && printed < topCount; i-- {
		s := stats[i]
		if s.Direct {
			continue
		}
		spec := "NOT Specialized"
		specExtra := ""
		if s.Specialized != "" {
			spec = "    Specialized"
			if s.Specialized != s.Hottest {
				specExtra = fmt.Sprintf("\t(specialized to %s weight %d)", s.Specialized, s.SpecializedWeight)
			}
		}
		typ := "interface"
		if !s.Interface {
			typ = " function"
		}
		fmt.Printf("\t(%s) (%s) %-40s -> %-40s (weight %d, %.2f%% of callsite weight)%s\t%s\n", spec, typ, s.Caller, s.Hottest, s.HottestWeight, pct(s.HottestWeight, s.Weight), specExtra, s.Pos)
		printed++
		topWeight += s.Weight
		topHottestWeight += s.HottestWeight
	}
	fmt.Printf("Top %d weight: %d (%.2f%% of indirect weight)\n", topCount, topWeight, pct(topWeight, weight.indirectFunc+weight.indirectMethod))
	fmt.Printf("Top %d hottest weight: %d (%.2f%% of indirect hottest weight)\n", topCount, topHottestWeight, pct(topHottestWeight, hottestWeight.indirectFunc+hottestWeight.indirectMethod))

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

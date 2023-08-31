// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package escape

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/ir"
	"cmd/compile/internal/logopt"
	"cmd/compile/internal/types"
	"cmd/internal/src"
	"fmt"
	"strings"
)

// Flag to allow disabling new Go 1.22 escape analysis for interface receivers.
const go122UseIfaceRecvEscapeAnalysis = true

// walkAll computes the minimal dereferences between all pairs of
// locations.
func (b *batch) walkAll() {
	// We use a work queue to keep track of locations that we need
	// to visit, and repeatedly walk until we reach a fixed point.
	//
	// We walk once from each location (including the heap), and
	// then re-enqueue each location on its transition from
	// transient->!transient, !escapes->escapes, and
	// !ifaceRecvEscape->ifaceRecvEscape, which can each
	// happen at most once. So we take Θ(len(e.allLocs)) walks.

	// LIFO queue, has enough room for b.allLocs, b.heapLoc,
	// and b.ifaceRecvLoc.
	todo := make([]*location, 0, len(b.allLocs)+2)
	enqueue := func(loc *location) {
		if !loc.queued {
			todo = append(todo, loc)
			loc.queued = true
		}
	}

	for _, loc := range b.allLocs {
		enqueue(loc)
	}
	enqueue(&b.heapLoc)
	if go122UseIfaceRecvEscapeAnalysis {
		enqueue(&b.ifaceRecvLoc)
	}

	if base.Debug.EscRecvDebug >= 4 {
		// TODO: maybe remove or maybe polish and move to another debug flag
		fmt.Println("walkAll: starting location graph:")
		details := func(loc *location) string {
			switch {
			case loc.escapes && !loc.ifaceRecvEscape:
				return " (escapes: y)"
			case !loc.escapes && loc.ifaceRecvEscape:
				return " (ifaceRecvEscape: y)"
			case loc.escapes && loc.ifaceRecvEscape:
				return " (escapes: y, ifaceRecvEscape: y)"
			default:
				return ""
			}
		}
		for _, loc := range todo {
			fmt.Printf("  location: %v%s in %v\n",
				b.explainLoc(loc), details(loc), loc.curfn)
			for _, e := range loc.edges {
				fmt.Printf("    incoming edge from: %v%s in %v with edge derefs=%d\n",
					b.explainLoc(e.src), details(e.src), e.src.curfn, e.derefs)
			}
		}
	}

	var walkgen uint32
	for len(todo) > 0 {
		root := todo[len(todo)-1]
		todo = todo[:len(todo)-1]
		root.queued = false

		walkgen++
		b.walkOne(root, walkgen, enqueue)
	}
}

// walkOne computes the minimal number of dereferences from root to
// all other locations.
func (b *batch) walkOne(root *location, walkgen uint32, enqueue func(*location)) {
	// The data flow graph has negative edges (from addressing
	// operations), so we use the Bellman-Ford algorithm. However,
	// we don't have to worry about infinite negative cycles since
	// we bound intermediate dereference counts to 0.

	root.walkgen = walkgen
	root.derefs = 0
	root.dst = nil

	todo := []*location{root} // LIFO queue
	for len(todo) > 0 {
		l := todo[len(todo)-1]
		todo = todo[:len(todo)-1]

		derefs := l.derefs

		// If l.derefs < 0, then l's address flows to root.
		addressOf := derefs < 0
		if addressOf {
			// For a flow path like "root = &l; l = x",
			// l's address flows to root, but x's does
			// not. We recognize this by lower bounding
			// derefs at 0.
			derefs = 0

			// If l's address flows to a non-transient
			// location, then l can't be transiently
			// allocated.
			if !root.transient && l.transient {
				l.transient = false
				enqueue(l)
			}
		}

		if b.outlives(root, l) {
			// l's value flows to root. If l is a function
			// parameter and root is the heap or a
			// corresponding result parameter, then record
			// that value flow for tagging the function
			// later.
			if l.isName(ir.PPARAM) {
				if (logopt.Enabled() || base.Flag.LowerM >= 2) && !l.escapes {
					if base.Flag.LowerM >= 2 {
						fmt.Printf("%s: parameter %v leaks to %s in %v with derefs=%d:\n", base.FmtPos(l.n.Pos()), l.n, b.explainLoc(root), ir.FuncName(l.curfn), derefs)
					}
					explanation := b.explainPath(root, l)
					if logopt.Enabled() {
						var e_curfn *ir.Func // TODO(mdempsky): Fix.
						// TODO: why is e_curfn used instead of l.curfn?
						logopt.LogOpt(l.n.Pos(), "leak", "escape", ir.FuncName(e_curfn),
							fmt.Sprintf("parameter %v leaks to %s with derefs=%d", l.n, b.explainLoc(root), derefs), explanation)
					}
				}
				l.leakTo(root, derefs)
			}

			// If l's address flows somewhere that
			// outlives it, then l needs to be heap
			// allocated.
			if addressOf && !l.escapes {
				if logopt.Enabled() || base.Flag.LowerM >= 2 {
					if base.Flag.LowerM >= 2 {
						fmt.Printf("%s: %v escapes to heap in %v:\n", base.FmtPos(l.n.Pos()), l.n, ir.FuncName(l.curfn))
					}
					explanation := b.explainPath(root, l)
					if logopt.Enabled() {
						var e_curfn *ir.Func // TODO(mdempsky): Fix.
						// TODO: why is e_curfn used instead of l.curfn?
						logopt.LogOpt(l.n.Pos(), "escape", "escape", ir.FuncName(e_curfn), fmt.Sprintf("%v escapes to heap", l.n), explanation)
					}
				}
				l.escapes = true
				enqueue(l)
				continue
			}
		}

		// Also check if l flows to the pseudo location for interface receivers
		// and has not already been marked as escaping.
		if root.ifaceRecvEscape && !l.escapes && !l.ifaceRecvEscape && go122UseIfaceRecvEscapeAnalysis {
			// TODO: we probably also could check for !root.escapes, which might be
			// an optimization but not a correctness change? If root.escapes is true,
			// this path will be handled as an escaping/outliving path and ifaceRecvEscape
			// is less interesting.

			if l.isName(ir.PPARAM) {
				// For this location l, see if we can prove it does not reach
				// the heap due to its use as an interface receiver.
				reaches := ifaceRecvPath(b, l)

				var leakToLoc *location
				switch reaches {
				case heapYes:
					leakToLoc = &b.heapLoc // l leaks to heap.
				case heapMaybe:
					leakToLoc = &b.ifaceRecvLoc // l leaks to pseudo location for interface receivers.
				case heapNo:
					if base.Debug.EscRecvDebug > 0 {
						// TODO: maybe make this similar to self-assign log? can use base.FmtPos(l.n.Pos()).
						fmt.Printf("walkOne: parameter %v does not escape via interface receiver\n", l.n)
					}
					leakToLoc = nil // l doesn't leak anywhere due to its use as an interface receiver.
				}
				if leakToLoc != nil {
					if (logopt.Enabled() || base.Flag.LowerM >= 2) && !l.escapes {
						if base.Flag.LowerM >= 2 {
							fmt.Printf("%s: parameter %v leaks to %s in %v with derefs=%d:\n", base.FmtPos(l.n.Pos()), l.n, b.explainLoc(leakToLoc), ir.FuncName(l.curfn), derefs)
						}
						explanation := b.explainPath(root, l)
						if logopt.Enabled() {
							var e_curfn *ir.Func // TODO(mdempsky): Fix.
							// TODO: why is e_curfn used instead of l.curfn?
							logopt.LogOpt(l.n.Pos(), "leak", "escape", ir.FuncName(e_curfn),
								fmt.Sprintf("parameter %v leaks to %s with derefs=%d", l.n, b.explainLoc(leakToLoc), derefs), explanation)
						}
					}
					l.leakTo(leakToLoc, derefs)
				}
			}

			if addressOf {
				// For this location l, see if we can prove it does not reach
				// the heap due to its use as an interface receiver.
				reaches := ifaceRecvPath(b, l)

				switch reaches {
				case heapYes:
					if logopt.Enabled() || base.Flag.LowerM >= 2 {
						if base.Flag.LowerM >= 2 {
							fmt.Printf("%s: %v escapes to heap in %v:\n", base.FmtPos(l.n.Pos()), l.n, ir.FuncName(l.curfn))
						}
						explanation := b.explainPath(root, l)
						if logopt.Enabled() {
							var e_curfn *ir.Func // TODO(mdempsky): Fix.
							// TODO: why is e_curfn used instead of l.curfn?
							logopt.LogOpt(l.n.Pos(), "escape", "escape", ir.FuncName(e_curfn), fmt.Sprintf("%v escapes to heap", l.n), explanation)
						}
					}
					l.escapes = true
					enqueue(l) // TODO: do we have a test that fails without this? Is it strictly needed for correctness?
					continue
				case heapMaybe:
					if logopt.Enabled() || base.Flag.LowerM >= 2 {
						if base.Flag.LowerM >= 2 {
							fmt.Printf("%s: %v might escape to heap in %v:\n", base.FmtPos(l.n.Pos()), l.n, ir.FuncName(l.curfn))
						}
						explanation := b.explainPath(root, l)
						if logopt.Enabled() {
							var e_curfn *ir.Func // TODO(mdempsky): Fix.
							// TODO: why is e_curfn used instead of src.l?
							logopt.LogOpt(l.n.Pos(), "escape", "escape", ir.FuncName(e_curfn), fmt.Sprintf("%v might escape to heap", l.n), explanation)
						}
					}
					// TODO: look at the testing/fstest.fsOnly example in escape_iface_recv_extracted.go,
					// which changes behavior based on removing these two lines.
					l.ifaceRecvEscape = true
					enqueue(l)
					continue
				case heapNo:
					if base.Debug.EscRecvDebug > 0 {
						// TODO: maybe make this similar to self-assign log?
						fmt.Printf("walkOne: %v does not escape via interface receiver (with addressOf: true)\n", l.n)
					}
				}
			}
		}

		for i, edge := range l.edges {
			if edge.src.escapes {
				continue
			}
			d := derefs + edge.derefs
			if edge.src.walkgen != walkgen || edge.src.derefs > d {
				edge.src.walkgen = walkgen
				edge.src.derefs = d
				edge.src.dst = l
				edge.src.dstEdgeIdx = i
				todo = append(todo, edge.src)
			}
		}
	}
}

// explainPath prints an explanation of how src flows to the walk root.
func (b *batch) explainPath(root, src *location) []*logopt.LoggedOpt {
	visited := make(map[*location]bool)
	pos := base.FmtPos(src.n.Pos())
	var explanation []*logopt.LoggedOpt
	for {
		// Prevent infinite loop.
		if visited[src] {
			if base.Flag.LowerM >= 2 {
				fmt.Printf("%s:   warning: truncated explanation due to assignment cycle; see golang.org/issue/35518\n", pos)
			}
			break
		}
		visited[src] = true
		dst := src.dst
		edge := &dst.edges[src.dstEdgeIdx]
		if edge.src != src {
			base.Fatalf("path inconsistency: %v != %v", edge.src, src)
		}

		explanation = b.explainFlow(pos, dst, src, edge.derefs, edge.notes, explanation)

		if dst == root {
			break
		}
		src = dst
	}

	return explanation
}

func (b *batch) explainFlow(pos string, dst, srcloc *location, derefs int, notes *note, explanation []*logopt.LoggedOpt) []*logopt.LoggedOpt {
	ops := "&"
	if derefs >= 0 {
		ops = strings.Repeat("*", derefs)
	}
	print := base.Flag.LowerM >= 2

	flow := fmt.Sprintf("   flow: %s ← %s%v:", b.explainLoc(dst), ops, b.explainLoc(srcloc))
	if print {
		fmt.Printf("%s:%s\n", pos, flow)
	}
	if logopt.Enabled() {
		var epos src.XPos
		if notes != nil {
			epos = notes.where.Pos()
		} else if srcloc != nil && srcloc.n != nil {
			epos = srcloc.n.Pos()
		}
		var e_curfn *ir.Func // TODO(mdempsky): Fix.
		explanation = append(explanation, logopt.NewLoggedOpt(epos, epos, "escflow", "escape", ir.FuncName(e_curfn), flow))
	}

	for note := notes; note != nil; note = note.next {
		if print {
			fmt.Printf("%s:     from %v (%v) at %s\n", pos, note.where, note.why, base.FmtPos(note.where.Pos()))
		}
		if logopt.Enabled() {
			var e_curfn *ir.Func // TODO(mdempsky): Fix.
			notePos := note.where.Pos()
			explanation = append(explanation, logopt.NewLoggedOpt(notePos, notePos, "escflow", "escape", ir.FuncName(e_curfn),
				fmt.Sprintf("     from %v (%v)", note.where, note.why)))
		}
	}
	return explanation
}

func (b *batch) explainLoc(l *location) string {
	if l == &b.heapLoc {
		return "<heap>"
	}
	if l == &b.ifaceRecvLoc {
		return "<interface receiver>"
	}
	if l.n == nil {
		// TODO(mdempsky): Omit entirely.
		return "<temp>"
	}
	if l.n.Op() == ir.ONAME {
		return fmt.Sprintf("%v", l.n)
	}
	return fmt.Sprintf("<storage for %v>", l.n)
}

// outlives reports whether values stored in l may survive beyond
// other's lifetime if stack allocated.
// TODO: slightly clarify this comment (e.g., may, if => result)
func (b *batch) outlives(l, other *location) bool {
	// The heap outlives everything.
	if l.escapes {
		return true
	}

	// We don't know what callers do with returned values, so
	// pessimistically we need to assume they flow to the heap and
	// outlive everything too.
	if l.isName(ir.PPARAMOUT) {
		// Exception: Directly called closures can return
		// locations allocated outside of them without forcing
		// them to the heap. For example:
		//
		//    var u int  // okay to stack allocate
		//    *(func() *int { return &u }()) = 42
		if containsClosure(other.curfn, l.curfn) && l.curfn != nil && l.curfn.ClosureCalled() {
			return false
		}

		return true
	}

	// If l and other are within the same function, then l
	// outlives other if it was declared outside other's loop
	// scope. For example:
	//
	//    var l *int
	//    for {
	//        l = new(int)
	//    }
	if l.curfn == other.curfn && l.loopDepth < other.loopDepth {
		return true
	}

	// If other is declared within a child closure of where l is
	// declared, then l outlives it. For example:
	//
	//    var l *int
	//    func() {
	//        l = new(int)
	//    }
	if containsClosure(l.curfn, other.curfn) {
		return true
	}

	return false
}

// containsClosure reports whether c is a closure contained within f.
func containsClosure(f, c *ir.Func) bool {
	// Common case.
	if f == c {
		return false
	}

	if f == nil || c == nil {
		// c is not a closure contained within f.
		// TODO: this can happen when checking ifaceRecvLoc. Better place to check?
		return false
	}

	// Closures within function Foo are named like "Foo.funcN..."
	// TODO(mdempsky): Better way to recognize this.
	fn := f.Sym().Name
	cn := c.Sym().Name
	return len(cn) > len(fn) && cn[:len(fn)] == fn && cn[len(fn)] == '.'
}

type reachesHeap int

const (
	heapUnknown reachesHeap = iota
	// Ordered from best to worst outcome (heapNo to heapYes)
	heapNo
	heapMaybe
	heapYes
)

func (i reachesHeap) String() string {
	// TODO: this uses our first cut terminology.
	switch i {
	case heapYes:
		return "escapes"
	case heapMaybe:
		return "mightEscape"
	case heapNo:
		return "noEscapes"
	default:
		return "unknown"
	}
}

// ifaceRecvCanEscape is a cache for the ifaceRecvPath function.
// TODO: we could also cache on location or node or type or similar to avoid repeated map lookup.
var ifaceRecvCanEscape = map[*types.Type]reachesHeap{}

// ifaceRecvPath accepts a location l that has a path to the interface receiver
// pseudo location in our data-flow graph and reports if the location must
// be treated as escaping to the heap.
//
// To do this, ifaceRecvPath examines the type associated with the location
// to see if it can prove that use as an interface receiver parameter
// in an interface method call cannot result in escaping to the heap,
// and if so, it reports reachesHeap as heapNo.
//
// For example, a location whose type does not have any methods might still
// have a path to the interface receiver pseudo location in the data-flow graph,
// but a value of that type cannot actually be used in an interface method call,
// and hence can never escape due to use as an interface receiver. (Our data-flow
// graph is a simplified model that ignores conditionals, does not generally
// distinguish between elements within a compound variable, and so on).
//
// If ifaceRecvPath encounters an interface in l's type but would have otherwise
// reported reachesHeap as heapNo, it reports heapMaybe to propagate the uncertainty
// about what types might be in the interface at run time. (If our data-flow graph
// later shows a type being assigned to the interface, the uncertainty can be
// resolved when examining that type).
//
// Otherwise, ifaceRecvPath reports reachesHeap as heapYes.
//
// ifaceRecvPath can be and is conservative, including it does not attempt to
// understand all types or all node ops. It might for example conservatively
// report heapYes when the true answer might be heapNo, but it should never
// do the reverse.
func ifaceRecvPath(b *batch, l *location) reachesHeap {
	if !go122UseIfaceRecvEscapeAnalysis {
		base.Fatalf("escape analysis: interface receiver analysis disabled but ifaceRecvPath was called")
	}
	if l.n == nil {
		// TODO: when does a nil node occur for a location?
		//  - 'teeHole' and 'later' holes have nil nodes.
		//  - ir.ORANGE in stmt.go creates a temp because the X in range X is eval'ed outside body.
		//  - others?
		//  - explainLoc reports nil node as '<temp>'.
		if base.Debug.EscRecvDebug >= 2 {
			fmt.Printf("ifaceRecvPath: encountered nil node in %s, reporting as escaping\n", ir.FuncName(l.curfn))
		}
		// TODO: is it reasonable to return heapNo? For now, we conservatively return heapYes.
		return heapYes
	}

	node := l.n
	switch node.Op() {
	case ir.OCONVIFACE:
		node = l.n.(*ir.ConvExpr).X
	case ir.ONAME, ir.OPTRLIT, ir.OSLICELIT, ir.ONEW:
		// Already the type we want to analyze.
	default:
		// TODO: support others like OMAKESLICE, maybe OADDSTR, OMAKEMAP, ...
		if base.Debug.EscRecvDebug > 0 {
			fmt.Printf("ifaceRecvPath: unhandled op %v, treating node as escaping in %v at %v\n",
				node.Op(), ir.FuncName(l.curfn), base.FmtPos(node.Pos()))
		}
		return heapYes
	}

	nodeType := node.Type()
	if reaches, ok := ifaceRecvCanEscape[nodeType]; ok {
		if base.Debug.EscRecvDebug > 0 {
			// TODO: maybe report at different debug levels based on result?
			fmt.Printf("ifaceRecvPath: using cached result of %v for type %v in node %v in %v at %v\n",
				reaches, nodeType, node, ir.FuncName(l.curfn), base.FmtPos(node.Pos()))
		}
		return reaches
	}

	type checkOne func(t *types.Type) reachesHeap
	var visit func(t *types.Type, check checkOne) reachesHeap
	visit = func(t *types.Type, check checkOne) reachesHeap {
		// Find the worst outcome (highest reachesHeap) for type t recursively,
		// stopping early if we find the worst possible (heapYes).
		// We purposefully handle a limited set of types.
		max := func(a, b reachesHeap) reachesHeap {
			if a > b {
				return a
			}
			return b
		}

		// We don't attempt to handle recursive types.
		if t.Recur() {
			// TODO: is this the right way to check for recursive types? Is it safe to set this flag temporarily?
			// TODO: we could gracefully handle recursive types.
			if base.Debug.EscRecvDebug >= 2 {
				fmt.Printf("ifaceRecvPath: encountered recursive type %v in node %v in %v at %v, reporting as escaping\n",
					t, node, ir.FuncName(l.curfn), base.FmtPos(node.Pos()))
			}
			return heapYes
		}
		t.SetRecur(true)
		defer t.SetRecur(false)

		switch {
		default:
			// We don't understand all types, and the conservative answer is
			// we reach the heap if we cannot prove otherwise.
			// TODO: add short/simple tests for more types (bool, unsafe.Pointer, ...)
			// TODO: would be nice if we could handle kind FUNC, including reflect.Value has an Equal func in its internal/abi.Type.
			// TODO: if needed, we could consider special case knowledge of reflect.Value or a narrow set of other reflect constructs.
			if base.Debug.EscRecvDebug > 0 {
				fmt.Printf("ifaceRecvPath: encountered unsupported kind %v for type %v in node %v in %v at %v, reporting as escaping\n",
					t.Kind(), t, node, ir.FuncName(l.curfn), base.FmtPos(node.Pos()))
			}
			return heapYes
		case t.IsInterface():
			// We don't know what concrete type will ultimately be inside an interface,
			// so we cannot analyze it now, reporting instead it might reach the heap.
			if base.Debug.EscRecvDebug >= 2 {
				fmt.Printf("ifaceRecvPath: encountered interface %v in node %v in %v at %v, reporting as might escape\n",
					t, node, ir.FuncName(l.curfn), base.FmtPos(node.Pos()))
			}
			return heapMaybe
		case t.IsScalar():
			// Numeric or bool.
			// TODO: alternative might be types.IsSimple[t.Kind()], which also includes types.TIDEAL?
			return check(t)
		case t.IsString():
			return check(t)
		case t.IsStruct():
			reaches := check(t)
			if reaches == heapYes {
				return heapYes
			}
			for _, field := range t.Fields().Slice() {
				childReaches := visit(field.Type, check)
				reaches = max(childReaches, reaches)
				if reaches == heapYes {
					return heapYes
				}
			}
			return reaches
		case t.IsPtr() || t.IsSlice():
			reaches := check(t)
			if reaches == heapYes {
				return heapYes
			}
			childReaches := visit(t.Elem(), check)
			return max(childReaches, reaches)
		}
	}

	// TODO: for now, we do three recursive walks, which we could collapse down.

	// First, walk to see if we have any methods anywhere,
	// returning heapUnknown in all cases (because we are not yet saying anything conclusive).
	// (Recall visit ensures we never see an interface type passed to our func literal here,
	// and visit returns at least heapMaybe if it does encounter an interface type).
	var haveMethods bool
	reaches := visit(nodeType, func(t *types.Type) reachesHeap {
		if t.AllMethods().Len() > 0 {
			haveMethods = true
			return heapUnknown
		}
		return heapUnknown
	})
	if reaches == heapYes {
		// Visit encountered a kind of type it did not understand or otherwise concluded that t
		// leaks to the heap when used as an interface receiver in an interface call.
		ifaceRecvCanEscape[nodeType] = reaches // cache result
		return heapYes
	}
	if reaches == heapUnknown && !haveMethods {
		// No interfaces encountered, and no methods anywhere, so the type cannot be leaked
		// due to use as an interface receiver in an interface call.
		if base.Debug.EscRecvDebug >= 2 {
			fmt.Printf("ifaceRecvPath: receiver cannot leak because its type %v has no methods in node %v in %v at %v\n",
				nodeType, node, ir.FuncName(l.curfn), base.FmtPos(node.Pos()))
		}
		ifaceRecvCanEscape[nodeType] = reaches // cache result
		return heapNo
	}

	// Second recursive walk to check if t contains any pointers anywhere,
	// including types that use pointers in their impl (like maps, interfaces, and so on).
	nodeHasPointers := nodeType.HasPointers()

	// Third, final, and most complex recursive walk: here, we check method-by-method if we can prove
	// for each type visited recursively:
	//   a. structurally the receiver in the method cannot leak due to use as an interface receiver
	//      because the type has no pointers anywhere and the receiver is not a pointer, or
	//   b. escape analysis is complete on the method, and it says the receiver does not leak.
	// This is currently conservative. (Also, when we say here that we prove the receiver cannot
	// leak due to use as an interface receiver, the value might leak for other reasons such as
	// if the value used as the receiver is also used as a parameter or otherwise escapes.
	// Our job here is just what happens due to use as interface receiver, and the rest of
	// escape analysis can prove the value escapes for other reasons).
	reaches = visit(nodeType, func(t *types.Type) reachesHeap {
		// For each type t visited, loop over each method, skipping methods that cannot cause this type
		// to leak when used as an interface receiver.
		// If we get to the end of the loop, we report reachesHeap as heapNo.
		for _, m := range t.AllMethods().Slice() {
			recv := m.Type.Recv()
			if recv == nil {
				if base.Debug.EscRecvDebug >= 2 {
					fmt.Printf("ifaceRecvPath: allowing method with nil m.Type.Recv: %v.%#v\n", t, m.Sym)
				}
				continue
			}
			if recv.Sym == nil || recv.Sym.IsBlank() {
				// Unnamed parameters are unused and therefore do not escape.
				if base.Debug.EscRecvDebug >= 2 {
					fmt.Printf("ifaceRecvPath: allowing method with unnamed or blank m.Type.Recv: %v.%#v\n", t, m.Sym)
				}
				continue
			}

			if !nodeHasPointers && !recv.Type.IsPtr() {
				// The nodeType does not have any pointers, and the receiver is not a pointer,
				// so neither the receiver nor its fields cannot be leaked just by using the receiver in an interface method call.
				if base.Debug.EscRecvDebug >= 2 {
					fmt.Printf("ifaceRecvPath: allowing method without pointer receiver given type does not have pointers for type %v for m.Type.Recv: %v.%#v\n", recv.Type, t, m.Sym)
				}
				continue
			}

			// Next, examine escape analysis results if we can.
			// TODO: currently, we allow use of escape analysis results for types in the same package,
			// but we probably should not. If the method of interest here on this type has not been called directly or
			// indirectly by the function we are analyzing, the method of interest here can have completed escape
			// analysis (or not) based on ordering of files or ordering of content within files, which is probably too
			// surprising & too sensitive to otherwise innocent changes. (Ordinary escape analysis visits functions
			// from the bottom of the call graph upward, but that does not take into account the implementation
			// behind an interface call, which can vary at run time).

			if !recv.Type.HasPointers() {
				// Escape analysis does not waste space recording a Note for function parameter types without any pointers
				// (because they cannot be leaked), so continue past this method.
				// If we were to check the recv.Note below, it would be empty, which is the encoding for escapes to heap.
				// TODO: HasPointers is recursive, so this is wasteful calling it repeatedly.
				if base.Debug.EscRecvDebug >= 2 {
					fmt.Printf("ifaceRecvPath: allowing method with receiver type %v that contains no pointers for m.Type.Recv: %v.%#v of type\n",
						recv.Type, t, m.Sym)
				}
				continue
			}

			// Nname points to the function name node as a types.Obj,
			// which represents an ir.Node.
			mNname := m.Nname
			if mNname == nil {
				// This can happen for example with the Read method in io.nopCloserWriterTo, which is defined via:
				//   type nopCloserWriterTo struct { Reader }
				if base.Debug.EscRecvDebug >= 2 {
					fmt.Printf("ifaceRecvPath: allowing method with nil Nname: %v.%#v\n", t, m.Sym)
				}
				continue
			}
			fn := mNname.(*ir.Name)

			// TODO: are these the right checks?
			isImported := fn.Sym().Pkg != types.LocalPkg
			isTagged := fn.Defn != nil && fn.Defn.Esc() == escFuncTagged
			canCheckEsc := isImported || isTagged
			if base.Debug.EscRecv < 2 {
				if base.Debug.EscRecvDebug > 0 {
					// TODO: change to "cannot check recorded paramater tags" or similar. change "recorded results:" below to "param tags:" or similar.
					fmt.Printf("ifaceRecvPath: cannot check recorded escape results for %v.%#v because not enabled. isTagged: %v, isImported: %v\n",
						t, m.Sym, isTagged, isImported)
				}
				return heapYes
			}
			if !canCheckEsc {
				if base.Debug.EscRecvDebug > 0 {
					fmt.Printf("ifaceRecvPath: cannot check recorded escape results for %v.%#v. isTagged: %v, isImported: %v\n",
						t, m.Sym, isTagged, isImported)
				}
				return heapYes
			}

			// Inspect the escape analysis results.
			esc := parseLeaks(recv.Note)
			if esc.Heap() >= 0 {
				if base.Debug.EscRecvDebug > 0 {
					fmt.Printf("ifaceRecvPath: recorded results: leak to heap of receiver for %v.%#v\n", t, m.Sym)
				}
				return heapYes
			}
			if m.Type.Results().NumFields() > numEscResults {
				// TODO: is something like this needed? Or would it already be marked as leak to heap if too many return params?
				if base.Debug.EscRecvDebug > 0 {
					fmt.Printf("ifaceRecvPath: recorded results: too many return results to track possible receiver leak for %v.%#v\n", t, m.Sym)
				}
				return heapYes
			}
			for i := 0; i < numEscResults; i++ {
				if esc.Result(i) >= 0 {
					if base.Debug.EscRecvDebug > 0 {
						fmt.Printf("ifaceRecvPath: recorded results: receiver leaks to return result for %v.%#v\n", t, m.Sym)
					}
					// TODO: this is likely conservative -- it might be OK based on call site.
					return heapYes
				}
			}
			if esc.IfaceRecv() >= 0 {
				// TODO: confirm we have explicit test e.g., struct Foo has Bar method with receiver converterd to Stringer)
				if base.Debug.EscRecvDebug > 0 {
					fmt.Printf("ifaceRecvPath: recorded results: receiver marked as esc.IfaceRecv for %v.%#v\n", t, m.Sym)
				}
				return heapMaybe
			}

			// We've concluded this recv does not leak.
			if base.Debug.EscRecvDebug >= 2 {
				fmt.Printf("ifaceRecvPath: recorded results: receiver does not leak based on recorded escape analysis results for %v.%#v. isTagged: %v, isImported: %v\n",
					t, m.Sym, isTagged, isImported)
			}
			continue
		}

		// We've made it through all the methods without returning heapYes or heapMaybe.
		return heapNo
	})

	ifaceRecvCanEscape[nodeType] = reaches // cache result
	return reaches
}

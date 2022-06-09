// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

func csedown(f *Func) {
	// csedown follow a different definition than cse of equivalent since it
	// create phis for unmatched arguments:
	// equivalent(v, w):
	//   v.op == w.op
	//   v.type == w.type
	//   v.aux == w.aux
	//   v.auxint == w.auxint
	//   len(v.Args) == len(w.Args)
	//   v.uses == 1
	//   w.uses == 1
	//   not v.op == OpPhi

	var rewrites int64

	for _, b := range f.Blocks {
		if len(b.Preds) <= 1 {
			continue
		}
	ValueLoop:
		for _, v := range b.Values {
			if v.Op != OpPhi {
				continue
			}

			// TODO: handle cases where not all branches are handelable
			// TODO: handle select if all users of pred.Args[0] are handelable
			phiArgs := v.Args
			pred := phiArgs[0]
			if pred.Uses != 1 {
				continue
			}
			switch pred.Op {
			case OpPhi, OpSelect0, OpSelect1, OpSelectN:
				continue
			}
			for _, pred2 := range phiArgs[1:] {
				if pred2.Uses != 1 ||
					pred.Op != pred2.Op ||
					pred.Type != pred2.Type ||
					pred.Aux != pred2.Aux ||
					pred.AuxInt != pred2.AuxInt ||
					len(pred.Args) != len(pred2.Args) {
					continue ValueLoop
				}
			}

			// Match let's replace
			// collect preds's arguments
			args := make([][]*Value, len(pred.Args))
			for i := range args {
				args[i] = make([]*Value, len(phiArgs))
			}
			for i, a := range phiArgs {
				// Cannonicalise commutative ops with lower id in Args[0]
				if opcodeTable[pred.Op].commutative && a.Args[1].ID < a.Args[0].ID {
					a.Args[0], a.Args[1] = a.Args[1], a.Args[0]
				}
				for j, w := range a.Args {
					if v == w {
						// Self dependent
						continue ValueLoop
					}
					// TODO: support memory
					if w.Type.IsFlags() ||
						(w.Type.IsStruct() && !isDecomposableValue(w)) ||
						w.Type.IsMemory() {
						continue ValueLoop
					}
					args[j][i] = w
				}
			}

			phiArgs = append([]*Value(nil), phiArgs...)

			// Make the CSEd value
			v.reset(pred.Op)
			v.Type = pred.Type
			v.Aux = pred.Aux
			v.AuxInt = pred.AuxInt
			rewrites++

			// Add arguments
		ArgLoop:
			for _, a := range args {
				c := a[0]
				// First just reuse the value if all preds use the same one
				for _, v := range a[1:] {
					if c != v {
						goto SearchPhi
					}
				}
				v.AddArg(c)
				continue
				// Secondly try to find a phi already doing what we need
			SearchPhi:
				for _, w := range b.Values {
					if v == w || w.Op != OpPhi || len(w.Args) != len(a) {
						continue
					}

					for i := range a {
						if a[i] != w.Args[i] {
							continue SearchPhi
						}
					}
					// We have found a matching phi
					v.AddArg(w)
					continue ArgLoop
				}

				// Third phi exists yet, let's make one
				w := b.NewValue0(c.Pos, OpPhi, c.Type)
				w.Args = a
				for _, v := range a {
					v.Uses++
				}
				v.AddArg(w)
			}

			// Reset all previous values so next iteration counters are correct
			for _, w := range phiArgs {
				w.reset(OpInvalid)
			}
			goto ValueLoop
		}
	}

	if f.pass.stats > 0 {
		f.LogStat("CSEDOWN REWRITES", rewrites)
	}
}

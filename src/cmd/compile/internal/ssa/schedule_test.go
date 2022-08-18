// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

import (
	"cmd/compile/internal/types"
	"testing"
)

func TestSchedule(t *testing.T) {
	c := testConfig(t)
	cases := []fun{
		c.Fun("entry",
			Bloc("entry",
				Valu("mem0", OpInitMem, types.TypeMem, 0, nil),
				Valu("ptr", OpConst64, c.config.Types.Int64, 0xABCD, nil),
				Valu("v", OpConst64, c.config.Types.Int64, 12, nil),
				Valu("mem1", OpStore, types.TypeMem, 0, c.config.Types.Int64, "ptr", "v", "mem0"),
				Valu("mem2", OpStore, types.TypeMem, 0, c.config.Types.Int64, "ptr", "v", "mem1"),
				Valu("mem3", OpStore, types.TypeMem, 0, c.config.Types.Int64, "ptr", "sum", "mem2"),
				Valu("l1", OpLoad, c.config.Types.Int64, 0, nil, "ptr", "mem1"),
				Valu("l2", OpLoad, c.config.Types.Int64, 0, nil, "ptr", "mem2"),
				Valu("sum", OpAdd64, c.config.Types.Int64, 0, nil, "l1", "l2"),
				Goto("exit")),
			Bloc("exit",
				Exit("mem3"))),
	}
	for _, c := range cases {
		schedule(c.f)
		if !isSingleLiveMem(c.f) {
			t.Error("single-live-mem restriction not enforced by schedule for func:")
			printFunc(c.f)
		}
	}
}

func isSingleLiveMem(f *Func) bool {
	for _, b := range f.Blocks {
		var liveMem *Value
		for _, v := range b.Values {
			for _, w := range v.Args {
				if w.Type.IsMemory() {
					if liveMem == nil {
						liveMem = w
						continue
					}
					if w != liveMem {
						return false
					}
				}
			}
			if v.Type.IsMemory() {
				liveMem = v
			}
		}
	}
	return true
}

func TestStoreOrder(t *testing.T) {
	// In the function below, v2 depends on v3 and v4, v4 depends on v3, and v3 depends on store v5.
	// storeOrder did not handle this case correctly.
	c := testConfig(t)
	fun := c.Fun("entry",
		Bloc("entry",
			Valu("mem0", OpInitMem, types.TypeMem, 0, nil),
			Valu("a", OpAdd64, c.config.Types.Int64, 0, nil, "b", "c"),                        // v2
			Valu("b", OpLoad, c.config.Types.Int64, 0, nil, "ptr", "mem1"),                    // v3
			Valu("c", OpNeg64, c.config.Types.Int64, 0, nil, "b"),                             // v4
			Valu("mem1", OpStore, types.TypeMem, 0, c.config.Types.Int64, "ptr", "v", "mem0"), // v5
			Valu("mem2", OpStore, types.TypeMem, 0, c.config.Types.Int64, "ptr", "a", "mem1"),
			Valu("ptr", OpConst64, c.config.Types.Int64, 0xABCD, nil),
			Valu("v", OpConst64, c.config.Types.Int64, 12, nil),
			Goto("exit")),
		Bloc("exit",
			Exit("mem2")))

	CheckFunc(fun.f)
	order := storeOrder(fun.f.Blocks[0].Values, fun.f.newSparseSet(fun.f.NumValues()), make([]int32, fun.f.NumValues()))

	// check that v2, v3, v4 is sorted after v5
	var ai, bi, ci, si int
	for i, v := range order {
		switch v.ID {
		case 2:
			ai = i
		case 3:
			bi = i
		case 4:
			ci = i
		case 5:
			si = i
		}
	}
	if ai < si || bi < si || ci < si {
		t.Logf("Func: %s", fun.f)
		t.Errorf("store order is wrong: got %v, want v2 v3 v4 after v5", order)
	}
}

func TestCarryChainOrder(t *testing.T) {
	// In the function below, v10 depends on v6, v6 depends on v5, and v2 depends on v8,
	// v8 depends on v7. But there is no dependency between the two carry chains. If they
	// are not scheduled properly, the carry value will be clobbered.
	c := testConfigARM64(t)
	fun := c.Fun("entry",
		Bloc("entry",
			Valu("v1", OpInitMem, types.TypeMem, 0, nil),
			Valu("x", OpARM64MOVDconst, c.config.Types.UInt64, 5, nil),
			Valu("y", OpARM64MOVDconst, c.config.Types.UInt64, 6, nil),
			Valu("z", OpARM64MOVDconst, c.config.Types.UInt64, 7, nil),
			Valu("v5", OpARM64ADDSflags, types.NewTuple(c.config.Types.UInt64, types.TypeFlags), 0, nil, "x", "z"), // x+z, set flags
			Valu("b", OpSelect1, types.TypeFlags, 0, nil, "v5"),
			Valu("v7", OpARM64ADDSflags, types.NewTuple(c.config.Types.UInt64, types.TypeFlags), 0, nil, "y", "z"), // y+z, set flags
			Valu("d", OpSelect1, types.TypeFlags, 0, nil, "v7"),
			Valu("a", OpSelect0, c.config.Types.UInt64, 0, nil, "v5"),
			Valu("v10", OpARM64ADCzerocarry, c.config.Types.UInt64, 0, nil, "b"), // 0+0+carry
			Valu("c", OpSelect0, c.config.Types.UInt64, 0, nil, "v7"),
			Valu("v12", OpARM64ADCzerocarry, c.config.Types.UInt64, 0, nil, "d"), // 0+0+carry
			Valu("v13", OpARM64ADD, c.config.Types.UInt64, 0, nil, "a", "c"),
			Valu("v14", OpARM64ADD, c.config.Types.UInt64, 0, nil, "v10", "v12"),
			Valu("v15", OpARM64AND, c.config.Types.UInt64, 0, nil, "v13", "v14"),
			Goto("exit")),
		Bloc("exit",
			Exit("v1")),
	)

	CheckFunc(fun.f)
	schedule(fun.f)

	// the expected order is v5, a, b, v10, v7, c, d, v12, check that b < v10 < v7 < v8 < v12.
	var ai, bi, ci, di, ei int
	for i, v := range fun.f.Blocks[0].Values {
		switch v.ID {
		case 6:
			ai = i
		case 10:
			bi = i
		case 7:
			ci = i
		case 8:
			di = i
		case 12:
			ei = i
		}
	}
	if !(ai < bi && bi < ci && ci < di && di < ei) {
		t.Logf("Func: %s", fun.f)
		t.Errorf("carry chain order is wrong: got %v, want v12 after v8 after v7 after v10 after v6,", fun.f.Blocks[0])
	}
}

package ssa

import (
	"fmt"
	"testing"
)

func TestCse(t *testing.T) {
	c := NewConfig("amd64", DummyFrontend{t})
	fun := Fun(c, "entry",
		Bloc("entry",
			Valu("mem", OpArg, TypeMem, 0, ".mem"),
			Valu("var1", OpConst, TypeInt64, 0, 25),
			Valu("var2", OpConst, TypeInt64, 0, 25),
			Valu("var3", OpConst, TypeInt64, 0, 25),
			Goto("b2")),
		Bloc("b2",
			Valu("v5", OpCopy, TypeInt64, 0, nil, "var1"),
			Valu("v6", OpCopy, TypeInt64, 0, nil, "var2"),
			Valu("v7", OpCopy, TypeInt64, 0, nil, "var3"),
			Goto("exit")),
		Bloc("exit",
			Exit("mem")))

	CheckFunc(fun.f)
	cse(fun.f)
	CheckFunc(fun.f)
	for _, b := range fun.f.Blocks {
		if b == fun.blocks["b2"] {
			arg := b.Values[0].Args[0]
			for _, v := range b.Values {
				if len(v.Args) != 1 {
					panic(fmt.Sprintf("expected arg length = 1, found %d", len(v.Args)))
				}

				if v.Args[0] != arg {
					panic(fmt.Sprintf("expected arg  = %v, found %v", arg, v.Args[0]))
				}
			}
		}
	}
}

func TestCsePhiArgs(t *testing.T) {
	c := NewConfig("amd64", DummyFrontend{t})
	fun := Fun(c, "entry",
		Bloc("entry",
			Valu("mem", OpArg, TypeMem, 0, ".mem"),
			Valu("p", OpConst, TypeBool, 0, true),
			Valu("var1", OpConst, TypeInt64, 0, 0),
			Valu("var2", OpConst, TypeInt64, 0, 0),
			Valu("var3", OpConst, TypeInt64, 0, 1),
			Valu("var4", OpConst, TypeInt64, 0, 2),
			If("p", "b2", "b3")),
		Bloc("b2",
			Goto("b3")),
		Bloc("b3",
			// var1 and var2 shouldn't be replaced by the same value due to being
			// arguments of different phis
			Valu("v5", OpPhi, TypeInt64, 0, nil, "var1", "var3"),
			Valu("v6", OpPhi, TypeInt64, 0, nil, "var2", "var4"),
			Goto("exit")),
		Bloc("exit",
			Exit("mem")))

	CheckFunc(fun.f)
	cse(fun.f)
	CheckFunc(fun.f)
	for _, b := range fun.f.Blocks {
		if b == fun.blocks["b3"] {
			phiArgs := make([]bool, fun.f.NumValues())
			for _, v := range b.Values {
				if v.Op == OpPhi {
					for _, a := range v.Args {
						phiArgs[a.ID] = true
					}
				}
			}
			uniqueArgCount := 0
			for _, b := range phiArgs {
				if b {
					uniqueArgCount++
				}
			}
			if uniqueArgCount != 4 {
				fmt.Printf("expected 4 unique phi arguments, found %d", uniqueArgCount)
			}

		}
	}
}

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkginit

import (
	"strings"

	"cmd/compile/internal/base"
	"cmd/compile/internal/ir"
	"cmd/compile/internal/typecheck"
	"cmd/compile/internal/types"
	"cmd/internal/src"
)

// instrumentGlobals declares a global array of _asan_global structures and initializes it.
func instrumentGlobals(fn *ir.Func) *ir.Name {
	asanGlobalStruct, asanLocationStruct, defStringstruct := createtypes()
	lname := typecheck.Lookup
	tconv := typecheck.ConvNop
	// Make a global array of asanGlobalStruct type.
	// var asanglobals []asanGlobalStruct
	arraytype := types.NewArray(asanGlobalStruct, int64(len(ShouldInstrumentGlobals)))
	symG := lname(".asanglobals")
	globals := typecheck.NewName(symG)
	globals.SetType(arraytype)
	globals.Class = ir.PEXTERN
	symG.Def = globals
	typecheck.Target.Externs = append(typecheck.Target.Externs, globals)
	// Make a global array of asanLocationStruct type.
	// var asanL []asanLocationStruct
	arraytype = types.NewArray(asanLocationStruct, int64(len(ShouldInstrumentGlobals)))
	symL := lname(".asanL")
	asanlocation := typecheck.NewName(symL)
	asanlocation.SetType(arraytype)
	asanlocation.Class = ir.PEXTERN
	symL.Def = asanlocation
	typecheck.Target.Externs = append(typecheck.Target.Externs, asanlocation)
	// Make two global string variables to pass the name of the global and its module name.
	// var asanName string
	// var asanModulename string
	symL = lname(".asanName")
	asanName := typecheck.NewName(symL) // nn
	asanName.SetType(types.Types[types.TSTRING])
	asanName.Class = ir.PEXTERN
	symL.Def = asanName
	typecheck.Target.Externs = append(typecheck.Target.Externs, asanName)

	symL = lname(".asanModulename")
	asanModulename := typecheck.NewName(symL)
	asanModulename.SetType(types.Types[types.TSTRING])
	asanModulename.Class = ir.PEXTERN
	symL.Def = asanModulename
	typecheck.Target.Externs = append(typecheck.Target.Externs, asanModulename)

	var init ir.Nodes
	var c ir.Node
	i := int64(0)
	for symName, n := range ShouldInstrumentGlobals {
		setField := func(f string, val ir.Node, i int64) {
			r := ir.NewAssignStmt(base.Pos, ir.NewSelectorExpr(base.Pos, ir.ODOT,
				ir.NewIndexExpr(base.Pos, globals, ir.NewInt(i)), lname(f)), val)
			init.Append(typecheck.Stmt(r))
		}
		// globals[i].beg = uintptr(unsafe.Pointer(&n))
		c = tconv(typecheck.NodAddr(n), types.Types[types.TUNSAFEPTR])
		c = tconv(c, types.Types[types.TUINTPTR])
		setField("beg", c, i)
		// Assign globals[i].size.
		g := n.(*ir.Name)
		size := g.Type().Size()
		c = tconv(ir.NewInt(size), types.Types[types.TUINTPTR])
		setField("size", c, i)
		// Assign globals[i].sizeWithRedzone.
		rzSize := GetRedzoneSizeForGlobal(size)
		sizeWithRz := rzSize + size
		c = tconv(ir.NewInt(sizeWithRz), types.Types[types.TUINTPTR])
		setField("sizeWithRedzone", c, i)
		// The C string type is terminated by a null charactor "\0", Go should use three-digit
		// octal "\000" or two-digit hexadecimal "\x00" to create null terminated string.
		// asanName = symName + "\000"
		// globals[i].name = (*defString)(unsafe.Pointer(&asanName)).data
		init.Append(typecheck.Stmt(ir.NewAssignStmt(base.Pos, asanName, ir.NewString(symName+"\000"))))
		c = tconv(typecheck.NodAddr(asanName), types.Types[types.TUNSAFEPTR])
		c = tconv(c, types.NewPtr(defStringstruct))
		c = ir.NewSelectorExpr(base.Pos, ir.ODOT, c, lname("data"))
		setField("name", c, i)
		// Set the name of package being compiled as a unique identifier of a module.
		// asanModulename = pkgName + "\000"
		init.Append(typecheck.Stmt(ir.NewAssignStmt(base.Pos, asanModulename, ir.NewString(types.LocalPkg.Name+"\000"))))
		c = tconv(typecheck.NodAddr(asanModulename), types.Types[types.TUNSAFEPTR])
		c = tconv(c, types.NewPtr(defStringstruct))
		c = ir.NewSelectorExpr(base.Pos, ir.ODOT, c, lname("data"))
		setField("moduleName", c, i)
		// Assign globals[i].location = uintptr(unsafe.Pointer(&asanL[i]))
		// TODO: How do we get the file name here?
		// We can not get the file name here, so set the name of package being compiled.
		// asanL[i].filename = pkgName + "\000"
		asanLi := ir.NewIndexExpr(base.Pos, asanlocation, ir.NewInt(i))
		init.Append(typecheck.Stmt(ir.NewAssignStmt(base.Pos, ir.NewSelectorExpr(base.Pos, ir.ODOT, asanLi, lname("filename")), c)))
		line := ir.NewInt(int64(n.Pos().Line()))
		init.Append(typecheck.Stmt(ir.NewAssignStmt(base.Pos, ir.NewSelectorExpr(base.Pos, ir.ODOT, asanLi, lname("line")), line)))
		col := ir.NewInt(int64(n.Pos().Col()))
		init.Append(typecheck.Stmt(ir.NewAssignStmt(base.Pos, ir.NewSelectorExpr(base.Pos, ir.ODOT, asanLi, lname("column")), col)))
		c = tconv(typecheck.NodAddr(asanLi), types.Types[types.TUNSAFEPTR])
		c = tconv(c, types.Types[types.TUINTPTR])
		setField("sourceLocation", c, i)
		i++
	}
	fn.Body.Append(init...)
	return globals
}

// createtypes creates the asanGlobal, asanLocation and defString struct type.
// Go compiler does not refer to the C types, we represent the struct field
// by a uintptr, then use type conversion to make copies of the data.
// E.g., (*defString)(asanGlobal.name).data to C string.
//
// Keep in sync with src/runtime/asan/asan.go.
// type asanGlobal struct {
//	beg               uintptr
//	size              uintptr
//	size_with_redzone uintptr
//	name              uintptr
//	moduleName        uintptr
//	hasDynamicInit    uintptr
//	sourceLocation    uintptr
//	odrIndicator      uintptr
// }
//
// type asanLocation struct {
//	filename uintptr
//	line     int32
//	column   int32
// }
//
// defString is synthesized struct type meant to capture the underlying
// implementations of string.
// type defString struct {
//	data uintptr
//	len  uintptr
// }

func createtypes() (*types.Type, *types.Type, *types.Type) {
	up := types.Types[types.TUINTPTR]
	i32 := types.Types[types.TINT32]
	fname := typecheck.Lookup
	nxp := src.NoXPos
	nfield := types.NewField
	asanGlobal := types.NewStruct(types.NoPkg, []*types.Field{
		nfield(nxp, fname("beg"), up),
		nfield(nxp, fname("size"), up),
		nfield(nxp, fname("sizeWithRedzone"), up),
		nfield(nxp, fname("name"), up),
		nfield(nxp, fname("moduleName"), up),
		nfield(nxp, fname("hasDynamicInit"), up),
		nfield(nxp, fname("sourceLocation"), up),
		nfield(nxp, fname("odrIndicator"), up),
	})
	types.CalcSize(asanGlobal)

	asanLocation := types.NewStruct(types.NoPkg, []*types.Field{
		nfield(nxp, fname("filename"), up),
		nfield(nxp, fname("line"), i32),
		nfield(nxp, fname("column"), i32),
	})
	types.CalcSize(asanLocation)

	defString := types.NewStruct(types.NoPkg, []*types.Field{
		types.NewField(nxp, fname("data"), up),
		types.NewField(nxp, fname("len"), up),
	})
	types.CalcSize(defString)

	return asanGlobal, asanLocation, defString
}

// Calculate redzone for globals.
func GetRedzoneSizeForGlobal(size int64) int64 {
	maxRZ := int64(1 << 18)
	minRZ := int64(32)
	redZone := (size / minRZ / 4) * minRZ
	switch {
	case redZone > maxRZ:
		redZone = maxRZ
	case redZone < minRZ:
		redZone = minRZ
	}
	// Round up to multiple of minRZ.
	if size%minRZ != 0 {
		redZone += minRZ - (size % minRZ)
	}
	return redZone
}

var ShouldInstrumentGlobals = make(map[string]ir.Node)

func canInstrumentGlobal(g ir.Node) bool {
	if g.Op() != ir.ONAME {
		return false
	}
	n := g.(*ir.Name)
	if n.Class == ir.PFUNC {
		return false
	}
	if n.Sym().Pkg != types.LocalPkg {
		return false
	}
	// Do not instrument function pointers.
	if strings.HasPrefix(n.Sym().Name, "x_cgo_") || strings.HasPrefix(n.Sym().Name, "_cgo_") ||
		strings.HasPrefix(n.Sym().Name, "__cgofn_") {
		return false
	}
	// Do not instrument globals that do not declared in the package being compiled.
	if !strings.HasPrefix(n.Linksym().Name, `"".`) {
		return false
	}
	return true
}

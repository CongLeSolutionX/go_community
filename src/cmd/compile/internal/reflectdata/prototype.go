package reflectdata

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/types"
	"cmd/internal/obj"
)

func PrepareDwarfTypes() {
	prepareStubTypes()
	obj.PrepareDwarfGenerateDebugInfo(base.Ctxt, func(s string) *obj.LSym {
		return types.TypeSymLookup(s).Linksym()
	})
}

func prepareStubTypes() {
	stubPkg := &types.Pkg{}
	_type := types.NewStruct(types.NoPkg, nil)
	_type.SetSym(&types.Sym{Pkg: stubPkg, Name: "runtime._type"})
	stringStructDWARF := types.NewStruct(types.NoPkg, nil)
	stringStructDWARF.SetSym(&types.Sym{Pkg: stubPkg, Name: "runtime.stringStructDWARF"})
	hchan := types.NewStruct(types.NoPkg, nil)
	hchan.SetSym(&types.Sym{Pkg: stubPkg, Name: "runtime.hchan"})
	sudog := types.NewStruct(types.NoPkg, nil)
	sudog.SetSym(&types.Sym{Pkg: stubPkg, Name: "runtime.sudog"})
	slice := types.NewStruct(types.NoPkg, nil)
	slice.SetSym(&types.Sym{Pkg: stubPkg, Name: "runtime.slice"})
	hmap := types.NewStruct(types.NoPkg, nil)
	hmap.SetSym(&types.Sym{Pkg: stubPkg, Name: "runtime.hmap"})
	bmap := types.NewStruct(types.NoPkg, nil)
	bmap.SetSym(&types.Sym{Pkg: stubPkg, Name: "runtime.bmap"})
	waitq := types.NewStruct(types.NoPkg, nil)
	waitq.SetSym(&types.Sym{Pkg: stubPkg, Name: "runtime.waitq"})
	mutex := types.NewStruct(types.NoPkg, nil)
	mutex.SetSym(&types.Sym{Pkg: stubPkg, Name: "runtime.mutex"})
	itab := types.NewStruct(types.NoPkg, nil)
	itab.SetSym(&types.Sym{Pkg: stubPkg, Name: "runtime.itab"})
	eface := types.NewStruct(types.NoPkg, nil)
	eface.SetSym(&types.Sym{Pkg: stubPkg, Name: "runtime.eface"})
	iface := types.NewStruct(types.NoPkg, nil)
	iface.SetSym(&types.Sym{Pkg: stubPkg, Name: "runtime.iface"})

	_type.SetFields([]*types.Field{
		makefield("size", types.Types[types.TUINTPTR]),
		makefield("ptrdata", types.Types[types.TUINTPTR]),
		makefield("hash", types.Types[types.TUINT32]),
		makefield("align", types.Types[types.TUINT8]),
		makefield("fieldAlign", types.Types[types.TUINT8]),
		makefield("kind", types.Types[types.TUINT8]),
		makefield("equal", types.NewSignature(types.NoPkg, nil, nil,
			[]*types.Field{makefield("", types.Types[types.TUNSAFEPTR]), makefield("", types.Types[types.TUNSAFEPTR])}, []*types.Field{makefield("", types.Types[types.TBOOL])})),
		makefield("gcdata", types.NewPtr(types.ByteType)),
		makefield("str", types.Types[types.TINT32]),
		makefield("ptrToThis", types.Types[types.TINT32]),
	})
	obj.StubTypes["runtime._type"] = DwarfType{_type}

	stringStructDWARF.SetFields([]*types.Field{
		makefield("str", types.NewPtr(types.ByteType)),
		makefield("len", types.Types[types.TINT]),
	})
	obj.StubTypes["runtime.stringStructDWARF"] = DwarfType{stringStructDWARF}

	slice.SetFields([]*types.Field{
		makefield("array", types.Types[types.TUNSAFEPTR]),
		makefield("len", types.Types[types.TINT]),
		makefield("cap", types.Types[types.TINT]),
	})
	obj.StubTypes["runtime.slice"] = DwarfType{slice}

	hmap.SetFields([]*types.Field{
		makefield("count", types.Types[types.TINT]),
		makefield("flags", types.Types[types.TUINT8]),
		makefield("B", types.Types[types.TUINT8]),
		makefield("noverflow", types.Types[types.TUINT16]),
		makefield("hash0", types.Types[types.TUINT32]), // Used in walk.go for OMAKEMAP.
		makefield("buckets", types.NewPtr(bmap)),       // Used in walk.go for OMAKEMAP.
		makefield("oldbuckets", types.NewPtr(bmap)),
		makefield("nevacuate", types.Types[types.TUINTPTR]),
		makefield("extra", types.Types[types.TUNSAFEPTR]),
	})
	obj.StubTypes["runtime.hmap"] = DwarfType{hmap}
	bmap.SetFields([]*types.Field{
		makefield("tophash", types.NewArray(types.Types[types.TUINT8], 8)),
	})
	obj.StubTypes["runtime.bmap"] = DwarfType{bmap}

	//type sudog struct {
	//	g *g
	//	next *sudog
	//	prev *sudog
	//	elem unsafe.Pointer // data element (may point to stack)

	//	acquiretime int64
	//	releasetime int64
	//	ticket      uint32

	//	isSelect bool
	//	success bool
	//
	//	parent   *sudog // semaRoot binary tree
	//	waitlink *sudog // g.waiting list or semaRoot
	//	waittail *sudog // semaRoot
	//	c        *hchan // channel
	//}

	sudog.SetFields([]*types.Field{
		makefield("g", types.Types[types.TUNSAFEPTR]),
		makefield("next", types.NewPtr(sudog)),
		makefield("prev", types.NewPtr(sudog)),
		makefield("elem", types.Types[types.TUNSAFEPTR]),
		makefield("acquiretime", types.Types[types.TINT64]), // Used in walk.go for OMAKEMAP.
		makefield("releasetime", types.Types[types.TINT64]), // Used in walk.go for OMAKEMAP.
		makefield("ticket", types.Types[types.TUINT32]),
		makefield("isSelect", types.Types[types.TBOOL]),
		makefield("success", types.Types[types.TBOOL]),
		makefield("parent", types.NewPtr(sudog)),
		makefield("waitlink", types.NewPtr(sudog)),
		makefield("waittail", types.NewPtr(sudog)),
		makefield("waittail", types.NewPtr(hchan)),
	})
	obj.StubTypes["runtime.sudog"] = DwarfType{sudog}

	//type waitq struct {
	//	first *sudog
	//	last  *sudog
	//}

	waitq.SetFields([]*types.Field{
		makefield("first", types.NewPtr(sudog)),
		makefield("last", types.NewPtr(sudog)),
	})
	obj.StubTypes["runtime.waitq"] = DwarfType{waitq}

	//type hchan struct {
	//	qcount   uint
	//	dataqsiz uint
	//	buf      unsafe.Pointer
	//	elemsize uint16
	//	closed   uint32
	//	elemtype *_type
	//	sendx    uint
	//	recvx    uint
	//	recvq    waitq
	//	sendq    waitq
	//	lock mutex
	//}

	mutex.SetFields([]*types.Field{
		makefield("key", types.Types[types.TUINTPTR]),
	})

	hchan.SetFields([]*types.Field{
		makefield("qcount", types.Types[types.TUINT]),
		makefield("dataqsiz", types.Types[types.TUINT]),
		makefield("buf", types.Types[types.TUNSAFEPTR]),
		makefield("elemsize", types.Types[types.TUNSAFEPTR]),
		makefield("closed", types.Types[types.TUINT32]), // Used in walk.go for OMAKEMAP.
		makefield("elemtype", types.NewPtr(_type)),      // Used in walk.go for OMAKEMAP.
		makefield("sendx", types.Types[types.TUINT]),
		makefield("recvx", types.Types[types.TUINT]),
		makefield("recvq", waitq),
		makefield("sendq", waitq),
		makefield("lock", mutex),
	})
	obj.StubTypes["runtime.hchan"] = DwarfType{hchan}

	//type eface struct {
	//	_type *_type
	//	data  unsafe.Pointer
	//}

	eface.SetFields([]*types.Field{
		makefield("_type", types.NewPtr(_type)),
		makefield("data", types.Types[types.TUNSAFEPTR]),
	})
	obj.StubTypes["runtime.eface"] = DwarfType{eface}

	//type itab struct {
	//	inter *interfacetype
	//	_type *_type
	//	hash  uint32 // copy of _type.hash. Used for type switches.
	//	_     [4]byte
	//	fun   [1]uintptr // variable sized. fun[0]==0 means _type does not implement inter.
	//}

	itab.SetFields([]*types.Field{
		makefield("inter", types.Types[types.TUNSAFEPTR]),
		makefield("_type", _type),
		makefield("hash", types.Types[types.TUINT32]),
		makefield("_", types.NewArray(types.ByteType, 4)),
		makefield("fun", types.NewArray(types.Types[types.TUINTPTR], 1)),
	})
	obj.StubTypes["runtime.itab"] = DwarfType{itab}

	//type iface struct {
	//	tab  *itab
	//	data unsafe.Pointer
	//}
	//
	iface.SetFields([]*types.Field{
		makefield("tab", types.NewPtr(itab)),
		makefield("data", types.Types[types.TUNSAFEPTR]),
	})
	obj.StubTypes["runtime.iface"] = DwarfType{itab}

	for _, typ := range obj.StubTypes {
		types.CalcSize(typ.(DwarfType).Type)
	}
}

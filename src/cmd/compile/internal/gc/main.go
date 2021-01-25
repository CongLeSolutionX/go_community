// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gc

import (
	"bufio"
	"bytes"
	"cmd/compile/internal/base"
	"cmd/compile/internal/deadcode"
	"cmd/compile/internal/devirtualize"
	"cmd/compile/internal/dwarfgen"
	"cmd/compile/internal/escape"
	"cmd/compile/internal/inline"
	"cmd/compile/internal/ir"
	"cmd/compile/internal/logopt"
	"cmd/compile/internal/noder"
	"cmd/compile/internal/pkginit"
	"cmd/compile/internal/reflectdata"
	"cmd/compile/internal/ssa"
	"cmd/compile/internal/ssagen"
	"cmd/compile/internal/typecheck"
	"cmd/compile/internal/types"
	"cmd/internal/dwarf"
	"cmd/internal/obj"
	"cmd/internal/objabi"
	"cmd/internal/src"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
)

func hidePanic() {
	if base.Debug.Panic == 0 && base.Errors() > 0 {
		// If we've already complained about things
		// in the program, don't bother complaining
		// about a panic too; let the user clean up
		// the code and try again.
		if err := recover(); err != nil {
			if err == "-h" {
				panic(err)
			}
			base.ErrorExit()
		}
	}
}

// Main parses flags and Go source files specified in the command-line
// arguments, type-checks the parsed Go package, compiles functions to machine
// code, and finally writes the compiled package definition to disk.
func Main(archInit func(*ssagen.ArchInfo)) {
	base.Timer.Start("fe", "init")

	defer hidePanic()

	archInit(&ssagen.Arch)

	base.Ctxt = obj.Linknew(ssagen.Arch.LinkArch)
	base.Ctxt.DiagFunc = base.Errorf
	base.Ctxt.DiagFlush = base.FlushErrors
	base.Ctxt.Bso = bufio.NewWriter(os.Stdout)

	// UseBASEntries is preferred because it shaves about 2% off build time, but LLDB, dsymutil, and dwarfdump
	// on Darwin don't support it properly, especially since macOS 10.14 (Mojave).  This is exposed as a flag
	// to allow testing with LLVM tools on Linux, and to help with reporting this bug to the LLVM project.
	// See bugs 31188 and 21945 (CLs 170638, 98075, 72371).
	base.Ctxt.UseBASEntries = base.Ctxt.Headtype != objabi.Hdarwin

	types.LocalPkg = types.NewPkg("", "")
	types.LocalPkg.Prefix = "\"\""

	// We won't know localpkg's height until after import
	// processing. In the mean time, set to MaxPkgHeight to ensure
	// height comparisons at least work until then.
	types.LocalPkg.Height = types.MaxPkgHeight

	// pseudo-package, for scoping
	types.BuiltinPkg = types.NewPkg("go.builtin", "") // TODO(gri) name this package go.builtin?
	types.BuiltinPkg.Prefix = "go.builtin"            // not go%2ebuiltin

	// pseudo-package, accessed by import "unsafe"
	ir.Pkgs.Unsafe = types.NewPkg("unsafe", "unsafe")

	// Pseudo-package that contains the compiler's builtin
	// declarations for package runtime. These are declared in a
	// separate package to avoid conflicts with package runtime's
	// actual declarations, which may differ intentionally but
	// insignificantly.
	ir.Pkgs.Runtime = types.NewPkg("go.runtime", "runtime")
	ir.Pkgs.Runtime.Prefix = "runtime"

	// pseudo-packages used in symbol tables
	ir.Pkgs.Itab = types.NewPkg("go.itab", "go.itab")
	ir.Pkgs.Itab.Prefix = "go.itab" // not go%2eitab

	// pseudo-package used for methods with anonymous receivers
	ir.Pkgs.Go = types.NewPkg("go", "")

<<<<<<< HEAD   (79f796 [dev.go2go] go/format: parse type parameters)
	Wasm := objabi.GOARCH == "wasm"

	// Whether the limit for stack-allocated objects is much smaller than normal.
	// This can be helpful for diagnosing certain causes of GC latency. See #27732.
	smallFrames := false
	jsonLogOpt := ""

	flag.BoolVar(&compiling_runtime, "+", false, "compiling runtime")
	flag.BoolVar(&compiling_std, "std", false, "compiling standard library")
	objabi.Flagcount("%", "debug non-static initializers", &Debug['%'])
	objabi.Flagcount("B", "disable bounds checking", &Debug['B'])
	objabi.Flagcount("C", "disable printing of columns in error messages", &Debug['C']) // TODO(gri) remove eventually
	flag.StringVar(&localimport, "D", "", "set relative `path` for local imports")
	objabi.Flagcount("E", "debug symbol export", &Debug['E'])
	objabi.Flagcount("G", "accept generic code", &Debug['G'])
	objabi.Flagfn1("I", "add `directory` to import search path", addidir)
	objabi.Flagcount("K", "debug missing line numbers", &Debug['K'])
	objabi.Flagcount("L", "show full file names in error messages", &Debug['L'])
	objabi.Flagcount("N", "disable optimizations", &Debug['N'])
	objabi.Flagcount("S", "print assembly listing", &Debug['S'])
	objabi.AddVersionFlag() // -V
	objabi.Flagcount("W", "debug parse tree after type checking", &Debug['W'])
	flag.StringVar(&asmhdr, "asmhdr", "", "write assembly header to `file`")
	flag.StringVar(&buildid, "buildid", "", "record `id` as the build id in the export metadata")
	flag.IntVar(&nBackendWorkers, "c", 1, "concurrency during compilation, 1 means no concurrency")
	flag.BoolVar(&pure_go, "complete", false, "compiling complete package (no C or assembly)")
	flag.StringVar(&debugstr, "d", "", "print debug information about items in `list`; try -d help")
	flag.BoolVar(&flagDWARF, "dwarf", !Wasm, "generate DWARF symbols")
	flag.BoolVar(&Ctxt.Flag_locationlists, "dwarflocationlists", true, "add location lists to DWARF in optimized mode")
	flag.IntVar(&genDwarfInline, "gendwarfinl", 2, "generate DWARF inline info records")
	objabi.Flagcount("e", "no limit on number of errors reported", &Debug['e'])
	objabi.Flagcount("h", "halt on error", &Debug['h'])
	objabi.Flagfn1("importmap", "add `definition` of the form source=actual to import map", addImportMap)
	objabi.Flagfn1("importcfg", "read import configuration from `file`", readImportCfg)
	flag.StringVar(&flag_installsuffix, "installsuffix", "", "set pkg directory `suffix`")
	objabi.Flagcount("j", "debug runtime-initialized variables", &Debug['j'])
	objabi.Flagcount("l", "disable inlining", &Debug['l'])
	flag.StringVar(&flag_lang, "lang", "", "release to compile for")
	flag.StringVar(&linkobj, "linkobj", "", "write linker-specific object to `file`")
	objabi.Flagcount("live", "debug liveness analysis", &debuglive)
	objabi.Flagcount("m", "print optimization decisions", &Debug['m'])
	if sys.MSanSupported(objabi.GOOS, objabi.GOARCH) {
		flag.BoolVar(&flag_msan, "msan", false, "build code compatible with C/C++ memory sanitizer")
	}
	flag.BoolVar(&nolocalimports, "nolocalimports", false, "reject local (relative) imports")
	flag.StringVar(&outfile, "o", "", "write output to `file`")
	flag.StringVar(&myimportpath, "p", "", "set expected package import `path`")
	flag.BoolVar(&writearchive, "pack", false, "write to file.a instead of file.o")
	objabi.Flagcount("r", "debug generated wrappers", &Debug['r'])
	if sys.RaceDetectorSupported(objabi.GOOS, objabi.GOARCH) {
		flag.BoolVar(&flag_race, "race", false, "enable race detector")
	}
	flag.StringVar(&spectre, "spectre", spectre, "enable spectre mitigations in `list` (all, index, ret)")
	if enableTrace {
		flag.BoolVar(&trace, "t", false, "trace type-checking")
	}
	flag.StringVar(&pathPrefix, "trimpath", "", "remove `prefix` from recorded source file paths")
	flag.BoolVar(&Debug_vlog, "v", false, "increase debug verbosity")
	objabi.Flagcount("w", "debug type checking", &Debug['w'])
	flag.BoolVar(&use_writebarrier, "wb", true, "enable write barrier")
	var flag_shared bool
	var flag_dynlink bool
	if supportsDynlink(thearch.LinkArch.Arch) {
		flag.BoolVar(&flag_shared, "shared", false, "generate code that can be linked into a shared library")
		flag.BoolVar(&flag_dynlink, "dynlink", false, "support references to Go symbols defined in other shared libraries")
		flag.BoolVar(&Ctxt.Flag_linkshared, "linkshared", false, "generate code that will be linked against Go shared libraries")
	}
	flag.StringVar(&cpuprofile, "cpuprofile", "", "write cpu profile to `file`")
	flag.StringVar(&memprofile, "memprofile", "", "write memory profile to `file`")
	flag.Int64Var(&memprofilerate, "memprofilerate", 0, "set runtime.MemProfileRate to `rate`")
	var goversion string
	flag.StringVar(&goversion, "goversion", "", "required version of the runtime")
	var symabisPath string
	flag.StringVar(&symabisPath, "symabis", "", "read symbol ABIs from `file`")
	flag.StringVar(&traceprofile, "traceprofile", "", "write an execution trace to `file`")
	flag.StringVar(&blockprofile, "blockprofile", "", "write block profile to `file`")
	flag.StringVar(&mutexprofile, "mutexprofile", "", "write mutex profile to `file`")
	flag.StringVar(&benchfile, "bench", "", "append benchmark times to `file`")
	flag.BoolVar(&smallFrames, "smallframes", false, "reduce the size limit for stack allocated objects")
	flag.BoolVar(&Ctxt.UseBASEntries, "dwarfbasentries", Ctxt.UseBASEntries, "use base address selection entries in DWARF")
	flag.BoolVar(&Ctxt.Flag_go115newobj, "go115newobj", true, "use new object file format")
	flag.StringVar(&jsonLogOpt, "json", "", "version,destination for JSON compiler/optimizer logging")

	objabi.Flagparse(usage)

	for _, f := range strings.Split(spectre, ",") {
		f = strings.TrimSpace(f)
		switch f {
		default:
			log.Fatalf("unknown setting -spectre=%s", f)
		case "":
			// nothing
		case "all":
			spectreIndex = true
			Ctxt.Retpoline = true
		case "index":
			spectreIndex = true
		case "ret":
			Ctxt.Retpoline = true
		}
	}

	if spectreIndex {
		switch objabi.GOARCH {
		case "amd64":
			// ok
		default:
			log.Fatalf("GOARCH=%s does not support -spectre=index", objabi.GOARCH)
		}
	}
=======
	base.DebugSSA = ssa.PhaseOption
	base.ParseFlags()
>>>>>>> BRANCH (945680 [dev.typeparams] test: fix excluded files lookup so it works)

	// Record flags that affect the build result. (And don't
	// record flags that don't, since that would cause spurious
	// changes in the binary.)
	dwarfgen.RecordFlags("B", "N", "l", "msan", "race", "shared", "dynlink", "dwarflocationlists", "dwarfbasentries", "smallframes", "spectre")

	if !base.EnableTrace && base.Flag.LowerT {
		log.Fatalf("compiler not built with support for -t")
	}

	// Enable inlining (after RecordFlags, to avoid recording the rewritten -l).  For now:
	//	default: inlining on.  (Flag.LowerL == 1)
	//	-l: inlining off  (Flag.LowerL == 0)
	//	-l=2, -l=3: inlining on again, with extra debugging (Flag.LowerL > 1)
	if base.Flag.LowerL <= 1 {
		base.Flag.LowerL = 1 - base.Flag.LowerL
	}

	if base.Flag.SmallFrames {
		ir.MaxStackVarSize = 128 * 1024
		ir.MaxImplicitStackVarSize = 16 * 1024
	}

	if base.Flag.Dwarf {
		base.Ctxt.DebugInfo = dwarfgen.Info
		base.Ctxt.GenAbstractFunc = dwarfgen.AbstractFunc
		base.Ctxt.DwFixups = obj.NewDwarfFixupTable(base.Ctxt)
	} else {
		// turn off inline generation if no dwarf at all
		base.Flag.GenDwarfInl = 0
		base.Ctxt.Flag_locationlists = false
	}
	if base.Ctxt.Flag_locationlists && len(base.Ctxt.Arch.DWARFRegisters) == 0 {
		log.Fatalf("location lists requested but register mapping not available on %v", base.Ctxt.Arch.Name)
	}

	types.ParseLangFlag()

	if base.Flag.SymABIs != "" {
		ssagen.ReadSymABIs(base.Flag.SymABIs, base.Ctxt.Pkgpath)
	}

	if base.Compiling(base.NoInstrumentPkgs) {
		base.Flag.Race = false
		base.Flag.MSan = false
	}

	ssagen.Arch.LinkArch.Init(base.Ctxt)
	startProfile()
	if base.Flag.Race || base.Flag.MSan {
		base.Flag.Cfg.Instrumenting = true
	}
	if base.Flag.Dwarf {
		dwarf.EnableLogging(base.Debug.DwarfInl != 0)
	}
	if base.Debug.SoftFloat != 0 {
		ssagen.Arch.SoftFloat = true
	}

	if base.Flag.JSON != "" { // parse version,destination from json logging optimization.
		logopt.LogJsonOption(base.Flag.JSON)
	}

	ir.EscFmt = escape.Fmt
	ir.IsIntrinsicCall = ssagen.IsIntrinsicCall
	inline.SSADumpInline = ssagen.DumpInline
	ssagen.InitEnv()
	ssagen.InitTables()

	types.PtrSize = ssagen.Arch.LinkArch.PtrSize
	types.RegSize = ssagen.Arch.LinkArch.RegSize
	types.MaxWidth = ssagen.Arch.MAXWIDTH

	typecheck.Target = new(ir.Package)

	typecheck.NeedITab = func(t, iface *types.Type) { reflectdata.ITabAddr(t, iface) }
	typecheck.NeedRuntimeType = reflectdata.NeedRuntimeType // TODO(rsc): TypeSym for lock?

	base.AutogeneratedPos = makePos(src.NewFileBase("<autogenerated>", "<autogenerated>"), 1, 0)

	typecheck.InitUniverse()

	// Parse and typecheck input.
	noder.LoadPackage(flag.Args())

	dwarfgen.RecordPackageName()
	ssagen.CgoSymABIs()

	// Build init task.
	if initTask := pkginit.Task(); initTask != nil {
		typecheck.Export(initTask)
	}

	// Eliminate some obviously dead code.
	// Must happen after typechecking.
	for _, n := range typecheck.Target.Decls {
		if n.Op() == ir.ODCLFUNC {
			deadcode.Func(n.(*ir.Func))
		}
	}

	// Compute Addrtaken for names.
	// We need to wait until typechecking is done so that when we see &x[i]
	// we know that x has its address taken if x is an array, but not if x is a slice.
	// We compute Addrtaken in bulk here.
	// After this phase, we maintain Addrtaken incrementally.
	if typecheck.DirtyAddrtaken {
		typecheck.ComputeAddrtaken(typecheck.Target.Decls)
		typecheck.DirtyAddrtaken = false
	}
	typecheck.IncrementalAddrtaken = true

<<<<<<< HEAD   (79f796 [dev.go2go] go/format: parse type parameters)
	// set via a -d flag
	Ctxt.Debugpcln = Debug_pctab
	if flagDWARF {
		dwarf.EnableLogging(Debug_gendwarfinl != 0)
	}

	if Debug_softfloat != 0 {
		thearch.SoftFloat = true
	}

	// enable inlining.  for now:
	//	default: inlining on.  (debug['l'] == 1)
	//	-l: inlining off  (debug['l'] == 0)
	//	-l=2, -l=3: inlining on again, with extra debugging (debug['l'] > 1)
	if Debug['l'] <= 1 {
		Debug['l'] = 1 - Debug['l']
	}

	if jsonLogOpt != "" { // parse version,destination from json logging optimization.
		logopt.LogJsonOption(jsonLogOpt)
	}

	ssaDump = os.Getenv("GOSSAFUNC")
	if ssaDump != "" {
		if strings.HasSuffix(ssaDump, "+") {
			ssaDump = ssaDump[:len(ssaDump)-1]
			ssaDumpStdout = true
		}
		spl := strings.Split(ssaDump, ":")
		if len(spl) > 1 {
			ssaDump = spl[0]
			ssaDumpCFG = spl[1]
		}
	}

	trackScopes = flagDWARF

	Widthptr = thearch.LinkArch.PtrSize
	Widthreg = thearch.LinkArch.RegSize

	// initialize types package
	// (we need to do this to break dependencies that otherwise
	// would lead to import cycles)
	types.Widthptr = Widthptr
	types.Dowidth = dowidth
	types.Fatalf = Fatalf
	types.Sconv = func(s *types.Sym, flag, mode int) string {
		return sconv(s, FmtFlag(flag), fmtMode(mode))
	}
	types.Tconv = func(t *types.Type, flag, mode int) string {
		return tconv(t, FmtFlag(flag), fmtMode(mode))
	}
	types.FormatSym = func(sym *types.Sym, s fmt.State, verb rune, mode int) {
		symFormat(sym, s, verb, fmtMode(mode))
	}
	types.FormatType = func(t *types.Type, s fmt.State, verb rune, mode int) {
		typeFormat(t, s, verb, fmtMode(mode))
	}
	types.TypeLinkSym = func(t *types.Type) *obj.LSym {
		return typenamesym(t).Linksym()
	}
	types.FmtLeft = int(FmtLeft)
	types.FmtUnsigned = int(FmtUnsigned)
	types.FErr = int(FErr)
	types.Ctxt = Ctxt

	initUniverse()

	dclcontext = PEXTERN
	nerrors = 0

	autogeneratedPos = makePos(src.NewFileBase("<autogenerated>", "<autogenerated>"), 1, 0)

	timings.Start("fe", "loadsys")
	loadsys()

	timings.Start("fe", "parse")
	lines := parseFiles(flag.Args(), Debug['G'] != 0)
	timings.Stop()
	timings.AddEvent(int64(lines), "lines")
	if Debug['G'] != 0 {
		// can only parse generic code for now
		if nerrors+nsavederrors != 0 {
			errorexit()
		}
		return
	}

	finishUniverse()

	recordPackageName()

	typecheckok = true

	// Process top-level declarations in phases.

	// Phase 1: const, type, and names and types of funcs.
	//   This will gather all the information about types
	//   and methods but doesn't depend on any of it.
	//
	//   We also defer type alias declarations until phase 2
	//   to avoid cycles like #18640.
	//   TODO(gri) Remove this again once we have a fix for #25838.

	// Don't use range--typecheck can add closures to xtop.
	timings.Start("fe", "typecheck", "top1")
	for i := 0; i < len(xtop); i++ {
		n := xtop[i]
		if op := n.Op; op != ODCL && op != OAS && op != OAS2 && (op != ODCLTYPE || !n.Left.Name.Param.Alias) {
			xtop[i] = typecheck(n, ctxStmt)
		}
	}

	// Phase 2: Variable assignments.
	//   To check interface assignments, depends on phase 1.

	// Don't use range--typecheck can add closures to xtop.
	timings.Start("fe", "typecheck", "top2")
	for i := 0; i < len(xtop); i++ {
		n := xtop[i]
		if op := n.Op; op == ODCL || op == OAS || op == OAS2 || op == ODCLTYPE && n.Left.Name.Param.Alias {
			xtop[i] = typecheck(n, ctxStmt)
		}
	}

	// Phase 3: Type check function bodies.
	// Don't use range--typecheck can add closures to xtop.
	timings.Start("fe", "typecheck", "func")
	var fcount int64
	for i := 0; i < len(xtop); i++ {
		n := xtop[i]
		if op := n.Op; op == ODCLFUNC || op == OCLOSURE {
			Curfn = n
			decldepth = 1
			saveerrors()
			typecheckslice(Curfn.Nbody.Slice(), ctxStmt)
			checkreturn(Curfn)
			if nerrors != 0 {
				Curfn.Nbody.Set(nil) // type errors; do not compile
			}
			// Now that we've checked whether n terminates,
			// we can eliminate some obviously dead code.
			deadcode(Curfn)
			fcount++
		}
	}
	// With all types checked, it's now safe to verify map keys. One single
	// check past phase 9 isn't sufficient, as we may exit with other errors
	// before then, thus skipping map key errors.
	checkMapKeys()
	timings.AddEvent(fcount, "funcs")

	if nsavederrors+nerrors != 0 {
		errorexit()
	}

	// Phase 4: Decide how to capture closed variables.
	// This needs to run before escape analysis,
	// because variables captured by value do not escape.
	timings.Start("fe", "capturevars")
	for _, n := range xtop {
		if n.Op == ODCLFUNC && n.Func.Closure != nil {
			Curfn = n
			capturevars(n)
		}
	}
	capturevarscomplete = true

	Curfn = nil

	if nsavederrors+nerrors != 0 {
		errorexit()
	}

	// Phase 5: Inlining
	timings.Start("fe", "inlining")
	if Debug_typecheckinl != 0 {
		// Typecheck imported function bodies if debug['l'] > 1,
=======
	if base.Debug.TypecheckInl != 0 {
		// Typecheck imported function bodies if Debug.l > 1,
>>>>>>> BRANCH (945680 [dev.typeparams] test: fix excluded files lookup so it works)
		// otherwise lazily when used or re-exported.
		typecheck.AllImportedBodies()
	}

	// Inlining
	base.Timer.Start("fe", "inlining")
	if base.Flag.LowerL != 0 {
		inline.InlinePackage()
	}

	// Devirtualize.
	for _, n := range typecheck.Target.Decls {
		if n.Op() == ir.ODCLFUNC {
			devirtualize.Func(n.(*ir.Func))
		}
	}
	ir.CurFunc = nil

	// Escape analysis.
	// Required for moving heap allocations onto stack,
	// which in turn is required by the closure implementation,
	// which stores the addresses of stack variables into the closure.
	// If the closure does not escape, it needs to be on the stack
	// or else the stack copier will not update it.
	// Large values are also moved off stack in escape analysis;
	// because large values may contain pointers, it must happen early.
	base.Timer.Start("fe", "escapes")
	escape.Funcs(typecheck.Target.Decls)

	// Collect information for go:nowritebarrierrec
	// checking. This must happen before transforming closures during Walk
	// We'll do the final check after write barriers are
	// inserted.
	if base.Flag.CompilingRuntime {
		ssagen.EnableNoWriteBarrierRecCheck()
	}

	// Prepare for SSA compilation.
	// This must be before CompileITabs, because CompileITabs
	// can trigger function compilation.
	typecheck.InitRuntime()
	ssagen.InitConfig()

	// Just before compilation, compile itabs found on
	// the right side of OCONVIFACE so that methods
	// can be de-virtualized during compilation.
	ir.CurFunc = nil
	reflectdata.CompileITabs()

	// Compile top level functions.
	// Don't use range--walk can add functions to Target.Decls.
	base.Timer.Start("be", "compilefuncs")
	fcount := int64(0)
	for i := 0; i < len(typecheck.Target.Decls); i++ {
		if fn, ok := typecheck.Target.Decls[i].(*ir.Func); ok {
			enqueueFunc(fn)
			fcount++
		}
	}
	base.Timer.AddEvent(fcount, "funcs")

	compileFunctions()

	if base.Flag.CompilingRuntime {
		// Write barriers are now known. Check the call graph.
		ssagen.NoWriteBarrierRecCheck()
	}

	// Finalize DWARF inline routine DIEs, then explicitly turn off
	// DWARF inlining gen so as to avoid problems with generated
	// method wrappers.
	if base.Ctxt.DwFixups != nil {
		base.Ctxt.DwFixups.Finalize(base.Ctxt.Pkgpath, base.Debug.DwarfInl != 0)
		base.Ctxt.DwFixups = nil
		base.Flag.GenDwarfInl = 0
	}

	// Write object data to disk.
	base.Timer.Start("be", "dumpobj")
	dumpdata()
	base.Ctxt.NumberSyms()
	dumpobj()
	if base.Flag.AsmHdr != "" {
		dumpasmhdr()
	}

	ssagen.CheckLargeStacks()
	typecheck.CheckFuncStack()

	if len(compilequeue) != 0 {
		base.Fatalf("%d uncompiled functions", len(compilequeue))
	}

	logopt.FlushLoggedOpts(base.Ctxt, base.Ctxt.Pkgpath)
	base.ExitIfErrors()

	base.FlushErrors()
	base.Timer.Stop()

	if base.Flag.Bench != "" {
		if err := writebench(base.Flag.Bench); err != nil {
			log.Fatalf("cannot write benchmark data: %v", err)
		}
	}
}

func writebench(filename string) error {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	fmt.Fprintln(&buf, "commit:", objabi.Version)
	fmt.Fprintln(&buf, "goos:", runtime.GOOS)
	fmt.Fprintln(&buf, "goarch:", runtime.GOARCH)
	base.Timer.Write(&buf, "BenchmarkCompile:"+base.Ctxt.Pkgpath+":")

	n, err := f.Write(buf.Bytes())
	if err != nil {
		return err
	}
	if n != buf.Len() {
		panic("bad writer")
	}

	return f.Close()
}

func makePos(b *src.PosBase, line, col uint) src.XPos {
	return base.Ctxt.PosTable.XPos(src.MakePos(b, line, col))
}

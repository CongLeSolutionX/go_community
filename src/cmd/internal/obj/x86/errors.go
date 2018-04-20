package x86

import (
	"cmd/internal/obj"
)

// errorKind records error cause information.
type errorKind int

const (
	// ekindArgs marks unmatched instruction error.
	// Either there is no ytab for specified arguments combination,
	// or there is argument check error.
	ekindArgs errorKind = iota

	// ekindInstruction marks matched instruction error.
	ekindInstruction

	// ekindMemArg marks matched instruction error at memory argument.
	ekindMemArg
)

// asmError carries assembler error info.
type asmError struct {
	kind errorKind
	text string
}

var (
	errMemScale = &asmError{
		kind: ekindMemArg,
		text: "scaling must be explicit and equal to 1/2/4/8",
	}

	errInvalidInstruction = &asmError{
		kind: ekindInstruction,
		text: "invalid instruction",
	}

	errModeNot64 = &asmError{
		kind: ekindInstruction,
		text: "can't encode in 32-bit mode",
	}

	errModeNot32 = &asmError{
		kind: ekindInstruction,
		text: "can't encode in 64-bit mode",
	}

	errIllegalArgs = &asmError{
		kind: ekindArgs,
		text: "illegal arguments combination",
	}

	errOffsetOverflow = &asmError{
		kind: ekindMemArg,
		text: "offset too large",
	}

	errMemIndex = &asmError{
		kind: ekindMemArg,
		text: "invalid index register",
	}

	errMemBaseModeNot64 = &asmError{
		kind: ekindMemArg,
		text: "invalid base register for 32-bit mode",
	}

	errMemIndexModeNot64 = &asmError{
		kind: ekindMemArg,
		text: "invalid index register for 32-bit mode",
	}

	errMemBase = &asmError{
		kind: ekindMemArg,
		text: "invalid base register",
	}

	errAVX2gather = &asmError{
		kind: ekindInstruction,
		text: "mask, index, and destination registers should be distinct",
	}
)

// printAsmError outputs p assembly error using ctxt.Diag.
func printAsmError(ctxt *obj.Link, ab *AsmBuf, p *obj.Prog) {
	ctxt.Diag("%v: %s", p, asmErrorMessage(ctxt, ab, p))
}

// asmErrorMessage returns error message for ab.err which occured during p encoding.
// Adds any useful context it can infer from specified arguments.
func asmErrorMessage(ctxt *obj.Link, ab *AsmBuf, p *obj.Prog) string {
	// withArg returns s annotated with argument info.
	withArg := func(s string, a *obj.Addr) string {
		return s + " in " + obj.Dconv(p, a)
	}

	switch ab.err.kind {
	case ekindInstruction:
		return ab.err.text

	case ekindArgs:
		for _, a := range collectProgArgs(p) {
			switch oclass(ctxt, a) {
			case Yxxx:
				return withArg("illegal argument", a)
			case yBadSPindex:
				return withArg("can't use SP as the index register", a)
			case yBadRegModeNot64:
				return withArg("invalid register for 32-bit mode", a)
			case yBadIndir:
				return withArg("illegal indirect argument", a)
			case yBadGlobalName:
				return withArg("global vars can't use index registers", a)
			case yBadAutoParamName:
				return withArg("auto/param arguments must be accesed with SP as base register", a)
			case yBadMemName:
				return withArg("unknown addr name in memory argument", a)
			case yBadMemIndexModeNot64:
				return withArg(errMemIndexModeNot64.text, a)
			case yBadGOTRef:
				return withArg("unexpected addr with @GOT ref", a)
			case yBadAddr:
				return withArg("illegal addr argument", a)
			case yBadConst:
				return withArg("unexpected const with symbol", a)
			case yBadArgType:
				return withArg("unexpected argument type", a)
			case yBadReg:
				return withArg("unknown register used", a)
			}
		}
		return ab.err.text

	case ekindMemArg:
		for _, a := range collectProgArgs(p) {
			if a.Type == obj.TYPE_MEM {
				return withArg(ab.err.text, a)
			}
		}
		return ab.err.text

	default:
		return "invalid instruction"
	}
}

// collectProgArgs returns p arguments slice.
// Should not be used in performance-sensitive code paths.
func collectProgArgs(p *obj.Prog) []*obj.Addr {
	args := make([]*obj.Addr, 0, 6)
	if p.From.Type != obj.TYPE_NONE {
		args = append(args, &p.From)
	}
	for i := range p.RestArgs {
		args = append(args, &p.RestArgs[i])
	}
	if p.To.Type != obj.TYPE_NONE {
		args = append(args, &p.To)
	}
	return args
}

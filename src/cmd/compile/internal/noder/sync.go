// UNREVIEWED

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package noder

import (
	"fmt"
	"runtime"
	"strings"
)

// enableSync controls whether sync markers are written into unified
// IR's export data format and also whether they're expected when
// reading them back in. They're inessential to the correct
// functioning of unified IR, but are helpful during development to
// detect mistakes.
const enableSync = true

// enableSyncFrames controls how many stack frames of caller
// information should be recorded at each sync point. This further
// grows the export data format size (but only for debug builds), but
// is very useful for identifying exactly what pair of reader/writer
// code has fallen out of sync.
const enableSyncFrames = 6

func fmtFrames(pcs ...uintptr) (res []string) {
	res = make([]string, 0, len(pcs))

	frames := runtime.CallersFrames(pcs)
	for {
		frame, more := frames.Next()
		if !more {
			return
		}

		// Trim package from function name. It's just redundant noise.
		name := strings.TrimPrefix(frame.Function, "cmd/compile/internal/noder.")

		res = append(res, fmt.Sprintf("%s:%v: %s +0x%v", frame.File, frame.Line, name, frame.PC-frame.Entry))
	}
}

// syncMarker is an enum type that represents markers that may be
// written to export data to ensure the reader and writer stay
// synchronized.
type syncMarker int

//go:generate stringer -type=syncMarker -trimprefix=sync

// TODO(mdempsky): Cleanup unneeded sync markers.

// TODO(mdempsky): Split these markers into public/stable markers, and
// private ones. Also, trim unused ones.
const (
	_ syncMarker = iota
	syncNode
	syncBool
	syncInt64
	syncUint64
	syncString
	syncPos
	syncPkg
	syncSym
	syncSelector
	syncKind
	syncType
	syncTypePkg
	syncSignature
	syncParam
	syncOp
	syncObject
	syncExpr
	syncStmt
	syncDecl
	syncConstDecl
	syncFuncDecl
	syncTypeDecl
	syncVarDecl
	syncPragma
	syncValue
	syncEOF
	syncMethod
	syncFuncBody
	syncUse
	syncUseObj
	syncObjectIdx
	syncTypeIdx
	syncBOF
	syncEntry
	syncOpenScope
	syncCloseScope
	syncGlobal
	syncLocal
	syncDefine
	syncDefLocal
	syncUseLocal
	syncDefGlobal
	syncUseGlobal
	syncTypeParams
	syncUseLabel
	syncDefLabel
	syncFuncLit
	syncCommonFunc
	syncBodyRef
	syncLinksymExt
	syncHack
	syncSetlineno
	syncName
	syncImportDecl
	syncDeclNames
	syncDeclName
	syncExprList
	syncExprs
	syncWrapname
	syncTypeExpr
	syncTypeExprOrNil
	syncChanDir
	syncParams
	syncCloseAnotherScope
	syncSum
	syncUnOp
	syncBinOp
	syncStructType
	syncInterfaceType
	syncPackname
	syncEmbedded
	syncStmts
	syncStmtsFall
	syncStmtFall
	syncBlockStmt
	syncIfStmt
	syncForStmt
	syncSwitchStmt
	syncRangeStmt
	syncCaseClause
	syncCommClause
	syncSelectStmt
	syncDecls
	syncLabeledStmt
	syncCompLit

	sync1
	sync2
	sync3
	sync4

	syncN
	syncDefImplicit
	syncUseName
	syncUseObjLocal
	syncAddLocal
	syncBothSignature
	syncSetUnderlying
	syncLinkname
	syncStmt1
	syncStmtsEnd
	syncDeclare
	syncTopDecls
	syncTopConstDecl
	syncTopFuncDecl
	syncTopTypeDecl
	syncTopVarDecl
	syncObject1
	syncAddBody
	syncLabel
	syncFuncExt
	syncMethExt
	syncOptLabel
	syncScalar
	syncStmtDecls
	syncDeclLocal
	syncObjLocal
	syncObjLocal1
	syncDeclareLocal
	syncPublic
	syncPrivate
	syncRelocs
	syncReloc
	syncUseReloc
	syncVarExt
	syncPkgDef
	syncTypeExt
	syncVal
	syncCodeObj
	syncPosBase
	syncLocalIdent
	syncTypeParamNames
	syncTypeParamBounds
	syncImplicitTypes
)

// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ld

import (
	"cmd/internal/objabi"
	"cmd/internal/sys"
)

func MakeTarget(ctxt *Link) *Target {
	return &Target{
		Arch:          ctxt.Arch,
		headType:      ctxt.HeadType,
		linkMode:      ctxt.LinkMode,
		buildMode:     ctxt.BuildMode,
		linkShared:    ctxt.linkShared,
		canUsePlugins: ctxt.canUsePlugins,
		isELF:         ctxt.IsELF,
	}
}

// Target holds the configuration we're building for.
type Target struct {
	Arch *sys.Arch

	headType objabi.HeadType

	linkMode  LinkMode
	buildMode BuildMode

	linkShared    bool
	canUsePlugins bool
	isELF         bool
}

//
// Target type functions
//

func (t *Target) IsShared() bool {
	return t.buildMode == BuildModeShared
}

func (t *Target) IsPlugin() bool {
	return t.buildMode == BuildModePlugin
}

func (t *Target) IsExternal() bool {
	return t.linkMode == LinkExternal
}

func (t *Target) IsPIE() bool {
	return t.buildMode == BuildModePIE
}

func (t *Target) IsSharedGoLink() bool {
	return t.linkShared
}

func (t *Target) CanUsePlugins() bool {
	return t.canUsePlugins
}

func (t *Target) IsElf() bool {
	return t.isELF
}

func (t *Target) IsDynlinkingGo() bool {
	return t.IsShared() || t.IsSharedGoLink() || t.IsPlugin() || t.CanUsePlugins()
}

//
// Processor functions
//

func (t *Target) IsArm() bool {
	return t.Arch.Family == sys.ARM
}

func (t *Target) IsAMD64() bool {
	return t.Arch.Family == sys.AMD64
}

func (t *Target) IsPPC64() bool {
	return t.Arch.Family == sys.PPC64
}

func (t *Target) IsS390X() bool {
	return t.Arch.Family == sys.S390X
}

//
// OS Functions
//

func (t *Target) IsDarwin() bool {
	return t.headType == objabi.Hdarwin
}

func (t *Target) IsWindows() bool {
	return t.headType == objabi.Hwindows
}

func (t *Target) IsPlan9() bool {
	return t.headType == objabi.Hplan9
}

func (t *Target) IsAix() bool {
	return t.headType == objabi.Haix
}

func (t *Target) IsSolaris() bool {
	return t.headType == objabi.Hsolaris
}

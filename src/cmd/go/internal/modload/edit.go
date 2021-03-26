// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package modload

import (
	"context"
	"reflect"
	"sort"

	"golang.org/x/mod/module"
	"golang.org/x/mod/semver"
)

func editRequirements(ctx context.Context, rs *Requirements, tryUpgrade, mustSelect []module.Version) (edited *Requirements, changed bool, err error) {
	mg, err := rs.Graph(ctx)
	if err != nil {
		return rs, false, err
	}

	limiter := newVersionLimiter(nil)
	if rs.depth == lazy {
		// The go.mod file records every relevant module explicitly.
		//
		// If we need to downgrade an existing root or a new root found in
		// tryUpgrade, we don't want to allow that downgrade to incidentally upgrade
		// a relevant module to some arbitrary version. However, we don't care about
		// arbitrary upgrades to otherwise-irrelevant modules.
		for _, m := range rs.rootModules {
			v := mg.Selected(m.Path)
			limiter.limitTo(module.Version{Path: m.Path, Version: v})
		}
	} else {
		// Eager go.mod files don't indicate which transitive dependencies are
		// actually relevant to the main module, so we have to assume that any
		// module that could have provided any package — that is, any module whose
		// selected version was not "none" — may be relevant.
		for _, m := range mg.BuildList() {
			limiter.limitTo(m)
		}
	}

	var eagerUpgrades []module.Version
	if rs.depth == lazy {
		for _, m := range tryUpgrade {
			if m.Path == Target.Path {
				// Target is already considered to be higher than any possible m, so we
				// won't be upgrading to it anyway and there is no point scanning its
				// dependencies.
				continue
			}

			summary, err := goModSummary(m)
			if err != nil {
				return rs, false, err
			}
			if summary.depth() == eager {
				// For efficiency, we'll load all of the eager upgrades as one big
				// graph, rather than loading the (potentially-overlapping) subgraph for
				// each upgrade individually.
				eagerUpgrades = append(eagerUpgrades, m)
				continue
			}

			for _, r := range summary.require {
				limiter.allow(r)
			}
		}
	} else {
		eagerUpgrades = tryUpgrade
	}

	// Compute the max versions for eager upgrades all together.
	// Since these modules are eager, we'll end up scanning all of their
	// transitive dependencies no matter which versions end up selected,
	// and since we have a large dependency graph to scan we might get
	// a significant benefit from not revisiting dependencies that are at
	// common versions among multiple upgrades.
	if len(eagerUpgrades) > 0 {
		upgradeGraph, err := readModGraph(ctx, eager, eagerUpgrades)
		if err != nil {
			if go117LazyTODO {
				// Compute the requirement path from a module path in tryUpgrade to the
				// error, and the requirement path (if any) from rs.rootModules to the
				// tryUpgrade module path. Return a *mvs.BuildListError showing the
				// concatenation of the paths (with an upgrade in the middle).
			}
			return rs, false, err
		}

		for _, r := range upgradeGraph.BuildList() {
			// Upgrading to m would upgrade to r, and the caller requested that we
			// try to upgrade to m, so it's ok to upgrade to r.
			limiter.allow(r)
		}
	}

	if len(mustSelect) > 0 {
		mustGraph, err := readModGraph(ctx, rs.depth, mustSelect)
		if err != nil {
			return rs, false, err
		}

		for _, r := range mustGraph.BuildList() {
			limiter.allow(r)
		}

		// The versions in mustSelect override whatever we would naively select —
		// we will downgrade other modules as needed in order to meet them.
		for _, m := range mustSelect {
			limiter.limitTo(m)
		}

		var conflicts []Conflict
		for _, m := range mustSelect {
			dq := limiter.check(m, rs.depth)
			switch {
			case dq.err != nil:
				return rs, false, err
			case dq.conflict != module.Version{}:
				conflicts = append(conflicts, Conflict{
					Source: m,
					Dep:    dq.conflict,
					Constraint: module.Version{
						Path:    dq.conflict.Path,
						Version: limiter.max[dq.conflict.Path],
					},
				})
			}
			limiter.selected[m.Path] = m.Version
		}
		if len(conflicts) > 0 {
			return rs, false, &ConstraintError{Conflicts: conflicts}
		}
	}

	for _, m := range tryUpgrade {
		if err := limiter.upgradeToward(ctx, m, rs.depth); err != nil {
			return rs, false, err
		}
	}
	for _, m := range rs.rootModules {
		if err := limiter.upgradeToward(ctx, m, rs.depth); err != nil {
			return rs, false, err
		}
	}

	rootModules := make([]module.Version, 0, len(limiter.selected))
	for path, v := range limiter.selected {
		if path != Target.Path {
			rootModules = append(rootModules, module.Version{Path: path, Version: v})
		}
	}
	module.Sort(rootModules)
	if reflect.DeepEqual(rootModules, rs.rootModules) {
		return rs, false, nil
	}

	// A module that is not even in the build list necessarily cannot provide
	// any imported packages. Mark as direct only the direct modules that are
	// still in the build list.
	//
	// TODO(bcmills): Would it make more sense to leave the direct map as-is
	// but allow it to refer to modules that are no longer in the build list?
	// That might complicate updateRoots, but it may be cleaner in other ways.
	direct := make(map[string]bool, len(rs.direct))
	for _, m := range rootModules {
		if rs.direct[m.Path] {
			direct[m.Path] = true
		}
	}
	return newRequirements(rs.depth, rootModules, direct), true, nil
}

// A versionLimiter tracks the versions that may be selected for each module
// subject to constraints on the maximum versions of transitive dependencies.
type versionLimiter struct {
	// max maps each module path to the maximum version that may be selected for
	// that path. Paths with no entry are unrestricted and assumed to be
	// irrelevant; irrelevant dependencies of lazy modules will not be followed
	// to check for conflicts.
	max map[string]string

	// selected maps each module path to a version of that path (if known) whose
	// transitive dependencies do not violate any max version. The version kept
	// is the highest one found during any call to upgradeToward for the given
	// module path.
	//
	// If a higher acceptable version is found during a call to upgradeToward for
	// some *other* module path, that does not update the selected version.
	// Ignoring those versions keeps the downgrades computed for two modules
	// together close to the individual downgrades that would be computed for each
	// module in isolation. (The only way one module can affect another is if the
	// final downgraded version of the one module explicitly requires a higher
	// version of the other.)
	//
	// Version "none" of every module is always known not to violate any max
	// version, so paths at version "none" are omitted.
	selected map[string]string

	// dqReason records whether and why each each encountered version is
	// disqualified.
	dqReason map[module.Version]dqState

	// requiring maps each not-yet-disqualified module version to the versions
	// that directly require it. If that version becomes disqualified, the
	// disqualification will be propagated to all of the versions in the list.
	requiring map[module.Version][]module.Version
}

// A dqState indicates whether and why a module version is “disqualified” from
// being used in a way that would incorporate its requirements.
//
// The zero dqState indicates that the module version is not known to be
// disqualified, either because it is ok or because we are currently traversing
// a cycle that includes it.
type dqState struct {
	err      error          // if non-nil, disqualified because the requirements of the module could not be read
	conflict module.Version // disqualified because the module (transitively) requires dep, which exceeds the maximum version constraint for its path
}

func (dq dqState) isDisqualified() bool {
	return dq != dqState{}
}

func newVersionLimiter(max map[string]string) *versionLimiter {
	return &versionLimiter{
		selected:  map[string]string{Target.Path: Target.Version},
		max:       max,
		dqReason:  map[module.Version]dqState{},
		requiring: map[module.Version][]module.Version{},
	}
}

// allow raises the limit for m.Path to at least m.Version.
//
// This may undo a previous call to limitTo.
func (l *versionLimiter) allow(m module.Version) {
	v, ok := l.max[m.Path]
	if !ok {
		// m.Path is already unlimited.
		return
	}
	if cmpVersion(v, m.Version) < 0 {
		l.max[m.Path] = m.Version
	}
}

// limitTo reduces the limit for m.Path to no higher than m.Version.
//
// This may undo a previous call to allow.
func (l *versionLimiter) limitTo(m module.Version) {
	v, ok := l.max[m.Path]
	if !ok || cmpVersion(v, m.Version) > 0 {
		l.max[m.Path] = m.Version
	}
}

// upgradeToward attempts to upgrade the selected version of m.Path as close as
// possible to m.Version without violating l's maximum version limits.
func (l *versionLimiter) upgradeToward(ctx context.Context, m module.Version, depth modDepth) error {
	selected, ok := l.selected[m.Path]
	if ok {
		if cmpVersion(selected, m.Version) >= 0 {
			// The selected version is already at least m, so no upgrade is needed.
			return nil
		}
	} else {
		selected = "none"
	}

	if l.check(m, depth).isDisqualified() {
		candidates, err := versions(ctx, m.Path, CheckAllowed)
		if err != nil {
			// This is likely a transient error reaching the repository,
			// rather than a permanent error with the retrieved version.
			//
			// TODO(golang.org/issue/31730, golang.org/issue/30134):
			// decode what to do based on the actual error.
			return err
		}

		// Skip to candidates < m.Version.
		i := sort.Search(len(candidates), func(i int) bool {
			return semver.Compare(candidates[i], m.Version) >= 0
		})
		candidates = candidates[:i]

		for l.check(m, depth).isDisqualified() {
			n := len(candidates)
			if n == 0 || cmpVersion(selected, candidates[n-1]) >= 0 {
				// We couldn't find a suitable candidate above the already-selected version.
				// Retain that version unmodified.
				return nil
			}
			m.Version, candidates = candidates[n-1], candidates[:n-1]
		}
	}

	l.selected[m.Path] = m.Version
	return nil
}

// check determines whether m (or its transitive dependencies) would violate l's
// maximum version limits if added to the module requirement graph.
func (l *versionLimiter) check(m module.Version, depth modDepth) dqState {
	if m.Version == "none" || m == Target {
		// version "none" has no requirements, and the dependencies of Target are
		// tautological.
		return dqState{}
	}

	if dq, seen := l.dqReason[m]; seen {
		return dq
	}
	l.dqReason[m] = dqState{}

	if max, ok := l.max[m.Path]; ok && cmpVersion(m.Version, max) > 0 {
		return l.disqualify(m, dqState{conflict: m})
	}

	summary, err := goModSummary(m)
	if err != nil {
		// If we can't load the requirements, we couldn't load the go.mod file.
		// There are a number of reasons this can happen, but this usually
		// means an older version of the module had a missing or invalid
		// go.mod file. For example, if example.com/mod released v2.0.0 before
		// migrating to modules (v2.0.0+incompatible), then added a valid go.mod
		// in v2.0.1, downgrading from v2.0.1 would cause this error.
		//
		// TODO(golang.org/issue/31730, golang.org/issue/30134): if the error
		// is transient (we couldn't download go.mod), return the error from
		// Downgrade. Currently, we can't tell what kind of error it is.
		return l.disqualify(m, dqState{err: err})
	}

	if summary.depth() == eager {
		depth = eager
	}
	for _, r := range summary.require {
		if depth == lazy {
			if _, relevant := l.max[r.Path]; !relevant {
				// r.Path is irrelevant, so we don't care at what version it is selected.
				// Because m is lazy, r's dependencies won't be followed.
				continue
			}
		}

		if dq := l.check(r, depth); dq.isDisqualified() {
			return l.disqualify(m, dq)
		}

		// r and its dependencies are (perhaps provisionally) ok.
		//
		// However, if there are cycles in the requirement graph, we may have only
		// checked a portion of the requirement graph so far, and r (and thus m) may
		// yet be disqualified by some path we have not yet visited. Remember this edge
		// so that we can disqualify m and its dependents if that occurs.
		l.requiring[r] = append(l.requiring[r], m)
	}

	return dqState{}
}

// disqualify records that m (or one of its transitive dependencies)
// violates l's maximum version limits.
func (l *versionLimiter) disqualify(m module.Version, dq dqState) dqState {
	if dq := l.dqReason[m]; dq.isDisqualified() {
		return dq
	}
	l.dqReason[m] = dq

	for _, p := range l.requiring[m] {
		l.disqualify(p, dqState{conflict: m})
	}
	// Now that we have disqualified the modules that depend on m, we can forget
	// about them — we won't need to disqualify them again.
	delete(l.requiring, m)
	return dq
}

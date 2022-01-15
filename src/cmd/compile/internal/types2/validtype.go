// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types2

// validType verifies that the given type does not "expand" infinitely
// producing a cycle in the type graph. Cycles are detected by marking
// defined types.
// (Cycles involving alias types, as in "type A = [10]A" are detected
// earlier, via the objDecl cycle detection mechanism.)
func (check *Checker) validType(typ *Named) {
	check.validType0(typ, nil, make(map[*Named]typeInfo), nil)
}

// A tparamEnv provides the environment for looking up the type arguments
// with which type parameters for a given generic type are instantiated.
type tparamEnv struct {
	tmap substMap
	link *tparamEnv
}

func (env *tparamEnv) push(typ *Named) *tparamEnv {
	targs := typ.TypeArgs()
	if targs == nil {
		return nil
	}

	n := targs.Len()
	tparams := typ.TypeParams()
	if tparams.Len() < n {
		// TODO(gri) how is this possible?
		n = tparams.Len()
	}
	tmap := make(substMap, n)
	for i := 0; i < n; i++ {
		tmap[typ.TypeParams().At(i)] = targs.At(i)
	}
	return &tparamEnv{tmap: tmap, link: env}
}

func (env *tparamEnv) pop() *tparamEnv {
	return env.link
}

func (env *tparamEnv) lookup(tpar *TypeParam) Type {
	for ; env != nil; env = env.link {
		// TODO(gri) can't use tmap.lookup because that returns the type parameter
		//           if no entry for it is found in the map
		//           => revisit implementation of lookup (at least document better)
		if targ, found := env.tmap[tpar]; found {
			return targ
		}
	}
	return nil
}

type typeInfo uint

func (check *Checker) validType0(typ Type, env *tparamEnv, info map[*Named]typeInfo, path []Object) typeInfo {
	const (
		unknown typeInfo = iota
		marked
		valid
		invalid
	)

	switch t := typ.(type) {
	case *Array:
		return check.validType0(t.elem, env, info, path)

	case *Struct:
		for _, f := range t.fields {
			if check.validType0(f.typ, env, info, path) == invalid {
				return invalid
			}
		}

	case *Union:
		for _, t := range t.terms {
			if check.validType0(t.typ, env, info, path) == invalid {
				return invalid
			}
		}

	case *Interface:
		for _, etyp := range t.embeddeds {
			if check.validType0(etyp, env, info, path) == invalid {
				return invalid
			}
		}

	case *Named:
		// Don't report a 2nd error if we already know the type is invalid
		// (e.g., if a cycle was detected earlier, via under).
		if t.underlying == Typ[Invalid] {
			info[t] = invalid
			return invalid
		}

		switch info[t] {
		case unknown:
			info[t] = marked
			info[t] = check.validType0(t.orig.fromRHS, env.push(t), info, append(path, t.obj))
		case marked:
			// cycle detected
			for i, tn := range path {
				if tn == t.obj {
					check.cycleError(path[i:])
					info[t] = invalid
					// don't modify imported types (leads to race condition, see #35049)
					if t.obj.pkg == check.pkg {
						t.underlying = Typ[Invalid]
					}
					return invalid
				}
			}
			panic("cycle start not found")
		}
		return info[t]

	case *TypeParam:
		if targ := env.lookup(t); targ != nil {
			return check.validType0(targ, env.pop(), info, path)
		}
	}

	return valid
}

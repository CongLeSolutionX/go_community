// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gc

import (
	"cmd/compile/internal/types"
	"cmd/internal/obj"
	"encoding/json"
	"io/ioutil"
	"log"
	"path"
	"sort"
	"strings"
)

var embedlist []*Node

var embedCfg struct {
	Patterns map[string][]string
	Files    map[string]string
}

func readEmbedCfg(file string) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalf("-embedcfg: %v", err)
	}
	if err := json.Unmarshal(data, &embedCfg); err != nil {
		log.Fatalf("%s: %v", file, err)
	}
	if embedCfg.Patterns == nil {
		log.Fatalf("%s: invalid embedcfg: missing Patterns", file)
	}
	if embedCfg.Files == nil {
		log.Fatalf("%s: invalid embedcfg: missing Files", file)
	}
}

func embedFileNameSplit(name string) (dir, elem string, isDir bool) {
	if name[len(name)-1] == '/' {
		isDir = true
		name = name[:len(name)-1]
	}
	i := len(name) - 1
	for i >= 0 && name[i] != '/' {
		i--
	}
	if i < 0 {
		return ".", name, isDir
	}
	return name[:i], name[i+1:], isDir
}

// embedFileLess implements the sort order for a list of embedded files.
// See the comment inside ../../../../embed/embed.go's Files struct for rationale.
func embedFileLess(x, y string) bool {
	xdir, xelem, _ := embedFileNameSplit(x)
	ydir, yelem, _ := embedFileNameSplit(y)
	return xdir < ydir || xdir == ydir && xelem < yelem
}

// initEmbed emits the init data for a //go:embed variable,
// which is either a string, a []byte, or an embed.FS.
func initEmbed(n *Node, fileSet map[string]bool) *Node {
	// TODO(mdempsky): We should just construct the
	// []byte/string/embed.FS value directly, rather than
	// initializing a static variable.

	switch n.Op {
	case OEMBEDBYTES, OEMBEDSTRING:
		if len(fileSet) != 1 {
			yyerrorl(n.Pos, "%v patterns matched %v files, want 1", n.Op, len(fileSet))
			return nil
		}

		var file string
		for key := range fileSet {
			file = key
			break
		}

		fsym, size, err := fileStringSym(n.Pos, embedCfg.Files[file], n.Op == OEMBEDSTRING, nil)
		if err != nil {
			yyerrorl(n.Pos, "embed %s: %v", file, err)
			return nil
		}

		typ := types.Types[TSTRING]
		if n.Op == OEMBEDBYTES {
			typ = types.NewSlice(types.Bytetype)
		}

		v := staticname(typ)
		sym := v.Sym.Linksym()
		off := 0
		off = dsymptr(sym, off, fsym, 0)       // data string
		off = duintptr(sym, off, uint64(size)) // len
		if n.Op == OEMBEDBYTES {
			duintptr(sym, off, uint64(size)) // cap for slice
		}
		return v

	case OEMBEDFILES:
		files := make([]string, 0, len(fileSet))
		dirSet := make(map[string]bool)
		for file := range fileSet {
			files = append(files, file)

			for dir := path.Dir(file); dir != "." && !dirSet[dir]; dir = path.Dir(dir) {
				dirSet[dir] = true
				files = append(files, dir+"/")
			}
		}
		sort.Slice(files, func(i, j int) bool {
			return embedFileLess(files[i], files[j])
		})

		fsType := resolve(oldname(embedpkg.Lookup("FS")))
		if fsType == nil || fsType.Op != OTYPE {
			Fatalf("bad embed.FS declaration: %v", fsType)
		}

		v := staticname(fsType.Type)

		slicedata := Ctxt.Lookup(`"".` + v.Sym.Name + `.files`)
		off := 0
		// []files pointed at by Files
		off = dsymptr(slicedata, off, slicedata, 3*Widthptr) // []file, pointing just past slice
		off = duintptr(slicedata, off, uint64(len(files)))
		off = duintptr(slicedata, off, uint64(len(files)))

		// embed/embed.go type file is:
		//	name string
		//	data string
		//	hash [16]byte
		// Emit one of these per file in the set.
		const hashSize = 16
		hash := make([]byte, hashSize)
		for _, file := range files {
			off = dsymptr(slicedata, off, stringsym(v.Pos, file), 0) // file string
			off = duintptr(slicedata, off, uint64(len(file)))
			if strings.HasSuffix(file, "/") {
				// entry for directory - no data
				off = duintptr(slicedata, off, 0)
				off = duintptr(slicedata, off, 0)
				off += hashSize
			} else {
				fsym, size, err := fileStringSym(v.Pos, embedCfg.Files[file], true, hash)
				if err != nil {
					yyerrorl(n.Pos, "embed %s: %v", file, err)
				}
				off = dsymptr(slicedata, off, fsym, 0) // data string
				off = duintptr(slicedata, off, uint64(size))
				off = int(slicedata.WriteBytes(Ctxt, int64(off), hash))
			}
		}
		ggloblsym(slicedata, int32(off), obj.RODATA|obj.LOCAL)
		sym := v.Sym.Linksym()
		dsymptr(sym, 0, slicedata, 0)

		return v
	}

	panic("unreachable")
}

func typecheckEmbed(n *Node) *Node {
	if n.IsDDD() {
		yyerror("invalid use of ... in call to %v", n.Op)
		n.Type = nil
		return n
	}

	// Build list of files to store.
	fileSet := make(map[string]bool)
	var bad bool
	for _, arg := range n.List.Slice() {
		if !Isconst(arg, CTSTR) || arg.Sym != nil || arg.IsDDD() {
			yyerror("arguments to %v must be string literals", n.Op)
			bad = true
			continue
		}

		pattern := arg.StringVal()
		files, ok := embedCfg.Patterns[pattern]
		if !ok {
			yyerror("invalid %v: build system did not map pattern: %q", n.Op, pattern)
			bad = true
			continue
		}
		for _, file := range files {
			if embedCfg.Files[file] == "" {
				yyerror("invalid %v: build system did not map file: %q", n.Op, file)
				bad = true
				continue
			}
			fileSet[file] = true
		}
	}
	if bad {
		n.Type = nil
		return n
	}

	vstat := initEmbed(n, fileSet)
	if vstat == nil {
		n.Type = nil
		return n
	}

	res := nod(OCONVNOP, vstat, nil)
	res.Type = vstat.Type
	res.Orig = n

	return typecheck(res, ctxExpr)
}

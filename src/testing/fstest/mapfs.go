// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fstest

import (
	"io"
	"io/fs"
	"path"
	"sort"
	"strings"
	"time"
)

// A MapFS is a simple in-memory file system for use in tests,
// represented as a map from path names (arguments to Open)
// to information about the files or directories they represent.
//
// The map need not include parent directories for files contained
// in the map; those will be synthesized if needed.
// But a directory can still be included by setting the MapFile.Mode's ModeDir bit;
// this may be necessary for detailed control over the directory's FileInfo
// or to create an empty directory.
//
// For convenience, file system accesses read directly from the map,
// so changes to the map affect later file system accesses.
// An implication is that file system accesses must not run concurrently
// with changes to the map, which would be a race.
// Another implication is that opening or reading a directory requires
// iterating over the entire map, so a MapFS should typically be used with not more
// than a few hundred entries or directory reads.
type MapFS map[string]*MapFile

// A MapFile describes a single file in a MapFS.
type MapFile struct {
	Data    []byte      // file content
	Mode    fs.FileMode // FileInfo.Mode
	ModTime time.Time   // FileInfo.ModTime
	Sys     interface{} // FileInfo.Sys
}

var _ fs.FS = MapFS(nil)
var _ fs.File = (*openMapFile)(nil)

// Open opens the named file.
func (fsys MapFS) Open(name string) (fs.File, error) {
	f := fsys[name]
	if f != nil && f.Mode&fs.ModeDir == 0 {
		// Ordinary file
		return &openMapFile{name, mapFileInfo{path.Base(name), f}, 0}, nil
	}

	// Directory, possibly synthesized.
	var list []mapFileInfo
	var elem string
	var need = make(map[string]bool)
	if name == "." {
		elem = "."
		for fname, f := range fsys {
			i := strings.Index(fname, "/")
			if i < 0 {
				list = append(list, mapFileInfo{fname, f})
			} else {
				need[fname[:i]] = true
			}
		}
	} else {
		elem = name[strings.LastIndex(name, "/")+1:]
		prefix := name + "/"
		for fname, f := range fsys {
			if strings.HasPrefix(fname, prefix) {
				felem := fname[len(prefix):]
				i := strings.Index(felem, "/")
				if i < 0 {
					list = append(list, mapFileInfo{felem, f})
				} else {
					need[fname[len(prefix):len(prefix)+i]] = true
				}
			}
		}
		if list == nil && len(need) == 0 && f == nil {
			return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
		}
	}
	for _, fi := range list {
		delete(need, fi.name)
	}
	for name := range need {
		list = append(list, mapFileInfo{name, &MapFile{Mode: fs.ModeDir}})
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].name < list[j].name
	})

	if f == nil {
		f = &MapFile{Mode: fs.ModeDir}
	}
	return &mapDir{name, mapFileInfo{elem, f}, list, 0}, nil
}

type mapFileInfo struct {
	name string
	f    *MapFile
}

func (i *mapFileInfo) Name() string       { return i.name }
func (i *mapFileInfo) Size() int64        { return int64(len(i.f.Data)) }
func (i *mapFileInfo) Mode() fs.FileMode  { return i.f.Mode }
func (i *mapFileInfo) ModTime() time.Time { return i.f.ModTime }
func (i *mapFileInfo) IsDir() bool        { return i.f.Mode&fs.ModeDir != 0 }
func (i *mapFileInfo) Sys() interface{}   { return i.f.Sys }

type openMapFile struct {
	path string
	mapFileInfo
	offset int64
}

func (f *openMapFile) Stat() (fs.FileInfo, error) { return &f.mapFileInfo, nil }

func (f *openMapFile) Close() error { return nil }

func (f *openMapFile) Read(b []byte) (int, error) {
	if f.offset >= int64(len(f.f.Data)) {
		return 0, io.EOF
	}
	if f.offset < 0 {
		return 0, &fs.PathError{Op: "read", Path: f.path, Err: fs.ErrInvalid}
	}
	n := copy(b, f.f.Data[f.offset:])
	f.offset += int64(n)
	return n, nil
}

func (f *openMapFile) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case 0:
		// offset += 0
	case 1:
		offset += f.offset
	case 2:
		offset += int64(len(f.f.Data))
	}
	if offset < 0 || offset > int64(len(f.f.Data)) {
		return 0, &fs.PathError{Op: "seek", Path: f.path, Err: fs.ErrInvalid}
	}
	f.offset = offset
	return offset, nil
}

func (f *openMapFile) ReadAt(b []byte, offset int64) (int, error) {
	if offset < 0 || offset > int64(len(f.f.Data)) {
		return 0, &fs.PathError{Op: "read", Path: f.path, Err: fs.ErrInvalid}
	}
	n := copy(b, f.f.Data[offset:])
	if n < len(b) {
		return n, io.EOF
	}
	return n, nil
}

type mapDir struct {
	path string
	mapFileInfo
	entry  []mapFileInfo
	offset int
}

func (d *mapDir) Stat() (fs.FileInfo, error) { return &d.mapFileInfo, nil }
func (d *mapDir) Close() error               { return nil }
func (d *mapDir) Read(b []byte) (int, error) {
	return 0, &fs.PathError{Op: "read", Path: d.path, Err: fs.ErrInvalid}
}

func (d *mapDir) ReadDir(count int) ([]fs.FileInfo, error) {
	n := len(d.entry) - d.offset
	if count > 0 && n > count {
		n = count
	}
	if n == 0 && count > 0 {
		return nil, io.EOF
	}
	list := make([]fs.FileInfo, n)
	for i := range list {
		list[i] = &d.entry[d.offset+i]
	}
	d.offset += n
	return list, nil
}

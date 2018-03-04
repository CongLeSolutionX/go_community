// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build js,wasm

package syscall

import (
	"syscall/js"
)

var jsProcess = js.Global.Get("process")
var jsFS = js.Global.Get("fs")
var constants = jsFS.Get("constants")

var (
	nodeWRONLY   = constants.Get("O_WRONLY").Int()
	nodeRDWR     = constants.Get("O_RDWR").Int()
	nodeCREATE   = constants.Get("O_CREAT").Int()
	nodeTRUNC    = constants.Get("O_TRUNC").Int()
	nodeAPPEND   = constants.Get("O_APPEND").Int()
	nodeEXCL     = constants.Get("O_EXCL").Int()
	nodeNONBLOCK = constants.Get("O_NONBLOCK").Int()
	nodeSYNC     = constants.Get("O_SYNC").Int()
)

func Open(path string, openmode int, perm uint32) (int, error) {
	if err := checkPath(path); err != nil {
		return 0, err
	}

	flags := 0
	if openmode&O_WRONLY != 0 {
		flags |= nodeWRONLY
	}
	if openmode&O_RDWR != 0 {
		flags |= nodeRDWR
	}
	if openmode&O_CREATE != 0 {
		flags |= nodeCREATE
	}
	if openmode&O_TRUNC != 0 {
		flags |= nodeTRUNC
	}
	if openmode&O_APPEND != 0 {
		flags |= nodeAPPEND
	}
	if openmode&O_EXCL != 0 {
		flags |= nodeEXCL
	}
	if openmode&O_NONBLOCK != 0 {
		flags |= nodeNONBLOCK
	}
	if openmode&O_SYNC != 0 {
		flags |= nodeSYNC
	}

	jsFD, err := fsCall("openSync", path, flags, perm)
	if err != nil {
		return 0, err
	}

	var entries []string
	if jsFS.Call("fstatSync", jsFD).Call("isDirectory").Bool() {
		dir := jsFS.Call("readdirSync", path)
		entries = make([]string, dir.Length())
		for i := range entries {
			entries[i] = dir.Index(i).String()
		}
	}

	return newFD(&jsFile{
		jsFD:    jsFD.Int(),
		path:    path,
		entries: entries,
	}), nil
}

func Mkdir(path string, perm uint32) error {
	if err := checkPath(path); err != nil {
		return err
	}
	_, err := fsCall("mkdirSync", path, perm)
	return err
}

func ReadDirent(fd int, buf []byte) (int, error) {
	f, err := fdToFile(fd)
	if err != nil {
		return 0, err
	}

	jsf := f.impl.(*jsFile)
	if jsf.entries == nil {
		return 0, EINVAL
	}

	n := 0
	for len(jsf.entries) > 0 {
		entry := jsf.entries[0]
		l := 2 + len(entry)
		if l > len(buf) {
			break
		}
		buf[0] = byte(l)
		buf[1] = byte(l >> 8)
		copy(buf[2:], entry)
		buf = buf[l:]
		n += l
		jsf.entries = jsf.entries[1:]
	}

	return n, nil
}

func setStat(st *Stat_t, jsSt js.Value) {
	st.Dev = int64(jsSt.Get("dev").Int())
	st.Ino = uint64(jsSt.Get("ino").Int())
	st.Mode = uint32(jsSt.Get("mode").Int())
	st.Nlink = uint32(jsSt.Get("nlink").Int())
	st.Uid = uint32(jsSt.Get("uid").Int())
	st.Gid = uint32(jsSt.Get("gid").Int())
	st.Rdev = int64(jsSt.Get("rdev").Int())
	st.Size = int64(jsSt.Get("size").Int())
	st.Blksize = int32(jsSt.Get("blksize").Int())
	st.Blocks = int32(jsSt.Get("blocks").Int())
	atime := int64(jsSt.Get("atimeMs").Int())
	st.Atime = atime / 1000
	st.AtimeNsec = (atime % 1000) * 1000000
	mtime := int64(jsSt.Get("mtimeMs").Int())
	st.Mtime = mtime / 1000
	st.MtimeNsec = (mtime % 1000) * 1000000
	ctime := int64(jsSt.Get("ctimeMs").Int())
	st.Ctime = ctime / 1000
	st.CtimeNsec = (ctime % 1000) * 1000000
}

func Stat(path string, st *Stat_t) error {
	if err := checkPath(path); err != nil {
		return err
	}
	jsSt, err := fsCall("statSync", path)
	if err != nil {
		return err
	}
	setStat(st, jsSt)
	return nil
}

func Lstat(path string, st *Stat_t) error {
	if err := checkPath(path); err != nil {
		return err
	}
	jsSt, err := fsCall("lstatSync", path)
	if err != nil {
		return err
	}
	setStat(st, jsSt)
	return nil
}

func Unlink(path string) error {
	if err := checkPath(path); err != nil {
		return err
	}
	_, err := fsCall("unlinkSync", path)
	return err
}

func Rmdir(path string) error {
	if err := checkPath(path); err != nil {
		return err
	}
	_, err := fsCall("rmdirSync", path)
	return err
}

func Chmod(path string, mode uint32) error {
	if err := checkPath(path); err != nil {
		return err
	}
	_, err := fsCall("chmodSync", path, mode)
	return err
}

func Fchmod(fd int, mode uint32) error {
	f, err := fdToFile(fd)
	if err != nil {
		return err
	}
	_, err = fsCall("fchmodSync", f.impl.(*jsFile).jsFD, mode)
	return err
}

func Chown(path string, uid, gid int) error {
	if err := checkPath(path); err != nil {
		return err
	}
	return ENOSYS
}

func Fchown(fd int, uid, gid int) error {
	return ENOSYS
}

func Lchown(path string, uid, gid int) error {
	if err := checkPath(path); err != nil {
		return err
	}
	return ENOSYS
}

func UtimesNano(path string, ts []Timespec) error {
	if err := checkPath(path); err != nil {
		return err
	}
	if len(ts) != 2 {
		return EINVAL
	}
	atime := ts[0].Sec
	mtime := ts[1].Sec
	_, err := fsCall("utimesSync", path, atime, mtime)
	return err
}

func Rename(from, to string) error {
	if err := checkPath(from); err != nil {
		return err
	}
	if err := checkPath(to); err != nil {
		return err
	}
	_, err := fsCall("renameSync", from, to)
	return err
}

func Truncate(path string, length int64) error {
	if err := checkPath(path); err != nil {
		return err
	}
	_, err := fsCall("truncateSync", path, length)
	return err
}

func Ftruncate(fd int, length int64) error {
	f, err := fdToFile(fd)
	if err != nil {
		return err
	}

	_, err = fsCall("ftruncateSync", f.impl.(*jsFile).jsFD, length)
	return err
}

func Getcwd(buf []byte) (n int, err error) {
	defer recoverErr(&err)
	cwd := jsProcess.Call("cwd").String()
	n = copy(buf, cwd)
	return n, nil
}

func Chdir(path string) (err error) {
	if err := checkPath(path); err != nil {
		return err
	}
	defer recoverErr(&err)
	jsProcess.Call("chdir", path)
	return
}

func Fchdir(fd int) error {
	f, err := fdToFile(fd)
	if err != nil {
		return err
	}

	return Chdir(f.impl.(*jsFile).path)
}

func Readlink(path string, buf []byte) (n int, err error) {
	if err := checkPath(path); err != nil {
		return 0, err
	}
	dst, err := fsCall("readlinkSync", path)
	if err != nil {
		return 0, err
	}
	n = copy(buf, dst.String())
	return n, nil
}

func Link(path, link string) error {
	if err := checkPath(path); err != nil {
		return err
	}
	if err := checkPath(link); err != nil {
		return err
	}
	_, err := fsCall("linkSync", path, link)
	return err
}

func Symlink(path, link string) error {
	if err := checkPath(path); err != nil {
		return err
	}
	if err := checkPath(link); err != nil {
		return err
	}
	_, err := fsCall("symlinkSync", path, link)
	return err
}

func Fsync(fd int) error {
	f, err := fdToFile(fd)
	if err != nil {
		return err
	}

	_, err = fsCall("fsyncSync", f.impl.(*jsFile).jsFD)
	return err
}

func fsCall(name string, args ...interface{}) (res js.Value, err error) {
	defer recoverErr(&err)
	res = jsFS.Call(name, args...)
	return
}

// checkPath checks that the path is not empty and that it contains no null characters.
func checkPath(path string) error {
	if path == "" {
		return EINVAL
	}
	for i := 0; i < len(path); i++ {
		if path[i] == '\x00' {
			return EINVAL
		}
	}
	return nil
}

func recoverErr(errPtr *error) {
	if err := recover(); err != nil {
		jsErr, ok := err.(js.Error)
		if !ok {
			panic(err)
		}
		errno, ok := errnoByCode[jsErr.Get("code").String()]
		if !ok {
			panic(err)
		}
		*errPtr = errnoErr(Errno(errno))
	}
}

// Provided by package runtime.
func now() (sec int64, nsec int32)

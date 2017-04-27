// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os

import (
	"internal/syscall/windows"
	"syscall"
	"unsafe"
)

// Stat returns the FileInfo structure describing file.
// If there is an error, it will be of type *PathError.
func (file *File) Stat() (FileInfo, error) {
	if file == nil {
		return nil, ErrInvalid
	}
	if file == nil {
		return nil, syscall.EINVAL
	}
	if file.isdir() {
		// I don't know any better way to do that for directory
		return statFullPath(file.dirinfo.path, file.dirinfo.path)
	}
	if file.name == DevNull {
		return &devNullStat, nil
	}

	ft, err := file.pfd.GetFileType()
	if err != nil {
		return nil, &PathError{"GetFileType", file.name, err}
	}
	switch ft {
	case syscall.FILE_TYPE_PIPE, syscall.FILE_TYPE_CHAR:
		return &fileStat{name: basename(file.name), filetype: ft}, nil
	}

	var d syscall.ByHandleFileInformation
	err = file.pfd.GetFileInformationByHandle(&d)
	if err != nil {
		return nil, &PathError{"GetFileInformationByHandle", file.name, err}
	}
	return &fileStat{
		name: basename(file.name),
		sys: syscall.Win32FileAttributeData{
			FileAttributes: d.FileAttributes,
			CreationTime:   d.CreationTime,
			LastAccessTime: d.LastAccessTime,
			LastWriteTime:  d.LastWriteTime,
			FileSizeHigh:   d.FileSizeHigh,
			FileSizeLow:    d.FileSizeLow,
		},
		filetype: ft,
		vol:      d.VolumeSerialNumber,
		idxhi:    d.FileIndexHigh,
		idxlo:    d.FileIndexLow,
	}, nil
}

// statFullPath is like Stat, but also accepts full path of the named file.
func statFullPath(name, fullpath string) (FileInfo, error) {
	// TODO: move _ERROR_CANT_RESOLVE_FILENAME into internal/syscall/windows
	const _ERROR_CANT_RESOLVE_FILENAME = 1921
	namep, err := syscall.UTF16PtrFromString(fixLongPath(name))
	if err != nil {
		return nil, &PathError{"Stat", name, err}
	}

	h, err := syscall.CreateFile(namep, 0, 0, nil,
		syscall.OPEN_EXISTING, syscall.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		if err == windows.ERROR_SHARING_VIOLATION {
			// try FindFirstFile now that CreateFile failed
			var fd syscall.Win32finddata
			h, err := syscall.FindFirstFile(namep, &fd)
			if err != nil {
				return nil, &PathError{"FindFirstFile", name, err}
			}
			syscall.FindClose(h)

			return &fileStat{
				name: basename(name),
				path: fullpath,
				sys: syscall.Win32FileAttributeData{
					FileAttributes: fd.FileAttributes,
					CreationTime:   fd.CreationTime,
					LastAccessTime: fd.LastAccessTime,
					LastWriteTime:  fd.LastWriteTime,
					FileSizeHigh:   fd.FileSizeHigh,
					FileSizeLow:    fd.FileSizeLow,
				},
			}, nil
		}
		// TODO: maybe just get rid of broken TestStatSymlinkLoop instead
		if err.(syscall.Errno) == _ERROR_CANT_RESOLVE_FILENAME {
			err = syscall.ELOOP
		}
		return nil, &PathError{"CreateFile", name, err}
	}
	defer syscall.CloseHandle(h)

	var d syscall.ByHandleFileInformation
	err = syscall.GetFileInformationByHandle(h, &d)
	if err != nil {
		return nil, &PathError{"GetFileInformationByHandle", name, err}
	}
	return &fileStat{
		name: basename(name),
		sys: syscall.Win32FileAttributeData{
			FileAttributes: d.FileAttributes,
			CreationTime:   d.CreationTime,
			LastAccessTime: d.LastAccessTime,
			LastWriteTime:  d.LastWriteTime,
			FileSizeHigh:   d.FileSizeHigh,
			FileSizeLow:    d.FileSizeLow,
		},
		vol:   d.VolumeSerialNumber,
		idxhi: d.FileIndexHigh,
		idxlo: d.FileIndexLow,
		// fileStat.path is used by os.SameFile to decide, if it needs
		// to fetch vol, idxhi and idxlo. But these are already set,
		// so set fileStat.path to "" to prevent os.SameFile doing it again.
		// Also do not set fileStat.filetype, because it is only used for
		// console and stdin/stdout. But you cannot call os.Stat for these.
	}, nil
}

// Stat returns a FileInfo structure describing the named file.
// If there is an error, it will be of type *PathError.
func Stat(name string) (FileInfo, error) {
	if len(name) == 0 {
		return nil, &PathError{"Stat", name, syscall.Errno(syscall.ERROR_PATH_NOT_FOUND)}
	}
	if name == DevNull {
		return &devNullStat, nil
	}
	if isAbs(name) {
		return statFullPath(name, name)
	}
	fullpath, err := syscall.FullPath(name)
	if err != nil {
		return nil, &PathError{"FullPath", name, err}
	}
	return statFullPath(name, fullpath)
}

// Lstat returns the FileInfo structure describing the named file.
// If the file is a symbolic link, the returned FileInfo
// describes the symbolic link. Lstat makes no attempt to follow the link.
// If there is an error, it will be of type *PathError.
func Lstat(name string) (FileInfo, error) {
	if len(name) == 0 {
		return nil, &PathError{"Lstat", name, syscall.Errno(syscall.ERROR_PATH_NOT_FOUND)}
	}
	if name == DevNull {
		return &devNullStat, nil
	}
	fs := &fileStat{name: basename(name)}
	namep, e := syscall.UTF16PtrFromString(fixLongPath(name))
	if e != nil {
		return nil, &PathError{"Lstat", name, e}
	}
	e = syscall.GetFileAttributesEx(namep, syscall.GetFileExInfoStandard, (*byte)(unsafe.Pointer(&fs.sys)))
	if e != nil {
		if e != windows.ERROR_SHARING_VIOLATION {
			return nil, &PathError{"GetFileAttributesEx", name, e}
		}
		// try FindFirstFile now that GetFileAttributesEx failed
		var fd syscall.Win32finddata
		h, e2 := syscall.FindFirstFile(namep, &fd)
		if e2 != nil {
			return nil, &PathError{"FindFirstFile", name, e}
		}
		syscall.FindClose(h)

		fs.sys.FileAttributes = fd.FileAttributes
		fs.sys.CreationTime = fd.CreationTime
		fs.sys.LastAccessTime = fd.LastAccessTime
		fs.sys.LastWriteTime = fd.LastWriteTime
		fs.sys.FileSizeHigh = fd.FileSizeHigh
		fs.sys.FileSizeLow = fd.FileSizeLow
	}
	fs.path = name
	if !isAbs(fs.path) {
		fs.path, e = syscall.FullPath(fs.path)
		if e != nil {
			return nil, e
		}
	}
	return fs, nil
}

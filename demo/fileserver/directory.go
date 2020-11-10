package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// errors
var (
	ErrDirectoryRead = fmt.Errorf("Failed to read directory")
	ErrFileRead      = fmt.Errorf("Failed to read file")
	ErrOutsideRoot   = fmt.Errorf("File points outside of the root directory")
	ErrHiddenFile    = fmt.Errorf("Hidden files are forbidden")
)

// Directory ...
type Directory string

// Open ...
func (root Directory) Open(relPath string) (http.File, error) {
	filename, isDir, err := root.resolve(relPath)
	if err != nil {
		return nil, err
	}
	if isDir {
		return nil, ErrFileRead
	}
	file, err := http.Dir(string(root)).Open(filename)
	if err != nil {
		return nil, ErrFileRead
	}
	return file, nil
}

// GetFileOrEntries ...
func (root Directory) GetFileOrEntries(relPath string) (file http.File, entries []*Entry, err error) {
	filename, isDir, err := root.resolve(relPath)
	if err != nil {
		return
	}
	if isDir {
		entries, err = root.getEntries(filename)
		return
	}
	file, err = http.Dir(string(root)).Open(filename)
	return
}

// GetEntries ...
func (root Directory) GetEntries(relPath string) ([]*Entry, error) {
	dir, isDir, err := root.resolve(relPath)
	if err != nil {
		return nil, err
	}
	if !isDir {
		return nil, ErrDirectoryRead
	}
	return root.getEntries(dir)
}

func (root Directory) getEntries(relPath string) ([]*Entry, error) {
	files, err := ioutil.ReadDir(relPath)
	if err != nil {
		return nil, ErrDirectoryRead
	}
	entries := make([]*Entry, 0, len(files)+1)
	if len(relPath) > 0 {
		entries = append(entries, newDirEntry("..", relPath))
	}
	for _, file := range files {
		entry := newEntry(file, relPath)
		if entry.Name[0] == '.' {
			continue
		}
		entry.FullName, _, err = root.resolve(entry.FullName)
		if err != nil {
			continue
		}
		entries = append(entries, entry)
	}
	sortEntries(entries)
	return entries, nil
}

func (root Directory) resolve(relPath string) (resolvedPath string, isDir bool, err error) {
	if root == "" {
		root = "."
	}
	if isHiddenFile(relPath) {
		err = ErrHiddenFile
		return
	}
	absRoot, err := filepath.Abs(string(root))
	if err != nil {
		return
	}
	filename := path.Join(absRoot, relPath)
	for {
		var fi os.FileInfo
		fi, err = os.Lstat(filename)
		if err != nil {
			return
		}
		if fi.Mode()&os.ModeSymlink == 0 {
			if !strings.HasPrefix(filename, absRoot) {
				err = ErrOutsideRoot
				return
			}
			resolvedPath = filepath.ToSlash(strings.TrimPrefix(filename, absRoot))
			isDir = fi.IsDir()
			if len(resolvedPath) == 0 {
				resolvedPath = "."
			} else if resolvedPath[0] == '/' {
				resolvedPath = resolvedPath[1:]
			}
			return
		}
		filename, err = os.Readlink(filename)
		if err != nil {
			return
		}
	}
}

func isHiddenFile(file string) bool {
	return len(file) > 1 && path.Base(file)[0] == '.'
}

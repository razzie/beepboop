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

// Directory ...
type Directory string

// GetEntries ...
func (root Directory) GetEntries(relPath string) ([]*Entry, error) {
	if root == "" {
		root = "."
	}
	absPath, err := root.absPath(relPath)
	if err != nil {
		return nil, err
	}
	files, err := ioutil.ReadDir(absPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to read directory")
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
		entry.FullName, err = root.resolveSymlinks(entry.FullName)
		if err != nil {
			continue
		}
		entries = append(entries, entry)
	}
	sortEntries(entries)
	return entries, nil
}

// Open ...
func (root Directory) Open(relPath string) (http.File, error) {
	if root == "" {
		root = "."
	}
	filename, err := root.resolveSymlinks(relPath)
	if err != nil {
		return nil, err
	}
	file, err := http.Dir(string(root)).Open(filename)
	if err != nil {
		return nil, fmt.Errorf("Failed to read file")
	}
	return file, nil
}

// GetFileOrEntries ...
func (root Directory) GetFileOrEntries(relPath string) (http.File, []*Entry, error) {
	if root == "" {
		root = "."
	}
	if isDir(path.Join(string(root), relPath)) {
		entries, err := root.GetEntries(relPath)
		return nil, entries, err
	}
	file, err := root.Open(relPath)
	return file, nil, err
}

func (root Directory) absPath(relPath string) (string, error) {
	absRoot, err := filepath.Abs(string(root))
	if err != nil {
		return "", err
	}
	return path.Join(absRoot, relPath), nil
}

func (root Directory) resolveSymlinks(relPath string) (string, error) {
	absRoot, err := filepath.Abs(string(root))
	if err != nil {
		return "", err
	}
	absRoot += "/"
	filename := path.Join(absRoot, relPath)
	for {
		fi, err := os.Lstat(filename)
		if err != nil {
			return "", err
		}
		if fi.Mode()&os.ModeSymlink == 0 {
			if !strings.HasPrefix(filename, absRoot) {
				return "", fmt.Errorf("File points outside of the root directory")
			}
			return strings.TrimPrefix(filename, absRoot), nil
		}
		filename, err = os.Readlink(filename)
		if err != nil {
			return "", err
		}
	}
}

func isDir(dir string) bool {
	fi, err := os.Stat(dir)
	if err != nil {
		return false
	}
	return fi.IsDir()
}

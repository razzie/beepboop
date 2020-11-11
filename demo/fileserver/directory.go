package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/razzie/beepboop"
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
	if relPath != "." {
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
	if len(relPath) > 1 && path.Base(relPath)[0] == '.' {
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

// DirectoryPage returns a beepboop.Page that handles the directory view
func DirectoryPage(root string) *beepboop.Page {
	contentTemplate, err := ioutil.ReadFile("demo/fileserver/template/directory.html")
	if err != nil {
		panic(err)
	}
	return &beepboop.Page{
		Path:            "/",
		ContentTemplate: string(contentTemplate),
		Handler: func(r *beepboop.PageRequest) *beepboop.View {
			return handleDirPage(r, Directory(root))
		},
	}
}

type dirView struct {
	Dir     string
	Entries []*Entry
}

func handleDirPage(r *beepboop.PageRequest, root Directory) *beepboop.View {
	r.Title = r.RelPath
	v := dirView{
		Dir: r.RelPath,
	}

	db := r.Context.DB
	if db != nil && db.GetCachedValue("dir:"+r.RelPath, &v.Entries) == nil {
		return r.Respond(v)
	}

	file, entries, err := root.GetFileOrEntries(r.RelPath)
	if err != nil {
		return r.ErrorView(err.Error(), http.StatusInternalServerError)
	}
	if file != nil {
		return r.FileView(file, "", false)
	}

	if db != nil {
		db.CacheValue("dir:"+r.RelPath, entries, false)
	}

	v.Entries = entries
	return r.Respond(v)
}

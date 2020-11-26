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
	ErrOutsideRoot     = fmt.Errorf("File points outside of the root directory")
	ErrSymlinkMaxDepth = fmt.Errorf("Symlink max depth exceeded")
	ErrHiddenFile      = fmt.Errorf("Hidden files are forbidden")
)

// Directory ...
type Directory string

// Open ...
func (root Directory) Open(relPath string) (http.File, error) {
	if isHiddenFile(relPath) {
		return nil, ErrHiddenFile
	}
	filename, _, err := root.resolve(relPath)
	if err != nil {
		return nil, err
	}
	if isHiddenFile(filename) {
		return nil, ErrHiddenFile
	}
	file, err := http.Dir(string(root)).Open(filename)
	if err != nil {
		return nil, fmt.Errorf("Failed to open: %s", relPath)
	}
	return file, nil
}

func (root Directory) resolve(relPath string) (resolvedPath string, isDir bool, err error) {
	if root == "" {
		root = "."
	}
	absRoot, err := filepath.Abs(string(root))
	if err != nil {
		return
	}
	var depth int
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
		if depth++; depth > 16 {
			err = ErrSymlinkMaxDepth
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
	uri := path.Clean(r.RelPath)
	r.Title = r.RelPath
	v := dirView{
		Dir: r.RelPath,
	}

	db := r.Context.DB
	if db != nil && db.GetCachedValue("dir:"+uri, &v.Entries) == nil {
		return r.Respond(v)
	}

	file, err := root.Open(r.RelPath)
	if err != nil {
		return r.ErrorView(err.Error(), http.StatusInternalServerError)
	}
	fi, err := file.Stat()
	if err != nil {
		file.Close()
		return r.ErrorView(err.Error(), http.StatusInternalServerError)
	}
	if !fi.IsDir() {
		return r.FileView(file, "", false)
	}
	defer file.Close()

	files, err := file.Readdir(-1)
	if err != nil {
		return r.ErrorView(err.Error(), http.StatusInternalServerError)
	}
	entries := make([]*Entry, 0, len(files)+1)
	if uri != "." {
		entries = append(entries, newDirEntry("..", uri))
	}
	for _, fi := range files {
		if isHiddenFile(fi.Name()) {
			continue
		}
		entries = append(entries, newEntry(fi, uri))
	}
	sortEntries(entries)

	if db != nil {
		db.CacheValue("dir:"+uri, entries, false)
	}

	v.Entries = entries
	return r.Respond(v)
}

func isHiddenFile(filename string) bool {
	filename = path.Base(filename)
	return len(filename) > 1 && strings.HasPrefix(filename, ".")
}

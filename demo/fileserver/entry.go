package main

import (
	"html/template"
	"net/http"
	"os"
	"path"
	"sort"
	"strings"
)

// Entry ...
type Entry struct {
	Prefix        template.HTML
	Name          string
	FullName      string
	PrimaryType   string
	SecondaryType string
	Size          int64
	Created       int64
	Modified      int64
	IsDirectory   bool
}

func newEntry(fi os.FileInfo, dir string) *Entry {
	if fi.IsDir() {
		return newDirEntry(fi.Name(), dir)
	}

	e := &Entry{
		Name:     fi.Name(),
		FullName: path.Join(dir, fi.Name()),
		Size:     fi.Size(),
		Created:  getCreationTime(fi).Unix(),
		Modified: fi.ModTime().Unix(),
	}
	e.PrimaryType, e.SecondaryType, e.Prefix = getFileTypeAndSymbol(e.FullName)
	return e
}

func newDirEntry(name, dir string) *Entry {
	return &Entry{
		Prefix:      "&#128194;",
		Name:        name,
		FullName:    path.Join(dir, name),
		IsDirectory: true,
	}
}

func getFileTypeAndSymbol(filename string) (string, string, template.HTML) {
	file, err := os.Open(filename)
	if err != nil {
		return "unknown", "", "&#128196;"
	}
	defer file.Close()

	var header [512]byte
	_, err = file.Read(header[:])
	if err != nil {
		return "unknown", "", "&#128196;"
	}
	mime := http.DetectContentType(header[:])

	if i := strings.Index(mime, "/"); i != -1 {
		primary, secondary := mime[:i], mime[i+1:]
		return primary, secondary, typeToSymbol(primary, secondary)
	}
	return mime, "", typeToSymbol(mime, "")
}

func typeToSymbol(primary, secondary string) template.HTML {
	switch primary {
	case "application":
		switch secondary {
		case "zip", "x-7z-compressed", "x-rar-compressed", "x-tar", "tar+gzip", "gzip", "x-bzip", "x-bzip2":
			return "&#128230;"
		case "vnd.microsoft.portable-executable", "vnd.debian.binary-package", "jar", "x-rpm":
			return "&#128187;"
		case "pdf", "msword", "vnd.openxmlformats-officedocument.wordprocessingml.document", "x-mobipocket-ebook", "epub+zip":
			return "&#128209;"
		case "x-iso9660-image", "x-cd-image", "x-raw-disk-image":
			return "&#128191;"
		case "vnd.ms-excel", "vnd.ms-powerpoint", "vnd.openxmlformats-officedocument.presentationml.presentation":
			return "&#128200;"
		}
	case "audio":
		return "&#127925;"
	case "font":
		return "&#9000;"
	case "image":
		return "&#127912;"
	case "model":
		return "&#127922;"
	case "text":
		return "&#128209;"
	case "video":
		return "&#127916;"
	}

	return "&#128196;"
}

func sortEntries(entries []*Entry) {
	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})
	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].IsDirectory && !entries[j].IsDirectory
	})
}

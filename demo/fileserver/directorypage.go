package main

import (
	"io/ioutil"
	"net/http"

	"github.com/razzie/beepboop"
)

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

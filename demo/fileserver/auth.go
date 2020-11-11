package main

import (
	"crypto/sha1"
	"encoding/hex"
	"io/ioutil"
	"net/http"
	"os"
	"path"

	"github.com/razzie/beepboop"
)

// AuthMiddleware returns a middleware that requests a password for password protected directories
func AuthMiddleware(root string) beepboop.Middleware {
	return func(r *beepboop.PageRequest) *beepboop.View {
		if r.PagePath == "/.auth/" {
			return nil
		}
		pwhash, dir, err := Directory(root).getPasswordHash(r.RelPath)
		if err != nil {
			return r.ErrorView(err.Error(), http.StatusInternalServerError)
		}
		if len(pwhash) > 0 {
			sess := r.Session()
			if accesscode, _ := sess.GetAccessCode("view", dir); accesscode != pwhash {
				//return r.ErrorView("Unauthorized", http.StatusUnauthorized)
				return r.RedirectView("/.auth/" + dir + "?r=/" + r.RelPath)
			}
		}
		return nil
	}
}

// AuthPage returns a beepboop.Page that handles password authentication for protected directories
func AuthPage(root string) *beepboop.Page {
	contentTemplate, err := ioutil.ReadFile("demo/fileserver/template/auth.html")
	if err != nil {
		panic(err)
	}
	return &beepboop.Page{
		Path:            "/.auth/",
		ContentTemplate: string(contentTemplate),
		Handler: func(r *beepboop.PageRequest) *beepboop.View {
			return handleAuthPage(r, Directory(root))
		},
	}
}

type authPageView struct {
	Error      string
	Directory  string
	AccessType string
	Redirect   string
}

func handleAuthPage(r *beepboop.PageRequest, root Directory) *beepboop.View {
	req := r.Request
	dir := path.Clean(r.RelPath)
	r.Title = dir
	v := &authPageView{
		Directory:  dir,
		AccessType: "view",
		Redirect:   req.URL.Query().Get("r"),
	}
	if len(v.Redirect) == 0 {
		v.Redirect = "/" + dir
	}
	if req.Method == "POST" {
		req.ParseForm()
		pw := req.FormValue("password")
		v.Redirect = req.FormValue("redirect")
		if pwhash, _, _ := root.getPasswordHash(dir); pwhash == hash([]byte(pw)) {
			r.Session().AddAccess("view", dir, pwhash)
			r.Log("Password accepted!")
			return r.RedirectView(v.Redirect)
		}
		v.Error = "Invalid password"
		return r.Respond(v, beepboop.WithErrorMessage(v.Error, http.StatusUnauthorized))
	}
	return r.Respond(v)
}

func (root Directory) getPasswordHash(relPath string) (pwhash, dir string, err error) {
	filename, isDir, err := root.resolve(relPath)
	if err != nil {
		return
	}
	dir = filename
	if !isDir {
		dir = path.Dir(filename)
	}
	pwfile := path.Join(string(root), dir, ".password")
	pw, err := ioutil.ReadFile(pwfile)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
			return
		}
		return
	}
	return hash(pw), dir, nil
}

func hash(p []byte) string {
	algorithm := sha1.New()
	algorithm.Write(p)
	return hex.EncodeToString(algorithm.Sum(nil))
}

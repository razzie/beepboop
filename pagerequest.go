package beepboop

import (
	"net/http"
)

// PageRequest ...
type PageRequest struct {
	Request  *http.Request
	RelPath  string
	RelURI   string
	Title    string
	renderer LayoutRenderer
}

// Respond returns the default page response View
func (r *PageRequest) Respond(data interface{}, opts ...ViewOption) *View {
	v := &View{
		StatusCode: http.StatusOK,
		Data:       data,
	}
	for _, opt := range opts {
		opt(v)
	}
	v.renderer = func(w http.ResponseWriter) {
		r.renderer(w, r.Request, r.Title, data, v.StatusCode)
	}
	return v
}

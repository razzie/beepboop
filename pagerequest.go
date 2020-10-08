package beepboop

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/mo7zayed/reqip"
	"github.com/mssola/user_agent"
)

// PageRequest ...
type PageRequest struct {
	Request   *http.Request
	RequestID string
	RelPath   string
	RelURI    string
	Title     string
	renderer  LayoutRenderer
	logger    *log.Logger
}

func (r *PageRequest) logRequest() {
	ip := reqip.GetClientIP(r.Request)
	hostnames, _ := net.LookupAddr(ip)
	ua := user_agent.New(r.Request.UserAgent())
	browser, ver := ua.Browser()

	r.logger.Printf("New request [%s]: %s\n - IP: %s\n - hostnames: %s\n - browser: %s",
		r.RequestID,
		r.Request.URL.Path,
		ip,
		strings.Join(hostnames, ", "),
		fmt.Sprintf("%s %s %s", ua.OS(), browser, ver))
}

// Log ...
func (r *PageRequest) Log(a ...interface{}) {
	prefix := fmt.Sprintf("[%s] ", r.RequestID)
	r.logger.Output(2, prefix+fmt.Sprint(a...))
}

// Logf ...
func (r *PageRequest) Logf(format string, a ...interface{}) {
	prefix := fmt.Sprintf("[%s] ", r.RequestID)
	r.logger.Output(2, prefix+fmt.Sprintf(format, a...))
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

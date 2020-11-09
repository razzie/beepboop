package beepboop

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/mssola/user_agent"
	"github.com/razzie/babble"
	"github.com/razzie/reqip"
)

// PageRequest ...
type PageRequest struct {
	Context   *Context
	Request   *http.Request
	RequestID string
	RelPath   string
	RelURI    string
	IsAPI     bool
	Title     string
	renderer  LayoutRenderer
	logged    bool
	token     *AccessToken
	access    AccessMap
}

func newPageRequest(page *Page, r *http.Request, ctx *Context, renderer LayoutRenderer) *PageRequest {
	return &PageRequest{
		Context:   ctx,
		Request:   r,
		RequestID: newRequestID(),
		RelPath:   strings.TrimPrefix(r.URL.Path, page.Path),
		RelURI:    strings.TrimPrefix(r.RequestURI, page.Path),
		Title:     page.Title,
		renderer:  renderer,
		IsAPI:     renderer == nil,
	}
}

func newRequestID() string {
	i := uint16(time.Now().UnixNano())
	babbler := babble.NewBabbler()
	return fmt.Sprintf("%s-%x", babbler.Babble(), i)
}

func (r *PageRequest) logRequest() {
	ip := reqip.GetClientIP(r.Request)
	ua := user_agent.New(r.Request.UserAgent())
	browser, ver := ua.Browser()

	logmsg := fmt.Sprintf("[%s]: %s %s\n • %s, %s %s %s",
		r.RequestID, r.Request.Method, r.Request.RequestURI,
		ip, ua.OS(), browser, ver)

	var hasLocation bool
	if r.Context.GeoIPClient != nil {
		loc, _ := r.Context.GeoIPClient.GetLocation(context.Background(), ip)
		if loc != nil {
			hasLocation = true
			logmsg += ", " + loc.String()
		}
	}
	if !hasLocation {
		hostnames, _ := net.LookupAddr(ip)
		logmsg += ", " + strings.Join(hostnames, ", ")
	}

	session, _ := r.Request.Cookie("session")
	if session != nil {
		logmsg += ", session: " + session.Value
	}

	r.Context.Logger.Print(logmsg)
	r.logged = true
}

func (r *PageRequest) logRequestNonblocking() {
	r.logged = true
	go r.logRequest()
}

// Log ...
func (r *PageRequest) Log(a ...interface{}) {
	if !r.logged {
		r.logRequestNonblocking()
	}
	prefix := fmt.Sprintf("[%s] ", r.RequestID)
	r.Context.Logger.Output(2, prefix+fmt.Sprint(a...))
}

// Logf ...
func (r *PageRequest) Logf(format string, a ...interface{}) {
	if !r.logged {
		r.logRequestNonblocking()
	}
	prefix := fmt.Sprintf("[%s] ", r.RequestID)
	r.Context.Logger.Output(2, prefix+fmt.Sprintf(format, a...))
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

// AccessToken returns an AccessToken from this page request
func (r *PageRequest) AccessToken() *AccessToken {
	if r.token == nil {
		r.token = newAccessTokenFromRequest(r)
	}
	return r.token
}

// AddAccess permits the requester to access the given resources
func (r *PageRequest) AddAccess(access AccessMap) error {
	token := r.AccessToken()
	if db := r.Context.DB; db != nil {
		if len(token.SessionID) == 0 {
			token.SessionID = r.RequestID
		}
		return db.addSessionAccess(token.SessionID, token.IP, access)
	}
	r.access.Merge(access)
	token.AccessMap.Merge(access)
	return nil
}

// RevokeAccess revokes the requester's access to the given resources
func (r *PageRequest) RevokeAccess(revoke AccessRevokeMap) error {
	token := r.AccessToken()
	if db := r.Context.DB; db != nil && len(token.SessionID) > 0 {
		return db.revokeSessionAccess(token.SessionID, token.IP, revoke)
	}
	r.access.Revoke(revoke, true)
	token.AccessMap.Revoke(revoke, false)
	return nil
}

func (r *PageRequest) updateViewAccess(view *View) {
	if r.token != nil {
		if len(r.token.SessionID) > 0 {
			cookie := r.token.getSessionCookie(r.Context.CookieExpiration)
			view.cookies = append(view.cookies, cookie)
			return
		}
		cookies := r.access.ToCookies(r.Context.CookieExpiration)
		view.cookies = append(view.cookies, cookies...)
	}
}

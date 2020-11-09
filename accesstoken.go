package beepboop

import (
	"net/http"
	"strings"
	"time"

	"github.com/razzie/reqip"
)

// AccessToken ...
type AccessToken struct {
	SessionID string
	IP        string
	AccessMap AccessMap
}

func newAccessTokenFromRequest(r *PageRequest) *AccessToken {
	token := new(AccessToken).fromCookies(r.Request.Cookies())
	token.IP = reqip.GetClientIP(r.Request)
	db := r.Context.DB
	if db != nil && len(token.SessionID) > 0 {
		dbToken, err := db.getAccessToken(token.SessionID, token.IP)
		if err == nil {
			token.AccessMap.Merge(dbToken.AccessMap)
		} else {
			r.Log(err)
		}
	}
	return token
}

func (token *AccessToken) fromCookies(cookies []*http.Cookie) *AccessToken {
	if token.AccessMap == nil {
		token.AccessMap = make(AccessMap)
	}
	for _, c := range cookies {
		if c.Name == "session" {
			token.SessionID = c.Value
			continue
		}
		access := strings.SplitN(c.Name, "-", 2)
		if len(access) < 2 {
			continue
		}
		token.AccessMap.Add(access[0], access[1], c.Value)
	}
	return token
}

func (token *AccessToken) getSessionCookie(expiration time.Duration) *http.Cookie {
	return &http.Cookie{
		Name:    "session",
		Value:   token.SessionID,
		Path:    "/",
		Expires: time.Now().Add(expiration),
	}
}

// ToCookies returns either a single SessionID cookie or a list of cookies
// containing access to the resources in the access token
func (token *AccessToken) ToCookies(expiration time.Duration) []*http.Cookie {
	if len(token.SessionID) > 0 {
		return []*http.Cookie{token.getSessionCookie(expiration)}
	}
	return token.AccessMap.ToCookies(expiration)
}

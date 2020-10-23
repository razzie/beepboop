package beepboop

import (
	"context"
	"log"

	"github.com/razzie/geoip-server/geoip"
	"golang.org/x/time/rate"
)

// Context ...
type Context struct {
	Context     context.Context
	DB          *DB
	Logger      *log.Logger
	GeoIPClient geoip.Client
	Limiters    map[string]*RateLimiter
}

// GetServiceLimiter returns the rate limiter for the given service and IP
func (ctx *Context) GetServiceLimiter(service, ip string) *rate.Limiter {
	if limiter, ok := ctx.Limiters[service]; ok {
		return limiter.Get(ip)
	}
	return nil
}

// ContextGetter ...
type ContextGetter func(context.Context) *Context

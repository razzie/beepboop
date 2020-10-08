package beepboop

import (
	"context"
	"log"
)

// Context ...
type Context struct {
	Context context.Context
	DB      *DB
	Logger  *log.Logger
}

// ContextGetter ...
type ContextGetter func(context.Context) *Context

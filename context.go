package gent

import (
	"net/http"
	"sync"
)

// Context stores details about a request. It does not implement context.Context.
type Context struct {
	cl  Requester
	mtx *sync.RWMutex

	fni int
	fns []func(*Context)

	Request  *http.Request
	Response *http.Response
	Values   map[string]any
	Errors   []error
}

// newRequestContext creates a request context.
func newRequestContext(
	cl Requester,
	req *http.Request,
	fns []func(*Context),
) *Context {
	return &Context{
		cl:      cl,
		mtx:     &sync.RWMutex{},
		fni:     0,
		fns:     fns,
		Request: req,
		Values:  map[string]any{},
	}
}

// Error appends an error to the context.
func (ctx *Context) Error(err error) {
	ctx.Errors = append(ctx.Errors, err)
}

// Lock locks the mutex in the context.
func (ctx *Context) Lock() {
	ctx.mtx.Lock()
}

// Unlock unlocks the mutex in the context.
func (ctx *Context) Unlock() {
	ctx.mtx.Unlock()
}

// Get retrieves some value from the context's store by a key. The
// operation locks the context's mutex for thread safety.
func (ctx *Context) Get(key string) (val any, ok bool) {
	ctx.mtx.RLock()
	val, ok = ctx.Values[key]
	ctx.mtx.RUnlock()
	return val, ok
}

// Set assigns some value to a key in the context's store. The operation
// locks the context's mutex for thread safety.
func (ctx *Context) Set(key string, val any) {
	ctx.mtx.Lock()
	ctx.Values[key] = val
	ctx.mtx.Unlock()
}

// Del removes some value from the context's store by a key. The
// operation locks the context's mutex for thread safety.
func (ctx *Context) Del(key string) {
	ctx.mtx.Lock()
	delete(ctx.Values, key)
	ctx.mtx.Unlock()
}

// Next runs the next middleware function on the context. If there are no more
// functions, it does nothing.
func (ctx *Context) Next() {
	if ctx.fni < len(ctx.fns) {
		ctx.fni++
		ctx.fns[ctx.fni-1](ctx)
		ctx.fni--
	}
}

// do uses the requester to perform the HTTP request and set the response.
func do(ctx *Context) {
	res, err := ctx.cl.Do(ctx.Request)
	if err != nil {
		ctx.Error(err)
	} else {
		ctx.Response = res
	}
}

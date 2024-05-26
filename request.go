package gent

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"sync"
)

// Request stores details about the request
type Request struct {
	ctx    context.Context
	mem    MemoryPool
	client HttpClient

	mtx    *sync.Mutex
	Values map[string]any

	fns []func(context.Context, *Request)
	fni int

	// Before Build
	Format      string
	Method      string
	Body        any
	Marshaler   Marshaler
	Headers     map[string]string
	QueryParams map[string][]string
	PathParams  []string

	// Before Execute
	Endpoint []byte
	Data     []byte
	Request  *http.Request
	Response *http.Response

	Errors []error
}

// newRequest creates a new Request context
func newRequest(
	ctx context.Context,
	mem MemoryPool,
	client HttpClient,
	format string,
	method string,
	body any,
	marshaler Marshaler,
	headers map[string]string,
	queryParams map[string][]string,
	pathParams []string,
	fns []func(context.Context, *Request),
) *Request {
	return &Request{
		ctx:         ctx,
		mem:         mem,
		client:      client,
		mtx:         &sync.Mutex{},
		Values:      map[string]any{},
		fns:         fns,
		fni:         0,
		Format:      format,
		Method:      method,
		Body:        body,
		Marshaler:   marshaler,
		Headers:     headers,
		QueryParams: queryParams,
		PathParams:  pathParams,
	}
}

// Lock locks the mutex within the context
func (r *Request) Lock() {
	r.mtx.Lock()
}

// Unlock unlocks the mutex within the context
func (r *Request) Unlock() {
	r.mtx.Unlock()
}

// Set assigns some value to a key in the context's Values store. The operation
// locks the context's mutex for thread safety.
func (r *Request) Set(key string, val any) {
	r.mtx.Lock()
	r.Values[key] = val
	r.mtx.Unlock()
}

// Get retrieves some value from the context's Values store by a key. The
// operation locks the context's mutex for thread safety.
func (r *Request) Get(key string) (val any, ok bool) {
	r.mtx.Lock()
	val, ok = r.Values[key]
	r.mtx.Unlock()
	return val, ok
}

// Next executes the next function on the function middleware on the request.
// If there are no more functions to call, it does nothing.
func (r *Request) Next() {
	if r.fni < len(r.fns) {
		r.fni++
		r.fns[r.fni-1](r.ctx, r)
		r.fni--
	}
}

// Error inserts an error to the errors slice.
func (r *Request) Error(err error) {
	r.Errors = append(r.Errors, err)
}

// prepare formats the endpoint of the request and marshals the request body if
// there is a body and a marshaler module provided.
func prepare(ctx context.Context, r *Request) {
	var endpoint, data []byte
	var contentType string
	var err error

	// create endpoint string
	endpoint, err = r.fmtEndpoint(r.Format, r.QueryParams, r.PathParams)
	if err != nil {
		r.Error(err)
		return
	}

	// create body content
	if r.Body != nil && r.Marshaler != nil {
		data, contentType, err = r.Marshaler.Marshal(r.Body)
		if err != nil {
			r.Error(err)
			return
		}
	}

	// create request
	body := bytes.NewReader(data)
	req, err := http.NewRequestWithContext(ctx, r.Method, string(endpoint), body)
	if err != nil {
		r.Error(err)
		return
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	for k, v := range r.Headers {
		req.Header.Set(k, v)
	}

	r.Endpoint = endpoint
	r.Data = data
	r.Request = req
	r.Next()
}

// execute creates an HTTP request from the method, url endpoint and content,
// adds the headers to the request and uses the client to do the request.
func execute(ctx context.Context, r *Request) {
	var err error

	r.Response, err = r.client.Do(r.Request)
	if err != nil {
		r.Error(err)
		return
	}
}

// fmtEndpoint formats the endpoint template to add path parameters and
// query parameters
func (r *Request) fmtEndpoint(
	format string,
	queryPrm map[string][]string,
	pathPrm []string,
) ([]byte, error) {
	wrt := newWrirter(r.mem)
	defer wrt.release()

	lst, cur, end := 0, 0, len(format)
	pathIdx := 0

	for cur < end {
		if format[cur] == '{' {
			if cur+1 == end || format[cur+1] != '}' {
				return nil, fmt.Errorf("illegal character/Invalid format in url")
			} else if pathIdx >= len(pathPrm) {
				return nil, fmt.Errorf("not enough parameters provided")
			} else {
				wrt.writeString(format[lst:cur])
				wrt.writeEscaped(pathPrm[pathIdx])
				pathIdx++
				cur += 2
				lst = cur
			}
		} else {
			cur++
		}
	}

	if pathIdx != len(pathPrm) {
		return nil, fmt.Errorf("too many parameters provided")
	}
	if lst < cur {
		wrt.writeString(format[lst:cur])
	}

	if len(queryPrm) > 0 {
		wrt.writeByte('?')
		first := true
		for k, vs := range queryPrm {
			for _, v := range vs {
				if !first {
					wrt.writeByte('&')
				}
				wrt.writeEscaped(k)
				wrt.writeByte('=')
				wrt.writeEscaped(v)
				first = false
			}
		}
	}

	return wrt.buf.build(nil), nil
}

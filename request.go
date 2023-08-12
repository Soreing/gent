package gent

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
)

// Request stores details about the request
type Request struct {
	ctx    context.Context
	mem    MemoryPool
	retr   Retrier
	client HttpClient
	errors []error

	fns []func(context.Context, *Request)
	fni int

	format      string
	method      string
	body        any
	marshaler   Marshaler
	headers     map[string]string
	queryParams map[string][]string
	pathParams  []string

	endpoint []byte
	data     []byte

	response *http.Response
}

// newRequest creates a new Request context
func newRequest(
	ctx context.Context,
	mem MemoryPool,
	retr Retrier,
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
		retr:        retr,
		client:      client,
		fns:         fns,
		fni:         0,
		format:      format,
		method:      method,
		body:        body,
		marshaler:   marshaler,
		headers:     headers,
		queryParams: queryParams,
		pathParams:  pathParams,
	}
}

// Next executes the next function on the function middleware on the request.
// If there are no more functions to call, it does nothing.
func (r *Request) Next() {
	if r.fni < len(r.fns) {
		r.fni++
		r.fns[r.fni-1](r.ctx, r)
	}
}

// Error inserts an error to the errors slice.
func (r *Request) Error(err error) {
	r.errors = append(r.errors, err)
}

// Errors retruns all the errors as a slice.
func (r *Request) Errors() []error {
	return r.errors
}

// prepare formats the endpoint of the request and marshals the request body if
// there is a body and a marshaller module provided.
func prepare(ctx context.Context, r *Request) {
	var err error

	r.endpoint, err = r.fmtEndpoint(
		r.format,
		r.queryParams,
		r.pathParams,
	)
	if err != nil {
		r.Error(err)
		return
	}

	if r.body != nil {
		if r.marshaler != nil {
			var ctype string
			r.data, ctype, err = r.marshaler.Marshal(r.body)
			if err != nil {
				r.Error(err)
				return
			}
			if _, ok := r.headers["Content-Type"]; !ok {
				r.headers["Content-Type"] = ctype
			}
		} else {
			r.Error(fmt.Errorf("marshaller is nil"))
			return
		}
	}

	r.Next()
}

// execute creates an HTTP request from the method, url endpoint and content,
// adds the headers to the request and uses the client to do the request.
func execute(ctx context.Context, r *Request) {
	req, err := http.NewRequestWithContext(
		ctx,
		r.method,
		string(r.endpoint),
		bytes.NewReader(r.data),
	)
	if err != nil {
		r.Error(err)
		return
	}

	for k, v := range r.headers {
		req.Header.Add(k, v)
	}

	r.response, err = r.client.Do(req)
	if err != nil {
		r.Error(err)
		return
	}
}

// executeWithRetrier executes the request in the context of a retrier. The
// request will be retried until the retrier's ShouldRetry function returns
// false.
func executeWithRetrier(ctx context.Context, r *Request) {
	err := r.retr.Run(ctx, func(ctx context.Context) (error, bool) {
		req, err := http.NewRequestWithContext(
			ctx,
			r.method,
			string(r.endpoint),
			bytes.NewReader(r.data),
		)
		if err != nil {
			return err, false
		}

		for k, v := range r.headers {
			req.Header.Add(k, v)
		}

		r.response, err = r.client.Do(req)
		return r.retr.ShouldRetry(r.response, err)
	})

	if err != nil {
		r.Error(err)
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
		begIdx := wrt.buf.len()
		for k, vs := range queryPrm {
			for _, v := range vs {
				if len(wrt.buf.page) > begIdx {
					wrt.writeByte('&')
				}
				wrt.writeEscaped(k)
				wrt.writeByte('=')
				wrt.writeEscaped(v)
			}
		}
	}

	return wrt.buf.build(nil), nil
}

func (r *Request) GetMethod() string {
	return r.method
}

func (r *Request) GetHeader(k string) (v string, ok bool) {
	v, ok = r.headers[k]
	return
}

func (r *Request) GetQueryParam(k string) (v []string, ok bool) {
	v, ok = r.queryParams[k]
	return
}

func (r *Request) GetEndpoint() []byte {
	return r.endpoint
}

func (r *Request) GetData() []byte {
	return r.data
}

func (r *Request) GetResponse() *http.Response {
	return r.response
}

func (r *Request) AddHeader(k, v string) {
	r.headers[k] = v
}

func (r *Request) DelHeader(k string) {
	delete(r.headers, k)
}

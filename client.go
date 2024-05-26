package gent

import (
	"context"
	"fmt"
	"net/http"
)

// MiddlewareStage enumerates the stage in the request where middlewares are
// attached. Middlewares can be added before the http.Request object is built,
// or before the http.Request object is sent.
type MiddlewareStage int

const (
	MDW_BeforeBuild MiddlewareStage = iota
	MDW_BeforeExecute
)

// Client wraps an http.Client with additional functionality.
type Client struct {
	mem    MemoryPool
	client HttpClient
	constr func() HttpClient
	l2mdw  []func(context.Context, *Request)
	l1mdw  []func(context.Context, *Request)
}

// NewClient creates a Client with options.
func NewClient(opts ...Option) *Client {
	cfg := newConfiguration(opts)

	return &Client{
		mem:    cfg.mempool,
		client: cfg.httpClient,
		constr: cfg.newClientFn,
		l2mdw:  []func(context.Context, *Request){},
		l1mdw:  []func(context.Context, *Request){},
	}
}

// Use attaches a middleware to the client's execution chain. Middlewares
// added before build run before the http.Request object is created from the
// http method, endpoint, headers, path query parameters and body. Middlewares
// added before execute run before the http.Request obeject is sent.
func (c *Client) Use(
	statge MiddlewareStage,
	middlewares ...func(context.Context, *Request),
) error {
	switch statge {
	case MDW_BeforeBuild:
		c.l2mdw = append(c.l2mdw, middlewares...)
	case MDW_BeforeExecute:
		c.l1mdw = append(c.l1mdw, middlewares...)
	default:
		return fmt.Errorf("invalid middleware stage")
	}
	return nil
}

// getClientForRequest returns an internal client to make a request with.
func (c *Client) getClientForRequest() HttpClient {
	if c.constr != nil {
		return c.constr()
	} else {
		return c.client
	}
}

// Do runs an http request with all the given parameters.
func (c *Client) Do(
	ctx context.Context,
	method string,
	endpoint string,
	body any,
	marshaler Marshaler,
	headers map[string]string,
	queryParam map[string][]string,
	pathParams ...string,
) (res *http.Response, err error) {

	fns := make(
		[]func(context.Context, *Request),
		0, len(c.l1mdw)+len(c.l2mdw)+2,
	)

	fns = append(fns, c.l2mdw...)
	fns = append(fns, prepare)
	fns = append(fns, c.l1mdw...)
	fns = append(fns, execute)

	cl := c.getClientForRequest()
	req := newRequest(
		ctx, c.mem, cl, endpoint, method,
		body, marshaler, headers,
		queryParam, pathParams,
		fns,
	)

	req.Next()

	if len(req.Errors) > 0 {
		return req.Response, req.Errors[0]
	} else {
		return req.Response, nil
	}
}

// Get runs an http GET request with all the given parameters.
func (c *Client) Get(
	ctx context.Context,
	endpoint string,
	body any,
	marshaler Marshaler,
	headers map[string]string,
	queryParam map[string][]string,
	pathParams ...string,
) (res *http.Response, err error) {
	return c.Do(
		ctx, http.MethodGet, endpoint,
		body, marshaler,
		headers, queryParam, pathParams...,
	)
}

// Post runs an http POST request with all the given parameters.
func (c *Client) Post(
	ctx context.Context,
	endpoint string,
	body any,
	marshaler Marshaler,
	headers map[string]string,
	queryParam map[string][]string,
	pathParams ...string,
) (res *http.Response, err error) {
	return c.Do(
		ctx, http.MethodPost, endpoint,
		body, marshaler,
		headers, queryParam, pathParams...,
	)
}

// Patch runs an http PATCH request with all the given parameters.
func (c *Client) Patch(
	ctx context.Context,
	endpoint string,
	body any,
	marshaler Marshaler,
	headers map[string]string,
	queryParam map[string][]string,
	pathParams ...string,
) (res *http.Response, err error) {
	return c.Do(
		ctx, http.MethodPatch, endpoint,
		body, marshaler,
		headers, queryParam, pathParams...,
	)
}

// Put runs an http PUT request with all the given parameters.
func (c *Client) Put(
	ctx context.Context,
	endpoint string,
	body any,
	marshaler Marshaler,
	headers map[string]string,
	queryParam map[string][]string,
	pathParams ...string,
) (res *http.Response, err error) {
	return c.Do(
		ctx, http.MethodPut, endpoint,
		body, marshaler,
		headers, queryParam, pathParams...,
	)
}

// Delete runs an http DELETE request with all the given parameters.
func (c *Client) Delete(
	ctx context.Context,
	endpoint string,
	body any,
	marshaler Marshaler,
	headers map[string]string,
	queryParam map[string][]string,
	pathParams ...string,
) (res *http.Response, err error) {
	return c.Do(
		ctx, http.MethodDelete, endpoint,
		body, marshaler,
		headers, queryParam, pathParams...,
	)
}

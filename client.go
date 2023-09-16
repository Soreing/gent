package gent

import (
	"context"
	"net/http"
)

// Client wraps an http.Client with additional functionality.
type Client struct {
	mem       MemoryPool
	retr      Retrier
	client    HttpClient
	constr    func() HttpClient
	functions []func(context.Context, *Request)
}

// NewClient creates a Client with options.
func NewClient(opts ...Option) *Client {
	cfg := newConfiguration(opts)

	functs := make(
		[]func(context.Context, *Request),
		0, len(cfg.hlmiddlewares)+1+len(cfg.llmiddlewares)+1,
	)

	functs = append(functs, cfg.hlmiddlewares...)
	functs = append(functs, prepare)
	functs = append(functs, cfg.llmiddlewares...)

	if cfg.retrier != nil {
		functs = append(functs, executeWithRetrier)
	} else {
		functs = append(functs, execute)
	}

	cl := &Client{
		mem:       cfg.mempool,
		retr:      cfg.retrier,
		client:    cfg.httpClient,
		constr:    cfg.newClientFn,
		functions: functs,
	}

	return cl
}

// getClientForRequest returns an internal client to make a request with
func (c *Client) getClientForRequest() HttpClient {
	if c.constr != nil {
		return c.constr()
	} else {
		return c.client
	}
}

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

	cl := c.getClientForRequest()
	req := newRequest(
		ctx, c.mem, c.retr, cl, endpoint, method,
		body, marshaler, headers,
		queryParam, pathParams,
		c.functions,
	)

	req.Next()

	res = req.GetResponse()
	errs := req.Errors()
	if len(errs) > 0 {
		return res, errs[0]
	} else {
		return res, nil
	}
}

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
		ctx, "GET", endpoint,
		body, marshaler,
		headers, queryParam, pathParams...,
	)
}

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
		ctx, "POST", endpoint,
		body, marshaler,
		headers, queryParam, pathParams...,
	)
}

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
		ctx, "PATCH", endpoint,
		body, marshaler,
		headers, queryParam, pathParams...,
	)
}

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
		ctx, "PUT", endpoint,
		body, marshaler,
		headers, queryParam, pathParams...,
	)
}

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
		ctx, "DELETE", endpoint,
		body, marshaler,
		headers, queryParam, pathParams...,
	)
}

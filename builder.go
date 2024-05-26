package gent

import (
	"context"
	"net/http"
)

// RequestBuilder allows gradual creation of http requests with functions to
// attach a body, headers, query parameters and path parameters.
type RequestBuilder struct {
	client      *Client
	method      string
	endpoint    string
	body        any
	marshaler   Marshaler
	headers     map[string]string
	queryParams map[string][]string
	pathParams  []string
}

// NewRequest creates a request builder.
func (c *Client) NewRequest(
	method string,
	endpoint string,
) *RequestBuilder {
	return &RequestBuilder{
		client:      c,
		method:      method,
		endpoint:    endpoint,
		body:        nil,
		marshaler:   nil,
		headers:     map[string]string{},
		queryParams: map[string][]string{},
		pathParams:  []string{},
	}
}

// WithBody adds a request body and a marshaler to the request builder. If the
// request builder already has a request body and marshaler, they get overwritten.
func (rb *RequestBuilder) WithBody(
	body any,
	marshaler Marshaler,
) *RequestBuilder {
	rb.body = body
	rb.marshaler = marshaler
	return rb
}

// WithHeader adds a header to the request builder. If the key was already
// assigned some value, it gets overwritten.
func (rb *RequestBuilder) WithHeader(
	key string,
	val string,
) *RequestBuilder {
	rb.headers[key] = val
	return rb
}

// WithHeaders adds a list of headers to the request builder. If any of the keys
// was already assigned some value, it gets overwritten.
func (rb *RequestBuilder) WithHeaders(
	headers map[string]string,
) *RequestBuilder {
	for k, v := range headers {
		rb.headers[k] = v
	}
	return rb
}

// WithQueryParameter adds a query parameter to the request builder. If the key
// was already assigned some value, it gets overwritten.
func (rb *RequestBuilder) WithQueryParameter(
	key string,
	val []string,
) *RequestBuilder {
	rb.queryParams[key] = val
	return rb
}

// WithQueryParameters adds a list of query parameters to the request builder.
// If any of the keys was already assigned some value, it gets overwritten.
func (rb *RequestBuilder) WithQueryParameters(
	queryParams map[string][]string,
) *RequestBuilder {
	for k, v := range queryParams {
		rb.queryParams[k] = v
	}
	return rb
}

// WithPathParameter adds a path parameter to the request builder. Each
// parameter gets appended to the list in the builder.
func (rb *RequestBuilder) WithPathParameters(
	pathParams ...string,
) *RequestBuilder {
	rb.pathParams = append(rb.pathParams, pathParams...)
	return rb
}

// Run runs the request with the client that created the request builder.
func (rb *RequestBuilder) Run(
	ctx context.Context,
) (res *http.Response, err error) {
	return rb.client.Do(
		ctx,
		rb.method, rb.endpoint,
		rb.body, rb.marshaler,
		rb.headers,
		rb.queryParams,
		rb.pathParams...,
	)
}

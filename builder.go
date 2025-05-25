package gent

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/url"
)

// ErrInvalidBodyType is returned by Marshaler functions when the object
// can not be marshaled into a byte array due to its type, or the RequestBuilder
// when the body requires a marshaler but none is provided
var ErrInvalidBodyType = errors.New("invalid body type")

// ErrInvalidFormat is returned by RequestBuilder.Build when the format has a
// trailing or incomplate placeholder {}, or if the number of placeholders does
// not match the number of path parameters
var ErrInvalidFormat = errors.New("invalid endpoint format")

// RequestBuilder allows gradual creation of http requests with functions to
// attach a body, headers, query parameters and path parameters.
type RequestBuilder struct {
	method    string
	format    string
	body      any
	marshaler Marshaler
	headers   map[string][]string
	queryPrms map[string][]string
	pathPrms  []string
}

// NewRequest creates a request builder.
func NewRequest(
	method string,
	format string,
) *RequestBuilder {
	return &RequestBuilder{
		method: method,
		format: format,
	}
}

// WithRawBody sets a byte array as the request body. If a body or marshaler
// is already set, it will overwrite it.
func (rb *RequestBuilder) WithRawBody(
	body []byte,
) *RequestBuilder {
	rb.marshaler = nil
	rb.body = body
	return rb
}

// WithBody adds a body and a marshaler to the request. If a body or marshaler
// is already set, it will overwrite it. The headers returned by the marshaler
// will not overwrite headers set by [WithHeader] or [WithHeaders]
func (rb *RequestBuilder) WithBody(
	body any,
	marshaler Marshaler,
) *RequestBuilder {
	rb.body = body
	rb.marshaler = marshaler
	return rb
}

// WithHeader adds a header to the request. If there was already a header set
// with the same key, it will overwrite it.
func (rb *RequestBuilder) WithHeader(
	key string,
	val string,
) *RequestBuilder {
	if rb.headers == nil {
		rb.headers = map[string][]string{}
	}
	rb.headers[key] = append(rb.headers[key], val)
	return rb
}

// WithQueryParameter adds a query parameter to the request. If there was
// already a parameter set with the same key, it will overwrite it.
func (rb *RequestBuilder) WithQueryParameter(
	key string,
	vals []string,
) *RequestBuilder {
	if rb.queryPrms == nil {
		rb.queryPrms = map[string][]string{}
	}
	rb.queryPrms[key] = append(rb.queryPrms[key], vals...)
	return rb
}

// WithPathParameter adds path parameters to the request. The parameters get
// escaped and appended to the list in the request builder. Path parameters
// replace {} placeholders in the request endpoint in the order they were added.
func (rb *RequestBuilder) WithPathParameters(
	params ...string,
) *RequestBuilder {
	slc := make([]string, 0, len(rb.pathPrms)+len(params))
	slc = append(slc, rb.pathPrms...)
	for _, param := range params {
		slc = append(slc, url.PathEscape(param))
	}

	rb.pathPrms = slc
	return rb
}

// Build returns a *http.Request from the values of the request builder.
func (rb *RequestBuilder) Build(
	ctx context.Context,
) (res *http.Request, err error) {
	buflen := len(rb.format)
	for _, param := range rb.pathPrms {
		buflen += len(param) - 2
	}

	// create request endpoint
	endp := make([]byte, 0, buflen)
	open, cursor, pidx := false, 0, 0
	for i, ch := range rb.format {
		if (open && ch != '}') || (!open && ch == '}') {
			return nil, ErrInvalidFormat
		} else if ch == '{' && pidx == len(rb.pathPrms) {
			return nil, ErrInvalidFormat
		} else if ch == '{' {
			open = true
		} else if ch == '}' {
			open = false
			endp = append(endp, rb.format[cursor:i-1]...)
			endp = append(endp, rb.pathPrms[pidx]...)
			cursor = i + 1
			pidx++
		}
	}
	if open || pidx != len(rb.pathPrms) {
		return nil, ErrInvalidFormat
	}
	endp = append(endp, rb.format[cursor:]...)

	// create body content
	var body []byte
	var bodyHdrs map[string][]string
	if rb.marshaler != nil {
		body, bodyHdrs, err = rb.marshaler(rb.body)
		if err != nil {
			return nil, err
		}
	} else if bytes, ok := rb.body.([]byte); bytes != nil && ok {
		body = bytes
	} else if rb.body != nil {
		return nil, ErrInvalidBodyType
	}

	// create request
	reader := bytes.NewReader(body)
	req, err := http.NewRequestWithContext(ctx, rb.method, string(endp), reader)
	if err != nil {
		return nil, err
	}

	// set query params
	if req.URL.RawQuery == "" {
		req.URL.RawQuery = url.Values(rb.queryPrms).Encode()
	} else {
		q := req.URL.Query()
		for k, v := range rb.queryPrms {
			q[k] = append(q[k], v...)
		}
		req.URL.RawQuery = q.Encode()
	}

	// set headers
	for key, vals := range rb.headers {
		for _, val := range vals {
			req.Header.Add(key, val)
		}
	}
	for key, vals := range bodyHdrs {
		for _, val := range vals {
			req.Header.Add(key, val)
		}
	}

	return req, nil
}

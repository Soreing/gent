package gent

import (
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Requester defines an HTTP client that can do requests.
type Requester interface {
	Do(r *http.Request) (*http.Response, error)
	CloseIdleConnections()
}

// Client wraps an http Client with additional features.
type Client struct {
	cl   Requester
	mdws []func(*Context)
}

// NewDefaultClient creates a Client from http.DefaultClient.
func NewDefaultClient() *Client {
	return &Client{cl: http.DefaultClient}
}

// NewClient creates a Client from the provided Requester.
func NewClient(client Requester) *Client {
	return &Client{cl: client}
}

// Use adds a middleware style handler function to the execution chain of
// the requests performed by the client which run in the order they were added
// before the client performs the request.
func (c *Client) Use(
	middlewares ...func(*Context),
) {
	c.mdws = append(c.mdws, middlewares...)
}

// Do sends an HTTP request and returns an HTTP response.
func (c *Client) Do(
	req *http.Request,
) (res *http.Response, err error) {
	fns := make([]func(*Context), 0, len(c.mdws)+1)
	fns = append(fns, c.mdws...)
	fns = append(fns, do)

	ctx := newRequestContext(c.cl, req, fns)
	ctx.Next()

	if len(ctx.Errors) > 0 {
		return ctx.Response, ctx.Errors[0]
	}
	return ctx.Response, nil
}

// Get sends a GET HTTP request to the specified URL.
func (c *Client) Get(
	url string,
) (res *http.Response, err error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	return c.Do(req)
}

// Head sends a HEAD HTTP request to the specified URL.
func (c *Client) Head(
	url string,
) (res *http.Response, err error) {
	req, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		return nil, err
	}

	return c.Do(req)
}

// Post sends a POST HTTP request to the specified URL with a content type
// header and a request body.
func (c *Client) Post(
	url string,
	contentType string,
	body io.Reader,
) (res *http.Response, err error) {
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)
	return c.Do(req)
}

// PostForm sends a POST HTTP request to the specified URL with a content type
// header of application/x-www-form-urlencoded and url encoded values as
// the request body.
func (c *Client) PostForm(
	url string,
	data url.Values,
) (res *http.Response, err error) {
	body := strings.NewReader(data.Encode())
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return c.Do(req)
}

// CloseIdleConnections closes idle connections on the underlying Requester.
func (c *Client) CloseIdleConnections() {
	c.cl.CloseIdleConnections()
}

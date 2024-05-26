package gent

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewRequestBuilder tests that a request builder can be created.
func TestNewRequestBuilder(t *testing.T) {
	tests := []struct {
		Name     string
		Client   *Client
		Method   string
		Endpoint string
	}{
		{
			Name:     "Creating new request builder",
			Client:   NewClient(),
			Method:   http.MethodGet,
			Endpoint: "http://localhost:8080",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			req := test.Client.NewRequest(test.Method, test.Endpoint)

			assert.NotNil(t, req.client)
			assert.Equal(t, test.Method, req.method)
			assert.Equal(t, test.Endpoint, req.endpoint)
			assert.Nil(t, req.body)
			assert.Nil(t, req.marshaler)
			assert.NotNil(t, req.headers)
			assert.NotNil(t, req.queryParams)
			assert.NotNil(t, req.pathParams)
		})
	}
}

// TestRequestWithBody tests adding a body and marshaler to the request builder.
func TestRequestWithBody(t *testing.T) {
	tests := []struct {
		Name      string
		Client    *Client
		Method    string
		Endpoint  string
		Body      any
		Marshaler Marshaler
	}{
		{
			Name:      "Request with body",
			Client:    NewClient(),
			Method:    http.MethodGet,
			Endpoint:  "http://localhost:8080",
			Body:      "{\"key\":\"value\"}",
			Marshaler: NewJSONMarshaler(),
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			req := test.Client.NewRequest(test.Method, test.Endpoint)

			req = req.WithBody(test.Body, test.Marshaler)

			assert.Equal(t, test.Body, req.body)
			assert.Equal(t, test.Marshaler, req.marshaler)
		})
	}
}

// TestRequestWithHeader tests adding a header to the request builder.
func TestRequestWithHeader(t *testing.T) {
	tests := []struct {
		Name     string
		Client   *Client
		Method   string
		Endpoint string
		Before   map[string]string
		Key      string
		Value    string
		After    map[string]string
	}{
		{
			Name:     "Adding header to request without headers",
			Client:   NewClient(),
			Method:   http.MethodGet,
			Endpoint: "http://localhost:8080",
			Before:   map[string]string{},
			Key:      "Authorization",
			Value:    "x.y.z",
			After: map[string]string{
				"Authorization": "x.y.z",
			},
		},
		{
			Name:     "Adding header to request with headers",
			Client:   NewClient(),
			Method:   http.MethodGet,
			Endpoint: "http://localhost:8080",
			Before: map[string]string{
				"X-Api-Key": "a1g2Q4GlqCXXcqbZ",
			},
			Key:   "Authorization",
			Value: "x.y.z",
			After: map[string]string{
				"X-Api-Key":     "a1g2Q4GlqCXXcqbZ",
				"Authorization": "x.y.z",
			},
		},
		{
			Name:     "Adding header to request with same key",
			Client:   NewClient(),
			Method:   http.MethodGet,
			Endpoint: "http://localhost:8080",
			Before: map[string]string{
				"X-Api-Key":     "a1g2Q4GlqCXXcqbZ",
				"Authorization": "a.b.c",
			},
			Key:   "Authorization",
			Value: "x.y.z",
			After: map[string]string{
				"X-Api-Key":     "a1g2Q4GlqCXXcqbZ",
				"Authorization": "x.y.z",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			req := test.Client.NewRequest(test.Method, test.Endpoint)
			req.headers = test.Before

			req = req.WithHeader(test.Key, test.Value)

			assert.Equal(t, test.After, req.headers)
		})
	}
}

// TestRequestWithHeaders tests adding headers to the request builder.
func TestRequestWithHeaders(t *testing.T) {
	tests := []struct {
		Name     string
		Client   *Client
		Method   string
		Endpoint string
		Before   map[string]string
		Headers  map[string]string
		After    map[string]string
	}{
		{
			Name:     "Adding headers to request without headers",
			Client:   NewClient(),
			Method:   http.MethodGet,
			Endpoint: "http://localhost:8080",
			Before:   map[string]string{},
			Headers: map[string]string{
				"Authorization": "x.y.z",
				"Date":          "Wed, 21 Oct 2015 07:28:00 GMT",
			},
			After: map[string]string{
				"Authorization": "x.y.z",
				"Date":          "Wed, 21 Oct 2015 07:28:00 GMT",
			},
		},
		{
			Name:     "Adding headers to request with headers",
			Client:   NewClient(),
			Method:   http.MethodGet,
			Endpoint: "http://localhost:8080",
			Before: map[string]string{
				"X-Api-Key": "a1g2Q4GlqCXXcqbZ",
			},
			Headers: map[string]string{
				"Authorization": "x.y.z",
				"Date":          "Wed, 21 Oct 2015 07:28:00 GMT",
			},
			After: map[string]string{
				"X-Api-Key":     "a1g2Q4GlqCXXcqbZ",
				"Authorization": "x.y.z",
				"Date":          "Wed, 21 Oct 2015 07:28:00 GMT",
			},
		},
		{
			Name:     "Adding headers to request with same keys",
			Client:   NewClient(),
			Method:   http.MethodGet,
			Endpoint: "http://localhost:8080",
			Before: map[string]string{
				"X-Api-Key":     "a1g2Q4GlqCXXcqbZ",
				"Authorization": "a.b.c",
				"Date":          "Tue, 04 Nov 2017 01:44:15 GMT",
			},
			Headers: map[string]string{
				"Authorization": "x.y.z",
				"Date":          "Wed, 21 Oct 2015 07:28:00 GMT",
			},
			After: map[string]string{
				"X-Api-Key":     "a1g2Q4GlqCXXcqbZ",
				"Authorization": "x.y.z",
				"Date":          "Wed, 21 Oct 2015 07:28:00 GMT",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			req := test.Client.NewRequest(test.Method, test.Endpoint)
			req.headers = test.Before

			req = req.WithHeaders(test.Headers)

			assert.Equal(t, test.After, req.headers)
		})
	}
}

// TestRequestWithQueryParameters tests adding a query parameter to the request builder.
func TestRequestWithQueryParameter(t *testing.T) {
	tests := []struct {
		Name     string
		Client   *Client
		Method   string
		Endpoint string
		Before   map[string][]string
		Key      string
		Value    []string
		After    map[string][]string
	}{
		{
			Name:     "Adding query param to request without params",
			Client:   NewClient(),
			Method:   http.MethodGet,
			Endpoint: "http://localhost:8080",
			Before:   map[string][]string{},
			Key:      "page",
			Value:    []string{"0"},
			After: map[string][]string{
				"page": {"0"},
			},
		},
		{
			Name:     "Adding query param to request with params",
			Client:   NewClient(),
			Method:   http.MethodGet,
			Endpoint: "http://localhost:8080",
			Before: map[string][]string{
				"items": {"100"},
			},
			Key:   "page",
			Value: []string{"0"},
			After: map[string][]string{
				"page":  {"0"},
				"items": {"100"},
			},
		},
		{
			Name:     "Adding query param to request with same key",
			Client:   NewClient(),
			Method:   http.MethodGet,
			Endpoint: "http://localhost:8080",
			Before: map[string][]string{
				"page":  {"1"},
				"items": {"100"},
			},
			Key:   "page",
			Value: []string{"0"},
			After: map[string][]string{
				"page":  {"0"},
				"items": {"100"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			req := test.Client.NewRequest(test.Method, test.Endpoint)
			req.queryParams = test.Before

			req = req.WithQueryParameter(test.Key, test.Value)

			assert.Equal(t, test.After, req.queryParams)
		})
	}
}

// TestRequestWithQueryParameters tests adding query parameters to the request builder.
func TestRequestWithQueryParameters(t *testing.T) {
	tests := []struct {
		Name     string
		Client   *Client
		Method   string
		Endpoint string
		Before   map[string][]string
		Params   map[string][]string
		After    map[string][]string
	}{
		{
			Name:     "Adding query params to request without params",
			Client:   NewClient(),
			Method:   http.MethodGet,
			Endpoint: "http://localhost:8080",
			Before:   map[string][]string{},
			Params: map[string][]string{
				"page":    {"0"},
				"orderby": {"id"},
			},
			After: map[string][]string{
				"page":    {"0"},
				"orderby": {"id"},
			},
		},
		{
			Name:     "Adding query params to request with params",
			Client:   NewClient(),
			Method:   http.MethodGet,
			Endpoint: "http://localhost:8080",
			Before: map[string][]string{
				"items": {"100"},
			},
			Params: map[string][]string{
				"page":    {"0"},
				"orderby": {"id"},
			},
			After: map[string][]string{
				"page":    {"0"},
				"orderby": {"id"},
				"items":   {"100"},
			},
		},
		{
			Name:     "Adding query params to request with same key",
			Client:   NewClient(),
			Method:   http.MethodGet,
			Endpoint: "http://localhost:8080",
			Before: map[string][]string{
				"page":    {"1"},
				"orderby": {"time"},
				"items":   {"100"},
			},
			Params: map[string][]string{
				"page":    {"0"},
				"orderby": {"id"},
			},
			After: map[string][]string{
				"page":    {"0"},
				"orderby": {"id"},
				"items":   {"100"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			req := test.Client.NewRequest(test.Method, test.Endpoint)
			req.queryParams = test.Before

			req = req.WithQueryParameters(test.Params)

			assert.Equal(t, test.After, req.queryParams)
		})
	}
}

// TestRequestWithPathParameters tests adding path parameters to the request builder.
func TestRequestWithPathParameters(t *testing.T) {
	tests := []struct {
		Name     string
		Client   *Client
		Method   string
		Endpoint string
		Before   []string
		Params   []string
		After    []string
	}{
		{
			Name:     "Adding path params to request without params",
			Client:   NewClient(),
			Method:   http.MethodGet,
			Endpoint: "http://localhost:8080",
			Before:   []string{},
			Params:   []string{"rcD3yVsj", "DiBdvVeU"},
			After:    []string{"rcD3yVsj", "DiBdvVeU"},
		},
		{
			Name:     "Adding query params to request with params",
			Client:   NewClient(),
			Method:   http.MethodGet,
			Endpoint: "http://localhost:8080",
			Before:   []string{"rnp7cd0w"},
			Params:   []string{"rcD3yVsj", "DiBdvVeU"},
			After:    []string{"rnp7cd0w", "rcD3yVsj", "DiBdvVeU"},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			req := test.Client.NewRequest(test.Method, test.Endpoint)
			req.pathParams = test.Before

			req = req.WithPathParameters(test.Params...)

			assert.Equal(t, test.After, req.pathParams)
		})
	}
}

// TestRunRequest tests running a request with the request builder.
func TestRunRequest(t *testing.T) {
	tests := []struct {
		Name        string
		HttpClient  *mockHttpHandler
		Method      string
		Format      string
		Body        any
		Marshaler   Marshaler
		Headers     map[string]string
		QueryParams map[string][]string
		PathParams  []string

		Data       []byte
		Endpoint   []byte
		StatusCode int
		Error      error
	}{
		{
			Name:        "Successful request",
			Method:      "POST",
			Format:      "http://localhost:8080/{}",
			Body:        map[string]any{"test": "test"},
			Marshaler:   NewJSONMarshaler(),
			Headers:     map[string]string{"Authorization": "Bearer x.y.z"},
			QueryParams: map[string][]string{"query": {"query"}},
			PathParams:  []string{"param"},
			Data:        []byte(`{"test":"test"}`),
			Endpoint:    []byte("http://localhost:8080/param?query=query"),
			StatusCode:  201,
			Error:       nil,

			HttpClient: &mockHttpHandler{
				code:    201,
				err:     nil,
				headers: map[string]string{},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			cl := NewClient(UseHttpClient(test.HttpClient))

			req := cl.NewRequest(
				test.Method, test.Format,
			).WithBody(
				test.Body, test.Marshaler,
			).WithHeaders(
				test.Headers,
			).WithQueryParameters(
				test.QueryParams,
			).WithPathParameters(
				test.PathParams...,
			)

			res, err := req.Run(context.Background())

			assert.Equal(t, test.Error, err)
			assert.Equal(t, test.Method, test.HttpClient.method)
			assert.Equal(t, test.Endpoint, test.HttpClient.endpoint)
			assert.Equal(t, test.Data, test.HttpClient.data)

			if test.Error != nil {
				assert.Nil(t, res)
			} else {
				assert.NotNil(t, res)
				assert.Equal(t, test.StatusCode, res.StatusCode)
			}
		})
	}
}

package gent

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestNewRequestBuilder tests creating a new request builder.
func TestNewRequest(t *testing.T) {
	tests := []struct {
		Name   string
		Method string
		Format string
	}{
		{
			Name:   "Creating new request",
			Method: http.MethodGet,
			Format: "http://localhost:8080",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			req := NewRequest(test.Method, test.Format)

			assert.Equal(t, test.Method, req.method)
			assert.Equal(t, test.Format, req.format)
			assert.Nil(t, req.body)
			assert.Nil(t, req.marshaler)
			assert.Nil(t, req.headers)
			assert.Nil(t, req.queryPrms)
			assert.Nil(t, req.pathPrms)

		})
	}
}

// TestNewRequestBuilder tests adding a byte array body to a request builder.
func TestRequestWithRawBody(t *testing.T) {
	tests := []struct {
		Name    string
		Builder *RequestBuilder
		Body    []byte
	}{
		{
			Name:    "Adding new request body",
			Builder: &RequestBuilder{},
			Body:    []byte("{\"key\":\"value\"}"),
		},
		{
			Name: "Overwriting existing request body",
			Builder: &RequestBuilder{
				body:      map[string]any{},
				marshaler: JsonMarshaler,
			},
			Body: []byte("{\"key\":\"value\"}"),
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			req := test.Builder.WithRawBody(test.Body)

			assert.Equal(t, test.Body, req.body)
			assert.Nil(t, req.marshaler)
		})
	}
}

// TestRequestWithBody tests adding body and marshaler to a request builder.
func TestRequestWithBody(t *testing.T) {
	tests := []struct {
		Name      string
		Builder   *RequestBuilder
		Body      any
		Marshaler Marshaler
	}{
		{
			Name:      "Adding new request body",
			Builder:   &RequestBuilder{},
			Body:      map[string]any{"Name": "John Smith"},
			Marshaler: JsonMarshaler,
		},
		{
			Name: "Overwriting existing request body",
			Builder: &RequestBuilder{
				body:      "placeholder",
				marshaler: XmlMarshaler,
			},
			Body:      map[string]any{"Name": "John Smith"},
			Marshaler: JsonMarshaler,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			req := test.Builder.WithBody(test.Body, test.Marshaler)

			assert.Equal(t, test.Body, req.body)
			//assert.Equal(t, test.Marshaler, req.marshaler)
		})
	}
}

// TestRequestWithHeader tests adding headers to a request builder.
func TestRequestWithHeader(t *testing.T) {
	tests := []struct {
		Name    string
		Builder *RequestBuilder
		Added   map[string][]string
		After   map[string][]string
	}{
		{
			Name:    "Adding headers to empty set",
			Builder: &RequestBuilder{},
			Added: map[string][]string{
				"Content-Type":   {"application/json"},
				"Content-Length": {"1024"},
			},
			After: map[string][]string{
				"Content-Type":   {"application/json"},
				"Content-Length": {"1024"},
			},
		},
		{
			Name: "Adding headers to populated set",
			Builder: &RequestBuilder{
				headers: map[string][]string{
					"Authorization": {"Bearer x.y.z"},
					"X-Api-Key":     {"cGxhY2Vob2xkZXI="},
				},
			},
			Added: map[string][]string{
				"Content-Type":   {"application/json"},
				"Content-Length": {"1024"},
			},
			After: map[string][]string{
				"Authorization":  {"Bearer x.y.z"},
				"X-Api-Key":      {"cGxhY2Vob2xkZXI="},
				"Content-Type":   {"application/json"},
				"Content-Length": {"1024"},
			},
		},
		{
			Name: "Adding value to existing header",
			Builder: &RequestBuilder{
				headers: map[string][]string{
					"Authorization": {"Bearer x.y.z"},
					"X-Api-Key":     {"cGxhY2Vob2xkZXI="},
				},
			},
			Added: map[string][]string{
				"Content-Type": {"application/json"},
				"X-Api-Key":    {"dGVzdGluZw=="},
			},
			After: map[string][]string{
				"Authorization": {"Bearer x.y.z"},
				"X-Api-Key":     {"cGxhY2Vob2xkZXI=", "dGVzdGluZw=="},
				"Content-Type":  {"application/json"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {

			for key, vals := range test.Added {
				for _, val := range vals {
					test.Builder.WithHeader(key, val)
				}
			}

			assert.Equal(t, test.After, test.Builder.headers)
		})
	}
}

// TestRequestWithQueryParameter tests adding query parameters to a request builder.
func TestRequestWithQueryParameter(t *testing.T) {
	tests := []struct {
		Name    string
		Builder *RequestBuilder
		Added   map[string][]string
		After   map[string][]string
	}{
		{
			Name:    "Adding parameters to empty set",
			Builder: &RequestBuilder{},
			Added: map[string][]string{
				"ids":   {"123", "456", "789"},
				"order": {"asc"},
			},
			After: map[string][]string{
				"ids":   {"123", "456", "789"},
				"order": {"asc"},
			},
		},
		{
			Name: "Adding parameters to populated set",
			Builder: &RequestBuilder{
				queryPrms: map[string][]string{
					"ids":   {"123", "456", "789"},
					"order": {"asc"},
				},
			},
			Added: map[string][]string{
				"page":  {"1"},
				"items": {"20"},
			},
			After: map[string][]string{
				"ids":   {"123", "456", "789"},
				"order": {"asc"},
				"page":  {"1"},
				"items": {"20"},
			},
		},
		{
			Name: "Adding value to existing parameters",
			Builder: &RequestBuilder{
				queryPrms: map[string][]string{
					"ids":   {"123", "456", "789"},
					"order": {"asc"},
				},
			},
			Added: map[string][]string{
				"ids":     {"abc", "def", "ghi"},
				"orderby": {"id"},
			},
			After: map[string][]string{
				"ids":     {"123", "456", "789", "abc", "def", "ghi"},
				"order":   {"asc"},
				"orderby": {"id"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {

			for key, vals := range test.Added {
				test.Builder.WithQueryParameter(key, vals)
			}

			assert.Equal(t, test.After, test.Builder.queryPrms)
		})
	}
}

// TestRequestWithPathParameters tests adding path parameters to a request builder.
func TestRequestWithPathParameters(t *testing.T) {
	tests := []struct {
		Name    string
		Builder *RequestBuilder
		Added   []string
		After   []string
	}{
		{
			Name:    "Adding parameters to empty set",
			Builder: &RequestBuilder{},
			Added:   []string{"123", "456"},
			After:   []string{"123", "456"},
		},
		{
			Name: "Adding parameters to populated set",
			Builder: &RequestBuilder{
				pathPrms: []string{"123", "456"},
			},
			Added: []string{"789", "abc"},
			After: []string{"123", "456", "789", "abc"},
		},
		{
			Name: "Adding parameters to be escaped",
			Builder: &RequestBuilder{
				pathPrms: []string{"123", "456"},
			},
			Added: []string{"Hello, Wold!"},
			After: []string{"123", "456", "Hello%2C%20Wold%21"},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {

			test.Builder.WithPathParameters(test.Added...)

			assert.Equal(t, test.After, test.Builder.pathPrms)
		})
	}
}

// TestRequestBuild tests building a request.
func TestRequestBuild(t *testing.T) {
	tests := []struct {
		Name          string
		Builder       *RequestBuilder
		Context       context.Context
		Error         error
		Method        string
		Body          []byte
		ContentLength int64
		Host          string
		Path          string
		QueryPrms     url.Values
		Headers       http.Header
	}{
		{
			Name: "GET request with details",
			Builder: &RequestBuilder{
				method:    http.MethodGet,
				format:    "https://localhost:8080/users",
				body:      nil,
				marshaler: nil,
				headers: map[string][]string{
					"Authorization": {"eyJhbGciOiJIUzI1NiJ9.e30.ZRrHA1JJJW8opsbCGfG"},
				},
				queryPrms: map[string][]string{
					"$select": {"id,name,email"},
				},
				pathPrms: nil,
			},
			Context:       context.Background(),
			Error:         nil,
			Method:        http.MethodGet,
			Body:          []byte(""),
			ContentLength: 0,
			Host:          "localhost:8080",
			Path:          "/users",
			QueryPrms: map[string][]string{
				"$select": {"id,name,email"},
			},
			Headers: map[string][]string{
				"Authorization": {"eyJhbGciOiJIUzI1NiJ9.e30.ZRrHA1JJJW8opsbCGfG"},
			},
		},
		{
			Name: "PUT request with details",
			Builder: &RequestBuilder{
				method: http.MethodPatch,
				format: "https://localhost:8080/users/{}/devices/{}/name",
				body: map[string]any{
					"Name": "My Phone",
				},
				marshaler: JsonMarshaler,
				headers: map[string][]string{
					"Authorization": {"eyJhbGciOiJIUzI1NiJ9.e30.ZRrHA1JJJW8opsbCGfG"},
				},
				queryPrms: nil,
				pathPrms:  []string{"4481e035-1711-419f-82bc-bfb72da06375", "01JW3ZFVR44BWEAJRW7TEQ3PK0"},
			},
			Context:       context.Background(),
			Error:         nil,
			Method:        http.MethodPatch,
			Body:          []byte(`{"Name":"My Phone"}`),
			ContentLength: 19,
			Host:          "localhost:8080",
			Path:          "/users/4481e035-1711-419f-82bc-bfb72da06375/devices/01JW3ZFVR44BWEAJRW7TEQ3PK0/name",
			QueryPrms:     map[string][]string{},
			Headers: map[string][]string{
				"Authorization": {"eyJhbGciOiJIUzI1NiJ9.e30.ZRrHA1JJJW8opsbCGfG"},
				"Content-Type":  {"application/json"},
			},
		},
		{
			Name: "Request format includes query",
			Builder: &RequestBuilder{
				method:    http.MethodGet,
				format:    "https://localhost:8080/users?orderby=id",
				body:      nil,
				marshaler: nil,
				headers: map[string][]string{
					"Authorization": {"eyJhbGciOiJIUzI1NiJ9.e30.ZRrHA1JJJW8opsbCGfG"},
				},
				queryPrms: map[string][]string{
					"$select": {"id,name,email"},
				},
				pathPrms: nil,
			},
			Context:       context.Background(),
			Error:         nil,
			Method:        http.MethodGet,
			Body:          []byte(""),
			ContentLength: 0,
			Host:          "localhost:8080",
			Path:          "/users",
			QueryPrms: map[string][]string{
				"$select": {"id,name,email"},
				"orderby": {"id"},
			},
			Headers: map[string][]string{
				"Authorization": {"eyJhbGciOiJIUzI1NiJ9.e30.ZRrHA1JJJW8opsbCGfG"},
			},
		},
		{
			Name: "Request body is bytes",
			Builder: &RequestBuilder{
				method:    http.MethodPost,
				format:    "https://localhost:8080/events",
				body:      []byte("UserUpdated"),
				marshaler: nil,
				headers:   nil,
				queryPrms: nil,
				pathPrms:  nil,
			},
			Context:       context.Background(),
			Error:         nil,
			Method:        http.MethodPost,
			Body:          []byte("UserUpdated"),
			ContentLength: 11,
			Host:          "localhost:8080",
			Path:          "/events",
			QueryPrms:     map[string][]string{},
			Headers:       map[string][]string{},
		},
		{
			Name: "Request body requires marshaler",
			Builder: &RequestBuilder{
				method:    http.MethodPost,
				format:    "https://localhost:8080/events",
				body:      time.Now(),
				marshaler: nil,
				headers:   nil,
				queryPrms: nil,
				pathPrms:  nil,
			},
			Context: context.Background(),
			Error:   ErrInvalidBodyType,
		},
		{
			Name: "Request body marshaling fails",
			Builder: &RequestBuilder{
				method:    http.MethodPost,
				format:    "https://localhost:8080/events",
				body:      time.Now(),
				marshaler: UrlEncodedMarshaler,
				headers:   nil,
				queryPrms: nil,
				pathPrms:  nil,
			},
			Context: context.Background(),
			Error:   ErrInvalidBodyType,
		},
		{
			Name: "Request format has trailing open bracket",
			Builder: &RequestBuilder{
				method:    http.MethodPut,
				format:    "https://localhost:8080/users/{}/devices/{",
				body:      nil,
				marshaler: nil,
				headers:   nil,
				queryPrms: nil,
				pathPrms:  []string{"123", "abc"},
			},
			Context: context.Background(),
			Error:   ErrInvalidFormat,
		},
		{
			Name: "Request format has unclosed bracket",
			Builder: &RequestBuilder{
				method:    http.MethodPut,
				format:    "https://localhost:8080/users/{}/devices/{/name",
				body:      nil,
				marshaler: nil,
				headers:   nil,
				queryPrms: nil,
				pathPrms:  []string{"123", "abc"},
			},
			Context: context.Background(),
			Error:   ErrInvalidFormat,
		},
		{
			Name: "Request format has unopened bracket",
			Builder: &RequestBuilder{
				method:    http.MethodPut,
				format:    "https://localhost:8080/users/{}/devices/}/name",
				body:      nil,
				marshaler: nil,
				headers:   nil,
				queryPrms: nil,
				pathPrms:  []string{"123", "abc"},
			},
			Context: context.Background(),
			Error:   ErrInvalidFormat,
		},
		{
			Name: "Request has too many path parameters",
			Builder: &RequestBuilder{
				method:    http.MethodPut,
				format:    "https://localhost:8080/users/{}/devices/{}/name",
				body:      nil,
				marshaler: nil,
				headers:   nil,
				queryPrms: nil,
				pathPrms:  []string{"123", "abc", "xyz"},
			},
			Context: context.Background(),
			Error:   ErrInvalidFormat,
		},
		{
			Name: "Request has not enough path parameters",
			Builder: &RequestBuilder{
				method:    http.MethodPut,
				format:    "https://localhost:8080/users/{}/devices/{}/name",
				body:      nil,
				marshaler: nil,
				headers:   nil,
				queryPrms: nil,
				pathPrms:  []string{"123"},
			},
			Context: context.Background(),
			Error:   ErrInvalidFormat,
		},
		{
			Name: "Request fails to build",
			Builder: &RequestBuilder{
				method:    http.MethodPut,
				format:    string([]byte{0}),
				body:      nil,
				marshaler: nil,
				headers:   nil,
				queryPrms: nil,
				pathPrms:  nil,
			},
			Context: context.Background(),
			Error: &url.Error{
				Op:  "parse",
				URL: "\x00",
				Err: errors.New("net/url: invalid control character in URL"),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {

			req, err := test.Builder.Build(test.Context)

			assert.Equal(t, test.Error, err)
			if err == nil {
				assert.NotNil(t, req)

				assert.Equal(t, test.Context, req.Context())
				assert.Equal(t, test.Method, req.Method)
				assert.Equal(t, test.ContentLength, req.ContentLength)

				assert.NotNil(t, req.URL)
				assert.Equal(t, test.Host, req.URL.Host)
				assert.Equal(t, test.Path, req.URL.Path)
				assert.Equal(t, test.QueryPrms, req.URL.Query())
				assert.Equal(t, test.Headers, req.Header)

				body, _ := io.ReadAll(req.Body)
				assert.Equal(t, test.Body, body)
			} else {
				assert.Nil(t, req)
			}
		})
	}
}

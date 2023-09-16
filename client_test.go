package gent

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCreateClient tests that a client can be created
func TestCreateClient(t *testing.T) {
	tests := []struct {
		Name string
	}{
		{Name: "Creating client with no options"},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			cl := NewClient()

			assert.NotNil(t, cl.mem)
			assert.NotNil(t, cl.client)
			assert.Nil(t, cl.retr)
			assert.Nil(t, cl.constr)
			assert.Equal(t, 2, len(cl.functions))
		})
	}
}

// TestCreateClient tests that a client can be created and configured
func TestCreateClientWithOptions(t *testing.T) {
	tests := []struct {
		Name         string
		MemPool      MemoryPool
		Retrier      Retrier
		Client       HttpClient
		Constructor  func() HttpClient
		HLMiddleware func(context.Context, *Request)
		LLMiddleware func(context.Context, *Request)
	}{
		{
			Name:         "Creating client with options",
			MemPool:      &mockMemPool{},
			Retrier:      &mockRetrier{},
			Client:       &mockHttpClient{},
			Constructor:  func() HttpClient { return &mockHttpClient{} },
			HLMiddleware: func(context.Context, *Request) {},
			LLMiddleware: func(context.Context, *Request) {},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			cl := NewClient(
				UseMemoryPool(test.MemPool),
				UseRetrier(test.Retrier),
				UseHttpClient(test.Client),
				UseHttpClientConstructor(test.Constructor),
				UseHighLevelMiddleware(test.HLMiddleware),
				UseLowLevelMiddleware(test.LLMiddleware),
			)

			assert.Equal(t, test.MemPool, cl.mem)
			assert.Equal(t, test.Client, cl.client)
			assert.NotNil(t, cl.constr)
			assert.Equal(t, 4, len(cl.functions))
		})
	}
}

// TestGetClientForRequest tests that a client is acquired from constructors
func TestGetClientForRequest(t *testing.T) {
	tests := []struct {
		Name           string
		Client         *Client
		InternalClient HttpClient
	}{
		{
			Name:           "Creating client without constructor",
			Client:         NewClient(),
			InternalClient: http.DefaultClient,
		},
		{
			Name: "Creating client with constructor",
			Client: NewClient(UseHttpClientConstructor(
				func() HttpClient { return &mockHttpClient{} },
			)),
			InternalClient: &mockHttpClient{},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			intrn := test.Client.getClientForRequest()

			assert.Equal(t, test.InternalClient, intrn)
		})
	}
}

// TestMakeRequest tests if a request can be made correctly
func TestMakeRequest(t *testing.T) {
	tests := []struct {
		Name        string
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
			Name:   "Successful request",
			Method: "POST",
			Format: "http://localhost:8080/{}",
			Body: map[string]any{
				"test": "test",
			},
			Marshaler: NewJSONMarshaler(),
			Headers: map[string]string{
				"Authorization": "Bearer x.y.z",
			},
			QueryParams: map[string][]string{
				"query": {"query"},
			},
			PathParams: []string{
				"param",
			},
			Data:       []byte(`{"test":"test"}`),
			Endpoint:   []byte("http://localhost:8080/param?query=query"),
			StatusCode: 201,
			Error:      nil,
		},
		{
			Name:   "Unsuccessful request",
			Method: "GET",
			Format: "http://localhost:8080/{}",
			Body: map[string]any{
				"test": "test",
			},
			Marshaler: NewJSONMarshaler(),
			Headers: map[string]string{
				"Authorization": "Bearer x.y.z",
			},
			QueryParams: map[string][]string{
				"query": {"query"},
			},
			PathParams: []string{
				"param",
			},
			Data:       []byte(`{"test":"test"}`),
			Endpoint:   []byte("http://localhost:8080/param?query=query"),
			StatusCode: 500,
			Error:      fmt.Errorf("request failed"),
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			intrn := &mockHttpHandler{
				code:    test.StatusCode,
				err:     test.Error,
				headers: map[string]string{},
			}
			cl := NewClient(UseHttpClient(intrn))

			res, err := cl.Do(
				context.TODO(),
				test.Method,
				test.Format,
				test.Body,
				test.Marshaler,
				test.Headers,
				test.QueryParams,
				test.PathParams...,
			)

			assert.Equal(t, test.Error, err)
			assert.Equal(t, test.Method, intrn.method)
			assert.Equal(t, test.Endpoint, intrn.endpoint)
			assert.Equal(t, test.Data, intrn.data)

			if test.Error != nil {
				assert.Nil(t, res)
			} else {
				assert.NotNil(t, res)
				assert.Equal(t, test.StatusCode, res.StatusCode)
			}
		})
	}
}

// TestGetRequest tests if a GET request can be made correctly
func TestGetRequest(t *testing.T) {
	tests := []struct {
		Name   string
		Format string
		Method string
		Error  error
	}{
		{
			Name:   "Get request",
			Format: "http://localhost:8080",
			Method: "GET",
			Error:  nil,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			intrn := &mockHttpHandler{
				code:    200,
				err:     test.Error,
				headers: map[string]string{},
			}
			cl := NewClient(UseHttpClient(intrn))

			_, err := cl.Get(
				context.TODO(),
				test.Format,
				nil, nil, nil, nil,
			)

			assert.Equal(t, test.Error, err)
			assert.Equal(t, test.Method, intrn.method)
		})
	}
}

// TestPostRequest tests if a POST request can be made correctly
func TestPostRequest(t *testing.T) {
	tests := []struct {
		Name   string
		Format string
		Method string
		Error  error
	}{
		{
			Name:   "Post request",
			Format: "http://localhost:8080",
			Method: "POST",
			Error:  nil,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			intrn := &mockHttpHandler{
				code:    200,
				err:     test.Error,
				headers: map[string]string{},
			}
			cl := NewClient(UseHttpClient(intrn))

			_, err := cl.Post(
				context.TODO(),
				test.Format,
				nil, nil, nil, nil,
			)

			assert.Equal(t, test.Error, err)
			assert.Equal(t, test.Method, intrn.method)
		})
	}
}

// TestPatchRequest tests if a PATCH request can be made correctly
func TestPatchRequest(t *testing.T) {
	tests := []struct {
		Name   string
		Format string
		Method string
		Error  error
	}{
		{
			Name:   "Patch request",
			Format: "http://localhost:8080",
			Method: "PATCH",
			Error:  nil,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			intrn := &mockHttpHandler{
				code:    200,
				err:     test.Error,
				headers: map[string]string{},
			}
			cl := NewClient(UseHttpClient(intrn))

			_, err := cl.Patch(
				context.TODO(),
				test.Format,
				nil, nil, nil, nil,
			)

			assert.Equal(t, test.Error, err)
			assert.Equal(t, test.Method, intrn.method)
		})
	}
}

// TestPutRequest tests if a PUT request can be made correctly
func TestPutRequest(t *testing.T) {
	tests := []struct {
		Name   string
		Format string
		Method string
		Error  error
	}{
		{
			Name:   "Put request",
			Format: "http://localhost:8080",
			Method: "PUT",
			Error:  nil,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			intrn := &mockHttpHandler{
				code:    200,
				err:     test.Error,
				headers: map[string]string{},
			}
			cl := NewClient(UseHttpClient(intrn))

			_, err := cl.Put(
				context.TODO(),
				test.Format,
				nil, nil, nil, nil,
			)

			assert.Equal(t, test.Error, err)
			assert.Equal(t, test.Method, intrn.method)
		})
	}
}

// TestDeleteRequest tests if a DELETE request can be made correctly
func TestDeleteRequest(t *testing.T) {
	tests := []struct {
		Name   string
		Format string
		Method string
		Error  error
	}{
		{
			Name:   "Delete request",
			Format: "http://localhost:8080",
			Method: "DELETE",
			Error:  nil,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			intrn := &mockHttpHandler{
				code:    200,
				err:     test.Error,
				headers: map[string]string{},
			}
			cl := NewClient(UseHttpClient(intrn))

			_, err := cl.Delete(
				context.TODO(),
				test.Format,
				nil, nil, nil, nil,
			)

			assert.Equal(t, test.Error, err)
			assert.Equal(t, test.Method, intrn.method)
		})
	}
}

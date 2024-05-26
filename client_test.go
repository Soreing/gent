package gent

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCreateClient tests that a client can be created.
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
			assert.Nil(t, cl.constr)
		})
	}
}

// TestCreateClient tests that a client can be created and configured.
func TestCreateClientWithOptions(t *testing.T) {
	tests := []struct {
		Name        string
		MemPool     MemoryPool
		Client      HttpClient
		Constructor func() HttpClient
	}{
		{
			Name:        "Creating client with options",
			MemPool:     &mockMemPool{},
			Client:      &mockHttpClient{},
			Constructor: func() HttpClient { return &mockHttpClient{} },
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			cl := NewClient(
				UseMemoryPool(test.MemPool),
				UseHttpClient(test.Client),
				UseHttpClientConstructor(test.Constructor),
			)

			assert.Equal(t, test.MemPool, cl.mem)
			assert.Equal(t, test.Client, cl.client)
			assert.NotNil(t, cl.constr)
		})
	}
}

// TestGetClientForRequest tests that a client is acquired from constructors.
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

// TestMakeRequest tests if a request can be made correctly.
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
			Method: http.MethodPost,
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

// TestGetRequest tests if a GET request can be made correctly.
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
			Method: http.MethodGet,
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

// TestPostRequest tests if a POST request can be made correctly.
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
			Method: http.MethodPost,
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

// TestPatchRequest tests if a PATCH request can be made correctly.
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
			Method: http.MethodPatch,
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

// TestPutRequest tests if a PUT request can be made correctly.
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
			Method: http.MethodPut,
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

// TestDeleteRequest tests if a DELETE request can be made correctly.
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
			Method: http.MethodDelete,
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

// TestUseBeforeBuildMiddleware tests adding middlewares before the build stage.
func TestUseBeforeBuildMiddleware(t *testing.T) {
	tests := []struct {
		Name     string
		Function func(context.Context, *Request)
		Error    error
	}{
		{
			Name:  "Middleware before build",
			Error: nil,

			Function: func(ctx context.Context, req *Request) {
				req.QueryParams["order"] = []string{"desc"}
				req.Next()
			},
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

			err := cl.Use(MDW_BeforeBuild, test.Function)

			assert.Equal(t, test.Error, err)
			assert.Equal(t, 0, len(cl.l1mdw))
			assert.Equal(t, 1, len(cl.l2mdw))
		})
	}
}

// TestUseBeforeExecuteMiddleware tests adding middlewares before the execute stage.
func TestUseBeforeExecuteMiddleware(t *testing.T) {
	tests := []struct {
		Name     string
		Function func(context.Context, *Request)
		Error    error
	}{
		{
			Name:  "Middleware before execute",
			Error: nil,

			Function: func(ctx context.Context, req *Request) {
				req.Request.Header.Set("Authorization", "x.y.z")
				req.Next()
			},
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

			err := cl.Use(MDW_BeforeExecute, test.Function)

			assert.Equal(t, test.Error, err)
			assert.Equal(t, 1, len(cl.l1mdw))
			assert.Equal(t, 0, len(cl.l2mdw))
		})
	}
}

// TestUseInvlaidMiddlewareStage tests adding middlewares to an invalid stage.
func TestUseInvlaidMiddlewareStage(t *testing.T) {
	tests := []struct {
		Name     string
		Function func(context.Context, *Request)
		Error    error
	}{
		{
			Name:  "Middleware invalid stage",
			Error: fmt.Errorf("invalid middleware stage"),

			Function: func(ctx context.Context, req *Request) {},
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

			err := cl.Use(MiddlewareStage(-1), test.Function)

			assert.Equal(t, test.Error, err)
			assert.Equal(t, 0, len(cl.l1mdw))
			assert.Equal(t, 0, len(cl.l2mdw))
		})
	}
}

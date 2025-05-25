package gent

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewDefaultClient tests creating a default client.
func TestNewDefaultClient(t *testing.T) {
	tests := []struct {
		Name string
	}{
		{Name: "New default client"},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			cl := NewDefaultClient()

			assert.Nil(t, cl.mdws)
			assert.Equal(t, http.DefaultClient, cl.cl)
		})
	}
}

// TestNewClient tests creating a client.
func TestNewClient(t *testing.T) {
	tests := []struct {
		Name      string
		Requester Requester
	}{
		{
			Name:      "New client",
			Requester: &mockRequester{},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			cl := NewClient(test.Requester)

			assert.Nil(t, cl.mdws)
			assert.Equal(t, test.Requester, cl.cl)
		})
	}
}

// TestClientUse tests adding middlewares to the client.
func TestClientUse(t *testing.T) {
	tests := []struct {
		Name      string
		Client    *Client
		Before    []func(*Context)
		Functions []func(*Context)
		After     []func(*Context)
	}{
		{
			Name:   "New client",
			Client: NewDefaultClient(),
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			cl := NewDefaultClient()
			cl.mdws = test.Before

			cl.Use(test.Functions...)

			assert.Equal(t, test.After, cl.mdws)
		})
	}
}

// TestClientDo tests making a request.
func TestClientDo(t *testing.T) {
	tests := []struct {
		Name      string
		Requester *mockRequester
		Url       string
		Error     bool
	}{
		{
			Name:      "Successful request",
			Url:       "https://localhost:8080",
			Requester: &mockRequester{},
			Error:     false,
		},
		{
			Name: "Request Failed",
			Url:  "https://localhost:8080",
			Requester: &mockRequester{
				RequestErr: fmt.Errorf("failed"),
			},
			Error: true,
		},
		{
			Name:      "Failed making request",
			Url:       string([]byte{0x0}),
			Requester: &mockRequester{},
			Error:     true,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			cl := NewClient(test.Requester)

			res, err := cl.Get(test.Url)

			if test.Error {
				assert.Nil(t, res)
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, res)
				assert.Equal(t, 1, test.Requester.CountCalled)
				assert.Equal(t, http.MethodGet, test.Requester.LastRequest.Method)
			}
		})
	}
}

// TestClientGet tests making a GET request.
func TestClientGet(t *testing.T) {
	tests := []struct {
		Name      string
		Requester *mockRequester
		Url       string
		Error     bool
	}{
		{
			Name:      "Successful request",
			Url:       "https://localhost:8080",
			Requester: &mockRequester{},
			Error:     false,
		},
		{
			Name:      "Failed making request",
			Url:       string([]byte{0x0}),
			Requester: &mockRequester{},
			Error:     true,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			cl := NewClient(test.Requester)

			res, err := cl.Get(test.Url)

			if test.Error {
				assert.Nil(t, res)
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, res)
				assert.Equal(t, 1, test.Requester.CountCalled)
				assert.Equal(t, http.MethodGet, test.Requester.LastRequest.Method)
			}
		})
	}
}

// TestClientHead tests making a HEAD request.
func TestClientHead(t *testing.T) {
	tests := []struct {
		Name      string
		Requester *mockRequester
		Url       string
		Error     bool
	}{
		{
			Name:      "Successful request",
			Url:       "https://localhost:8080",
			Requester: &mockRequester{},
			Error:     false,
		},
		{
			Name:      "Failed making request",
			Url:       string([]byte{0x0}),
			Requester: &mockRequester{},
			Error:     true,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			cl := NewClient(test.Requester)

			res, err := cl.Head(test.Url)

			if test.Error {
				assert.Nil(t, res)
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, res)
				assert.Equal(t, 1, test.Requester.CountCalled)
				assert.Equal(t, http.MethodHead, test.Requester.LastRequest.Method)
			}
		})
	}
}

// TestClientPost tests making a POST request.
func TestClientPost(t *testing.T) {
	tests := []struct {
		Name      string
		Requester *mockRequester
		Url       string
		Error     bool
	}{
		{
			Name:      "Successful request",
			Url:       "https://localhost:8080",
			Requester: &mockRequester{},
			Error:     false,
		},
		{
			Name:      "Failed making request",
			Url:       string([]byte{0x0}),
			Requester: &mockRequester{},
			Error:     true,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			cl := NewClient(test.Requester)

			cype := "application/json"
			body := []byte(`{"name": "John Smith"}`)
			res, err := cl.Post(test.Url, cype, bytes.NewBuffer(body))

			if test.Error {
				assert.Nil(t, res)
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, res)
				assert.Equal(t, 1, test.Requester.CountCalled)
				assert.Equal(t, http.MethodPost, test.Requester.LastRequest.Method)

				rbody, _ := io.ReadAll(test.Requester.LastRequest.Body)
				assert.Equal(t, body, rbody)
				assert.Equal(t, cype, test.Requester.LastRequest.Header.Get("Content-Type"))
			}
		})
	}
}

// TestClientPostForm tests making a POST form request.
func TestClientPostForm(t *testing.T) {
	tests := []struct {
		Name      string
		Requester *mockRequester
		Url       string
		Error     bool
	}{
		{
			Name:      "Successful request",
			Url:       "https://localhost:8080",
			Requester: &mockRequester{},
			Error:     false,
		},
		{
			Name:      "Failed making request",
			Url:       string([]byte{0x0}),
			Requester: &mockRequester{},
			Error:     true,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			cl := NewClient(test.Requester)

			vals := url.Values{}
			vals.Add("name", "John Smith")
			res, err := cl.PostForm(test.Url, vals)

			if test.Error {
				assert.Nil(t, res)
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, res)
				assert.Equal(t, 1, test.Requester.CountCalled)
				assert.Equal(t, http.MethodPost, test.Requester.LastRequest.Method)

				rbody, _ := io.ReadAll(test.Requester.LastRequest.Body)
				assert.Equal(t, []byte(vals.Encode()), rbody)
				assert.Equal(
					t, "application/x-www-form-urlencoded",
					test.Requester.LastRequest.Header.Get("Content-Type"),
				)
			}
		})
	}
}

// TestClientCloseIdleConnections tests closing idle connections.
func TestClientCloseIdleConnections(t *testing.T) {
	tests := []struct {
		Name string
	}{
		{Name: "Close"},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			req := &mockRequester{}
			cl := NewClient(req)

			cl.CloseIdleConnections()

			assert.Equal(t, 1, req.ClosedCount)
		})
	}
}

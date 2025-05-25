package gent

import (
	"bytes"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestNewRequestContext tests creating a request context.
func TestNewRequestContext(t *testing.T) {
	tests := []struct {
		Name      string
		Client    Requester
		Request   *http.Request
		Functions []func(*Context)
	}{
		{
			Client:    http.DefaultClient,
			Request:   &http.Request{},
			Functions: []func(*Context){func(*Context) { fmt.Println("Test") }},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			ctx := newRequestContext(test.Client, test.Request, test.Functions)

			assert.NotNil(t, ctx.mtx)
			assert.NotNil(t, ctx.Values)
			assert.Nil(t, ctx.Response)

			assert.Equal(t, test.Client, ctx.cl)
			assert.Equal(t, test.Request, ctx.Request)
			assert.Equal(t, test.Functions, ctx.fns)
			assert.Equal(t, 0, ctx.fni)
		})
	}
}

// TestContextErrors tests adding an error to the context.
func TestContextErrors(t *testing.T) {
	tests := []struct {
		Name   string
		Errors []error
	}{
		{
			Name: "Setting one error",
			Errors: []error{
				fmt.Errorf("Error"),
			},
		},
		{
			Name: "Setting multiple errors",
			Errors: []error{
				fmt.Errorf("Error1"),
				fmt.Errorf("Error1"),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			req := &Context{}

			for _, e := range test.Errors {
				req.Error(e)
			}

			assert.Equal(t, test.Errors, req.Errors)
		})
	}
}

// TestContextLock tests locking the context.
func TestContextLock(t *testing.T) {
	tests := []struct {
		Name string
	}{
		{Name: "Lock"},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			req := Context{mtx: &sync.RWMutex{}}

			req.Lock()

			ok := req.mtx.TryLock()
			assert.Equal(t, false, ok)
		})
	}
}

// TestContextUnlock tests unlocking the context.
func TestContextUnlock(t *testing.T) {
	tests := []struct {
		Name string
	}{
		{Name: "Unlock"},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			req := Context{mtx: &sync.RWMutex{}}
			req.mtx.Lock()

			req.Unlock()

			ok := req.mtx.TryLock()
			assert.Equal(t, true, ok)
		})
	}
}

// TestContextGet tests getting values from the context.
func TestContextGet(t *testing.T) {
	tests := []struct {
		Name   string
		Values map[string]any
		Key    string
		Exists bool
		Value  any
	}{
		{
			Name: "Value exists",
			Values: map[string]any{
				"retries":  "0",
				"attempts": 1,
			},
			Key:    "retries",
			Exists: true,
			Value:  "0",
		},
		{
			Name: "Value does not exist",
			Values: map[string]any{
				"attempts": 1,
			},
			Key:    "retries",
			Exists: false,
			Value:  nil,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			req := Context{
				mtx:    &sync.RWMutex{},
				Values: test.Values,
			}

			val, ok := req.Get(test.Key)

			assert.Equal(t, test.Exists, ok)
			assert.Equal(t, test.Value, val)
		})
	}
}

// TestContextSet tests setting values in the context.
func TestContextSet(t *testing.T) {
	tests := []struct {
		Name   string
		Before map[string]any
		Key    string
		Value  string
		After  map[string]any
	}{
		{
			Name: "Set value",
			Before: map[string]any{
				"attempts": 1,
			},
			Key:   "retries",
			Value: "0",
			After: map[string]any{
				"retries":  "0",
				"attempts": 1,
			},
		},
		{
			Name: "Overwrite value",
			Before: map[string]any{
				"retries":  "0",
				"attempts": 1,
			},
			Key:   "retries",
			Value: "1",
			After: map[string]any{
				"retries":  "1",
				"attempts": 1,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			req := Context{
				mtx:    &sync.RWMutex{},
				Values: test.Before,
			}

			req.Set(test.Key, test.Value)

			val, ok := req.Values[test.Key]
			assert.Equal(t, true, ok)
			assert.Equal(t, test.Value, val)
		})
	}
}

// TestSetRequestValue tests deleting values from the context.
func TestContextDel(t *testing.T) {
	tests := []struct {
		Name   string
		Before map[string]any
		Key    string
		Value  string
		After  map[string]any
	}{
		{
			Name: "Delete existing value",
			Before: map[string]any{
				"retries":  "0",
				"attempts": 1,
			},
			Key: "retries",
			After: map[string]any{
				"attempts": 1,
			},
		},
		{
			Name: "Delete missing value",
			Before: map[string]any{
				"retries":  "0",
				"attempts": 1,
			},
			Key: "errors",
			After: map[string]any{
				"retries":  "0",
				"attempts": 1,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			req := Context{
				mtx:    &sync.RWMutex{},
				Values: test.Before,
			}

			req.Del(test.Key)

			assert.Equal(t, test.After, req.Values)
		})
	}
}

// TestNextFunction tests if calling next in a middleware moves to the next
// function
func TestContextNext(t *testing.T) {
	tests := []struct {
		Name string
	}{
		{Name: "Next"},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			executed := false
			ctx := newRequestContext(&mockRequester{}, &http.Request{},
				[]func(*Context){func(ctx *Context) {
					executed = true
					ctx.Next()
				}},
			)

			ctx.Next()

			assert.Equal(t, true, executed)

		})
	}
}

// TestDo tests doing a request
func TestDo(t *testing.T) {
	tests := []struct {
		Name   string
		Client *mockRequester
	}{
		{
			Name: "Successful request",
			Client: &mockRequester{
				Delay:      time.Millisecond,
				StatusCode: 200,
			},
		},
		{
			Name: "Failed to make request",
			Client: &mockRequester{
				RequestErr: fmt.Errorf("failed"),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			req, _ := http.NewRequest(
				http.MethodGet, "http://localhost:8080", bytes.NewReader(nil),
			)
			ctx := newRequestContext(test.Client, req, nil)

			do(ctx)

			if test.Client.RequestErr != nil {
				assert.Equal(t, 1, len(ctx.Errors))
				assert.Equal(t, test.Client.RequestErr, ctx.Errors[0])
				assert.Nil(t, ctx.Response)
			} else {
				assert.Equal(t, 0, len(ctx.Errors))
				assert.NotNil(t, ctx.Response)
			}
		})
	}
}

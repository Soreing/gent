package gent

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestNewRequest tests if a request can be created and that the object's fields
// are properly initialized
func TestNewRequest(t *testing.T) {
	tests := []struct {
		Name        string
		Context     context.Context
		MemPool     MemoryPool
		Client      HttpClient
		Format      string
		Method      string
		Body        any
		Marshaler   Marshaler
		Headers     map[string]string
		QueryParams map[string][]string
		PathParams  []string
		Functions   []func(context.Context, *Request)
	}{
		{
			Name:        "New Request",
			Context:     context.TODO(),
			MemPool:     NewDefaultMemPool(),
			Client:      http.DefaultClient,
			Format:      "format",
			Method:      "method",
			Body:        map[string]any{},
			Marshaler:   NewJSONMarshaler(),
			Headers:     map[string]string{},
			QueryParams: map[string][]string{},
			PathParams:  []string{},
			Functions:   []func(context.Context, *Request){},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			req := newRequest(
				test.Context, test.MemPool, test.Client,
				test.Format, test.Method, test.Body, test.Marshaler,
				test.Headers, test.QueryParams, test.PathParams,
				test.Functions,
			)

			assert.NotNil(t, req.ctx)
			assert.NotNil(t, req.mem)
			assert.NotNil(t, req.client)
			assert.NotNil(t, req.mtx)
			assert.NotNil(t, req.Values)
			assert.NotNil(t, req.fns)
			assert.NotNil(t, req.Body)
			assert.NotNil(t, req.Marshaler)
			assert.NotNil(t, req.Headers)
			assert.NotNil(t, req.QueryParams)
			assert.NotNil(t, req.PathParams)

			assert.Nil(t, req.Endpoint)
			assert.Nil(t, req.Data)
			assert.Nil(t, req.Request)
			assert.Nil(t, req.Response)

			assert.Equal(t, 0, len(req.Errors))
			assert.Equal(t, 0, req.fni)
			assert.Equal(t, test.Format, req.Format)
			assert.Equal(t, test.Method, req.Method)

		})
	}
}

// TestNextFunction tests if calling next in a middleware moves to the next
// function
func TestNextFunction(t *testing.T) {
	tests := []struct {
		Name      string
		Functions []func(context.Context, *Request)
	}{
		{
			Name: "New Request",
			Functions: []func(context.Context, *Request){
				func(c context.Context, r *Request) {
					c.Value("ch").(chan int) <- 1
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			ch := make(chan int, 1)
			dl := time.Now().Add(time.Second)
			ctx := context.WithValue(context.TODO(), "ch", ch)
			ctx, cncl := context.WithDeadline(ctx, dl)
			defer cncl()

			req := &Request{
				ctx: ctx,
				fns: test.Functions,
			}

			go req.Next()

			select {
			case <-ctx.Done():
				t.Errorf("waiting for middlewares timed out")
			case <-ch:
			}

		})
	}
}

// TestSetError tests if adding errors get set in the request accurately
func TestAddError(t *testing.T) {
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
			req := &Request{}

			for _, e := range test.Errors {
				req.Error(e)
			}

			assert.Equal(t, test.Errors, req.Errors)
		})
	}
}

// TestGetErrors tests if getting errors from the request is accurate
func TestGetErrors(t *testing.T) {
	tests := []struct {
		Name   string
		Errors []error
	}{
		{
			Name: "Getting with one error",
			Errors: []error{
				fmt.Errorf("Error"),
			},
		},
		{
			Name: "Getting with multiple errors",
			Errors: []error{
				fmt.Errorf("Error1"),
				fmt.Errorf("Error1"),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			req := &Request{
				Errors: test.Errors,
			}

			errs := req.Errors

			assert.Equal(t, test.Errors, errs)
		})
	}
}

// TestPrepareRequest tests if preparing the request accurately creates an
// endpoint and a data byte array from the body and sets it to the request's
// fields
func TestPrepareRequest(t *testing.T) {
	tests := []struct {
		Name       string
		Method     string
		Format     string
		Headers    map[string]string
		Body       any
		Marshaler  Marshaler
		Endpoint   []byte
		Data       []byte
		CTHeader   string
		CTHeaderOk bool
		Errors     []error
	}{
		{
			Name:       "Endpoint created",
			Method:     http.MethodPost,
			Format:     "http://localhost:8080",
			Headers:    map[string]string{},
			Body:       nil,
			Marshaler:  nil,
			Endpoint:   []byte(`http://localhost:8080`),
			Data:       nil,
			CTHeader:   "",
			CTHeaderOk: false,
			Errors:     []error{},
		},
		{
			Name:       "Endpoint creation failed",
			Method:     http.MethodPost,
			Format:     "http://localhost:8080/{}",
			Headers:    map[string]string{},
			Body:       nil,
			Marshaler:  nil,
			Endpoint:   nil,
			Data:       nil,
			CTHeader:   "",
			CTHeaderOk: false,
			Errors: []error{
				fmt.Errorf("not enough parameters provided"),
			},
		},
		{
			Name:    "Body marshaled",
			Method:  http.MethodPost,
			Format:  "http://localhost:8080",
			Headers: map[string]string{},
			Body: map[string]any{
				"id":   123,
				"name": "John",
			},
			Marshaler:  NewJSONMarshaler(),
			Endpoint:   []byte(`http://localhost:8080`),
			Data:       []byte(`{"id":123,"name":"John"}`),
			CTHeader:   "application/json",
			CTHeaderOk: true,
			Errors:     []error{},
		},
		{
			Name:    "Body failed to marshal",
			Method:  http.MethodPost,
			Format:  "http://localhost:8080",
			Headers: map[string]string{},
			Body: map[string]any{
				"id":   123,
				"name": "John",
			},
			Marshaler:  NewFormMarshaler(),
			Endpoint:   []byte(`http://localhost:8080`),
			Data:       nil,
			CTHeader:   "",
			CTHeaderOk: false,
			Errors: []error{
				fmt.Errorf("invalid body type"),
			},
		},
		{
			Name:   "Body marshaled with existing content type header",
			Method: http.MethodPost,
			Format: "http://localhost:8080",
			Headers: map[string]string{
				"Content-Type": "application/merge-patch+json",
			},
			Body: map[string]any{
				"id":   123,
				"name": "John",
			},
			Marshaler:  NewJSONMarshaler(),
			Endpoint:   []byte(`http://localhost:8080`),
			Data:       []byte(`{"id":123,"name":"John"}`),
			CTHeader:   "application/merge-patch+json",
			CTHeaderOk: true,
			Errors:     []error{},
		},
		{
			Name:    "Failed to create request",
			Method:  "\n",
			Format:  "http://localhost:8080",
			Headers: map[string]string{},
			Body: map[string]any{
				"id":   123,
				"name": "John",
			},
			Marshaler:  NewJSONMarshaler(),
			Endpoint:   []byte(`http://localhost:8080`),
			Data:       nil,
			CTHeader:   "",
			CTHeaderOk: false,
			Errors: []error{
				fmt.Errorf("net/http: invalid method \"\\n\""),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			req := &Request{
				mem:       NewDefaultMemPool(),
				Method:    test.Method,
				Format:    test.Format,
				Body:      test.Body,
				Marshaler: test.Marshaler,
				Headers:   test.Headers,
				Errors:    []error{},
			}

			prepare(context.TODO(), req)

			if len(test.Errors) == 0 {
				assert.NotNil(t, req.Request)
				assert.NotNil(t, req.Request)
				assert.Equal(t, test.Endpoint, req.Endpoint)
				assert.Equal(t, test.Data, req.Data)
				assert.Equal(t, test.CTHeader, req.Request.Header.Get("Content-Type"))
			}
			assert.Equal(t, test.Errors, req.Errors)
		})
	}
}

// TestExecuteRequest tests if executing the request accurately sends a request
// after applying the headers and the response is accurately set
func TestExecuteRequest(t *testing.T) {
	tests := []struct {
		Name     string
		Client   *mockHttpHandler
		Timeout  time.Duration
		Method   string
		Endpoint []byte
		Errors   []error
	}{
		{
			Name: "Successful request",
			Client: &mockHttpHandler{
				dur:     time.Millisecond,
				code:    200,
				headers: map[string]string{},
			},
			Timeout:  time.Second,
			Method:   "GET",
			Endpoint: []byte(`http://localhost:8080?query=true`),
			Errors:   nil,
		},
		{
			Name: "Failed to make request",
			Client: &mockHttpHandler{
				dur:     time.Millisecond,
				code:    500,
				err:     fmt.Errorf("failed to make request"),
				headers: map[string]string{},
			},
			Timeout:  time.Second,
			Method:   "GET",
			Endpoint: []byte(`http://localhost:8080?query=true`),
			Errors:   []error{fmt.Errorf("failed to make request")},
		},
		{
			Name: "Making request times out",
			Client: &mockHttpHandler{
				dur:     time.Minute,
				code:    200,
				headers: map[string]string{},
			},
			Timeout:  time.Second,
			Method:   "GET",
			Endpoint: []byte(`http://localhost:8080`),
			Errors:   []error{context.DeadlineExceeded},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			ctx, cncl := context.WithTimeout(context.TODO(), test.Timeout)
			defer cncl()

			httpreq, _ := http.NewRequestWithContext(
				ctx, test.Method, string(test.Endpoint), bytes.NewReader(nil),
			)
			req := &Request{
				client:  test.Client,
				Request: httpreq,
			}

			execute(ctx, req)

			assert.Equal(t, test.Errors, req.Errors)
			if len(test.Errors) == 0 {
				assert.NotNil(t, req.Response)
			}
		})
	}
}

// TestFormatEndpoint tests
func TestFormatEndpoint(t *testing.T) {
	tests := []struct {
		Name        string
		Format      string
		QueryParams map[string][]string
		PathParams  []string
		Endpoint    []byte
		Error       error
	}{
		{
			Name:        "No Parameters",
			Format:      "http://localhost:8080",
			QueryParams: map[string][]string{},
			PathParams:  []string{},
			Endpoint:    []byte(`http://localhost:8080`),
			Error:       nil,
		},
		{
			Name:        "Using path parameters at the end",
			Format:      "http://localhost:8080/{}",
			QueryParams: map[string][]string{},
			PathParams: []string{
				"employees",
			},
			Endpoint: []byte(`http://localhost:8080/employees`),
			Error:    nil,
		},
		{
			Name:        "Using path parameters in the middle",
			Format:      "http://localhost:8080/{}/emails",
			QueryParams: map[string][]string{},
			PathParams: []string{
				"employees",
			},
			Endpoint: []byte(`http://localhost:8080/employees/emails`),
			Error:    nil,
		},
		{
			Name:        "Not enough path parameters",
			Format:      "http://localhost:8080/{}",
			QueryParams: map[string][]string{},
			PathParams:  []string{},
			Endpoint:    nil,
			Error:       fmt.Errorf("not enough parameters provided"),
		},
		{
			Name:        "Too many path parameters",
			Format:      "http://localhost:8080",
			QueryParams: map[string][]string{},
			PathParams: []string{
				"employees",
			},
			Endpoint: nil,
			Error:    fmt.Errorf("too many parameters provided"),
		},
		{
			Name:        "Incorrect path parameter format in the middle",
			Format:      "http://localhost:8080/{/emails",
			QueryParams: map[string][]string{},
			PathParams: []string{
				"employees",
			},
			Endpoint: nil,
			Error:    fmt.Errorf("illegal character/Invalid format in url"),
		},
		{
			Name:        "Incorrect path parameter format at the end",
			Format:      "http://localhost:8080/{",
			QueryParams: map[string][]string{},
			PathParams: []string{
				"employees",
			},
			Endpoint: nil,
			Error:    fmt.Errorf("illegal character/Invalid format in url"),
		},
		{
			Name:   "Using query parameters",
			Format: "http://localhost:8080",
			QueryParams: map[string][]string{
				"width": {"100"},
				// "height": {"100"}, Apparently the order can be random
			},
			PathParams: []string{},
			// Endpoint:   []byte("http://localhost:8080?width=100&height=100"),
			Endpoint: []byte("http://localhost:8080?width=100"),
			Error:    nil,
		},
		{
			Name:   "Using query parameter lists",
			Format: "http://localhost:8080",
			QueryParams: map[string][]string{
				"ids": {"123", "456", "789"},
			},
			PathParams: []string{},
			Endpoint:   []byte("http://localhost:8080?ids=123&ids=456&ids=789"),
			Error:      nil,
		},
		{
			Name:   "Having escaped characters",
			Format: "http://localhost:8080/{}",
			QueryParams: map[string][]string{
				"@id": {"1 3", "4!6"},
			},
			PathParams: []string{
				"employee mails",
			},
			Endpoint: []byte("http://localhost:8080/employee%20mails?%40id=1%203&%40id=4%216"),
			Error:    nil,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			req := &Request{
				mem: NewDefaultMemPool(),
			}

			endpoint, err := req.fmtEndpoint(
				test.Format,
				test.QueryParams,
				test.PathParams,
			)

			assert.Equal(t, test.Error, err)
			assert.Equal(t, test.Endpoint, endpoint)
		})
	}
}

// TestLockRequest tests locking the request mutex
func TestLockRequest(t *testing.T) {
	tests := []struct {
		Name string
	}{
		{Name: "Lock Mutex"},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			req := Request{
				mtx:    &sync.Mutex{},
				Values: map[string]any{},
			}

			req.Lock()

			ok := req.mtx.TryLock()
			assert.Equal(t, false, ok)
		})
	}
}

// TestUnlockRequest tests unlocking the request mutex
func TestUnlockRequest(t *testing.T) {
	tests := []struct {
		Name string
	}{
		{Name: "Unlock Mutex"},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			req := Request{
				mtx:    &sync.Mutex{},
				Values: map[string]any{},
			}
			req.mtx.Lock()

			req.Unlock()

			ok := req.mtx.TryLock()
			assert.Equal(t, true, ok)
		})
	}
}

// TestSetRequestValue tests setting a value in the request
func TestSetRequestValue(t *testing.T) {
	tests := []struct {
		Name   string
		Before map[string]any
		Key    string
		Value  string
		After  map[string]any
	}{
		{
			Name:   "Set value",
			Before: map[string]any{},
			Key:    "retries",
			Value:  "0",
			After: map[string]any{
				"retries": "0",
			},
		},
		{
			Name: "Overwrite value",
			Before: map[string]any{
				"retries": "0",
			},
			Key:   "retries",
			Value: "1",
			After: map[string]any{
				"retries": "1",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			req := Request{
				mtx:    &sync.Mutex{},
				Values: map[string]any{},
			}

			req.Set(test.Key, test.Value)

			val, ok := req.Values[test.Key]
			assert.Equal(t, true, ok)
			assert.Equal(t, test.Value, val)
		})
	}
}

// TestGetRequestValue tests getting a value in the request
func TestGetRequestValue(t *testing.T) {
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
				"retries": "0",
			},
			Key:    "retries",
			Exists: true,
			Value:  "0",
		},
		{
			Name:   "Value does not exist",
			Values: map[string]any{},
			Key:    "retries",
			Exists: false,
			Value:  nil,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			req := Request{
				mtx:    &sync.Mutex{},
				Values: test.Values,
			}

			val, ok := req.Get(test.Key)

			assert.Equal(t, test.Exists, ok)
			assert.Equal(t, test.Value, val)
		})
	}
}

package gent

import (
	"context"
	"fmt"
	"net/http"
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
		Retrier     Retrier
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
			Retrier:     NewBasicRetrier(0, func(int) time.Duration { return time.Second }),
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
				test.Context, test.MemPool, test.Retrier, test.Client,
				test.Format, test.Method, test.Body, test.Marshaler,
				test.Headers, test.QueryParams, test.PathParams,
				test.Functions,
			)

			assert.NotNil(t, req.ctx)
			assert.NotNil(t, req.mem)
			assert.NotNil(t, req.retr)
			assert.NotNil(t, req.client)
			assert.NotNil(t, req.fns)
			assert.NotNil(t, req.body)
			assert.NotNil(t, req.marshaler)
			assert.NotNil(t, req.headers)
			assert.NotNil(t, req.queryParams)
			assert.NotNil(t, req.pathParams)

			assert.Nil(t, req.endpoint)
			assert.Nil(t, req.data)
			assert.Nil(t, req.response)

			assert.Equal(t, 0, len(req.errors))
			assert.Equal(t, 0, req.fni)
			assert.Equal(t, test.Format, req.format)
			assert.Equal(t, test.Method, req.method)

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

			assert.Equal(t, test.Errors, req.errors)
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
				errors: test.Errors,
			}

			errs := req.Errors()

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
			Name:    "Marshaler is nil",
			Format:  "http://localhost:8080",
			Headers: map[string]string{},
			Body: map[string]any{
				"id":   123,
				"name": "John",
			},
			Marshaler:  nil,
			Endpoint:   []byte(`http://localhost:8080`),
			Data:       nil,
			CTHeader:   "",
			CTHeaderOk: false,
			Errors: []error{
				fmt.Errorf("marshaller is nil"),
			},
		},
		{
			Name:   "Body marshaled with existing content type header",
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
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			req := &Request{
				mem:       NewDefaultMemPool(),
				format:    test.Format,
				body:      test.Body,
				marshaler: test.Marshaler,
				headers:   test.Headers,
			}

			prepare(context.TODO(), req)
			hdr, _ := req.headers["Content-Type"]

			assert.Equal(t, test.Endpoint, req.endpoint)
			assert.Equal(t, test.Data, req.data)
			assert.Equal(t, test.CTHeader, hdr)
		})
	}
}

// TestExecuteRequest tests if executing the request accurately sends a request
// after applying the headers and the response is accurately set
func TestExecuteRequest(t *testing.T) {
	tests := []struct {
		Name     string
		Client   *mockHttpHandler
		Headers  map[string]string
		Timeout  time.Duration
		Method   string
		Endpoint []byte
		Data     []byte
		Errors   []error
	}{
		{
			Name: "Successful request",
			Client: &mockHttpHandler{
				dur:     time.Millisecond,
				code:    200,
				headers: map[string]string{},
			},
			Headers: map[string]string{
				"Authorization": "x.y.z",
				"X-Api-Key":     "awcef79a4wcn9fy",
			},
			Timeout:  time.Second,
			Method:   "GET",
			Endpoint: []byte(`http://localhost:8080?query=true`),
			Data:     []byte{},
			Errors:   nil,
		},
		{
			Name: "Failed to create request",
			Client: &mockHttpHandler{
				dur:     time.Millisecond,
				code:    200,
				headers: map[string]string{},
			},
			Headers:  map[string]string{},
			Timeout:  time.Second,
			Method:   "\n",
			Endpoint: nil,
			Data:     nil,
			Errors: []error{
				fmt.Errorf("net/http: invalid method \"\\n\""),
			},
		},
		{
			Name: "Failed to make request",
			Client: &mockHttpHandler{
				dur:     time.Millisecond,
				code:    500,
				err:     fmt.Errorf("failed to make request"),
				headers: map[string]string{},
			},
			Headers:  map[string]string{},
			Timeout:  time.Second,
			Method:   "GET",
			Endpoint: []byte(`http://localhost:8080?query=true`),
			Data:     []byte{},
			Errors: []error{
				fmt.Errorf("failed to make request"),
			},
		},
		{
			Name: "Making request times out",
			Client: &mockHttpHandler{
				dur:     time.Minute,
				code:    200,
				headers: map[string]string{},
			},
			Headers:  map[string]string{},
			Timeout:  time.Second,
			Method:   "GET",
			Endpoint: []byte(`http://localhost:8080`),
			Data:     []byte{},
			Errors: []error{
				context.DeadlineExceeded,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			req := &Request{
				mem:      NewDefaultMemPool(),
				client:   test.Client,
				method:   test.Method,
				endpoint: test.Endpoint,
				data:     test.Data,
				headers:  test.Headers,
			}

			ctx, cncl := context.WithTimeout(context.TODO(), test.Timeout)
			defer cncl()

			execute(ctx, req)

			assert.Equal(t, test.Errors, req.errors)
			assert.Equal(t, test.Endpoint, req.endpoint)
			assert.Equal(t, test.Data, req.data)
			assert.Equal(t, test.Data, test.Client.data)
			assert.Equal(t, test.Endpoint, test.Client.endpoint)
			assert.Equal(t, test.Headers, test.Client.headers)
			if len(test.Errors) == 0 {
				assert.NotNil(t, req.response)
			}
		})
	}
}

// TestExecuteRequestWithRetrier tests if executing the request accurately sends
// a request after applying the headers and the response is accurately set. It
// also checks if the request is retried on failure
func TestExecuteRequestWithRetrier(t *testing.T) {
	tests := []struct {
		Name     string
		Client   *mockHttpHandler
		Retrier  Retrier
		Headers  map[string]string
		Timeout  time.Duration
		Method   string
		Endpoint []byte
		Data     []byte
		Called   int
		Errors   []error
	}{
		{
			Name: "Successful request",
			Client: &mockHttpHandler{
				dur:     time.Millisecond,
				code:    200,
				headers: map[string]string{},
			},
			Retrier: NewBasicRetrier(5, func(int) time.Duration {
				return time.Millisecond
			}),
			Headers: map[string]string{
				"Authorization": "x.y.z",
				"X-Api-Key":     "awcef79a4wcn9fy",
			},
			Timeout:  time.Second,
			Method:   "GET",
			Endpoint: []byte(`http://localhost:8080?query=true`),
			Data:     []byte{},
			Called:   1,
			Errors:   []error(nil),
		},
		{
			Name: "Failed to create request",
			Client: &mockHttpHandler{
				dur:     time.Millisecond,
				code:    200,
				headers: map[string]string{},
			},
			Retrier: NewBasicRetrier(5, func(int) time.Duration {
				return time.Millisecond
			}),
			Headers:  map[string]string{},
			Timeout:  time.Second,
			Method:   "\n",
			Endpoint: nil,
			Data:     nil,
			Called:   0,
			Errors: []error{
				fmt.Errorf("net/http: invalid method \"\\n\""),
			},
		},
		{
			Name: "Failed to make request",
			Client: &mockHttpHandler{
				dur:     time.Millisecond,
				code:    500,
				err:     fmt.Errorf("failed to make request"),
				headers: map[string]string{},
			},
			Retrier: NewBasicRetrier(5, func(int) time.Duration {
				return time.Millisecond
			}),
			Headers:  map[string]string{},
			Timeout:  time.Second,
			Method:   "GET",
			Endpoint: []byte(`http://localhost:8080?query=true`),
			Data:     []byte{},
			Called:   6,
			Errors: []error{
				fmt.Errorf(
					"failed after max retries: %w",
					fmt.Errorf("failed to make request"),
				),
			},
		},
		{
			Name: "Making request times out",
			Client: &mockHttpHandler{
				dur:     time.Minute,
				code:    200,
				headers: map[string]string{},
			},
			Retrier: NewBasicRetrier(5, func(int) time.Duration {
				return time.Millisecond
			}),
			Headers:  map[string]string{},
			Timeout:  time.Second,
			Method:   "GET",
			Endpoint: []byte(`http://localhost:8080`),
			Data:     []byte{},
			Called:   1,
			Errors: []error{
				context.DeadlineExceeded,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			req := &Request{
				mem:      NewDefaultMemPool(),
				client:   test.Client,
				retr:     test.Retrier,
				method:   test.Method,
				endpoint: test.Endpoint,
				data:     test.Data,
				headers:  test.Headers,
			}

			ctx, cncl := context.WithTimeout(context.TODO(), test.Timeout)
			defer cncl()

			executeWithRetrier(ctx, req)

			assert.Equal(t, test.Errors, req.errors)
			assert.Equal(t, test.Endpoint, req.endpoint)
			assert.Equal(t, test.Data, req.data)
			assert.Equal(t, test.Data, test.Client.data)
			assert.Equal(t, test.Endpoint, test.Client.endpoint)
			assert.Equal(t, test.Headers, test.Client.headers)
			assert.Equal(t, test.Called, test.Client.called)
			if test.Errors == nil {
				assert.NotNil(t, req.response)
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

// TestGetMethod tests if request method can be fetched correctly
func TestGetMethod(t *testing.T) {
	tests := []struct {
		Name   string
		Method string
	}{
		{
			Name:   "Get Method",
			Method: "GET",
		},
		{
			Name:   "Post Method",
			Method: "POST",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			req := &Request{method: test.Method}
			val := req.GetMethod()
			assert.Equal(t, test.Method, val)
		})
	}
}

// TestGetHeader tests if request method can be fetched correctly
func TestGetHeader(t *testing.T) {
	tests := []struct {
		Name    string
		Headers map[string]string
		Key     string
		Value   string
		Exists  bool
	}{
		{
			Name: "Header exists",
			Headers: map[string]string{
				"Authorization": "Bearer x.y.z",
				"Content-Type":  "application/json",
			},
			Key:    "Authorization",
			Value:  "Bearer x.y.z",
			Exists: true,
		},
		{
			Name:    "Header does not exist",
			Headers: map[string]string{},
			Key:     "Authorization",
			Value:   "",
			Exists:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			req := &Request{headers: test.Headers}
			val, ok := req.GetHeader(test.Key)
			assert.Equal(t, test.Value, val)
			assert.Equal(t, test.Exists, ok)
		})
	}
}

// TestGetQueryParam tests if request query parameters can be fetched correctly
func TestGetQueryParam(t *testing.T) {
	tests := []struct {
		Name   string
		Params map[string][]string
		Key    string
		Value  []string
		Exists bool
	}{
		{
			Name: "Param exists",
			Params: map[string][]string{
				"ids": {"123", "456", "789"},
			},
			Key:    "ids",
			Value:  []string{"123", "456", "789"},
			Exists: true,
		},
		{
			Name:   "Param does not exist",
			Params: map[string][]string{},
			Key:    "ids",
			Value:  nil,
			Exists: false,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			req := &Request{queryParams: test.Params}
			val, ok := req.GetQueryParam(test.Key)
			assert.Equal(t, test.Value, val)
			assert.Equal(t, test.Exists, ok)
		})
	}
}

// TestGetEndpoint tests if request endpoint can be fetched correctly
func TestGetEndpoint(t *testing.T) {
	tests := []struct {
		Name     string
		Endpoint []byte
	}{
		{
			Name:     "Get Endpoint",
			Endpoint: []byte("http://localhost:8080"),
		},
		{
			Name:     "Endpoint is nil",
			Endpoint: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			req := &Request{endpoint: test.Endpoint}
			val := req.GetEndpoint()
			assert.Equal(t, test.Endpoint, val)
		})
	}
}

// TestGetData tests if request data can be fetched correctly
func TestGetData(t *testing.T) {
	tests := []struct {
		Name string
		Data []byte
	}{
		{
			Name: "Get Data",
			Data: []byte(`{"id":"123"}`),
		},
		{
			Name: "Data is nil",
			Data: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			req := &Request{data: test.Data}
			val := req.GetData()
			assert.Equal(t, test.Data, val)
		})
	}
}

// TestGetResponse tests if request response can be fetched correctly
func TestGetResponse(t *testing.T) {
	tests := []struct {
		Name     string
		Response *http.Response
	}{
		{
			Name:     "Get Response",
			Response: &http.Response{},
		},
		{
			Name:     "Response is nil",
			Response: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			req := &Request{response: test.Response}
			val := req.GetResponse()
			assert.Equal(t, test.Response, val)
		})
	}
}

// TestAddHeader tests if request headers can be added correctly
func TestAddHeader(t *testing.T) {
	tests := []struct {
		Name    string
		Headers map[string]string
		Key     string
		Value   string
	}{
		{
			Name:    "Header does not exist",
			Headers: map[string]string{},
			Key:     "Authorization",
			Value:   "Bearer x.y.z",
		},
		{
			Name: "Header already exists",
			Headers: map[string]string{
				"Authorization": "something",
			},
			Key:   "Authorization",
			Value: "Bearer x.y.z",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			req := &Request{headers: test.Headers}
			req.AddHeader(test.Key, test.Value)
			val, ok := req.headers[test.Key]
			assert.Equal(t, true, ok)
			assert.Equal(t, test.Value, val)
		})
	}
}

// TestRemoveHeader tests if request headers can be removed correctly
func TestRemoveHeader(t *testing.T) {
	tests := []struct {
		Name    string
		Headers map[string]string
		Key     string
	}{
		{
			Name:    "Header does not exist",
			Headers: map[string]string{},
			Key:     "Authorization",
		},
		{
			Name: "Header exists",
			Headers: map[string]string{
				"Authorization": "something",
			},
			Key: "Authorization",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			req := &Request{headers: test.Headers}
			req.RemoveHeader(test.Key)
			_, ok := req.headers[test.Key]
			assert.Equal(t, false, ok)
		})
	}
}

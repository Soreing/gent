package gent

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type mockHttpHandler struct {
	dur  time.Duration
	err  error
	code int

	data     []byte
	endpoint []byte
	headers  map[string]string
	called   int
}

func (m *mockHttpHandler) Do(r *http.Request) (*http.Response, error) {
	m.called++
	m.data, _ = ioutil.ReadAll(r.Body)
	m.endpoint = []byte(r.URL.String())

	for k, v := range r.Header {
		if len(v) > 0 {
			m.headers[k] = v[0]
		}
	}

	select {
	case <-r.Context().Done():
		return nil, r.Context().Err()
	case <-time.NewTimer(m.dur).C:
		if m.err != nil {
			return nil, m.err
		} else {
			rec := httptest.NewRecorder()
			res := rec.Result()
			res.StatusCode = m.code
			return res, nil
		}
	}
}

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

			if req.ctx == nil {
				t.Errorf("expected req.ctx to not be nil")
			}
			if req.mem == nil {
				t.Errorf("expected req.mem to not be nil")
			}
			if req.retr == nil {
				t.Errorf("expected req.retr to not be nil")
			}
			if req.client == nil {
				t.Errorf("expected req.client to not be nil")
			}
			if len(req.errors) != 0 {
				t.Errorf(
					"expected len(req.errors) to be %d, but it's %d",
					0, len(req.errors),
				)
			}
			if req.fns == nil {
				t.Errorf("expected req.fns to not be nil")
			}
			if req.fni != 0 {
				t.Errorf(
					"expected req.fni to be %d, but it's %d",
					0, req.fni,
				)
			}
			if req.format != test.Format {
				t.Errorf(
					"expected req.format to be %s, but it's %s",
					test.Format, req.format,
				)
			}
			if req.method != test.Method {
				t.Errorf(
					"expected req.method to be %s, but it's %s",
					test.Method, req.method,
				)
			}
			if req.body == nil {
				t.Errorf("expected req.body to not be nil")
			}
			if req.marshaler == nil {
				t.Errorf("expected req.marshaler to not be nil")
			}
			if req.headers == nil {
				t.Errorf("expected req.headers to not be nil")
			}
			if req.queryParams == nil {
				t.Errorf("expected req.queryParams to not be nil")
			}
			if req.pathParams == nil {
				t.Errorf("expected req.pathParams to not be nil")
			}
			if req.endpoint != nil {
				t.Errorf("expected req.pathParams to be nil")
			}
			if req.data != nil {
				t.Errorf("expected req.pathParams to be nil")
			}
			if req.response != nil {
				t.Errorf("expected req.pathParams to be nil")
			}
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
					val := c.Value("ch")
					ch := val.(chan int)
					ch <- 1
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			dlCtx, cncl := context.WithDeadline(
				context.TODO(),
				time.Now().Add(time.Second),
			)
			defer cncl()

			ch := make(chan int, 1)
			ctx := context.WithValue(dlCtx, "ch", ch)

			req := &Request{
				ctx: ctx,
				fns: test.Functions,
			}

			go func() {
				req.Next()
			}()

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

			if len(req.errors) != len(test.Errors) {
				t.Errorf(
					"expected len(req.errors) to be %d, but it's %d",
					len(test.Errors),
					len(req.errors),
				)
			} else {
				for i := range test.Errors {
					if req.errors[i].Error() != test.Errors[i].Error() {
						t.Errorf("expected contents of req.errors to match")
						break
					}
				}
			}
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

			if len(errs) != len(test.Errors) {
				t.Errorf(
					"expected len(errs) to be %d, but it's %d",
					len(test.Errors),
					len(errs),
				)
			} else {
				for i := range test.Errors {
					if errs[i].Error() != test.Errors[i].Error() {
						t.Errorf("expected contents of errs to match")
						break
					}
				}
			}
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
			Data:       nil,
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
			Data:       []byte(`{"id":123,"name":"John"}`),
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

			if string(req.endpoint) != string(test.Endpoint) {
				t.Errorf(
					"expected req.endpoint to be %s, but it's %s",
					string(test.Endpoint),
					string(req.endpoint),
				)
			}
			if string(req.data) != string(req.data) {
				t.Errorf(
					"expected req.data to be %s, but it's %s",
					string(test.Data),
					string(req.data),
				)
			}
			hdr, ok := req.headers["Content-Type"]
			if ok != test.CTHeaderOk {
				t.Errorf(
					"expected content type header state to be %t, but it's %t",
					test.CTHeaderOk,
					ok,
				)
			} else if test.CTHeaderOk && hdr != test.CTHeader {
				t.Errorf(
					"expected content type header to be %s, but it's %s",
					test.CTHeader,
					hdr,
				)
			}
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
			Data:     nil,
			Errors:   []error{},
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
			Data:     nil,
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
			Data:     nil,
			Errors: []error{
				fmt.Errorf("context deadline exceeded"),
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

			if len(req.errors) != len(test.Errors) {
				t.Errorf(
					"expected len(req.errors) to be %d, but it's %d",
					len(test.Errors),
					len(req.errors),
				)
			} else {
				for i := range test.Errors {
					if req.errors[i].Error() != test.Errors[i].Error() {
						t.Errorf("expected errors to match")
					}
				}
			}
			if len(test.Errors) == 0 && req.response == nil {
				t.Errorf("expected req.response to not be nil")
			}
			if string(test.Client.data) != string(test.Data) {
				t.Errorf(
					"expected sent data to be %s, but it's %s",
					string(test.Data),
					string(test.Client.data),
				)
			}
			if string(test.Client.endpoint) != string(test.Endpoint) {
				t.Errorf(
					"expected sent endpoint to be %s, but it's %s",
					string(test.Endpoint),
					string(test.Client.endpoint),
				)
			}

			if string(req.endpoint) != string(test.Endpoint) {
				t.Errorf(
					"expected req.endpoint to be %s, but it's %s",
					string(test.Endpoint),
					string(req.endpoint),
				)
			}
			if string(req.data) != string(req.data) {
				t.Errorf(
					"expected req.data to be %s, but it's %s",
					string(test.Data),
					string(req.data),
				)
			}
			for k, v := range test.Headers {
				if hv, ok := test.Client.headers[k]; !ok {
					t.Errorf("expected header[%s] to exist", k)
				} else if hv != v {
					t.Errorf(
						"expected header[%s] to be %s, but it's %s",
						k, v, hv,
					)
				}
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
			Data:     nil,
			Called:   1,
			Errors:   []error{},
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
			Data:     nil,
			Called:   6,
			Errors: []error{
				fmt.Errorf("failed after max retries: failed to make request"),
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
			Data:     nil,
			Called:   1,
			Errors: []error{
				fmt.Errorf("context deadline exceeded"),
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

			if len(req.errors) != len(test.Errors) {
				t.Errorf(
					"expected len(req.errors) to be %d, but it's %d",
					len(test.Errors),
					len(req.errors),
				)
			} else {
				for i := range test.Errors {
					if req.errors[i].Error() != test.Errors[i].Error() {
						t.Errorf("expected errors to match")
					}
				}
			}
			if len(test.Errors) == 0 && req.response == nil {
				t.Errorf("expected req.response to not be nil")
			}
			if string(test.Client.data) != string(test.Data) {
				t.Errorf(
					"expected sent data to be %s, but it's %s",
					string(test.Data),
					string(test.Client.data),
				)
			}
			if string(test.Client.endpoint) != string(test.Endpoint) {
				t.Errorf(
					"expected sent endpoint to be %s, but it's %s",
					string(test.Endpoint),
					string(test.Client.endpoint),
				)
			}
			if test.Client.called != test.Called {
				t.Errorf(
					"expected client called to be %d, but it's %d",
					test.Called,
					test.Client.called,
				)
			}
			if string(req.endpoint) != string(test.Endpoint) {
				t.Errorf(
					"expected req.endpoint to be %s, but it's %s",
					string(test.Endpoint),
					string(req.endpoint),
				)
			}
			if string(req.data) != string(req.data) {
				t.Errorf(
					"expected req.data to be %s, but it's %s",
					string(test.Data),
					string(req.data),
				)
			}
			for k, v := range test.Headers {
				if hv, ok := test.Client.headers[k]; !ok {
					t.Errorf("expected header[%s] to exist", k)
				} else if hv != v {
					t.Errorf(
						"expected header[%s] to be %s, but it's %s",
						k, v, hv,
					)
				}
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
				"width":  {"100"},
				"height": {"100"},
			},
			PathParams: []string{},
			Endpoint:   []byte("http://localhost:8080?width=100&height=100"),
			Error:      nil,
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

			if test.Error != nil && err == nil {
				t.Errorf("expected err to not be nil")
			} else if test.Error == nil && err != nil {
				t.Errorf("expected err to be nil")
			} else if test.Error != nil && err.Error() != test.Error.Error() {
				t.Errorf(
					"expected err to be %s but it's %s",
					test.Error.Error(),
					err.Error(),
				)
			}

			if string(endpoint) != string(test.Endpoint) {
				t.Errorf(
					"expected endpoint to be %s but it's %s",
					string(test.Endpoint),
					string(endpoint),
				)
			}
		})
	}
}

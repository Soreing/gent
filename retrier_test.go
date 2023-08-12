package gent

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"
)

// TestCreatingBasicRetrier tests the creation of a basic retrier
func TestCreatingBasicRetrier(t *testing.T) {
	tests := []struct {
		Name  string
		Max   int
		Delay func(int) time.Duration
	}{
		{
			Name: "Create basic retrier",
			Max:  2,
			Delay: func(int) time.Duration {
				return time.Second
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			ret := NewBasicRetrier(test.Max, test.Delay)

			if ret.retr == nil {
				t.Errorf("expected ret.retr to not be nil")
			}
			if ret.retryCodes != nil {
				t.Errorf("expected ret.retr to be nil")
			}
		})
	}
}

// TestCreatingStatusCodeRetrier tests the creation of a status code retrier
func TestCreatingStatusCodeRetrier(t *testing.T) {
	tests := []struct {
		Name  string
		Max   int
		Delay func(int) time.Duration
		Codes []int
	}{
		{
			Name: "Create status code retrier with nil codes",
			Max:  2,
			Delay: func(int) time.Duration {
				return time.Second
			},
			Codes: nil,
		},
		{
			Name: "Create status code retrier with status codes",
			Max:  2,
			Delay: func(int) time.Duration {
				return time.Second
			},
			Codes: []int{425, 418, 500},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			ret := NewStatusCodeRetrier(test.Max, test.Delay, test.Codes)

			if ret.retr == nil {
				t.Errorf("expected ret.retr to not be nil")
			}
			if test.Codes == nil && ret.retryCodes != nil {
				t.Errorf("expected ret.retryCodes to be nil")
			} else if test.Codes != nil && ret.retryCodes == nil {
				t.Errorf("expected ret.retryCodes to not be nil")
			} else if len(ret.retryCodes) != len(test.Codes) {
				t.Errorf(
					"expected len(ret.retryCodes) to be %d but it's %d",
					len(test.Codes),
					len(ret.retryCodes),
				)
			} else {
				for i, c := range test.Codes {
					if ret.retryCodes[i] != c {
						t.Errorf("expected contents of ret.retryCodes to match input")
					}
				}
			}
		})
	}
}

// TestRunTask tests if the retrier can run tasks
func TestRunTask(t *testing.T) {
	tests := []struct {
		Name  string
		Work  func(ctx context.Context) (error, bool)
		Error error
	}{
		{
			Name: "Running work function",
			Work: func(ctx context.Context) (error, bool) {
				val := ctx.Value("ch")
				ch := val.(chan int)
				ch <- 0
				return nil, false
			},
			Error: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			ret := NewBasicRetrier(0, func(int) time.Duration {
				return time.Second
			})

			dlCtx, cncl := context.WithDeadline(
				context.TODO(),
				time.Now().Add(time.Second),
			)
			defer cncl()

			ch := make(chan int, 1)
			ctx := context.WithValue(dlCtx, "ch", ch)
			err := ret.Run(ctx, test.Work)

			if err != test.Error {
				t.Errorf("expected err to be %v but it's %v", test.Error, err)
			} else {
				select {
				case <-ch:
				case <-ctx.Done():
					t.Errorf("waiting for result timed out")
				}
			}
		})
	}
}

// TestShouldRetryTask tests if the retrier can run tasks
func TestShouldRetryTask(t *testing.T) {
	tests := []struct {
		Name       string
		RetryCodes []int
		Response   *http.Response
		ErrorIn    error
		ErrorOut   error
		Retry      bool
	}{
		{
			Name:       "Successful request",
			RetryCodes: []int{},
			Response:   &http.Response{StatusCode: 200},
			ErrorIn:    nil,
			ErrorOut:   nil,
			Retry:      false,
		},
		{
			Name:       "Failed request with error",
			RetryCodes: []int{},
			Response:   nil,
			ErrorIn:    fmt.Errorf("request failed"),
			ErrorOut:   fmt.Errorf("request failed"),
			Retry:      true,
		},
		{
			Name:       "Failed request with retriable error code",
			RetryCodes: []int{500},
			Response:   &http.Response{StatusCode: 500},
			ErrorIn:    nil,
			ErrorOut:   fmt.Errorf("request failed with status code 500"),
			Retry:      true,
		},
		{
			Name:       "Failed request with non retriable error code",
			RetryCodes: []int{},
			Response:   &http.Response{StatusCode: 500},
			ErrorIn:    nil,
			ErrorOut:   fmt.Errorf("request failed with status code 500"),
			Retry:      false,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			ret := NewStatusCodeRetrier(0, func(int) time.Duration {
				return time.Second
			}, test.RetryCodes)

			err, retry := ret.ShouldRetry(test.Response, test.ErrorIn)

			if test.ErrorOut != nil && err == nil {
				t.Errorf("expected err to not be nil")
			} else if test.ErrorOut == nil && err != nil {
				t.Errorf("expected err to be nil")
			} else if test.ErrorOut != nil && err.Error() != test.ErrorOut.Error() {
				t.Errorf(
					"expected err to be %s but it's %s",
					test.ErrorOut.Error(),
					err.Error(),
				)
			}
			if retry != test.Retry {
				t.Errorf("expected retry to be %t but it's %t", test.Retry, retry)
			}
		})
	}
}

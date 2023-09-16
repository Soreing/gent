package gent

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

			assert.NotNil(t, ret.retr)
			assert.Nil(t, ret.retryCodes)
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

			assert.NotNil(t, ret.retr)
			assert.Equal(t, test.Codes, ret.retryCodes)
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
				ctx.Value("ch").(chan int) <- 0
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

			ch := make(chan int, 1)
			dl := time.Now().Add(time.Second)
			ctx := context.WithValue(context.TODO(), "ch", ch)
			ctx, cncl := context.WithDeadline(ctx, dl)
			defer cncl()

			err := ret.Run(ctx, test.Work)

			assert.Equal(t, test.Error, err)

			if test.Error == nil {
				select {
				case <-ctx.Done():
					t.Errorf("waiting for result timed out")
				case n := <-ch:
					assert.Equal(t, 0, n)
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

			assert.Equal(t, test.ErrorOut, err)
			assert.Equal(t, test.Retry, retry)
		})
	}
}

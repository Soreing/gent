package gent

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	sr "github.com/Soreing/retrier"
)

// retrier implements logic for for retrying requests
type retrier struct {
	retr       *sr.Retrier
	retryCodes []int
}

// NewBasicRetrier creates a retrier that retries requests up to an upper limit
// and waits for some duration between retries defined by the delay function.
// The basic retrier will only retry when the error is not nil.
func NewBasicRetrier(
	max int,
	delayf func(int) time.Duration,
) *retrier {
	return &retrier{
		retr: sr.NewRetrier(max, delayf),
	}
}

// NewStatusCodeRetrier creates a retrier that retries requests up to an upper
// limit and waits for some duration between retries defined by the delay
// function. The basic retrier will only retry when the error is not nil.
func NewStatusCodeRetrier(
	max int,
	delayf func(int) time.Duration,
	retryCodes []int,
) *retrier {
	return &retrier{
		retr:       sr.NewRetrier(max, delayf),
		retryCodes: retryCodes,
	}
}

// Run executes the task in the context of the retrier.
func (r *retrier) Run(
	ctx context.Context,
	work func(ctx context.Context) (error, bool),
) error {
	return r.retr.RunCtx(ctx, work)
}

// ShouldRetry evaluates whether the request should be retried based on the
// error and the response received. All errors are retried, and optionally
// status codes above 299 can be retried if they are in the retryable codes
// list.
func (r *retrier) ShouldRetry(
	res *http.Response,
	err error,
) (error, bool) {
	if err != nil {
		return err, true
	} else if res.StatusCode > 299 {
		e := fmt.Errorf(
			"request failed with status code " +
				strconv.Itoa(res.StatusCode),
		)

		for _, code := range r.retryCodes {
			if res.StatusCode == code {
				return e, true
			}
		}

		return e, false
	} else {
		return nil, false
	}
}

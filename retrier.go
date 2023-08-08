package gent

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	sr "github.com/Soreing/retrier"
)

type retrier struct {
	retr       *sr.Retrier
	retryCodes []int
}

func NewBasicRetrier(
	max int,
	delayf func(int) time.Duration,
) *retrier {
	return &retrier{
		retr: sr.NewRetrier(max, delayf),
	}
}

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

func (r *retrier) Run(
	ctx context.Context,
	work func(ctx context.Context) (error, bool),
) error {
	return r.retr.RunCtx(ctx, work)
}

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

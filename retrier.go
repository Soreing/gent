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
	retr *sr.Retrier
}

func NewRetrier(
	max int,
	delayf func(int) time.Duration,
) *retrier {
	return &retrier{
		retr: sr.NewRetrier(max, delayf),
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
		return fmt.Errorf(
			"request failed with status code " +
				strconv.Itoa(res.StatusCode),
		), true
	} else {
		return nil, false
	}
}

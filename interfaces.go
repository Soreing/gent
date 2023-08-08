package gent

import (
	"context"
	"net/http"
)

// HttpClient defines features of an internal http clients that is used to
// execute http requests.
type HttpClient interface {
	Do(r *http.Request) (*http.Response, error)
}

// MemoryPool defines a pool that can be used to acquire memory and release
// memory as byte arrays.
type MemoryPool interface {
	Acquire() []byte
	Release(...[]byte)
}

// Retrier defines a wrapper that can retry a request
type Retrier interface {
	Run(context.Context, func(ctx context.Context) (error, bool)) error
	ShouldRetry(*http.Response, error) (error, bool)
}

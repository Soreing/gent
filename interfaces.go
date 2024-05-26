package gent

import (
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

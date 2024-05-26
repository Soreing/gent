package gent

import (
	"net/http"
)

// Configuration is a collection of options that apply to the client.
type Configuration struct {
	mempool     MemoryPool
	httpClient  HttpClient
	newClientFn func() HttpClient
}

// newConfiguration creates default configs and applies options
func newConfiguration(opts []Option) *Configuration {
	cfg := &Configuration{
		httpClient: http.DefaultClient,
	}

	for _, opt := range opts {
		opt.Configure(cfg)
	}

	if cfg.mempool == nil {
		cfg.mempool = NewDefaultMemPool()
	}

	return cfg
}

// Option defines objects that can change a Configuration.
type Option interface {
	Configure(c *Configuration)
}

// UseMemoryPool creates an option for setting the client's memory pool.
func UseMemoryPool(pool MemoryPool) Option {
	return &memPoolOption{
		pool: pool,
	}
}

// UseHttpClient creates an option for setting the internal http client.
func UseHttpClient(client HttpClient) Option {
	return &httpClientOption{
		client: client,
	}
}

// UseHttpClientConstructor creates an option for setting the constructor to
// create a new http client for each request.
func UseHttpClientConstructor(constr func() HttpClient) Option {
	return &httpClientConstructorOption{
		constr: constr,
	}
}

type memPoolOption struct {
	pool MemoryPool
}

func (o *memPoolOption) Configure(c *Configuration) {
	c.mempool = o.pool
}

type httpClientOption struct {
	client HttpClient
}

func (o *httpClientOption) Configure(c *Configuration) {
	c.httpClient = o.client
}

type httpClientConstructorOption struct {
	constr func() HttpClient
}

func (o *httpClientConstructorOption) Configure(c *Configuration) {
	c.newClientFn = o.constr
}

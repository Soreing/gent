package gent

import (
	"context"
	"net/http"
)

// Configuration is a collection of options that apply to the client.
type Configuration struct {
	mempool       MemoryPool
	retrier       Retrier
	httpClient    HttpClient
	newClientFn   func() HttpClient
	hlmiddlewares []func(context.Context, *Request)
	llmiddlewares []func(context.Context, *Request)
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

// UseRetrier creates an option for adding a retrier that retries the request
// untill it succeeds.
func UseRetrier(retr Retrier) Option {
	return &retrierOption{
		retr: retr,
	}
}

// UseHighLevelMiddleware creates an option for adding a high level middleware
// that is executed before the endpoint url and request data are processed.
func UseHighLevelMiddleware(mdw func(context.Context, *Request)) Option {
	return &highLevelMiddlewareOption{
		mdw: mdw,
	}
}

// UseLowLevelMiddleware creates an option for adding a low level middleware
// that is executed before the headers are added and the request is executed.
func UseLowLevelMiddleware(mdw func(context.Context, *Request)) Option {
	return &lowLevelMiddlewareOption{
		mdw: mdw,
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

type retrierOption struct {
	retr Retrier
}

func (o *retrierOption) Configure(c *Configuration) {
	c.retrier = o.retr
}

type highLevelMiddlewareOption struct {
	mdw func(context.Context, *Request)
}

func (o *highLevelMiddlewareOption) Configure(c *Configuration) {
	c.hlmiddlewares = append(c.hlmiddlewares, o.mdw)
}

type lowLevelMiddlewareOption struct {
	mdw func(context.Context, *Request)
}

func (o *lowLevelMiddlewareOption) Configure(c *Configuration) {
	c.llmiddlewares = append(c.llmiddlewares, o.mdw)
}

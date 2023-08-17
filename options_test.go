package gent

import (
	"context"
	"testing"
	"time"
)

type mockMemPool struct {
	MemoryPool
}

type mockHttpClient struct {
	HttpClient
}

type mockRetrier struct {
	retrier
}

// TestMemoryPoolOption tests that memory pool options can be created and that
// they apply the configuration accurately.
func TestMemoryPoolOption(t *testing.T) {
	tests := []struct {
		Name string
	}{
		{Name: "Use memory pool"},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			pool := &mockMemPool{}
			opt := UseMemoryPool(pool)
			cfg := newConfiguration([]Option{opt})

			if cfg.mempool == nil {
				t.Errorf("expected cfg.mempool to not be nil")
			} else if _, ok := cfg.mempool.(*mockMemPool); !ok {
				t.Errorf(
					"expected cfg.mempool to be of type %T but it's %T",
					pool,
					cfg.mempool,
				)
			}
		})
	}
}

// TestHttpClientOption tests that http client options can be created and that
// they apply the configuration accurately.
func TestHttpClientOption(t *testing.T) {
	tests := []struct {
		Name string
	}{
		{Name: "Use http client"},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			client := &mockHttpClient{}
			opt := UseHttpClient(client)
			cfg := newConfiguration([]Option{opt})

			if cfg.httpClient == nil {
				t.Errorf("expected cfg.httpClient to not be nil")
			} else if _, ok := cfg.httpClient.(*mockHttpClient); !ok {
				t.Errorf(
					"expected cfg.httpClient to be of type %T but it's %T",
					client,
					cfg.httpClient,
				)
			}
		})
	}
}

// TestHttpClientConstructorOption tests that http client constructor options
// can be created and that they apply the configuration accurately.
func TestHttpClientConstructorOption(t *testing.T) {
	tests := []struct {
		Name string
	}{
		{Name: "Use http client constructor"},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			client := &mockHttpClient{}
			opt := UseHttpClientConstructor(func() HttpClient {
				return client
			})

			cfg := newConfiguration([]Option{opt})

			if cfg.newClientFn == nil {
				t.Errorf("expected cfg.httpClient to not be nil")
			} else {
				newClient := cfg.newClientFn()
				if _, ok := newClient.(*mockHttpClient); !ok {
					t.Errorf(
						"expected client to be of type %T but it's %T",
						client,
						newClient,
					)
				}
			}
		})
	}
}

// TestRetrierOption tests if the retrier option can be created and that it
// applies the configuration correctly
func TestRetrierOption(t *testing.T) {
	tests := []struct {
		Name string
	}{
		{Name: "Use retrier constructor"},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			retr := &mockRetrier{}
			opt := UseRetrier(retr)

			cfg := newConfiguration([]Option{opt})

			if cfg.retrier == nil {
				t.Errorf("expected cfg.retrier to not be nil")
			} else {
				if _, ok := cfg.retrier.(*mockRetrier); !ok {
					t.Errorf(
						"expected cfg.retrier to be of type %T but it's %T",
						retr,
						cfg.retrier,
					)
				}
			}
		})
	}
}

// TestLowLevelMiddlewareOption tests that low level middleware options can be
// created and that they apply the configuration accurately.
func TestLowLevelMiddlewareOption(t *testing.T) {
	tests := []struct {
		Name        string
		Middlewares []func(context.Context, *Request)
	}{
		{
			Name: "Use one low level middleware",
			Middlewares: []func(context.Context, *Request){
				func(c context.Context, r *Request) {
					val := c.Value("ch")
					ch := val.(chan int)
					ch <- 0
				},
			},
		},
		{
			Name: "Use multiple low level middlewares",
			Middlewares: []func(context.Context, *Request){
				func(c context.Context, r *Request) {
					val := c.Value("ch")
					ch := val.(chan int)
					ch <- 0
				},
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
			opts := []Option{}
			for _, mdw := range test.Middlewares {
				opts = append(opts, UseLowLevelMiddleware(mdw))
			}

			cfg := newConfiguration(opts)

			if cfg.llmiddlewares == nil {
				t.Errorf("expected cfg.llmiddlewares to not be nil")
			} else if len(cfg.llmiddlewares) != len(test.Middlewares) {
				t.Errorf(
					"expected len(cfg.llmiddlewares) to be of %d but it's %d",
					len(test.Middlewares),
					len(cfg.llmiddlewares),
				)
			} else {
				dlCtx, cncl := context.WithDeadline(
					context.TODO(),
					time.Now().Add(time.Second),
				)
				defer cncl()

				ch := make(chan int, len(cfg.llmiddlewares))
				ctx := context.WithValue(dlCtx, "ch", ch)

				go func() {
					for _, mdw := range cfg.llmiddlewares {
						mdw(ctx, nil)
					}
				}()

				for i := range cfg.llmiddlewares {
					select {
					case <-ctx.Done():
						t.Errorf("waiting for middlewares timed out")
					case n := <-ch:
						if n != i {
							t.Errorf("expected mdw value to be of %d but it's %d", i, n)
						}
					}
				}

			}
		})
	}
}

// TestHighLevelMiddlewareOption tests that high level middleware options can be
// created and that they apply the configuration accurately.
func TestHighLevelMiddlewareOption(t *testing.T) {
	tests := []struct {
		Name        string
		Middlewares []func(context.Context, *Request)
	}{
		{
			Name: "Use one high level middleware",
			Middlewares: []func(context.Context, *Request){
				func(c context.Context, r *Request) {
					val := c.Value("ch")
					ch := val.(chan int)
					ch <- 0
				},
			},
		},
		{
			Name: "Use multiple high level middlewares",
			Middlewares: []func(context.Context, *Request){
				func(c context.Context, r *Request) {
					val := c.Value("ch")
					ch := val.(chan int)
					ch <- 0
				},
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
			opts := []Option{}
			for _, mdw := range test.Middlewares {
				opts = append(opts, UseHighLevelMiddleware(mdw))
			}

			cfg := newConfiguration(opts)

			if cfg.hlmiddlewares == nil {
				t.Errorf("expected cfg.hlmiddlewares to not be nil")
			} else if len(cfg.hlmiddlewares) != len(test.Middlewares) {
				t.Errorf(
					"expected len(cfg.hlmiddlewares) to be of %d but it's %d",
					len(test.Middlewares),
					len(cfg.hlmiddlewares),
				)
			} else {
				dlCtx, cncl := context.WithDeadline(
					context.TODO(),
					time.Now().Add(time.Second),
				)
				defer cncl()

				ch := make(chan int, len(cfg.hlmiddlewares))
				ctx := context.WithValue(dlCtx, "ch", ch)

				go func() {
					for _, mdw := range cfg.hlmiddlewares {
						mdw(ctx, nil)
					}
				}()

				for i := range cfg.hlmiddlewares {
					select {
					case <-ctx.Done():
						t.Errorf("waiting for middlewares timed out")
					case n := <-ch:
						if n != i {
							t.Errorf("expected mdw value to be of %d but it's %d", i, n)
						}
					}
				}

			}
		})
	}
}

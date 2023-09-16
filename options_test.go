package gent

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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

			assert.NotNil(t, cfg.mempool)
			assert.Equal(t, cfg.mempool, pool)
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

			assert.NotNil(t, cfg.httpClient)
			assert.Equal(t, cfg.httpClient, client)
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
			opt := UseHttpClientConstructor(func() HttpClient { return client })
			cfg := newConfiguration([]Option{opt})

			if assert.NotNil(t, cfg.newClientFn) {
				newClient := cfg.newClientFn()
				assert.Equal(t, client, newClient)
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

			assert.NotNil(t, cfg.retrier)
			assert.Equal(t, retr, cfg.retrier)
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
					c.Value("ch").(chan int) <- 0
				},
			},
		},
		{
			Name: "Use multiple low level middlewares",
			Middlewares: []func(context.Context, *Request){
				func(c context.Context, r *Request) {
					c.Value("ch").(chan int) <- 0
				},
				func(c context.Context, r *Request) {
					c.Value("ch").(chan int) <- 1
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

			assert.NotNil(t, cfg.llmiddlewares)
			assert.Equal(t, len(test.Middlewares), len(cfg.llmiddlewares))

			ch := make(chan int, 1)
			dl := time.Now().Add(time.Second)
			ctx := context.WithValue(context.TODO(), "ch", ch)
			ctx, cncl := context.WithDeadline(ctx, dl)
			defer cncl()

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
					assert.Equal(t, i, n)
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
					c.Value("ch").(chan int) <- 0
				},
			},
		},
		{
			Name: "Use multiple high level middlewares",
			Middlewares: []func(context.Context, *Request){
				func(c context.Context, r *Request) {
					c.Value("ch").(chan int) <- 0
				},
				func(c context.Context, r *Request) {
					c.Value("ch").(chan int) <- 1
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

			assert.NotNil(t, cfg.hlmiddlewares)
			assert.Equal(t, len(test.Middlewares), len(cfg.hlmiddlewares))

			ch := make(chan int, 1)
			dl := time.Now().Add(time.Second)
			ctx := context.WithValue(context.TODO(), "ch", ch)
			ctx, cncl := context.WithDeadline(ctx, dl)
			defer cncl()

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
					assert.Equal(t, i, n)
				}
			}
		})
	}
}

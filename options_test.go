package gent

import (
	"testing"

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

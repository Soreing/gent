package gent

import "testing"

// TestCreateDefaultMempool tests that mempools can be created with default options
func TestCreateDefaultMempool(t *testing.T) {
	tests := []struct {
		Name string
	}{
		{
			Name: "Default memory pool",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {

			mempool := NewDefaultMemPool()

			if mempool == nil {
				t.Errorf("expected mempool to not be nil")
			} else {
				if mempool.pageSize != default_page_size {
					t.Errorf(
						"expected mempool.pageSize to be %d but it's %d",
						default_page_size,
						mempool.pageSize,
					)
				}
				if len(mempool.pool) != 0 {
					t.Errorf(
						"expected len(mempool.pool) to be %d but it's %d",
						0,
						len(mempool.pool),
					)
				}
				if cap(mempool.pool) != default_pool_size {
					t.Errorf(
						"expected cap(mempool.pool) to be %d but it's %d",
						default_pool_size,
						cap(mempool.pool),
					)
				}
			}
		})
	}
}

// TestCreateMempool tests that mempools can be created with options
func TestCreateMempool(t *testing.T) {
	tests := []struct {
		Name     string
		PageSize int
		PoolSize int
	}{
		{
			Name:     "New memory pool",
			PageSize: 512,
			PoolSize: 50,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {

			mempool := NewMemPool(test.PageSize, test.PoolSize)

			if mempool == nil {
				t.Errorf("expected mempool to not be nil")
			} else {
				if mempool.pageSize != test.PageSize {
					t.Errorf(
						"expected mempool.pageSize to be %d but it's %d",
						test.PageSize,
						mempool.pageSize,
					)
				}
				if len(mempool.pool) != 0 {
					t.Errorf(
						"expected len(mempool.pool) to be %d but it's %d",
						0,
						len(mempool.pool),
					)
				}
				if cap(mempool.pool) != test.PoolSize {
					t.Errorf(
						"expected cap(mempool.pool) to be %d but it's %d",
						test.PoolSize,
						cap(mempool.pool),
					)
				}
			}
		})
	}
}

// TestAcquireMemory tests that mempools can return memory pages
func TestAcquireMemory(t *testing.T) {
	tests := []struct {
		Name          string
		PageSize      int
		PoolSize      int
		PagesBefore   int
		PagesAcquired int
		PagesAfter    int
	}{
		{
			Name:          "Pool is empty",
			PageSize:      512,
			PoolSize:      5,
			PagesBefore:   0,
			PagesAcquired: 1,
			PagesAfter:    0,
		},
		{
			Name:          "Pool is populated",
			PageSize:      512,
			PoolSize:      5,
			PagesBefore:   3,
			PagesAcquired: 1,
			PagesAfter:    2,
		},
		{
			Name:          "Acquiring multiple pages",
			PageSize:      512,
			PoolSize:      5,
			PagesBefore:   5,
			PagesAcquired: 3,
			PagesAfter:    2,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {

			mempool := NewMemPool(test.PageSize, test.PoolSize)
			for i := 0; i < test.PagesBefore; i++ {
				mempool.pool <- make([]byte, test.PageSize)
			}

			var page []byte
			for i := 0; i < test.PagesAcquired; i++ {
				page = mempool.Acquire()
			}

			if page == nil {
				t.Errorf("expected page to not be nil")
			} else if cap(page) != test.PageSize {
				t.Errorf(
					"expected cap(page) to be %d but it's %d",
					test.PageSize,
					cap(page),
				)
			}
			if len(mempool.pool) != test.PagesAfter {
				t.Errorf(
					"expected len(mempool.pool) to be %d but it's %d",
					test.PagesAfter,
					len(mempool.pool),
				)
			}
		})
	}
}

// TestReleaseMemory tests that mempools can release memory pages
func TestReleaseMemory(t *testing.T) {
	tests := []struct {
		Name          string
		PageSize      int
		PoolSize      int
		PagesBefore   int
		PagesReleased int
		PagesAfter    int
	}{
		{
			Name:          "Pool is full",
			PageSize:      512,
			PoolSize:      5,
			PagesBefore:   5,
			PagesReleased: 1,
			PagesAfter:    5,
		},
		{
			Name:          "Pool has space",
			PageSize:      512,
			PoolSize:      5,
			PagesBefore:   2,
			PagesReleased: 1,
			PagesAfter:    3,
		},
		{
			Name:          "Releasing multiple pages",
			PageSize:      512,
			PoolSize:      5,
			PagesBefore:   1,
			PagesReleased: 3,
			PagesAfter:    4,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {

			mempool := NewMemPool(test.PageSize, test.PoolSize)
			for i := 0; i < test.PagesBefore; i++ {
				mempool.pool <- make([]byte, test.PageSize)
			}

			var pages [][]byte
			for i := 0; i < test.PagesReleased; i++ {
				pages = append(pages, make([]byte, test.PageSize))
			}
			mempool.Release(pages...)

			if len(mempool.pool) != test.PagesAfter {
				t.Errorf(
					"expected len(mempool.pool) to be %d but it's %d",
					test.PagesAfter,
					len(mempool.pool),
				)
			}
		})
	}
}

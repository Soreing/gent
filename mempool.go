package gent

const default_page_size = 256
const default_pool_size = 100

// MemPool manages pooling memory allocations for reusing dyanmic memory.
type MemPool struct {
	pageSize int
	pool     chan []byte
}

// NewDefaultMemPool creates a memory pool with default page and pool size.
func NewDefaultMemPool() *MemPool {
	return &MemPool{
		pageSize: default_page_size,
		pool:     make(chan []byte, default_pool_size),
	}
}

// NewMemPool creates a memory pool with a channel that provides byte arrays
// with the configured page size. The pool's channel has an upper limit of
// pool size.
func NewMemPool(
	pageSize int,
	poolSize int,
) *MemPool {
	return &MemPool{
		pageSize: pageSize,
		pool:     make(chan []byte, poolSize),
	}
}

// Acquire returns a byte array from the pool, or creates a new one if the
// pool is empty.
func (m *MemPool) Acquire() []byte {
	select {
	case buf := <-m.pool:
		return buf
	default:
		return make([]byte, 0, m.pageSize)
	}
}

// Release resets and inserts byte arrays into the pool. If the pool's channel
// is full, the byte array is discarded.
func (c *MemPool) Release(mem ...[]byte) {
	for i := range mem {
		select {
		case c.pool <- mem[i][:0]:
		default:
		}
	}
}

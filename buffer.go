package gent

// buffer manages pages of byte arrays and allows for expansion without the need
// to copy existing data into a larger allocation.
type buffer struct {
	page  []byte
	store [][]byte
	bytes int
}

// newBuffer creates a buffer with a provided byte array as the first page.
func newBuffer(page []byte) *buffer {
	return &buffer{
		page:  page,
		bytes: 0,
	}
}

// build merges the pages into a single byte array. If a buffer with sufficient
// space is provided, it will be used to store the result, otherwise a new
// byte array will be allocated.
func (b *buffer) build(buf []byte) []byte {
	size := b.bytes + len(b.page)

	var res []byte
	if cap(buf)-len(buf) >= size {
		res = buf
	} else {
		res = make([]byte, 0, size)
	}

	for _, b := range b.store {
		res = append(res, b...)
	}
	return append(res, b.page...)
}

// add stores the current page and sets the given byte array as the current
// page, expanding the buffer.
func (b *buffer) add(buf []byte) {
	b.store = append(b.store, b.page)
	b.bytes += len(b.page)
	b.page = buf
}

// len returns the total byte size of the buffer
func (b *buffer) len() int {
	return b.bytes + len(b.page)
}

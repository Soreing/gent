package gent

// writer manages a buffer with the ability to expand it using a memory pool
type writer struct {
	mem MemoryPool
	buf *buffer
}

// newWriter creates a new writer
func newWrirter(
	mem MemoryPool,
) writer {
	return writer{
		mem: mem,
		buf: newBuffer(mem.Acquire()),
	}
}

// writeByte writes a single byte to the buffer
func (w writer) writeByte(byt byte) {
	buf, mem := w.buf, w.mem
	if len(buf.page) < cap(buf.page) {
		buf.page = append(buf.page, byt)
	} else {
		newb := mem.Acquire()
		newb = append(newb, byt)
		buf.add(newb)
	}
}

// writeString writes a raw string to the buffer as it is
func (w writer) writeString(str string) {
	buf, mem := w.buf, w.mem
	space := cap(buf.page) - len(buf.page)
	if space >= len(str) {
		buf.page = append(buf.page, str...)
	} else {
		beg, end := space, space
		buf.page = append(buf.page, str[0:space]...)
		for beg < len(str) {
			newb := mem.Acquire()

			end += cap(newb)
			if end > len(str) {
				end = len(str)
			}

			newb = append(newb, str[beg:end]...)
			buf.add(newb)
			beg = end
		}
	}
}

// writeEscaped writes a url escaped string to the buffer
func (w writer) writeEscaped(str string) {
	beg, end := 0, 0
	for end < len(str) {
		if shouldEscape(str[end]) {
			w.writeString(str[beg:end])
			w.writeString(escape(str[end]))
			beg = end + 1
		}
		end++
	}
	w.writeString(str[beg:end])
}

// release releases all the pages held by the buffer
func (w writer) release() {
	buf, mem := w.buf, w.mem
	mem.Release(buf.page)
	mem.Release(buf.store...)
}

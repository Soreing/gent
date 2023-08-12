package gent

import (
	"testing"
)

// TestCreateBuffer tests that when the buffer is created, the current page
// should be set as the parameter that was provided in teh constructor
func TestCreateBuffer(t *testing.T) {
	tests := []struct {
		Name string
		Page []byte
	}{
		{
			Name: "Empty byte array",
			Page: make([]byte, 0, 256),
		},
		{
			Name: "Byte array with content",
			Page: []byte{1, 2, 3, 4, 5},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			buf := newBuffer(test.Page)
			if buf.page == nil {
				t.Errorf("expected buf.page to not be nil")
			} else {
				if len(buf.page) != len(test.Page) {
					t.Errorf(
						"expected len(buf.page) to be %d but it's %d",
						len(test.Page),
						len(buf.page),
					)
				}
				if cap(buf.page) != cap(test.Page) {
					t.Errorf(
						"expected cap(buf.page) to be %d but it's %d",
						len(test.Page),
						len(buf.page),
					)
				}
				for i := range test.Page {
					if test.Page[i] != buf.page[i] {
						t.Errorf("expected the contents of buf.page to match page")
					}
				}
			}
			if buf.store != nil {
				t.Errorf("expected buf.store to be nil")
			}
			if buf.bytes != 0 {
				t.Errorf("expected buf.bytes to be %d, but it's %d", 0, buf.bytes)
			}
		})
	}
}

// TestAddToBuffer tests that when a new byte array is added to the buffer, the
// byte arrays are stored, the length is tracked accurately and that the current
// page is set as the byte array just added to the buffer.
func TestAddToBuffer(t *testing.T) {
	tests := []struct {
		Name  string
		Pages [][]byte
	}{
		{
			Name:  "Empty byte array",
			Pages: [][]byte{make([]byte, 0, 256)},
		},
		{
			Name:  "Byte array with content",
			Pages: [][]byte{{1, 2, 3, 4, 5}},
		},
		{
			Name: "Multiple byte arrays with content",
			Pages: [][]byte{
				{1, 2, 3, 4, 5},
				{6, 7, 8},
				{9},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			buf := newBuffer([]byte{})
			bytes := buf.bytes

			for i, pg := range test.Pages {
				buf.add(pg)

				if buf.page == nil {
					t.Errorf("expected buf.page to not be nil")
				} else {
					if len(buf.page) != len(pg) {
						t.Errorf(
							"expected len(buf.page) to be %d but it's %d",
							len(pg),
							len(buf.page),
						)
					}
					if cap(buf.page) != cap(pg) {
						t.Errorf(
							"expected cap(buf.page) to be %d but it's %d",
							len(pg),
							len(buf.page),
						)
					}
					for i := range pg {
						if pg[i] != buf.page[i] {
							t.Errorf("expected the contents of buf.page to match pg")
						}
					}
				}

				if buf.store == nil {
					t.Errorf("expected buf.store to not be nil")
				} else if len(buf.store) != i+1 {
					t.Errorf(
						"expected len(buf.store) to be %d but it's %d",
						i+1,
						len(buf.store),
					)
				}

				if buf.bytes != bytes {
					t.Errorf("expected buf.bytes to be %d, but it's %d", bytes, buf.bytes)
				}

				bytes += len(pg)
			}
		})
	}
}

// TestBufferLength tests that the buffer length returns the correct length
// when the buffer expands
func TestBufferLength(t *testing.T) {
	tests := []struct {
		Name  string
		Pages [][]byte
	}{
		{
			Name: "Multiple byte arrays",
			Pages: [][]byte{
				{1, 2, 3, 4, 5},
				{},
				{6, 7, 8},
				{9},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			buf := newBuffer([]byte{})
			bytes := buf.bytes

			for _, pg := range test.Pages {
				buf.add(pg)
				bytes += len(pg)

				if buf.len() != bytes {
					t.Errorf("expected buf.len() to be %d, but it's %d", bytes, buf.len())
				}
			}
		})
	}
}

// TestBuildBuffer tests that the buffer builds the pages into the correct
// contiguous byte array
func TestBuildBuffer(t *testing.T) {
	tests := []struct {
		Name   string
		Pages  [][]byte
		Result []byte
	}{
		{
			Name:   "Empty byte array",
			Pages:  [][]byte{make([]byte, 0, 256)},
			Result: []byte(""),
		},
		{
			Name:   "One byte array",
			Pages:  [][]byte{[]byte("Test")},
			Result: []byte("Test"),
		},
		{
			Name: "Multiple byte arrays",
			Pages: [][]byte{
				[]byte("Hello, "),
				[]byte("World"),
				[]byte("!"),
			},
			Result: []byte("Hello, World!"),
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			var buf *buffer
			bytes := 0

			for i, pg := range test.Pages {
				bytes += len(pg)
				if i == 0 {
					buf = newBuffer(pg)
				} else {
					buf.add(pg)
				}
			}

			res := buf.build(nil)
			if len(res) != bytes {
				t.Errorf("expected len(res) to be %d, but it's %d", bytes, len(res))
			} else if string(res) != string(test.Result) {
				t.Errorf("expected res to be %s, but it's %s", string(test.Result), string(res))
			}
		})
	}
}

// TestBuildExistingBuffer tests that the buffer builds the pages into the
// provided byte array
func TestBuildExistingBuffer(t *testing.T) {
	tests := []struct {
		Name   string
		Slice  []byte
		Pages  [][]byte
		Result []byte
	}{
		{
			Name:   "No buffer provided",
			Slice:  nil,
			Pages:  [][]byte{make([]byte, 0, 256)},
			Result: []byte(""),
		},
		{
			Name:   "Buffer has sufficient size",
			Slice:  make([]byte, 0, 10),
			Pages:  [][]byte{[]byte("Test")},
			Result: []byte("Test"),
		},
		{
			Name:  "Buffer does without sufficient size",
			Slice: []byte{},
			Pages: [][]byte{
				[]byte("Hello, "),
				[]byte("World"),
				[]byte("!"),
			},
			Result: []byte("Hello, World!"),
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			var buf *buffer
			for i, pg := range test.Pages {
				if i == 0 {
					buf = newBuffer(pg)
				} else {
					buf.add(pg)
				}
			}

			res := buf.build(test.Slice)
			if string(res) != string(test.Result) {
				t.Errorf("expected res to be %s, but it's %s", string(test.Result), string(res))
			}
			if len(res) >= cap(test.Slice) && cap(res) != len(res) {
				t.Errorf("expected cap(res) to be %d, but it's %d", len(res), cap(res))
			} else if len(res) < cap(test.Slice) && cap(res) != cap(test.Slice) {
				t.Errorf("expected cap(res) to be %d, but it's %d", cap(test.Slice), cap(res))
			}
		})
	}
}

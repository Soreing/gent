package gent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCreateBuffer tests that when the buffer is created, the current page
// should be set as the parameter that was provided in the constructor
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

			if assert.NotNil(t, buf.page) {
				assert.Equal(t, len(test.Page), len(buf.page))
				assert.Equal(t, cap(test.Page), cap(buf.page))
				assert.Equal(t, test.Page, buf.page)
			}
			assert.Nil(t, buf.store)
			assert.Equal(t, 0, buf.bytes)
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

				if assert.NotNil(t, buf.page) {
					assert.Equal(t, len(pg), len(buf.page))
					assert.Equal(t, cap(pg), cap(buf.page))
					assert.Equal(t, pg, buf.page)
				}
				assert.NotNil(t, buf.store)
				assert.Equal(t, i+1, len(buf.store))
				assert.Equal(t, bytes, buf.bytes)

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

				assert.Equal(t, bytes, buf.len())
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
			Result: []byte(nil),
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

			assert.Equal(t, bytes, len(res))
			assert.Equal(t, test.Result, res)
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
			Result: []byte(nil),
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

			assert.Equal(t, test.Result, res)
			if len(res) >= cap(test.Slice) {
				assert.Equal(t, len(res), cap(res))
			} else {
				assert.Equal(t, cap(test.Slice), cap(res))
			}
		})
	}
}

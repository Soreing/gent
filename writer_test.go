package gent

import "testing"

// TestNewWriter tests if a writer can be created and that the object's fields
// are properly initialized
func TestNewWriter(t *testing.T) {
	tests := []struct {
		Name    string
		MemPool MemoryPool
	}{
		{
			Name:    "New writer",
			MemPool: NewDefaultMemPool(),
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			wrt := newWrirter(test.MemPool)

			if wrt.mem == nil {
				t.Errorf("expected wrt.mem to not be nil")
			}
			if wrt.buf == nil {
				t.Errorf("expected wrt.buf to not be nil")
			}
		})
	}
}

// TestWriteByte tests if a byte can be written into a buffer in any state
func TestWriteByte(t *testing.T) {
	tests := []struct {
		Name        string
		MemPool     MemoryPool
		InitialData string
		StoreLength int
		PageLength  int
	}{
		{
			Name:        "Write byte into empty buffer",
			MemPool:     NewMemPool(8, 100),
			InitialData: "",
			StoreLength: 0,
			PageLength:  1,
		},
		{
			Name:        "Write byte into buffer with space",
			MemPool:     NewMemPool(8, 100),
			InitialData: "Data",
			StoreLength: 0,
			PageLength:  5,
		},
		{
			Name:        "Write byte into full buffer",
			MemPool:     NewMemPool(8, 100),
			InitialData: "FullPage",
			StoreLength: 1,
			PageLength:  1,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			wrt := newWrirter(test.MemPool)
			wrt.buf.page = append(wrt.buf.page, test.InitialData...)
			wrt.writeByte('a')

			if len(wrt.buf.store) != test.StoreLength {
				t.Errorf(
					"expected len(wrt.buf.store) to be %d but it's %d",
					test.StoreLength,
					len(wrt.buf.store),
				)
			}
			if len(wrt.buf.page) != test.PageLength {
				t.Errorf(
					"expected len(wrt.buf.page) to be %d but it's %d",
					test.PageLength,
					len(wrt.buf.page),
				)
			}
		})
	}
}

// TestWriteString tests if a string can be written into a buffer in any state
func TestWriteString(t *testing.T) {
	tests := []struct {
		Name        string
		MemPool     MemoryPool
		InitialData string
		String      string
		StoreLength int
		PageLength  int
	}{
		{
			Name:        "Write string into empty buffer",
			MemPool:     NewMemPool(10, 100),
			InitialData: "",
			String:      "Test",
			StoreLength: 0,
			PageLength:  4,
		},
		{
			Name:        "Write string into buffer with space",
			MemPool:     NewMemPool(10, 100),
			InitialData: "Data",
			String:      "Test",
			StoreLength: 0,
			PageLength:  8,
		},
		{
			Name:        "Write string into full buffer",
			MemPool:     NewMemPool(10, 100),
			InitialData: "_FullPage_",
			String:      "Test",
			StoreLength: 1,
			PageLength:  4,
		},
		{
			Name:        "Write multi page string into buffer",
			MemPool:     NewMemPool(10, 100),
			InitialData: "",
			String:      "The Quick Brown Fox Jumped Over The Lazy Dog",
			StoreLength: 4,
			PageLength:  4,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			wrt := newWrirter(test.MemPool)
			wrt.buf.page = append(wrt.buf.page, test.InitialData...)
			wrt.writeString(test.String)

			if len(wrt.buf.store) != test.StoreLength {
				t.Errorf(
					"expected len(wrt.buf.store) to be %d but it's %d",
					test.StoreLength,
					len(wrt.buf.store),
				)
			}
			if len(wrt.buf.page) != test.PageLength {
				t.Errorf(
					"expected len(wrt.buf.page) to be %d but it's %d",
					test.PageLength,
					len(wrt.buf.page),
				)
			}
		})
	}
}

// TestWriteEscaped tests if a string can be written into a buffer
// and that invalid characters are escaped properly
func TestWriteEscaped(t *testing.T) {
	tests := []struct {
		Name    string
		MemPool MemoryPool
		String  string
		Result  string
	}{
		{
			Name:    "No escaped characters",
			MemPool: NewMemPool(10, 100),
			String:  "Test",
			Result:  "Test",
		},
		{
			Name:    "Escaped character at the end",
			MemPool: NewMemPool(10, 100),
			String:  "Test!",
			Result:  "Test%21",
		},
		{
			Name:    "Escaped character in the middle",
			MemPool: NewMemPool(10, 100),
			String:  "Another Test",
			Result:  "Another%20Test",
		},
		{
			Name:    "Multiple escaped characters",
			MemPool: NewMemPool(10, 100),
			String:  "Hello, World!",
			Result:  "Hello%2C%20World%21",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			wrt := newWrirter(test.MemPool)
			wrt.writeEscaped(test.String)

			res := string(wrt.buf.build(nil))

			if res != test.Result {
				t.Errorf(
					"expected result to be %s but it's %s",
					test.Result,
					res,
				)
			}
		})
	}
}

// TestReleaseWriter tests if a writer properly releases the entire buffer into
// the memory pool
func TestReleaseWriter(t *testing.T) {
	tests := []struct {
		Name    string
		MemPool MemoryPool
		Pages   int
	}{
		{
			Name:    "No new pages",
			MemPool: NewMemPool(10, 100),
			Pages:   0,
		},
		{
			Name:    "Some new pages",
			MemPool: NewMemPool(10, 100),
			Pages:   4,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			wrt := newWrirter(test.MemPool)
			for i := 0; i < test.Pages; i++ {
				wrt.buf.add([]byte{})
			}

			wrt.release()
			mem := wrt.mem.(*MemPool)

			if len(mem.pool) != test.Pages+1 {
				t.Errorf(
					"expected len(mem.pool) to be %d but it's %d",
					test.Pages+1,
					len(mem.pool),
				)
			}
		})
	}
}

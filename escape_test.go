package gent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestEscape tests the result of escaping bytes
func TestEscape(t *testing.T) {
	tests := []struct {
		Name string
		In   byte
		Out  string
	}{
		{
			Name: "Escape space",
			In:   ' ',
			Out:  "%20",
		},
		{
			Name: "Escape newline",
			In:   '\n',
			Out:  "%0A",
		},
		{
			Name: "Escape first",
			In:   '\u0000',
			Out:  "%00",
		},
		{
			Name: "Escape last",
			In:   '\u007f',
			Out:  "%7F",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			seq := escape(test.In)
			assert.Equal(t, test.Out, seq)
		})
	}
}

// TestShouldEscape tests if it's correctly identified that bytes should be escaped
func TestShouldEscape(t *testing.T) {
	tests := []struct {
		Name string
		In   byte
		Out  bool
	}{
		{
			Name: "Escape lowercase start",
			In:   'a',
			Out:  false,
		},
		{
			Name: "Escape lowercase end",
			In:   'z',
			Out:  false,
		},
		{
			Name: "Escape uppercase start",
			In:   'A',
			Out:  false,
		},
		{
			Name: "Escape uppercase end",
			In:   'Z',
			Out:  false,
		},
		{
			Name: "Escape digits start",
			In:   '0',
			Out:  false,
		},
		{
			Name: "Escape digits end",
			In:   '9',
			Out:  false,
		},
		{
			Name: "Escape dash",
			In:   '-',
			Out:  false,
		},
		{
			Name: "Escape underscore",
			In:   '_',
			Out:  false,
		},
		{
			Name: "Escape dot",
			In:   '.',
			Out:  false,
		},
		{
			Name: "Escape tilde",
			In:   '~',
			Out:  false,
		},
		{
			Name: "Escape anything else",
			In:   '@',
			Out:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			esc := shouldEscape(test.In)
			assert.Equal(t, test.Out, esc)
		})
	}
}
